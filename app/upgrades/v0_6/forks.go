//nolint:revive,stylecheck
package v0_6

import (
	"bytes"
	"slices"

	sdkmath "cosmossdk.io/math"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	bridgetypes "github.com/mezo-org/mezod/x/bridge/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/mezo-org/mezod/app/upgrades"
	bridgekeeper "github.com/mezo-org/mezod/x/bridge/keeper"
	"github.com/mezo-org/mezod/x/evm/keeper"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
)

func RunForkLogic(ctx sdk.Context, keepers *upgrades.Keepers) {
	ctx.Logger().Info("running v0.6.0 fork logic")

	updateBridgePrecompileVersion(ctx, keepers.EvmKeeper)
	enableERC20Bridge(ctx, keepers.BridgeKeeper)
	enableBTCSupplyAssertion(ctx, keepers.BridgeKeeper, keepers.BankKeeper)

	ctx.Logger().Info("v0.6.0 fork logic applied successfully")
}

func updateBridgePrecompileVersion(ctx sdk.Context, evmKeeper *keeper.Keeper) {
	params := evmKeeper.GetParams(ctx)

	ctx.Logger().Info(
		"begin bridge precompile version update",
		"precompilesVersions",
		params.PrecompilesVersions,
	)

	bridgeVersionInfoIndex := slices.IndexFunc(
		params.PrecompilesVersions,
		func(versionInfo *evmtypes.PrecompileVersionInfo) bool {
			// Compare bytes just in case to avoid any potential issues
			// with string comparison.
			return bytes.Equal(
				evmtypes.HexAddressToBytes(versionInfo.PrecompileAddress),
				evmtypes.HexAddressToBytes(evmtypes.AssetsBridgePrecompileAddress),
			)
		},
	)

	// 2 is the value of evmtypes.AssetsBridgePrecompileLatestVersion at
	// the time of this upgrade. We avoid using the constant directly to
	// as it will change in the future so the actual value of the version
	// used during this upgrade would perish over time.
	params.PrecompilesVersions[bridgeVersionInfoIndex].Version = 2

	err := evmKeeper.SetParams(ctx, params)
	if err != nil {
		// Do not halt consensus in case of failure.
		// Just live without the Bridge precompile version update.
		ctx.Logger().Error(
			"failed to set bridge precompile version; abandoning",
			"error",
			err,
		)
		return
	}

	ctx.Logger().Info(
		"bridge precompile version updated",
		"precompilesVersions",
		evmKeeper.GetParams(ctx).PrecompilesVersions,
	)
}

func enableERC20Bridge(ctx sdk.Context, bridgeKeeper bridgekeeper.Keeper) {
	params := bridgeKeeper.GetParams(ctx)

	ctx.Logger().Info(
		"enabling ERC20 bridge",
		"maxErc20TokensMappings", params.MaxErc20TokensMappings,
		"sourceBTCToken", evmtypes.BytesToHexAddress(bridgeKeeper.GetSourceBTCToken(ctx)),
	)

	params.MaxErc20TokensMappings = bridgetypes.DefaultMaxERC20TokensMappings
	err := bridgeKeeper.SetParams(ctx, params)
	if err != nil {
		// Do not halt consensus in case of failure.
		// Just live without the mapping limit.
		ctx.Logger().Error(
			"failed to set max ERC20 token mappings count; abandoning",
			"error",
			err,
		)
		return
	}

	sepoliaTBTC := "0x517f2982701695D4E52f1ECFBEf3ba31Df470161"
	bridgeKeeper.SetSourceBTCToken(ctx, evmtypes.HexAddressToBytes(sepoliaTBTC))

	ctx.Logger().Info(
		"ERC20 bridge enabled",
		"maxErc20TokensMappings", bridgeKeeper.GetParams(ctx).MaxErc20TokensMappings,
		"sourceBTCToken", evmtypes.BytesToHexAddress(bridgeKeeper.GetSourceBTCToken(ctx)),
	)
}

func enableBTCSupplyAssertion(
	ctx sdk.Context,
	bridgeKeeper bridgekeeper.Keeper,
	bankKeeper bankkeeper.Keeper,
) {
	params := bridgeKeeper.GetParams(ctx)
	btcSupply := bankKeeper.GetSupply(ctx, evmtypes.DefaultEVMDenom)

	ctx.Logger().Info(
		"enabling BTC supply assertion",
		"btcSupplyAssertionEnabled", params.BtcSupplyAssertionEnabled,
		"btcMinted", bridgeKeeper.GetBTCMinted(ctx),
		"btcBurnt", bridgeKeeper.GetBTCBurnt(ctx),
		"btcSupply", btcSupply.Amount,
	)

	err := bridgeKeeper.IncreaseBTCMinted(ctx, btcSupply.Amount)
	if err != nil {
		// Do not halt consensus in case of failure.
		// Just live without the counter.
		ctx.Logger().Error(
			"failed to increase BTC minted counter; abandoning",
			"error",
			err,
		)
		return
	}

	// Just for sanity, set the burnt counter to 0.
	err = bridgeKeeper.IncreaseBTCBurnt(ctx, sdkmath.NewInt(0))
	if err != nil {
		// Do not halt consensus in case of failure.
		// Just live without the counter.
		ctx.Logger().Error(
			"failed to increase BTC burnt counter; abandoning",
			"error",
			err,
		)
		return
	}

	params.BtcSupplyAssertionEnabled = true
	err = bridgeKeeper.SetParams(ctx, params)
	if err != nil {
		// Do not halt consensus in case of failure.
		// Just live without the assertion.
		ctx.Logger().Error(
			"failed to set BTC supply assertion flag; abandoning",
			"error",
			err,
		)
		return
	}

	ctx.Logger().Info(
		"BTC supply assertion enabled",
		"btcSupplyAssertionEnabled", bridgeKeeper.GetParams(ctx).BtcSupplyAssertionEnabled,
		"btcMinted", bridgeKeeper.GetBTCMinted(ctx),
		"btcBurnt", bridgeKeeper.GetBTCBurnt(ctx),
		"btcSupply", btcSupply.Amount,
	)
}
