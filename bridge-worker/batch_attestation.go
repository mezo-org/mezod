package bridgeworker

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/mezo-org/mezod/ethereum/bindings/portal"
	"github.com/mezo-org/mezod/x/bridge/types"
)

const (
	// TODO: Consider making this variable configurable.
	defaultBatchAttestationSubmissionReadyEventsFetchFrequency = 30 * time.Second

	// batchAttestationSubmissionBackoff is a backoff time used between retries
	// when submitting a attestBridgeOutWithSignatures transaction.
	batchAttestationSubmissionBackoff = 1 * time.Minute
)

type batchAttestationJob struct {
	env *environment

	server *Server

	submissionReadyEventsFetchFrequency time.Duration
}

func newBatchAttestationJob(
	env *environment,
	server *Server,
) *batchAttestationJob {
	return &batchAttestationJob{
		env:                                 env,
		server:                              server,
		submissionReadyEventsFetchFrequency: defaultBatchAttestationSubmissionReadyEventsFetchFrequency,
	}
}

func (baj *batchAttestationJob) run(ctx context.Context) {
	runCtx, cancelRunCtx := context.WithCancel(ctx)
	defer cancelRunCtx()

	go func() {
		defer cancelRunCtx()
		baj.submitBatchAttestations(runCtx)
		baj.env.logger.Warn("batch attestations submitting routine stopped")
	}()

	go func() {
		defer cancelRunCtx()
		baj.processBatchAttestationFinalityChecks(runCtx)
		baj.env.logger.Warn("batch attestation finality checks routine stopped")
	}()

	<-runCtx.Done()
	baj.env.logger.Info("batch attestation job stopped")
}

// nolint:unused
func (baj *batchAttestationJob) submitBatchAttestations(ctx context.Context) {
	ticker := time.NewTicker(baj.submissionReadyEventsFetchFrequency)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// TODO: Check if splitting getting unlock sequences and getting
			//       attestation data (`AssetsUnlock` entry + signatures) is faster
			//       than getting attestation data with just a single command.
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

				assetsUnlockedEvent := portal.MezoBridgeAssetsUnlocked{
					UnlockSequenceNumber: entry.UnlockSequence.BigInt(),
					Recipient:            entry.Recipient,
					Token:                common.HexToAddress(entry.Token),
					Amount:               entry.Amount.BigInt(),
					Chain:                uint8(entry.Chain),
				}

				isConfirmed, err := baj.env.mezoBridgeContract.ConfirmedUnlocks(
					unlockSequence.BigInt(),
				)
				if err != nil {
					batchAttestationLogger.Error(
						"failed to check if unlock is already confirmed",
						"err", err,
					)
					continue
				}

				if isConfirmed {
					batchAttestationLogger.Info("unlock is already confirmed")
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
					baj.queueWithdrawalFinalityCheck(
						assetsUnlockedEvent.UnlockSequenceNumber,
						entry,
					)

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

				// Run a retry loop with a few attempts.
				for i := 0; i < 5; i++ {
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
							"unlock is already confirmed; skipping",
						)

						// TODO: Set status to "processed" in the database. The unlock
						//       has already been confirmed, either through individual
						//       attestations or batch attestation. However, we have
						//       no guarantee it's finalized so we should schedule it
						//       for checking.
						err := baj.server.setBatchAttestationStatus(unlockSequence, "processed")
						if err != nil {
							batchAttestationSubmissionLogger.Error(
								"failed to set batch attestation status",
								"err", err,
							)
						}

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
					baj.queueWithdrawalFinalityCheck(
						assetsUnlockedEvent.UnlockSequenceNumber,
						entry,
					)

					batchAttestationSubmissionLogger.Info(
						"submitted attest bridge-out with signatures transaction",
						"tx_hash", tx.Hash().Hex(),
					)

					break
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
	buffer := make([]byte, len(signatures)*65)

	for i, s := range signatures {
		decoded, err := hexutil.Decode(s)
		if err != nil {
			return nil, err
		}

		if len(decoded) != 65 {
			return nil, fmt.Errorf(
				"incorrect length of signature (%d); expected 65 bytes",
				len(decoded),
			)
		}
		copy(buffer[i*65:(i+1)*65], decoded)
	}

	return buffer, nil
}

// nolint:unused
func (baj *batchAttestationJob) queueWithdrawalFinalityCheck(
	_ *big.Int,
	_ *types.AssetsUnlockedEvent,
) {
	// TODO: Implement
}

// nolint:unused
func (baj *batchAttestationJob) processBatchAttestationFinalityChecks(ctx context.Context) {
	// TODO: Implement the same way as in BTC withdrawal job.
	<-ctx.Done()
}
