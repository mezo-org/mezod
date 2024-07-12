package validatorpool

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v12/precompile"
	poakeeper "github.com/evmos/evmos/v12/x/poa/keeper"
)

// KickMethodName is the name of the kick method. It matches the name
// of the method in the contract ABI.
const KickMethodName = "kick"

// kickMethod is the implementation of the kick method that registers
// a validator candidates application as pending

// The method has the following input arguments:
// - operator: the EVM address identifying the validator.
type kickMethod struct {
	poaKeeper poakeeper.Keeper
}

func newKickMethod(poaKeeper poakeeper.Keeper) *kickMethod {
	return &kickMethod{
		poaKeeper: poaKeeper,
	}
}

func (km *kickMethod) MethodName() string {
	return KickMethodName
}

func (km *kickMethod) MethodType() precompile.MethodType {
	return precompile.Write
}

func (km *kickMethod) RequiredGas(_ []byte) (uint64, bool) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (km *kickMethod) Payable() bool {
	return false
}

func (km *kickMethod) Run(_ *precompile.RunContext, inputs precompile.MethodInputs) (precompile.MethodOutputs, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 1); err != nil {
		return nil, err
	}

	_, ok := inputs[1].(common.Address)
	if !ok {
		return nil, fmt.Errorf("operator argument must be common.Address")
	}

	// err := km.poaKeeper.Kick(context.SdkCtx(), operator)
	// if err != nil {
	// 	return nil, err
	// }

	return nil, nil
}
