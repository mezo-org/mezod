package validatorpool

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v12/precompile"
)

// ValidatorLeftName is the name of the ValidatorLeft event. It matches the name
// of the event in the contract ABI.
const ValidatorLeftEventName = "ValidatorLeft"

// validatorLeftEvent is the implementation of the ValidatorLeft event that contains
// the following arguments:
// - operator (indexed): is the address identifying the validators operator
type validatorLeftEvent struct {
	operator common.Address
}

func newValidatorLeftEvent(operator common.Address) *validatorLeftEvent {
	return &validatorLeftEvent{
		operator: operator,
	}
}

func (te *validatorLeftEvent) EventName() string {
	return ValidatorLeftEventName
}

func (te *validatorLeftEvent) Arguments() []*precompile.EventArgument {
	return []*precompile.EventArgument{
		{
			Indexed: true,
			Value:   te.operator,
		},
	}
}
