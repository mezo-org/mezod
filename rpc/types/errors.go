package types

// Spec-reserved JSON-RPC error codes introduced by the `eth_simulateV1`
// execution-apis spec. Every -38xxx value below appears in the simulate
// schema, and the -32xxx values are the subset used inside simulate call
// paths per go-ethereum's `internal/ethapi/errors.go`. Kept together here
// so the simulate driver and RPC layer can branch on a single coded
// vocabulary.
const (
	SimErrCodeNonceTooLow              = -38010
	SimErrCodeNonceTooHigh             = -38011
	SimErrCodeIntrinsicGas             = -38013
	SimErrCodeInsufficientFunds        = -38014
	SimErrCodeBlockGasLimitReached     = -38015
	SimErrCodeBlockNumberInvalid       = -38020
	SimErrCodeBlockTimestampInvalid    = -38021
	SimErrCodePrecompileNotMovable     = -38022
	SimErrCodeAddressAlreadyOverridden = -38023
	SimErrCodeSenderIsNotEOA           = -38024
	SimErrCodeMaxInitCodeSizeExceeded  = -38025
	SimErrCodeClientLimitExceeded      = -38026

	SimErrCodeReverted       = -32000
	SimErrCodeFeeCapTooLow   = -32005
	SimErrCodeVMError        = -32015
	SimErrCodeMethodNotFound = -32601
	SimErrCodeInvalidParams  = -32602
	SimErrCodeInternalError  = -32603
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

// NewSimNotImplementedError returns a `-32601` "method not found" error
// for the `eth_simulateV1` stub.
func NewSimNotImplementedError() *RPCError {
	return &RPCError{
		Code:    SimErrCodeMethodNotFound,
		Message: "eth_simulateV1 is not implemented",
	}
}
