package validatorpool

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v12/precompile"
)

// AcceptOwnershipMethodName is the name of the acceptOwnership method. It matches the name
// of the method in the contract ABI.
const AcceptOwnershipMethodName = "acceptOwnership"

// acceptOwnershipMethod is the implementation of the acceptOwnership method that accepts
// a pending ownership transfer
type acceptOwnershipMethod struct {
	keeper PoaKeeper
}

func newAcceptOwnershipMethod(pk PoaKeeper) *acceptOwnershipMethod {
	return &acceptOwnershipMethod{
		keeper: pk,
	}
}

func (m *acceptOwnershipMethod) MethodName() string {
	return AcceptOwnershipMethodName
}

func (m *acceptOwnershipMethod) MethodType() precompile.MethodType {
	return precompile.Write
}

func (m *acceptOwnershipMethod) RequiredGas(_ []byte) (uint64, bool) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (m *acceptOwnershipMethod) Payable() bool {
	return false
}

func (m *acceptOwnershipMethod) Run(context *precompile.RunContext, inputs precompile.MethodInputs) (precompile.MethodOutputs, error) {
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
		newOwnershipTransferredEvent(
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

// ownershipTransferredEvent is the implementation of the OwnershipTransferred event that contains
// the following arguments:
// - previousOwner (indexed): is the EVM address of the now previous owner
// - newOwner (indexed): is the EVM address of the new (now current) owner
type ownershipTransferredEvent struct {
	previousOwner, newOwner common.Address
}

func newOwnershipTransferredEvent(previousOwner, newOwner common.Address) *ownershipTransferredEvent {
	return &ownershipTransferredEvent{
		previousOwner: previousOwner,
		newOwner:      newOwner,
	}
}

func (e *ownershipTransferredEvent) EventName() string {
	return OwnershipTransferredEventName
}

func (e *ownershipTransferredEvent) Arguments() []*precompile.EventArgument {
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
