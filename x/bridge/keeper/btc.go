package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/mezo-org/mezod/x/bridge/types"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
)

// GetSourceBTCToken returns the BTC token address on the source chain.
// AssetsLocked events carrying this token address are directly mapped to the
// Mezo native denomination - BTC.
func (k Keeper) GetSourceBTCToken(ctx sdk.Context) string {
	sourceBTCToken := ctx.KVStore(k.storeKey).Get(types.SourceBTCTokenKey)
	return evmtypes.BytesToHexAddress(sourceBTCToken)
}

func (k Keeper) setSourceBTCToken(ctx sdk.Context, sourceBTCToken string) {
	ctx.KVStore(k.storeKey).Set(
		types.SourceBTCTokenKey,
		evmtypes.HexAddressToBytes(sourceBTCToken),
	)
}
