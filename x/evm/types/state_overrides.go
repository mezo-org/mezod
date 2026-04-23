package types

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// StateOverride is the collection of overridden accounts for eth_call and
// the per-block state overrides carried by eth_simulateV1.
type StateOverride map[common.Address]OverrideAccount

// OverrideAccount indicates the overriding fields of account during the
// execution of a message call.
type OverrideAccount struct {
	Nonce            *hexutil.Uint64             `json:"nonce"`
	Code             *hexutil.Bytes              `json:"code"`
	Balance          *hexutil.Big                `json:"balance"`
	State            map[common.Hash]common.Hash `json:"state"`
	StateDiff        map[common.Hash]common.Hash `json:"stateDiff"`
	MovePrecompileTo *common.Address             `json:"movePrecompileToAddress,omitempty"`
}
