package keeper

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/tracing"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
)

// simTransferTopic is keccak256("Transfer(address,address,uint256)").
var simTransferTopic = common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")

// simTransferAddress is the ERC-7528 pseudo-address used as the emitter of
// synthetic native-value Transfer logs.
var simTransferAddress = common.HexToAddress("0xEeeeeEeeeEeEeeEeEeEeeEEEeeeeEeeeeeeeEEeE")

// simTracer captures EVM logs across nested call frames; on traceTransfers,
// it also synthesizes ERC-7528 Transfer logs for value-bearing CALL /
// CALLCODE / SELFDESTRUCT edges. DELEGATECALL is excluded because the value
// belongs to the caller's frame.
type simTracer struct {
	logs [][]*ethtypes.Log
	// preserved across reset for monotonic Log.Index.
	count uint

	traceTransfers bool
	blockNumber    uint64
	blockHash      common.Hash
	txHash         common.Hash
	txIdx          uint
}

func newSimTracer(traceTransfers bool, blockNumber uint64, blockHash common.Hash) *simTracer {
	return &simTracer{
		traceTransfers: traceTransfers,
		blockNumber:    blockNumber,
		blockHash:      blockHash,
	}
}

func (t *simTracer) Hooks() *tracing.Hooks {
	return &tracing.Hooks{
		OnEnter: t.onEnter,
		OnExit:  t.onExit,
		OnLog:   t.onLog,
	}
}

// reset clears the per-tx log buffers and updates the tx-context fields
// stamped onto subsequent captured logs. count is preserved.
func (t *simTracer) reset(txHash common.Hash, txIdx int) {
	t.logs = nil
	t.txHash = txHash
	t.txIdx = uint(txIdx) //nolint:gosec
}

// Logs returns the root frame's accumulated logs. Returns nil if the root
// frame reverted, or before any frame opened.
func (t *simTracer) Logs() []*ethtypes.Log {
	if len(t.logs) == 0 {
		return nil
	}
	return t.logs[0]
}

func (t *simTracer) onEnter(_ int, typ byte, from common.Address, to common.Address, _ []byte, _ uint64, value *big.Int) {
	t.logs = append(t.logs, []*ethtypes.Log{})
	if value == nil || value.Sign() <= 0 {
		return
	}
	op := vm.OpCode(typ)
	if op == vm.DELEGATECALL || op == vm.STATICCALL {
		return
	}
	// Skip mezo custom precompiles: each emits its own ERC-20 Transfer
	// event, so a synthetic log here would double-count.
	if _, deny := mezoCustomPrecompileAddrs[to]; deny {
		return
	}
	t.captureTransfer(from, to, value)
}

func (t *simTracer) onExit(depth int, _ []byte, _ uint64, _ error, reverted bool) {
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

func (t *simTracer) onLog(log *ethtypes.Log) {
	t.captureLog(log.Address, log.Topics, log.Data)
}

func (t *simTracer) captureLog(address common.Address, topics []common.Hash, data []byte) {
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

func (t *simTracer) captureTransfer(from, to common.Address, value *big.Int) {
	if !t.traceTransfers {
		return
	}
	topics := []common.Hash{
		simTransferTopic,
		common.BytesToHash(from.Bytes()),
		common.BytesToHash(to.Bytes()),
	}
	t.captureLog(simTransferAddress, topics, common.BigToHash(value).Bytes())
}
