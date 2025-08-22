package sidecar

import (
	"fmt"
	"math/big"

	"cosmossdk.io/log"
	"github.com/ethereum/go-ethereum/accounts/abi"
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
		return false, fmt.Errorf("couldn't get confirmedLocks: %w", err)
	}
	if ok {
		return true, nil
	}

	return av.checkOwnAttestation(attestation)
}

func (av *attestationValidator) checkOwnAttestation(
	attestation *portal.MezoBridgeAssetsUnlocked,
) (bool, error) {
	encoded, err := abiEncodeAttestation(attestation)
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

func abiEncodeAttestation(attestation *portal.MezoBridgeAssetsUnlocked) ([]byte, error) {
	uint256Type, err := abi.NewType("uint256", "uint256", nil)
	if err != nil {
		return nil, err
	}
	bytesType, err := abi.NewType("bytes", "bytes", nil)
	if err != nil {
		return nil, err
	}
	addressType, err := abi.NewType("address", "address", nil)
	if err != nil {
		return nil, err
	}
	uint8Type, err := abi.NewType("uint8", "uint8", nil)
	if err != nil {
		return nil, err
	}

	return abi.Arguments{
		{Type: uint256Type},
		{Type: bytesType},
		{Type: addressType},
		{Type: uint256Type},
		{Type: uint8Type},
	}.Pack(
		attestation.UnlockSequenceNumber,
		attestation.Recipient,
		attestation.Token,
		attestation.Amount,
		attestation.Chain,
	)
}
