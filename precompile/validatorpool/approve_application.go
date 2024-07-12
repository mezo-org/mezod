package validatorpool

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v12/precompile"
	poakeeper "github.com/evmos/evmos/v12/x/poa/keeper"
)

// ApproveApplicationMethodName is the name of the approveApplication method. It matches the name
// of the method in the contract ABI.
const ApproveApplicationMethodName = "approveApplication"

// approveApplicationMethod is the implementation of the approveApplication method that approves
// a pending validator application.

// The method has the following input arguments:
// - operator: the EVM address identifying the validator.
type approveApplicationMethod struct {
	poaKeeper poakeeper.Keeper
}

func newApproveApplicationMethod(poaKeeper poakeeper.Keeper) *approveApplicationMethod {
	return &approveApplicationMethod{
		poaKeeper: poaKeeper,
	}
}

func (aam *approveApplicationMethod) MethodName() string {
	return ApproveApplicationMethodName
}

func (aam *approveApplicationMethod) MethodType() precompile.MethodType {
	return precompile.Write
}

func (aam *approveApplicationMethod) RequiredGas(_ []byte) (uint64, bool) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (aam *approveApplicationMethod) Payable() bool {
	return false
}

func (aam *approveApplicationMethod) Run(_ *precompile.RunContext, inputs precompile.MethodInputs) (precompile.MethodOutputs, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 1); err != nil {
		return nil, err
	}

	_, ok := inputs[0].(common.Address)
	if !ok {
		return nil, fmt.Errorf("operator argument must be common.Address")
	}

	// err := aam.poaKeeper.ApproveApplication(context.SdkCtx(), operator)
	// if err != nil {
	// 	return nil, err
	// }

	return nil, nil
}
