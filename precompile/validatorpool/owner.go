package validatorpool

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/mezo-org/mezod/precompile"
)

// OwnerMethodName is the name of the owner method. It matches the name
// of the method in the contract ABI.
const OwnerMethodName = "owner"

// OwnerMethod is the implementation of the owner method that returns
// the current owner
type OwnerMethod struct {
	keeper PoaKeeper
}

func newOwnerMethod(pk PoaKeeper) *OwnerMethod {
	return &OwnerMethod{
		keeper: pk,
	}
}

func (m *OwnerMethod) MethodName() string {
	return OwnerMethodName
}

func (m *OwnerMethod) MethodType() precompile.MethodType {
	return precompile.Read
}

func (m *OwnerMethod) RequiredGas(_ []byte) (uint64, bool) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (m *OwnerMethod) Payable() bool {
	return false
}

func (m *OwnerMethod) Run(context *precompile.RunContext, inputs precompile.MethodInputs) (precompile.MethodOutputs, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 0); err != nil {
		return nil, err
	}

	owner := m.keeper.GetOwner(
		context.SdkCtx(),
	)

	return precompile.MethodOutputs{precompile.TypesConverter.Address.FromSDK(owner)}, nil
}

// CandidateOwnerMethodName is the name of the candidateOwner method. It matches the name
// of the method in the contract ABI.
const CandidateOwnerMethodName = "candidateOwner"

// CandidateOwnerMethod is the implementation of the candidateOwner method that returns
// the pending ownership candidate
type CandidateOwnerMethod struct {
	keeper PoaKeeper
}

func newCandidateOwnerMethod(pk PoaKeeper) *CandidateOwnerMethod {
	return &CandidateOwnerMethod{
		keeper: pk,
	}
}

func (m *CandidateOwnerMethod) MethodName() string {
	return CandidateOwnerMethodName
}

func (m *CandidateOwnerMethod) MethodType() precompile.MethodType {
	return precompile.Read
}

func (m *CandidateOwnerMethod) RequiredGas(_ []byte) (uint64, bool) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (m *CandidateOwnerMethod) Payable() bool {
	return false
}

func (m *CandidateOwnerMethod) Run(context *precompile.RunContext, inputs precompile.MethodInputs) (precompile.MethodOutputs, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 0); err != nil {
		return nil, err
	}

	candidateOwner := m.keeper.GetCandidateOwner(
		context.SdkCtx(),
	)

	return precompile.MethodOutputs{precompile.TypesConverter.Address.FromSDK(candidateOwner)}, nil
}

// TransferOwnershipMethodName is the name of the transferOwnership method. It matches the name
// of the method in the contract ABI.
const TransferOwnershipMethodName = "transferOwnership"

// TransferOwnershipMethod is the implementation of the transferOwnership method that begins
// the ownership transfer process to another account

// The method has the following input arguments:
// - newOwner: the EVM address identifying the new owner.
type TransferOwnershipMethod struct {
	keeper PoaKeeper
}

func newTransferOwnershipMethod(pk PoaKeeper) *TransferOwnershipMethod {
	return &TransferOwnershipMethod{
		keeper: pk,
	}
}

func (m *TransferOwnershipMethod) MethodName() string {
	return TransferOwnershipMethodName
}

func (m *TransferOwnershipMethod) MethodType() precompile.MethodType {
	return precompile.Write
}

func (m *TransferOwnershipMethod) RequiredGas(_ []byte) (uint64, bool) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (m *TransferOwnershipMethod) Payable() bool {
	return false
}

func (m *TransferOwnershipMethod) Run(context *precompile.RunContext, inputs precompile.MethodInputs) (precompile.MethodOutputs, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 1); err != nil {
		return nil, err
	}

	newOwner, ok := inputs[0].(common.Address)
	if !ok {
		return nil, fmt.Errorf("newOwner argument must be common.Address")
	}

	err := m.keeper.TransferOwnership(
		context.SdkCtx(),
		precompile.TypesConverter.Address.ToSDK(context.MsgSender()),
		precompile.TypesConverter.Address.ToSDK(newOwner),
	)
	if err != nil {
		return nil, err
	}

	// emit ownershipTransferStarted event
	err = context.EventEmitter().Emit(
		NewOwnershipTransferStartedEvent(context.MsgSender(), newOwner),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to emit ownershipTransferStarted event: [%w]", err)
	}

	return precompile.MethodOutputs{true}, nil
}

// OwnershipTransferStartedName is the name of the OwnershipTransferStarted event. It matches the name
// of the event in the contract ABI.
const OwnershipTransferStartedEventName = "OwnershipTransferStarted"

// OwnershipTransferStartedEvent is the implementation of the OwnershipTransferStarted event that contains
// the following arguments:
// - previousOwner (indexed): is the EVM address of the current (soon to be previous) owner,
// - newOwner (indexed): is the EVM address of the new owner
type OwnershipTransferStartedEvent struct {
	previousOwner, newOwner common.Address
}

func NewOwnershipTransferStartedEvent(previousOwner, newOwner common.Address) *OwnershipTransferStartedEvent {
	return &OwnershipTransferStartedEvent{
		previousOwner: previousOwner,
		newOwner:      newOwner,
	}
}

func (e *OwnershipTransferStartedEvent) EventName() string {
	return OwnershipTransferStartedEventName
}

func (e *OwnershipTransferStartedEvent) Arguments() []*precompile.EventArgument {
	return []*precompile.EventArgument{
		{
			Indexed: true,
			Value:   e.previousOwner,
		},
		{
			Indexed: true,
			Value:   e.newOwner,
		},
	}
}

// AcceptOwnershipMethodName is the name of the acceptOwnership method. It matches the name
// of the method in the contract ABI.
const AcceptOwnershipMethodName = "acceptOwnership"

// AcceptOwnershipMethod is the implementation of the acceptOwnership method that accepts
// a pending ownership transfer
type AcceptOwnershipMethod struct {
	keeper PoaKeeper
}

func newAcceptOwnershipMethod(pk PoaKeeper) *AcceptOwnershipMethod {
	return &AcceptOwnershipMethod{
		keeper: pk,
	}
}

func (m *AcceptOwnershipMethod) MethodName() string {
	return AcceptOwnershipMethodName
}

func (m *AcceptOwnershipMethod) MethodType() precompile.MethodType {
	return precompile.Write
}

func (m *AcceptOwnershipMethod) RequiredGas(_ []byte) (uint64, bool) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (m *AcceptOwnershipMethod) Payable() bool {
	return false
}

func (m *AcceptOwnershipMethod) Run(context *precompile.RunContext, inputs precompile.MethodInputs) (precompile.MethodOutputs, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 0); err != nil {
		return nil, err
	}

	// get owner before calling AcceptOwnership in the keeper
	previousOwner := m.keeper.GetOwner(context.SdkCtx())

	err := m.keeper.AcceptOwnership(
		context.SdkCtx(),
		precompile.TypesConverter.Address.ToSDK(context.MsgSender()),
	)
	if err != nil {
		return nil, err
	}

	// emit ownershipTransferred event
	err = context.EventEmitter().Emit(
		NewOwnershipTransferredEvent(
			precompile.TypesConverter.Address.FromSDK(previousOwner),
			context.MsgSender(),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to emit ownershipTransferred event: [%w]", err)
	}

	return precompile.MethodOutputs{true}, nil
}

// OwnershipTransferredName is the name of the OwnershipTransferred event. It matches the name
// of the event in the contract ABI.
const OwnershipTransferredEventName = "OwnershipTransferred"

// OwnershipTransferredEvent is the implementation of the OwnershipTransferred event that contains
// the following arguments:
// - previousOwner (indexed): is the EVM address of the now previous owner
// - newOwner (indexed): is the EVM address of the new (now current) owner
type OwnershipTransferredEvent struct {
	previousOwner, newOwner common.Address
}

func NewOwnershipTransferredEvent(previousOwner, newOwner common.Address) *OwnershipTransferredEvent {
	return &OwnershipTransferredEvent{
		previousOwner: previousOwner,
		newOwner:      newOwner,
	}
}

func (e *OwnershipTransferredEvent) EventName() string {
	return OwnershipTransferredEventName
}

func (e *OwnershipTransferredEvent) Arguments() []*precompile.EventArgument {
	return []*precompile.EventArgument{
		{
			Indexed: true,
			Value:   e.previousOwner,
		},
		{
			Indexed: true,
			Value:   e.newOwner,
		},
	}
}
