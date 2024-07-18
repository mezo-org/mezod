package validatorpool

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v12/precompile"
)

// OwnershipTransferredName is the name of the OwnershipTransferred event. It matches the name
// of the event in the contract ABI.
const OwnershipTransferredEventName = "OwnershipTransferred"

// ownershipTransferredEvent is the implementation of the OwnershipTransferred event that contains
// the following arguments:
// - previousOwner (indexed): is the EVM address of the now previous owner
// - newOwner (indexed): is the EVM address of the new (now current) owner
type ownershipTransferredEvent struct {
	previousOwner, newOwner common.Address
}

func newOwnershipTransferredEvent(previousOwner, newOwner common.Address) *ownershipTransferredEvent {
	return &ownershipTransferredEvent{
		previousOwner: previousOwner,
		newOwner:      newOwner,
	}
}

func (e *ownershipTransferredEvent) EventName() string {
	return OwnershipTransferredEventName
}

func (e *ownershipTransferredEvent) Arguments() []*precompile.EventArgument {
	return []*precompile.EventArgument{
		{
			Indexed: true,
			Value:   e.previousOwner,
		},
		{
			Indexed: true,
			Value:   e.newOwner,
		},
	}
}
