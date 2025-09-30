package bridgeworker

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/mezo-org/mezod/ethereum/bindings/portal"
	"github.com/mezo-org/mezod/x/bridge/types"
)

const (
	// signatureByteSize is the size in bytes of a single ECDSA signature.
	signatureByteSize = 65

	// batchAttestationProcessingFrequency is the frequency at which new
	// attestation-ready unlock entries are fetched from the server and processed.
	batchAttestationProcessingFrequency = 1 * time.Minute

	// batchAttestationSubmissionBackoff is a backoff time used between retries
	// when submitting a attestBridgeOutWithSignatures transaction.
	batchAttestationSubmissionBackoff = 1 * time.Minute
)

// batchAttestationFinalityCheck is a struct that contains an AssetsUnlock event
// and the Ethereum height at which the batch attestation finality check for it
// was scheduled.
type batchAttestationFinalityCheck struct {
	event             *types.AssetsUnlockedEvent
	scheduledAtHeight *big.Int // nil means unscheduled
}

func (wfc *batchAttestationFinalityCheck) key() string {
	return wfc.event.UnlockSequence.String()
}

type batchAttestationJob struct {
	env *environment

	server *Server

	batchAttestationFinalityChecksMutex sync.Mutex
	batchAttestationFinalityChecks      map[string]*batchAttestationFinalityCheck
}

func newBatchAttestationJob(
	env *environment,
	server *Server,
) *batchAttestationJob {
	return &batchAttestationJob{
		env:                            env,
		server:                         server,
		batchAttestationFinalityChecks: make(map[string]*batchAttestationFinalityCheck),
	}
}

func (baj *batchAttestationJob) run(ctx context.Context) {
	runCtx, cancelRunCtx := context.WithCancel(ctx)
	defer cancelRunCtx()

	go func() {
		defer cancelRunCtx()
		baj.submitBatchAttestations(runCtx)
		baj.env.logger.Warn("batch attestations submission routine stopped")
	}()

	go func() {
		defer cancelRunCtx()
		baj.processBatchAttestationFinalityChecks(runCtx)
		baj.env.logger.Warn("batch attestation finality checks routine stopped")
	}()

	<-runCtx.Done()
	baj.env.logger.Info("batch attestation job stopped")
}

func (baj *batchAttestationJob) submitBatchAttestations(ctx context.Context) {
	ticker := time.NewTicker(batchAttestationProcessingFrequency)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			readySequences, err := baj.server.getBatchAttestationReadyUnlockSequences()
			if err != nil {
				baj.env.logger.Error(
					"failed to get unlock sequences for batch attestation",
					"err", err,
				)
				continue
			}

			if len(readySequences) == 0 {
				baj.env.logger.Info(
					"no unlock sequences ready for batch attestations",
				)
				continue
			}

			for _, unlockSequence := range readySequences {
				batchAttestationLogger := baj.env.logger.With(
					"unlock_sequence", unlockSequence,
				)

				entry, signatures, err := baj.server.getBatchAttestationData(unlockSequence)
				if err != nil {
					batchAttestationLogger.Error(
						"failed to get batch attestation data",
						"err", err,
					)
					continue
				}

				// Check if unlock sequences match, just in case.
				if !entry.UnlockSequence.Equal(unlockSequence) {
					batchAttestationLogger.Error(
						"unexpected unlock sequence in retrieved unlock entry",
					)
					continue
				}

				isConfirmed, err := baj.env.mezoBridgeContract.ConfirmedUnlocks(
					unlockSequence.BigInt(),
				)
				if err != nil {
					batchAttestationLogger.Error(
						"failed to check if unlock entry is already confirmed",
						"err", err,
					)
					continue
				}

				if isConfirmed {
					batchAttestationLogger.Info("unlock entry is already confirmed")
					// Set status to "processed" in the database. The unlock
					// has already been confirmed, either through individual
					// attestations or batch attestation.
					err := baj.server.setBatchAttestationStatus(unlockSequence, "processed")
					if err != nil {
						batchAttestationLogger.Error(
							"failed to set batch attestation status",
							"err", err,
						)
					}

					// Schedule finality check. If the transaction is reverted,
					// the unlock entry can still return to a state where it
					// must be submitted again.
					baj.queueBatchAttestationFinalityCheck(entry)
					continue
				}

				batchAttestationLogger.Info("starting batch attestation execution")

				attestationThreshold, err := baj.env.mezoBridgeContract.AttestationThreshold()
				if err != nil {
					batchAttestationLogger.Error(
						"failed to get attestation threshold",
						"err", err,
					)
					continue
				}

				if uint64(len(signatures)) < attestationThreshold.Uint64() {
					batchAttestationLogger.Error(
						"number of signatures below attestation threshold",
						"number_signatures", len(signatures),
						"attestation_threshold", attestationThreshold.Uint64(),
					)
					continue
				}

				// The server storing signatures should return the exact number of
				// signatures needed. Truncate just in case the attestation threshold
				// changed.
				signatures = signatures[:attestationThreshold.Uint64()]

				concatenatedSignatures, err := concatenateSignatures(signatures)
				if err != nil {
					batchAttestationLogger.Error(
						"failed to concatenate signatures",
						"err", err,
					)
					continue
				}

				assetsUnlockedEvent := portal.MezoBridgeAssetsUnlocked{
					UnlockSequenceNumber: entry.UnlockSequence.BigInt(),
					Recipient:            entry.Recipient,
					Token:                common.HexToAddress(entry.Token),
					Amount:               entry.Amount.BigInt(),
					Chain:                uint8(entry.Chain),
				}

				attestationSuccessful := false

				// Run a retry loop with `3` attempts with a short backoff
				// sleeping period between them. We cannot spend too much time
				// on attesting a single unlock entry. Batch attestations are
				// time-based and validators will proceed with costly individual
				// attestations if we do not submit batch attestation transaction
				// on time.
				for i := 0; i < 3; i++ {
					batchAttestationSubmissionLogger := batchAttestationLogger.With("iteration", i)

					if i > 0 {
						// use a backoff for subsequent iterations as they are
						// most likely retries
						select {
						case <-time.After(batchAttestationSubmissionBackoff):
						case <-ctx.Done():
							batchAttestationSubmissionLogger.Warn(
								"stopping batch attestation submission backoff " +
									"wait due to context cancellation",
							)
							return
						}
					}

					// check if the unlock is still unconfirmed
					isConfirmed, err := baj.env.mezoBridgeContract.ConfirmedUnlocks(
						unlockSequence.BigInt(),
					)
					if err != nil {
						batchAttestationSubmissionLogger.Error(
							"failed to check if unlock is still unconfirmed; "+
								"retrying",
							"err", err,
						)
						continue
					}

					if isConfirmed {
						batchAttestationSubmissionLogger.Info(
							"unlock entry is already confirmed; skipping",
						)
						// Set status to "processed" in the database. The unlock
						// has already been confirmed, either through individual
						// attestations or batch attestation.
						err := baj.server.setBatchAttestationStatus(unlockSequence, "processed")
						if err != nil {
							batchAttestationSubmissionLogger.Error(
								"failed to set batch attestation status",
								"err", err,
							)
						}

						// Schedule finality check. If the transaction is reverted,
						// the unlock entry can still return to a state where it
						// must be submitted again.
						baj.queueBatchAttestationFinalityCheck(entry)

						attestationSuccessful = true
						break
					}

					batchAttestationSubmissionLogger.Info(
						"submitting batch attestation transaction",
					)

					tx, err := baj.env.mezoBridgeContract.AttestBridgeOutWithSignatures(
						assetsUnlockedEvent,
						concatenatedSignatures,
					)
					if err != nil {
						batchAttestationSubmissionLogger.Error(
							"batch attestation transaction submission failed; "+
								"retrying",
							"err", err,
						)
						continue
					}

					// Consider the attestation as processed.
					err = baj.server.setBatchAttestationStatus(unlockSequence, "processed")
					if err != nil {
						batchAttestationSubmissionLogger.Error(
							"failed to set batch attestation status",
							"err", err,
						)
					}

					// Schedule finality check. If the transaction is reverted,
					// the unlock entry can still return to a state where it
					// must be submitted again.
					baj.queueBatchAttestationFinalityCheck(entry)

					attestationSuccessful = true

					batchAttestationSubmissionLogger.Info(
						"submitted attest bridge-out with signatures transaction",
						"tx_hash", tx.Hash().Hex(),
					)

					break
				}

				if !attestationSuccessful {
					batchAttestationLogger.Error(
						"all batch attestation attempts failed",
					)
					// All attempts at batch attestation failed. This particular
					// unlock entry will have to be processed again. We did not
					// set its status to "processed", so it remains
					// "ready_for_submission" and will be picked up for
					// submission in the next round of processing.
				}
			}
		case <-ctx.Done():
			baj.env.logger.Warn(
				"stopping batch attestation submission loop due to context " +
					"cancellation",
			)
			return
		}
	}
}

// concatenateSignatures takes signatures stored as hex-strings and returns their
// byte representations concatenated into a single array of bytes. The input
// signatures should be prepended with `0x` and be 132-char long.
func concatenateSignatures(signatures []string) ([]byte, error) {
	buffer := make([]byte, len(signatures)*signatureByteSize)

	for i, s := range signatures {
		decoded, err := hexutil.Decode(s)
		if err != nil {
			return nil, err
		}

		if len(decoded) != signatureByteSize {
			return nil, fmt.Errorf(
				"incorrect length of signature (%d); expected %d bytes",
				len(decoded),
				signatureByteSize,
			)
		}
		copy(buffer[i*signatureByteSize:(i+1)*signatureByteSize], decoded)
	}

	return buffer, nil
}

func (baj *batchAttestationJob) queueBatchAttestationFinalityCheck(
	event *types.AssetsUnlockedEvent,
) {
	baj.batchAttestationFinalityChecksMutex.Lock()
	defer baj.batchAttestationFinalityChecksMutex.Unlock()

	check := &batchAttestationFinalityCheck{
		event:             event,
		scheduledAtHeight: nil,
	}

	// This should never happen. Check just in case.
	if _, ok := baj.batchAttestationFinalityChecks[check.key()]; ok {
		return
	}

	baj.batchAttestationFinalityChecks[check.key()] = check

	baj.env.logger.Info(
		"queued batch attestation finality check",
		"unlock_sequence", event.UnlockSequence.String(),
	)
}

func (baj *batchAttestationJob) processBatchAttestationFinalityChecks(ctx context.Context) {
	tickerChan := baj.env.chain.WatchBlocks(ctx)

	for {
		select {
		case height := <-tickerChan:
			currentFinalizedBlock, err := baj.env.chain.FinalizedBlock(ctx)
			if err != nil {
				baj.env.logger.Error(
					"cannot get finalized block during batch attestation "+
						"finality checks - skipping current iteration",
					"err", err,
				)
				continue
			}

			checksToExecute := make([]*batchAttestationFinalityCheck, 0)

			// Lock the mutex but do not hold it for too long. Just schedule
			// new checks and determine which ones are ready to be executed.
			baj.batchAttestationFinalityChecksMutex.Lock()
			for _, check := range baj.batchAttestationFinalityChecks {
				checkLogger := baj.env.logger.With(
					"unlock_sequence", check.event.UnlockSequence.String(),
				)

				// First, schedule the check if needed.
				if check.scheduledAtHeight == nil {
					//nolint:gosec
					check.scheduledAtHeight = big.NewInt(int64(height))

					checkLogger.Info(
						"batch attestation finality check scheduled",
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
			baj.batchAttestationFinalityChecksMutex.Unlock()

			for _, check := range checksToExecute {
				checkLogger := baj.env.logger.With(
					"unlock_sequence", check.event.UnlockSequence.String(),
				)

				checkLogger.Info(
					"executing batch attestation finality check",
					"scheduled_at_height", check.scheduledAtHeight.String(),
					"current_finalized_block", currentFinalizedBlock.String(),
				)

				isConfirmed, err := baj.env.mezoBridgeContract.ConfirmedUnlocks(
					check.event.UnlockSequence.BigInt(),
				)
				if err != nil {
					checkLogger.Error(
						"error while checking batch attestation finality "+
							"- retry will be done during next iteration",
						"error", err,
					)
					// Continue to the next check without removing the current one
					// from the queue upon error. This way, the check will be retried
					// during the next iteration.
					continue
				}

				if isConfirmed {
					checkLogger.Info(
						"batch attestation confirmed during finality check",
					)
				} else {
					// If the unlock entry is not confirmed, we need to set its
					// status to "ready_for_submission", so that it's processed
					// again.
					err := baj.server.setBatchAttestationStatus(
						check.event.UnlockSequence,
						"ready_for_submission",
					)
					if err != nil {
						checkLogger.Error(
							"failed to set batch attestation status",
							"err", err,
						)
					}

					checkLogger.Info(
						"batch attestation still unconfirmed during finality check; " +
							"it will have to be resubmitted",
					)
				}

				// Batch attestation check is done, remove the check from the queue.
				baj.batchAttestationFinalityChecksMutex.Lock()
				delete(baj.batchAttestationFinalityChecks, check.key())
				baj.batchAttestationFinalityChecksMutex.Unlock()
			}
		case <-ctx.Done():
			baj.env.logger.Warn(
				"stopping batch attestation finality checks due to context " +
					"cancellation",
			)
			return
		}
	}
}
