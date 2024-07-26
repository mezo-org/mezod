package validatorpool

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v12/precompile"
)

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

func NewKickMethod(pk PoaKeeper) *KickMethod {
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

	operator, ok := inputs[1].(common.Address)
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
		return nil, fmt.Errorf("failed to emitv ValidatorKicked event: [%w]", err)
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
