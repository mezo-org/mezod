package validatorpool

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v12/precompile"
)

// LeaveMethodName is the name of the leave method. It matches the name
// of the method in the contract ABI.
const LeaveMethodName = "leave"

// LeaveMethod is the implementation of the leave method that removes
// msg.sender from the validator pool
type LeaveMethod struct {
	keeper PoaKeeper
}

func newLeaveMethod(pk PoaKeeper) *LeaveMethod {
	return &LeaveMethod{
		keeper: pk,
	}
}

func (m *LeaveMethod) MethodName() string {
	return LeaveMethodName
}

func (m *LeaveMethod) MethodType() precompile.MethodType {
	return precompile.Write
}

func (m *LeaveMethod) RequiredGas(_ []byte) (uint64, bool) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (m *LeaveMethod) Payable() bool {
	return false
}

func (m *LeaveMethod) Run(context *precompile.RunContext, inputs precompile.MethodInputs) (precompile.MethodOutputs, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 0); err != nil {
		return nil, err
	}

	err := m.keeper.Leave(
		context.SdkCtx(),
		precompile.TypesConverter.Address.ToSDK(context.MsgSender()),
	)
	if err != nil {
		return nil, err
	}

	// emit event
	err = context.EventEmitter().Emit(
		NewValidatorLeftEvent(context.MsgSender()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to emit validatorLeft event: [%w]", err)
	}

	return precompile.MethodOutputs{true}, nil
}

// ValidatorLeftName is the name of the ValidatorLeft event. It matches the name
// of the event in the contract ABI.
const ValidatorLeftEventName = "ValidatorLeft"

// ValidatorLeftEvent is the implementation of the ValidatorLeft event that contains
// the following arguments:
// - operator (indexed): is the address identifying the validators operator
type ValidatorLeftEvent struct {
	operator common.Address
}

func NewValidatorLeftEvent(operator common.Address) *ValidatorLeftEvent {
	return &ValidatorLeftEvent{
		operator: operator,
	}
}

func (e *ValidatorLeftEvent) EventName() string {
	return ValidatorLeftEventName
}

func (e *ValidatorLeftEvent) Arguments() []*precompile.EventArgument {
	return []*precompile.EventArgument{
		{
			Indexed: true,
			Value:   e.operator,
		},
	}
}
