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
// Note, state and stateDiff can't be specified at the same time. If state is
// set, message execution will only use the data in the given state. Otherwise
// if stateDiff is set, all diff will be applied first and then execute the
// call message.
// MovePrecompileTo relocates the precompile at this address to the given
// destination address for the duration of the call. The source address loses
// its precompile binding; its regular account fields (Nonce, Code, Balance,
// State, StateDiff) may then be overridden in the same request. Only standard
// Ethereum precompiles (0x01-0x0A) can be moved — mezo's custom precompiles
// at 0x7b7c... are rejected.
type OverrideAccount struct {
	Nonce            *hexutil.Uint64             `json:"nonce"`
	Code             *hexutil.Bytes              `json:"code"`
	Balance          *hexutil.Big                `json:"balance"`
	State            map[common.Hash]common.Hash `json:"state"`
	StateDiff        map[common.Hash]common.Hash `json:"stateDiff"`
	MovePrecompileTo *common.Address             `json:"movePrecompileToAddress,omitempty"`
}
