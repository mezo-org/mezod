package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v12/x/poa/types"
)

func (k Keeper) GetHistoricalInfo(
	_ sdk.Context,
	_ int64,
) (types.HistoricalInfo, bool) {
	// TODO: Implement GetHistoricalInfo function.
	return types.HistoricalInfo{}, false
}

func (k Keeper) SetHistoricalInfo(
	_ sdk.Context,
	_ int64,
	_ *types.HistoricalInfo,
) {
	// TODO: Implement SetHistoricalInfo function.
	// TODO: Uncomment tests in x/evm/keeper/state_transition_test.go
	panic("not implemented")
}
