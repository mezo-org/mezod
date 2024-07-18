package validatorpool

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v12/precompile"
)

// LeaveMethodName is the name of the leave method. It matches the name
// of the method in the contract ABI.
const LeaveMethodName = "leave"

// leaveMethod is the implementation of the leave method that removes
// msg.sender from the validator pool
type leaveMethod struct {
	keeper PoaKeeper
}

func newLeaveMethod(pk PoaKeeper) *leaveMethod {
	return &leaveMethod{
		keeper: pk,
	}
}

func (m *leaveMethod) MethodName() string {
	return LeaveMethodName
}

func (m *leaveMethod) MethodType() precompile.MethodType {
	return precompile.Write
}

func (m *leaveMethod) RequiredGas(_ []byte) (uint64, bool) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (m *leaveMethod) Payable() bool {
	return false
}

func (m *leaveMethod) Run(context *precompile.RunContext, inputs precompile.MethodInputs) (precompile.MethodOutputs, error) {
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
		newValidatorLeftEvent(context.MsgSender()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to emit validatorLeft event: [%w]", err)
	}

	return precompile.MethodOutputs{true}, nil
}

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
