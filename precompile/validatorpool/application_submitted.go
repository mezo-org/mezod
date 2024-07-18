package validatorpool

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v12/precompile"
)

// ApplicationSubmittedEventName is the name of the ApplicationSubmitted event.
// It matches the name of the event in the contract ABI.
const ApplicationSubmittedEventName = "ApplicationSubmitted"

// applicationSubmitted is the implementation of the ApplicationSubmitted
// event that contains the following arguments:
// - operator (indexed): is the address identifying the validator,
// - consPubKey (indexed): is the consensus public key of the validator used tovote on blocks.
type applicationSubmittedEvent struct {
	operator, consPubKey common.Address
}

func newApplicationSubmittedEvent(operator, consPubKey common.Address) *applicationSubmittedEvent {
	return &applicationSubmittedEvent{
		operator:   operator,
		consPubKey: consPubKey,
	}
}

func (e *applicationSubmittedEvent) EventName() string {
	return ApplicationSubmittedEventName
}

func (e *applicationSubmittedEvent) Arguments() []*precompile.EventArgument {
	return []*precompile.EventArgument{
		{
			Indexed: true,
			Value:   e.operator,
		},
		{
			Indexed: true,
			Value:   e.consPubKey,
		},
	}
}
