package types

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"
)

// MaxSimulateBlocks caps the number of simulated blocks in a single
// request. The RPC layer counts submitted blocks; the keeper counts the
// span from base.Number to the last sim block including gap-fills.
const MaxSimulateBlocks = 256

// MaxSimulateCalls caps the cumulative call count across all simulated
// blocks (post-sanitize, gap-fills included).
const MaxSimulateCalls = 1000

// SimTimestampIncrement is the default gap, in seconds, between
// sequential simulated blocks when the caller omits Time overrides.
// Matches mezo's ~3s average CometBFT block time so callers who let
// the sim fabricate timestamps land in a realistic ballpark.
const SimTimestampIncrement = 3

// CountSimCalls returns the cumulative number of calls across the
// supplied simulated blocks.
func CountSimCalls(blocks []SimBlock) int {
	var n int
	for i := range blocks {
		n += len(blocks[i].Calls)
	}
	return n
}

// SimOpts are the top-level options passed to `eth_simulateV1`. Fields not
// yet consumed by the driver are still parsed so unknown-call behavior is
// deterministic and future phases can extend without reworking the
// unmarshal path.
type SimOpts struct {
	BlockStateCalls        []SimBlock `json:"blockStateCalls"`
	TraceTransfers         bool       `json:"traceTransfers"`
	Validation             bool       `json:"validation"`
	ReturnFullTransactions bool       `json:"returnFullTransactions"`
}

// SimBlock is a batch of calls executed sequentially inside one simulated
// block, with optional block and state overrides.
type SimBlock struct {
	BlockOverrides *SimBlockOverrides `json:"blockOverrides,omitempty"`
	StateOverrides StateOverride      `json:"stateOverrides,omitempty"`
	Calls          []TransactionArgs  `json:"calls"`
}

// SimBlockOverrides overrides header fields for a simulated block. All
// fields are optional; unset fields inherit from the parent simulated (or
// base) header. Fields for EIPs the mezo chain model does not support
// (EIP-4788 beacon root, EIP-4895 withdrawals, blob-gas fields) are parsed
// so the driver can explicitly reject them rather than silently ignore
// them.
type SimBlockOverrides struct {
	Number        *hexutil.Big          `json:"number,omitempty"`
	Difficulty    *hexutil.Big          `json:"difficulty,omitempty"`
	Time          *hexutil.Uint64       `json:"time,omitempty"`
	GasLimit      *hexutil.Uint64       `json:"gasLimit,omitempty"`
	FeeRecipient  *common.Address       `json:"feeRecipient,omitempty"`
	PrevRandao    *common.Hash          `json:"prevRandao,omitempty"`
	BaseFeePerGas *hexutil.Big          `json:"baseFeePerGas,omitempty"`
	BlobBaseFee   *hexutil.Big          `json:"blobBaseFee,omitempty"`
	BeaconRoot    *common.Hash          `json:"beaconRoot,omitempty"`
	Withdrawals   *ethtypes.Withdrawals `json:"withdrawals,omitempty"`
}

// SimCallResult is the result of one simulated call. Error carries the
// spec-reserved JSON-RPC code + message + optional hex data payload
// directly; the wire format does not go through any translation between
// driver and client.
type SimCallResult struct {
	ReturnValue hexutil.Bytes   `json:"returnData"`
	Logs        []*ethtypes.Log `json:"logs"`
	GasUsed     hexutil.Uint64  `json:"gasUsed"`
	Status      hexutil.Uint64  `json:"status"`
	Error       *SimError       `json:"error,omitempty"`
}

// MarshalJSON forces `logs` to serialize as `[]` rather than `null` when
// empty, matching the execution-apis spec.
func (r SimCallResult) MarshalJSON() ([]byte, error) {
	type alias SimCallResult
	if r.Logs == nil {
		r.Logs = []*ethtypes.Log{}
	}
	return json.Marshal(alias(r))
}

// SimBlockResult envelopes a single simulated block over the wire.
// Block holds the spec-shaped header fields; Calls holds the per-call
// results. MarshalJSON flattens Block + Calls into one JSON object so
// the spec emits header keys at the top level alongside `calls`.
//
// Two construction paths populate Block:
//   - keeper-side: NewSimBlockResult takes the typed *ethtypes.Block +
//     senders + chain config and renders the map via marshalSimBlock.
//   - rpc-side: UnmarshalJSON populates Block from the gRPC payload
//     bytes (already-rendered field map).
//
// Both produce the same in-memory shape, so MarshalJSON has a single
// branch.
type SimBlockResult struct {
	Block map[string]interface{} `json:"-"`
	Calls []SimCallResult        `json:"calls"`
}

// NewSimBlockResult builds a SimBlockResult from the keeper-side typed
// inputs. senders must be keyed by call index in the same order
// block.Transactions() returns; ethtypes.NewBlock preserves input order
// so call-index alignment is stable. When fullTx is true, marshalSimBlock
// patches each tx object's `from` from senders[i] — the synthetic
// transactions the keeper assembles are unsigned LegacyTx values, so
// signature recovery would otherwise yield the zero address.
func NewSimBlockResult(
	block *ethtypes.Block,
	senders []common.Address,
	fullTx bool,
	chainConfig *params.ChainConfig,
	calls []SimCallResult,
) *SimBlockResult {
	return &SimBlockResult{
		Block: marshalSimBlock(block, senders, fullTx, chainConfig),
		Calls: calls,
	}
}

func (r SimBlockResult) MarshalJSON() ([]byte, error) {
	out := make(map[string]interface{}, len(r.Block)+1)
	for k, v := range r.Block {
		out[k] = v
	}
	if r.Calls == nil {
		out["calls"] = []SimCallResult{}
	} else {
		out["calls"] = r.Calls
	}
	return json.Marshal(out)
}

// UnmarshalJSON extracts the `calls` field into `Calls` and stores the
// remaining fields in `Block`. Unknown fields are tolerated to allow the
// type to evolve alongside the spec without breaking clients.
func (r *SimBlockResult) UnmarshalJSON(data []byte) error {
	raw := map[string]json.RawMessage{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if callsRaw, ok := raw["calls"]; ok {
		if err := json.Unmarshal(callsRaw, &r.Calls); err != nil {
			return err
		}
		delete(raw, "calls")
	}
	if r.Block == nil {
		r.Block = map[string]interface{}{}
	}
	for k, v := range raw {
		var anyVal interface{}
		if err := json.Unmarshal(v, &anyVal); err != nil {
			return err
		}
		r.Block[k] = anyVal
	}
	return nil
}

// marshalSimBlock builds the spec-shaped envelope map for a simulated
// block: header fields plus a `transactions` array (hashes when fullTx is
// false, full tx objects otherwise). When fullTx is true, `from` on each
// tx object is patched from senders[i].
func marshalSimBlock(
	block *ethtypes.Block,
	senders []common.Address,
	fullTx bool,
	config *params.ChainConfig,
) map[string]interface{} {
	out := marshalEthBlock(block, true, fullTx, config)
	if !fullTx {
		return out
	}
	raw, ok := out["transactions"].([]interface{})
	if !ok {
		return out
	}
	for i, tx := range raw {
		rpcTx := tx.(*rpcTransaction) //nolint:forcetypeassert
		if i < len(senders) {
			rpcTx.From = senders[i]
		}
	}
	return out
}

// UnmarshalSimOpts decodes a SimulateV1Request.opts payload and rejects
// overrides for EIPs that mezo does not support. All rejections are
// returned as *SimError carrying -32602 (invalid params) so the gRPC
// handler surfaces the spec-reserved code without translation.
func UnmarshalSimOpts(data []byte) (*SimOpts, error) {
	var opts SimOpts
	if err := json.Unmarshal(data, &opts); err != nil {
		return nil, NewSimInvalidParams(fmt.Sprintf("invalid simOpts: %s", err.Error()))
	}
	for bi, block := range opts.BlockStateCalls {
		if block.BlockOverrides == nil {
			continue
		}
		if block.BlockOverrides.BeaconRoot != nil {
			return nil, NewSimInvalidParams(fmt.Sprintf(
				"block %d: BlockOverrides.BeaconRoot is not supported on mezo (no beacon chain)", bi,
			))
		}
		if block.BlockOverrides.Withdrawals != nil {
			return nil, NewSimInvalidParams(fmt.Sprintf(
				"block %d: BlockOverrides.Withdrawals is not supported on mezo (no EL-CL withdrawal queue)", bi,
			))
		}
		if block.BlockOverrides.BlobBaseFee != nil {
			return nil, NewSimInvalidParams(fmt.Sprintf(
				"block %d: BlockOverrides.BlobBaseFee is not supported on mezo (blob transactions rejected)", bi,
			))
		}
	}
	return &opts, nil
}

// BuildSimCallResult maps a MsgEthereumTxResponse to a spec-shaped
// per-call result. VM failures land on Error as a *SimError carrying
// the spec-reserved code directly: 3 for revert, -32015 for other VM
// errors.
func BuildSimCallResult(res *MsgEthereumTxResponse) SimCallResult {
	status := uint64(1)
	if res.Failed() {
		status = 0
	}
	logs := LogsToEthereum(res.Logs)
	if logs == nil {
		logs = []*ethtypes.Log{}
	}
	out := SimCallResult{
		ReturnValue: res.Ret,
		Logs:        logs,
		GasUsed:     hexutil.Uint64(res.GasUsed),
		Status:      hexutil.Uint64(status),
	}
	if res.Failed() {
		if res.VmError == vm.ErrExecutionReverted.Error() {
			out.Error = NewSimReverted(res.Ret)
		} else {
			out.Error = NewSimVMError(res.VmError)
		}
	}
	return out
}

// The block / header marshaling helpers below (marshalEthHeader,
// marshalEthBlock, rpcTransaction, newRPCTransaction*) are copied from
// go-ethereum's internal/ethapi/api.go (LGPL), reproduced here because
// the upstream package sits behind Go's `internal/` visibility rule.
// Drift from upstream:
//   - rpcTransaction is unexported and the doc comment notes that From
//     starts at the zero address until marshalSimBlock patches it from
//     the senders slice (simulated transactions are unsigned).
//   - marshalEthBlock drops the upstream `receipts` parameter; the
//     keeper seals receipt-derived header fields (TxHash/ReceiptHash/
//     Bloom) via ethtypes.NewBlock before this is called.

// marshalEthHeader converts the given header to the RPC output.
func marshalEthHeader(head *ethtypes.Header) map[string]interface{} {
	result := map[string]interface{}{
		"number":           (*hexutil.Big)(head.Number),
		"hash":             head.Hash(),
		"parentHash":       head.ParentHash,
		"nonce":            head.Nonce,
		"mixHash":          head.MixDigest,
		"sha3Uncles":       head.UncleHash,
		"logsBloom":        head.Bloom,
		"stateRoot":        head.Root,
		"miner":            head.Coinbase,
		"difficulty":       (*hexutil.Big)(head.Difficulty),
		"extraData":        hexutil.Bytes(head.Extra),
		"gasLimit":         hexutil.Uint64(head.GasLimit),
		"gasUsed":          hexutil.Uint64(head.GasUsed),
		"timestamp":        hexutil.Uint64(head.Time),
		"transactionsRoot": head.TxHash,
		"receiptsRoot":     head.ReceiptHash,
	}
	if head.BaseFee != nil {
		result["baseFeePerGas"] = (*hexutil.Big)(head.BaseFee)
	}
	if head.WithdrawalsHash != nil {
		result["withdrawalsRoot"] = head.WithdrawalsHash
	}
	if head.BlobGasUsed != nil {
		result["blobGasUsed"] = hexutil.Uint64(*head.BlobGasUsed)
	}
	if head.ExcessBlobGas != nil {
		result["excessBlobGas"] = hexutil.Uint64(*head.ExcessBlobGas)
	}
	if head.ParentBeaconRoot != nil {
		result["parentBeaconBlockRoot"] = head.ParentBeaconRoot
	}
	return result
}

// marshalEthBlock converts the given block to the RPC output. When
// inclTx is true the `transactions` field is set; fullTx selects
// between hash-only and *rpcTransaction objects (whose `from` is the
// zero address until the caller patches it from a senders slice, since
// simulated transactions are unsigned).
func marshalEthBlock(block *ethtypes.Block, inclTx bool, fullTx bool, config *params.ChainConfig) map[string]interface{} {
	fields := marshalEthHeader(block.Header())
	fields["size"] = hexutil.Uint64(block.Size())

	if inclTx {
		formatTx := func(_ int, tx *ethtypes.Transaction) interface{} {
			return tx.Hash()
		}
		if fullTx {
			formatTx = func(idx int, _ *ethtypes.Transaction) interface{} {
				return newRPCTransactionFromBlockIndex(block, uint64(idx), config) //nolint:gosec
			}
		}
		txs := block.Transactions()
		transactions := make([]interface{}, len(txs))
		for i, tx := range txs {
			transactions[i] = formatTx(i, tx)
		}
		fields["transactions"] = transactions
	}
	uncles := block.Uncles()
	uncleHashes := make([]common.Hash, len(uncles))
	for i, uncle := range uncles {
		uncleHashes[i] = uncle.Hash()
	}
	fields["uncles"] = uncleHashes
	if block.Header().WithdrawalsHash != nil {
		fields["withdrawals"] = block.Withdrawals()
	}
	return fields
}

// rpcTransaction mirrors go-ethereum's internal/ethapi.RPCTransaction
// JSON shape. Lives here rather than rpc/types because rpc/types
// already depends on x/evm/types.
type rpcTransaction struct {
	BlockHash        *common.Hash         `json:"blockHash"`
	BlockNumber      *hexutil.Big         `json:"blockNumber"`
	From             common.Address       `json:"from"`
	Gas              hexutil.Uint64       `json:"gas"`
	GasPrice         *hexutil.Big         `json:"gasPrice"`
	GasFeeCap        *hexutil.Big         `json:"maxFeePerGas,omitempty"`
	GasTipCap        *hexutil.Big         `json:"maxPriorityFeePerGas,omitempty"`
	Hash             common.Hash          `json:"hash"`
	Input            hexutil.Bytes        `json:"input"`
	Nonce            hexutil.Uint64       `json:"nonce"`
	To               *common.Address      `json:"to"`
	TransactionIndex *hexutil.Uint64      `json:"transactionIndex"`
	Value            *hexutil.Big         `json:"value"`
	Type             hexutil.Uint64       `json:"type"`
	Accesses         *ethtypes.AccessList `json:"accessList,omitempty"`
	ChainID          *hexutil.Big         `json:"chainId,omitempty"`
	V                *hexutil.Big         `json:"v"`
	R                *hexutil.Big         `json:"r"`
	S                *hexutil.Big         `json:"s"`
}

// newRPCTransactionFromBlockIndex returns a transaction that will
// serialize to the RPC representation, located at the supplied block
// index.
func newRPCTransactionFromBlockIndex(b *ethtypes.Block, index uint64, config *params.ChainConfig) *rpcTransaction {
	txs := b.Transactions()
	if index >= uint64(len(txs)) {
		return nil
	}
	return newRPCTransaction(txs[index], b.Hash(), b.NumberU64(), index, b.BaseFee(), config.ChainID)
}

// newRPCTransaction populates the location fields of an rpcTransaction.
// Simulated transactions are unsigned; signature recovery yields the
// zero address, so the caller patches From from the senders slice
// after marshaling.
func newRPCTransaction(
	tx *ethtypes.Transaction,
	blockHash common.Hash,
	blockNumber, index uint64,
	baseFee *big.Int,
	chainID *big.Int,
) *rpcTransaction {
	v, r, s := tx.RawSignatureValues()
	result := &rpcTransaction{
		Type:     hexutil.Uint64(tx.Type()),
		Gas:      hexutil.Uint64(tx.Gas()),
		GasPrice: (*hexutil.Big)(tx.GasPrice()),
		Hash:     tx.Hash(),
		Input:    hexutil.Bytes(tx.Data()),
		Nonce:    hexutil.Uint64(tx.Nonce()),
		To:       tx.To(),
		Value:    (*hexutil.Big)(tx.Value()),
		V:        (*hexutil.Big)(v),
		R:        (*hexutil.Big)(r),
		S:        (*hexutil.Big)(s),
		ChainID:  (*hexutil.Big)(chainID),
	}
	if blockHash != (common.Hash{}) {
		result.BlockHash = &blockHash
		result.BlockNumber = (*hexutil.Big)(new(big.Int).SetUint64(blockNumber))
		idx := hexutil.Uint64(index)
		result.TransactionIndex = &idx
	}
	switch tx.Type() {
	case ethtypes.AccessListTxType:
		al := tx.AccessList()
		result.Accesses = &al
		result.ChainID = (*hexutil.Big)(tx.ChainId())
	case ethtypes.DynamicFeeTxType:
		al := tx.AccessList()
		result.Accesses = &al
		result.ChainID = (*hexutil.Big)(tx.ChainId())
		result.GasFeeCap = (*hexutil.Big)(tx.GasFeeCap())
		result.GasTipCap = (*hexutil.Big)(tx.GasTipCap())
		if baseFee != nil && blockHash != (common.Hash{}) {
			price := math.BigMin(new(big.Int).Add(tx.GasTipCap(), baseFee), tx.GasFeeCap())
			result.GasPrice = (*hexutil.Big)(price)
		} else {
			result.GasPrice = (*hexutil.Big)(tx.GasFeeCap())
		}
	}
	return result
}
