package types

import evmtypes "github.com/mezo-org/mezod/x/evm/types"

const (
	// ModuleName defines the module name
	ModuleName = "bridge"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName
)

var (
	// ParamsKey is a standalone key for module params.
	ParamsKey = []byte{0x10}

	// AssetsLockedSequenceTipKey is a standalone key for the assets locked sequence tip.
	AssetsLockedSequenceTipKey = []byte{0x20}

	// SourceBTCTokenKey is a standalone key for the BTC token address on the
	// source chain. AssetsLocked events carrying this token address are
	// directly mapped to the Mezo native denomination - BTC.
	//
	// For example, if the source chain for the Mezo bridge is Ethereum,
	// and we use TBTC as the BTC token, this key should point to the TBTC
	// token address on the Ethereum chain.
	SourceBTCTokenKey = []byte{0x30}

	// ERC20TokenMappingKeyPrefix is a prefix used to construct a key to an
	// ERC20 token mapping, supported by the bridge. Specifically, a key to an
	// ERC20 token mapping is constructed by taking this prefix and appending the
	// corresponding token contract address on the source chain. In other words,
	// this prefix is used to construct the following mapping:
	// ERC20TokenMappingKeyPrefix + SourceERC20Token -> MezoERC20Token
	ERC20TokenMappingKeyPrefix = []byte{0x40}
)

// GetERC20TokenMappingKey gets the key for an ERC20 token mapping by the
// corresponding source token address.
func GetERC20TokenMappingKey(sourceERC20Token string) []byte {
	sourceERC20TokenBytes := evmtypes.HexAddressToBytes(sourceERC20Token)
	return append(ERC20TokenMappingKeyPrefix, sourceERC20TokenBytes...)
}
