package testbeds

import (
	"embed"
	"fmt"
	"math/big"

	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/ethereum/go-ethereum/common"
	"github.com/mezo-org/mezod/precompile"
	mezotypes "github.com/mezo-org/mezod/types"
	evmkeeper "github.com/mezo-org/mezod/x/evm/keeper"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
)

//go:embed abi.json
var filesystem embed.FS

const EvmAddress = evmtypes.TestBedStrippedERC20PrecompileAddress

//nolint:unused
var chainID *big.Int

// NewPrecompileVersionMap creates a new version map for the BTC token precompile.
func NewPrecompileVersionMap(
	bankKeeper bankkeeper.Keeper,
	authzkeeper authzkeeper.Keeper,
	evmkeeper evmkeeper.Keeper,
	id string,
) (*precompile.VersionMap, error) {
	contractV1, err := NewPrecompile(bankKeeper, authzkeeper, evmkeeper, id)
	if err != nil {
		return nil, err
	}

	return precompile.NewVersionMap(
		map[int]*precompile.Contract{
			0: contractV1, // returning v1 as v0 is legacy to support this precompile before versioning was introduced
			evmtypes.TestBedStrippedERC20PrecompileLatestVersion: contractV1,
		},
	), nil
}

// NewPrecompile creates a new StrippedERC20 token precompile.
func NewPrecompile(bankKeeper bankkeeper.Keeper, authzkeeper authzkeeper.Keeper, evmkeeper evmkeeper.Keeper, id string) (*precompile.Contract, error) {
	contractAbi, err := precompile.LoadAbiFile(filesystem, "abi.json")
	if err != nil {
		return nil, fmt.Errorf("failed to load abi file: [%w]", err)
	}
	chainID, err = mezotypes.ParseChainID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to parse chain ID: [%w]", err)
	}

	contract := precompile.NewContract(
		contractAbi,
		common.HexToAddress(EvmAddress),
		EvmByteCode,
	)

	methods := newPrecompileMethods(bankKeeper, authzkeeper, evmkeeper)
	contract.RegisterMethods(methods...)

	return contract, nil
}

// newPrecompileMethods builds the list of methods for the BTC token precompile.
// All methods returned by this function are registered in the BTC token precompile.
func newPrecompileMethods(bankKeeper bankkeeper.Keeper, authzkeeper authzkeeper.Keeper, _ evmkeeper.Keeper) []precompile.Method {
	return []precompile.Method{
		newTransferMethod(bankKeeper, authzkeeper),
		newTransferWithRevertMethod(bankKeeper, authzkeeper),
	}
}
