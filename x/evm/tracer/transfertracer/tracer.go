// Package transfertracer captures EVM logs and synthesises ERC-7528
// Transfer logs at the pseudo-address for native value-transfer call edges.
package transfertracer

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/tracing"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/eth/tracers"

	evmtypes "github.com/mezo-org/mezod/x/evm/types"
)

// TransferTopic is keccak256("Transfer(address,address,uint256)").
var TransferTopic = common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")

// TransferAddress is the ERC-7528 pseudo-address used as the emitter of
// synthetic native-value Transfer logs.
var TransferAddress = common.HexToAddress("0xEeeeeEeeeEeEeeEeEeEeeEEEeeeeEeeeeeeeEEeE")

// denyList is the set of addresses for which a synthetic Transfer log must
// not be emitted: each mezo custom precompile already emits its own ERC-20
// Transfer event, and the synthetic log would double-count.
var denyList = func() map[common.Address]struct{} {
	set := make(map[common.Address]struct{}, len(evmtypes.DefaultPrecompilesVersions))
	for _, pv := range evmtypes.DefaultPrecompilesVersions {
		set[common.HexToAddress(pv.PrecompileAddress)] = struct{}{}
	}
	return set
}()

// Tracer captures EVM logs across nested call frames. On call entry it
// pushes a fresh frame; on success it merges the frame's logs into the
// parent; on revert it drops the frame's logs. When traceTransfers is
// true, OnEnter additionally emits a synthetic ERC-7528 Transfer log for
// value-bearing CALL / CALLCODE / SELFDESTRUCT edges (DELEGATECALL is
// excluded because the value belongs to the caller's frame).
type Tracer struct {
	logs [][]*ethtypes.Log
	// count is preserved across Reset so log.Index stays monotonic across
	// calls within a single Tracer lifetime.
	count uint

	traceTransfers bool
	blockNumber    uint64
	blockHash      common.Hash
	txHash         common.Hash
	txIdx          uint
}

// New constructs a Tracer for one simulated block.
func New(traceTransfers bool, blockNumber uint64, blockHash common.Hash) *Tracer {
	return &Tracer{
		traceTransfers: traceTransfers,
		blockNumber:    blockNumber,
		blockHash:      blockHash,
	}
}

// Hooks returns the tracing hook surface; the same value drives both
// vm.Config.Tracer (OnEnter / OnExit) and statedb.SetLogger (OnLog).
func (t *Tracer) Hooks() *tracing.Hooks {
	return &tracing.Hooks{
		OnEnter: t.onEnter,
		OnExit:  t.onExit,
		OnLog:   t.onLog,
	}
}

// Tracer wraps Hooks() in the eth/tracers.Tracer envelope expected by the
// keeper's applyMessageWithConfig.
func (t *Tracer) Tracer() *tracers.Tracer {
	return &tracers.Tracer{Hooks: t.Hooks()}
}

// Reset clears the per-tx log buffers and updates the tx-context fields
// stamped onto subsequent captured logs. count is preserved.
func (t *Tracer) Reset(txHash common.Hash, txIdx int) {
	t.logs = nil
	t.txHash = txHash
	t.txIdx = uint(txIdx) //nolint:gosec
}

// Logs returns the root frame's accumulated logs. Returns nil if the root
// frame reverted, or before any frame opened.
func (t *Tracer) Logs() []*ethtypes.Log {
	if len(t.logs) == 0 {
		return nil
	}
	return t.logs[0]
}

func (t *Tracer) onEnter(_ int, typ byte, from common.Address, to common.Address, _ []byte, _ uint64, value *big.Int) {
	t.logs = append(t.logs, []*ethtypes.Log{})
	if value == nil || value.Sign() <= 0 {
		return
	}
	op := vm.OpCode(typ)
	if op == vm.DELEGATECALL || op == vm.STATICCALL {
		return
	}
	if _, deny := denyList[to]; deny {
		return
	}
	t.captureTransfer(from, to, value)
}

func (t *Tracer) onExit(depth int, _ []byte, _ uint64, _ error, reverted bool) {
	if depth == 0 {
		if reverted && len(t.logs) > 0 {
			t.logs[0] = nil
		}
		return
	}
	size := len(t.logs)
	if size <= 1 {
		return
	}
	frame := t.logs[size-1]
	t.logs = t.logs[:size-1]
	if !reverted {
		t.logs[size-2] = append(t.logs[size-2], frame...)
	}
}

func (t *Tracer) onLog(log *ethtypes.Log) {
	t.captureLog(log.Address, log.Topics, log.Data)
}

func (t *Tracer) captureLog(address common.Address, topics []common.Hash, data []byte) {
	if len(t.logs) == 0 {
		return
	}
	idx := len(t.logs) - 1
	t.logs[idx] = append(t.logs[idx], &ethtypes.Log{
		Address:     address,
		Topics:      topics,
		Data:        data,
		BlockNumber: t.blockNumber,
		BlockHash:   t.blockHash,
		TxHash:      t.txHash,
		TxIndex:     t.txIdx,
		Index:       t.count,
	})
	t.count++
}

func (t *Tracer) captureTransfer(from, to common.Address, value *big.Int) {
	if !t.traceTransfers {
		return
	}
	topics := []common.Hash{
		TransferTopic,
		common.BytesToHash(from.Bytes()),
		common.BytesToHash(to.Bytes()),
	}
	t.captureLog(TransferAddress, topics, common.BigToHash(value).Bytes())
}
