package validatorpool

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v12/precompile"
)

// OwnershipTransferStartedName is the name of the OwnershipTransferStarted event. It matches the name
// of the event in the contract ABI.
const OwnershipTransferStartedEventName = "OwnershipTransferStarted"

// oOwnershipTransferStartedEvent is the implementation of the OwnershipTransferStarted event that contains
// the following arguments:
// - previousOwner (indexed): is the EVM address of the current (soon to be previous) owner,
// - newOwner (indexed): is the EVM address of the new owner
type ownershipTransferStartedEvent struct {
	previousOwner, newOwner common.Address
}

func newOwnershipTransferStartedEvent(previousOwner, newOwner common.Address) *ownershipTransferStartedEvent {
	return &ownershipTransferStartedEvent{
		previousOwner: previousOwner,
		newOwner:      newOwner,
	}
}

func (e *ownershipTransferStartedEvent) EventName() string {
	return OwnershipTransferStartedEventName
}

func (e *ownershipTransferStartedEvent) Arguments() []*precompile.EventArgument {
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
