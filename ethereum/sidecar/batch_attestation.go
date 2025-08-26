package sidecar

import (
	"context"
	"crypto/ecdsa"
	"time"

	"cosmossdk.io/log"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	ethconnect "github.com/mezo-org/mezod/ethereum"
	"github.com/mezo-org/mezod/ethereum/bindings/portal"
)

var (
	defaultBatchAttestationTimeout = 5 * time.Minute
	defaultBatchAttestationCheck   = 15 * time.Second
	defaultRetrySendSignature      = 5 * time.Second
)

type BridgeWorker interface {
	// expect no returned payload,
	// just an error eventually
	SendSignature(address common.Address, signature string) error
}

type batchAttestation struct {
	logger         log.Logger
	privateKey     *ecdsa.PrivateKey
	address        common.Address
	bridgeWorker   BridgeWorker
	bridgeContract ethconnect.BridgeContract
}

func newBatchAttestation(
	logger log.Logger,
	privateKey *ecdsa.PrivateKey,
	address common.Address,
	bridgeWorker BridgeWorker,
	bridgeContract ethconnect.BridgeContract,
) *batchAttestation {
	return &batchAttestation{
		logger:         logger,
		address:        address,
		privateKey:     privateKey,
		bridgeWorker:   bridgeWorker,
		bridgeContract: bridgeContract,
	}
}

// batchAttestation assume the attestation has been
// validated before being called.
func (ba *batchAttestation) TryAttest(
	ctx context.Context,
	attestation *portal.MezoBridgeAssetsUnlocked,
) (bool, error) {
	// main timeout, for the overall time spent
	// trying waiting for the bridge worker to submit the attestations
	cancelCtx, cancel := context.WithTimeout(context.Background(), defaultBatchAttestationTimeout)
	defer cancel()

	// first send the attestestation signature.
	// if there's an error we can fallback to
	ok, err := ba.sendPayload(ctx, cancelCtx, attestation)
	if !ok || err != nil {
		// if either there was an error = ctx canceled or sending
		// payload is not ok == reached default batch timeout
		return ok, err
	}

	checkTicker := time.NewTicker(defaultBatchAttestationCheck)
	defer checkTicker.Stop()

	for {
		select {
		case <-checkTicker.C:
			ok := ba.isConfirmed(attestation)
			if ok {
				return true, nil
			}
			// else we just continue to wait for the confirmation
		case <-cancelCtx.Done():
			ba.logger.Info("stopping batch attestation slot wait due timeout")
			return false, nil
		case <-ctx.Done():
			ba.logger.Info("stopping batch attestation slot wait due to context cancellation")
			return false, ctx.Err()
		}
	}
}

func (ba *batchAttestation) sendPayload(
	ctx context.Context,
	cancelCtx context.Context,
	attestation *portal.MezoBridgeAssetsUnlocked,
) (bool, error) {
	signature, err := ba.signPayload(attestation)
	if err != nil {
		// we panic here, there's no reason than a bug or misconfiguration
		// for not being able to sign the payload here, so let just exit early
		ba.logger.Error("couldn't sign batch attestation payload", "attestation", attestation, "error", err)
		panic("unable to sign batch attestation payload")
	}

	// we operate this in a loop just to handle retries in case
	// of transcient network failure.
	retryTicker := time.NewTicker(defaultRetrySendSignature)
	defer retryTicker.Stop()

	for {
		select {
		case <-retryTicker.C:
			err := ba.bridgeWorker.SendSignature(ba.address, signature)
			if err != nil {
				ba.logger.Info("couldn't send signature to bridge worker",
					"attestation", attestation, "error", err)
				continue
			}
			return true, nil
		case <-cancelCtx.Done():
			ba.logger.Info("stopping batch attestation slot wait due timeout")
			return false, nil
		case <-ctx.Done():
			ba.logger.Info("stopping batch attestation slot wait due to context cancellation")
			return false, ctx.Err()
		}
	}
}

func (ba *batchAttestation) signPayload(attestation *portal.MezoBridgeAssetsUnlocked) (string, error) {
	abiEncoded, err := abiEncodeAttestation(attestation)
	if err != nil {
		return "", err
	}

	signature, err := crypto.Sign(accounts.TextHash(abiEncoded), ba.privateKey)
	if err != nil {
		return "", err
	}

	return hexutil.Encode(signature), nil
}

func (ba *batchAttestation) isConfirmed(attestation *portal.MezoBridgeAssetsUnlocked) bool {
	ok, err := ba.bridgeContract.ConfirmedUnlocks(attestation.UnlockSequenceNumber)
	if err != nil {
		ba.logger.Error("couldn't get confirmedLocks", "error", err)
	}

	return ok
}
