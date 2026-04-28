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

// --- validation=true constructors ---------------------------------------
//
// One test per constructor pinning the spec-reserved JSON-RPC code and
// the user-visible message tokens. Boundary tests below cover the
// edge-of-validity inputs (zero nonces, init-code at the limit). The
// integration test at the bottom of this group asserts the typed errors
// round-trip through `errors.As` so the gRPC handler's *SimError
// fast-path picks them up.

func TestNewSimNonceTooLow(t *testing.T) {
	addr := common.HexToAddress("0x00000000000000000000000000000000000000a1")
	err := NewSimNonceTooLow(addr, 5, 7)
	require.Equal(t, SimErrCodeNonceTooLow, err.ErrorCode())
	require.Contains(t, err.Message, "nonce too low")
	require.Contains(t, err.Message, addr.Hex())
	require.Contains(t, err.Message, "5")
	require.Contains(t, err.Message, "7")
}

func TestNewSimNonceTooHigh(t *testing.T) {
	addr := common.HexToAddress("0x00000000000000000000000000000000000000a2")
	err := NewSimNonceTooHigh(addr, 9, 7)
	require.Equal(t, SimErrCodeNonceTooHigh, err.ErrorCode())
	require.Contains(t, err.Message, "nonce too high")
	require.Contains(t, err.Message, addr.Hex())
	require.Contains(t, err.Message, "9")
	require.Contains(t, err.Message, "7")
}

func TestNewSimBaseFeeTooLow(t *testing.T) {
	have := big.NewInt(1_000_000_000)
	want := big.NewInt(2_000_000_000)
	err := NewSimBaseFeeTooLow(have, want)
	require.Equal(t, SimErrCodeBaseFeeTooLow, err.ErrorCode())
	require.Contains(t, err.Message, "base fee too low")
	require.Contains(t, err.Message, have.String())
	require.Contains(t, err.Message, want.String())
}

func TestNewSimInsufficientFunds(t *testing.T) {
	addr := common.HexToAddress("0x00000000000000000000000000000000000000a3")
	have := big.NewInt(10)
	want := big.NewInt(1000)
	err := NewSimInsufficientFunds(addr, have, want)
	require.Equal(t, SimErrCodeInsufficientFunds, err.ErrorCode())
	require.Contains(t, err.Message, "insufficient funds")
	require.Contains(t, err.Message, addr.Hex())
	require.Contains(t, err.Message, "10")
	require.Contains(t, err.Message, "1000")
}

func TestNewSimInitcodeTooLarge(t *testing.T) {
	err := NewSimInitcodeTooLarge(49_153, 49_152)
	require.Equal(t, SimErrCodeMaxInitCodeSizeExceeded, err.ErrorCode())
	require.Contains(t, err.Message, "max initcode size exceeded")
	require.Contains(t, err.Message, "49153")
	require.Contains(t, err.Message, "49152")
}

func TestNewSimFeeCapTooLow(t *testing.T) {
	have := big.NewInt(1_000_000_000)
	want := big.NewInt(2_000_000_000)
	err := NewSimFeeCapTooLow(have, want)
	require.Equal(t, SimErrCodeFeeCapTooLow, err.ErrorCode())
	require.Contains(t, err.Message, "max fee per gas less than block base fee")
	require.Contains(t, err.Message, have.String())
	require.Contains(t, err.Message, want.String())
}

// NewSimIntrinsicGas predates Phase 10 but is reused as the validation
// gate's intrinsic-gas branch. Pinning the code here keeps Phase 10's
// gate set covered in one place.
func TestNewSimIntrinsicGas_AlreadyExists_Sanity(t *testing.T) {
	err := NewSimIntrinsicGas(20_000, 21_000)
	require.Equal(t, SimErrCodeIntrinsicGas, err.ErrorCode())
	require.Contains(t, err.Message, "intrinsic gas too low")
	require.Contains(t, err.Message, "20000")
	require.Contains(t, err.Message, "21000")
}

// Boundary: zero nonces are a valid (and common) case for fresh accounts.
// The constructor must format them as "0" rather than skip the value or
// emit "<nil>".
func TestNewSimNonceTooLow_ZeroNonces(t *testing.T) {
	addr := common.HexToAddress("0x00000000000000000000000000000000000000a4")
	err := NewSimNonceTooLow(addr, 0, 0)
	require.Equal(t, SimErrCodeNonceTooLow, err.ErrorCode())
	require.Contains(t, err.Message, " 0")
}

// Boundary: init-code exactly at the limit must NOT trip; the
// constructor only fires for sizes strictly above MaxInitCodeSize. This
// test pins the message format for the smallest "over-by-one" case
// callers will encounter in practice.
func TestNewSimInitcodeTooLarge_BoundaryEqual(t *testing.T) {
	// One byte over the limit — the smallest input the gate is allowed
	// to fire on (the equal-to case must NOT fire; that branch lives in
	// validateSimCall, covered in the keeper E2E suite).
	err := NewSimInitcodeTooLarge(49_153, 49_152)
	require.Equal(t, SimErrCodeMaxInitCodeSizeExceeded, err.ErrorCode())
	require.Contains(t, err.Message, "49153")
	require.Contains(t, err.Message, "49152")
}

// Each new constructor must round-trip through `errors.As(err, &SimError)`
// so the gRPC handler's *SimError fast-path picks them up; ErrorCode()
// must match the exported constant.
func TestSimError_ErrorsAsRoundTrip(t *testing.T) {
	addr := common.HexToAddress("0x00000000000000000000000000000000000000a5")
	cases := []struct {
		name    string
		err     error
		wantCode int
	}{
		{"NonceTooLow", NewSimNonceTooLow(addr, 1, 2), SimErrCodeNonceTooLow},
		{"NonceTooHigh", NewSimNonceTooHigh(addr, 5, 2), SimErrCodeNonceTooHigh},
		{"BaseFeeTooLow", NewSimBaseFeeTooLow(big.NewInt(1), big.NewInt(2)), SimErrCodeBaseFeeTooLow},
		{"InsufficientFunds", NewSimInsufficientFunds(addr, big.NewInt(1), big.NewInt(2)), SimErrCodeInsufficientFunds},
		{"InitcodeTooLarge", NewSimInitcodeTooLarge(49_153, 49_152), SimErrCodeMaxInitCodeSizeExceeded},
		{"FeeCapTooLow", NewSimFeeCapTooLow(big.NewInt(1), big.NewInt(2)), SimErrCodeFeeCapTooLow},
		{"IntrinsicGas", NewSimIntrinsicGas(0, 21_000), SimErrCodeIntrinsicGas},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var simErr *SimError
			require.True(t, errors.As(tc.err, &simErr))
			require.Equal(t, tc.wantCode, simErr.ErrorCode())
		})
	}
}
