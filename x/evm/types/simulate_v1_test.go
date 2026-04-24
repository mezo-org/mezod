package types_test

import (
	"encoding/json"
	"testing"

	"github.com/ethereum/go-ethereum/common"
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

func TestSimBlockResult_UnmarshalJSON(t *testing.T) {
	raw := []byte(`{"number":"0x5","hash":"0xabcd","calls":[{"returnData":"0x","logs":[],"gasUsed":"0x5208","status":"0x1"}],"extraField":"preserved"}`)
	var r types.SimBlockResult
	require.NoError(t, json.Unmarshal(raw, &r))
	require.Len(t, r.Calls, 1)
	require.Equal(t, hexutil.Uint64(0x5208), r.Calls[0].GasUsed)
	require.Equal(t, "0x5", r.Block["number"])
	require.Equal(t, "0xabcd", r.Block["hash"])
	require.Equal(t, "preserved", r.Block["extraField"])
}

func TestSimCallResult_MarshalsEmptyLogsAsArray(t *testing.T) {
	r := types.SimCallResult{
		ReturnValue: hexutil.Bytes{0xde, 0xad},
		GasUsed:     hexutil.Uint64(21000),
		Status:      hexutil.Uint64(1),
		Logs:        nil,
	}
	data, err := json.Marshal(r)
	require.NoError(t, err)
	require.Contains(t, string(data), `"logs":[]`)
	require.NotContains(t, string(data), `"logs":null`)
}

func TestSimCallResult_MarshalsPopulatedLogs(t *testing.T) {
	topic := common.HexToHash("0xaaaa")
	r := types.SimCallResult{
		Logs: []*ethtypes.Log{{
			Address: common.HexToAddress("0x3333333333333333333333333333333333333333"),
			Topics:  []common.Hash{topic},
		}},
		Status: hexutil.Uint64(1),
	}
	data, err := json.Marshal(r)
	require.NoError(t, err)
	require.Contains(t, string(data), `"address":"0x3333333333333333333333333333333333333333"`)
}

func TestSimCallResult_ErrorOmitEmpty(t *testing.T) {
	r := types.SimCallResult{Status: hexutil.Uint64(1)}
	data, err := json.Marshal(r)
	require.NoError(t, err)
	require.NotContains(t, string(data), `"error"`)

	r2 := types.SimCallResult{
		Status: hexutil.Uint64(0),
		Error: &types.SimError{
			Message: "boom",
			Code:    -32000,
			Data:    "0xdead",
		},
	}
	data2, err := json.Marshal(r2)
	require.NoError(t, err)
	require.Contains(t, string(data2), `"error"`)
	require.Contains(t, string(data2), `"code":-32000`)
	require.Contains(t, string(data2), `"data":"0xdead"`)
}

func TestSimOpts_JSONRoundTrip(t *testing.T) {
	from := common.HexToAddress("0x1111111111111111111111111111111111111111")
	to := common.HexToAddress("0x2222222222222222222222222222222222222222")
	number := (*hexutil.Big)(common.Big1)
	time := hexutil.Uint64(100)

	opts := types.SimOpts{
		BlockStateCalls: []types.SimBlock{
			{
				BlockOverrides: &types.SimBlockOverrides{
					Number: number,
					Time:   &time,
				},
				Calls: []types.TransactionArgs{
					{From: &from, To: &to},
				},
			},
		},
		TraceTransfers:         true,
		Validation:             false,
		ReturnFullTransactions: true,
	}

	data, err := json.Marshal(opts)
	require.NoError(t, err)

	var decoded types.SimOpts
	require.NoError(t, json.Unmarshal(data, &decoded))
	require.Len(t, decoded.BlockStateCalls, 1)
	require.True(t, decoded.TraceTransfers)
	require.True(t, decoded.ReturnFullTransactions)
	require.False(t, decoded.Validation)
	require.NotNil(t, decoded.BlockStateCalls[0].BlockOverrides)
	require.Equal(t, "0x1", decoded.BlockStateCalls[0].BlockOverrides.Number.String())
	require.Equal(t, hexutil.Uint64(100), *decoded.BlockStateCalls[0].BlockOverrides.Time)
	require.Equal(t, from, *decoded.BlockStateCalls[0].Calls[0].From)
}

func TestSimBlockOverrides_AllFields(t *testing.T) {
	number := (*hexutil.Big)(common.Big1)
	difficulty := (*hexutil.Big)(common.Big2)
	time := hexutil.Uint64(200)
	gasLimit := hexutil.Uint64(30000000)
	feeRecipient := common.HexToAddress("0x4444444444444444444444444444444444444444")
	prevRandao := common.HexToHash("0xbeef")
	baseFee := (*hexutil.Big)(common.Big3)
	blobBaseFee := (*hexutil.Big)(common.Big1)
	beaconRoot := common.HexToHash("0xfeed")

	overrides := types.SimBlockOverrides{
		Number:        number,
		Difficulty:    difficulty,
		Time:          &time,
		GasLimit:      &gasLimit,
		FeeRecipient:  &feeRecipient,
		PrevRandao:    &prevRandao,
		BaseFeePerGas: baseFee,
		BlobBaseFee:   blobBaseFee,
		BeaconRoot:    &beaconRoot,
	}
	data, err := json.Marshal(overrides)
	require.NoError(t, err)

	var decoded types.SimBlockOverrides
	require.NoError(t, json.Unmarshal(data, &decoded))
	require.Equal(t, number.String(), decoded.Number.String())
	require.Equal(t, difficulty.String(), decoded.Difficulty.String())
	require.Equal(t, time, *decoded.Time)
	require.Equal(t, gasLimit, *decoded.GasLimit)
	require.Equal(t, feeRecipient, *decoded.FeeRecipient)
	require.Equal(t, prevRandao, *decoded.PrevRandao)
	require.Equal(t, baseFee.String(), decoded.BaseFeePerGas.String())
	require.Equal(t, blobBaseFee.String(), decoded.BlobBaseFee.String())
	require.Equal(t, beaconRoot, *decoded.BeaconRoot)
}
