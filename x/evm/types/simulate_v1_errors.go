package types

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// Spec-reserved JSON-RPC error codes for eth_simulateV1. Names and values
// mirror the execution-apis spec; see the canonical error list (message
// text + condition) at:
//
//	https://github.com/ethereum/execution-apis/blob/main/src/eth/execute.yaml
const (
	// SimErrCodeReverted pins to the spec's enforced `const: 3` for
	// CallResultFailure.error.code (execute.yaml schema). Geth diverges
	// here and reuses its legacy -32000 from eth_call; reth follows the
	// spec. We follow the spec / reth.
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
