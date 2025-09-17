package sidecar

import (
	"fmt"
	"math/big"

	"cosmossdk.io/log"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	ethconnect "github.com/mezo-org/mezod/ethereum"
	"github.com/mezo-org/mezod/ethereum/bindings/portal"
)

type attestationValidator struct {
	logger         log.Logger
	bridgeContract ethconnect.BridgeContract
	address        common.Address
}

func newAttestationValidation(
	logger log.Logger,
	bridgeContract ethconnect.BridgeContract,
	validatorAddress common.Address,
) *attestationValidator {
	return &attestationValidator{
		logger:         logger,
		bridgeContract: bridgeContract,
		address:        validatorAddress,
	}
}

func (av *attestationValidator) IsConfirmed(
	attestation *portal.MezoBridgeAssetsUnlocked,
) (bool, error) {
	ok, err := av.bridgeContract.ConfirmedUnlocks(attestation.UnlockSequenceNumber)
	if err != nil {
		return false, fmt.Errorf("couldn't get confirmedUnlocks: %w", err)
	}
	if ok {
		return true, nil
	}

	return av.checkOwnAttestation(attestation)
}

func (av *attestationValidator) checkOwnAttestation(
	attestation *portal.MezoBridgeAssetsUnlocked,
) (bool, error) {
	encoded, err := portal.AbiEncodeAttestation(attestation)
	if err != nil {
		return false, fmt.Errorf("couldn't ABI encode attestation: %w", err)
	}

	hash := crypto.Keccak256Hash(encoded)

	bitmap, err := av.bridgeContract.Attestations(hash)
	if err != nil {
		return false, fmt.Errorf("couldn't get confirmedUnlock: %w", err)
	}

	validatorID, err := av.bridgeContract.ValidatorIDs(av.address)
	if err != nil {
		return false, fmt.Errorf("couldn't get validator ID: %w", err)
	}

	mask := new(big.Int).Lsh(big.NewInt(1), uint(validatorID))

	if new(big.Int).And(bitmap, mask).Int64() == 0 {
		return false, nil
	}

	return true, nil
}
