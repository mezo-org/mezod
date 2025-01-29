package assetsbridge

import (
	"embed"
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"

	bridgetypes "github.com/mezo-org/mezod/x/bridge/types"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"

	"github.com/ethereum/go-ethereum/common"
	"github.com/mezo-org/mezod/precompile"
)

//go:embed abi.json
var filesystem embed.FS

// EvmAddress is the EVM address of the Assets Bridge precompile. The address is
// prefixed with 0x7b7c which was used to derive Mezo chain ID. This prefix is
// used to avoid potential collisions with EVM native precompiles.
const EvmAddress = evmtypes.AssetsBridgePrecompileAddress

// NewPrecompileVersionMap creates a new version map for the assets bridge precompile.
func NewPrecompileVersionMap(poaKeeper PoaKeeper, bridgeKeeper BridgeKeeper) (
	*precompile.VersionMap,
	error,
) {
	// v1 is just BTC observability.
	contractV1, err := NewPrecompile(
		poaKeeper,
		bridgeKeeper,
		&Settings{
			Observability:   true,
			BTCManagement:   false,
			ERC20Management: false,
		},
	)
	if err != nil {
		return nil, err
	}

	// v2 is BTC observability, BTC management, and ERC20 management.
	contractV2, err := NewPrecompile(
		poaKeeper,
		bridgeKeeper,
		&Settings{
			Observability:   true,
			BTCManagement:   true,
			ERC20Management: true,
		},
	)
	if err != nil {
		return nil, err
	}

	return precompile.NewVersionMap(
		map[int]*precompile.Contract{
			0: contractV1, // returning v1 as v0 is legacy to support this precompile before versioning was introduced
			1: contractV1,
			evmtypes.AssetsBridgePrecompileLatestVersion: contractV2,
		},
	), nil
}

type Settings struct {
	Observability   bool // enable methods related to the bridge observability
	BTCManagement   bool // enable methods related to the BTC bridging management
	ERC20Management bool // enable methods related to the ERC20 bridging management
}

// NewPrecompile creates a new Assets Bridge precompile.
func NewPrecompile(
	poaKeeper PoaKeeper,
	bridgeKeeper BridgeKeeper,
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

	contract.RegisterMethods(methods...)

	return contract, nil
}

type AssetsLockedEvent struct {
	SequenceNumber *big.Int       `abi:"sequenceNumber"`
	Recipient      common.Address `abi:"recipient"`
	TBTCAmount     *big.Int       `abi:"tbtcAmount"`
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
	GetSourceBTCToken(ctx sdk.Context) string
	CreateERC20TokenMapping(ctx sdk.Context, mapping *bridgetypes.ERC20TokenMapping) error
	DeleteERC20TokenMapping(ctx sdk.Context, sourceToken string) error
	GetERC20TokensMappings(ctx sdk.Context) []*bridgetypes.ERC20TokenMapping
	GetERC20TokenMapping(ctx sdk.Context, sourceToken string) (*bridgetypes.ERC20TokenMapping, bool)
	GetParams(ctx sdk.Context) bridgetypes.Params
}
