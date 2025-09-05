package bridgeworker

import (
	"context"
	"fmt"
	"math/big"
	"sort"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/mezo-org/mezod/ethereum/bindings/portal"
)

const (
	// bitcoinTargetChain is a numerical value representing Bitcoin target chain.
	bitcoinTargetChain = uint8(1)

	// TODO: Should the following be options?
	defaultBatchSize                    = uint64(1000)
	defaultRequestsPerMinute            = uint64(600) // 10 requests per second
	assetsUnlockConfirmedLookBackBlocks = 216000      // ~30 days

	withdrawalProcessBackoff = 10 * time.Second
)

type withdrawalFinalityCheck struct {
	event             *portal.MezoBridgeAssetsUnlockConfirmed
	scheduledAtHeight *big.Int // nil means unscheduled
}

// TODO: Check if hashing is a good key method. Maybe simply using unlock sequence could be better.
func (wfc *withdrawalFinalityCheck) key() string {
	hash, _ := computeAttestationKey(
		wfc.event.UnlockSequenceNumber,
		wfc.event.Recipient[:],
		wfc.event.Token,
		wfc.event.Amount,
		wfc.event.Chain,
	)
	return hash.Hex()
}

func (bw *BridgeWorker) observeBitcoinWithdrawals(ctx context.Context) error {
	finalizedBlock, err := bw.chain.FinalizedBlock(ctx)
	if err != nil {
		return fmt.Errorf("failed to get finalized block: [%w]", err)
	}
	endBlock := finalizedBlock.Uint64()

	startBlock := uint64(0)
	if endBlock > assetsUnlockConfirmedLookBackBlocks {
		startBlock = endBlock - assetsUnlockConfirmedLookBackBlocks
	}

	recentEvents, err := bw.fetchAssetsUnlockConfirmedEvents(
		startBlock,
		endBlock,
	)
	if err != nil {
		return fmt.Errorf(
			"failed to fetch AssetsUnlockConfirmed events: [%w]",
			err,
		)
	}

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
		}
	}

	bw.btcWithdrawalLastProcessedBlock = endBlock

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

func (bw *BridgeWorker) processNewAssetsUnlockConfirmedEvents(
	ctx context.Context,
) error {
	finalizedBlock, err := bw.chain.FinalizedBlock(ctx)
	if err != nil {
		return fmt.Errorf("cannot get current block: [%w]", err)
	}
	endBlock := finalizedBlock.Uint64()

	if endBlock > bw.btcWithdrawalLastProcessedBlock {
		events, err := bw.fetchAssetsUnlockConfirmedEvents(
			bw.btcWithdrawalLastProcessedBlock+1,
			endBlock,
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
			}
		}

		bw.btcWithdrawalLastProcessedBlock = endBlock
	}

	return nil
}

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

				// retry loop with small backoff
				for i := 0; ; i++ {
					withdrawalProcessLogger := withdrawalLogger.With("iteration", i)

					if i > 0 {
						// use a small backoff for subsequent iterations as they
						// are most likely retries
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

					withdrawalProcessLogger.Info(
						"Bitcoin withdrawal submitted",
						"tx_hash", tx.Hash().Hex(),
					)
					break
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
		"entry_hash", check.key(),
		"unlock_sequence", event.UnlockSequenceNumber.String(),
	)
}

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

				key := check.key()

				checkLogger.Info(
					"executing Bitcoin withdrawal finality check",
					"scheduled_at_height", check.scheduledAtHeight.String(),
					"current_finalized_block", currentFinalizedBlock.String(),
					"entry_hash", key,
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
				delete(bw.withdrawalFinalityChecks, key)
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
	_ *portal.MezoBridgeAssetsUnlockConfirmed,
) (
	portal.MezoBridgeAssetsUnlocked,
	[20]byte,
	portal.BitcoinTxUTXO,
	error,
) {
	// TODO: Implement
	return portal.MezoBridgeAssetsUnlocked{},
		[20]byte{},
		portal.BitcoinTxUTXO{},
		fmt.Errorf("unimplemented")
}

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
