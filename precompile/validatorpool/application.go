package validatorpool

import (
	"encoding/json"
	"fmt"

	cryptocdc "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v12/precompile"
	poatypes "github.com/evmos/evmos/v12/x/poa/types"
	"github.com/tendermint/tendermint/crypto/ed25519"
)

// SubmitApplicationMethodName is the name of the submitApplication method. It matches the name
// of the method in the contract ABI.
const SubmitApplicationMethodName = "submitApplication"

// submitApplicationMethod is the implementation of the submitApplication method that registers
// a validator candidates application as pending

// The method has the following input arguments:
// - consPubKey: the consensus public key of the validator used to vote on blocks
// - operator: the EVM address identifying the validator
// - description: the validators description info
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
	// check method inputs
	if err := precompile.ValidateMethodInputsCount(inputs, 3); err != nil {
		return nil, err
	}

	consPubKeyBytes, ok := inputs[0].([32]byte)
	if !ok {
		return nil, fmt.Errorf("consPubKey argument must be bytes32")
	}

	operator, ok := inputs[1].(common.Address)
	if !ok {
		return nil, fmt.Errorf("operator argument must be common.Address")
	}

	descBytes, ok := inputs[2].([]byte)
	if !ok {
		return nil, fmt.Errorf("description argument must be bytes")
	}

	tmpk := ed25519.PubKey(consPubKeyBytes[:])
	consPubKey, err := cryptocdc.FromTmPubKeyInterface(tmpk)
	if err != nil {
		return nil, err
	}

	var description poatypes.Description
	err = json.Unmarshal(descBytes, &description)
	if err != nil {
		return nil, err
	}

	validator := poatypes.NewValidator(
		types.ValAddress(precompile.TypesConverter.Address.ToSDK(operator)),
		consPubKey,
		description,
	)

	err = m.keeper.SubmitApplication(
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
			consPubKeyBytes,
			descBytes,
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
// - consPubKey (indexed): is the consensus public key of the validator used to vote on blocks.
// - description: is the validators description info
type applicationSubmittedEvent struct {
	consPubKey  [32]byte
	operator    common.Address
	description []byte
}

func newApplicationSubmittedEvent(operator common.Address, consPubKey [32]byte, description []byte) *applicationSubmittedEvent {
	return &applicationSubmittedEvent{
		operator:    operator,
		consPubKey:  consPubKey,
		description: description,
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
		{
			Indexed: false,
			Value:   e.description,
		},
	}
}

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
