package types

import "errors"

// Sentinel errors returned by keeper-side validation of `eth_call` /
// `eth_simulateV1` state overrides.
//
// Callers surfacing errors on the JSON-RPC wire (see rpc/types) translate
// these with errors.Is into the spec-reserved JSON-RPC error codes from
// execution-apis/src/eth/execute.yaml. Keeper code stays free of the RPC
// error vocabulary.
var (
	// ErrOverrideStateAndStateDiff is returned when a single account
	// specifies both `state` (full replacement) and `stateDiff` (partial
	// patch) overrides. Mirrors go-ethereum's
	// "account has both 'state' and 'stateDiff'" guard in override.go.
	ErrOverrideStateAndStateDiff = errors.New(
		"account has both state and stateDiff overrides",
	)

	// ErrOverrideAccountTaintedByPrecompile is returned when an account
	// already reserved as a MovePrecompileTo destination in the current
	// request appears again as a regular override source. Mirrors
	// go-ethereum's `dirtyAddrs` check in override.go.
	ErrOverrideAccountTaintedByPrecompile = errors.New(
		"account has already been overridden by a precompile",
	)

	// ErrOverrideDestAlreadyOverridden is returned when a MovePrecompileTo
	// destination also carries a regular account override in the same
	// request. Mirrors go-ethereum's `diff.has(dst)` check in override.go.
	ErrOverrideDestAlreadyOverridden = errors.New(
		"destination account is already overridden",
	)

	// ErrOverrideMovePrecompileSelfReference is returned when
	// MovePrecompileToAddress names its own source address as the
	// destination. Surfaces as the spec code -38022 at the RPC layer.
	ErrOverrideMovePrecompileSelfReference = errors.New(
		"MovePrecompileToAddress referenced itself in replacement",
	)

	// ErrOverrideMovePrecompileDuplicateDest is returned when two or more
	// MovePrecompileTo entries target the same destination address.
	// Surfaces as the spec code -38023 at the RPC layer.
	ErrOverrideMovePrecompileDuplicateDest = errors.New(
		"multiple MovePrecompileToAddress entries reference the same destination",
	)

	// ErrOverrideMoveMezoCustomPrecompile is returned when a caller tries
	// to relocate one of the chain's custom precompiles (0x7b7c...). The
	// set is derived from DefaultPrecompilesVersions.
	ErrOverrideMoveMezoCustomPrecompile = errors.New(
		"cannot move mezo custom precompile",
	)

	// ErrOverrideNotAPrecompile is returned when MovePrecompileTo names a
	// source address that is neither a stdlib precompile for the active
	// fork nor a denylisted mezo custom precompile.
	ErrOverrideNotAPrecompile = errors.New(
		"account is not a precompile",
	)
)
