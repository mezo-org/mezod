package types

// Spec-reserved JSON-RPC error codes introduced by the `eth_simulateV1`
// execution-apis spec (`src/eth/execute.yaml`). The -38xxx values appear
// in the simulate schema; the -32xxx values are the subset used inside
// simulate call paths. Kept together here so the simulate driver and
// RPC layer can branch on a single coded vocabulary.
//
// Every constant below mirrors the spec's textual description exactly —
// the name describes the condition the spec assigns to that code, not
// a mezo-internal interpretation. See `ethereum/execution-apis` ref.
const (
	// -38010: "Transactions nonce is too low"
	SimErrCodeNonceTooLow = -38010
	// -38011: "Transactions nonce is too high"
	SimErrCodeNonceTooHigh = -38011
	// -38012: "Transactions baseFeePerGas is too low" — emitted when a
	// tx's effective gas price falls below the block's (overridden or
	// derived) baseFee under `validation=true`.
	SimErrCodeBaseFeeTooLow = -38012
	// -38013: "Not enough gas provided to pay for intrinsic gas".
	SimErrCodeIntrinsicGas = -38013
	// -38014: "Insufficient funds to pay for gas fees and value".
	SimErrCodeInsufficientFunds = -38014
	// -38015: "Block gas limit exceeded by the block's transactions" —
	// cumulative `gasUsed` inside a simulated block overflowed the
	// block's gasLimit.
	SimErrCodeBlockGasLimitReached = -38015
	// -38020: "Block number in sequence did not increase".
	SimErrCodeBlockNumberInvalid = -38020
	// -38021: "Block timestamp in sequence did not increase or stay the same".
	SimErrCodeBlockTimestampInvalid = -38021
	// -38022: "MovePrecompileToAddress referenced itself in replacement"
	// — the override sets `movePrecompileToAddress` equal to the
	// account's own address.
	SimErrCodeMovePrecompileSelfReference = -38022
	// -38023: "Multiple MovePrecompileToAddress referencing the same
	// address to replace" — two separate state overrides point their
	// `movePrecompileToAddress` at the same destination.
	SimErrCodeMovePrecompileDuplicateDest = -38023
	// -38024: "Sender is not an EOA" — unused by mezo simulate (overrides
	// may well install code at the `from` address), kept for spec parity.
	SimErrCodeSenderIsNotEOA = -38024
	// -38025: "Max init code size exceeded" (EIP-3860).
	SimErrCodeMaxInitCodeSizeExceeded = -38025
	// -38026: "Client adjustable limit exceeded" — span of simulated
	// blocks above the 256 cap, or other operator-configured bound.
	SimErrCodeClientLimitExceeded = -38026

	// -32005: "Transactions maxFeePerGas is too low" — `gasFeeCap` below
	// the derived block baseFee under `validation=true`.
	SimErrCodeFeeCapTooLow = -32005
	// -32015: "Execution error" — VM-level failure (invalid opcode, OOG,
	// etc.). Per-call error with `message` prefix `"vm execution error"`.
	SimErrCodeVMError = -32015
	// -32016: "Timeout" — per-request evaluation exceeded `RPCEVMTimeout`.
	SimErrCodeTimeout = -32016
	// -32601: "Method not found" — JSON-RPC 2.0 reserved.
	SimErrCodeMethodNotFound = -32601
	// -32602: "Missing or invalid parameters" — JSON-RPC 2.0 reserved;
	// spec assigns it to `eth_simulateV1` too. Used for move-precompile
	// validation failures that the spec's -38xxx range doesn't cover
	// (e.g., source address is not a precompile), matching go-ethereum.
	SimErrCodeInvalidParams = -32602
	// -32603: "The Ethereum node encountered an internal error".
	SimErrCodeInternalError = -32603

	// 3: "execution reverted" — EIP-140 alignment; matches the spec's
	// `CallResultFailure.error.code = 3` pin for reverted calls.
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

// NewSimNotImplementedError returns a `-32603` "internal error" carrying
// a "not yet implemented" message, used by the `eth_simulateV1` Phase 1
// stub. -32603 is the spec-listed internal-error code for the method
// (`src/eth/execute.yaml`), and is a better fit than -32601 because the
// method IS registered on the `eth_` namespace — returning "method not
// found" would mislead clients that cache unavailable-method signals.
func NewSimNotImplementedError() *RPCError {
	return &RPCError{
		Code:    SimErrCodeInternalError,
		Message: "eth_simulateV1 is not yet implemented",
	}
}
