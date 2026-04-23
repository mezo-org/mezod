package types_test

import (
	"encoding/json"
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/stretchr/testify/require"

	"github.com/mezo-org/mezod/x/evm/types"
)

func TestUnmarshalSimOpts_EmptyIsValid(t *testing.T) {
	opts, err := types.UnmarshalSimOpts([]byte(`{}`))
	require.NoError(t, err)
	require.NotNil(t, opts)
	require.Empty(t, opts.BlockStateCalls)
}

func TestUnmarshalSimOpts_UnknownFieldsTolerated(t *testing.T) {
	opts, err := types.UnmarshalSimOpts([]byte(`{"blockStateCalls":[],"futureFlag":true}`))
	require.NoError(t, err)
	require.NotNil(t, opts)
}

func TestUnmarshalSimOpts_RejectsBeaconRootOverride(t *testing.T) {
	data := []byte(`{"blockStateCalls":[{"blockOverrides":{"beaconRoot":"0x0000000000000000000000000000000000000000000000000000000000000001"},"calls":[]}]}`)
	_, err := types.UnmarshalSimOpts(data)
	require.Error(t, err)
	require.Contains(t, err.Error(), "BeaconRoot")
}

func TestUnmarshalSimOpts_RejectsWithdrawalsOverride(t *testing.T) {
	data := []byte(`{"blockStateCalls":[{"blockOverrides":{"withdrawals":[]},"calls":[]}]}`)
	_, err := types.UnmarshalSimOpts(data)
	require.Error(t, err)
	require.Contains(t, err.Error(), "Withdrawals")
}

func TestUnmarshalSimOpts_RejectsBlobBaseFee(t *testing.T) {
	data := []byte(`{"blockStateCalls":[{"blockOverrides":{"blobBaseFee":"0x1"},"calls":[]}]}`)
	_, err := types.UnmarshalSimOpts(data)
	require.Error(t, err)
	require.Contains(t, err.Error(), "BlobBaseFee")
}

func TestUnmarshalSimOpts_InvalidJSON(t *testing.T) {
	_, err := types.UnmarshalSimOpts([]byte(`{not json`))
	require.Error(t, err)
}

func TestBuildSimCallResult_Success(t *testing.T) {
	res := &types.MsgEthereumTxResponse{
		Ret:     []byte{0xde, 0xad, 0xbe, 0xef},
		GasUsed: 21_000,
	}
	out := types.BuildSimCallResult(res)
	require.Equal(t, hexutil.Uint64(1), out.Status)
	require.Equal(t, hexutil.Uint64(21_000), out.GasUsed)
	require.Equal(t, hexutil.Bytes{0xde, 0xad, 0xbe, 0xef}, out.ReturnValue)
	require.Nil(t, out.Error)
	require.NotNil(t, out.Logs, "Logs must default to empty slice, never nil")
	require.Empty(t, out.Logs)
}

func TestBuildSimCallResult_RevertWithData(t *testing.T) {
	revertData := []byte{0x08, 0xc3, 0x79, 0xa0}
	res := &types.MsgEthereumTxResponse{
		Ret:     revertData,
		GasUsed: 12_345,
		VmError: vm.ErrExecutionReverted.Error(),
	}
	out := types.BuildSimCallResult(res)
	require.Equal(t, hexutil.Uint64(0), out.Status)
	require.Equal(t, hexutil.Uint64(12_345), out.GasUsed)
	require.NotNil(t, out.Error)
	require.Equal(t, types.SimErrCodeReverted, out.Error.ErrorCode())
	require.Equal(t, "execution reverted", out.Error.Message)
	require.Equal(t, hexutil.Encode(revertData), out.Error.Data)
}

func TestBuildSimCallResult_VMError(t *testing.T) {
	res := &types.MsgEthereumTxResponse{
		GasUsed: 30_000,
		VmError: "out of gas",
	}
	out := types.BuildSimCallResult(res)
	require.Equal(t, hexutil.Uint64(0), out.Status)
	require.NotNil(t, out.Error)
	require.Equal(t, types.SimErrCodeVMError, out.Error.ErrorCode())
	require.Equal(t, "out of gas", out.Error.Message)
	require.Empty(t, out.Error.Data)
}

func TestSimBlockResult_MarshalJSON_FlattensBlockFields(t *testing.T) {
	r := types.SimBlockResult{
		Block: map[string]interface{}{
			"number":    "0x1",
			"hash":      "0xabc",
			"gasUsed":   "0x5208",
			"timestamp": "0x64",
		},
		Calls: []types.SimCallResult{{
			Status:  hexutil.Uint64(1),
			GasUsed: hexutil.Uint64(21_000),
			Logs:    []*ethtypes.Log{},
		}},
	}
	data, err := json.Marshal(r)
	require.NoError(t, err)

	var decoded map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &decoded))
	require.Equal(t, "0x1", decoded["number"], "block fields must appear at top level")
	require.Equal(t, "0xabc", decoded["hash"])
	require.Equal(t, "0x5208", decoded["gasUsed"])
	require.Contains(t, decoded, "calls")
	calls, ok := decoded["calls"].([]interface{})
	require.True(t, ok)
	require.Len(t, calls, 1)
}

func TestSimBlockResult_MarshalJSON_NilCallsBecomesEmptySlice(t *testing.T) {
	r := types.SimBlockResult{
		Block: map[string]interface{}{"number": "0x1"},
		Calls: nil,
	}
	data, err := json.Marshal(r)
	require.NoError(t, err)

	var decoded map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &decoded))
	calls, ok := decoded["calls"].([]interface{})
	require.True(t, ok, "nil Calls must marshal as an empty array, not null")
	require.Empty(t, calls)
}
