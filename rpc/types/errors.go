package types

// Spec-reserved JSON-RPC error codes for `eth_simulateV1`. Names and
// values mirror the execution-apis spec; see the canonical error list
// (message text + condition) at:
//
//	https://github.com/ethereum/execution-apis/blob/main/src/eth/execute.yaml
//
// Also exported here is the generic `RPCError` type for any call path
// that needs to surface a structured `{code, message, data}` payload
// rather than collapsing to bare `-32000`.
const (
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

	SimErrCodeFeeCapTooLow   = -32005
	SimErrCodeVMError        = -32015
	SimErrCodeTimeout        = -32016
	SimErrCodeMethodNotFound = -32601
	SimErrCodeInvalidParams  = -32602
	SimErrCodeInternalError  = -32603

	// SimErrCodeReverted pins to the spec's enforced `const: 3` for
	// `CallResultFailure.error.code` (execute.yaml schema). Geth diverges
	// here and reuses its legacy `-32000` from `eth_call`; reth follows
	// the spec. We follow the spec / reth.
	SimErrCodeReverted = 3
)

// RPCError is a structured JSON-RPC error carrying a numeric code plus
// optional data payload. Geth's rpc package populates the wire-level
// `{code, message, data}` fields from `ErrorCode()` / `ErrorData()` methods.
type RPCError struct {
	Code    int
	Message string
	Data    interface{}
}

func (e *RPCError) Error() string          { return e.Message }
func (e *RPCError) ErrorCode() int         { return e.Code }
func (e *RPCError) ErrorData() interface{} { return e.Data }

// NewSimNotImplementedError is the Phase 1 stub response for
// `eth_simulateV1`: `-32603` (the spec-listed internal-error code for
// this method) carrying "not yet implemented". Preferred over `-32601`
// because the method IS registered on the `eth_` namespace.
func NewSimNotImplementedError() *RPCError {
	return &RPCError{
		Code:    SimErrCodeInternalError,
		Message: "eth_simulateV1 is not yet implemented",
	}
}
