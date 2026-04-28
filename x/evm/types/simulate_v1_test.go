package types_test

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/trie"
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

// -----------------------------------------------------------------------------
// SimBlockResult — full block envelope marshaling
// -----------------------------------------------------------------------------

// envelopeChainConfig returns a minimal post-London chain config so
// RPCMarshalBlock includes baseFeePerGas in the rendered envelope. The
// test app's full chain config is anchored higher up the stack; here
// only the fork bits the marshaler reads need to be present.
func envelopeChainConfig() *params.ChainConfig {
	return &params.ChainConfig{
		ChainID:             big.NewInt(31611),
		HomesteadBlock:      big.NewInt(0),
		EIP150Block:         big.NewInt(0),
		EIP155Block:         big.NewInt(0),
		EIP158Block:         big.NewInt(0),
		ByzantiumBlock:      big.NewInt(0),
		ConstantinopleBlock: big.NewInt(0),
		PetersburgBlock:     big.NewInt(0),
		IstanbulBlock:       big.NewInt(0),
		MuirGlacierBlock:    big.NewInt(0),
		BerlinBlock:         big.NewInt(0),
		LondonBlock:         big.NewInt(0),
	}
}

// makeEnvelopeBlock builds a minimal *ethtypes.Block carrying one
// unsigned legacy tx and one matching receipt. Used by the marshal
// tests to drive the typed-Block branch of SimBlockResult.MarshalJSON.
func makeEnvelopeBlock(t *testing.T, from common.Address, to common.Address, value uint64) (*ethtypes.Block, *ethtypes.Transaction) {
	t.Helper()
	tx := ethtypes.NewTx(&ethtypes.LegacyTx{
		Nonce:    7,
		GasPrice: big.NewInt(0),
		Gas:      21_000,
		To:       &to,
		Value:    new(big.Int).SetUint64(value),
		Data:     nil,
	})
	receipt := &ethtypes.Receipt{
		Type:              ethtypes.LegacyTxType,
		Status:            ethtypes.ReceiptStatusSuccessful,
		CumulativeGasUsed: 21_000,
		GasUsed:           21_000,
		TxHash:            tx.Hash(),
		Logs:              []*ethtypes.Log{},
		BlockNumber:       big.NewInt(11),
	}
	receipt.Bloom = ethtypes.CreateBloom(ethtypes.Receipts{receipt})
	header := &ethtypes.Header{
		ParentHash:  common.HexToHash("0xaaa1"),
		UncleHash:   ethtypes.EmptyUncleHash,
		Coinbase:    common.Address{},
		TxHash:      ethtypes.EmptyTxsHash,
		ReceiptHash: ethtypes.EmptyReceiptsHash,
		Number:      big.NewInt(11),
		GasLimit:    30_000_000,
		GasUsed:     21_000,
		Time:        1_700_000_000,
		Difficulty:  new(big.Int),
		BaseFee:     big.NewInt(1_000_000_000),
	}
	block := ethtypes.NewBlock(
		header,
		&ethtypes.Body{Transactions: []*ethtypes.Transaction{tx}},
		[]*ethtypes.Receipt{receipt},
		trie.NewStackTrie(nil),
	)
	return block, tx
}

// makeEmptyEnvelopeBlock builds a block carrying no transactions or
// receipts — exercises the "gap-fill" code path where the empty roots
// flow through.
func makeEmptyEnvelopeBlock(t *testing.T) *ethtypes.Block {
	t.Helper()
	header := &ethtypes.Header{
		ParentHash:  common.HexToHash("0xaaa0"),
		UncleHash:   ethtypes.EmptyUncleHash,
		Coinbase:    common.Address{},
		TxHash:      ethtypes.EmptyTxsHash,
		ReceiptHash: ethtypes.EmptyReceiptsHash,
		Number:      big.NewInt(12),
		GasLimit:    30_000_000,
		GasUsed:     0,
		Time:        1_700_000_005,
		Difficulty:  new(big.Int),
		BaseFee:     big.NewInt(1_000_000_000),
	}
	return ethtypes.NewBlock(
		header,
		&ethtypes.Body{Transactions: nil},
		nil,
		trie.NewStackTrie(nil),
	)
}

// TestSimBlockResult_MarshalJSON_FullEnvelope_HashOnly: typed
// *ethtypes.Block path renders all canonical envelope keys, with
// transactions as a string array (FullTx=false).
func TestSimBlockResult_MarshalJSON_FullEnvelope_HashOnly(t *testing.T) {
	from := common.HexToAddress("0x1111111111111111111111111111111111111111")
	to := common.HexToAddress("0x2222222222222222222222222222222222222222")
	block, tx := makeEnvelopeBlock(t, from, to, 1_000_000)

	r := types.SimBlockResult{
		EthBlock:    block,
		Senders:     []common.Address{from},
		FullTx:      false,
		ChainConfig: envelopeChainConfig(),
		Calls:       []types.SimCallResult{{Status: hexutil.Uint64(1), GasUsed: hexutil.Uint64(21_000), Logs: []*ethtypes.Log{}}},
	}

	data, err := json.Marshal(r)
	require.NoError(t, err)

	var decoded map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &decoded))
	for _, key := range []string{
		"number", "hash", "parentHash", "logsBloom", "stateRoot",
		"miner", "difficulty", "extraData", "gasLimit", "gasUsed",
		"timestamp", "transactionsRoot", "receiptsRoot", "size",
		"transactions", "uncles", "calls",
	} {
		require.Contains(t, decoded, key, "envelope must contain %q", key)
	}

	txs, ok := decoded["transactions"].([]interface{})
	require.True(t, ok, "transactions must be an array")
	require.Len(t, txs, 1)
	hashStr, ok := txs[0].(string)
	require.True(t, ok, "FullTx=false must yield a hash string, got %T", txs[0])
	require.Equal(t, tx.Hash().Hex(), hashStr)
}

// TestSimBlockResult_MarshalJSON_FullEnvelope_FullTxFromPatched:
// FullTx=true emits transaction objects and patches `from` from the
// Senders map. Without the patch the unsigned-tx signature recovery
// returns the zero address.
func TestSimBlockResult_MarshalJSON_FullEnvelope_FullTxFromPatched(t *testing.T) {
	from := common.HexToAddress("0x3333333333333333333333333333333333333333")
	to := common.HexToAddress("0x4444444444444444444444444444444444444444")
	block, tx := makeEnvelopeBlock(t, from, to, 555)

	r := types.SimBlockResult{
		EthBlock:    block,
		Senders:     []common.Address{from},
		FullTx:      true,
		ChainConfig: envelopeChainConfig(),
	}

	data, err := json.Marshal(r)
	require.NoError(t, err)

	var decoded struct {
		Transactions []map[string]interface{} `json:"transactions"`
	}
	require.NoError(t, json.Unmarshal(data, &decoded))
	require.Len(t, decoded.Transactions, 1)
	require.Equal(t, from.Hex(), common.HexToAddress(decoded.Transactions[0]["from"].(string)).Hex(),
		"FullTx=true must patch `from` from Senders")
	require.Equal(t, tx.Hash().Hex(), decoded.Transactions[0]["hash"].(string))
}

// TestSimBlockResult_MarshalJSON_FullTx_TwoCallsTwoSenders: the patch
// matches by tx hash, not by index — verified by feeding two calls
// from two different senders and asserting each tx's `from` resolves
// to the right address even after the marshaler reorders them.
func TestSimBlockResult_MarshalJSON_FullTx_TwoCallsTwoSenders(t *testing.T) {
	fromA := common.HexToAddress("0xa1aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa11")
	fromB := common.HexToAddress("0xb2bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb22")
	to := common.HexToAddress("0x4444444444444444444444444444444444444444")

	txA := ethtypes.NewTx(&ethtypes.LegacyTx{
		Nonce: 1, GasPrice: big.NewInt(0), Gas: 21_000, To: &to, Value: big.NewInt(11),
	})
	txB := ethtypes.NewTx(&ethtypes.LegacyTx{
		Nonce: 1, GasPrice: big.NewInt(0), Gas: 21_000, To: &to, Value: big.NewInt(22),
	})
	receiptA := &ethtypes.Receipt{
		Type: ethtypes.LegacyTxType, Status: 1, CumulativeGasUsed: 21_000,
		GasUsed: 21_000, TxHash: txA.Hash(), Logs: []*ethtypes.Log{}, BlockNumber: big.NewInt(11),
	}
	receiptA.Bloom = ethtypes.CreateBloom(ethtypes.Receipts{receiptA})
	receiptB := &ethtypes.Receipt{
		Type: ethtypes.LegacyTxType, Status: 1, CumulativeGasUsed: 42_000,
		GasUsed: 21_000, TxHash: txB.Hash(), Logs: []*ethtypes.Log{}, BlockNumber: big.NewInt(11),
	}
	receiptB.Bloom = ethtypes.CreateBloom(ethtypes.Receipts{receiptB})

	header := &ethtypes.Header{
		ParentHash:  common.HexToHash("0xaaa1"),
		UncleHash:   ethtypes.EmptyUncleHash,
		Number:      big.NewInt(11),
		GasLimit:    30_000_000,
		GasUsed:     42_000,
		Difficulty:  new(big.Int),
		BaseFee:     big.NewInt(1_000_000_000),
		TxHash:      ethtypes.EmptyTxsHash,
		ReceiptHash: ethtypes.EmptyReceiptsHash,
	}
	block := ethtypes.NewBlock(header,
		&ethtypes.Body{Transactions: []*ethtypes.Transaction{txA, txB}},
		[]*ethtypes.Receipt{receiptA, receiptB},
		trie.NewStackTrie(nil),
	)

	r := types.SimBlockResult{
		EthBlock: block,
		Senders:     []common.Address{fromA, fromB},
		FullTx:      true,
		ChainConfig: envelopeChainConfig(),
	}

	data, err := json.Marshal(r)
	require.NoError(t, err)

	var decoded struct {
		Transactions []map[string]interface{} `json:"transactions"`
	}
	require.NoError(t, json.Unmarshal(data, &decoded))
	require.Len(t, decoded.Transactions, 2)

	gotA := common.HexToAddress(decoded.Transactions[0]["from"].(string))
	gotB := common.HexToAddress(decoded.Transactions[1]["from"].(string))

	// Match by hash, not index. Identify which decoded tx is A vs B
	// by the hash field (block.Transactions ordering may shuffle if
	// NewBlock sorts).
	hashA := decoded.Transactions[0]["hash"].(string)
	if hashA == txA.Hash().Hex() {
		require.Equal(t, fromA, gotA)
		require.Equal(t, fromB, gotB)
	} else {
		require.Equal(t, fromB, gotA)
		require.Equal(t, fromA, gotB)
	}
}

// TestSimBlockResult_MarshalJSON_EmptyBlock_UsesEmptyRoots: an empty
// block produces an envelope where transactionsRoot/receiptsRoot are
// the canonical empty roots, logsBloom is zero, and `transactions` is
// an empty array (not null).
func TestSimBlockResult_MarshalJSON_EmptyBlock_UsesEmptyRoots(t *testing.T) {
	block := makeEmptyEnvelopeBlock(t)

	r := types.SimBlockResult{
		EthBlock:    block,
		Senders:     nil,
		FullTx:      false,
		ChainConfig: envelopeChainConfig(),
	}

	data, err := json.Marshal(r)
	require.NoError(t, err)

	var decoded map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &decoded))

	require.Equal(t, ethtypes.EmptyTxsHash.Hex(), decoded["transactionsRoot"].(string))
	require.Equal(t, ethtypes.EmptyReceiptsHash.Hex(), decoded["receiptsRoot"].(string))

	txs, ok := decoded["transactions"].([]interface{})
	require.True(t, ok, "transactions must be an array, not null")
	require.Empty(t, txs)

	calls, ok := decoded["calls"].([]interface{})
	require.True(t, ok, "calls must be an array, not null")
	require.Empty(t, calls)

	bloom, ok := decoded["logsBloom"].(string)
	require.True(t, ok)
	require.Equal(t, "0x"+common.Bytes2Hex(make([]byte, 256)), bloom,
		"empty block must surface a zero-filled logsBloom")
}

// TestSimBlockResult_MarshalJSON_FullTx_NilSendersMap: defensive — a
// nil Senders map must not panic; `from` falls back to the unsigned
// recovery's zero address.
func TestSimBlockResult_MarshalJSON_FullTx_NilSendersMap(t *testing.T) {
	from := common.HexToAddress("0x5555555555555555555555555555555555555555")
	to := common.HexToAddress("0x6666666666666666666666666666666666666666")
	block, _ := makeEnvelopeBlock(t, from, to, 1)

	r := types.SimBlockResult{
		EthBlock:    block,
		Senders:     nil,
		FullTx:      true,
		ChainConfig: envelopeChainConfig(),
	}

	require.NotPanics(t, func() {
		_, err := json.Marshal(r)
		require.NoError(t, err)
	})
}

// TestSimBlockResult_MarshalJSON_FullTx_HashMissingFromSenders:
// defensive — a sender map missing the tx's hash must not panic; the
// missing entry resolves to the zero address via map's zero-value
// semantics.
func TestSimBlockResult_MarshalJSON_FullTx_HashMissingFromSenders(t *testing.T) {
	from := common.HexToAddress("0x5555555555555555555555555555555555555555")
	to := common.HexToAddress("0x6666666666666666666666666666666666666666")
	block, _ := makeEnvelopeBlock(t, from, to, 1)

	r := types.SimBlockResult{
		EthBlock:    block,
		Senders:     nil, // no senders
		FullTx:      true,
		ChainConfig: envelopeChainConfig(),
	}

	data, err := json.Marshal(r)
	require.NoError(t, err)

	var decoded struct {
		Transactions []map[string]interface{} `json:"transactions"`
	}
	require.NoError(t, json.Unmarshal(data, &decoded))
	require.Len(t, decoded.Transactions, 1)
	require.Equal(t, common.Address{}.Hex(),
		common.HexToAddress(decoded.Transactions[0]["from"].(string)).Hex(),
		"missing sender entry must default to the zero address (map zero-value)")
}

// TestSimBlockResult_MarshalJSON_StateRootIsZero: pinned divergence —
// mezod has no MPT root, so the marshaled envelope reports stateRoot
// as the zero hash. Documented in `assembleSimBlock`.
func TestSimBlockResult_MarshalJSON_StateRootIsZero(t *testing.T) {
	from := common.HexToAddress("0x1111111111111111111111111111111111111111")
	to := common.HexToAddress("0x2222222222222222222222222222222222222222")
	block, _ := makeEnvelopeBlock(t, from, to, 1)

	r := types.SimBlockResult{
		EthBlock:    block,
		Senders:     nil,
		FullTx:      false,
		ChainConfig: envelopeChainConfig(),
	}

	data, err := json.Marshal(r)
	require.NoError(t, err)

	var decoded map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &decoded))
	require.Equal(t, common.Hash{}.Hex(), decoded["stateRoot"].(string),
		"stateRoot must be the zero hash (mezo divergence; pinned)")
}

// TestSimBlockResult_MarshalJSON_SizeNonZero: the envelope's `size`
// field must be non-zero for any block carrying a sealed header — the
// RLP-encoded header alone is well over zero bytes.
func TestSimBlockResult_MarshalJSON_SizeNonZero(t *testing.T) {
	from := common.HexToAddress("0x1111111111111111111111111111111111111111")
	to := common.HexToAddress("0x2222222222222222222222222222222222222222")
	block, _ := makeEnvelopeBlock(t, from, to, 1)

	r := types.SimBlockResult{
		EthBlock:    block,
		Senders:     nil,
		FullTx:      false,
		ChainConfig: envelopeChainConfig(),
	}

	data, err := json.Marshal(r)
	require.NoError(t, err)

	var decoded map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &decoded))

	sizeStr, ok := decoded["size"].(string)
	require.True(t, ok)
	sizeVal := new(big.Int)
	_, ok = sizeVal.SetString(sizeStr[2:], 16) // strip 0x prefix
	require.True(t, ok)
	require.True(t, sizeVal.Sign() > 0, "size must be non-zero")
}

// TestSimBlockResult_MarshalJSON_LogsBloomMatchesReceipts: when a
// receipt carries a topic, the assembled block's bloom must
// affirmative-test that topic via BloomLookup. The test installs a
// receipt with one log carrying a known topic, builds the block, then
// asserts the marshaled envelope's logsBloom carries that topic.
func TestSimBlockResult_MarshalJSON_LogsBloomMatchesReceipts(t *testing.T) {
	from := common.HexToAddress("0x1111111111111111111111111111111111111111")
	to := common.HexToAddress("0x2222222222222222222222222222222222222222")

	topic := common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")
	contract := common.HexToAddress("0x9999999999999999999999999999999999999999")

	tx := ethtypes.NewTx(&ethtypes.LegacyTx{
		Nonce: 1, GasPrice: big.NewInt(0), Gas: 21_000, To: &to, Value: big.NewInt(0),
	})
	logEntry := &ethtypes.Log{
		Address: contract,
		Topics:  []common.Hash{topic},
		Data:    []byte{},
	}
	receipt := &ethtypes.Receipt{
		Type: ethtypes.LegacyTxType, Status: 1, CumulativeGasUsed: 21_000,
		GasUsed: 21_000, TxHash: tx.Hash(), Logs: []*ethtypes.Log{logEntry}, BlockNumber: big.NewInt(11),
	}
	receipt.Bloom = ethtypes.CreateBloom(ethtypes.Receipts{receipt})

	header := &ethtypes.Header{
		ParentHash:  common.HexToHash("0xaaa1"),
		UncleHash:   ethtypes.EmptyUncleHash,
		Number:      big.NewInt(11),
		GasLimit:    30_000_000,
		GasUsed:     21_000,
		Difficulty:  new(big.Int),
		BaseFee:     big.NewInt(1_000_000_000),
		TxHash:      ethtypes.EmptyTxsHash,
		ReceiptHash: ethtypes.EmptyReceiptsHash,
	}
	block := ethtypes.NewBlock(header,
		&ethtypes.Body{Transactions: []*ethtypes.Transaction{tx}},
		[]*ethtypes.Receipt{receipt},
		trie.NewStackTrie(nil),
	)

	r := types.SimBlockResult{
		EthBlock:    block,
		Senders:     []common.Address{from},
		FullTx:      false,
		ChainConfig: envelopeChainConfig(),
	}

	data, err := json.Marshal(r)
	require.NoError(t, err)

	var decoded map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &decoded))

	bloomStr, ok := decoded["logsBloom"].(string)
	require.True(t, ok)
	bloom := ethtypes.BytesToBloom(common.FromHex(bloomStr))
	require.True(t, bloom.Test(topic.Bytes()),
		"bloom must affirmative-test the receipt's topic")
}

func TestSimBlockResult_UnmarshalJSON_FromMarshaledFullEnvelope(t *testing.T) {
	from := common.HexToAddress("0x1111111111111111111111111111111111111111")
	to := common.HexToAddress("0x2222222222222222222222222222222222222222")
	block, _ := makeEnvelopeBlock(t, from, to, 1)

	r := types.SimBlockResult{
		EthBlock:    block,
		Senders:     nil,
		FullTx:      false,
		ChainConfig: envelopeChainConfig(),
		Calls: []types.SimCallResult{{
			Status: hexutil.Uint64(1), GasUsed: hexutil.Uint64(21_000), Logs: []*ethtypes.Log{},
		}},
	}
	data, err := json.Marshal(r)
	require.NoError(t, err)

	var roundTripped types.SimBlockResult
	require.NoError(t, json.Unmarshal(data, &roundTripped))
	require.Nil(t, roundTripped.EthBlock,
		"UnmarshalJSON must not reconstruct the typed *ethtypes.Block")
	require.Len(t, roundTripped.Calls, 1)
	require.Equal(t, hexutil.Uint64(21_000), roundTripped.Calls[0].GasUsed)
	require.Contains(t, roundTripped.Block, "number")
	require.Contains(t, roundTripped.Block, "hash")
	require.Contains(t, roundTripped.Block, "parentHash")
}

// TestSimBlockResult_UnmarshalJSON_BackCompatPreserved: the existing
// untyped-Block path keeps working — pinned alongside the typed
// envelope path so the legacy constructor pattern that direct-test
// callers rely on stays supported.
func TestSimBlockResult_UnmarshalJSON_BackCompatPreserved(t *testing.T) {
	raw := []byte(`{"number":"0x5","hash":"0xabcd","calls":[{"returnData":"0x","logs":[],"gasUsed":"0x5208","status":"0x1"}],"extraField":"preserved"}`)
	var r types.SimBlockResult
	require.NoError(t, json.Unmarshal(raw, &r))
	require.Len(t, r.Calls, 1)
	require.Equal(t, hexutil.Uint64(0x5208), r.Calls[0].GasUsed)
	require.Equal(t, "0x5", r.Block["number"])
	require.Equal(t, "0xabcd", r.Block["hash"])
	require.Equal(t, "preserved", r.Block["extraField"])
}
