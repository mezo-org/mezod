package types

import (
	"errors"

	evmtypes "github.com/mezo-org/mezod/x/evm/types"
)

// TranslateOverrideError maps a state-override validation error returned by
// the x/evm keeper (applyStateOverrides) into a JSON-RPC RPCError carrying
// the spec-reserved error code for eth_simulateV1. The mapping table is:
//
//   - ErrOverrideMovePrecompileSelfReference  → -38022
//   - ErrOverrideMovePrecompileDuplicateDest  → -38023
//   - ErrOverrideStateAndStateDiff            → -32602 (invalid params)
//   - ErrOverrideAccountTaintedByPrecompile   → -32602
//   - ErrOverrideDestAlreadyOverridden        → -32602
//   - ErrOverrideMoveMezoCustomPrecompile     → -32602
//   - ErrOverrideNotAPrecompile               → -32602
//
// Returns nil when err is nil or does not wrap any of the recognized
// override sentinels, leaving the caller to fall back to its default error
// handling (typically an opaque -32000 or the transport's own framing).
func TranslateOverrideError(err error) *RPCError {
	if err == nil {
		return nil
	}

	var code int
	switch {
	case errors.Is(err, evmtypes.ErrOverrideMovePrecompileSelfReference):
		code = SimErrCodeMovePrecompileSelfReference
	case errors.Is(err, evmtypes.ErrOverrideMovePrecompileDuplicateDest):
		code = SimErrCodeMovePrecompileDuplicateDest
	case errors.Is(err, evmtypes.ErrOverrideStateAndStateDiff),
		errors.Is(err, evmtypes.ErrOverrideAccountTaintedByPrecompile),
		errors.Is(err, evmtypes.ErrOverrideDestAlreadyOverridden),
		errors.Is(err, evmtypes.ErrOverrideMoveMezoCustomPrecompile),
		errors.Is(err, evmtypes.ErrOverrideNotAPrecompile):
		code = SimErrCodeInvalidParams
	default:
		return nil
	}

	return &RPCError{Code: code, Message: err.Error()}
}
