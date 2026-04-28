package transfertracer

import (
	"errors"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/stretchr/testify/require"

	evmtypes "github.com/mezo-org/mezod/x/evm/types"
)

const fxBlockNumber = uint64(100)

var (
	fxBlockHash = common.HexToHash("0xb10cb10cb10cb10cb10cb10cb10cb10cb10cb10cb10cb10cb10cb10cb10cb10c")
	fxTxHash    = common.HexToHash("0x7a7a7a7a7a7a7a7a7a7a7a7a7a7a7a7a7a7a7a7a7a7a7a7a7a7a7a7a7a7a7a7a")
	fxTxIdx     = 3

	fxFrom    = common.HexToAddress("0xaaaa000000000000000000000000000000000001")
	fxTo      = common.HexToAddress("0xbbbb000000000000000000000000000000000002")
	fxNested1 = common.HexToAddress("0xcccc000000000000000000000000000000000003")
	fxNested2 = common.HexToAddress("0xdddd000000000000000000000000000000000004")
	fxValue   = big.NewInt(123)
)

// newFixtureTracer builds a Tracer with traceTransfers enabled and the
// per-tx context primed. Each test that needs a fresh frame stack
// constructs its own.
func newFixtureTracer(t *testing.T, traceTransfers bool) *Tracer {
	t.Helper()
	tt := New(traceTransfers, fxBlockNumber, fxBlockHash)
	tt.Reset(fxTxHash, fxTxIdx)
	return tt
}

// TestTracer_PlainValueTransfer covers the happy path: a single CALL
// frame at depth 0 with non-zero value emits exactly one synthetic
// ERC-7528 log carrying the standard Transfer topic, indexed sender /
// recipient, and the value as 32-byte data. The captured log also
// inherits every per-block / per-tx context field the tracer holds.
func TestTracer_PlainValueTransfer(t *testing.T) {
	tt := newFixtureTracer(t, true)

	tt.Hooks().OnEnter(0, byte(vm.CALL), fxFrom, fxTo, nil, 0, fxValue)
	tt.Hooks().OnExit(0, nil, 0, nil, false)

	logs := tt.Logs()
	require.Len(t, logs, 1, "expected one synthetic log for the value-bearing CALL")
	log := logs[0]

	require.Equal(t, TransferAddress, log.Address, "synthetic log must use the ERC-7528 pseudo-address")
	require.Equal(t, []common.Hash{
		TransferTopic,
		common.BytesToHash(fxFrom.Bytes()),
		common.BytesToHash(fxTo.Bytes()),
	}, log.Topics)
	require.Equal(t, common.BigToHash(fxValue).Bytes(), log.Data)

	// Stamped block / tx context.
	require.Equal(t, fxBlockNumber, log.BlockNumber)
	require.Equal(t, fxBlockHash, log.BlockHash)
	require.Equal(t, fxTxHash, log.TxHash)
	require.Equal(t, uint(fxTxIdx), log.TxIndex) //nolint:gosec
	require.Equal(t, uint(0), log.Index)
}

// TestTracer_TraceTransfersFalseSuppressesSynthetic locks the boolean
// branch in captureTransfer: with traceTransfers=false a value-bearing
// CALL must NOT produce a synthetic log. Real EVM logs (delivered via
// OnLog) still flow through, so the tracer remains useful as a log
// capture for the driver even when synthetic emission is off.
func TestTracer_TraceTransfersFalseSuppressesSynthetic(t *testing.T) {
	tt := newFixtureTracer(t, false)

	tt.Hooks().OnEnter(0, byte(vm.CALL), fxFrom, fxTo, nil, 0, fxValue)
	// Real log emitted from inside the call: must still be captured.
	tt.Hooks().OnLog(&ethtypes.Log{
		Address: fxTo,
		Topics:  []common.Hash{common.HexToHash("0xdead")},
		Data:    []byte{0x01},
	})
	tt.Hooks().OnExit(0, nil, 0, nil, false)

	logs := tt.Logs()
	require.Len(t, logs, 1, "only the real EVM log should remain — synthetic must be suppressed")
	require.Equal(t, fxTo, logs[0].Address)
	require.NotEqual(t, TransferAddress, logs[0].Address)
}

// TestTracer_DelegateCallNoSyntheticLog pins the DELEGATECALL exclusion:
// the value belongs to the caller's frame, so no synthetic Transfer is
// emitted for the delegated edge.
func TestTracer_DelegateCallNoSyntheticLog(t *testing.T) {
	tt := newFixtureTracer(t, true)

	tt.Hooks().OnEnter(0, byte(vm.CALL), fxFrom, fxTo, nil, 0, fxValue)
	tt.Hooks().OnEnter(1, byte(vm.DELEGATECALL), fxTo, fxNested1, nil, 0, fxValue)
	tt.Hooks().OnExit(1, nil, 0, nil, false)
	tt.Hooks().OnExit(0, nil, 0, nil, false)

	logs := tt.Logs()
	require.Len(t, logs, 1, "only the outer CALL must produce a synthetic log; DELEGATECALL must not")
	require.Equal(t, common.BytesToHash(fxTo.Bytes()), logs[0].Topics[2],
		"synthetic log recipient must be the outer CALL's `to`, not the DELEGATECALL target")
}

// TestTracer_StaticCallNoSyntheticLog pins the STATICCALL guard. The
// EVM rejects value transfer on STATICCALL before the tracer fires in
// real execution, but the code carries a defensive guard regardless;
// this test prevents a refactor from dropping it.
func TestTracer_StaticCallNoSyntheticLog(t *testing.T) {
	tt := newFixtureTracer(t, true)

	tt.Hooks().OnEnter(0, byte(vm.STATICCALL), fxFrom, fxTo, nil, 0, fxValue)
	tt.Hooks().OnExit(0, nil, 0, nil, false)

	require.Empty(t, tt.Logs(), "STATICCALL with value must never yield a synthetic log")
}

// TestTracer_CallCodeAndCreatesEmitSynthetic locks the exclusion list
// to exactly {DELEGATECALL, STATICCALL}. CALLCODE, CREATE, and CREATE2
// all carry real value movement and must produce a synthetic log when
// value > 0.
func TestTracer_CallCodeAndCreatesEmitSynthetic(t *testing.T) {
	cases := []struct {
		name string
		op   vm.OpCode
	}{
		{"CALLCODE", vm.CALLCODE},
		{"CREATE", vm.CREATE},
		{"CREATE2", vm.CREATE2},
		{"SELFDESTRUCT", vm.SELFDESTRUCT},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			tt := newFixtureTracer(t, true)
			tt.Hooks().OnEnter(0, byte(c.op), fxFrom, fxTo, nil, 0, fxValue)
			tt.Hooks().OnExit(0, nil, 0, nil, false)

			logs := tt.Logs()
			require.Len(t, logs, 1, "%s with value > 0 must emit a synthetic log", c.name)
			require.Equal(t, TransferAddress, logs[0].Address)
			require.Equal(t, common.BigToHash(fxValue).Bytes(), logs[0].Data)
		})
	}
}

// TestTracer_ZeroValueNoSyntheticLog covers the value-sign guard: CALL
// with zero (or nil) value must not emit a synthetic log even when
// traceTransfers is on.
func TestTracer_ZeroValueNoSyntheticLog(t *testing.T) {
	tt := newFixtureTracer(t, true)

	tt.Hooks().OnEnter(0, byte(vm.CALL), fxFrom, fxTo, nil, 0, big.NewInt(0))
	tt.Hooks().OnExit(0, nil, 0, nil, false)
	require.Empty(t, tt.Logs(), "value=0 must not emit a synthetic log")

	tt = newFixtureTracer(t, true)
	tt.Hooks().OnEnter(0, byte(vm.CALL), fxFrom, fxTo, nil, 0, nil)
	tt.Hooks().OnExit(0, nil, 0, nil, false)
	require.Empty(t, tt.Logs(), "value=nil must not emit a synthetic log")
}

// TestTracer_NestedRevertDropsFrameLogs covers the per-frame revert
// rule: a 3-level call where the middle frame reverts must drop only
// the middle frame's logs (and any of its descendants merged before
// the revert hit), preserving the outer and innermost-success-before-
// revert frames as the merge order dictates.
//
// Layout:
//
//	depth 0  CALL value=v       -> emits synthetic log L0
//	depth 1  CALL value=v       -> emits synthetic log L1
//	depth 2  CALL value=v       -> emits synthetic log L2
//	depth 2  exit success       -> L2 merges into depth 1 frame
//	depth 1  exit REVERTED      -> L1 + L2 dropped
//	depth 0  exit success       -> L0 stays
//
// Final root frame = [L0]. Asserting recipient address pins which log
// survived (each frame uses a distinct fxTo).
func TestTracer_NestedRevertDropsFrameLogs(t *testing.T) {
	tt := newFixtureTracer(t, true)

	tt.Hooks().OnEnter(0, byte(vm.CALL), fxFrom, fxTo, nil, 0, fxValue)         // L0
	tt.Hooks().OnEnter(1, byte(vm.CALL), fxTo, fxNested1, nil, 0, fxValue)      // L1
	tt.Hooks().OnEnter(2, byte(vm.CALL), fxNested1, fxNested2, nil, 0, fxValue) // L2
	tt.Hooks().OnExit(2, nil, 0, nil, false)                                    // success
	tt.Hooks().OnExit(1, nil, 0, errors.New("revert"), true)                    // REVERTED — drops L1 + L2
	tt.Hooks().OnExit(0, nil, 0, nil, false)

	logs := tt.Logs()
	require.Len(t, logs, 1, "only the outermost frame's log must survive the middle-frame revert")
	require.Equal(t, common.BytesToHash(fxTo.Bytes()), logs[0].Topics[2],
		"surviving log must be L0 (recipient = fxTo); L1 (->fxNested1) and L2 (->fxNested2) must be dropped")
}

// TestTracer_TopLevelRevertDropsAll covers the depth=0 revert path:
// when the root frame reverts, Logs() returns nil for the whole tx.
func TestTracer_TopLevelRevertDropsAll(t *testing.T) {
	tt := newFixtureTracer(t, true)

	tt.Hooks().OnEnter(0, byte(vm.CALL), fxFrom, fxTo, nil, 0, fxValue)
	tt.Hooks().OnLog(&ethtypes.Log{Address: fxTo, Topics: []common.Hash{common.HexToHash("0xfeed")}})
	tt.Hooks().OnExit(0, nil, 0, errors.New("revert"), true)

	require.Nil(t, tt.Logs(), "root-frame revert must surface as nil logs")
}

// TestTracer_SelfdestructWithBalanceEmitsSynthetic: the EVM dispatches
// OnEnter with typ=byte(vm.SELFDESTRUCT) and the contract's balance as
// value when SELFDESTRUCT moves funds to a beneficiary. The tracer
// must treat this as a real value-transfer edge.
func TestTracer_SelfdestructWithBalanceEmitsSynthetic(t *testing.T) {
	tt := newFixtureTracer(t, true)
	balance := big.NewInt(99_999)

	tt.Hooks().OnEnter(0, byte(vm.CALL), fxFrom, fxTo, nil, 0, big.NewInt(0))
	tt.Hooks().OnEnter(1, byte(vm.SELFDESTRUCT), fxTo, fxNested1, nil, 0, balance)
	tt.Hooks().OnExit(1, nil, 0, nil, false)
	tt.Hooks().OnExit(0, nil, 0, nil, false)

	logs := tt.Logs()
	require.Len(t, logs, 1, "SELFDESTRUCT with balance must emit a synthetic log")
	require.Equal(t, common.BytesToHash(fxTo.Bytes()), logs[0].Topics[1],
		"sender topic must be the self-destructing contract")
	require.Equal(t, common.BytesToHash(fxNested1.Bytes()), logs[0].Topics[2],
		"recipient topic must be the SELFDESTRUCT beneficiary")
	require.Equal(t, common.BigToHash(balance).Bytes(), logs[0].Data)
}

// TestTracer_DenyListSuppressesSyntheticForAllPrecompiles iterates the
// canonical mezo precompile registry and asserts NO synthetic log is
// emitted when value flows TO any of those addresses. The mezo BTC and
// MEZO precompiles emit their own ERC-20 Transfer events from inside
// the precompile run; surfacing a second synthetic log from the tracer
// would double-count the value movement.
func TestTracer_DenyListSuppressesSyntheticForAllPrecompiles(t *testing.T) {
	require.NotEmpty(t, evmtypes.DefaultPrecompilesVersions,
		"fixture relies on at least one canonical mezo precompile being registered")

	for _, pv := range evmtypes.DefaultPrecompilesVersions {
		precompile := common.HexToAddress(pv.PrecompileAddress)
		t.Run(precompile.Hex(), func(t *testing.T) {
			tt := newFixtureTracer(t, true)
			tt.Hooks().OnEnter(0, byte(vm.CALL), fxFrom, precompile, nil, 0, fxValue)
			tt.Hooks().OnExit(0, nil, 0, nil, false)

			require.Empty(t, tt.Logs(),
				"value transfer TO mezo precompile %s must NOT emit a synthetic log; the precompile already emits its own Transfer",
				precompile.Hex(),
			)
		})
	}
}

// TestTracer_RealLogCapturedWithContext: a real EVM log piped via
// OnLog must be captured into the active frame stamped with the
// per-block / per-tx context the tracer holds. Pins the captureLog
// pathway used by the StateDB.SetLogger -> OnLog wiring in the
// simulate driver.
func TestTracer_RealLogCapturedWithContext(t *testing.T) {
	tt := newFixtureTracer(t, true)
	emitted := &ethtypes.Log{
		Address: fxTo,
		Topics:  []common.Hash{common.HexToHash("0xfeedface")},
		Data:    []byte{0xAB, 0xCD},
	}

	tt.Hooks().OnEnter(0, byte(vm.CALL), fxFrom, fxTo, nil, 0, big.NewInt(0))
	tt.Hooks().OnLog(emitted)
	tt.Hooks().OnExit(0, nil, 0, nil, false)

	logs := tt.Logs()
	require.Len(t, logs, 1)
	got := logs[0]
	require.Equal(t, emitted.Address, got.Address)
	require.Equal(t, emitted.Topics, got.Topics)
	require.Equal(t, emitted.Data, got.Data)
	require.Equal(t, fxBlockNumber, got.BlockNumber)
	require.Equal(t, fxBlockHash, got.BlockHash)
	require.Equal(t, fxTxHash, got.TxHash)
	require.Equal(t, uint(fxTxIdx), got.TxIndex) //nolint:gosec
}

// TestTracer_LogIndexMonotonicAcrossReset pins the documented invariant:
// Reset clears per-tx log buffers but does NOT reset the request-scoped
// log index counter. Across two simulated calls (Reset between them),
// the captured logs must carry monotonically increasing Index values
// starting from 0.
func TestTracer_LogIndexMonotonicAcrossReset(t *testing.T) {
	tt := newFixtureTracer(t, true)

	tt.Hooks().OnEnter(0, byte(vm.CALL), fxFrom, fxTo, nil, 0, fxValue)
	tt.Hooks().OnExit(0, nil, 0, nil, false)
	first := tt.Logs()
	require.Len(t, first, 1)
	require.Equal(t, uint(0), first[0].Index, "first captured log starts the request-scoped counter at 0")

	// New tx within the same simulated request: per-tx context flips,
	// per-tx logs clear, but the index counter must keep climbing.
	nextTxHash := common.HexToHash("0xbeefbeefbeefbeefbeefbeefbeefbeefbeefbeefbeefbeefbeefbeefbeefbeef")
	tt.Reset(nextTxHash, fxTxIdx+1)
	require.Empty(t, tt.Logs(), "Reset must clear the per-tx log buffer")

	tt.Hooks().OnEnter(0, byte(vm.CALL), fxFrom, fxTo, nil, 0, fxValue)
	tt.Hooks().OnExit(0, nil, 0, nil, false)
	second := tt.Logs()
	require.Len(t, second, 1)
	require.Equal(t, uint(1), second[0].Index,
		"index counter must persist across Reset so logs stay monotonic across the simulated request")
	require.Equal(t, nextTxHash, second[0].TxHash, "Reset updates the stamped TxHash")
	require.Equal(t, uint(fxTxIdx+1), second[0].TxIndex, "Reset updates the stamped TxIndex") //nolint:gosec
}

// TestTracer_LogsBeforeAnyFrameOpened: calling Logs() before any
// OnEnter must return nil rather than panicking. Defensive coverage of
// the empty-stack branch in Logs().
func TestTracer_LogsBeforeAnyFrameOpened(t *testing.T) {
	tt := New(true, fxBlockNumber, fxBlockHash)
	require.Nil(t, tt.Logs())
}

// TestTracer_HooksAndTracerWired sanity-checks that Hooks() and Tracer()
// expose non-nil callbacks. The simulate driver relies on Hooks() for
// the StateDB.SetLogger path and Tracer() for the
// applyMessageWithConfig path; a nil hook pointer in either would
// silently disable the tracer at runtime.
func TestTracer_HooksAndTracerWired(t *testing.T) {
	tt := New(true, fxBlockNumber, fxBlockHash)
	hooks := tt.Hooks()
	require.NotNil(t, hooks)
	require.NotNil(t, hooks.OnEnter)
	require.NotNil(t, hooks.OnExit)
	require.NotNil(t, hooks.OnLog)

	wrapped := tt.Tracer()
	require.NotNil(t, wrapped)
	require.NotNil(t, wrapped.Hooks)
	require.NotNil(t, wrapped.Hooks.OnEnter)
	require.NotNil(t, wrapped.Hooks.OnExit)
	require.NotNil(t, wrapped.Hooks.OnLog)
}
