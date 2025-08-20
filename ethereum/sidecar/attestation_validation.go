package sidecar

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	ethconnect "github.com/mezo-org/mezod/ethereum"
	"github.com/mezo-org/mezod/ethereum/bindings/portal"
)

var (
	ErrInvalidAttestation      = errors.New("invalid attestation")
	ErrValidatorNotInTheBitmap = errors.New("validatorId not in the bitmap yet")
)

type AttestationValidation struct {
	bridgeContract ethconnect.BridgeContract
	address        common.Address
}

func NewAttestationValidation(
	bridgeContract ethconnect.BridgeContract,
	validatorAddress common.Address,
) *AttestationValidation {
	return &AttestationValidation{
		bridgeContract: bridgeContract,
		address:        validatorAddress,
	}
}

func (av *AttestationValidation) IsConfirmed(
	attestation *portal.MezoBridgeAssetsUnlocked,
) error {
	ok, err := av.bridgeContract.ValidateAssetsUnlocked(*attestation)
	if err != nil {
		return fmt.Errorf("couldn't call validateAssetsUnlocked: %w", err)
	}
	if !ok {
		// this is a specific case of error
		// the attestation is not valid, we specify it down the line
		return ErrInvalidAttestation
	}

	ok, err = av.bridgeContract.ConfirmedUnlocks(attestation.Amount)
	if err != nil {
		return fmt.Errorf("couldn't get confirmedLocks: %w", err)
	}
	if ok {
		return nil
	}

	encoded, err := abiEncodeAttestation(attestation)
	if err != nil {
		return fmt.Errorf("couldn't ABI encode attestation: %w", err)
	}

	hash := crypto.Keccak256Hash(encoded)

	bitmap, err := av.bridgeContract.Attestations(hash)
	if err != nil {
		return fmt.Errorf("couldn't get confirmedLock: %w", err)
	}

	validatorId, err := av.bridgeContract.ValidatorIDs(av.address)
	if err != nil {
		return fmt.Errorf("couldn't get validator ID: %w", err)
	}

	mask := new(big.Int).Lsh(big.NewInt(1), uint(validatorId))

	if new(big.Int).And(bitmap, mask).Int64() != 0 {
		return ErrValidatorNotInTheBitmap
	}

	return nil
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
