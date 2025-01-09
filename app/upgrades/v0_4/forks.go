//nolint:revive,stylecheck
package v0_4

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/holiman/uint256"
	"github.com/mezo-org/mezod/app/upgrades"
	"github.com/mezo-org/mezod/precompile/assetsbridge"
	"github.com/mezo-org/mezod/x/evm/keeper"
	"github.com/mezo-org/mezod/x/evm/statedb"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
)

func RunForkLogic(ctx sdk.Context, keepers *upgrades.Keepers) {
	ctx.Logger().Info("running v0.4.0 fork logic")

	updateStorageRootStrategy(ctx, keepers.EvmKeeper)
	setBridgePrecompileByteCode(ctx, keepers.EvmKeeper)

	ctx.Logger().Info("v0.4.0 fork logic applied successfully")
}

func updateStorageRootStrategy(ctx sdk.Context, evmKeeper *keeper.Keeper) {
	ctx.Logger().Info(
		"begin storage root strategy update",
		"currentStrategy",
		evmKeeper.GetStorageRootStrategy(ctx),
	)

	// The pre-fork strategy on testnet is evmtypes.StorageRootStrategyDummyHash.
	err := evmKeeper.SetStorageRootStrategy(
		ctx,
		evmtypes.StorageRootStrategyEmptyHash,
	)
	if err != nil {
		// Do not halt consensus in case of failure. Just live with the old strategy.
		ctx.Logger().Error(
			"failed to update storage root strategy; abandoning",
			"error",
			err,
		)
		return
	}

	ctx.Logger().Info(
		"storage root strategy updated",
		"currentStrategy",
		evmKeeper.GetStorageRootStrategy(ctx),
	)
}

func setBridgePrecompileByteCode(ctx sdk.Context, evmKeeper *keeper.Keeper) {
	address := common.HexToAddress(assetsbridge.EvmAddress)
	code := common.Hex2Bytes(assetsbridge.EvmByteCode)
	codeHash := crypto.Keccak256Hash(code)

	account := evmKeeper.GetAccount(ctx, address)
	ctx.Logger().Info(
		"begin bridge precompile byte code setting",
		"isContract",
		account != nil && account.IsContract(),
	)

	// Execute only if the precompile is registered.
	if !evmKeeper.IsCustomPrecompile(address) {
		ctx.Logger().Error(
			"bridge precompile not registered; abandoning",
		)
		return
	}

	err := evmKeeper.SetAccount(ctx, address, statedb.Account{
		Nonce:    0,
		Balance:  uint256.NewInt(0),
		CodeHash: codeHash.Bytes(),
	})
	if err != nil {
		// Do not halt consensus in case of failure.
		// Just live without the precompile's byte code.
		ctx.Logger().Error(
			"failed to set precompile byte code; abandoning",
			"error",
			err,
		)
		return
	}

	evmKeeper.SetCode(ctx, codeHash.Bytes(), code)

	account = evmKeeper.GetAccount(ctx, address)
	ctx.Logger().Info(
		"bridge precompile byte code set",
		"isContract",
		account != nil && account.IsContract(),
	)
}
