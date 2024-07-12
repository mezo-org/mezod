package validatorpool

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v12/precompile"
	poakeeper "github.com/evmos/evmos/v12/x/poa/keeper"
)

// TransferOwnershipMethodName is the name of the transferOwnership method. It matches the name
// of the method in the contract ABI.
const TransferOwnershipMethodName = "transferOwnership"

// transferOwnershipMethod is the implementation of the transferOwnership method that begins
// the ownership transfer process to another account

// The method has the following input arguments:
// - newOwner: the EVM address identifying the new owner.
type transferOwnershipMethod struct {
	poaKeeper poakeeper.Keeper
}

func newTransferOwnershipMethod(poaKeeper poakeeper.Keeper) *transferOwnershipMethod {
	return &transferOwnershipMethod{
		poaKeeper: poaKeeper,
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

func (tom *transferOwnershipMethod) Run(_ *precompile.RunContext, inputs precompile.MethodInputs) (precompile.MethodOutputs, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 1); err != nil {
		return nil, err
	}

	_, ok := inputs[1].(common.Address)
	if !ok {
		return nil, fmt.Errorf("newOwner argument must be common.Address")
	}

	// err := tom.poaKeeper.TransferOwnership(context.SdkCtx(), newOwner)
	// if err != nil {
	// 	return nil, err
	// }

	return nil, nil
}
