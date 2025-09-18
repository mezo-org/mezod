package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
)

func NewERC20TokenMapping(
	sourceToken, mezoToken []byte,
) *ERC20TokenMapping {
	return &ERC20TokenMapping{
		SourceToken: evmtypes.BytesToHexAddress(sourceToken),
		MezoToken:   evmtypes.BytesToHexAddress(mezoToken),
	}
}

// SourceTokenBytes returns the source token EVM address as bytes.
func (m *ERC20TokenMapping) SourceTokenBytes() []byte {
	return evmtypes.HexAddressToBytes(m.SourceToken)
}

// MezoTokenBytes returns the Mezo token EVM address as bytes.
func (m *ERC20TokenMapping) MezoTokenBytes() []byte {
	return evmtypes.HexAddressToBytes(m.MezoToken)
}

// MustMarshalERC20TokenMapping marshals an ERC20TokenMapping to bytes.
// It panics on error.
func MustMarshalERC20TokenMapping(
	cdc codec.BinaryCodec,
	mapping ERC20TokenMapping,
) []byte {
	return cdc.MustMarshal(&mapping)
}

// MustUnmarshalERC20TokenMapping unmarshals an ERC20TokenMapping from bytes.
// It panics on error.
func MustUnmarshalERC20TokenMapping(
	cdc codec.BinaryCodec,
	value []byte,
) ERC20TokenMapping {
	mapping, err := UnmarshalERC20TokenMapping(cdc, value)
	if err != nil {
		panic(err)
	}
	return mapping
}

// UnmarshalERC20TokenMapping unmarshals an ERC20TokenMapping from bytes.
func UnmarshalERC20TokenMapping(
	cdc codec.BinaryCodec,
	value []byte,
) (ERC20TokenMapping, error) {
	var mapping ERC20TokenMapping
	err := cdc.Unmarshal(value, &mapping)
	return mapping, err
}

// MustMarshalAssetsUnlockedEvent marshals an AssetsUnlockedEvent to bytes.
// It panics on error.
func MustMarshalAssetsUnlockedEvent(
	cdc codec.BinaryCodec,
	event AssetsUnlockedEvent,
) []byte {
	return cdc.MustMarshal(&event)
}

// MustUnmarshalAssetsUnlockedEvent unmarshals an AssetsUnlockedEvent from bytes.
// It panics on error.
func MustUnmarshalAssetsUnlockedEvent(
	cdc codec.BinaryCodec,
	value []byte,
) AssetsUnlockedEvent {
	event, err := UnmarshalAssetsUnlockedEvent(cdc, value)
	if err != nil {
		panic(err)
	}
	return event
}

// UnmarshalAssetsUnlockedEvent unmarshals an AssetsUnlockedEvent from bytes.
func UnmarshalAssetsUnlockedEvent(
	cdc codec.BinaryCodec,
	value []byte,
) (AssetsUnlockedEvent, error) {
	var event AssetsUnlockedEvent
	err := cdc.Unmarshal(value, &event)
	return event, err
}
