package btctoken

import (
	"embed"
	"fmt"
	"math/big"

	evmtypes "github.com/mezo-org/mezod/x/evm/types"

	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/ethereum/go-ethereum/common"
	"github.com/mezo-org/mezod/precompile"
	mezotypes "github.com/mezo-org/mezod/types"
	evmkeeper "github.com/mezo-org/mezod/x/evm/keeper"
)

//go:embed abi.json
var filesystem embed.FS

// EvmAddress is the EVM address of the BTC token precompile. Token address is
// prefixed with 0x7b7c which was used to derive Mezo chain ID. This prefix is
// used to avoid potential collisions with EVM native precompiles.
const EvmAddress = evmtypes.BTCTokenPrecompileAddress

// Parsed chain ID represented as a big integer.
// E.g. mezo_31612-1 is parsed to 31612.
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
			0:                                        contractV1, // returning v1 as v0 is legacy to support this precompile before versioning was introduced
			evmtypes.BTCTokenPrecompileLatestVersion: contractV1,
		},
	), nil
}

// NewPrecompile creates a new BTC token precompile.
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
func newBasePrecompileMethods(bankKeeper bankkeeper.Keeper, authzkeeper authzkeeper.Keeper, evmkeeper evmkeeper.Keeper) []precompile.Method {
	return []precompile.Method{
		newBalanceOfMethod(bankKeeper),
		newTotalSupplyMethod(bankKeeper),
		newNameMethod(),
		newSymbolMethod(),
		newDecimalsMethod(),
		newApproveMethod(bankKeeper, authzkeeper),
		newTransferMethod(bankKeeper, authzkeeper),
		newTransferWithRevertMethod(bankKeeper, authzkeeper),
		newTransferFromMethod(bankKeeper, authzkeeper),
		newAllowanceMethod(authzkeeper),
		newPermitMethod(bankKeeper, authzkeeper, evmkeeper),
		newNonceMethod(evmkeeper),
		newDomainSeparatorMethod(),
		newPermitTypehashMethod(),
	}
}
