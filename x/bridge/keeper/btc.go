package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/mezo-org/mezod/x/bridge/types"
)

// GetSourceBTCToken returns the BTC token address on the source chain.
// AssetsLocked events carrying this token address are directly mapped to the
// Mezo native denomination - BTC.
func (k Keeper) GetSourceBTCToken(ctx sdk.Context) []byte {
	return ctx.KVStore(k.storeKey).Get(types.SourceBTCTokenKey)
}

func (k Keeper) setSourceBTCToken(ctx sdk.Context, sourceBTCToken []byte) {
	ctx.KVStore(k.storeKey).Set(types.SourceBTCTokenKey, sourceBTCToken)
}
