package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/mezo-org/mezod/x/bridge/types"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func (k Keeper) InitGenesis(
	ctx sdk.Context,
	genState types.GenesisState,
	accountKeeper types.AccountKeeper,
) {
	// Ensure x/bridge module account is created, if not already exist.
	if acc := accountKeeper.GetModuleAccount(ctx, types.ModuleName); acc == nil {
		panic("the x/bridge module account has not been set")
	}

	err := k.SetParams(ctx, genState.Params)
	if err != nil {
		panic(errorsmod.Wrapf(err, "error setting params"))
	}

	k.setAssetsLockedSequenceTip(ctx, genState.AssetsLockedSequenceTip)
	k.SetSourceBTCToken(ctx, evmtypes.HexAddressToBytes(genState.SourceBtcToken))
	k.setERC20TokensMappings(ctx, genState.Erc20TokensMappings)

	err = k.IncreaseBTCMinted(ctx, genState.InitialBtcSupply)
	if err != nil {
		panic(errorsmod.Wrapf(err, "error setting params"))
	}
}

// ExportGenesis returns the module's exported genesis
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	return &types.GenesisState{
		Params:                  k.GetParams(ctx),
		AssetsLockedSequenceTip: k.GetAssetsLockedSequenceTip(ctx),
		SourceBtcToken:          evmtypes.BytesToHexAddress(k.GetSourceBTCToken(ctx)),
		Erc20TokensMappings:     k.GetERC20TokensMappings(ctx),
		InitialBtcSupply:        k.GetBTCMinted(ctx).Sub(k.GetBTCBurnt(ctx)),
	}
}
