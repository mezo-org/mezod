package validatorpool

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v12/precompile"
)

// TransferOwnershipMethodName is the name of the transferOwnership method. It matches the name
// of the method in the contract ABI.
const TransferOwnershipMethodName = "transferOwnership"

// transferOwnershipMethod is the implementation of the transferOwnership method that begins
// the ownership transfer process to another account

// The method has the following input arguments:
// - newOwner: the EVM address identifying the new owner.
type transferOwnershipMethod struct {
	keeper PoaKeeper
}

func newTransferOwnershipMethod(pk PoaKeeper) *transferOwnershipMethod {
	return &transferOwnershipMethod{
		keeper: pk,
	}
}

func (m *transferOwnershipMethod) MethodName() string {
	return TransferOwnershipMethodName
}

func (m *transferOwnershipMethod) MethodType() precompile.MethodType {
	return precompile.Write
}

func (m *transferOwnershipMethod) RequiredGas(_ []byte) (uint64, bool) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (m *transferOwnershipMethod) Payable() bool {
	return false
}

func (m *transferOwnershipMethod) Run(context *precompile.RunContext, inputs precompile.MethodInputs) (precompile.MethodOutputs, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 1); err != nil {
		return nil, err
	}

	newOwner, ok := inputs[1].(common.Address)
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
		newOwnershipTransferStartedEvent(context.MsgSender(), newOwner),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to emit ownershipTransferStarted event: [%w]", err)
	}

	return precompile.MethodOutputs{true}, nil
}

// OwnershipTransferStartedName is the name of the OwnershipTransferStarted event. It matches the name
// of the event in the contract ABI.
const OwnershipTransferStartedEventName = "OwnershipTransferStarted"

// oOwnershipTransferStartedEvent is the implementation of the OwnershipTransferStarted event that contains
// the following arguments:
// - previousOwner (indexed): is the EVM address of the current (soon to be previous) owner,
// - newOwner (indexed): is the EVM address of the new owner
type ownershipTransferStartedEvent struct {
	previousOwner, newOwner common.Address
}

func newOwnershipTransferStartedEvent(previousOwner, newOwner common.Address) *ownershipTransferStartedEvent {
	return &ownershipTransferStartedEvent{
		previousOwner: previousOwner,
		newOwner:      newOwner,
	}
}

func (e *ownershipTransferStartedEvent) EventName() string {
	return OwnershipTransferStartedEventName
}

func (e *ownershipTransferStartedEvent) Arguments() []*precompile.EventArgument {
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
