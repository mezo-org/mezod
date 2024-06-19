package dualstaking

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v12/x/dualstaking/keeper"
	"github.com/evmos/evmos/v12/x/dualstaking/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	for _, position := range genState.Stakes {
		k.SetStake(ctx, position)
	}
	for _, position := range genState.Delegations {
		k.SetDelegation(ctx, position)
	}
}

// ExportGenesis returns the module's exported genesis
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	// var stakes []types.Stake
	// var delegations []types.Delegation

	genesis := types.DefaultGenesis()

	// Add all staking positions to the genesis state

	return genesis
}
