package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestJSONRPCErrorImplementsInterfaces(t *testing.T) {
	err := &JSONRPCError{
		Code:    SimErrCodeInsufficientFunds,
		Message: "insufficient funds",
		Data:    "0xdead",
	}
	require.Equal(t, "insufficient funds", err.Error())
	require.Equal(t, SimErrCodeInsufficientFunds, err.ErrorCode())
	require.Equal(t, "0xdead", err.ErrorData())
}

func TestNewSimNotImplementedError(t *testing.T) {
	err := NewSimNotImplementedError()
	require.NotNil(t, err)
	require.Equal(t, SimErrCodeMethodNotFound, err.ErrorCode())
	require.Contains(t, err.Error(), "eth_simulateV1")
}

func TestSimErrorCodeValues(t *testing.T) {
	require.Equal(t, -38010, SimErrCodeNonceTooLow)
	require.Equal(t, -38011, SimErrCodeNonceTooHigh)
	require.Equal(t, -38013, SimErrCodeIntrinsicGas)
	require.Equal(t, -38014, SimErrCodeInsufficientFunds)
	require.Equal(t, -38015, SimErrCodeBlockGasLimitReached)
	require.Equal(t, -38020, SimErrCodeBlockNumberInvalid)
	require.Equal(t, -38021, SimErrCodeBlockTimestampInvalid)
	require.Equal(t, -38022, SimErrCodePrecompileNotMovable)
	require.Equal(t, -38023, SimErrCodeAddressAlreadyOverridden)
	require.Equal(t, -38024, SimErrCodeSenderIsNotEOA)
	require.Equal(t, -38025, SimErrCodeMaxInitCodeSizeExceeded)
	require.Equal(t, -38026, SimErrCodeClientLimitExceeded)
	require.Equal(t, -32000, SimErrCodeReverted)
	require.Equal(t, -32005, SimErrCodeFeeCapTooLow)
	require.Equal(t, -32015, SimErrCodeVMError)
	require.Equal(t, -32601, SimErrCodeMethodNotFound)
	require.Equal(t, -32602, SimErrCodeInvalidParams)
	require.Equal(t, -32603, SimErrCodeInternalError)
}
