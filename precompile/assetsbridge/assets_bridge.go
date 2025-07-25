package assetsbridge

import (
	"context"
	"embed"
	"fmt"
	"math/big"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/mezo-org/mezod/precompile"
	bridgetypes "github.com/mezo-org/mezod/x/bridge/types"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
)

//go:embed abi.json
var filesystem embed.FS

// EvmAddress is the EVM address of the Assets Bridge precompile. The address is
// prefixed with 0x7b7c which was used to derive Mezo chain ID. This prefix is
// used to avoid potential collisions with EVM native precompiles.
const EvmAddress = evmtypes.AssetsBridgePrecompileAddress

// NewPrecompileVersionMap creates a new version map for the assets bridge precompile.
func NewPrecompileVersionMap(
	poaKeeper PoaKeeper,
	bridgeKeeper BridgeKeeper,
	evmKeeper EvmKeeper,
	authzKeeper AuthzKeeper,
) (
	*precompile.VersionMap,
	error,
) {
	// v1 is just BTC observability.
	contractV1, err := NewPrecompile(
		poaKeeper,
		bridgeKeeper,
		evmKeeper,
		authzKeeper,
		&Settings{
			Observability:   true,
			BTCManagement:   false,
			ERC20Management: false,
			SequenceTipView: false,
			BridgeOut:       false,
		},
	)
	if err != nil {
		return nil, err
	}

	// v2 is BTC observability, BTC management, and ERC20 management.
	contractV2, err := NewPrecompile(
		poaKeeper,
		bridgeKeeper,
		evmKeeper,
		authzKeeper,
		&Settings{
			Observability:   true,
			BTCManagement:   true,
			ERC20Management: true,
			SequenceTipView: false,
			BridgeOut:       false,
		},
	)
	if err != nil {
		return nil, err
	}

	// v3 is BTC observability, BTC management, ERC20 management, and sequence tip view
	contractV3, err := NewPrecompile(
		poaKeeper,
		bridgeKeeper,
		evmKeeper,
		authzKeeper,
		&Settings{
			Observability:   true,
			BTCManagement:   true,
			ERC20Management: true,
			SequenceTipView: true,
			BridgeOut:       false,
		},
	)
	if err != nil {
		return nil, err
	}

	// v4 is BTC observability, BTC management, ERC20 management, and sequence tip view,
	// and bridgeOut method implementation
	contractV4, err := NewPrecompile(
		poaKeeper,
		bridgeKeeper,
		evmKeeper,
		authzKeeper,
		&Settings{
			Observability:   true,
			BTCManagement:   true,
			ERC20Management: true,
			SequenceTipView: true,
			BridgeOut:       true,
		},
	)
	if err != nil {
		return nil, err
	}

	return precompile.NewVersionMap(
		map[int]*precompile.Contract{
			0: contractV1, // returning v1 as v0 is legacy to support this precompile before versioning was introduced
			1: contractV1,
			2: contractV2,
			3: contractV3,
			evmtypes.AssetsBridgePrecompileLatestVersion: contractV4,
		},
	), nil
}

type Settings struct {
	Observability   bool // enable methods related to the bridge observability
	BTCManagement   bool // enable methods related to the BTC bridging management
	ERC20Management bool // enable methods related to the ERC20 bridging management
	SequenceTipView bool // enable the method to expose the sequence tip
	BridgeOut       bool // enable the bridgeOut method
}

// NewPrecompile creates a new Assets Bridge precompile.
func NewPrecompile(
	poaKeeper PoaKeeper,
	bridgeKeeper BridgeKeeper,
	evmKeeper EvmKeeper,
	authzKeeper AuthzKeeper,
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

	var methods []precompile.Method

	if settings.Observability {
		methods = append(methods, newBridgeMethod())
	}

	if settings.BTCManagement {
		methods = append(methods, newGetSourceBTCTokenMethod(bridgeKeeper))
	}

	if settings.ERC20Management {
		methods = append(methods, newCreateERC20TokenMappingMethod(poaKeeper, bridgeKeeper))
		methods = append(methods, newDeleteERC20TokenMappingMethod(poaKeeper, bridgeKeeper))
		methods = append(methods, newGetERC20TokenMappingMethod(bridgeKeeper))
		methods = append(methods, newGetERC20TokensMappingsMethod(bridgeKeeper))
		methods = append(methods, newGetMaxERC20TokensMappingsMethod(bridgeKeeper))
	}

	if settings.SequenceTipView {
		methods = append(methods, newGetCurrentSequenceTipMethod(bridgeKeeper))
	}

	if settings.BridgeOut {
		methods = append(methods, newBridgeOutMethod(bridgeKeeper, evmKeeper, authzKeeper))
	}

	contract.RegisterMethods(methods...)

	return contract, nil
}

type AssetsLockedEvent struct {
	SequenceNumber *big.Int       `abi:"sequenceNumber"`
	Recipient      common.Address `abi:"recipient"`
	Amount         *big.Int       `abi:"amount"`
	Token          common.Address `abi:"token"`
}

// PackEventsToInput packs given `AssetsLocked` events into an input of the
// `bridge` function.
func PackEventsToInput(events []AssetsLockedEvent) ([]byte, error) {
	abi, err := precompile.LoadAbiFile(filesystem, "abi.json")
	if err != nil {
		return nil, fmt.Errorf("failed to load ABI file: [%w]", err)
	}

	packedData, err := abi.Pack("bridge", events)
	if err != nil {
		return nil, fmt.Errorf("failed to pack ABI: [%w]", err)
	}

	return packedData, nil
}

type PoaKeeper interface {
	CheckOwner(ctx sdk.Context, sender sdk.AccAddress) error
}

type BridgeKeeper interface {
	GetAssetsLockedSequenceTip(ctx sdk.Context) math.Int
	GetSourceBTCToken(ctx sdk.Context) []byte
	CreateERC20TokenMapping(ctx sdk.Context, sourceToken, mezoToken []byte) error
	DeleteERC20TokenMapping(ctx sdk.Context, sourceToken []byte) error
	GetERC20TokensMappings(ctx sdk.Context) []*bridgetypes.ERC20TokenMapping
	GetERC20TokenMapping(ctx sdk.Context, sourceToken []byte) (*bridgetypes.ERC20TokenMapping, bool)
	GetParams(ctx sdk.Context) bridgetypes.Params
	SaveAssetsUnlocked(
		ctx sdk.Context,
		token []byte,
		amount math.Int,
		chain uint8,
		recipient []byte,
	) (*bridgetypes.AssetsUnlockedEvent, error)
}

type EvmKeeper interface {
	// ExecuteContractCall executes an EVM contract call.
	ExecuteContractCall(
		ctx sdk.Context,
		call evmtypes.ContractCall,
	) (*evmtypes.MsgEthereumTxResponse, error)
	// IsContract returns if the account contains contract code.
	IsContract(ctx sdk.Context, address []byte) bool
}

type BankKeeper interface {
	AllBalances(context.Context, *banktypes.QueryAllBalancesRequest) (*banktypes.QueryAllBalancesResponse, error)
}

type AuthzKeeper interface {
	GetAuthorization(ctx context.Context, grantee, granter sdk.AccAddress, msgType string) (authz.Authorization, *time.Time)
	DispatchActions(ctx context.Context, grantee sdk.AccAddress, msgs []sdk.Msg) ([][]byte, error)
}
