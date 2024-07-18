package validatorpool

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v12/precompile"
)

// ApproveApplicationMethodName is the name of the approveApplication method. It matches the name
// of the method in the contract ABI.
const ApproveApplicationMethodName = "approveApplication"

// approveApplicationMethod is the implementation of the approveApplication method that approves
// a pending validator application.

// The method has the following input arguments:
// - operator: the EVM address identifying the validator.
type approveApplicationMethod struct {
	keeper PoaKeeper
}

func newApproveApplicationMethod(pk PoaKeeper) *approveApplicationMethod {
	return &approveApplicationMethod{
		keeper: pk,
	}
}

func (m *approveApplicationMethod) MethodName() string {
	return ApproveApplicationMethodName
}

func (m *approveApplicationMethod) MethodType() precompile.MethodType {
	return precompile.Write
}

func (m *approveApplicationMethod) RequiredGas(_ []byte) (uint64, bool) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (m *approveApplicationMethod) Payable() bool {
	return false
}

func (m *approveApplicationMethod) Run(context *precompile.RunContext, inputs precompile.MethodInputs) (precompile.MethodOutputs, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 1); err != nil {
		return nil, err
	}

	operator, ok := inputs[0].(common.Address)
	if !ok {
		return nil, fmt.Errorf("operator argument must be of type common.Address")
	}

	err := m.keeper.ApproveApplication(
		context.SdkCtx(),
		precompile.TypesConverter.Address.ToSDK(context.MsgSender()),
		types.ValAddress(precompile.TypesConverter.Address.ToSDK(operator)),
	)
	if err != nil {
		return nil, err
	}

	// emit events
	err = context.EventEmitter().Emit(
		newApplicationApprovedEvent(
			operator,
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to emit ApplicationApproved event: [%w]", err)
	}
	err = context.EventEmitter().Emit(
		newValidatorJoinedEvent(
			operator,
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to emit ValidatorJoined event: [%w]", err)
	}

	return precompile.MethodOutputs{true}, nil
}

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
