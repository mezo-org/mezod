package validatorpool

import (
	"github.com/evmos/evmos/v12/precompile"
	poakeeper "github.com/evmos/evmos/v12/x/poa/keeper"
)

// LeaveMethodName is the name of the leave method. It matches the name
// of the method in the contract ABI.
const LeaveMethodName = "leave"

// leaveMethod is the implementation of the leave method that removes
// msg.sender from the validator pool
type leaveMethod struct {
	poaKeeper poakeeper.Keeper
}

func newLeaveMethod(poaKeeper poakeeper.Keeper) *leaveMethod {
	return &leaveMethod{
		poaKeeper: poaKeeper,
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

func (lm *leaveMethod) Run(_ *precompile.RunContext, inputs precompile.MethodInputs) (precompile.MethodOutputs, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 0); err != nil {
		return nil, err
	}

	// err := lm.poaKeeper.Leave(context.SdkCtx(), context.MsgSender())
	// if err != nil {
	// 	return nil, err
	// }

	return nil, nil
}
