package types

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
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

// SimBlockResult envelopes a single simulated block. The marshal path
// uses EthBlock + Senders + FullTx + ChainConfig (populated by the
// driver); UnmarshalJSON populates Block instead so gRPC round-trips
// keep working without reconstructing a *ethtypes.Block from JSON.
type SimBlockResult struct {
	EthBlock    *ethtypes.Block                `json:"-"`
	Senders     map[common.Hash]common.Address `json:"-"`
	FullTx      bool                           `json:"-"`
	ChainConfig *params.ChainConfig            `json:"-"`
	Block       map[string]interface{}         `json:"-"`
	Calls       []SimCallResult                `json:"calls"`
}

func (r SimBlockResult) MarshalJSON() ([]byte, error) {
	var out map[string]interface{}
	switch {
	case r.EthBlock != nil:
		out = RPCMarshalBlock(r.EthBlock, true, r.FullTx, r.ChainConfig)
		if r.FullTx {
			if raw, ok := out["transactions"].([]interface{}); ok {
				for _, tx := range raw {
					rpcTx, ok := tx.(*RPCTransaction)
					if !ok {
						return nil, errors.New("simulated transaction result has invalid type")
					}
					rpcTx.From = r.Senders[rpcTx.Hash]
				}
			}
		}
	default:
		out = make(map[string]interface{}, len(r.Block)+1)
		for k, v := range r.Block {
			out[k] = v
		}
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
