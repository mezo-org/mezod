//nolint:revive,stylecheck
package v2_0

import (
	"context"

	sdkmath "cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/mezo-org/mezod/app/upgrades"
	"github.com/mezo-org/mezod/utils"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	keepers *upgrades.Keepers,
) upgradetypes.UpgradeHandler {
	return func(
		ctx context.Context,
		_ upgradetypes.Plan,
		fromVM module.VersionMap,
	) (module.VersionMap, error) {
		sdkCtx := sdk.UnwrapSDKContext(ctx)

		sdkCtx.Logger().Info("running v2.0.0 upgrade handler")

		// For mainnet, there are no particular state upgrades that must be executed globally during this upgrade.
		// The RunMigrations call may still execute module-specific state upgrades. This is not the
		// case for this upgrade but it's a Cosmos SDK requirement to call it here.

		// For testnet, mint the MEZO supply. This is needed because testnet, unlike mainnet,
		// did not have the MEZO supply minted at genesis. Moreover, testnet MEZO supply
		// was not minted immediately after the MEZO token was introduced in v1.0.0. Now is
		// the time to do so with the v2.0.0 testnet upgrade.
		if utils.IsTestnet(sdkCtx.ChainID()) {
			mezoSupply, _ := sdkmath.NewIntFromString("1000000000000000000000000000") // 1B MEZO
			coins := sdk.NewCoins(sdk.NewCoin(utils.MezoDenom, mezoSupply))

			// Mint the MEZO supply to the EVM module.
			err := keepers.BankKeeper.MintCoins(
				ctx,
				evmtypes.ModuleName,
				coins,
			)
			if err != nil {
				return nil, err
			}

			// Transfer the MEZO supply to the PoA owner account.
			err = keepers.BankKeeper.SendCoinsFromModuleToAccount(
				ctx,
				evmtypes.ModuleName,
				keepers.PoaKeeper.GetOwner(sdkCtx),
				coins,
			)
			if err != nil {
				return nil, err
			}

			sdkCtx.Logger().Info("Testnet MEZO supply minted and transferred to the PoA owner account")

			params := keepers.EvmKeeper.GetParams(sdkCtx)
			// 1 is the value of evmtypes.MEZOTokenPrecompileLatestVersion at
			// the time of this upgrade. We avoid using the constant directly
			// as it will change in the future so the actual value of the version
			// used during this upgrade would perish over time.
			params.PrecompilesVersions = append(
				params.PrecompilesVersions,
				&evmtypes.PrecompileVersionInfo{
					PrecompileAddress: evmtypes.MEZOTokenPrecompileAddress,
					Version:           1,
				},
			)
			err = keepers.EvmKeeper.SetParams(sdkCtx, params)
			if err != nil {
				return nil, err
			}

			sdkCtx.Logger().Info("Testnet MEZO token precompile enabled")
		}

		return mm.RunMigrations(ctx, configurator, fromVM)
	}
}
