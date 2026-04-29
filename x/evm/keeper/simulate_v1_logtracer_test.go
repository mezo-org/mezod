package keeper

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

func newFixtureTracer(t *testing.T, traceTransfers bool) *simTracer {
	t.Helper()
	tt := newSimTracer(traceTransfers, fxBlockNumber, fxBlockHash)
	tt.reset(fxTxHash, fxTxIdx)
	return tt
}

func TestSimTracer_PlainValueTransfer(t *testing.T) {
	tt := newFixtureTracer(t, true)

	tt.Hooks().OnEnter(0, byte(vm.CALL), fxFrom, fxTo, nil, 0, fxValue)
	tt.Hooks().OnExit(0, nil, 0, nil, false)

	logs := tt.Logs()
	require.Len(t, logs, 1)
	log := logs[0]

	require.Equal(t, simTransferAddress, log.Address)
	require.Equal(t, []common.Hash{
		simTransferTopic,
		common.BytesToHash(fxFrom.Bytes()),
		common.BytesToHash(fxTo.Bytes()),
	}, log.Topics)
	require.Equal(t, common.BigToHash(fxValue).Bytes(), log.Data)

	require.Equal(t, fxBlockNumber, log.BlockNumber)
	require.Equal(t, fxBlockHash, log.BlockHash)
	require.Equal(t, fxTxHash, log.TxHash)
	require.Equal(t, uint(fxTxIdx), log.TxIndex) //nolint:gosec
	require.Equal(t, uint(0), log.Index)
}

// With traceTransfers=false a value-bearing CALL must NOT produce a synthetic
// log, but real EVM logs delivered via OnLog must still flow through.
func TestSimTracer_TraceTransfersFalseSuppressesSynthetic(t *testing.T) {
	tt := newFixtureTracer(t, false)

	tt.Hooks().OnEnter(0, byte(vm.CALL), fxFrom, fxTo, nil, 0, fxValue)
	tt.Hooks().OnLog(&ethtypes.Log{
		Address: fxTo,
		Topics:  []common.Hash{common.HexToHash("0xdead")},
		Data:    []byte{0x01},
	})
	tt.Hooks().OnExit(0, nil, 0, nil, false)

	logs := tt.Logs()
	require.Len(t, logs, 1)
	require.Equal(t, fxTo, logs[0].Address)
	require.NotEqual(t, simTransferAddress, logs[0].Address)
}

// DELEGATECALL value belongs to the caller's frame, so no synthetic log is
// emitted for the delegated edge.
func TestSimTracer_DelegateCallNoSyntheticLog(t *testing.T) {
	tt := newFixtureTracer(t, true)

	tt.Hooks().OnEnter(0, byte(vm.CALL), fxFrom, fxTo, nil, 0, fxValue)
	tt.Hooks().OnEnter(1, byte(vm.DELEGATECALL), fxTo, fxNested1, nil, 0, fxValue)
	tt.Hooks().OnExit(1, nil, 0, nil, false)
	tt.Hooks().OnExit(0, nil, 0, nil, false)

	logs := tt.Logs()
	require.Len(t, logs, 1)
	require.Equal(t, common.BytesToHash(fxTo.Bytes()), logs[0].Topics[2])
}

// EVM rejects value transfer on STATICCALL before the tracer fires in real
// execution, but the tracer carries a defensive guard regardless; this test
// prevents a refactor from dropping it.
func TestSimTracer_StaticCallNoSyntheticLog(t *testing.T) {
	tt := newFixtureTracer(t, true)

	tt.Hooks().OnEnter(0, byte(vm.STATICCALL), fxFrom, fxTo, nil, 0, fxValue)
	tt.Hooks().OnExit(0, nil, 0, nil, false)

	require.Empty(t, tt.Logs())
}

// Locks the exclusion list to exactly {DELEGATECALL, STATICCALL}.
func TestSimTracer_CallCodeAndCreatesEmitSynthetic(t *testing.T) {
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
			require.Len(t, logs, 1)
			require.Equal(t, simTransferAddress, logs[0].Address)
			require.Equal(t, common.BigToHash(fxValue).Bytes(), logs[0].Data)
		})
	}
}

func TestSimTracer_ZeroValueNoSyntheticLog(t *testing.T) {
	tt := newFixtureTracer(t, true)

	tt.Hooks().OnEnter(0, byte(vm.CALL), fxFrom, fxTo, nil, 0, big.NewInt(0))
	tt.Hooks().OnExit(0, nil, 0, nil, false)
	require.Empty(t, tt.Logs())

	tt = newFixtureTracer(t, true)
	tt.Hooks().OnEnter(0, byte(vm.CALL), fxFrom, fxTo, nil, 0, nil)
	tt.Hooks().OnExit(0, nil, 0, nil, false)
	require.Empty(t, tt.Logs())
}

// Per-frame revert rule: a 3-level call where the middle frame reverts must
// drop only the middle frame's logs (and any descendants merged before the
// revert hit), preserving the outer frame.
//
//	depth 0  CALL value=v       -> emits synthetic log L0
//	depth 1  CALL value=v       -> emits synthetic log L1
//	depth 2  CALL value=v       -> emits synthetic log L2
//	depth 2  exit success       -> L2 merges into depth 1 frame
//	depth 1  exit REVERTED      -> L1 + L2 dropped
//	depth 0  exit success       -> L0 stays
func TestSimTracer_NestedRevertDropsFrameLogs(t *testing.T) {
	tt := newFixtureTracer(t, true)

	tt.Hooks().OnEnter(0, byte(vm.CALL), fxFrom, fxTo, nil, 0, fxValue)         // L0
	tt.Hooks().OnEnter(1, byte(vm.CALL), fxTo, fxNested1, nil, 0, fxValue)      // L1
	tt.Hooks().OnEnter(2, byte(vm.CALL), fxNested1, fxNested2, nil, 0, fxValue) // L2
	tt.Hooks().OnExit(2, nil, 0, nil, false)
	tt.Hooks().OnExit(1, nil, 0, errors.New("revert"), true)
	tt.Hooks().OnExit(0, nil, 0, nil, false)

	logs := tt.Logs()
	require.Len(t, logs, 1)
	require.Equal(t, common.BytesToHash(fxTo.Bytes()), logs[0].Topics[2],
		"surviving log must be L0 (recipient = fxTo); L1 and L2 must be dropped")
}

func TestSimTracer_TopLevelRevertDropsAll(t *testing.T) {
	tt := newFixtureTracer(t, true)

	tt.Hooks().OnEnter(0, byte(vm.CALL), fxFrom, fxTo, nil, 0, fxValue)
	tt.Hooks().OnLog(&ethtypes.Log{Address: fxTo, Topics: []common.Hash{common.HexToHash("0xfeed")}})
	tt.Hooks().OnExit(0, nil, 0, errors.New("revert"), true)

	require.Nil(t, tt.Logs())
}

// EVM dispatches OnEnter with typ=byte(vm.SELFDESTRUCT) and the contract's
// balance as value when SELFDESTRUCT moves funds to a beneficiary; the tracer
// must treat this as a real value-transfer edge.
func TestSimTracer_SelfdestructWithBalanceEmitsSynthetic(t *testing.T) {
	tt := newFixtureTracer(t, true)
	balance := big.NewInt(99_999)

	tt.Hooks().OnEnter(0, byte(vm.CALL), fxFrom, fxTo, nil, 0, big.NewInt(0))
	tt.Hooks().OnEnter(1, byte(vm.SELFDESTRUCT), fxTo, fxNested1, nil, 0, balance)
	tt.Hooks().OnExit(1, nil, 0, nil, false)
	tt.Hooks().OnExit(0, nil, 0, nil, false)

	logs := tt.Logs()
	require.Len(t, logs, 1)
	require.Equal(t, common.BytesToHash(fxTo.Bytes()), logs[0].Topics[1])
	require.Equal(t, common.BytesToHash(fxNested1.Bytes()), logs[0].Topics[2])
	require.Equal(t, common.BigToHash(balance).Bytes(), logs[0].Data)
}

// Mezo custom precompiles emit their own ERC-20 Transfer events, so a
// synthetic log on value flowing TO any of them would double-count.
func TestSimTracer_DenyListSuppressesSyntheticForAllPrecompiles(t *testing.T) {
	require.NotEmpty(t, evmtypes.DefaultPrecompilesVersions,
		"fixture relies on at least one canonical mezo precompile being registered")

	for _, pv := range evmtypes.DefaultPrecompilesVersions {
		precompile := common.HexToAddress(pv.PrecompileAddress)
		t.Run(precompile.Hex(), func(t *testing.T) {
			tt := newFixtureTracer(t, true)
			tt.Hooks().OnEnter(0, byte(vm.CALL), fxFrom, precompile, nil, 0, fxValue)
			tt.Hooks().OnExit(0, nil, 0, nil, false)

			require.Empty(t, tt.Logs())
		})
	}
}

// Real EVM logs piped via OnLog must be captured into the active frame
// stamped with the per-block / per-tx context the tracer holds.
func TestSimTracer_RealLogCapturedWithContext(t *testing.T) {
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

// reset clears per-tx log buffers but does NOT reset the request-scoped log
// index counter; pin that across two calls.
func TestSimTracer_LogIndexMonotonicAcrossReset(t *testing.T) {
	tt := newFixtureTracer(t, true)

	tt.Hooks().OnEnter(0, byte(vm.CALL), fxFrom, fxTo, nil, 0, fxValue)
	tt.Hooks().OnExit(0, nil, 0, nil, false)
	first := tt.Logs()
	require.Len(t, first, 1)
	require.Equal(t, uint(0), first[0].Index)

	nextTxHash := common.HexToHash("0xbeefbeefbeefbeefbeefbeefbeefbeefbeefbeefbeefbeefbeefbeefbeefbeef")
	tt.reset(nextTxHash, fxTxIdx+1)
	require.Empty(t, tt.Logs())

	tt.Hooks().OnEnter(0, byte(vm.CALL), fxFrom, fxTo, nil, 0, fxValue)
	tt.Hooks().OnExit(0, nil, 0, nil, false)
	second := tt.Logs()
	require.Len(t, second, 1)
	require.Equal(t, uint(1), second[0].Index,
		"index counter must persist across reset")
	require.Equal(t, nextTxHash, second[0].TxHash)
	require.Equal(t, uint(fxTxIdx+1), second[0].TxIndex) //nolint:gosec
}

func TestSimTracer_LogsBeforeAnyFrameOpened(t *testing.T) {
	tt := newSimTracer(true, fxBlockNumber, fxBlockHash)
	require.Nil(t, tt.Logs())
}

func TestSimTracer_HooksWired(t *testing.T) {
	tt := newSimTracer(true, fxBlockNumber, fxBlockHash)
	hooks := tt.Hooks()
	require.NotNil(t, hooks)
	require.NotNil(t, hooks.OnEnter)
	require.NotNil(t, hooks.OnExit)
	require.NotNil(t, hooks.OnLog)
}
