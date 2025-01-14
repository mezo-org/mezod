package maintenance

import (
	"embed"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/mezo-org/mezod/precompile"
	"github.com/mezo-org/mezod/x/evm/statedb"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
)

//go:embed abi.json
var filesystem embed.FS

// EvmAddress is the EVM address of the maintenance precompile. The address is
// prefixed with 0x7b7c which was used to derive Mezo chain ID. This prefix is
// used to avoid potential collisions with EVM native precompiles.
const EvmAddress = "0x7b7c000000000000000000000000000000000013"

// NewPrecompile creates a new maintenance precompile.
func NewPrecompile(
	poaKeeper PoaKeeper,
	evmKeeper EvmKeeper,
) (*precompile.Contract, error) {
	contractAbi, err := precompile.LoadAbiFile(filesystem, "abi.json")
	if err != nil {
		return nil, fmt.Errorf("failed to load abi file: [%w]", err)
	}

	contract := precompile.NewContract(
		contractAbi,
		common.HexToAddress(EvmAddress),
		EvmByteCode,
	)

	methods := newPrecompileMethods(poaKeeper, evmKeeper)
	contract.RegisterMethods(methods...)

	return contract, nil
}

// newPrecompileMethods builds the list of methods for the maintenance precompile.
// All methods returned by this function are registered in the maintenance precompile.
func newPrecompileMethods(
	poaKeeper PoaKeeper,
	evmKeeper EvmKeeper,
) []precompile.Method {
	return []precompile.Method{
		newSetSupportNonEIP155TxsMethod(poaKeeper, evmKeeper),
		newGetSupportNonEIP155TxsMethod(evmKeeper),
		newSetPrecompileByteCodeMethod(poaKeeper, evmKeeper),
	}
}

type PoaKeeper interface {
	CheckOwner(ctx sdk.Context, sender sdk.AccAddress) error
}

type EvmKeeper interface {
	GetParams(ctx sdk.Context) (params evmtypes.Params)
	SetParams(ctx sdk.Context, params evmtypes.Params) error
	SetCode(ctx sdk.Context, codeHash, code []byte)
	GetCode(ctx sdk.Context, codeHash common.Hash) []byte
	IsCustomPrecompile(address common.Address) bool
	GetAccount(ctx sdk.Context, addr common.Address) *statedb.Account
	SetAccount(ctx sdk.Context, addr common.Address, account statedb.Account) error
}
