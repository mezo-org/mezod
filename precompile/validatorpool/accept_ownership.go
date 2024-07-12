package validatorpool

import (
	"github.com/evmos/evmos/v12/precompile"
	poakeeper "github.com/evmos/evmos/v12/x/poa/keeper"
)

// AcceptOwnershipMethodName is the name of the acceptOwnership method. It matches the name
// of the method in the contract ABI.
const AcceptOwnershipMethodName = "acceptOwnership"

// acceptOwnershipMethod is the implementation of the acceptOwnership method that accepts
// a pending ownership transfer
type acceptOwnershipMethod struct {
	poaKeeper poakeeper.Keeper
}

func newAcceptOwnershipMethod(poaKeeper poakeeper.Keeper) *acceptOwnershipMethod {
	return &acceptOwnershipMethod{
		poaKeeper: poaKeeper,
	}
}

func (aom *acceptOwnershipMethod) MethodName() string {
	return AcceptOwnershipMethodName
}

func (aom *acceptOwnershipMethod) MethodType() precompile.MethodType {
	return precompile.Write
}

func (aom *acceptOwnershipMethod) RequiredGas(_ []byte) (uint64, bool) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (aom *acceptOwnershipMethod) Payable() bool {
	return false
}

func (aom *acceptOwnershipMethod) Run(_ *precompile.RunContext, inputs precompile.MethodInputs) (precompile.MethodOutputs, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 0); err != nil {
		return nil, err
	}

	// err := aom.poaKeeper.AcceptOwnership(context.SdkCtx(), validator)
	// if err != nil {
	// 	return nil, err
	// }

	return nil, nil
}
