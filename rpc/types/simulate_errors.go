package types

// Spec-reserved JSON-RPC error codes for `eth_simulateV1`, matching the
// execution-apis spec and go-ethereum's `internal/ethapi/errors.go`.
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

// JSONRPCError is a structured JSON-RPC error carrying a numeric code plus
// optional data payload. Geth's rpc package populates the wire-level
// `{code, message, data}` fields from `ErrorCode()` / `ErrorData()` methods.
type JSONRPCError struct {
	Code    int
	Message string
	Data    interface{}
}

func (e *JSONRPCError) Error() string          { return e.Message }
func (e *JSONRPCError) ErrorCode() int         { return e.Code }
func (e *JSONRPCError) ErrorData() interface{} { return e.Data }

// NewSimNotImplementedError returns a `-32601` "method not found" error.
func NewSimNotImplementedError() *JSONRPCError {
	return &JSONRPCError{
		Code:    SimErrCodeMethodNotFound,
		Message: "eth_simulateV1 is not implemented",
	}
}
