package btctoken

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
)

//go:embed abi.json
var filesystem embed.FS

// EvmAddress is the EVM address of the BTC token precompile. Token address is
// prefixed with 0x7b7c which was used to derive Mezo chain ID. This prefix is
// used to avoid potential collisions with EVM native precompiles.
const EvmAddress = "0x7b7c000000000000000000000000000000000000"

// EvmByteCode is the EVM bytecode of the BTC token precompile. This code is
// returned by eth_getCode and ensures the precompile address is detected as a
// smart contract by external services. note: It shoould NOT contain a 0x prefix
const EvmByteCode = "7b7c000000000000000000000000000000000000"

// Parsed chain ID represented as a big integer.
// E.g. mezo_31612-1 is parsed to 31612.
var chainID *big.Int

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
	)

	methods := newPrecompileMethods(bankKeeper, authzkeeper, evmkeeper)
	contract.RegisterMethods(methods...)

	return contract, nil
}

// newPrecompileMethods builds the list of methods for the BTC token precompile.
// All methods returned by this function are registered in the BTC token precompile.
func newPrecompileMethods(bankKeeper bankkeeper.Keeper, authzkeeper authzkeeper.Keeper, evmkeeper evmkeeper.Keeper) []precompile.Method {
	return []precompile.Method{
		newBalanceOfMethod(bankKeeper),
		newTotalSupplyMethod(bankKeeper),
		newNameMethod(),
		newSymbolMethod(),
		newDecimalsMethod(),
		newApproveMethod(bankKeeper, authzkeeper),
		newTransferMethod(bankKeeper, authzkeeper),
		newTransferFromMethod(bankKeeper, authzkeeper),
		newAllowanceMethod(authzkeeper),
		newPermitMethod(bankKeeper, authzkeeper, evmkeeper),
		newNonceMethod(evmkeeper),
		newDomainSeparatorMethod(),
		newPermitTypehashMethod(),
	}
}
