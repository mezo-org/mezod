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
	k.setAssetsUnlockedSequenceTip(ctx, genState.AssetsUnlockedSequenceTip)
	k.SetSourceBTCToken(ctx, evmtypes.HexAddressToBytes(genState.SourceBtcToken))
	k.setERC20TokensMappings(ctx, genState.Erc20TokensMappings)

	// Initialize assets unlocked events
	for _, event := range genState.AssetsUnlockedEvents {
		k.saveAssetsUnlocked(ctx, event)
	}

	// Initialize minimum bridge out amount for Bitcoin chain
	k.SetMinBridgeOutAmountForBitcoinChain(ctx, genState.BitcoinChainMinBridgeOutAmount)

	// Initialize token minimum bridge out amounts
	for _, tokenAmount := range genState.TokenMinBridgeOutAmounts {
		err := k.SetMinBridgeOutAmount(ctx, tokenAmount.Token, tokenAmount.Amount)
		if err != nil {
			panic(errorsmod.Wrapf(err, "error setting min bridge out amount"))
		}
	}

	k.SetPauser(ctx, genState.Pauser)
	k.setLastOutflowReset(ctx, genState.LastOutflowReset)

	for _, outflowLimit := range genState.CurrentOutflowLimits {
		k.SetOutflowLimit(ctx, outflowLimit.Token, outflowLimit.Limit)
	}

	for _, outflowAmount := range genState.CurrentOutflowAmounts {
		k.increaseCurrentOutflow(ctx, outflowAmount.Token, outflowAmount.Amount)
	}

	err = k.IncreaseBTCMinted(ctx, genState.InitialBtcSupply)
	if err != nil {
		panic(errorsmod.Wrapf(err, "error setting params"))
	}
}

// ExportGenesis returns the module's exported genesis
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	return &types.GenesisState{
		Params:                         k.GetParams(ctx),
		AssetsLockedSequenceTip:        k.GetAssetsLockedSequenceTip(ctx),
		AssetsUnlockedSequenceTip:      k.GetAssetsUnlockedSequenceTip(ctx),
		SourceBtcToken:                 evmtypes.BytesToHexAddress(k.GetSourceBTCToken(ctx)),
		Erc20TokensMappings:            k.GetERC20TokensMappings(ctx),
		InitialBtcSupply:               k.GetBTCMinted(ctx).Sub(k.GetBTCBurnt(ctx)),
		AssetsUnlockedEvents:           k.GetAllAssetsUnlockedEvents(ctx),
		BitcoinChainMinBridgeOutAmount: k.GetMinBridgeOutAmountForBitcoinChain(ctx),
		TokenMinBridgeOutAmounts:       k.GetAllMinBridgeOutAmount(ctx),
		Pauser:                         k.GetPauser(ctx),
		LastOutflowReset:               k.getLastOutflowReset(ctx),
		CurrentOutflowLimits:           k.GetAllCurrentOutflowLimits(ctx),
		CurrentOutflowAmounts:          k.GetAllCurrentOutflowAmounts(ctx),
	}
}
