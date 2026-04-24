package types

import (
	"errors"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestSimErrorImplementsInterfaces(t *testing.T) {
	err := &SimError{
		Code:    SimErrCodeInsufficientFunds,
		Message: "insufficient funds",
		Data:    "0xdead",
	}
	require.Equal(t, "insufficient funds", err.Error())
	require.Equal(t, SimErrCodeInsufficientFunds, err.ErrorCode())
	require.Equal(t, "0xdead", err.ErrorData())
}

func TestSimErrorErrorDataEmpty(t *testing.T) {
	err := &SimError{Code: SimErrCodeVMError, Message: "stack overflow"}
	require.Nil(t, err.ErrorData())
}

func TestSimErrorIsErrorInterface(t *testing.T) {
	var err error = NewSimInvalidParams("boom")
	var simErr *SimError
	require.True(t, errors.As(err, &simErr))
	require.Equal(t, SimErrCodeInvalidParams, simErr.ErrorCode())
	require.Equal(t, "boom", simErr.Message)
}

func TestSimErrorCodeValues(t *testing.T) {
	require.Equal(t, 3, SimErrCodeReverted)
	require.Equal(t, -32005, SimErrCodeFeeCapTooLow)
	require.Equal(t, -32015, SimErrCodeVMError)
	require.Equal(t, -32016, SimErrCodeTimeout)
	require.Equal(t, -32601, SimErrCodeMethodNotFound)
	require.Equal(t, -32602, SimErrCodeInvalidParams)
	require.Equal(t, -32603, SimErrCodeInternalError)
	require.Equal(t, -38010, SimErrCodeNonceTooLow)
	require.Equal(t, -38011, SimErrCodeNonceTooHigh)
	require.Equal(t, -38012, SimErrCodeBaseFeeTooLow)
	require.Equal(t, -38013, SimErrCodeIntrinsicGas)
	require.Equal(t, -38014, SimErrCodeInsufficientFunds)
	require.Equal(t, -38015, SimErrCodeBlockGasLimitReached)
	require.Equal(t, -38020, SimErrCodeBlockNumberInvalid)
	require.Equal(t, -38021, SimErrCodeBlockTimestampInvalid)
	require.Equal(t, -38022, SimErrCodeMovePrecompileSelfReference)
	require.Equal(t, -38023, SimErrCodeMovePrecompileDuplicateDest)
	require.Equal(t, -38024, SimErrCodeSenderIsNotEOA)
	require.Equal(t, -38025, SimErrCodeMaxInitCodeSizeExceeded)
	require.Equal(t, -38026, SimErrCodeClientLimitExceeded)
}

func TestNewSimInvalidBlockNumber(t *testing.T) {
	err := NewSimInvalidBlockNumber(big.NewInt(5), big.NewInt(10))
	require.Equal(t, SimErrCodeBlockNumberInvalid, err.ErrorCode())
	require.Contains(t, err.Message, "strictly increasing")
	require.Contains(t, err.Message, "5 <= 10")
}

func TestNewSimInvalidBlockTimestamp(t *testing.T) {
	err := NewSimInvalidBlockTimestamp(100, 200)
	require.Equal(t, SimErrCodeBlockTimestampInvalid, err.ErrorCode())
	require.Contains(t, err.Message, "100 <= 200")
}

func TestNewSimClientLimitExceeded(t *testing.T) {
	err := NewSimClientLimitExceeded(big.NewInt(500), 256)
	require.Equal(t, SimErrCodeClientLimitExceeded, err.ErrorCode())
	require.Contains(t, err.Message, "span 500 > max 256")
}

func TestNewSimMovePrecompileSelfRef(t *testing.T) {
	addr := common.HexToAddress("0x0000000000000000000000000000000000000005")
	err := NewSimMovePrecompileSelfRef(addr)
	require.Equal(t, SimErrCodeMovePrecompileSelfReference, err.ErrorCode())
	require.Contains(t, err.Message, "referenced itself")
	require.Contains(t, err.Message, addr.Hex())
}

func TestNewSimMovePrecompileDupDest(t *testing.T) {
	addr := common.HexToAddress("0x0000000000000000000000000000000000001234")
	err := NewSimMovePrecompileDupDest(addr)
	require.Equal(t, SimErrCodeMovePrecompileDuplicateDest, err.ErrorCode())
	require.Contains(t, err.Message, "reference the same destination")
}

func TestInvalidParamsFamilyUsesMinus32602(t *testing.T) {
	addr := common.HexToAddress("0x0000000000000000000000000000000000000042")
	for _, err := range []*SimError{
		NewSimStateAndStateDiff(addr),
		NewSimAccountTainted(addr),
		NewSimDestAlreadyOverridden(addr),
		NewSimMoveMezoCustom(addr),
		NewSimNotAPrecompile(addr),
	} {
		require.Equal(t, SimErrCodeInvalidParams, err.ErrorCode())
		require.Contains(t, err.Message, addr.Hex())
	}
}

func TestNewSimReverted(t *testing.T) {
	t.Run("with data", func(t *testing.T) {
		err := NewSimReverted([]byte{0xde, 0xad, 0xbe, 0xef})
		require.Equal(t, SimErrCodeReverted, err.ErrorCode())
		require.Equal(t, "execution reverted", err.Message)
		require.Equal(t, "0xdeadbeef", err.Data)
		require.Equal(t, "0xdeadbeef", err.ErrorData())
	})
	t.Run("no data", func(t *testing.T) {
		err := NewSimReverted(nil)
		require.Equal(t, "", err.Data)
		require.Nil(t, err.ErrorData())
	})
}

func TestNewSimVMError(t *testing.T) {
	err := NewSimVMError("out of gas")
	require.Equal(t, SimErrCodeVMError, err.ErrorCode())
	require.Equal(t, "out of gas", err.Message)
}
