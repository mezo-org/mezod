package keeper_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"sync"
	"testing"

	rpctypes "github.com/mezo-org/mezod/rpc/types"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/status"
)

// validSimErrCodes enumerates every spec-reserved JSON-RPC error code the
// simulate driver is allowed to surface on a SimulateV1Response.Error
// field. Any value outside this set is treated as a leak of an
// unstructured / unclassified failure and fails the fuzz invariant.
//
// Update this set whenever a new SimErrCode* is added in
// x/evm/types/simulate_v1_errors.go. Forgetting to update will cause the
// fuzz target to flag legitimate codes as leaks. SimErrCodeMethodNotFound
// is deliberately absent: it is fired only by the JSON-RPC kill switch,
// and the keeper handler should never surface it. If it does, the fuzz
// target should flag it as a layering bug.
var validSimErrCodes = map[int32]struct{}{
	evmtypes.SimErrCodeReverted:                    {},
	evmtypes.SimErrCodeFeeCapTooLow:                {},
	evmtypes.SimErrCodeVMError:                     {},
	evmtypes.SimErrCodeTimeout:                     {},
	evmtypes.SimErrCodeInvalidParams:               {},
	evmtypes.SimErrCodeInternalError:               {},
	evmtypes.SimErrCodeNonceTooLow:                 {},
	evmtypes.SimErrCodeNonceTooHigh:                {},
	evmtypes.SimErrCodeBaseFeeTooLow:               {},
	evmtypes.SimErrCodeIntrinsicGas:                {},
	evmtypes.SimErrCodeInsufficientFunds:           {},
	evmtypes.SimErrCodeBlockGasLimitReached:        {},
	evmtypes.SimErrCodeBlockNumberInvalid:          {},
	evmtypes.SimErrCodeBlockTimestampInvalid:       {},
	evmtypes.SimErrCodeMovePrecompileSelfReference: {},
	evmtypes.SimErrCodeMovePrecompileDuplicateDest: {},
	evmtypes.SimErrCodeSenderIsNotEOA:              {},
	evmtypes.SimErrCodeMaxInitCodeSizeExceeded:     {},
	evmtypes.SimErrCodeClientLimitExceeded:         {},
}

// fuzzSuite holds the lazily-initialized KeeperTestSuite shared across
// fuzz iterations. Setting up the full app (genesis, keepers, EVM
// config) costs hundreds of milliseconds; doing it once and reusing
// the keeper for read-mostly simulate calls keeps fuzz throughput
// usable. The simulate driver does not commit to underlying state, so
// reuse across iterations is safe.
var (
	fuzzSuiteOnce sync.Once
	fuzzSuite     *KeeperTestSuite
	fuzzSuiteErr  error
)

// getFuzzSuite returns the shared test suite, initializing it on first
// call. The supplied *testing.T is wired into the suite so testify-
// style Require() / Assert() calls inside the suite's setup do not
// panic with "'Require' must not be called before 'Run' or 'SetT'".
// Subsequent fuzz iterations re-bind the suite's T to the active
// fuzz subtest's T to preserve correct failure attribution.
func getFuzzSuite(t *testing.T) *KeeperTestSuite {
	fuzzSuiteOnce.Do(func() {
		// Use a defer/recover so a setup panic does not leave
		// fuzzSuite nil while sync.Once is satisfied — that would
		// hide the real failure mode behind a useless nil-pointer
		// dereference on every subsequent iteration.
		defer func() {
			if r := recover(); r != nil {
				fuzzSuiteErr = fmt.Errorf("KeeperTestSuite setup panicked: %v", r)
			}
		}()
		s := new(KeeperTestSuite)
		s.enableFeemarket = false
		s.enableLondonHF = true
		// testify's Suite.Require() requires the suite's T to be
		// set first; SetupTestWithT calls Require() internally.
		s.SetT(t)
		s.SetupTestWithT(t)
		fuzzSuite = s
	})
	if fuzzSuiteErr != nil {
		t.Fatalf("fuzz suite unavailable: %v", fuzzSuiteErr)
	}
	// Re-bind T so the active fuzz subtest owns suite.Require().
	fuzzSuite.SetT(t)
	return fuzzSuite
}

// fuzzSimulateV1Request mirrors KeeperTestSuite.simulateV1Request but
// without testify dependencies, since fuzz iterations run under
// *testing.T inside the f.Fuzz callback rather than the suite's own
// T. Marshaling errors are reported via tb.Fatalf — they are not part
// of the fuzz invariant.
func fuzzSimulateV1Request(tb testing.TB, suite *KeeperTestSuite, optsJSON []byte) *evmtypes.SimulateV1Request {
	tb.Helper()

	bn := rpctypes.BlockNumber(suite.ctx.BlockHeight())
	bnh := rpctypes.BlockNumberOrHash{BlockNumber: &bn}
	bnhBz, err := json.Marshal(bnh)
	if err != nil {
		tb.Fatalf("marshal block number: %v", err)
	}

	return &evmtypes.SimulateV1Request{
		Opts:              optsJSON,
		BlockNumberOrHash: bnhBz,
		GasCap:            21_000_000,
		ProposerAddress:   sdk.ConsAddress(suite.ctx.BlockHeader().ProposerAddress),
		ChainId:           suite.app.EvmKeeper.ChainID().Int64(),
	}
}

// FuzzSimulateV1Opts mutates the JSON `opts` payload of an
// eth_simulateV1 request and asserts two invariants on every input:
//
//  1. evmtypes.UnmarshalSimOpts and the public Keeper.SimulateV1
//     handler never panic.
//  2. When UnmarshalSimOpts accepts the input, the handler returns a
//     non-nil *SimulateV1Response whose Error field is either nil or a
//     *SimError whose Code is one of the spec-reserved
//     SimErrCode* constants. A gRPC codes.Internal status from the
//     handler is treated as an unstructured leak of an internal error
//     and fails the fuzz.
//
// Seed corpus comes from x/evm/keeper/testdata/simulate_v1_fuzz/, which
// holds the params[0] payload of every request line in the
// ethereum/execution-apis eth_simulateV1 fixtures (91 entries) plus a
// handful of hand-picked corner cases added via f.Add below.
func FuzzSimulateV1Opts(f *testing.F) {
	// Seed: every harvested execution-apis fixture.
	corpusDir := filepath.Join("testdata", "simulate_v1_fuzz")
	entries, err := os.ReadDir(corpusDir)
	if err != nil {
		f.Fatalf("read corpus dir %s: %v", corpusDir, err)
	}
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(corpusDir, e.Name()))
		if err != nil {
			f.Fatalf("read corpus entry %s: %v", e.Name(), err)
		}
		f.Add(string(data))
	}

	// Hand-picked corner cases that exercise unmarshal-path edges the
	// upstream fixtures do not (empty bodies, oversized arrays, deeply
	// nested block override fields, EIP surfaces explicitly rejected
	// by mezo).
	for _, s := range []string{
		``,
		`{}`,
		`[]`,
		`null`,
		`{"blockStateCalls":[]}`,
		`{"blockStateCalls":null}`,
		`{"blockStateCalls":[{}]}`,
		`{"blockStateCalls":[{"calls":[]}]}`,
		`{"blockStateCalls":[{"calls":null}]}`,
		`{"blockStateCalls":[{"blockOverrides":{}}]}`,
		`{"blockStateCalls":[{"blockOverrides":{"beaconRoot":"0x0000000000000000000000000000000000000000000000000000000000000000"}}]}`,
		`{"blockStateCalls":[{"blockOverrides":{"withdrawals":[]}}]}`,
		`{"blockStateCalls":[{"blockOverrides":{"blobBaseFee":"0x1"}}]}`,
		`{"validation":true,"traceTransfers":true,"returnFullTransactions":true,"blockStateCalls":[]}`,
		`{"blockStateCalls":[{"calls":[{}]}]}`,
		`{"blockStateCalls":[{"calls":[{"from":"0x1111111111111111111111111111111111111111","to":"0x2222222222222222222222222222222222222222"}]}]}`,
	} {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, optsJSON string) {
		// Catch panics from either path so the fuzz harness reports
		// the offending input rather than crashing the worker. The
		// recovered value is logged with a stack trace so a real bug
		// surfaced by fuzzing carries its origin alongside the input.
		// We deliberately fail fast via t.Fatalf; if a panic occurs
		// mid-call the test writer should re-run after fixing the
		// input. The shared suite reuse is safe under no-panic
		// conditions; under a panic the test process exits and
		// `go test -fuzz` re-spawns a fresh worker.
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("panic for input %q: %v\n%s", optsJSON, r, debug.Stack())
			}
		}()

		// Parse path. Rejections from UnmarshalSimOpts are an
		// acceptable terminal state: the wire layer would surface
		// them as a structured -32602 to the caller. Tighter than the
		// handler-path check below — UnmarshalSimOpts only ever wraps
		// with NewSimInvalidParams, so anything else is a contract
		// violation in the unmarshaler.
		opts, err := evmtypes.UnmarshalSimOpts([]byte(optsJSON))
		if err != nil {
			var simErr *evmtypes.SimError
			if !errors.As(err, &simErr) {
				t.Fatalf("UnmarshalSimOpts returned non-SimError: %T %v", err, err)
			}
			if simErr.Code != evmtypes.SimErrCodeInvalidParams {
				t.Fatalf("UnmarshalSimOpts SimError.Code %d != %d (SimErrCodeInvalidParams)",
					simErr.Code, evmtypes.SimErrCodeInvalidParams)
			}
			return
		}
		_ = opts

		// Handler path. Single shared suite — see getFuzzSuite for
		// the rationale.
		suite := getFuzzSuite(t)
		req := fuzzSimulateV1Request(t, suite, []byte(optsJSON))

		resp, err := suite.app.EvmKeeper.SimulateV1(suite.ctx, req)
		if err != nil {
			// gRPC-level errors mean the handler refused to produce
			// a structured response. codes.Internal is the leak
			// path the fuzz target is meant to surface; other gRPC
			// codes (e.g. InvalidArgument for nil request, which is
			// unreachable here since req is always non-nil) are
			// also unexpected.
			if st, ok := status.FromError(err); ok {
				t.Fatalf("handler returned gRPC %s: %s (input=%q)",
					st.Code(), st.Message(), optsJSON)
			}
			t.Fatalf("handler returned non-status error %T: %v (input=%q)",
				err, err, optsJSON)
		}
		if resp == nil {
			t.Fatalf("handler returned nil response without error (input=%q)", optsJSON)
		}
		if resp.Error != nil {
			if _, ok := validSimErrCodes[resp.Error.Code]; !ok {
				t.Fatalf("response SimError.Code %d not in known set (msg=%q, input=%q)",
					resp.Error.Code, resp.Error.Message, optsJSON)
			}
			return
		}
		// Success path: the keeper marshals SimBlockResult slices to
		// JSON. Confirm the payload is valid JSON so a downstream
		// JSON-RPC layer would not choke.
		if len(resp.Result) > 0 && !json.Valid(resp.Result) {
			t.Fatalf("handler Result not valid JSON: %q (input=%q)", resp.Result, optsJSON)
		}
	})
}
