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
const EvmAddress = evmtypes.MaintenancePrecompileAddress

// NewPrecompileVersionMap creates a new version map for the maintenance precompile.
func NewPrecompileVersionMap(
	poaKeeper PoaKeeper,
	evmKeeper EvmKeeper,
) (*precompile.VersionMap, error) {
	// v1 is just the EVM settings.
	contractV1, err := NewPrecompile(poaKeeper, evmKeeper, &Settings{
		EVM:         true,
		Precompiles: false,
	})
	if err != nil {
		return nil, err
	}

	// v2 is the EVM settings and the precompiles settings.
	contractV2, err := NewPrecompile(poaKeeper, evmKeeper, &Settings{
		EVM:         true,
		Precompiles: true,
	})
	if err != nil {
		return nil, err
	}

	return precompile.NewVersionMap(
		map[int]*precompile.Contract{
			0: contractV1, // returning v1 as v0 is legacy to support this precompile before versioning was introduced
			1: contractV1,
			evmtypes.MaintenancePrecompileLatestVersion: contractV2,
		},
	), nil
}

type Settings struct {
	EVM         bool // enable methods related to the evm
	Precompiles bool // enable methods related to the precompiles
}

// NewPrecompile creates a new maintenance precompile.
func NewPrecompile(
	poaKeeper PoaKeeper,
	evmKeeper EvmKeeper,
	settings *Settings,
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

	methods := newPrecompileMethods(poaKeeper, evmKeeper, settings)
	contract.RegisterMethods(methods...)

	return contract, nil
}

// newPrecompileMethods builds the list of methods for the maintenance precompile.
// All methods returned by this function are registered in the maintenance precompile.
func newPrecompileMethods(
	poaKeeper PoaKeeper,
	evmKeeper EvmKeeper,
	settings *Settings,
) []precompile.Method {
	var methods []precompile.Method

	if settings.EVM {
		methods = append(methods, newSetSupportNonEIP155TxsMethod(poaKeeper, evmKeeper))
		methods = append(methods, newGetSupportNonEIP155TxsMethod(evmKeeper))
	}

	if settings.Precompiles {
		methods = append(methods, newSetPrecompileByteCodeMethod(poaKeeper, evmKeeper))
	}

	// TODO: unclear, do we need to add this function behind EVM/Precompiles flag?
	methods = append(methods, newSetFeeChainSplitterAddressMethod(poaKeeper, evmKeeper))
	methods = append(methods, newGetFeeChainSplitterAddressMethod(evmKeeper))

	return methods
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
