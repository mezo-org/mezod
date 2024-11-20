package bridge

import (
	"embed"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/mezo-org/mezod/precompile"
)

// TODO: Add the `Bridge` precompile.

//go:embed abi.json
var filesystem embed.FS

// EvmAddress is the EVM address of the Bridge precompile. The address is
// prefixed with 0x7b7c which was used to derive Mezo chain ID. This prefix is
// used to avoid potential collisions with EVM native precompiles.
const EvmAddress = "0x7b7C000000000000000000000000000000000012"

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
