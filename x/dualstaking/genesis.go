package dualstaking

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v12/x/dualstaking/keeper"
	"github.com/evmos/evmos/v12/x/dualstaking/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	for _, position := range genState.StakingPositions {
		k.SetStakingPosition(ctx, position)
	}
	for _, position := range genState.DelegationPositions {
		k.SetDelegationPosition(ctx, position)
	}
}

// ExportGenesis returns the module's exported genesis
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	// var stakingPositions []types.StakingPosition
	// var delegationPositions []types.DelegationPosition

	genesis := types.DefaultGenesis()

	// Add all staking positions to the genesis state


	return genesis
}
