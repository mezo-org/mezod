//go:build debugprecompile

package btctoken

import (
	"embed"

	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/mezo-org/mezod/precompile"
	evmkeeper "github.com/mezo-org/mezod/x/evm/keeper"
)

//go:embed abi_debug.json
var filesystem embed.FS

const filePath = "abi_debug.json"

func newPrecompileMethods(bankKeeper bankkeeper.Keeper, authzkeeper authzkeeper.Keeper, evmkeeper evmkeeper.Keeper) []precompile.Method {
	return append(
		newBasePrecompileMethods(bankKeeper, authzkeeper, evmkeeper),
		newTransferWithRevertMethod(bankKeeper, authzkeeper),
	)
}
