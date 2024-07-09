package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v12/x/poa/types"
)

func (k Keeper) GetHistoricalInfo(
	ctx sdk.Context,
	height int64,
) (types.HistoricalInfo, bool) {
	// TODO: Implement GetHistoricalInfo function.
	return types.HistoricalInfo{}, false
}