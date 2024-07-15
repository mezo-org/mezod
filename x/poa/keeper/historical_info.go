package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v12/x/poa/types"
)

// GetHistoricalInfo gets the historical info at a given height
func (k Keeper) GetHistoricalInfo(
	ctx sdk.Context,
	height int64,
) (types.HistoricalInfo, bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetHistoricalInfoKey(height)

	value := store.Get(key)
	if value == nil {
		return types.HistoricalInfo{}, false
	}

	return types.MustUnmarshalHistoricalInfo(k.cdc, value), true
}

// SetHistoricalInfo sets the historical info at a given height
func (k Keeper) SetHistoricalInfo(
	ctx sdk.Context,
	height int64,
	hi *types.HistoricalInfo,
) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetHistoricalInfoKey(height)
	value := k.cdc.MustMarshal(hi)
	store.Set(key, value)
}
