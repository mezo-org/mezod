package validatorpool

import (
	"fmt"

	"github.com/cometbft/cometbft/crypto/ed25519"
	cryptocdc "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/mezo-org/mezod/precompile"
	poatypes "github.com/mezo-org/mezod/x/poa/types"
)

// SubmitApplicationMethodName is the name of the submitApplication method. It matches the name
// of the method in the contract ABI.
const SubmitApplicationMethodName = "submitApplication"

// SubmitApplicationMethod is the implementation of the submitApplication method that registers
// a validator candidates application as pending

// The method has the following input arguments:
// - consPubKey: the consensus public key of the validator used to vote on blocks
// - description: the validators description info
type SubmitApplicationMethod struct {
	keeper PoaKeeper
}

func newSubmitApplicationMethod(pk PoaKeeper) *SubmitApplicationMethod {
	return &SubmitApplicationMethod{
		keeper: pk,
	}
}

func (m *SubmitApplicationMethod) MethodName() string {
	return SubmitApplicationMethodName
}

func (m *SubmitApplicationMethod) MethodType() precompile.MethodType {
	return precompile.Write
}

func (m *SubmitApplicationMethod) RequiredGas(_ []byte) (uint64, bool) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (m *SubmitApplicationMethod) Payable() bool {
	return false
}

func (m *SubmitApplicationMethod) Run(context *precompile.RunContext, inputs precompile.MethodInputs) (precompile.MethodOutputs, error) {
	// check method inputs
	if err := precompile.ValidateMethodInputsCount(inputs, 2); err != nil {
		return nil, err
	}

	consPubKeyBytes, ok := inputs[0].([32]byte)
	if !ok {
		return nil, fmt.Errorf("consPubKey argument must be type bytes32")
	}

	description, ok := inputs[1].(Description)
	if !ok {
		return nil, fmt.Errorf("description argument must be type Description")
	}

	operator := context.MsgSender()

	// Here we assume consPubKeyBytes is a valid ED25519 key, without performing any additional validation.
	// We may need to add validation here.
	tmpk := ed25519.PubKey(consPubKeyBytes[:])
	consPubKey, err := cryptocdc.FromCmtPubKeyInterface(tmpk)
	if err != nil {
		return nil, err
	}

	validator, err := poatypes.NewValidator(
		types.ValAddress(precompile.TypesConverter.Address.ToSDK(operator)),
		consPubKey,
		poatypes.Description(description),
	)
	if err != nil {
		return nil, err
	}

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
		NewApplicationSubmittedEvent(
			operator,
			consPubKeyBytes,
			description,
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
type ApplicationSubmittedEvent struct {
	operator    common.Address
	consPubKey  [32]byte
	description Description
}

func NewApplicationSubmittedEvent(operator common.Address, consPubKey [32]byte, description Description) *ApplicationSubmittedEvent {
	return &ApplicationSubmittedEvent{
		operator:    operator,
		consPubKey:  consPubKey,
		description: description,
	}
}

func (e *ApplicationSubmittedEvent) EventName() string {
	return ApplicationSubmittedEventName
}

func (e *ApplicationSubmittedEvent) Arguments() []*precompile.EventArgument {
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

// ApproveApplicationMethod is the implementation of the approveApplication method that approves
// a pending validator application.

// The method has the following input arguments:
// - operator: the EVM address identifying the validator.
type ApproveApplicationMethod struct {
	keeper PoaKeeper
}

func newApproveApplicationMethod(pk PoaKeeper) *ApproveApplicationMethod {
	return &ApproveApplicationMethod{
		keeper: pk,
	}
}

func (m *ApproveApplicationMethod) MethodName() string {
	return ApproveApplicationMethodName
}

func (m *ApproveApplicationMethod) MethodType() precompile.MethodType {
	return precompile.Write
}

func (m *ApproveApplicationMethod) RequiredGas(_ []byte) (uint64, bool) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (m *ApproveApplicationMethod) Payable() bool {
	return false
}

func (m *ApproveApplicationMethod) Run(context *precompile.RunContext, inputs precompile.MethodInputs) (precompile.MethodOutputs, error) {
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
		NewApplicationApprovedEvent(
			operator,
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to emit ApplicationApproved event: [%w]", err)
	}
	err = context.EventEmitter().Emit(
		NewValidatorJoinedEvent(
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

// ApplicationApprovedEvent is the implementation of the ApplicationApproved event that contains
// the following arguments:
// - operator (indexed): is the address identifying the validators operator
type ApplicationApprovedEvent struct {
	operator common.Address
}

func NewApplicationApprovedEvent(operator common.Address) *ApplicationApprovedEvent {
	return &ApplicationApprovedEvent{
		operator: operator,
	}
}

func (e *ApplicationApprovedEvent) EventName() string {
	return ApplicationApprovedEventName
}

func (e *ApplicationApprovedEvent) Arguments() []*precompile.EventArgument {
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
type ValidatorJoinedEvent struct {
	operator common.Address
}

func NewValidatorJoinedEvent(operator common.Address) *ValidatorJoinedEvent {
	return &ValidatorJoinedEvent{
		operator: operator,
	}
}

func (e *ValidatorJoinedEvent) EventName() string {
	return ValidatorJoinedEventName
}

func (e *ValidatorJoinedEvent) Arguments() []*precompile.EventArgument {
	return []*precompile.EventArgument{
		{
			Indexed: true,
			Value:   e.operator,
		},
	}
}

// ApplicationsMethodName is the name of the Applications method. It matches the name
// of the method in the contract ABI.
const ApplicationsMethodName = "applications"

// ApplicationsMethod is the implementation of the applications method that returns
// the the operators addresses of all existing applications
type ApplicationsMethod struct {
	keeper PoaKeeper
}

func newApplicationsMethod(pk PoaKeeper) *ApplicationsMethod {
	return &ApplicationsMethod{
		keeper: pk,
	}
}

func (m *ApplicationsMethod) MethodName() string {
	return ApplicationsMethodName
}

func (m *ApplicationsMethod) MethodType() precompile.MethodType {
	return precompile.Read
}

func (m *ApplicationsMethod) RequiredGas(_ []byte) (uint64, bool) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (m *ApplicationsMethod) Payable() bool {
	return false
}

func (m *ApplicationsMethod) Run(context *precompile.RunContext, inputs precompile.MethodInputs) (precompile.MethodOutputs, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 0); err != nil {
		return nil, err
	}

	applications := m.keeper.GetAllApplications(
		context.SdkCtx(),
	)

	operators := make([]common.Address, len(applications))

	for i, application := range applications {
		validator := application.GetValidator()
		valAddress := validator.GetOperator()
		operators[i] = precompile.TypesConverter.Address.FromSDK(types.AccAddress(valAddress))
	}

	return precompile.MethodOutputs{operators}, nil
}

// ApplicationMethodName is the name of the applications method. It matches the name
// of the method in the contract ABI.
const ApplicationMethodName = "application"

// ApplicationMethod is the implementation of the application method that returns
// an application for a given operator address
type ApplicationMethod struct {
	keeper PoaKeeper
}

func newApplicationMethod(pk PoaKeeper) *ApplicationMethod {
	return &ApplicationMethod{
		keeper: pk,
	}
}

func (m *ApplicationMethod) MethodName() string {
	return ApplicationMethodName
}

func (m *ApplicationMethod) MethodType() precompile.MethodType {
	return precompile.Read
}

func (m *ApplicationMethod) RequiredGas(_ []byte) (uint64, bool) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (m *ApplicationMethod) Payable() bool {
	return false
}

func (m *ApplicationMethod) Run(context *precompile.RunContext, inputs precompile.MethodInputs) (precompile.MethodOutputs, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 1); err != nil {
		return nil, err
	}

	operator, ok := inputs[0].(common.Address)
	if !ok {
		return nil, fmt.Errorf("operator argument must be of type common.Address")
	}

	application, found := m.keeper.GetApplication(
		context.SdkCtx(),
		types.ValAddress(precompile.TypesConverter.Address.ToSDK(operator)),
	)
	if !found {
		return nil, fmt.Errorf("application does not exist")
	}

	val := application.GetValidator()
	var consPubKey [32]byte
	copy(consPubKey[:], val.GetConsPubKey().Bytes())

	return precompile.MethodOutputs{consPubKey, val.Description}, nil
}

// CleanupApplicationsMethodName is the name of the cleanupApplications method. It matches the name
// of the method in the contract ABI.
const CleanupApplicationsMethodName = "cleanupApplications"

// CleanupApplicationsMethod is the implementation of the cleanupApplications method that removes
// all applications

// The method has no arguments
type CleanupApplicationsMethod struct {
	keeper PoaKeeper
}

func newCleanupApplicationsMethod(pk PoaKeeper) *CleanupApplicationsMethod {
	return &CleanupApplicationsMethod{
		keeper: pk,
	}
}

func (m *CleanupApplicationsMethod) MethodName() string {
	return SubmitApplicationMethodName
}

func (m *CleanupApplicationsMethod) MethodType() precompile.MethodType {
	return precompile.Write
}

func (m *CleanupApplicationsMethod) RequiredGas(_ []byte) (uint64, bool) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (m *CleanupApplicationsMethod) Payable() bool {
	return false
}

func (m *CleanupApplicationsMethod) Run(context *precompile.RunContext, inputs precompile.MethodInputs) (precompile.MethodOutputs, error) {
	// check method inputs
	if err := precompile.ValidateMethodInputsCount(inputs, 0); err != nil {
		return nil, err
	}

	err := m.keeper.CleanupApplications(
		context.SdkCtx(),
		precompile.TypesConverter.Address.ToSDK(context.MsgSender()),
	)
	if err != nil {
		return nil, err
	}

	// emit event
	err = context.EventEmitter().Emit(NewApplicationsCleanedEvent())
	if err != nil {
		return nil, fmt.Errorf("failed to emit ApplicationsCleaned event: [%w]", err)
	}

	return precompile.MethodOutputs{true}, nil
}

// ApplicationsCleanedName is the name of the ApplicationsCleaned event. It matches the name
// of the event in the contract ABI.
const ApplicationsCleanedEventName = "ApplicationsCleaned"

// ApplicationsCleanedEvent is the implementation of the ApplicationsCleaned event
type ApplicationsCleanedEvent struct{}

func NewApplicationsCleanedEvent() *ApplicationsCleanedEvent {
	return &ApplicationsCleanedEvent{}
}

func (e *ApplicationsCleanedEvent) EventName() string {
	return ApplicationsCleanedEventName
}

func (e *ApplicationsCleanedEvent) Arguments() []*precompile.EventArgument {
	return []*precompile.EventArgument{}
}
