package bridge

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v12/x/bridge/keeper"
	"github.com/evmos/evmos/v12/x/bridge/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(_ sdk.Context, _ keeper.Keeper, _ types.GenesisState) {
}

// ExportGenesis returns the module's exported genesis
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.Params = k.GetParams(ctx)

	return genesis
}
