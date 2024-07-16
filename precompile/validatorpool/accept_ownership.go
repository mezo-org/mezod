package validatorpool

import (
	"github.com/evmos/evmos/v12/precompile"
)

// AcceptOwnershipMethodName is the name of the acceptOwnership method. It matches the name
// of the method in the contract ABI.
const AcceptOwnershipMethodName = "acceptOwnership"

// acceptOwnershipMethod is the implementation of the acceptOwnership method that accepts
// a pending ownership transfer
type acceptOwnershipMethod struct {
	keeper ValidatorPool
}

func newAcceptOwnershipMethod(vp ValidatorPool) *acceptOwnershipMethod {
	return &acceptOwnershipMethod{
		keeper: vp,
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

func (aom *acceptOwnershipMethod) Run(context *precompile.RunContext, inputs precompile.MethodInputs) (precompile.MethodOutputs, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 0); err != nil {
		return nil, err
	}

	err := aom.keeper.AcceptOwnership(
		context.SdkCtx(),
		precompile.TypesConverter.Address.ToSDK(context.MsgSender()),
	)
	if err != nil {
		return nil, err
	}

	return precompile.MethodOutputs{true}, nil
}
