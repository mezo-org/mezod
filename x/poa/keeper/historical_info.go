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

func (k Keeper) SetHistoricalInfo(
	ctx sdk.Context,
	height int64,
	historicalInfo *types.HistoricalInfo,
) {
	// TODO: Implement SetHistoricalInfo function.
	// TODO: Uncomment tests in x/evm/keeper/state_transition_test.go
	panic("not implemented")
}