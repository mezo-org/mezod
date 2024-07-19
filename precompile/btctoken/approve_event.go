package btctoken

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v12/precompile"
)

// ApprovalEventName is the name of the Approval event. It matches the name
// of the event in the contract ABI.
const ApprovalEventName = "Approval"

// ApprovalEvent is the implementation of the Approval event that contains
// the following arguments:
// - owner (indexed): the address of BTC owner,
// - to (indexed): the address of spender,
// - amount (non-indexed): the amount of tokens approved.
type ApprovalEvent struct {
	from, to common.Address
	amount   *big.Int
}

func NewApprovalEvent(from, to common.Address, amount *big.Int) *ApprovalEvent {
	return &ApprovalEvent{
		from:   from,
		to:     to,
		amount: amount,
	}
}

func (ae *ApprovalEvent) EventName() string {
	return ApprovalEventName
}

func (ae *ApprovalEvent) Arguments() []*precompile.EventArgument {
	return []*precompile.EventArgument{
		{
			Indexed: true,
			Value:   ae.from,
		},
		{
			Indexed: true,
			Value:   ae.to,
		},
		{
			Indexed: false,
			Value:   ae.amount,
		},
	}
}
