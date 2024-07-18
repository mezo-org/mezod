package validatorpool

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v12/precompile"
)

// ValidatorJoinedName is the name of the ValidatorJoined event. It matches the name
// of the event in the contract ABI.
const ValidatorJoinedEventName = "ValidatorJoined"

// validatorJoinedEvent is the implementation of the ValidatorJoined event that contains
// the following arguments:
// - operator (indexed): is the EVM address identifying the validators operator,
type validatorJoinedEvent struct {
	operator common.Address
}

func newValidatorJoinedEvent(operator common.Address) *validatorJoinedEvent {
	return &validatorJoinedEvent{
		operator: operator,
	}
}

func (e *validatorJoinedEvent) EventName() string {
	return ValidatorJoinedEventName
}

func (e *validatorJoinedEvent) Arguments() []*precompile.EventArgument {
	return []*precompile.EventArgument{
		{
			Indexed: true,
			Value:   e.operator,
		},
	}
}
