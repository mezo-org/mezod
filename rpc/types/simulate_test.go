package types

import (
	"encoding/json"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"

	evmtypes "github.com/mezo-org/mezod/x/evm/types"
)

func TestSimOptsJSONRoundTrip(t *testing.T) {
	from := common.HexToAddress("0x1111111111111111111111111111111111111111")
	to := common.HexToAddress("0x2222222222222222222222222222222222222222")
	number := (*hexutil.Big)(common.Big1)
	time := hexutil.Uint64(100)

	opts := SimOpts{
		BlockStateCalls: []SimBlock{
			{
				BlockOverrides: &BlockOverrides{
					Number: number,
					Time:   &time,
				},
				Calls: []evmtypes.TransactionArgs{
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

	var decoded SimOpts
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

func TestSimCallResultMarshalsEmptyLogsAsArray(t *testing.T) {
	r := SimCallResult{
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

func TestSimCallResultMarshalsPopulatedLogs(t *testing.T) {
	topic := common.HexToHash("0xaaaa")
	r := SimCallResult{
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

func TestSimCallResultErrorOmitEmpty(t *testing.T) {
	r := SimCallResult{Status: hexutil.Uint64(1)}
	data, err := json.Marshal(r)
	require.NoError(t, err)
	require.NotContains(t, string(data), `"error"`)

	r2 := SimCallResult{
		Status: hexutil.Uint64(0),
		Error: &evmtypes.SimError{
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

func TestSimBlockResultMarshalJSON(t *testing.T) {
	r := SimBlockResult{
		Block: map[string]interface{}{
			"number": "0x1",
			"hash":   "0xdeadbeef",
		},
		Calls: []SimCallResult{{Status: hexutil.Uint64(1)}},
	}
	data, err := json.Marshal(r)
	require.NoError(t, err)
	require.Contains(t, string(data), `"number":"0x1"`)
	require.Contains(t, string(data), `"hash":"0xdeadbeef"`)
	require.Contains(t, string(data), `"calls":[`)
	require.Contains(t, string(data), `"logs":[]`)
}

func TestSimBlockResultMarshalEmptyCallsAsArray(t *testing.T) {
	r := SimBlockResult{
		Block: map[string]interface{}{"number": "0x1"},
		Calls: nil,
	}
	data, err := json.Marshal(r)
	require.NoError(t, err)
	require.Contains(t, string(data), `"calls":[]`)
}

func TestSimBlockResultUnmarshalJSON(t *testing.T) {
	raw := []byte(`{"number":"0x5","hash":"0xabcd","calls":[{"returnData":"0x","logs":[],"gasUsed":"0x5208","status":"0x1"}],"extraField":"preserved"}`)
	var r SimBlockResult
	require.NoError(t, json.Unmarshal(raw, &r))
	require.Len(t, r.Calls, 1)
	require.Equal(t, hexutil.Uint64(0x5208), r.Calls[0].GasUsed)
	require.Equal(t, "0x5", r.Block["number"])
	require.Equal(t, "0xabcd", r.Block["hash"])
	require.Equal(t, "preserved", r.Block["extraField"])
}

func TestSimOptsTolerateUnknownFields(t *testing.T) {
	raw := []byte(`{"blockStateCalls":[],"traceTransfers":true,"unknownField":"ignored"}`)
	var opts SimOpts
	require.NoError(t, json.Unmarshal(raw, &opts))
	require.True(t, opts.TraceTransfers)
	require.Empty(t, opts.BlockStateCalls)
}

func TestBlockOverridesAllFields(t *testing.T) {
	number := (*hexutil.Big)(common.Big1)
	difficulty := (*hexutil.Big)(common.Big2)
	time := hexutil.Uint64(200)
	gasLimit := hexutil.Uint64(30000000)
	feeRecipient := common.HexToAddress("0x4444444444444444444444444444444444444444")
	prevRandao := common.HexToHash("0xbeef")
	baseFee := (*hexutil.Big)(common.Big3)
	blobBaseFee := (*hexutil.Big)(common.Big1)
	beaconRoot := common.HexToHash("0xfeed")

	overrides := BlockOverrides{
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

	var decoded BlockOverrides
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
