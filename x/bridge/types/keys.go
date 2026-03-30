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

	// AssetsUnlockedKeyPrefix is the key prefix for the assets unlocked key.
	AssetsUnlockedKeyPrefix = []byte{0x80}

	// OutflowLimitKeyPrefix is the key prefix for per-token outflow limits.
	OutflowLimitKeyPrefix = []byte{0x90}

	// CurrentOutflowKeyPrefix is the key prefix for tracking current outflow per token.
	CurrentOutflowKeyPrefix = []byte{0x91}

	// LastOutflowResetKey is a standalone key for tracking when outflow was last reset.
	LastOutflowResetKey = []byte{0x92}

	// PauserKey is a standalone key for the pauser address.
	PauserKey = []byte{0x93}

	// MinBridgeOutAmountKeyPrefix is the key prefix for the minimum bridge-out
	// amount.
	MinBridgeOutAmountKeyPrefix = []byte{0x94}

	// MinBridgeOutAmountForBitcoinChainKey is a standalone key for the minimum
	// bridge-out amount that applies specifically to the Bitcoin chain.
	MinBridgeOutAmountForBitcoinChainKey = []byte{0x95}

	// TripartyControllerKeyPrefix is a prefix used to construct a key to a
	// triparty controller entry. A key is constructed by taking this prefix
	// and appending the controller address.
	TripartyControllerKeyPrefix = []byte{0xA0}

	// TripartyPausedKey is a standalone key for the triparty paused flag.
	// If the key is present in the store, triparty bridging is paused.
	TripartyPausedKey = []byte{0xA1}

	// TripartyBlockDelayKey is a standalone key for the triparty block delay.
	TripartyBlockDelayKey = []byte{0xA2}

	// TripartyPerRequestLimitKey is a standalone key for the triparty
	// per-request limit.
	TripartyPerRequestLimitKey = []byte{0xA3}

	// TripartyWindowLimitKey is a standalone key for the triparty window
	// limit.
	TripartyWindowLimitKey = []byte{0xA4}

	// TripartyRequestKeyPrefix is a prefix used to construct a key to a
	// pending triparty bridge request.
	TripartyRequestKeyPrefix = []byte{0xA5}

	// TripartySequenceTipKey is a standalone key for the last assigned
	// triparty request sequence number.
	TripartySequenceTipKey = []byte{0xA6}
)

// GetERC20TokenMappingKey gets the key for an ERC20 token mapping by the
// corresponding source token address.
func GetERC20TokenMappingKey(sourceERC20Token []byte) []byte {
	return append(ERC20TokenMappingKeyPrefix, sourceERC20Token...)
}

// GetAssetsUnlockedKey gets the key for an AssetsUnlocked event.
func GetAssetsUnlockedKey(unlockSequence math.Int) []byte {
	return append(AssetsUnlockedKeyPrefix, unlockSequence.BigInt().Bytes()...)
}

// GetOutflowLimitKey gets the key for an outflow limit by token address.
func GetOutflowLimitKey(token []byte) []byte {
	return append(OutflowLimitKeyPrefix, token...)
}

// GetCurrentOutflowKey gets the key for current outflow tracking by token address.
func GetCurrentOutflowKey(token []byte) []byte {
	return append(CurrentOutflowKeyPrefix, token...)
}

// GetMinBridgeOutAmountKey gets the key for minimum bridge-out amount by the
// given Mezo token address.
func GetMinBridgeOutAmountKey(mezoToken []byte) []byte {
	return append(MinBridgeOutAmountKeyPrefix, mezoToken...)
}

// GetTripartyControllerKey gets the key for a triparty controller by address.
func GetTripartyControllerKey(controller []byte) []byte {
	return append(TripartyControllerKeyPrefix, controller...)
}

// GetTripartyBridgeRequestKey gets the key for a pending triparty bridge
// request by its sequence number.
func GetTripartyBridgeRequestKey(sequence math.Int) []byte {
	return append(TripartyRequestKeyPrefix, sequence.BigInt().Bytes()...)
}
