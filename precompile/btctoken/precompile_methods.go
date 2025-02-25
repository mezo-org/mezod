//go:build !debugprecompile

package btctoken

import (
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/mezo-org/mezod/precompile"
	evmkeeper "github.com/mezo-org/mezod/x/evm/keeper"
)

func newPrecompileMethods(bankKeeper bankkeeper.Keeper, authzkeeper authzkeeper.Keeper, evmkeeper evmkeeper.Keeper) []precompile.Method {
	return newBasePrecompileMethods(bankKeeper, authzkeeper, evmkeeper)
}
