package keeper

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/holiman/uint256"
	"github.com/mezo-org/mezod/x/evm/statedb"
)

// stateOverride is the collection of overridden accounts for eth_call.
type stateOverride map[common.Address]overrideAccount

// overrideAccount indicates the overriding fields of account during the
// execution of a message call.
type overrideAccount struct {
	Nonce     *hexutil.Uint64              `json:"nonce"`
	Code      *hexutil.Bytes               `json:"code"`
	Balance   **hexutil.Big                `json:"balance"`
	State     *map[common.Hash]common.Hash `json:"state"`
	StateDiff *map[common.Hash]common.Hash `json:"stateDiff"`
}

// applyStateOverrides modifies the given StateDB according to the provided
// state overrides. This is used by eth_call to simulate calls with modified
// account state.
func applyStateOverrides(db *statedb.StateDB, overrides stateOverride) error {
	for addr, override := range overrides {
		if override.Nonce != nil {
			db.SetNonce(addr, uint64(*override.Nonce))
		}

		if override.Code != nil {
			db.SetCode(addr, *override.Code)
		}

		if override.Balance != nil {
			balance, _ := uint256.FromBig((*big.Int)(*override.Balance))
			db.SetBalance(addr, balance, tracing.BalanceChangeUnspecified)
		}

		if override.State != nil && override.StateDiff != nil {
			return fmt.Errorf(
				"account %s has both state and stateDiff overrides", addr.Hex(),
			)
		}

		if override.State != nil {
			db.SetStorage(addr, *override.State)
		}

		if override.StateDiff != nil {
			for key, val := range *override.StateDiff {
				db.SetState(addr, key, val)
			}
		}
	}
	return nil
}
