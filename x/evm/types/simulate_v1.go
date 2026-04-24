package types

import (
	"encoding/json"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
)

// SimOpts is the driver-side view of SimOpts (see rpc/types/simulate_v1.go).
// The shape matches the execution-apis spec; fields not yet used by the
// driver are still parsed so unknown-call behavior is deterministic and
// future phases can extend without reworking the unmarshal path.
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

// SimBlockOverrides mirrors rpc/types.BlockOverrides. Fields for EIPs the
// mezo chain model does not support (EIP-4788 beacon root, EIP-4895
// withdrawals, blob-gas fields) are parsed so the driver can explicitly
// reject them rather than silently ignore them.
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

// SimCallResult mirrors rpc/types.SimCallResult on the keeper side.
// Error carries a spec-reserved JSON-RPC code + message + optional hex
// data payload directly; the wire format does not go through any
// translation on the way to the client.
type SimCallResult struct {
	ReturnValue hexutil.Bytes   `json:"returnData"`
	Logs        []*ethtypes.Log `json:"logs"`
	GasUsed     hexutil.Uint64  `json:"gasUsed"`
	Status      hexutil.Uint64  `json:"status"`
	Error       *SimError       `json:"error,omitempty"`
}

// SimBlockResult envelopes a single simulated block. Block carries the
// header fields as an untyped map; Calls carries the per-call outputs.
// The custom MarshalJSON embeds Block's fields at the top level alongside
// `calls`, matching the execution-apis response shape and the
// rpc/types.SimBlockResult unmarshaler on the other side of the gRPC
// wire.
type SimBlockResult struct {
	Block map[string]interface{} `json:"-"`
	Calls []SimCallResult        `json:"calls"`
}

// MarshalJSON flattens Block's fields into the top-level object and
// appends `calls`. Keep this in sync with
// rpc/types.SimBlockResult.UnmarshalJSON, which reads the same shape.
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
