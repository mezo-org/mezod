package types

import (
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// Spec-reserved JSON-RPC error codes for eth_simulateV1. Names and values
// mirror the geth execution spec (ethereum/execution-apis); see the
// canonical error list (message text + condition) at:
//
//	https://github.com/ethereum/execution-apis/blob/main/src/eth/execute.yaml
const (
	// SimErrCodeReverted pins to the spec's enforced `const: 3` for
	// CallResultFailure.error.code (geth execution spec's `execute.yaml`
	// schema). Geth diverges here and reuses its legacy -32000 from
	// eth_call; reth follows the spec. We follow the spec / reth.
	SimErrCodeReverted = 3

	SimErrCodeFeeCapTooLow   = -32005
	SimErrCodeVMError        = -32015
	SimErrCodeTimeout        = -32016
	SimErrCodeMethodNotFound = -32601
	SimErrCodeInvalidParams  = -32602
	SimErrCodeInternalError  = -32603

	SimErrCodeNonceTooLow                 = -38010
	SimErrCodeNonceTooHigh                = -38011
	SimErrCodeBaseFeeTooLow               = -38012
	SimErrCodeIntrinsicGas                = -38013
	SimErrCodeInsufficientFunds           = -38014
	SimErrCodeBlockGasLimitReached        = -38015
	SimErrCodeBlockNumberInvalid          = -38020
	SimErrCodeBlockTimestampInvalid       = -38021
	SimErrCodeMovePrecompileSelfReference = -38022
	SimErrCodeMovePrecompileDuplicateDest = -38023
	SimErrCodeSenderIsNotEOA              = -38024
	SimErrCodeMaxInitCodeSizeExceeded     = -38025
	SimErrCodeClientLimitExceeded         = -38026
)

// The methods below implement geth's JSON-RPC error interface
// (same shape as RevertError in errors.go) so the RPC server emits
// {code, message, data} verbatim when a *SimError bubbles to the
// top of the call stack.
//
// Never return a typed-nil *SimError through an `error`-typed
// variable: the interface-nil gotcha makes the caller's nil-check
// false even when the pointer is nil. Return untyped nil or a
// non-nil *SimError.

func (e *SimError) Error() string  { return e.Message }
func (e *SimError) ErrorCode() int { return int(e.Code) }

func (e *SimError) ErrorData() any {
	if e.Data == "" {
		return nil
	}
	return e.Data
}

// NewSimInvalidParams wraps a free-form message as an invalid-params
// failure (-32602). Used for catch-all parse/validation rejections.
func NewSimInvalidParams(message string) *SimError {
	return &SimError{Code: SimErrCodeInvalidParams, Message: message}
}

// NewSimInvalidBlockNumber reports a block-number regression in the
// simulated chain (-38020). `num` is the offending number, `prev` is the
// prior block's.
func NewSimInvalidBlockNumber(num, prev *big.Int) *SimError {
	return &SimError{
		Code: SimErrCodeBlockNumberInvalid,
		Message: fmt.Sprintf(
			"simulate: block numbers must be strictly increasing: %s <= %s",
			num.String(), prev.String(),
		),
	}
}

// NewSimInvalidBlockTimestamp reports a timestamp regression in the
// simulated chain (-38021).
func NewSimInvalidBlockTimestamp(ts, prev uint64) *SimError {
	return &SimError{
		Code: SimErrCodeBlockTimestampInvalid,
		Message: fmt.Sprintf(
			"simulate: block timestamps must be strictly increasing: %d <= %d",
			ts, prev,
		),
	}
}

// NewSimClientLimitExceeded reports that the requested simulated chain
// span exceeds maxSimulateBlocks (-38026).
func NewSimClientLimitExceeded(span *big.Int, maxSpan int64) *SimError {
	return &SimError{
		Code: SimErrCodeClientLimitExceeded,
		Message: fmt.Sprintf(
			"simulate: too many blocks: span %s > max %d",
			span.String(), maxSpan,
		),
	}
}

// NewSimBlockGasLimitReached reports that a call's requested or defaulted
// gas would push cumulative block gas past the simulated block's gas limit
// (-38015).
func NewSimBlockGasLimitReached(requested, remaining uint64) *SimError {
	return &SimError{
		Code: SimErrCodeBlockGasLimitReached,
		Message: fmt.Sprintf(
			"simulate: block gas limit reached: requested %d > remaining %d",
			requested, remaining,
		),
	}
}

// NewSimIntrinsicGas reports a call whose explicit gas limit falls below
// the call's intrinsic-gas requirement (-38013). `required` may be 0 when
// the catch site does not have access to the computed intrinsic value;
// the message stays informative as "have X, need 0".
func NewSimIntrinsicGas(provided, required uint64) *SimError {
	return &SimError{
		Code: SimErrCodeIntrinsicGas,
		Message: fmt.Sprintf(
			"simulate: intrinsic gas too low: have %d, need %d",
			provided, required,
		),
	}
}

// NewSimForkSpanUnsupported reports a simulated chain span that crosses a
// fork activation boundary, which mezod does not yet support. Reuses
// SimErrCodeClientLimitExceeded (-38026) — the spec does not reserve a
// fork-boundary code, and "client-imposed limit on fork-uniformity" is a
// defensible fit. Distinguishable from NewSimClientLimitExceeded by
// message text.
func NewSimForkSpanUnsupported() *SimError {
	return &SimError{
		Code:    SimErrCodeClientLimitExceeded,
		Message: "simulate: span crosses a fork boundary; not yet supported",
	}
}

// NewSimMovePrecompileSelfRef reports a MovePrecompileTo override whose
// destination is the source address (-38022).
func NewSimMovePrecompileSelfRef(addr common.Address) *SimError {
	return &SimError{
		Code: SimErrCodeMovePrecompileSelfReference,
		Message: fmt.Sprintf(
			"MovePrecompileToAddress referenced itself in replacement: %s",
			addr.Hex(),
		),
	}
}

// NewSimMovePrecompileDupDest reports two MovePrecompileTo overrides
// that target the same destination (-38023).
func NewSimMovePrecompileDupDest(dest common.Address) *SimError {
	return &SimError{
		Code: SimErrCodeMovePrecompileDuplicateDest,
		Message: fmt.Sprintf(
			"multiple MovePrecompileToAddress entries reference the same destination: %s",
			dest.Hex(),
		),
	}
}

// NewSimStateAndStateDiff reports an account specifying both `state`
// (full replacement) and `stateDiff` (partial patch) overrides (-32602).
func NewSimStateAndStateDiff(addr common.Address) *SimError {
	return NewSimInvalidParams(fmt.Sprintf(
		"account has both state and stateDiff overrides: %s", addr.Hex(),
	))
}

// NewSimAccountTainted reports an account that was already reserved as a
// MovePrecompileTo destination and then appears as a regular override
// source in the same request (-32602).
func NewSimAccountTainted(addr common.Address) *SimError {
	return NewSimInvalidParams(fmt.Sprintf(
		"account has already been overridden by a precompile: %s", addr.Hex(),
	))
}

// NewSimDestAlreadyOverridden reports a MovePrecompileTo destination
// that also carries a regular account override in the same request
// (-32602).
func NewSimDestAlreadyOverridden(dest common.Address) *SimError {
	return NewSimInvalidParams(fmt.Sprintf(
		"destination account is already overridden: %s", dest.Hex(),
	))
}

// NewSimMoveMezoCustom reports an attempt to relocate one of the mezo
// custom precompiles (0x7b7c…), forbidden by chain policy (-32602).
func NewSimMoveMezoCustom(addr common.Address) *SimError {
	return NewSimInvalidParams(fmt.Sprintf(
		"cannot move mezo custom precompile: %s", addr.Hex(),
	))
}

// NewSimNotAPrecompile reports a MovePrecompileTo source address that is
// neither a stdlib precompile for the active fork nor a denylisted mezo
// custom precompile (-32602).
func NewSimNotAPrecompile(addr common.Address) *SimError {
	return NewSimInvalidParams(fmt.Sprintf(
		"account is not a precompile: %s", addr.Hex(),
	))
}

// NewSimReverted reports an EVM revert. `data` is the raw revert-return
// bytes; the constructor hex-encodes them as required by the spec
// (CallResultFailure.error.data). Data is omitted when the revert has
// no payload.
func NewSimReverted(data []byte) *SimError {
	e := &SimError{Code: SimErrCodeReverted, Message: "execution reverted"}
	if len(data) > 0 {
		e.Data = hexutil.Encode(data)
	}
	return e
}

// NewSimVMError reports a non-revert VM failure (out of gas, invalid
// opcode, stack overflow, etc.) with the raw EVM error string
// (-32015).
func NewSimVMError(vmErr string) *SimError {
	return &SimError{Code: SimErrCodeVMError, Message: vmErr}
}

// NewSimMethodNotFound reports the method-disabled kill switch (-32601).
// The message matches the JSON-RPC framework's "method not registered"
// shape so an operator hiding the endpoint is indistinguishable from a
// node that does not implement it.
func NewSimMethodNotFound(method string) *SimError {
	return &SimError{
		Code:    SimErrCodeMethodNotFound,
		Message: fmt.Sprintf("the method %s does not exist/is not available", method),
	}
}

// NewSimCallLimitExceeded reports that the cumulative call count
// exceeds MaxSimulateCalls (-38026).
func NewSimCallLimitExceeded(total, maxCalls int) *SimError {
	return &SimError{
		Code: SimErrCodeClientLimitExceeded,
		Message: fmt.Sprintf(
			"client limit exceeded: %d calls > max %d",
			total, maxCalls,
		),
	}
}

// NewSimBlockCountExceeded reports that the number of submitted blocks
// exceeds MaxSimulateBlocks (-38026).
func NewSimBlockCountExceeded(total, maxBlocks int) *SimError {
	return &SimError{
		Code: SimErrCodeClientLimitExceeded,
		Message: fmt.Sprintf(
			"client limit exceeded: %d blocks > max %d",
			total, maxBlocks,
		),
	}
}

// NewSimTimeout reports that the request hit its evm-timeout deadline
// (-32016).
func NewSimTimeout(timeout time.Duration) *SimError {
	return &SimError{
		Code:    SimErrCodeTimeout,
		Message: fmt.Sprintf("execution aborted (timeout = %s)", timeout),
	}
}

// NewSimNonceTooLow reports a call whose nonce is below the sender's
// state nonce (-38010). Fires only when validation=true.
func NewSimNonceTooLow(addr common.Address, have, want uint64) *SimError {
	return &SimError{
		Code: SimErrCodeNonceTooLow,
		Message: fmt.Sprintf(
			"nonce too low: address %s, tx: %d state: %d",
			addr.Hex(), have, want,
		),
	}
}

// NewSimNonceTooHigh reports a call whose nonce exceeds the sender's
// state nonce (-38011). Fires only when validation=true.
func NewSimNonceTooHigh(addr common.Address, have, want uint64) *SimError {
	return &SimError{
		Code: SimErrCodeNonceTooHigh,
		Message: fmt.Sprintf(
			"nonce too high: address %s, tx: %d state: %d",
			addr.Hex(), have, want,
		),
	}
}

// NewSimBaseFeeTooLow reports that the caller-supplied
// BlockOverrides.BaseFeePerGas falls below the chain-computed
// eip1559.CalcBaseFee floor for the parent (-38012). Distinct from
// -32005, which targets the transaction's gasFeeCap rather than the
// block's overridden baseFee.
func NewSimBaseFeeTooLow(have, want *big.Int) *SimError {
	return &SimError{
		Code: SimErrCodeBaseFeeTooLow,
		Message: fmt.Sprintf(
			"base fee too low: override %s, floor %s",
			have.String(), want.String(),
		),
	}
}

// NewSimInsufficientFunds reports a call whose sender lacks the funds
// required to cover gasLimit*gasPrice + value (-38014).
func NewSimInsufficientFunds(addr common.Address, have, want *big.Int) *SimError {
	return &SimError{
		Code: SimErrCodeInsufficientFunds,
		Message: fmt.Sprintf(
			"insufficient funds: address %s have %s, want %s",
			addr.Hex(), have.String(), want.String(),
		),
	}
}

// NewSimInitcodeTooLarge reports a CREATE call whose init-code size
// exceeds params.MaxInitCodeSize on a Shanghai-active fork (-38025).
func NewSimInitcodeTooLarge(have, limit int) *SimError {
	return &SimError{
		Code: SimErrCodeMaxInitCodeSizeExceeded,
		Message: fmt.Sprintf(
			"max initcode size exceeded: code size %d limit %d",
			have, limit,
		),
	}
}

// NewSimFeeCapTooLow reports a call whose gasFeeCap is below the
// active block base fee (-32005). Fires only when validation=true.
func NewSimFeeCapTooLow(have, want *big.Int) *SimError {
	return &SimError{
		Code: SimErrCodeFeeCapTooLow,
		Message: fmt.Sprintf(
			"max fee per gas less than block base fee: maxFeePerGas: %s, baseFee: %s",
			have.String(), want.String(),
		),
	}
}
