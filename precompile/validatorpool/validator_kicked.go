package validatorpool

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v12/precompile"
)

// ValidatorKickedName is the name of the ValidatorKicked event. It matches the name
// of the event in the contract ABI.
const ValidatorKickedEventName = "ValidatorKicked"

// validatorKickedEvent is the implementation of the ValidatorKicked event that contains
// the following arguments:
// - operator (indexed): is the address identifying the validators operator
type validatorKickedEvent struct {
	operator common.Address
}

func newValidatorKickedEvent(operator common.Address) *validatorKickedEvent {
	return &validatorKickedEvent{
		operator: operator,
	}
}

func (e *validatorKickedEvent) EventName() string {
	return ValidatorKickedEventName
}

func (e *validatorKickedEvent) Arguments() []*precompile.EventArgument {
	return []*precompile.EventArgument{
		{
			Indexed: true,
			Value:   e.operator,
		},
	}
}
