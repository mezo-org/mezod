package validatorpool

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/mezo-org/mezod/precompile"
)

// ValidatorsMethodName is the name of the validators method. It matches the name
// of the method in the contract ABI.
const ValidatorsMethodName = "validators"

// ValidatorsMethod is the implementation of the validators method that returns
// the operator addresses of all existing validators
type ValidatorsMethod struct {
	keeper PoaKeeper
}

func newValidatorsMethod(pk PoaKeeper) *ValidatorsMethod {
	return &ValidatorsMethod{
		keeper: pk,
	}
}

func (m *ValidatorsMethod) MethodName() string {
	return ValidatorsMethodName
}

func (m *ValidatorsMethod) MethodType() precompile.MethodType {
	return precompile.Read
}

func (m *ValidatorsMethod) RequiredGas(_ []byte) (uint64, bool) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (m *ValidatorsMethod) Payable() bool {
	return false
}

func (m *ValidatorsMethod) Run(context *precompile.RunContext, inputs precompile.MethodInputs) (precompile.MethodOutputs, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 0); err != nil {
		return nil, err
	}

	validators := m.keeper.GetAllValidators(
		context.SdkCtx(),
	)

	operators := make([]common.Address, len(validators))

	for i, validator := range validators {
		operator := validator.GetOperator()
		operators[i] = precompile.TypesConverter.Address.FromSDK(types.AccAddress(operator))
	}

	return precompile.MethodOutputs{operators}, nil
}

// ValidatorMethodName is the name of the validators method. It matches the name
// of the method in the contract ABI.
const ValidatorMethodName = "validator"

// ValidatorMethod is the implementation of the validator method that returns
// the validator information for the given operator address
type ValidatorMethod struct {
	keeper PoaKeeper
}

func newValidatorMethod(pk PoaKeeper) *ValidatorMethod {
	return &ValidatorMethod{
		keeper: pk,
	}
}

func (m *ValidatorMethod) MethodName() string {
	return ValidatorMethodName
}

func (m *ValidatorMethod) MethodType() precompile.MethodType {
	return precompile.Read
}

func (m *ValidatorMethod) RequiredGas(_ []byte) (uint64, bool) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (m *ValidatorMethod) Payable() bool {
	return false
}

func (m *ValidatorMethod) Run(context *precompile.RunContext, inputs precompile.MethodInputs) (precompile.MethodOutputs, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 1); err != nil {
		return nil, err
	}

	operator, ok := inputs[0].(common.Address)
	if !ok {
		return nil, fmt.Errorf("operator argument must be of type common.Address")
	}

	validator, found := m.keeper.GetValidator(
		context.SdkCtx(),
		types.ValAddress(precompile.TypesConverter.Address.ToSDK(operator)),
	)
	if !found {
		return nil, fmt.Errorf("validator does not exist")
	}

	var consPubKey [32]byte
	copy(consPubKey[:], validator.GetConsPubKey().Bytes())

	return precompile.MethodOutputs{consPubKey, validator.Description}, nil
}

// KickMethodName is the name of the kick method. It matches the name
// of the method in the contract ABI.
const KickMethodName = "kick"

// KickMethod is the implementation of the kick method that registers
// a validator candidates application as pending

// The method has the following input arguments:
// - operator: the address identifying the validator.
type KickMethod struct {
	keeper PoaKeeper
}

func newKickMethod(pk PoaKeeper) *KickMethod {
	return &KickMethod{
		keeper: pk,
	}
}

func (m *KickMethod) MethodName() string {
	return KickMethodName
}

func (m *KickMethod) MethodType() precompile.MethodType {
	return precompile.Write
}

func (m *KickMethod) RequiredGas(_ []byte) (uint64, bool) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (m *KickMethod) Payable() bool {
	return false
}

func (m *KickMethod) Run(context *precompile.RunContext, inputs precompile.MethodInputs) (precompile.MethodOutputs, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 1); err != nil {
		return nil, err
	}

	operator, ok := inputs[0].(common.Address)
	if !ok {
		return nil, fmt.Errorf("operator argument must be common.Address")
	}

	err := m.keeper.Kick(
		context.SdkCtx(),
		precompile.TypesConverter.Address.ToSDK(context.MsgSender()),
		types.ValAddress(precompile.TypesConverter.Address.ToSDK(operator)),
	)
	if err != nil {
		return nil, err
	}

	// emit event
	err = context.EventEmitter().Emit(
		NewValidatorKickedEvent(operator),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to emit ValidatorKicked event: [%w]", err)
	}

	return precompile.MethodOutputs{true}, nil
}

// ValidatorKickedName is the name of the ValidatorKicked event. It matches the name
// of the event in the contract ABI.
const ValidatorKickedEventName = "ValidatorKicked"

// ValidatorKickedEvent is the implementation of the ValidatorKicked event that contains
// the following arguments:
// - operator (indexed): is the address identifying the validators operator
type ValidatorKickedEvent struct {
	operator common.Address
}

func NewValidatorKickedEvent(operator common.Address) *ValidatorKickedEvent {
	return &ValidatorKickedEvent{
		operator: operator,
	}
}

func (e *ValidatorKickedEvent) EventName() string {
	return ValidatorKickedEventName
}

func (e *ValidatorKickedEvent) Arguments() []*precompile.EventArgument {
	return []*precompile.EventArgument{
		{
			Indexed: true,
			Value:   e.operator,
		},
	}
}

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
