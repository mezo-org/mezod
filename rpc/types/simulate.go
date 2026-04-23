package types

import (
	"encoding/json"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	evmtypes "github.com/mezo-org/mezod/x/evm/types"
)

// SimOpts are the top-level options passed to `eth_simulateV1`.
type SimOpts struct {
	BlockStateCalls        []SimBlock `json:"blockStateCalls"`
	TraceTransfers         bool       `json:"traceTransfers"`
	Validation             bool       `json:"validation"`
	ReturnFullTransactions bool       `json:"returnFullTransactions"`
}

// SimBlock is a batch of calls executed sequentially inside one simulated
// block, with optional block / state overrides.
type SimBlock struct {
	BlockOverrides *BlockOverrides            `json:"blockOverrides,omitempty"`
	StateOverrides *StateOverride             `json:"stateOverrides,omitempty"`
	Calls          []evmtypes.TransactionArgs `json:"calls"`
}

// BlockOverrides overrides header fields for a simulated block. Fields are
// all optional; unset fields inherit from the parent simulated (or base)
// header.
type BlockOverrides struct {
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
// directly; the keeper side emits the same *evmtypes.SimError instance
// into the per-call JSON, so no translation happens at unmarshal time.
type SimCallResult struct {
	ReturnValue hexutil.Bytes      `json:"returnData"`
	Logs        []*ethtypes.Log    `json:"logs"`
	GasUsed     hexutil.Uint64     `json:"gasUsed"`
	Status      hexutil.Uint64     `json:"status"`
	Error       *evmtypes.SimError `json:"error,omitempty"`
}

// MarshalJSON forces `Logs` to serialize as `[]` rather than `null` when
// empty, matching the execution-apis spec.
func (r SimCallResult) MarshalJSON() ([]byte, error) {
	type alias SimCallResult
	if r.Logs == nil {
		r.Logs = []*ethtypes.Log{}
	}
	return json.Marshal(alias(r))
}

// SimBlockResult is the envelope returned for each simulated block. `Block`
// carries the full block payload (populated via `RPCMarshalBlock` by the
// driver) and is marshaled at the top level by the custom `MarshalJSON`
// below; the `json:"-"` tag keeps the default encoder from emitting a
// nested `block` key.
type SimBlockResult struct {
	Block map[string]interface{} `json:"-"`
	Calls []SimCallResult        `json:"calls"`
}

// MarshalJSON embeds the block fields at the top level and appends
// `calls`, matching the execution-apis response shape.
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
