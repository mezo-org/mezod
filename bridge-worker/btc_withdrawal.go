package bridgeworker

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	sdkmath "cosmossdk.io/math"

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

	bw.logger.Info(
		"finished initial search for live wallets",
		"number_of_live_wallets", len(liveWallets),
	)

	bw.liveWalletsMutex.Lock()
	bw.liveWallets = liveWallets
	bw.liveWalletsMutex.Unlock()

	// Signal that initial fetching of wallets ready.
	close(bw.liveWalletsReady)

	// Start a ticker to periodically update the wallets.
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
	// Work on a copy of live wallets to avoid blocking for too long.
	// Live wallets are represented by `[20]byte` arrays, so `copy` will
	// deep-copy them.
	bw.liveWalletsMutex.Lock()

	walletPublicKeyHashes := make([][20]byte, len(bw.liveWallets))
	copy(walletPublicKeyHashes, bw.liveWallets)

	bw.liveWalletsMutex.Unlock()

	// Keep only wallets which are still live.
	updatedLiveWallets := [][20]byte{}

	for _, walletPublicKeyHash := range walletPublicKeyHashes {
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

	bw.logger.Info(
		"finished updating live wallets",
		"number_of_live_wallets", len(updatedLiveWallets),
	)

	bw.liveWalletsMutex.Lock()
	bw.liveWallets = updatedLiveWallets
	bw.liveWalletsMutex.Unlock()

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
		isPendingWithdrawal, err := bw.isPendingBTCWithdrawal(ctx, event)
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
			err := bw.processNewAssetsUnlockConfirmedEvents(ctx)
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
// withdrawals at the end of the queue.
func (bw *BridgeWorker) enqueueBtcWithdrawal(
	event *portal.MezoBridgeAssetsUnlockConfirmed,
) {
	bw.btcWithdrawalMutex.Lock()
	defer bw.btcWithdrawalMutex.Unlock()

	bw.btcWithdrawalQueue = append(bw.btcWithdrawalQueue, *event)

	bw.logger.Debug(
		"enqueued BTC withdrawal",
		"unlock_sequence", event.UnlockSequenceNumber.String(),
	)
}

// dequeueBtcWithdrawal removes an AssetsUnlockConfirmed event representing
// a Bitcoin withdrawal from the front of the queue.
func (bw *BridgeWorker) dequeueBtcWithdrawal() *portal.MezoBridgeAssetsUnlockConfirmed {
	bw.btcWithdrawalMutex.Lock()
	defer bw.btcWithdrawalMutex.Unlock()
	if len(bw.btcWithdrawalQueue) == 0 {
		return nil
	}

	event := bw.btcWithdrawalQueue[0]
	bw.btcWithdrawalQueue = bw.btcWithdrawalQueue[1:]

	bw.logger.Debug(
		"dequeued BTC withdrawal",
		"unlock_sequence", event.UnlockSequenceNumber.String(),
	)

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

	assetsUnlockEvents, err := bw.assetsUnlockedEndpoint.GetAssetsUnlockedEvents(
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

	isPendingBTCWithdrawal, err := bw.mezoBridgeContract.PendingBTCWithdrawals(hash)
	if err != nil {
		return false, fmt.Errorf(
			"failed to get pending BTC withdrawals info: [%w]",
			err,
		)
	}

	if !isPendingBTCWithdrawal {
		return false, nil
	}

	redemptionParameters, err := bw.tbtcBridgeContract.RedemptionParameters()
	if err != nil {
		return false, fmt.Errorf(
			"failed to get redemption parameters: [%w]",
			err,
		)
	}

	redemptionDustThresholdBtcPrecision := new(big.Int).SetUint64(
		redemptionParameters.RedemptionDustThreshold,
	)
	redemptionDustThresholdErc20Precision := btcToErc20Amount(
		redemptionDustThresholdBtcPrecision,
	)

	if event.Amount.Cmp(redemptionDustThresholdErc20Precision) < 0 {
		bw.logger.Warn(
			"found BTC withdrawal below redemption dust threshold",
			"unlock_sequence", event.UnlockSequenceNumber.String(),
			"amount", event.Amount,
			"redemption_dust_threshold_erc20_precision", redemptionDustThresholdErc20Precision,
		)

		return false, nil
	}

	return true, nil
}

// processNewAssetsUnlockConfirmedEvents fetches new AssetsUnlockConfirmed
// events representing pending Bitcoin withdrawals and puts them into a queue.
// It is intended to be run periodically.
func (bw *BridgeWorker) processNewAssetsUnlockConfirmedEvents(ctx context.Context) error {
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
			isPendingWithdrawal, err := bw.isPendingBTCWithdrawal(ctx, event)
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
				isPending, err := bw.isPendingBTCWithdrawal(ctx, event)
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
				entry, walletPublicKeyHash, mainUTXO, err := bw.prepareBtcWithdrawal(
					ctx,
					event,
				)
				if err != nil {
					withdrawalLogger.Error(
						"failed to prepare withdrawal; re-queuing",
						"error",
						err,
					)
					bw.enqueueBtcWithdrawal(event)
					continue
				}

				withdrawalLogger.Info(
					"found redemption wallet",
					"wallet_public_key_hash", hex.EncodeToString(walletPublicKeyHash[:]),
				)

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
					ok, err := bw.isPendingBTCWithdrawal(ctx, event)
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
						*entry,
						walletPublicKeyHash,
						*mainUTXO,
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

				pending, err := bw.isPendingBTCWithdrawal(ctx, check.event)
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

// prepareBtcWithdrawal selects the redeeming wallet and prepares arguments
// needed to execute withdrawBTC: AssetsUnlock entry, wallet public key hash and
// wallet main UTXO.
func (bw *BridgeWorker) prepareBtcWithdrawal(
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

	assetsUnlockEvents, err := bw.assetsUnlockedEndpoint.GetAssetsUnlockedEvents(
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
	bw.liveWalletsMutex.Lock()

	walletPublicKeyHashes := make([][20]byte, len(bw.liveWallets))
	copy(walletPublicKeyHashes, bw.liveWallets)

	bw.liveWalletsMutex.Unlock()

	// Select a wallet that can cover the redemption.
	for _, walletPublicKeyHash := range walletPublicKeyHashes {
		wallet, err := bw.tbtcBridgeContract.Wallets(walletPublicKeyHash)
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

		mainUTXO, err := bw.determineWalletMainUtxo(walletPublicKeyHash)
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

		redemptionRequest, err := bw.tbtcBridgeContract.PendingRedemptions(redemptionKey)
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

	return nil, [20]byte{}, nil, fmt.Errorf("cannot find wallet to cover withdrawal")
}

// determineWalletMainUtxo determines the plain-text wallet main UTXO
// currently registered in the Bridge on-chain contract. The returned
// main UTXO can be nil if the wallet does not have a main UTXO registered
// in the Bridge at the moment.
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
