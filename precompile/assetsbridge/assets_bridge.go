package assetsbridge

import (
	"embed"
	"fmt"
	"math/big"

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
func NewPrecompileVersionMap() (*precompile.VersionMap, error) {
	contractV1, err := NewPrecompile()
	if err != nil {
		return nil, err
	}

	return precompile.NewVersionMap(
		map[int]*precompile.Contract{
			0: contractV1, // returning v1 as v0 is legacy to support this precompile before versioning was introduced
			evmtypes.AssetsBridgePrecompileLatestVersion: contractV1,
		},
	), nil
}

// NewPrecompile creates a new Assets Bridge precompile.
func NewPrecompile() (*precompile.Contract, error) {
	contractAbi, err := precompile.LoadAbiFile(filesystem, "abi.json")
	if err != nil {
		return nil, fmt.Errorf("failed to load abi file: [%w]", err)
	}

	contract := precompile.NewContract(
		contractAbi,
		common.HexToAddress(EvmAddress),
		EvmByteCode,
	)

	contract.RegisterMethods(newBridgeMethod())

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
