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
	keeper ValidatorPool
}

func newTransferOwnershipMethod(vp ValidatorPool) *transferOwnershipMethod {
	return &transferOwnershipMethod{
		keeper: vp,
	}
}

func (tom *transferOwnershipMethod) MethodName() string {
	return TransferOwnershipMethodName
}

func (tom *transferOwnershipMethod) MethodType() precompile.MethodType {
	return precompile.Write
}

func (tom *transferOwnershipMethod) RequiredGas(_ []byte) (uint64, bool) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (tom *transferOwnershipMethod) Payable() bool {
	return false
}

func (tom *transferOwnershipMethod) Run(context *precompile.RunContext, inputs precompile.MethodInputs) (precompile.MethodOutputs, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 1); err != nil {
		return nil, err
	}

	newOwner, ok := inputs[1].(common.Address)
	if !ok {
		return nil, fmt.Errorf("newOwner argument must be common.Address")
	}

	err := tom.keeper.TransferOwnership(
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
