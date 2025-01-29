package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/mezo-org/mezod/x/bridge/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	err := k.SetParams(ctx, genState.Params)
	if err != nil {
		panic(errorsmod.Wrapf(err, "error setting params"))
	}

	k.setAssetsLockedSequenceTip(ctx, genState.AssetsLockedSequenceTip)
	k.setSourceBTCToken(ctx, genState.SourceBtcToken)
	k.setERC20TokensMappings(ctx, genState.Erc20TokensMappings)
}

// ExportGenesis returns the module's exported genesis
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	return &types.GenesisState{
		Params:                  k.GetParams(ctx),
		AssetsLockedSequenceTip: k.GetAssetsLockedSequenceTip(ctx),
		SourceBtcToken:          k.GetSourceBTCToken(ctx),
		Erc20TokensMappings:     k.GetERC20TokensMappings(ctx),
	}
}
