package backend

import (
	rpctypes "github.com/mezo-org/mezod/rpc/types"
)

// SimulateV1 runs `eth_simulateV1`.
func (b *Backend) SimulateV1(
	_ rpctypes.SimOpts,
	_ *rpctypes.BlockNumberOrHash,
) ([]*rpctypes.SimBlockResult, error) {
	return nil, rpctypes.NewSimNotImplementedError()
}
