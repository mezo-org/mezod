package types

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

	// SupportedERC20TokenKeyPrefix is a prefix used to construct a key to a
	// Mezo ERC20 token, supported by the bridge. Specifically, a key to a
	// Mezo ERC20 token is constructed by taking this prefix and appending the
	// corresponding token contract address on the source chain. In other words,
	// this prefix is used to construct the following mapping:
	// SupportedERC20TokenKeyPrefix + SourceERC20Token -> MezoERC20Token
	SupportedERC20TokenKeyPrefix = []byte{0x40}
)

// GetSupportedERC20TokenKey gets the key for a Mezo ERC20 bridgeable token
// address by the corresponding source token address.
func GetSupportedERC20TokenKey(sourceERC20Token []byte) []byte {
	return append(SupportedERC20TokenKeyPrefix, sourceERC20Token...)
}
