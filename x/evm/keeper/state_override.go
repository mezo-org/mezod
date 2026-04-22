package keeper

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"
	"github.com/holiman/uint256"

	rpctypes "github.com/mezo-org/mezod/rpc/types"
	"github.com/mezo-org/mezod/x/evm/statedb"
	"github.com/mezo-org/mezod/x/evm/types"
)

// stateOverride is the collection of overridden accounts for eth_call.
type stateOverride map[common.Address]overrideAccount

// overrideAccount indicates the overriding fields of account during the
// execution of a message call.
type overrideAccount struct {
	Nonce            *hexutil.Uint64             `json:"nonce"`
	Code             *hexutil.Bytes              `json:"code"`
	Balance          *hexutil.Big                `json:"balance"`
	State            map[common.Hash]common.Hash `json:"state"`
	StateDiff        map[common.Hash]common.Hash `json:"stateDiff"`
	MovePrecompileTo *common.Address             `json:"movePrecompileToAddress,omitempty"`
}

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

func invalidParams(format string, args ...any) error {
	return &rpctypes.RPCError{
		Code:    rpctypes.SimErrCodeInvalidParams,
		Message: fmt.Sprintf(format, args...),
	}
}

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
// On the first invariant violation returns *rpctypes.RPCError carrying the
// spec-reserved code (-38022, -38023) or -32602 for the mezo-custom denylist
// and the two anti-nondeterminism guards that mirror go-ethereum's
// override.go dirtyAddrs / diff.has checks.
func applyStateOverrides(
	db *statedb.StateDB,
	overrides stateOverride,
	rules params.Rules,
) (map[common.Address]common.Address, error) {
	var moves map[common.Address]common.Address
	// Destinations that received a moved precompile in this request.
	moveDests := make(map[common.Address]struct{})

	for addr, override := range overrides {
		if _, dirty := moveDests[addr]; dirty {
			return nil, invalidParams(
				"account %s has already been overridden by a precompile",
				addr.Hex(),
			)
		}

		if override.MovePrecompileTo != nil {
			dest := *override.MovePrecompileTo

			if dest == addr {
				return nil, &rpctypes.RPCError{
					Code: rpctypes.SimErrCodeMovePrecompileSelfReference,
					Message: fmt.Sprintf(
						"MovePrecompileToAddress referenced itself in replacement: %s",
						addr.Hex(),
					),
				}
			}

			if _, exists := moveDests[dest]; exists {
				return nil, &rpctypes.RPCError{
					Code: rpctypes.SimErrCodeMovePrecompileDuplicateDest,
					Message: fmt.Sprintf(
						"Multiple MovePrecompileToAddress referencing the same address to replace: %s",
						dest.Hex(),
					),
				}
			}

			if _, destOverridden := overrides[dest]; destOverridden {
				return nil, invalidParams("account %s is already overridden", dest.Hex())
			}

			// Denylist check before the DefaultPrecompiles lookup so
			// mezo custom addresses surface the specific error rather
			// than "is not a precompile" (they aren't stdlib entries).
			if _, isCustom := mezoCustomPrecompileAddrs[addr]; isCustom {
				return nil, invalidParams("cannot move mezo custom precompile: %s", addr.Hex())
			}

			if _, isStdlib := vm.DefaultPrecompiles(rules)[addr]; !isStdlib {
				return nil, invalidParams("account %s is not a precompile", addr.Hex())
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
			return nil, invalidParams(
				"account %s has both state and stateDiff overrides", addr.Hex(),
			)
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
