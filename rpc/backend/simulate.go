package backend

import (
	rpctypes "github.com/mezo-org/mezod/rpc/types"
)

// SimulateV1 runs `eth_simulateV1`.
//
// TODO: When the real implementation lands, state-override validation
// failures must be surfaced as spec-reserved JSON-RPC codes (-38022,
// -38023, -32602). The design: the gRPC SimulateV1Response carries a
// structured override_error_kind enum (populated in the keeper from the
// x/evm/types ErrOverrideXxx sentinels); this function translates the
// enum here via rpctypes.TranslateOverrideKind. Error identity cannot
// cross gRPC (status.Error destroys Go types), so the wire contract
// must be explicit data.
func (b *Backend) SimulateV1(
	_ rpctypes.SimOpts,
	_ *rpctypes.BlockNumberOrHash,
) ([]*rpctypes.SimBlockResult, error) {
	return nil, rpctypes.NewSimNotImplementedError()
}
