package eth

import (
	rpctypes "github.com/mezo-org/mezod/rpc/types"
)

// SimulateV1 simulates a sequence of calls grouped by simulated block.
// Registered as `eth_simulateV1` on the `eth_` namespace.
func (e *PublicAPI) SimulateV1(
	opts rpctypes.SimOpts,
	blockNrOrHash *rpctypes.BlockNumberOrHash,
) ([]*rpctypes.SimBlockResult, error) {
	e.logger.Debug("eth_simulateV1")
	return e.backend.SimulateV1(opts, blockNrOrHash)
}
