package validatorpool

import (
	"embed"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v12/precompile"
	poakeeper "github.com/evmos/evmos/v12/x/poa/keeper"
)

//go:embed abi.json
var filesystem embed.FS

// EvmAddress is the EVM address of the validator pool precompile.
// EVM native precompiles reserve the addresses from 0x...01 to 0x...09.
// We use the opposite range (0x1... to 0x9...) for custom Mezo precompiles to
// avoid collisions.
const EvmAddress = "0x2000000000000000000000000000000000000000"

// NewPrecompile creates a new validator pool precompile.
func NewPrecompile(poaKeeper poakeeper.Keeper) (*precompile.Contract, error) {
	contractAbi, err := precompile.LoadAbiFile(filesystem, "abi.json")
	if err != nil {
		return nil, fmt.Errorf("failed to load abi file: [%w]", err)
	}

	contract := precompile.NewContract(
		contractAbi,
		common.HexToAddress(EvmAddress),
	)

	methods := newPrecompileMethods(poaKeeper)
	contract.RegisterMethods(methods...)

	return contract, nil
}

// newPrecompileMethods builds the list of methods for the validator pool precompile.
// All methods returned by this function are registered in the validator pool precompile.
func newPrecompileMethods(_ poakeeper.Keeper) []precompile.Method {
	return []precompile.Method{
		// newSubmitApplicationMethod(poaKeeper),
		// newApproveApplicationMethod(poaKeeper),
		// newKickMethod(poaKeeper),
		// newLeaveMethod(poaKeeper),
		// newTransferOwnershipMethod(poaKeeper),
		// newAcceptOwnershipMethod(poaKeeper),
		// newSlotsMethod(poaKeeper),
	}
}
