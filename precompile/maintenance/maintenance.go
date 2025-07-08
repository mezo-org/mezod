package maintenance

import (
	"embed"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/mezo-org/mezod/precompile"
	"github.com/mezo-org/mezod/x/evm/statedb"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
	feemarkettypes "github.com/mezo-org/mezod/x/feemarket/types"
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
	feeMarketKeeper FeeMarketKeeper,
) (*precompile.VersionMap, error) {
	// v1 is just the EVM settings.
	contractV1, err := NewPrecompile(poaKeeper, evmKeeper, feeMarketKeeper, &Settings{
		EVM:              true,
		Precompiles:      false,
		ChainFeeSplitter: false,
		GasPrice:         false,
	})
	if err != nil {
		return nil, err
	}

	// v2 is the EVM settings and the precompiles settings.
	contractV2, err := NewPrecompile(poaKeeper, evmKeeper, feeMarketKeeper, &Settings{
		EVM:              true,
		Precompiles:      true,
		ChainFeeSplitter: false,
		GasPrice:         false,
	})
	if err != nil {
		return nil, err
	}

	// v3 is the EVM settings, the precompiles settings and the chain fee splitter settings.
	contractV3, err := NewPrecompile(poaKeeper, evmKeeper, feeMarketKeeper, &Settings{
		EVM:              true,
		Precompiles:      true,
		ChainFeeSplitter: true,
		GasPrice:         false,
	})
	if err != nil {
		return nil, err
	}

	// v4 is the EVM settings, the precompiles settings, the chain fee splitter settings,
	// and the gas price settings.
	contractV4, err := NewPrecompile(poaKeeper, evmKeeper, feeMarketKeeper, &Settings{
		EVM:              true,
		Precompiles:      true,
		ChainFeeSplitter: true,
		GasPrice:         true,
	})
	if err != nil {
		return nil, err
	}

	return precompile.NewVersionMap(
		map[int]*precompile.Contract{
			0: contractV1, // returning v1 as v0 is legacy to support this precompile before versioning was introduced
			1: contractV1,
			2: contractV2,
			3: contractV3,
			evmtypes.MaintenancePrecompileLatestVersion: contractV4,
		},
	), nil
}

type Settings struct {
	EVM              bool // enable methods related to the evm
	Precompiles      bool // enable methods related to the precompiles
	ChainFeeSplitter bool // enable methods related to the chain fee splitter
	GasPrice         bool // enable methods related to the gas price
}

// NewPrecompile creates a new maintenance precompile.
func NewPrecompile(
	poaKeeper PoaKeeper,
	evmKeeper EvmKeeper,
	feeMarketKeeper FeeMarketKeeper,
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

	methods := newPrecompileMethods(poaKeeper, evmKeeper, feeMarketKeeper, settings)
	contract.RegisterMethods(methods...)

	return contract, nil
}

// newPrecompileMethods builds the list of methods for the maintenance precompile.
// All methods returned by this function are registered in the maintenance precompile.
func newPrecompileMethods(
	poaKeeper PoaKeeper,
	evmKeeper EvmKeeper,
	feeMarketKeeper FeeMarketKeeper,
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

	if settings.ChainFeeSplitter {
		methods = append(methods, newSetChainFeeSplitterAddressMethod(poaKeeper, evmKeeper))
		methods = append(methods, newGetChainFeeSplitterAddressMethod(evmKeeper))
	}

	if settings.GasPrice {
		methods = append(methods, newSetMinGasPriceMethod(poaKeeper, feeMarketKeeper))
		methods = append(methods, newGetMinGasPriceMethod(feeMarketKeeper))
	}

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

type FeeMarketKeeper interface {
	GetParams(ctx sdk.Context) (params feemarkettypes.Params)
	SetParams(ctx sdk.Context, params feemarkettypes.Params) error
}
