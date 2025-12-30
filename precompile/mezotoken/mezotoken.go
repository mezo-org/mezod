package mezotoken

import (
	"embed"
	"fmt"

	"github.com/mezo-org/mezod/utils"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/ethereum/go-ethereum/common"
	"github.com/mezo-org/mezod/precompile"
	"github.com/mezo-org/mezod/precompile/erc20"
	mezotypes "github.com/mezo-org/mezod/types"
	evmkeeper "github.com/mezo-org/mezod/x/evm/keeper"
)

//go:embed abi.json
var filesystem embed.FS

const (
	// EvmAddress is the EVM address of the MEZO token precompile. Token address is
	// prefixed with 0x7b7c which was used to derive Mezo chain ID. This prefix is
	// used to avoid potential collisions with EVM native precompiles.
	EvmAddress = evmtypes.MEZOTokenPrecompileAddress

	Decimals = uint8(18)
	Symbol   = "MEZO"
	Name     = "MEZO"
)

// PoaKeeper defines the expected interface for the POA keeper.
type PoaKeeper interface {
	CheckOwner(ctx sdk.Context, sender sdk.AccAddress) error
}

type Settings struct {
	Minting bool // enable methods related to minting (setMinter, getMinter, mint)
}

// NewPrecompileVersionMap creates a new version map for the MEZO token precompile.
func NewPrecompileVersionMap(
	bankKeeper bankkeeper.Keeper,
	authzkeeper authzkeeper.Keeper,
	evmkeeper evmkeeper.Keeper,
	poaKeeper PoaKeeper,
	id string,
) (*precompile.VersionMap, error) {
	// v1 is the base ERC20 functionality without minting
	contractV1, err := NewPrecompile(bankKeeper, authzkeeper, evmkeeper, poaKeeper, id, &Settings{
		Minting: false,
	})
	if err != nil {
		return nil, err
	}

	// v2 adds minting functionality (setMinter, getMinter, mint)
	contractV2, err := NewPrecompile(bankKeeper, authzkeeper, evmkeeper, poaKeeper, id, &Settings{
		Minting: true,
	})
	if err != nil {
		return nil, err
	}

	return precompile.NewVersionMap(
		map[int]*precompile.Contract{
			1: contractV1,
			evmtypes.MEZOTokenPrecompileLatestVersion: contractV2,
		},
	), nil
}

// NewPrecompile creates a new MEZO token precompile.
func NewPrecompile(
	bankKeeper bankkeeper.Keeper,
	authzkeeper authzkeeper.Keeper,
	evmkeeper evmkeeper.Keeper,
	poaKeeper PoaKeeper,
	id string,
	settings *Settings,
) (*precompile.Contract, error) {
	contractAbi, err := precompile.LoadAbiFile(filesystem, "abi.json")
	if err != nil {
		return nil, fmt.Errorf("failed to load abi file: [%w]", err)
	}

	chainID, err := mezotypes.ParseChainID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to parse chain ID: [%w]", err)
	}

	evmAddress := common.HexToAddress(EvmAddress)
	denom := utils.MezoDenom

	domainSeparator, err := erc20.BuildDomainSeparator(chainID, Name, "1", evmAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to build domain separator: [%w]", err)
	}

	nonceKey := evmtypes.PrecompileMEZONonceKey()

	contract := precompile.NewContract(
		contractAbi,
		evmAddress,
		EvmByteCode,
	)

	methods := newPrecompileMethods(
		bankKeeper,
		authzkeeper,
		evmkeeper,
		poaKeeper,
		denom,
		domainSeparator,
		nonceKey,
		settings,
	)
	contract.RegisterMethods(methods...)

	return contract, nil
}

// newPrecompileMethods builds the list of methods for the MEZO token precompile.
// All methods returned by this function are registered in the MEZO token precompile.
func newPrecompileMethods(
	bankKeeper bankkeeper.Keeper,
	authzkeeper authzkeeper.Keeper,
	evmkeeper evmkeeper.Keeper,
	poaKeeper PoaKeeper,
	denom string,
	domainSeparator []byte,
	nonceKey []byte,
	settings *Settings,
) []precompile.Method {
	methods := []precompile.Method{
		erc20.NewBalanceOfMethod(bankKeeper, denom),
		erc20.NewTotalSupplyMethod(bankKeeper, denom),
		erc20.NewNameMethod(Name),
		erc20.NewSymbolMethod(Symbol),
		erc20.NewDecimalsMethod(Decimals),
		erc20.NewApproveMethod(bankKeeper, authzkeeper, denom),
		erc20.NewTransferMethod(bankKeeper, authzkeeper, evmkeeper, denom),
		erc20.NewTransferFromMethod(bankKeeper, authzkeeper, evmkeeper, denom),
		erc20.NewAllowanceMethod(authzkeeper, denom),
		erc20.NewPermitMethod(bankKeeper, authzkeeper, evmkeeper, denom, domainSeparator, nonceKey),
		erc20.NewNonceMethod(evmkeeper, nonceKey),
		erc20.NewNoncesMethod(evmkeeper, nonceKey),
		erc20.NewDomainSeparatorMethod(domainSeparator),
		erc20.NewPermitTypehashMethod(),
	}

	if settings.Minting {
		methods = append(methods, NewSetMinterMethod(evmkeeper, poaKeeper))
		methods = append(methods, NewGetMinterMethod(evmkeeper))
		methods = append(methods, NewMintMethod(bankKeeper, evmkeeper, denom))
	}

	return methods
}
