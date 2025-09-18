package portal

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	ethconfig "github.com/keep-network/keep-common/pkg/chain/ethereum"
	ethereumgen "github.com/mezo-org/mezod/ethereum/bindings/portal/ethereum/gen"
	ethereumabi "github.com/mezo-org/mezod/ethereum/bindings/portal/ethereum/gen/abi"
	ethereumcontract "github.com/mezo-org/mezod/ethereum/bindings/portal/ethereum/gen/contract"
	sepoliagen "github.com/mezo-org/mezod/ethereum/bindings/portal/sepolia/gen"
)

// TODO: The current bindings implementation uses a simplified approach for handling multiple Ethereum networks.
//       A more robust solution would involve creating a facade component within the `portal` package
//       that dynamically loads the appropriate contract bindings at runtime, abstracting this complexity
//       from client code.
//
//       Currently, we're using Mainnet contract bindings for both Sepolia and Mainnet networks in the
//       `mezod/ethereum` package. While this introduces some technical debt, it's an intentional trade-off
//       since the contract interfaces are currently identical across networks. The only difference is in the
//       contract addresses, which are properly configured per environment.
//
//       As long as this approach is in place, we optimize the binary size by generating bindings only for Mainnet,
//       while Sepolia bindings are limited to address definitions.
//
//       This approach is sustainable as long as there are no differences between Mainnet and Sepolia contracts.
//       Once such differences emerge, we need to revisit it.

func MezoBridgeAddress(network ethconfig.Network) string {
	switch network {
	case ethconfig.Sepolia:
		return sepoliagen.MezoBridgeAddress
	case ethconfig.Mainnet:
		return ethereumgen.MezoBridgeAddress
	default:
		panic("unknown ethereum network")
	}
}

type (
	MezoBridge                      = ethereumcontract.MezoBridge
	MezoBridgeAssetsLocked          = ethereumabi.MezoBridgeAssetsLocked
	MezoBridgeAssetsUnlockConfirmed = ethereumabi.MezoBridgeAssetsUnlockConfirmed
	MezoBridgeAssetsUnlocked        = ethereumabi.MezoBridgeAssetsUnlocked
	BitcoinTxUTXO                   = ethereumabi.BitcoinTxUTXO
)

var NewMezoBridge = ethereumcontract.NewMezoBridge

func AttestationDigestHash(attestation *MezoBridgeAssetsUnlocked, chainID *big.Int) ([]byte, error) {
	abiEncoded, err := AbiEncodeAttestationWithChainID(attestation, chainID)
	if err != nil {
		return nil, err
	}

	digest := crypto.Keccak256(abiEncoded)

	return accounts.TextHash(digest), nil
}

func AbiEncodeAttestation(attestation *MezoBridgeAssetsUnlocked) ([]byte, error) {
	return AbiEncodeAttestationWithChainID(attestation, nil)
}

// abiEncodeAttestationWithChainID is used to encode the attestation with the chain ID
// which is used to produce a signature for the batch attestation process.
func AbiEncodeAttestationWithChainID(attestation *MezoBridgeAssetsUnlocked, chainID *big.Int) ([]byte, error) {
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
