package bridgeworker

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"math/big"
	"sort"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/mezo-org/mezod/bridge-worker/bitcoin"
	"github.com/mezo-org/mezod/ethereum/bindings/portal"
	"github.com/mezo-org/mezod/ethereum/bindings/tbtc"
)

const (
	// bitcoinTargetChain is a numerical value representing Bitcoin target chain.
	bitcoinTargetChain = uint8(1)

	// assetsUnlockConfirmedLookBackBlocks is the number of blocks used when
	// fetching AssetsUnlockConfirmed events from the Ethereum chain. It defines
	// how far back we look when searching for events.
	assetsUnlockConfirmedLookBackBlocks = 216000 // ~30 days

	// withdrawalProcessBackoff is a backoff time used between retries when
	// submitting a withdrawBTC transaction.
	withdrawalProcessBackoff = 1 * time.Minute

	// liveWalletsUpdatePeriod defines how frequently we should update the list
	// of live wallets.
	liveWalletsUpdatePeriod = 1 * time.Hour

	// walletStateLive represents Live wallet state from the tBTC Bridge contract.
	walletStateLive = uint8(1)
)

func (bw *BridgeWorker) observeLiveWallets(ctx context.Context) error {
	finalizedBlock, err := bw.chain.FinalizedBlock(ctx)
	if err != nil {
		return fmt.Errorf("failed to get finalized block: [%w]", err)
	}
	endBlock := finalizedBlock.Uint64()

	// TODO: Only look at 1 month of blocks; allow to provide a list of wallets
	//       in configuration. Check the wallets on the provided list.
	recentEvents, err := bw.fetchNewWalletRegisteredEvents(
		0,
		endBlock,
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

		wallet, err := bw.tbtcBridgeContract.Wallets(walletPublicKeyHash)
		if err != nil {
			return fmt.Errorf("failed to get wallet: [%w]", err)
		}

		if wallet.State == walletStateLive {
			liveWallets = append(liveWallets, walletPublicKeyHash)
		}
	}

	bw.liveWalletsLastProcessedBlock = endBlock

	bw.liveWalletsMutex.Lock()
	bw.liveWallets = liveWallets
	bw.liveWalletsMutex.Unlock()

	// TODO: Should we block the Bitcoin withdrawal routine
	//       until the initial search for live wallets is complete?

	bw.logger.Info(
		"finished initial search for live wallets",
		"number_of_live_wallets", len(liveWallets),
	)

	ticker := time.NewTicker(liveWalletsUpdatePeriod)
	defer ticker.Stop()
	tickerChan := ticker.C

	for {
		select {
		case <-ctx.Done(): // Handle context cancellation
			bw.logger.Warn(
				"stopping live wallet observation due to context cancellation",
			)
			return nil
		case <-tickerChan:
			err := bw.updateLiveWallets(ctx)
			if err != nil {
				bw.logger.Error("failed to update live wallets", "err", err)
			}
		}
	}
}

func (bw *BridgeWorker) fetchNewWalletRegisteredEvents(
	startBlock uint64,
	endBlock uint64,
) ([]*tbtc.BridgeNewWalletRegistered, error) {
	bw.logger.Info(
		"fetching NewWalletRegistered events from range",
		"start_block", startBlock,
		"end_block", endBlock,
	)

	result := make([]*tbtc.BridgeNewWalletRegistered, 0)

	ticker := time.NewTicker(time.Minute / time.Duration(bw.requestsPerMinute)) //nolint:gosec
	defer ticker.Stop()

	events, err := bw.tbtcBridgeContract.PastNewWalletRegisteredEvents(
		startBlock,
		&endBlock,
		nil,
		nil,
	)
	if err != nil {
		bw.logger.Warn(
			"failed to fetch NewWalletRegistered events from the entire "+
				"range; falling back to batched events fetch",
			"start_block", startBlock,
			"end_block", endBlock,
			"err", err,
		)

		batchStartBlock := startBlock

		for batchStartBlock <= endBlock {
			batchEndBlock := batchStartBlock + bw.batchSize
			if batchEndBlock > endBlock {
				batchEndBlock = endBlock
			}

			bw.logger.Info(
				"fetching a batch of NewWalletRegistered events from range",
				"batch_start_block", batchStartBlock,
				"batch_end_block", batchEndBlock,
			)

			<-ticker.C

			batchEvents, batchErr := bw.tbtcBridgeContract.PastNewWalletRegisteredEvents(
				batchStartBlock,
				&batchEndBlock,
				nil,
				nil,
			)
			if batchErr != nil {
				return nil, fmt.Errorf(
					"batched NewWalletRegistered fetch failed: [%w]; giving up",
					batchErr,
				)
			}

			result = append(result, batchEvents...)

			batchStartBlock = batchEndBlock + 1
		}
	} else {
		result = append(result, events...)
	}

	return result, nil
}

func (bw *BridgeWorker) updateLiveWallets(ctx context.Context) error {
	// Keep only wallets which are still live.
	updatedLiveWallets := [][20]byte{}

	for _, walletPublicKeyHash := range bw.liveWallets {
		wallet, err := bw.tbtcBridgeContract.Wallets(walletPublicKeyHash)
		if err != nil {
			return fmt.Errorf("failed to get wallet: [%w]", err)
		}

		if wallet.State == walletStateLive {
			updatedLiveWallets = append(updatedLiveWallets, walletPublicKeyHash)
		}
	}

	// Look for new live wallets.
	finalizedBlock, err := bw.chain.FinalizedBlock(ctx)
	if err != nil {
		return fmt.Errorf("failed to get finalized block: [%w]", err)
	}
	endBlock := finalizedBlock.Uint64()

	if endBlock > bw.liveWalletsLastProcessedBlock {
		events, err := bw.fetchNewWalletRegisteredEvents(
			bw.liveWalletsLastProcessedBlock+1,
			endBlock,
		)
		if err != nil {
			return fmt.Errorf(
				"failed to fetch NewWalletRegistered events: [%w]",
				err,
			)
		}

		for _, event := range events {
			walletPublicKeyHash := event.WalletPubKeyHash

			wallet, err := bw.tbtcBridgeContract.Wallets(walletPublicKeyHash)
			if err != nil {
				return fmt.Errorf("failed to get wallet: [%w]", err)
			}

			if wallet.State == walletStateLive {
				updatedLiveWallets = append(updatedLiveWallets, walletPublicKeyHash)
			}
		}

		bw.liveWalletsLastProcessedBlock = endBlock
	}

	bw.liveWalletsMutex.Lock()
	bw.liveWallets = updatedLiveWallets
	bw.liveWalletsMutex.Unlock()

	bw.logger.Info(
		"finished updating live wallets",
		"number_of_live_wallets", len(bw.liveWallets),
	)

	return nil
}

// withdrawalFinalityCheck is a struct that contains an AssetsUnlockConfirmed
// and the Ethereum height at which the withdrawal finality check for it was
// scheduled.
type withdrawalFinalityCheck struct {
	event             *portal.MezoBridgeAssetsUnlockConfirmed
	scheduledAtHeight *big.Int // nil means unscheduled
}

func (wfc *withdrawalFinalityCheck) key() string {
	return wfc.event.UnlockSequenceNumber.String()
}

// observeBitcoinWithdrawals monitors AssetsUnlockConfirmed events, filters
// events representing pending Bitcoin withdrawals and puts them into a queue.
func (bw *BridgeWorker) observeBitcoinWithdrawals(ctx context.Context) error {
	// Use the current block rather than finalized block to speed up event
	// processing. Event processing should handle the effects of a possible
	// reorg (e.g. an event being duplicated in a queue), although there is a
	// small risk of skipping an event. In that case it would require a manual
	// execution of `withdrawBTC`.
	currentBlock, err := bw.chain.BlockCounter().CurrentBlock()
	if err != nil {
		return fmt.Errorf("failed to get current block: [%w]", err)
	}

	startBlock := uint64(0)
	if currentBlock > assetsUnlockConfirmedLookBackBlocks {
		startBlock = currentBlock - assetsUnlockConfirmedLookBackBlocks
	}

	recentEvents, err := bw.fetchAssetsUnlockConfirmedEvents(
		startBlock,
		currentBlock,
	)
	if err != nil {
		return fmt.Errorf(
			"failed to fetch AssetsUnlockConfirmed events: [%w]",
			err,
		)
	}

	pendingWithdrawalCount := 0

	for _, event := range recentEvents {
		isPendingWithdrawal, err := bw.isPendingBTCWithdrawal(event)
		if err != nil {
			return fmt.Errorf(
				"failed to check if event represents pending BTC withdrawal: [%w]",
				err,
			)
		}

		if isPendingWithdrawal {
			bw.enqueueBtcWithdrawal(event)
			pendingWithdrawalCount++
		}
	}

	bw.btcWithdrawalLastProcessedBlock = currentBlock

	bw.logger.Info(
		"initial search for pending BTC withdrawals done",
		"confirmed_withdrawal_events", len(recentEvents),
		"pending_btc_withdrawals", pendingWithdrawalCount,
	)

	// Start a ticker to periodically check the current block number
	tickerChan := bw.chain.WatchBlocks(ctx)

	for {
		select {
		case <-ctx.Done(): // Handle context cancellation
			bw.logger.Warn(
				"stopping BTC withdrawals routine due to context cancellation",
			)
			return nil
		case <-tickerChan:
			// Process incoming AssetsUnlockedConfirmed events
			err := bw.processNewAssetsUnlockConfirmedEvents()
			if err != nil {
				bw.logger.Error(
					"failed to process newly emitted AssetsUnlockConfirmed events",
					"err", err,
				)
			}
		}
	}
}

// enqueueBtcWithdrawal puts AssetsUnlockConfirmed events representing Bitcoin
// withdrawals into a queue.
func (bw *BridgeWorker) enqueueBtcWithdrawal(
	event *portal.MezoBridgeAssetsUnlockConfirmed,
) {
	bw.btcWithdrawalMutex.Lock()
	defer bw.btcWithdrawalMutex.Unlock()

	bw.btcWithdrawalQueue = append(bw.btcWithdrawalQueue, *event)

	// order events by unlock sequence in ascending order
	sort.Slice(bw.btcWithdrawalQueue, func(i, j int) bool {
		return bw.btcWithdrawalQueue[i].UnlockSequenceNumber.Cmp(
			bw.btcWithdrawalQueue[j].UnlockSequenceNumber,
		) < 0
	})
}

// dequeueBtcWithdrawal removes an AssetsUnlockConfirmed event representing
// a Bitcoin withdrawals from the queue.
func (bw *BridgeWorker) dequeueBtcWithdrawal() *portal.MezoBridgeAssetsUnlockConfirmed {
	bw.btcWithdrawalMutex.Lock()
	defer bw.btcWithdrawalMutex.Unlock()
	if len(bw.btcWithdrawalQueue) == 0 {
		return nil
	}

	event := bw.btcWithdrawalQueue[0]
	bw.btcWithdrawalQueue = bw.btcWithdrawalQueue[1:]

	return &event
}

// fetchAssetsUnlockConfirmedEvents fetches AssetsUnlockConfirmed events from
// the Ethereum MezoBridge contract from the given range.
func (bw *BridgeWorker) fetchAssetsUnlockConfirmedEvents(
	startBlock uint64,
	endBlock uint64,
) ([]*portal.MezoBridgeAssetsUnlockConfirmed, error) {
	bw.logger.Info(
		"fetching AssetsUnlockConfirmed events from range",
		"start_block", startBlock,
		"end_block", endBlock,
	)

	result := make([]*portal.MezoBridgeAssetsUnlockConfirmed, 0)

	ticker := time.NewTicker(time.Minute / time.Duration(bw.requestsPerMinute)) //nolint:gosec
	defer ticker.Stop()

	events, err := bw.mezoBridgeContract.PastAssetsUnlockConfirmedEvents(
		startBlock,
		&endBlock,
		nil,
		nil,
		nil,
	)
	if err != nil {
		bw.logger.Warn(
			"failed to fetch AssetsUnlockConfirmed events from the entire "+
				"range; falling back to batched events fetch",
			"start_block", startBlock,
			"end_block", endBlock,
			"err", err,
		)

		batchStartBlock := startBlock

		for batchStartBlock <= endBlock {
			batchEndBlock := batchStartBlock + bw.batchSize
			if batchEndBlock > endBlock {
				batchEndBlock = endBlock
			}

			bw.logger.Info(
				"fetching a batch of AssetsUnlockConfirmed events from range",
				"batch_start_block", batchStartBlock,
				"batch_end_block", batchEndBlock,
			)

			<-ticker.C

			batchEvents, batchErr := bw.mezoBridgeContract.PastAssetsUnlockConfirmedEvents(
				batchStartBlock,
				&batchEndBlock,
				nil,
				nil,
				nil,
			)
			if batchErr != nil {
				return nil, fmt.Errorf(
					"batched AssetsUnlockConfirmed fetch failed: [%w]; giving up",
					batchErr,
				)
			}

			result = append(result, batchEvents...)

			batchStartBlock = batchEndBlock + 1
		}
	} else {
		result = append(result, events...)
	}

	return result, nil
}

// isPendingBTCWithdrawal checks whether an AssetsUnlockConfirmed event
// represents a pending Bitcoin withdrawal.
func (bw *BridgeWorker) isPendingBTCWithdrawal(
	event *portal.MezoBridgeAssetsUnlockConfirmed,
) (bool, error) {
	if event.Chain != bitcoinTargetChain {
		return false, nil
	}

	hash, err := computeAttestationKey(
		event.UnlockSequenceNumber,
		event.Recipient[:],
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

	isPendingBTCWithdrawal, err := bw.mezoBridgeContract.PendingBTCWithdrawals(hash)
	if err != nil {
		return false, fmt.Errorf(
			"failed to get pending BTC withdrawals info: [%w]",
			err,
		)
	}

	return isPendingBTCWithdrawal, nil
}

// processNewAssetsUnlockConfirmedEvents fetches new AssetsUnlockConfirmed
// events representing pending Bitcoin withdrawals and puts them into a queue.
// It is intended to be run periodically.
func (bw *BridgeWorker) processNewAssetsUnlockConfirmedEvents() error {
	// Use the current block rather than finalized block to speed up event
	// processing. Event processing should handle the effects of a possible
	// reorg (e.g. an event being duplicated in a queue), although there is a
	// small risk of skipping an event. In that case it would require a manual
	// execution of `withdrawBTC`.
	currentBlock, err := bw.chain.BlockCounter().CurrentBlock()
	if err != nil {
		return fmt.Errorf("cannot get current block: [%w]", err)
	}

	newPendingBtcWithdrawals := 0

	if currentBlock > bw.btcWithdrawalLastProcessedBlock {
		events, err := bw.fetchAssetsUnlockConfirmedEvents(
			bw.btcWithdrawalLastProcessedBlock+1,
			currentBlock,
		)
		if err != nil {
			return fmt.Errorf(
				"failed to fetch AssetsUnlockConfirmed events: [%w]",
				err,
			)
		}

		for _, event := range events {
			isPendingWithdrawal, err := bw.isPendingBTCWithdrawal(event)
			if err != nil {
				return fmt.Errorf(
					"failed to check if event represents pending BTC "+
						"withdrawal: [%w]",
					err,
				)
			}

			if isPendingWithdrawal {
				bw.enqueueBtcWithdrawal(event)
				newPendingBtcWithdrawals++
			}
		}

		bw.btcWithdrawalLastProcessedBlock = currentBlock
	}

	bw.logger.Info(
		"search for new pending BTC withdrawals done",
		"new_pending_btc_withdrawals", newPendingBtcWithdrawals,
	)

	return nil
}

// processBtcWithdrawalQueue processes pending Bitcoin withdrawals. It removes
// AssetsUnlockConfirmed events from the queue, prepares data needed to perform
// a withdrawal and submit the withdrawBTC transaction.
func (bw *BridgeWorker) processBtcWithdrawalQueue(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			for event := bw.dequeueBtcWithdrawal(); event != nil; event = bw.dequeueBtcWithdrawal() {
				withdrawalLogger := bw.logger.With(
					"unlock_sequence", event.UnlockSequenceNumber.String(),
				)

				// verify if still pending
				isPending, err := bw.isPendingBTCWithdrawal(event)
				if err != nil {
					withdrawalLogger.Error(
						"failed to check if withdrawal is pending; re-queuing",
						"error", err,
					)
					bw.enqueueBtcWithdrawal(event)
					continue
				}

				if !isPending {
					withdrawalLogger.Info("withdrawal no longer pending; skipping")
					continue
				}

				withdrawalLogger.Info("starting Bitcoin withdrawal submission")

				// Preparing Bitcoin withdrawal data is time-consuming. Call it
				// only once per event processing.
				entry, walletPublicKeyHash, mainUTXO, err := bw.prepareBtcWithdrawal(event)
				if err != nil {
					withdrawalLogger.Error(
						"failed to prepare withdrawal; re-queuing",
						"error",
						err,
					)
					bw.enqueueBtcWithdrawal(event)
					continue
				}

				withdrawingSuccessful := false

				// Run a retry loop with a few attempts. If all the attempts
				// fail, quit retrying and put the event back into the queue.
				// We need to start over the event processing as the data
				// needed for `withdrawBTC` might have changed in the meantime.
				for i := 0; i < 5; i++ {
					withdrawalProcessLogger := withdrawalLogger.With("iteration", i)

					if i > 0 {
						// use a backoff for subsequent iterations as they are
						// most likely retries
						select {
						case <-time.After(withdrawalProcessBackoff):
						case <-ctx.Done():
							withdrawalProcessLogger.Warn(
								"stopping withdrawal submission process backoff " +
									"wait due to context cancellation",
							)
							return
						}
					}

					// check if still pending
					ok, err := bw.isPendingBTCWithdrawal(event)
					if err != nil {
						withdrawalProcessLogger.Error(
							"failed to check if pending; retrying",
							"error", err,
						)
						continue
					}
					if !ok {
						withdrawalProcessLogger.Info(
							"withdrawal no longer pending; skipping",
						)
						withdrawingSuccessful = true
						break
					}

					withdrawalProcessLogger.Info(
						"submitting Bitcoin withdrawal transaction",
					)

					tx, err := bw.mezoBridgeContract.WithdrawBTC(
						entry,
						walletPublicKeyHash,
						mainUTXO,
					)
					if err != nil {
						withdrawalProcessLogger.Error(
							"withdrawal transaction submission failed; retrying",
							"error", err,
						)
						continue
					}

					// schedule finality check
					bw.queueWithdrawalFinalityCheck(event)

					withdrawingSuccessful = true

					withdrawalProcessLogger.Info(
						"Bitcoin withdrawal submitted",
						"tx_hash", tx.Hash().Hex(),
					)
					break
				}

				if !withdrawingSuccessful {
					withdrawalLogger.Error(
						"all withdrawal attempts failed; re-queuing",
					)
					bw.enqueueBtcWithdrawal(event)
				}
			}
		case <-ctx.Done():
			bw.logger.Warn(
				"stopping Bitcoin withdrawal queue loop due to context " +
					"cancellation",
			)
			return
		}
	}
}

// queueWithdrawalFinalityCheck constructs a Bitcoin withdrawal finality check
// and puts it into a queue.
func (bw *BridgeWorker) queueWithdrawalFinalityCheck(
	event *portal.MezoBridgeAssetsUnlockConfirmed,
) {
	bw.withdrawalFinalityChecksMutex.Lock()
	defer bw.withdrawalFinalityChecksMutex.Unlock()

	check := &withdrawalFinalityCheck{
		event:             event,
		scheduledAtHeight: nil,
	}

	// This should never happen. Check just in case.
	if _, ok := bw.withdrawalFinalityChecks[check.key()]; ok {
		return
	}

	bw.withdrawalFinalityChecks[check.key()] = check

	bw.logger.Info(
		"queued Bitcoin withdrawal finality check",
		"unlock_sequence", event.UnlockSequenceNumber.String(),
	)
}

// processWithdrawalFinalityChecks selects Bitcoin withdrawal finality checks
// that have reached their scheduled height and executes them.
func (bw *BridgeWorker) processWithdrawalFinalityChecks(ctx context.Context) {
	tickerChan := bw.chain.WatchBlocks(ctx)

	for {
		select {
		case height := <-tickerChan:
			currentFinalizedBlock, err := bw.chain.FinalizedBlock(ctx)
			if err != nil {
				bw.logger.Error(
					"cannot get finalized block during withdrawal finality checks - "+
						"skipping current iteration",
					"error", err,
				)
				continue
			}

			checksToExecute := make([]*withdrawalFinalityCheck, 0)

			// Lock the mutex but do not hold it for too long. Just schedule
			// new checks and determine which ones are ready to be executed.
			bw.withdrawalFinalityChecksMutex.Lock()
			for _, check := range bw.withdrawalFinalityChecks {
				checkLogger := bw.logger.With(
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
			bw.withdrawalFinalityChecksMutex.Unlock()

			for _, check := range checksToExecute {
				checkLogger := bw.logger.With(
					"unlock_sequence", check.event.UnlockSequenceNumber.String(),
				)

				checkLogger.Info(
					"executing Bitcoin withdrawal finality check",
					"scheduled_at_height", check.scheduledAtHeight.String(),
					"current_finalized_block", currentFinalizedBlock.String(),
				)

				pending, err := bw.isPendingBTCWithdrawal(check.event)
				if err != nil {
					checkLogger.Error(
						"error while checking Bitcoin withdrawal finality - retry will be "+
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
						"withdrawal confirmed during finality check",
					)
				} else {
					// If the withdrawal is still pending, we need to re-queue it
					// so the withdrawal loop can pick it up again.
					bw.enqueueBtcWithdrawal(check.event)
					checkLogger.Info(
						"withdrawal still pending during finality check; " +
							"re-queued",
					)
				}

				// Withdrawal check is done, remove the check from the queue.
				bw.withdrawalFinalityChecksMutex.Lock()
				delete(bw.withdrawalFinalityChecks, check.key())
				bw.withdrawalFinalityChecksMutex.Unlock()
			}
		case <-ctx.Done():
			bw.logger.Warn(
				"stopping Bitcoin withdrawal finality checks due to context " +
					"cancellation",
			)
			return
		}
	}
}

func (bw *BridgeWorker) prepareBtcWithdrawal(
	event *portal.MezoBridgeAssetsUnlockConfirmed,
) (
	portal.MezoBridgeAssetsUnlocked,
	[20]byte, // wallet pubkey hash to return
	portal.BitcoinTxUTXO, // main UTXO to return
	error,
) {
	assetsUnlocked := portal.MezoBridgeAssetsUnlocked{
		UnlockSequenceNumber: event.UnlockSequenceNumber,
		Recipient:            event.Recipient[:], // TODO: We must get the Recipient from the `mezod` endpoint via gRPC call.
		Token:                event.Token,
		Amount:               event.Amount,
		Chain:                event.Chain,
	}

	// Work on a copy of live wallets to avoid blocking for too long.
	// Live wallets are represented by `[20]byte` arrays, so `copy` will
	// deep-copy them.

	bw.liveWalletsMutex.Lock()
	wallets := make([][20]byte, len(bw.liveWallets))
	copy(wallets, bw.liveWallets)
	bw.liveWalletsMutex.Unlock()

	fmt.Println(len(wallets))

	// TODO: Continue with the implementation.
	var walletPKH [20]byte
	var mainUTXO portal.BitcoinTxUTXO // TODO: When checking wallet amount take decimal precision into consideration (1e18 in the event vs 1e8 in mainUTXO)
	return assetsUnlocked, walletPKH, mainUTXO, nil
}

func (bw *BridgeWorker) determineWalletMainUtxo(
	walletPublicKeyHash [20]byte,
) (*bitcoin.UnspentTransactionOutput, error) {
	walletChainData, err := bw.tbtcBridgeContract.Wallets(walletPublicKeyHash)
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
	txHashes, err := bw.btcChain.GetTxHashesForPublicKeyHash(walletPublicKeyHash)
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

		transaction, err := bw.btcChain.GetTransaction(txHash)
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
						OutputIndex:     uint32(outputIndex),
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

func computeMainUtxoHash(mainUtxo *bitcoin.UnspentTransactionOutput) [32]byte {
	outputIndexBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(outputIndexBytes, mainUtxo.Outpoint.OutputIndex)

	valueBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(valueBytes, uint64(mainUtxo.Value))

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
