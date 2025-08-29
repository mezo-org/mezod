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
	return abiEncodeAttestationWithChainID(attestation, nil)
}

// abiEncodeAttestationWithChainID is used to encode the attestation with the chain ID
// which is used to produce a signature for the batch attestation process.
func abiEncodeAttestationWithChainID(attestation *portal.MezoBridgeAssetsUnlocked, chainID *big.Int) ([]byte, error) {
	var argumentsTypes abi.Arguments
	var arguments []any

	if chainID != nil {
		uint256Type, err := abi.NewType("uint256", "uint256", nil)
		if err != nil {
			return nil, err
		}
		argumentsTypes = append(argumentsTypes, abi.Argument{Type: uint256Type})
		arguments = append(arguments, chainID)
	}

	// Create tuple type for AssetsUnlocked struct
	tupleType, err := abi.NewType("tuple", "tuple", []abi.ArgumentMarshaling{
		{Name: "unlockSequenceNumber", Type: "uint256"},
		{Name: "recipient", Type: "bytes"},
		{Name: "token", Type: "address"},
		{Name: "amount", Type: "uint256"},
		{Name: "chain", Type: "uint8"},
	})
	if err != nil {
		return nil, err
	}

	// Add the tuple as a single argument instead of individual fields
	argumentsTypes = append(argumentsTypes, abi.Argument{Type: tupleType})

	// Create the struct as a single tuple argument
	assetsUnlockedTuple := struct {
		UnlockSequenceNumber *big.Int
		Recipient            []byte
		Token                common.Address
		Amount               *big.Int
		Chain                uint8
	}{
		UnlockSequenceNumber: attestation.UnlockSequenceNumber,
		Recipient:            attestation.Recipient,
		Token:                attestation.Token,
		Amount:               attestation.Amount,
		Chain:                attestation.Chain,
	}

	arguments = append(arguments, assetsUnlockedTuple)

	return argumentsTypes.Pack(arguments...)
}
