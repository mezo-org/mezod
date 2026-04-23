package types

import (
	"errors"
	"fmt"
	"testing"

	evmtypes "github.com/mezo-org/mezod/x/evm/types"
	"github.com/stretchr/testify/require"
)

func TestTranslateOverrideError_NilReturnsNil(t *testing.T) {
	require.Nil(t, TranslateOverrideError(nil))
}

func TestTranslateOverrideError_UnknownReturnsNil(t *testing.T) {
	require.Nil(t, TranslateOverrideError(errors.New("something unrelated")))
}

func TestTranslateOverrideError_KindToCode(t *testing.T) {
	tests := []struct {
		name     string
		sentinel error
		wantCode int
	}{
		{
			name:     "self reference maps to -38022",
			sentinel: evmtypes.ErrOverrideMovePrecompileSelfReference,
			wantCode: SimErrCodeMovePrecompileSelfReference,
		},
		{
			name:     "duplicate destination maps to -38023",
			sentinel: evmtypes.ErrOverrideMovePrecompileDuplicateDest,
			wantCode: SimErrCodeMovePrecompileDuplicateDest,
		},
		{
			name:     "state-and-stateDiff conflict maps to -32602",
			sentinel: evmtypes.ErrOverrideStateAndStateDiff,
			wantCode: SimErrCodeInvalidParams,
		},
		{
			name:     "account tainted by precompile maps to -32602",
			sentinel: evmtypes.ErrOverrideAccountTaintedByPrecompile,
			wantCode: SimErrCodeInvalidParams,
		},
		{
			name:     "destination already overridden maps to -32602",
			sentinel: evmtypes.ErrOverrideDestAlreadyOverridden,
			wantCode: SimErrCodeInvalidParams,
		},
		{
			name:     "moving mezo custom precompile maps to -32602",
			sentinel: evmtypes.ErrOverrideMoveMezoCustomPrecompile,
			wantCode: SimErrCodeInvalidParams,
		},
		{
			name:     "not-a-precompile maps to -32602",
			sentinel: evmtypes.ErrOverrideNotAPrecompile,
			wantCode: SimErrCodeInvalidParams,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			wrapped := fmt.Errorf("%w: 0xdeadbeef", tc.sentinel)
			got := TranslateOverrideError(wrapped)
			require.NotNil(t, got)
			require.Equal(t, tc.wantCode, got.Code)
			require.Equal(t, wrapped.Error(), got.Message)
		})
	}
}
