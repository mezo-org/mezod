//nolint:revive,stylecheck
package v5_0

import (
	"bytes"
	"context"
	"fmt"
	"slices"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/ethereum/go-ethereum/common"
	"github.com/mezo-org/mezod/app/upgrades"
	"github.com/mezo-org/mezod/utils"
	evmkeeper "github.com/mezo-org/mezod/x/evm/keeper"
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

		sdkCtx.Logger().Info("running v5.0.0 upgrade handler")

		err := updateMEZOPrecompileVersion(sdkCtx, keepers.EvmKeeper)
		if err != nil {
			return nil, fmt.Errorf("failed to update MEZO precompile version: %w", err)
		}

		err = migrateBTCFromLegacyTigrisContracts(sdkCtx, keepers)
		if err != nil {
			return nil, fmt.Errorf("failed to migrate BTC from legacy Tigris contracts: %w", err)
		}

		return mm.RunMigrations(ctx, configurator, fromVM)
	}
}

func updateMEZOPrecompileVersion(ctx sdk.Context, evmKeeper *evmkeeper.Keeper) error {
	params := evmKeeper.GetParams(ctx)

	ctx.Logger().Info(
		"begin MEZO precompile version update",
		"precompilesVersions",
		params.PrecompilesVersions,
	)

	mezoVersionInfoIndex := slices.IndexFunc(
		params.PrecompilesVersions,
		func(versionInfo *evmtypes.PrecompileVersionInfo) bool {
			// Compare bytes just in case to avoid any potential issues
			// with string comparison.
			return bytes.Equal(
				evmtypes.HexAddressToBytes(versionInfo.PrecompileAddress),
				evmtypes.HexAddressToBytes(evmtypes.MEZOTokenPrecompileAddress),
			)
		},
	)

	// 2 is the value of evmtypes.MEZOTokenPrecompileLatestVersion at
	// the time of this upgrade. We avoid using the constant directly
	// as it will change in the future so the actual value of the version
	// used during this upgrade would perish over time.
	params.PrecompilesVersions[mezoVersionInfoIndex].Version = 2

	err := evmKeeper.SetParams(ctx, params)
	if err != nil {
		return err
	}

	ctx.Logger().Info(
		"MEZO precompile version updated",
		"precompilesVersions",
		evmKeeper.GetParams(ctx).PrecompilesVersions,
	)

	return nil
}

func migrateBTCFromLegacyTigrisContracts(
	ctx sdk.Context,
	keepers *upgrades.Keepers,
) error {
	ctx.Logger().Info("begin btc migration from legacy Tigris contracts")

	var legacyVoter, legacyRewardsDistributor, legacyChainFeeBuffer common.Address
	//nolint:gocritic
	if utils.IsTestnet(ctx.ChainID()) {
		legacyVoter = common.HexToAddress("0x72F8dd7F44fFa19E45955aa20A5486E8EB255738")
		legacyRewardsDistributor = common.HexToAddress("0x10B0E7b3411F4A38ca2F6BB697aA28D607924729")
		legacyChainFeeBuffer = common.HexToAddress("0xFb14491168f6BbfFFf9880E936C4D0189e8105dA")
	} else if utils.IsMainnet(ctx.ChainID()) {
		legacyVoter = common.HexToAddress("0x3A4a6919F70e5b0aA32401747C471eCfe2322C1b")
		legacyRewardsDistributor = common.HexToAddress("0x535E01F948458E0b64F9dB2A01Da6F32E240140f")
		legacyChainFeeBuffer = common.HexToAddress("0xf286EA706A2512d2B9232FE7F8b2724880230b45")
	} else {
		ctx.Logger().Info("unknown chain; skipping btc migration from legacy Tigris contracts")
		return nil
	}

	err := migrateBTCBalance(ctx, keepers.BankKeeper, legacyVoter, legacyChainFeeBuffer)
	if err != nil {
		return fmt.Errorf("failed to migrate btc balance from legacy voter: %w", err)
	}

	err = migrateBTCBalance(ctx, keepers.BankKeeper, legacyRewardsDistributor, legacyChainFeeBuffer)
	if err != nil {
		return fmt.Errorf("failed to migrate btc balance from legacy rewards distributor: %w", err)
	}

	ctx.Logger().Info(
		"btc migration from legacy Tigris contracts completed",
		"legacyVoter", legacyVoter.Hex(),
		"legacyRewardsDistributor", legacyRewardsDistributor.Hex(),
		"legacyChainFeeBuffer", legacyChainFeeBuffer.Hex(),
	)

	return nil
}

func migrateBTCBalance(
	ctx sdk.Context,
	bankKeeper bankkeeper.Keeper,
	from, to common.Address,
) error {
	balance := bankKeeper.GetBalance(ctx, sdk.AccAddress(from.Bytes()), utils.BaseDenom)

	if balance.IsZero() {
		ctx.Logger().Info("no btc balance to migrate", "from", from.Hex())
		return nil
	}

	err := bankKeeper.SendCoinsFromAccountToModule(
		ctx,
		sdk.AccAddress(from.Bytes()),
		evmtypes.ModuleName,
		sdk.NewCoins(balance),
	)
	if err != nil {
		return fmt.Errorf("failed to send coins from account to module: %w", err)
	}

	err = bankKeeper.SendCoinsFromModuleToAccount(
		ctx,
		evmtypes.ModuleName,
		sdk.AccAddress(to.Bytes()),
		sdk.NewCoins(balance),
	)
	if err != nil {
		return fmt.Errorf("failed to send coins from module to account: %w", err)
	}

	ctx.Logger().Info(
		"btc balance migrated",
		"from", from.Hex(),
		"to", to.Hex(),
		"balance", balance,
	)

	return nil
}
