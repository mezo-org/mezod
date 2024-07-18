package validatorpool

import (
	"fmt"

	"github.com/evmos/evmos/v12/precompile"
)

// LeaveMethodName is the name of the leave method. It matches the name
// of the method in the contract ABI.
const LeaveMethodName = "leave"

// leaveMethod is the implementation of the leave method that removes
// msg.sender from the validator pool
type leaveMethod struct {
	keeper PoaKeeper
}

func newLeaveMethod(pk PoaKeeper) *leaveMethod {
	return &leaveMethod{
		keeper: pk,
	}
}

func (lm *leaveMethod) MethodName() string {
	return LeaveMethodName
}

func (lm *leaveMethod) MethodType() precompile.MethodType {
	return precompile.Write
}

func (lm *leaveMethod) RequiredGas(_ []byte) (uint64, bool) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (lm *leaveMethod) Payable() bool {
	return false
}

func (lm *leaveMethod) Run(context *precompile.RunContext, inputs precompile.MethodInputs) (precompile.MethodOutputs, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 0); err != nil {
		return nil, err
	}

	err := lm.keeper.Leave(
		context.SdkCtx(),
		precompile.TypesConverter.Address.ToSDK(context.MsgSender()),
	)
	if err != nil {
		return nil, err
	}

	// emit event
	err = context.EventEmitter().Emit(
		newValidatorLeftEvent(context.MsgSender()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to emit validatorLeft event: [%w]", err)
	}

	return precompile.MethodOutputs{true}, nil
}
