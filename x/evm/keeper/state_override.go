package keeper

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"
	"github.com/holiman/uint256"

	"github.com/mezo-org/mezod/x/evm/statedb"
	"github.com/mezo-org/mezod/x/evm/types"
)

// mezoCustomPrecompileAddrs is the set of mezo custom precompile source
// addresses that callers are NOT allowed to relocate via MovePrecompileTo.
// Derived from types.DefaultPrecompilesVersions so the denylist cannot drift
// out of sync with the chain's registered custom precompiles.
var mezoCustomPrecompileAddrs = func() map[common.Address]struct{} {
	set := make(map[common.Address]struct{}, len(types.DefaultPrecompilesVersions))
	for _, pv := range types.DefaultPrecompilesVersions {
		set[common.HexToAddress(pv.PrecompileAddress)] = struct{}{}
	}
	return set
}()

// applyStateOverrides validates the override set and mutates the StateDB for
// nonce / code / balance / state / stateDiff entries. The returned map holds
// validated MovePrecompileTo relocations (source → destination) that the
// caller must install on the EVM's precompile registry before the message
// executes; the map is nil when no MovePrecompileTo entries are present.
//
// Source-address eligibility is checked against vm.DefaultPrecompiles(rules)
// (stdlib precompiles only) with mezoCustomPrecompileAddrs rejecting the
// chain's own custom precompiles.
//
// On the first invariant violation returns a *types.SimError carrying the
// spec-reserved JSON-RPC code; the returned error rides the plain `error`
// channel so callers that don't care about the code (eth_call,
// eth_estimateGas) propagate it unchanged, while callers that do care
// (SimulateV1) branch with errors.As.
func applyStateOverrides(
	db *statedb.StateDB,
	overrides types.StateOverride,
	rules params.Rules,
) (map[common.Address]common.Address, error) {
	var moves map[common.Address]common.Address
	// Destinations that received a moved precompile in this request.
	moveDests := make(map[common.Address]struct{})

	for addr, override := range overrides {
		if _, dirty := moveDests[addr]; dirty {
			return nil, types.NewSimAccountTainted(addr)
		}

		if override.MovePrecompileTo != nil {
			dest := *override.MovePrecompileTo

			if dest == addr {
				return nil, types.NewSimMovePrecompileSelfRef(addr)
			}

			if _, exists := moveDests[dest]; exists {
				return nil, types.NewSimMovePrecompileDupDest(dest)
			}

			if _, destOverridden := overrides[dest]; destOverridden {
				return nil, types.NewSimDestAlreadyOverridden(dest)
			}

			// Denylist check before the DefaultPrecompiles lookup so
			// mezo custom addresses surface the specific error rather
			// than "is not a precompile" (they aren't stdlib entries).
			if _, isCustom := mezoCustomPrecompileAddrs[addr]; isCustom {
				return nil, types.NewSimMoveMezoCustom(addr)
			}

			if _, isStdlib := vm.DefaultPrecompiles(rules)[addr]; !isStdlib {
				return nil, types.NewSimNotAPrecompile(addr)
			}

			if moves == nil {
				moves = make(map[common.Address]common.Address)
			}
			moves[addr] = dest
			moveDests[dest] = struct{}{}
		}

		if override.Nonce != nil {
			db.SetNonce(addr, uint64(*override.Nonce))
		}

		if override.Code != nil {
			db.SetCode(addr, *override.Code)
		}

		if override.Balance != nil {
			balance, _ := uint256.FromBig((*big.Int)(override.Balance))
			db.OverrideBalance(addr, balance, tracing.BalanceChangeUnspecified)
		}

		if override.State != nil && override.StateDiff != nil {
			return nil, types.NewSimStateAndStateDiff(addr)
		}

		if override.State != nil {
			db.OverrideStorage(addr, override.State)
		}

		if override.StateDiff != nil {
			for key, val := range override.StateDiff {
				db.SetState(addr, key, val)
			}
		}
	}
	return moves, nil
}
