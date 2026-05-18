package statedb

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/tracing"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/holiman/uint256"
)

// AddTracingHooks installs tracing hooks and returns a cleanup function.
func (s *StateDB) AddTracingHooks(hooks *tracing.Hooks) func() {
	if hooks == nil {
		return func() {}
	}
	for _, existing := range s.tracingHooks {
		if existing == hooks {
			return func() {}
		}
	}
	s.tracingHooks = append(s.tracingHooks, hooks)
	return func() {
		for i, existing := range s.tracingHooks {
			if existing == hooks {
				s.tracingHooks = append(s.tracingHooks[:i], s.tracingHooks[i+1:]...)
				return
			}
		}
	}
}

func (s *StateDB) traceLog(log *ethtypes.Log) {
	for _, hooks := range s.tracingHooks {
		if hooks.OnLog != nil {
			hooks.OnLog(log)
		}
	}
}

func (s *StateDB) traceBalanceChange(
	addr common.Address,
	prev *uint256.Int,
	next *uint256.Int,
	reason tracing.BalanceChangeReason,
) {
	if prev.Cmp(next) == 0 {
		return
	}
	prevBig := prev.ToBig()
	nextBig := next.ToBig()
	for _, hooks := range s.tracingHooks {
		if hooks.OnBalanceChange != nil {
			hooks.OnBalanceChange(addr, prevBig, nextBig, reason)
		}
	}
}

func (s *StateDB) traceNonceChange(
	addr common.Address,
	prev uint64,
	next uint64,
	reason tracing.NonceChangeReason,
) {
	for _, hooks := range s.tracingHooks {
		if hooks.OnNonceChangeV2 != nil {
			hooks.OnNonceChangeV2(addr, prev, next, reason)
		} else if hooks.OnNonceChange != nil {
			hooks.OnNonceChange(addr, prev, next)
		}
	}
}

func (s *StateDB) traceCodeChange(
	addr common.Address,
	prev []byte,
	code []byte,
	codeHash common.Hash,
	reason tracing.CodeChangeReason,
) {
	if len(s.tracingHooks) == 0 {
		return
	}
	prevHash := crypto.Keccak256Hash(prev)
	if prevHash == codeHash {
		return
	}
	for _, hooks := range s.tracingHooks {
		if hooks.OnCodeChangeV2 != nil {
			hooks.OnCodeChangeV2(addr, prevHash, prev, codeHash, code, reason)
		} else if hooks.OnCodeChange != nil {
			hooks.OnCodeChange(addr, prevHash, prev, codeHash, code)
		}
	}
}

func (s *StateDB) traceSelfDestructCodeChange(addr common.Address, prev []byte) {
	if len(s.tracingHooks) == 0 {
		return
	}
	if len(prev) == 0 {
		return
	}
	prevHash := crypto.Keccak256Hash(prev)
	for _, hooks := range s.tracingHooks {
		if hooks.OnCodeChangeV2 != nil {
			hooks.OnCodeChangeV2(addr, prevHash, prev, ethtypes.EmptyCodeHash, nil, tracing.CodeChangeSelfDestruct)
		} else if hooks.OnCodeChange != nil {
			hooks.OnCodeChange(addr, prevHash, prev, ethtypes.EmptyCodeHash, nil)
		}
	}
}

func (s *StateDB) traceStorageChange(addr common.Address, key common.Hash, prev common.Hash, next common.Hash) {
	if prev == next {
		return
	}
	for _, hooks := range s.tracingHooks {
		if hooks.OnStorageChange != nil {
			hooks.OnStorageChange(addr, key, prev, next)
		}
	}
}
