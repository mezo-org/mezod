//nolint:revive,stylecheck
package v0_4

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/mezo-org/mezod/app/upgrades"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
)

func RunForkLogic(ctx sdk.Context, keepers *upgrades.Keepers) {
	ctx.Logger().Info("running v0.4.0 fork logic")

	k := keepers.EvmKeeper

	ctx.Logger().Info(
		"begin storage root strategy update",
		"currentStrategy",
		k.GetStorageRootStrategy(ctx),
	)

	// The pre-fork strategy on testnet is evmtypes.StorageRootStrategyDummyHash.
	err := k.SetStorageRootStrategy(
		ctx,
		evmtypes.StorageRootStrategyEmptyHash,
	)
	if err != nil {
		// Do not handle consensus in case of failure. Just live with the old strategy.
		ctx.Logger().Error(
			"failed to update storage root strategy",
			"error",
			err,
		)
		return
	}

	ctx.Logger().Info(
		"storage root strategy updated",
		"currentStrategy",
		k.GetStorageRootStrategy(ctx),
	)

	ctx.Logger().Info("v0.4.0 fork logic applied successfully")
}
