package validatorpool

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v12/precompile"
	poatypes "github.com/evmos/evmos/v12/x/poa/types"
)

// SubmitApplicationMethodName is the name of the submitApplication method. It matches the name
// of the method in the contract ABI.
const SubmitApplicationMethodName = "submitApplication"

// submitApplicationMethod is the implementation of the submitApplication method that registers
// a validator candidates application as pending

// The method has the following input arguments:
// - consPubKey: the consensus public key of the validator used to vote on blocks
// - operator: the EVM address identifying the validator.
type submitApplicationMethod struct {
	keeper PoaKeeper
}

func newSubmitApplicationMethod(pk PoaKeeper) *submitApplicationMethod {
	return &submitApplicationMethod{
		keeper: pk,
	}
}

func (m *submitApplicationMethod) MethodName() string {
	return SubmitApplicationMethodName
}

func (m *submitApplicationMethod) MethodType() precompile.MethodType {
	return precompile.Write
}

func (m *submitApplicationMethod) RequiredGas(_ []byte) (uint64, bool) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (m *submitApplicationMethod) Payable() bool {
	return false
}

func (m *submitApplicationMethod) Run(context *precompile.RunContext, inputs precompile.MethodInputs) (precompile.MethodOutputs, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 2); err != nil {
		return nil, err
	}

	consPubKey, ok := inputs[0].(common.Address)
	if !ok {
		return nil, fmt.Errorf("consPubKey argument must be common.Address")
	}

	operator, ok := inputs[1].(common.Address)
	if !ok {
		return nil, fmt.Errorf("operator argument must be common.Address")
	}

	validator := poatypes.Validator{
		OperatorAddress: types.ValAddress(precompile.TypesConverter.Address.ToSDK(operator)),
		ConsensusPubkey: consPubKey.String(),
	}

	err := m.keeper.SubmitApplication(
		context.SdkCtx(),
		precompile.TypesConverter.Address.ToSDK(context.MsgSender()),
		validator,
	)
	if err != nil {
		return nil, err
	}

	// emit event
	err = context.EventEmitter().Emit(
		newApplicationSubmittedEvent(
			operator,
			consPubKey,
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to emit ApplicationSubmitted event: [%w]", err)
	}

	return precompile.MethodOutputs{true}, nil
}

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
