package validatorpool

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v12/precompile"
)

// ApplicationApprovedName is the name of the ApplicationApproved event. It matches the name
// of the event in the contract ABI.
const ApplicationApprovedEventName = "ApplicationApproved"

// applicationApprovedEvent is the implementation of the ApplicationApproved event that contains
// the following arguments:
// - operator (indexed): is the address identifying the validators operator
type applicationApprovedEvent struct {
	operator common.Address
}

func newApplicationApprovedEvent(operator common.Address) *applicationApprovedEvent {
	return &applicationApprovedEvent{
		operator: operator,
	}
}

func (e *applicationApprovedEvent) EventName() string {
	return ApplicationApprovedEventName
}

func (e *applicationApprovedEvent) Arguments() []*precompile.EventArgument {
	return []*precompile.EventArgument{
		{
			Indexed: true,
			Value:   e.operator,
		},
	}
}
