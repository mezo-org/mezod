package bridgeworker

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
	"sync"
	"time"

	sdkmath "cosmossdk.io/math"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/mezo-org/mezod/bridge-worker/bitcoin"
	"github.com/mezo-org/mezod/ethereum/bindings/portal"
	"github.com/mezo-org/mezod/ethereum/bindings/tbtc"
	"github.com/mezo-org/mezod/utils"
	bridgetypes "github.com/mezo-org/mezod/x/bridge/types"
)

const (
	// bitcoinTargetChain is a numerical value representing Bitcoin target chain.
	bitcoinTargetChain = uint8(1)

	// assetsUnlockConfirmedLookBackBlocks is the number of blocks used when
	// fetching AssetsUnlockConfirmed events from the Ethereum chain. It defines
	// how far back we look when searching for events.
	assetsUnlockConfirmedLookBackBlocks = 216000 // ~30 days

	// assetsUnlockConfirmedProcessingPeriod is the time period defining how
	// frequently new AssetsUnlockConfirmed events should be processed.
	assetsUnlockConfirmedProcessingPeriod = 1 * time.Minute

	// btcWithdrawalProcessBackoff is a backoff time used between retries when
	// submitting a withdrawBTC transaction.
	btcWithdrawalProcessBackoff = 1 * time.Minute

	// liveWalletsUpdatePeriod defines how frequently we should update the list
	// of live wallets.
	liveWalletsUpdatePeriod = 1 * time.Hour

	// walletStateLive represents Live wallet state from the tBTC Bridge contract.
	walletStateLive = uint8(1)
)

// AssetsUnlockedEndpoint is a client enabling communication with the `Mezo`
// chain.
type AssetsUnlockedEndpoint interface {
	// GetAssetsUnlockedEvents gets the AssetsUnlocked events from the Mezo
	// chain. The requested range of events is inclusive on the lower side and
	// exclusive on the upper side.
	GetAssetsUnlockedEvents(
		ctx context.Context,
		sequenceStart sdkmath.Int,
		sequenceEnd sdkmath.Int,
	) ([]bridgetypes.AssetsUnlockedEvent, error)
}

func newBTCWithdrawalJob(
	env *environment,
	assetsUnlockEndpoint string,
	queueCheckFrequency time.Duration,
) *btcWithdrawalJob {
	// The messages handled by the bridge-worker contain custom types.
	// Add codecs so that the messages can be marshaled/unmarshalled.
	assetsUnlockedGrpcEndpoint, err := NewAssetsUnlockedGrpcEndpoint(
		assetsUnlockEndpoint,
		codectypes.NewInterfaceRegistry(),
	)
	if err != nil {
		panic(fmt.Sprintf("failed to create assets unlocked endpoint: %v", err))
	}

	go func() {
		// Test the connection to the assets unlocked endpoint by verifying we
		// can successfully execute `GetAssetsUnlockedEvents`.
		ctxWithTimeout, cancel := context.WithTimeout(
			context.Background(),
			requestTimeout,
		)
		defer cancel()

		_, err := assetsUnlockedGrpcEndpoint.GetAssetsUnlockedEvents(
			ctxWithTimeout,
			sdkmath.NewInt(1),
			sdkmath.NewInt(2),
		)
		if err != nil {
			env.logger.Error(
				"assets unlocked endpoint connection test failed; possible "+
					"problem with configuration or connectivity",
				"error", err,
			)
		} else {
			env.logger.Info(
				"assets unlocked endpoint connection test completed successfully",
			)
		}
	}()

	redemptionParameters, err := env.tbtcBridgeContract.RedemptionParameters()
	if err != nil {
		panic(fmt.Sprintf("failed to get tBTC bridge redemption parameters: %v", err))
	}

	redemptionDustThresholdErc20Precision := btcToErc20Amount(
		new(big.Int).SetUint64(redemptionParameters.RedemptionDustThreshold),
	)

	tbtcToken, err := env.mezoBridgeContract.TbtcToken()
	if err != nil {
		panic(fmt.Sprintf("failed to get tBTC token: %v", err))
	}

	return &btcWithdrawalJob{
		env:                                   env,
		assetsUnlockedEndpoint:                assetsUnlockedGrpcEndpoint,
		liveWalletsReady:                      make(chan struct{}),
		btcWithdrawalQueue:                    []portal.MezoBridgeAssetsUnlockConfirmed{},
		btcWithdrawalFinalityChecks:           map[string]*btcWithdrawalFinalityCheck{},
		btcWithdrawalQueueCheckFrequency:      queueCheckFrequency,
		redemptionDustThresholdErc20Precision: redemptionDustThresholdErc20Precision,
		tbtcToken:                             tbtcToken,
	}
}

type btcWithdrawalJob struct {
	env *environment

	assetsUnlockedEndpoint AssetsUnlockedEndpoint

	// Channel used to indicate whether the initial fetching of live wallets
	// is done. Once this channel is closed we can proceed with the Bitcoin
	// withdrawing routines.
	liveWalletsReady chan struct{}

	liveWalletsMutex sync.Mutex
	liveWallets      [][20]byte

	liveWalletsLastProcessedBlock uint64 // single-routine use; no mutex locking needed.

	btcWithdrawalLastProcessedBlock uint64

	btcWithdrawalMutex sync.Mutex
	btcWithdrawalQueue []portal.MezoBridgeAssetsUnlockConfirmed

	btcWithdrawalFinalityChecksMutex sync.Mutex
	btcWithdrawalFinalityChecks      map[string]*btcWithdrawalFinalityCheck

	btcWithdrawalQueueCheckFrequency time.Duration

	// The `redemptionDustThreshold` and `tbtcToken` parameters change very
	// rarely. Therefore we only read them only once, at program start.
	redemptionDustThresholdErc20Precision *big.Int
	tbtcToken                             common.Address
}

func (bwj *btcWithdrawalJob) run(ctx context.Context) {
	runCtx, cancelRunCtx := context.WithCancel(ctx)
	defer cancelRunCtx()

	go func() {
		defer cancelRunCtx()
		err := bwj.observeLiveWallets(runCtx)
		if err != nil {
			bwj.env.logger.Error(
				"live wallets observation routine failed",
				"error", err,
			)
		}

		bwj.env.logger.Warn("live wallets observation routine stopped")
	}()

	// Wait until the initial fetching of live wallets is done.
	bwj.env.logger.Info("waiting for initial fetching of live wallet")
	select {
	case <-bwj.liveWalletsReady:
		bwj.env.logger.Info("initial fetching of live wallets completed")
	case <-runCtx.Done():
		bwj.env.logger.Warn(
			"context canceled while waiting; exiting without launching " +
				"BTC withdrawal routines",
		)
		return
	}

	go func() {
		defer cancelRunCtx()
		err := bwj.observeBTCWithdrawals(runCtx)
		if err != nil {
			bwj.env.logger.Error(
				"BTC withdrawal observation routine failed",
				"error", err,
			)
		}

		bwj.env.logger.Warn("BTC withdrawal observation routine stopped")
	}()

	go func() {
		defer cancelRunCtx()
		bwj.processBTCWithdrawalQueue(runCtx)
		bwj.env.logger.Warn("BTC withdrawal processing loop stopped")
	}()

	go func() {
		defer cancelRunCtx()
		bwj.processBTCWithdrawalFinalityChecks(runCtx)
		bwj.env.logger.Warn("BTC withdrawal finality checks loop stopped")
	}()

	<-runCtx.Done()
	bwj.env.logger.Info("BTC withdrawal job stopped")
}

func (bwj *btcWithdrawalJob) observeLiveWallets(ctx context.Context) error {
	finalizedBlock, err := bwj.env.chain.FinalizedBlock(ctx)
	if err != nil {
		return fmt.Errorf("failed to get finalized block: [%w]", err)
	}
	endBlock := finalizedBlock.Uint64()

	recentEvents, err := utils.WithBatchEventFetch(
		bwj.newWalletRegisteredEvents,
		0,
		endBlock,
		bwj.env.requestsPerMinute,
		bwj.env.batchSize,
	)
	if err != nil {
		return fmt.Errorf(
			"failed to fetch NewWalletRegistered events: [%w]",
			err,
		)
	}

	liveWallets := [][20]byte{}
	for _, event := range recentEvents {
		walletPublicKeyHash := event.WalletPubKeyHash

		wallet, err := bwj.env.tbtcBridgeContract.Wallets(walletPublicKeyHash)
		if err != nil {
			return fmt.Errorf("failed to get wallet: [%w]", err)
		}

		if wallet.State == walletStateLive {
			liveWallets = append(liveWallets, walletPublicKeyHash)
		}
	}

	bwj.liveWalletsLastProcessedBlock = endBlock

	bwj.env.logger.Info(
		"finished initial search for live wallets",
		"number_of_live_wallets", len(liveWallets),
	)

	bwj.liveWalletsMutex.Lock()
	bwj.liveWallets = liveWallets
	bwj.liveWalletsMutex.Unlock()

	// Signal that initial fetching of wallets ready.
	close(bwj.liveWalletsReady)

	// Start a ticker to periodically update the wallets.
	ticker := time.NewTicker(liveWalletsUpdatePeriod)
	defer ticker.Stop()
	tickerChan := ticker.C

	for {
		select {
		case <-ctx.Done(): // Handle context cancellation
			bwj.env.logger.Warn(
				"stopping live wallet observation due to context cancellation",
			)
			return nil
		case <-tickerChan:
			err := bwj.updateLiveWallets(ctx)
			if err != nil {
				bwj.env.logger.Error(
					"failed to update live wallets",
					"error", err,
				)
			}
		}
	}
}

func (bwj *btcWithdrawalJob) updateLiveWallets(ctx context.Context) error {
	// Work on a copy of live wallets to avoid blocking for too long.
	// Live wallets are represented by `[20]byte` arrays, so `copy` will
	// deep-copy them.
	bwj.liveWalletsMutex.Lock()

	walletPublicKeyHashes := make([][20]byte, len(bwj.liveWallets))
	copy(walletPublicKeyHashes, bwj.liveWallets)

	bwj.liveWalletsMutex.Unlock()

	// Keep only wallets which are still live.
	updatedLiveWallets := [][20]byte{}

	for _, walletPublicKeyHash := range walletPublicKeyHashes {
		wallet, err := bwj.env.tbtcBridgeContract.Wallets(walletPublicKeyHash)
		if err != nil {
			return fmt.Errorf("failed to get wallet: [%w]", err)
		}

		if wallet.State == walletStateLive {
			updatedLiveWallets = append(updatedLiveWallets, walletPublicKeyHash)
		}
	}

	// Look for new live wallets.
	finalizedBlock, err := bwj.env.chain.FinalizedBlock(ctx)
	if err != nil {
		return fmt.Errorf("failed to get finalized block: [%w]", err)
	}
	endBlock := finalizedBlock.Uint64()

	if endBlock > bwj.liveWalletsLastProcessedBlock {
		events, err := utils.WithBatchEventFetch(
			bwj.newWalletRegisteredEvents,
			bwj.liveWalletsLastProcessedBlock+1,
			endBlock,
			bwj.env.requestsPerMinute,
			bwj.env.batchSize,
		)
		if err != nil {
			return fmt.Errorf(
				"failed to fetch NewWalletRegistered events: [%w]",
				err,
			)
		}

		for _, event := range events {
			walletPublicKeyHash := event.WalletPubKeyHash

			wallet, err := bwj.env.tbtcBridgeContract.Wallets(walletPublicKeyHash)
			if err != nil {
				return fmt.Errorf("failed to get wallet: [%w]", err)
			}

			if wallet.State == walletStateLive {
				updatedLiveWallets = append(updatedLiveWallets, walletPublicKeyHash)
			}
		}

		bwj.liveWalletsLastProcessedBlock = endBlock
	}

	bwj.env.logger.Info(
		"finished updating live wallets",
		"number_of_live_wallets", len(updatedLiveWallets),
	)

	bwj.liveWalletsMutex.Lock()
	bwj.liveWallets = updatedLiveWallets
	bwj.liveWalletsMutex.Unlock()

	return nil
}

// btcWithdrawalFinalityCheck is a struct that contains an AssetsUnlockConfirmed
// and the Ethereum height at which the BTC withdrawal finality check for it was
// scheduled.
type btcWithdrawalFinalityCheck struct {
	event             *portal.MezoBridgeAssetsUnlockConfirmed
	scheduledAtHeight *big.Int // nil means unscheduled
}

func (wfc *btcWithdrawalFinalityCheck) key() string {
	return wfc.event.UnlockSequenceNumber.String()
}

// observeBTCWithdrawals monitors AssetsUnlockConfirmed events, filters
// events representing pending BTC withdrawals and puts them into a queue.
func (bwj *btcWithdrawalJob) observeBTCWithdrawals(ctx context.Context) error {
	// Use the current block rather than finalized block to speed up event
	// processing. Event processing should handle the effects of a possible
	// reorg (e.g. an event being duplicated in a queue), although there is a
	// small risk of skipping an event. In that case it would require a manual
	// execution of `withdrawBTC`.
	currentBlock, err := bwj.env.chain.BlockCounter().CurrentBlock()
	if err != nil {
		return fmt.Errorf("failed to get current block: [%w]", err)
	}

	startBlock := uint64(0)
	if currentBlock > assetsUnlockConfirmedLookBackBlocks {
		startBlock = currentBlock - assetsUnlockConfirmedLookBackBlocks
	}

	recentEvents, err := utils.WithBatchEventFetch(
		bwj.assetsUnlockConfirmedEvents,
		startBlock,
		currentBlock,
		bwj.env.requestsPerMinute,
		bwj.env.batchSize,
	)
	if err != nil {
		return fmt.Errorf(
			"failed to fetch AssetsUnlockConfirmed events: [%w]",
			err,
		)
	}

	pendingBTCWithdrawalCount := 0

	for _, event := range recentEvents {
		isPendingBTCWithdrawal, err := bwj.isPendingBTCWithdrawal(ctx, event)
		if err != nil {
			return fmt.Errorf(
				"failed to check if event represents pending BTC withdrawal: [%w]",
				err,
			)
		}

		if isPendingBTCWithdrawal {
			bwj.enqueueBTCWithdrawal(event)
			pendingBTCWithdrawalCount++
		}
	}

	bwj.btcWithdrawalLastProcessedBlock = currentBlock

	bwj.env.logger.Info(
		"initial search for pending BTC withdrawals done",
		"assets_unlock_confirmed_events", len(recentEvents),
		"pending_btc_withdrawals", pendingBTCWithdrawalCount,
	)

	// Start a ticker to process the new events periodically.
	ticker := time.NewTicker(assetsUnlockConfirmedProcessingPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done(): // Handle context cancellation
			bwj.env.logger.Warn(
				"stopping BTC withdrawals routine due to context cancellation",
			)
			return nil
		case <-ticker.C:
			// Process incoming AssetsUnlockedConfirmed events
			err := bwj.processNewAssetsUnlockConfirmedEvents(ctx)
			if err != nil {
				bwj.env.logger.Error(
					"failed to process newly emitted AssetsUnlockConfirmed events",
					"error", err,
				)
			}
		}
	}
}

// enqueueBTCWithdrawal puts AssetsUnlockConfirmed events representing BTC
// withdrawals at the end of the queue.
func (bwj *btcWithdrawalJob) enqueueBTCWithdrawal(
	event *portal.MezoBridgeAssetsUnlockConfirmed,
) {
	bwj.btcWithdrawalMutex.Lock()
	defer bwj.btcWithdrawalMutex.Unlock()

	bwj.btcWithdrawalQueue = append(bwj.btcWithdrawalQueue, *event)

	bwj.env.logger.Debug(
		"enqueued BTC withdrawal",
		"unlock_sequence", event.UnlockSequenceNumber.String(),
	)
}

// dequeueBTCWithdrawal removes an AssetsUnlockConfirmed event representing
// a BTC withdrawal from the front of the queue.
func (bwj *btcWithdrawalJob) dequeueBTCWithdrawal() *portal.MezoBridgeAssetsUnlockConfirmed {
	bwj.btcWithdrawalMutex.Lock()
	defer bwj.btcWithdrawalMutex.Unlock()
	if len(bwj.btcWithdrawalQueue) == 0 {
		return nil
	}

	event := bwj.btcWithdrawalQueue[0]
	bwj.btcWithdrawalQueue = bwj.btcWithdrawalQueue[1:]

	bwj.env.logger.Debug(
		"dequeued BTC withdrawal",
		"unlock_sequence", event.UnlockSequenceNumber.String(),
	)

	return &event
}

// isPendingBTCWithdrawal checks whether an AssetsUnlockConfirmed event
// represents a pending BTC withdrawal.
func (bwj *btcWithdrawalJob) isPendingBTCWithdrawal(
	ctx context.Context,
	event *portal.MezoBridgeAssetsUnlockConfirmed,
) (bool, error) {
	if event.Chain != bitcoinTargetChain {
		return false, nil
	}

	// When calculating the attestation key, we must get the recipient field
	// from `mezod`. We cannot read it from the passed `AssetsUnlockConfirmed`
	// event as it only contains a hash of recipient.
	unlockSequence := event.UnlockSequenceNumber

	assetsUnlockEvents, err := bwj.assetsUnlockedEndpoint.GetAssetsUnlockedEvents(
		ctx,
		sdkmath.NewIntFromBigInt(new(big.Int).Set(unlockSequence)),
		sdkmath.NewIntFromBigInt(new(big.Int).Add(
			new(big.Int).Set(unlockSequence),
			big.NewInt(1)),
		),
	)
	if err != nil {
		return false, fmt.Errorf(
			"failed to get AssetsUnlock event from Mezo: [%w]",
			err,
		)
	}

	if len(assetsUnlockEvents) != 1 {
		return false, fmt.Errorf(
			"incorrect number of AssetsUnlock events returned from Mezo",
		)
	}

	recipient := assetsUnlockEvents[0].Recipient

	hash, err := computeAttestationKey(
		event.UnlockSequenceNumber,
		recipient,
		event.Token,
		event.Amount,
		event.Chain,
	)
	if err != nil {
		return false, fmt.Errorf(
			"failed to calculate attestation key: [%w]",
			err,
		)
	}

	isPendingBTCWithdrawal, err := bwj.env.mezoBridgeContract.PendingBTCWithdrawals(hash)
	if err != nil {
		return false, fmt.Errorf(
			"failed to get pending BTC withdrawals info: [%w]",
			err,
		)
	}

	if !isPendingBTCWithdrawal {
		return false, nil
	}

	if event.Amount.Cmp(bwj.redemptionDustThresholdErc20Precision) < 0 {
		bwj.env.logger.Warn(
			"found BTC withdrawal below redemption dust threshold",
			"unlock_sequence", event.UnlockSequenceNumber.String(),
			"amount", event.Amount.String(),
			"redemption_dust_threshold_erc20_precision", bwj.redemptionDustThresholdErc20Precision.String(),
		)

		return false, nil
	}

	return true, nil
}

// processNewAssetsUnlockConfirmedEvents fetches new AssetsUnlockConfirmed
// events representing pending BTC withdrawals and puts them into a queue.
// It is intended to be run periodically.
func (bwj *btcWithdrawalJob) processNewAssetsUnlockConfirmedEvents(ctx context.Context) error {
	// Use the current block rather than finalized block to speed up event
	// processing. Event processing should handle the effects of a possible
	// reorg (e.g. an event being duplicated in a queue), although there is a
	// small risk of skipping an event. In that case it would require a manual
	// execution of `withdrawBTC`.
	currentBlock, err := bwj.env.chain.BlockCounter().CurrentBlock()
	if err != nil {
		return fmt.Errorf("cannot get current block: [%w]", err)
	}

	newPendingBTCWithdrawals := 0

	if currentBlock > bwj.btcWithdrawalLastProcessedBlock {
		events, err := utils.WithBatchEventFetch(
			bwj.assetsUnlockConfirmedEvents,
			bwj.btcWithdrawalLastProcessedBlock+1,
			currentBlock,
			bwj.env.requestsPerMinute,
			bwj.env.batchSize,
		)
		if err != nil {
			return fmt.Errorf(
				"failed to fetch AssetsUnlockConfirmed events: [%w]",
				err,
			)
		}

		for _, event := range events {
			isPendingBTCWithdrawal, err := bwj.isPendingBTCWithdrawal(ctx, event)
			if err != nil {
				return fmt.Errorf(
					"failed to check if event represents pending BTC "+
						"withdrawal: [%w]",
					err,
				)
			}

			if isPendingBTCWithdrawal {
				bwj.enqueueBTCWithdrawal(event)
				newPendingBTCWithdrawals++
			}
		}

		bwj.btcWithdrawalLastProcessedBlock = currentBlock
	}

	bwj.env.logger.Info(
		"search for new pending BTC withdrawals done",
		"new_pending_btc_withdrawals", newPendingBTCWithdrawals,
	)

	return nil
}

func (bwj *btcWithdrawalJob) newWalletRegisteredEvents(
	startHeight, endHeight uint64,
) ([]*tbtc.BridgeNewWalletRegistered, error) {
	bwj.env.logger.Info(
		"fetching NewWalletRegistered events",
		"start_height", startHeight,
		"end_height", endHeight,
	)
	return bwj.env.tbtcBridgeContract.PastNewWalletRegisteredEvents(startHeight, &endHeight, nil, nil)
}

func (bwj *btcWithdrawalJob) assetsUnlockConfirmedEvents(
	startHeight, endHeight uint64,
) ([]*portal.MezoBridgeAssetsUnlockConfirmed, error) {
	bwj.env.logger.Info(
		"fetching AssetsUnlockConfirmed events",
		"start_height", startHeight,
		"end_height", endHeight,
	)
	return bwj.env.mezoBridgeContract.PastAssetsUnlockConfirmedEvents(
		startHeight,
		&endHeight,
		nil,
		nil,
		[]common.Address{bwj.tbtcToken}, // fetch only events with tBTC as token
	)
}

// processBTCWithdrawalQueue processes pending BTC withdrawals. It removes
// AssetsUnlockConfirmed events from the queue, prepares data needed to perform
// a BTC withdrawal and submit the withdrawBTC transaction.
func (bwj *btcWithdrawalJob) processBTCWithdrawalQueue(ctx context.Context) {
	ticker := time.NewTicker(bwj.btcWithdrawalQueueCheckFrequency)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			for event := bwj.dequeueBTCWithdrawal(); event != nil; event = bwj.dequeueBTCWithdrawal() {
				btcWithdrawalLogger := bwj.env.logger.With(
					"unlock_sequence", event.UnlockSequenceNumber.String(),
				)

				// verify if still pending
				isPending, err := bwj.isPendingBTCWithdrawal(ctx, event)
				if err != nil {
					btcWithdrawalLogger.Error(
						"failed to check if BTC withdrawal is pending; re-queuing",
						"error", err,
					)
					bwj.enqueueBTCWithdrawal(event)
					continue
				}

				if !isPending {
					btcWithdrawalLogger.Info(
						"BTC withdrawal no longer pending; skipping",
					)
					continue
				}

				btcWithdrawalLogger.Info("starting BTC withdrawal submission")

				// Preparing BTC withdrawal data is time-consuming. Call it
				// only once per event processing.
				entry, walletPublicKeyHash, mainUTXO, err := bwj.prepareBTCWithdrawal(
					ctx,
					event,
				)
				if err != nil {
					btcWithdrawalLogger.Error(
						"failed to prepare BTC withdrawal; re-queuing",
						"error",
						err,
					)
					bwj.enqueueBTCWithdrawal(event)
					continue
				}

				btcWithdrawalLogger.Info(
					"found redemption wallet",
					"wallet_public_key_hash", hex.EncodeToString(walletPublicKeyHash[:]),
				)

				withdrawingSuccessful := false

				// Run a retry loop with a few attempts. If all the attempts
				// fail, quit retrying and put the event back into the queue.
				// We need to start over the event processing as the data
				// needed for `withdrawBTC` might have changed in the meantime.
				for i := 0; i < 5; i++ {
					btcWithdrawalProcessLogger := btcWithdrawalLogger.With("iteration", i)

					if i > 0 {
						// use a backoff for subsequent iterations as they are
						// most likely retries
						select {
						case <-time.After(btcWithdrawalProcessBackoff):
						case <-ctx.Done():
							btcWithdrawalProcessLogger.Warn(
								"stopping BTC withdrawal submission process " +
									"backoff wait due to context cancellation",
							)
							return
						}
					}

					// check if still pending
					ok, err := bwj.isPendingBTCWithdrawal(ctx, event)
					if err != nil {
						btcWithdrawalProcessLogger.Error(
							"failed to check if BTC withdrawal is pending; "+
								"retrying",
							"error", err,
						)
						continue
					}
					if !ok {
						btcWithdrawalProcessLogger.Info(
							"BTC withdrawal no longer pending; skipping",
						)
						withdrawingSuccessful = true
						break
					}

					btcWithdrawalProcessLogger.Info(
						"submitting BTC withdrawal transaction",
					)

					tx, err := bwj.env.mezoBridgeContract.WithdrawBTC(
						*entry,
						walletPublicKeyHash,
						*mainUTXO,
					)
					if err != nil {
						btcWithdrawalProcessLogger.Error(
							"BTC withdrawal transaction submission failed; "+
								"retrying",
							"error", err,
						)
						continue
					}

					// schedule finality check
					bwj.queueBTCWithdrawalFinalityCheck(event)

					withdrawingSuccessful = true

					btcWithdrawalProcessLogger.Info(
						"BTC withdrawal submitted",
						"tx_hash", tx.Hash().Hex(),
					)
					break
				}

				if !withdrawingSuccessful {
					btcWithdrawalLogger.Error(
						"all BTC withdrawal attempts failed; re-queuing",
					)
					bwj.enqueueBTCWithdrawal(event)
				}
			}
		case <-ctx.Done():
			bwj.env.logger.Warn(
				"stopping BTC withdrawal queue loop due to context " +
					"cancellation",
			)
			return
		}
	}
}

// queueBTCWithdrawalFinalityCheck constructs a BTC withdrawal finality check
// and puts it into a queue.
func (bwj *btcWithdrawalJob) queueBTCWithdrawalFinalityCheck(
	event *portal.MezoBridgeAssetsUnlockConfirmed,
) {
	bwj.btcWithdrawalFinalityChecksMutex.Lock()
	defer bwj.btcWithdrawalFinalityChecksMutex.Unlock()

	check := &btcWithdrawalFinalityCheck{
		event:             event,
		scheduledAtHeight: nil,
	}

	// This should never happen. Check just in case.
	if _, ok := bwj.btcWithdrawalFinalityChecks[check.key()]; ok {
		return
	}

	bwj.btcWithdrawalFinalityChecks[check.key()] = check

	bwj.env.logger.Info(
		"queued BTC withdrawal finality check",
		"unlock_sequence", event.UnlockSequenceNumber.String(),
	)
}

// processBTCWithdrawalFinalityChecks selects Bitcoin withdrawal finality checks
// that have reached their scheduled height and executes them.
func (bwj *btcWithdrawalJob) processBTCWithdrawalFinalityChecks(ctx context.Context) {
	tickerChan := bwj.env.chain.WatchBlocks(ctx)

	for {
		select {
		case height := <-tickerChan:
			currentFinalizedBlock, err := bwj.env.chain.FinalizedBlock(ctx)
			if err != nil {
				bwj.env.logger.Error(
					"cannot get finalized block during BTC withdrawal finality "+
						"checks - skipping current iteration",
					"error", err,
				)
				continue
			}

			checksToExecute := make([]*btcWithdrawalFinalityCheck, 0)

			// Lock the mutex but do not hold it for too long. Just schedule
			// new checks and determine which ones are ready to be executed.
			bwj.btcWithdrawalFinalityChecksMutex.Lock()
			for _, check := range bwj.btcWithdrawalFinalityChecks {
				checkLogger := bwj.env.logger.With(
					"unlock_sequence", check.event.UnlockSequenceNumber.String(),
				)

				// First, schedule the check if needed.
				if check.scheduledAtHeight == nil {
					//nolint:gosec
					check.scheduledAtHeight = big.NewInt(int64(height))

					checkLogger.Info(
						"BTC withdrawal finality check scheduled",
						"scheduled_at_height", check.scheduledAtHeight.String(),
						"current_finalized_block", currentFinalizedBlock.String(),
					)
				}

				// Then, see if check's scheduled height fell into a finalized epoch
				// and is ready to be executed.
				if currentFinalizedBlock.Cmp(check.scheduledAtHeight) >= 0 {
					checksToExecute = append(checksToExecute, check)
				}
			}
			bwj.btcWithdrawalFinalityChecksMutex.Unlock()

			for _, check := range checksToExecute {
				checkLogger := bwj.env.logger.With(
					"unlock_sequence", check.event.UnlockSequenceNumber.String(),
				)

				checkLogger.Info(
					"executing BTC withdrawal finality check",
					"scheduled_at_height", check.scheduledAtHeight.String(),
					"current_finalized_block", currentFinalizedBlock.String(),
				)

				pending, err := bwj.isPendingBTCWithdrawal(ctx, check.event)
				if err != nil {
					checkLogger.Error(
						"error while checking BTC withdrawal finality - retry will be "+
							"done during next iteration",
						"error", err,
					)
					// Continue to the next check without removing the current one
					// from the queue upon error. This way, the check will be retried
					// during the next iteration.
					continue
				}

				if !pending {
					checkLogger.Info(
						"BTC withdrawal confirmed during finality check",
					)
				} else {
					// If the withdrawal is still pending, we need to re-queue it
					// so the withdrawal loop can pick it up again.
					bwj.enqueueBTCWithdrawal(check.event)
					checkLogger.Info(
						"BTC withdrawal still pending during finality check; " +
							"re-queued",
					)
				}

				// Withdrawal check is done, remove the check from the queue.
				bwj.btcWithdrawalFinalityChecksMutex.Lock()
				delete(bwj.btcWithdrawalFinalityChecks, check.key())
				bwj.btcWithdrawalFinalityChecksMutex.Unlock()
			}
		case <-ctx.Done():
			bwj.env.logger.Warn(
				"stopping BTC withdrawal finality checks due to context " +
					"cancellation",
			)
			return
		}
	}
}

// prepareBTCWithdrawal selects the redeeming wallet and prepares arguments
// needed to execute withdrawBTC: AssetsUnlock entry, wallet public key hash and
// wallet main UTXO.
func (bwj *btcWithdrawalJob) prepareBTCWithdrawal(
	ctx context.Context,
	event *portal.MezoBridgeAssetsUnlockConfirmed,
) (
	*portal.MezoBridgeAssetsUnlocked,
	[20]byte, // wallet PKH
	*portal.BitcoinTxUTXO, // main UTXO
	error,
) {
	// We must get the recipient field from `mezod`. We cannot read it from
	// the passed `AssetsUnlockConfirmed` event as it only contains a hash of
	// recipient.
	unlockSequence := event.UnlockSequenceNumber

	assetsUnlockEvents, err := bwj.assetsUnlockedEndpoint.GetAssetsUnlockedEvents(
		ctx,
		sdkmath.NewIntFromBigInt(new(big.Int).Set(unlockSequence)),
		sdkmath.NewIntFromBigInt(new(big.Int).Add(
			new(big.Int).Set(unlockSequence),
			big.NewInt(1)),
		),
	)
	if err != nil {
		return nil, [20]byte{}, nil, fmt.Errorf(
			"failed to get AssetsUnlock event from Mezo: [%w]",
			err,
		)
	}

	if len(assetsUnlockEvents) != 1 {
		return nil, [20]byte{}, nil, fmt.Errorf(
			"incorrect number of AssetsUnlock events returned from Mezo",
		)
	}

	recipient := assetsUnlockEvents[0].Recipient

	// Build the AssetsUnlocked entry.
	assetsUnlocked := portal.MezoBridgeAssetsUnlocked{
		UnlockSequenceNumber: event.UnlockSequenceNumber,
		Recipient:            recipient,
		Token:                event.Token,
		Amount:               event.Amount,
		Chain:                event.Chain,
	}

	// Work on a copy of live wallets to avoid blocking for too long.
	// Live wallets are represented by `[20]byte` arrays, so `copy` will
	// deep-copy them.
	bwj.liveWalletsMutex.Lock()

	walletPublicKeyHashes := make([][20]byte, len(bwj.liveWallets))
	copy(walletPublicKeyHashes, bwj.liveWallets)

	bwj.liveWalletsMutex.Unlock()

	// Select a wallet that can cover the redemption.
	for _, walletPublicKeyHash := range walletPublicKeyHashes {
		wallet, err := bwj.env.tbtcBridgeContract.Wallets(walletPublicKeyHash)
		if err != nil {
			return nil, [20]byte{}, nil, fmt.Errorf(
				"failed to get wallet: [%w]",
				err,
			)
		}

		// Check if the state of the wallet is still live.
		if wallet.State != walletStateLive {
			continue
		}

		mainUTXO, err := bwj.determineWalletMainUtxo(walletPublicKeyHash)
		if err != nil {
			return nil, [20]byte{}, nil, fmt.Errorf(
				"failed to determine wallet main UTXO: [%w]",
				err,
			)
		}

		// Wallet has no main UTXO yet.
		if mainUTXO == nil {
			continue
		}

		walletBalanceBtcPrecision := big.NewInt(mainUTXO.Value)
		walletBalanceBtcPrecision.Sub(
			walletBalanceBtcPrecision,
			new(big.Int).SetUint64(wallet.PendingRedemptionsValue),
		)

		// This should never happen - check just in case.
		if walletBalanceBtcPrecision.Sign() < 0 {
			continue
		}

		walletBalanceErc20Precision := btcToErc20Amount(walletBalanceBtcPrecision)

		// Wallet does not have enough funds.
		if walletBalanceErc20Precision.Cmp(event.Amount) < 0 {
			continue
		}

		outputScript, err := bitcoin.NewScriptFromVarLenData(recipient)
		if err != nil {
			return nil, [20]byte{}, nil, fmt.Errorf(
				"failed to create recipient output script: [%w]",
				err,
			)
		}

		// Check if there are is no pending redemption request for the given
		// output script from the wallet.
		redemptionKey, err := computeRedemptionKey(walletPublicKeyHash, outputScript)
		if err != nil {
			return nil, [20]byte{}, nil, fmt.Errorf(
				"failed to compute redemption key: [%w]",
				err,
			)
		}

		redemptionRequest, err := bwj.env.tbtcBridgeContract.PendingRedemptions(redemptionKey)
		if err != nil {
			return nil, [20]byte{}, nil, fmt.Errorf(
				"failed to get pending redemption request: [%w]",
				err,
			)
		}

		// There is already a pending redemption.
		if redemptionRequest.RequestedAt != 0 {
			continue
		}

		utxo := portal.BitcoinTxUTXO{
			TxHash:        mainUTXO.Outpoint.TransactionHash,
			TxOutputIndex: mainUTXO.Outpoint.OutputIndex,
			TxOutputValue: uint64(mainUTXO.Value), // #nosec G115
		}

		return &assetsUnlocked, walletPublicKeyHash, &utxo, nil
	}

	return nil, [20]byte{}, nil, fmt.Errorf(
		"cannot find wallet to cover BTC withdrawal",
	)
}

// determineWalletMainUtxo determines the plain-text wallet main UTXO
// currently registered in the Bridge on-chain contract. The returned
// main UTXO can be nil if the wallet does not have a main UTXO registered
// in the Bridge at the moment.
func (bwj *btcWithdrawalJob) determineWalletMainUtxo(
	walletPublicKeyHash [20]byte,
) (*bitcoin.UnspentTransactionOutput, error) {
	walletChainData, err := bwj.env.tbtcBridgeContract.Wallets(walletPublicKeyHash)
	if err != nil {
		return nil, fmt.Errorf("cannot get on-chain data for wallet: [%v]", err)
	}

	// Valid case when the wallet doesn't have a main UTXO registered into
	// the Bridge.
	if walletChainData.MainUtxoHash == [32]byte{} {
		return nil, nil
	}

	// The wallet main UTXO registered in the Bridge almost always comes
	// from the latest BTC transaction made by the wallet. However, there may
	// be cases where the BTC transaction was made but their SPV proof is
	// not yet submitted to the Bridge thus the registered main UTXO points
	// to the second last BTC transaction. In theory, such a gap between
	// the actual latest BTC transaction and the registered main UTXO in
	// the Bridge may be even wider. To cover the worst possible cases, we
	// must rely on the full transaction history. Due to performance reasons,
	// we are first taking just the transactions hashes (fast call) and then
	// fetch full transaction data (time-consuming calls) starting from
	// the most recent transactions as there is a high chance the main UTXO
	// comes from there.
	txHashes, err := bwj.env.btcChain.GetTxHashesForPublicKeyHash(walletPublicKeyHash)
	if err != nil {
		return nil, fmt.Errorf("cannot get transactions history for wallet: [%v]", err)
	}

	walletP2PKH, err := bitcoin.PayToPublicKeyHash(walletPublicKeyHash)
	if err != nil {
		return nil, fmt.Errorf("cannot construct P2PKH for wallet: [%v]", err)
	}
	walletP2WPKH, err := bitcoin.PayToWitnessPublicKeyHash(walletPublicKeyHash)
	if err != nil {
		return nil, fmt.Errorf("cannot construct P2WPKH for wallet: [%v]", err)
	}

	// Start iterating from the latest transaction as the chance it matches
	// the wallet main UTXO is the highest.
	for i := len(txHashes) - 1; i >= 0; i-- {
		txHash := txHashes[i]

		transaction, err := bwj.env.btcChain.GetTransaction(txHash)
		if err != nil {
			return nil, fmt.Errorf(
				"cannot get transaction with hash [%s]: [%v]",
				txHash.String(),
				err,
			)
		}

		// Iterate over transaction's outputs and find the one that targets
		// the wallet public key hash.
		for outputIndex, output := range transaction.Outputs {
			script := output.PublicKeyScript
			matchesWallet := bytes.Equal(script, walletP2PKH) ||
				bytes.Equal(script, walletP2WPKH)

			// Once the right output is found, check whether their hash
			// matches the main UTXO hash stored on-chain. If so, this
			// UTXO is the one we are looking for.
			if matchesWallet {
				utxo := &bitcoin.UnspentTransactionOutput{
					Outpoint: &bitcoin.TransactionOutpoint{
						TransactionHash: transaction.Hash(),
						OutputIndex:     uint32(outputIndex), // #nosec G115
					},
					Value: output.Value,
				}

				if computeMainUtxoHash(utxo) ==
					walletChainData.MainUtxoHash {
					return utxo, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("main UTXO not found")
}

// btcToErc20Amount converts amount in Bitcoin precision (1e8) to ERC20 token
// precision (1e18). It effectively multiplies the Bitcoin precision amount by
// 1e10.
func btcToErc20Amount(btcPrecisionAmount *big.Int) *big.Int {
	return new(big.Int).Mul(
		new(big.Int).Set(btcPrecisionAmount),
		big.NewInt(10_000_000_000), // × 1e10
	)
}

// computeMainUtxoHash computes the hash of the provided main UTXO
// according to the on-chain Bridge rules.
func computeMainUtxoHash(mainUtxo *bitcoin.UnspentTransactionOutput) [32]byte {
	outputIndexBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(outputIndexBytes, mainUtxo.Outpoint.OutputIndex)

	valueBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(valueBytes, uint64(mainUtxo.Value)) // #nosec G115

	mainUtxoHash := crypto.Keccak256Hash(
		append(
			append(
				mainUtxo.Outpoint.TransactionHash[:],
				outputIndexBytes...,
			), valueBytes...,
		),
	)

	return mainUtxoHash
}

// computeRedemptionKey calculates a redemption key for the given redemption
// request which is an identifier for a redemption at the given time
// on-chain.
func computeRedemptionKey(
	walletPublicKeyHash [20]byte,
	redeemerOutputScript bitcoin.Script,
) (*big.Int, error) {
	// The Bridge contract builds the redemption key using the length-prefixed
	// redeemer output script.
	prefixedRedeemerOutputScript, err := redeemerOutputScript.ToVarLenData()
	if err != nil {
		return nil, fmt.Errorf("cannot build prefixed redeemer output script: [%v]", err)
	}

	redeemerOutputScriptHash := crypto.Keccak256Hash(prefixedRedeemerOutputScript)

	redemptionKey := crypto.Keccak256Hash(
		append(redeemerOutputScriptHash[:], walletPublicKeyHash[:]...),
	)

	return redemptionKey.Big(), nil
}

// computeAttestationKey computes the attestation key for the data representing
// an AssetsUnlocked entry.
func computeAttestationKey(
	unlockSeq *big.Int,
	recipient []byte,
	token common.Address,
	amount *big.Int,
	chain uint8,
) (common.Hash, error) {
	type assetsUnlockedTuple struct {
		UnlockSequenceNumber *big.Int       `abi:"unlockSequenceNumber"`
		Recipient            []byte         `abi:"recipient"`
		Token                common.Address `abi:"token"`
		Amount               *big.Int       `abi:"amount"`
		Chain                uint8          `abi:"chain"`
	}

	tupleType, err := abi.NewType("tuple", "tuple", []abi.ArgumentMarshaling{
		{Name: "unlockSequenceNumber", Type: "uint256"}, // unlockSequenceNumber
		{Name: "recipient", Type: "bytes"},              // recipient
		{Name: "token", Type: "address"},                // token
		{Name: "amount", Type: "uint256"},               // amount
		{Name: "chain", Type: "uint8"},                  // chain
	})
	if err != nil {
		return common.Hash{}, err
	}

	entry := assetsUnlockedTuple{
		UnlockSequenceNumber: unlockSeq,
		Recipient:            recipient,
		Token:                token,
		Amount:               amount,
		Chain:                chain,
	}

	bytes, err := (abi.Arguments{{Type: tupleType}}).Pack(entry)
	if err != nil {
		return common.Hash{}, err
	}

	return crypto.Keccak256Hash(bytes), nil
}
