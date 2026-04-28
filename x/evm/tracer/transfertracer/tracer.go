// Package transfertracer implements the eth_simulateV1 ERC-7528 transfer
// tracer: a per-frame log capture that records every real EVM log and,
// when traceTransfers is enabled, emits synthetic ERC-20 Transfer logs at
// the ERC-7528 pseudo-address for native value-transfer call edges.
//
// Direct port of go-ethereum's internal/ethapi/logtracer.go, adapted to:
//
//   - mezo's per-call Reset (request-scoped log index counter, mirroring
//     geth's choice not to reset the counter on Reset);
//   - mezo's custom precompiles at 0x7b7c…00..15, which already emit
//     their own ERC-20 Transfer events on value-transfer call edges;
//     to avoid double-counting, the tracer skips synthetic emission when
//     the call target is one of those precompiles.
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

var (
	// transferTopic is keccak256("Transfer(address,address,uint256)").
	transferTopic = common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")

	// transferAddress is the ERC-7528 pseudo-address used as the emitter
	// of synthetic native-value Transfer logs.
	transferAddress = common.HexToAddress("0xEeeeeEeeeEeEeeEeEeEeeEEEeeeeEeeeeeeeEEeE")

	// denyList is the set of mezo custom precompile addresses that the
	// tracer must NOT emit a synthetic Transfer log for. These precompiles
	// emit their own ERC-20 Transfer event from inside the precompile run
	// (see precompile/erc20/transfer.go), and the synthetic log would
	// double-count the value movement.
	denyList = func() map[common.Address]struct{} {
		set := make(map[common.Address]struct{}, len(evmtypes.DefaultPrecompilesVersions))
		for _, pv := range evmtypes.DefaultPrecompilesVersions {
			set[common.HexToAddress(pv.PrecompileAddress)] = struct{}{}
		}
		return set
	}()
)

// Tracer captures EVM logs across nested call frames. On call entry it
// pushes a fresh frame; on success it merges the frame's logs into the
// parent; on revert it drops the frame's logs (matching real-chain log
// semantics). When traceTransfers is true, OnEnter additionally emits a
// synthetic ERC-7528 Transfer log for value-bearing CALL / CALLCODE /
// SELFDESTRUCT edges (DELEGATECALL is excluded because the value belongs
// to the caller's frame).
type Tracer struct {
	// logs holds the per-frame log buffers; the active frame is at the
	// top of the slice. A request that revert at the root flips logs[0]
	// to nil (drop everything).
	logs [][]*ethtypes.Log

	// count is the request-scoped log index counter. Mirrors the geth
	// reference impl: Reset clears per-tx logs but does NOT reset count,
	// so log.Index is monotonic across calls within a single Tracer
	// lifetime.
	count uint

	traceTransfers bool
	blockNumber    uint64
	// blockTimestamp matches geth's logtracer field; ethtypes.Log on
	// the v1.14.8 fork mezod tracks does not yet carry a BlockTimestamp,
	// so the field is accepted at construction but not stamped onto
	// captured logs. Kept as a struct field for forward-compatibility
	// with the v1.16.9 upgrade.
	blockTimestamp uint64 //nolint:unused
	blockHash      common.Hash
	txHash         common.Hash
	txIdx          uint
}

// New constructs a Tracer for one simulated block.
func New(traceTransfers bool, blockNumber uint64, blockTimestamp uint64, blockHash common.Hash) *Tracer {
	return &Tracer{
		traceTransfers: traceTransfers,
		blockNumber:    blockNumber,
		blockTimestamp: blockTimestamp,
		blockHash:      blockHash,
	}
}

// Hooks returns the tracing hook surface. The same *tracing.Hooks is the
// one routed via vm.Config.Tracer (drives OnEnter/OnExit) and via
// statedb.SetLogger (drives OnLog from inside StateDB.AddLog).
func (t *Tracer) Hooks() *tracing.Hooks {
	return &tracing.Hooks{
		OnEnter: t.onEnter,
		OnExit:  t.onExit,
		OnLog:   t.onLog,
	}
}

// Tracer wraps Hooks() in the eth/tracers.Tracer envelope expected by
// the keeper's applyMessageWithConfig.
func (t *Tracer) Tracer() *tracers.Tracer {
	return &tracers.Tracer{Hooks: t.Hooks()}
}

// Reset prepares the tracer for a new transaction (call) within the same
// block. Clears the per-tx log buffers and updates the tx-context fields
// stamped onto subsequent captured logs. The request-scoped count is
// preserved on purpose so log indices stay monotonic across calls.
func (t *Tracer) Reset(txHash common.Hash, txIdx int) {
	t.logs = nil
	t.txHash = txHash
	t.txIdx = uint(txIdx) //nolint:gosec
}

// Logs returns the root frame's accumulated logs (real + synthetic),
// stamped with the per-block / per-tx context that was current at
// emission time. Returns nil if the root frame reverted, or before any
// frame opened.
func (t *Tracer) Logs() []*ethtypes.Log {
	if len(t.logs) == 0 {
		return nil
	}
	return t.logs[0]
}

func (t *Tracer) onEnter(depth int, typ byte, from common.Address, to common.Address, _ []byte, _ uint64, value *big.Int) {
	t.logs = append(t.logs, []*ethtypes.Log{})
	if value == nil || value.Sign() <= 0 {
		return
	}
	op := vm.OpCode(typ)
	// DELEGATECALL leaves value with the caller's frame; STATICCALL
	// cannot transfer value (the EVM rejects it before reaching the
	// tracer, but guard defensively). CALL, CALLCODE, CREATE, CREATE2,
	// SELFDESTRUCT, and the implicit top-level tx all carry real value.
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
		transferTopic,
		common.BytesToHash(from.Bytes()),
		common.BytesToHash(to.Bytes()),
	}
	t.captureLog(transferAddress, topics, common.BigToHash(value).Bytes())
}
