package sidecar

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"
	"time"

	"cosmossdk.io/log"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	ethconnect "github.com/mezo-org/mezod/ethereum"
	"github.com/mezo-org/mezod/ethereum/bindings/portal"
	bridgetypes "github.com/mezo-org/mezod/x/bridge/types"
)

var (
	batchAttestationTimeout = 10 * time.Minute
	batchAttestationCheck   = 15 * time.Second
	retrySubmitAttestation  = 5 * time.Second

	ErrBridgeWorkerNotSet = errors.New("bridge worker not set")
)

type BridgeWorker interface {
	// expect no returned payload,
	// just an error eventually
	SubmitAttestation(attestation *bridgetypes.AssetsUnlockedEvent, signature string) error
}

type batchAttestation struct {
	logger         log.Logger
	privateKey     *ecdsa.PrivateKey
	bridgeWorker   BridgeWorker
	bridgeContract ethconnect.BridgeContract
	chainID        *big.Int
}

func newBatchAttestation(
	logger log.Logger,
	privateKey *ecdsa.PrivateKey,
	bridgeWorker BridgeWorker,
	bridgeContract ethconnect.BridgeContract,
	chainID *big.Int,
) *batchAttestation {
	return &batchAttestation{
		logger:         logger,
		privateKey:     privateKey,
		bridgeWorker:   bridgeWorker,
		bridgeContract: bridgeContract,
		chainID:        chainID,
	}
}

// batchAttestation assume the attestation has been
// validated before being called.
func (ba *batchAttestation) TryAttest(
	ctx context.Context,
	originalAttestation *bridgetypes.AssetsUnlockedEvent,
	attestation *portal.MezoBridgeAssetsUnlocked,
) (bool, error) {
	if ba.bridgeWorker == nil {
		return false, ErrBridgeWorkerNotSet
	}
	// main timeout, for the overall time spent
	// trying waiting for the bridge worker to submit the attestations
	attestCtx, cancelAttestCtx := context.WithTimeout(ctx, batchAttestationTimeout)
	defer cancelAttestCtx()

	// first send the attestestation signature.
	// if there's an error we can fallback to
	err := ba.sendPayload(attestCtx, originalAttestation, attestation)
	if err != nil {
		return false, fmt.Errorf("failed to send payload: %w", err)
	}

	checkTicker := time.NewTicker(batchAttestationCheck)
	defer checkTicker.Stop()

	for {
		select {
		case <-checkTicker.C:
			ok := ba.isConfirmed(attestation)
			if ok {
				return true, nil
			}
			// else we just continue to wait for the confirmation
		case <-attestCtx.Done():
			return false, fmt.Errorf("batch attestation terminated: %w", attestCtx.Err())
		}
	}
}

func (ba *batchAttestation) sendPayload(
	ctx context.Context,
	originalAttestation *bridgetypes.AssetsUnlockedEvent,
	attestation *portal.MezoBridgeAssetsUnlocked,
) error {
	signature, err := ba.signPayload(attestation)
	if err != nil {
		// we panic here, there's no reason other than a bug or misconfiguration
		// for not being able to sign the payload here, so let just exit early
		panic(fmt.Sprintf("unable to sign batch attestation payload: %v", err))
	}

	// we operate this in a loop just to handle retries in case
	// of transcient network failure.
	retryTicker := time.NewTicker(retrySubmitAttestation)
	defer retryTicker.Stop()

	sendCtx, cancelSendCtx := context.WithTimeout(ctx, batchAttestationTimeout/5)
	defer cancelSendCtx()

	for {
		select {
		case <-retryTicker.C:
			err := ba.bridgeWorker.SubmitAttestation(originalAttestation, signature)
			if err != nil {
				ba.logger.Warn(
					"failed to send attestation signature to the bridge worker",
					"unlock_sequence", attestation.UnlockSequenceNumber.String(),
					"error", err,
				)
				continue
			}
			return nil
		case <-sendCtx.Done():
			return fmt.Errorf("payload send terminated: %w", sendCtx.Err())
		}
	}
}

func attestationDigestHash(attestation *portal.MezoBridgeAssetsUnlocked, chainID *big.Int) ([]byte, error) {
	abiEncoded, err := abiEncodeAttestationWithChainID(attestation, chainID)
	if err != nil {
		return nil, err
	}

	digest := crypto.Keccak256(abiEncoded)

	return accounts.TextHash(digest), nil
}

func (ba *batchAttestation) signPayload(attestation *portal.MezoBridgeAssetsUnlocked) (string, error) {
	digestHash, err := attestationDigestHash(attestation, ba.chainID)
	if err != nil {
		return "", err
	}

	signature, err := crypto.Sign(digestHash, ba.privateKey)
	if err != nil {
		return "", err
	}

	return hexutil.Encode(signature), nil
}

func (ba *batchAttestation) isConfirmed(attestation *portal.MezoBridgeAssetsUnlocked) bool {
	ok, err := ba.bridgeContract.ConfirmedUnlocks(attestation.UnlockSequenceNumber)
	if err != nil {
		ba.logger.Error(
			"failed to get confirmedUnlocks during batch attestation",
			"unlock_sequence", attestation.UnlockSequenceNumber.String(),
			"error", err,
		)
		return false
	}

	return ok
}
