package types

import (
	"cosmossdk.io/math"
)

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

	BTCMintedKey = []byte{0x50}

	BTCBurntKey = []byte{0x60}

	// AssetsUnlockedSequenceTipKey is a standalone key for the assets unlocked sequence tip.
	AssetsUnlockedSequenceTipKey = []byte{0x70}

	// AssetsUnlockedSequenceKeyPrefix is the key prefix for the assets unlocked sequence.
	AssetsUnlockedSequenceKeyPrefix = []byte{0x80}
)

// GetERC20TokenMappingKey gets the key for an ERC20 token mapping by the
// corresponding source token address.
func GetERC20TokenMappingKey(sourceERC20Token []byte) []byte {
	return append(ERC20TokenMappingKeyPrefix, sourceERC20Token...)
}

// GetAssetUnlockedKey gets the key for an AssetsUnlocked event.
func GetAssetsUnlockedKey(sequence math.Int) []byte {
	return append(AssetsUnlockedSequenceKeyPrefix, sequence.BigInt().Bytes()...)
}
