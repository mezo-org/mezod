package validatorpool

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v12/precompile"
)

// ApproveApplicationMethodName is the name of the approveApplication method. It matches the name
// of the method in the contract ABI.
const ApproveApplicationMethodName = "approveApplication"

// approveApplicationMethod is the implementation of the approveApplication method that approves
// a pending validator application.

// The method has the following input arguments:
// - operator: the EVM address identifying the validator.
type approveApplicationMethod struct {
	keeper PoaKeeper
}

func newApproveApplicationMethod(pk PoaKeeper) *approveApplicationMethod {
	return &approveApplicationMethod{
		keeper: pk,
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

func (aam *approveApplicationMethod) Run(context *precompile.RunContext, inputs precompile.MethodInputs) (precompile.MethodOutputs, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 1); err != nil {
		return nil, err
	}

	operator, ok := inputs[0].(common.Address)
	if !ok {
		return nil, fmt.Errorf("operator argument must be of type common.Address")
	}

	err := aam.keeper.ApproveApplication(
		context.SdkCtx(),
		precompile.TypesConverter.Address.ToSDK(context.MsgSender()),
		types.ValAddress(precompile.TypesConverter.Address.ToSDK(operator)),
	)
	if err != nil {
		return nil, err
	}

	// emit events
	err = context.EventEmitter().Emit(
		newApplicationApprovedEvent(
			operator,
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to emit ApplicationApproved event: [%w]", err)
	}
	err = context.EventEmitter().Emit(
		newValidatorJoinedEvent(
			operator,
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to emit ValidatorJoined event: [%w]", err)
	}

	return precompile.MethodOutputs{true}, nil
}
