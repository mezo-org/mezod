package validatorpool

import (
	"fmt"

	cryptocdc "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v12/precompile"
	poatypes "github.com/evmos/evmos/v12/x/poa/types"
	"github.com/tendermint/tendermint/crypto/ed25519"
)

type Description = struct {
	Moniker         string `json:"moniker"`
	Identity        string `json:"identity"`
	Website         string `json:"website"`
	SecurityContact string `json:"securityContact"`
	Details         string `json:"details"`
}

// SubmitApplicationMethodName is the name of the submitApplication method. It matches the name
// of the method in the contract ABI.
const SubmitApplicationMethodName = "submitApplication"

// SubmitApplicationMethod is the implementation of the submitApplication method that registers
// a validator candidates application as pending

// The method has the following input arguments:
// - consPubKey: the consensus public key of the validator used to vote on blocks
// - operator: the EVM address identifying the validator
// - description: the validators description info
type SubmitApplicationMethod struct {
	keeper PoaKeeper
}

func NewSubmitApplicationMethod(pk PoaKeeper) *SubmitApplicationMethod {
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

	description, ok := inputs[2].(Description)
	if !ok {
		return nil, fmt.Errorf("description argument must be Description")
	}

	// Here we assume consPubKeyBytes is a valid ED25519 key, without performing any additional validation.
	// We may need to add validation here.
	tmpk := ed25519.PubKey(consPubKeyBytes[:])
	consPubKey, err := cryptocdc.FromTmPubKeyInterface(tmpk)
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

func NewApproveApplicationMethod(pk PoaKeeper) *ApproveApplicationMethod {
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

// GetApplicationsMethodName is the name of the getApplications method. It matches the name
// of the method in the contract ABI.
const GetApplicationsMethodName = "getApplications"

// getApplicationsMethod is the implementation of the getApplications method that returns
// the current getApplications
type GetApplicationsMethod struct {
	keeper PoaKeeper
}

func NewGetApplicationsMethod(pk PoaKeeper) *GetApplicationsMethod {
	return &GetApplicationsMethod{
		keeper: pk,
	}
}

func (m *GetApplicationsMethod) MethodName() string {
	return GetApplicationsMethodName
}

func (m *GetApplicationsMethod) MethodType() precompile.MethodType {
	return precompile.Read
}

func (m *GetApplicationsMethod) RequiredGas(_ []byte) (uint64, bool) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (m *GetApplicationsMethod) Payable() bool {
	return false
}

func (m *GetApplicationsMethod) Run(context *precompile.RunContext, inputs precompile.MethodInputs) (precompile.MethodOutputs, error) {
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

// GetApplicationMethodName is the name of the getApplications method. It matches the name
// of the method in the contract ABI.
const GetApplicationMethodName = "getApplication"

// getApplicationMethod is the implementation of the getApplication method that returns
// the current getApplication
type GetApplicationMethod struct {
	keeper PoaKeeper
}

func NewGetApplicationMethod(pk PoaKeeper) *GetApplicationsMethod {
	return &GetApplicationsMethod{
		keeper: pk,
	}
}

func (m *GetApplicationMethod) MethodName() string {
	return GetApplicationsMethodName
}

func (m *GetApplicationMethod) MethodType() precompile.MethodType {
	return precompile.Read
}

func (m *GetApplicationMethod) RequiredGas(_ []byte) (uint64, bool) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (m *GetApplicationMethod) Payable() bool {
	return false
}

func (m *GetApplicationMethod) Run(context *precompile.RunContext, inputs precompile.MethodInputs) (precompile.MethodOutputs, error) {
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

	return precompile.MethodOutputs{operator, consPubKey, val.Description}, nil
}
