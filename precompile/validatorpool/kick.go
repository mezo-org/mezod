package validatorpool

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v12/precompile"
)

// KickMethodName is the name of the kick method. It matches the name
// of the method in the contract ABI.
const KickMethodName = "kick"

// kickMethod is the implementation of the kick method that registers
// a validator candidates application as pending

// The method has the following input arguments:
// - operator: the address identifying the validator.
type kickMethod struct {
	keeper ValidatorPool
}

func newKickMethod(vp ValidatorPool) *kickMethod {
	return &kickMethod{
		keeper: vp,
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

func (km *kickMethod) Run(context *precompile.RunContext, inputs precompile.MethodInputs) (precompile.MethodOutputs, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 1); err != nil {
		return nil, err
	}

	operator, ok := inputs[1].(common.Address)
	if !ok {
		return nil, fmt.Errorf("operator argument must be common.Address")
	}

	err := km.keeper.Kick(
		context.SdkCtx(),
		precompile.TypesConverter.Address.ToSDK(context.MsgSender()),
		types.ValAddress(precompile.TypesConverter.Address.ToSDK(operator)),
	)
	if err != nil {
		return nil, err
	}

	// emit event
	err = context.EventEmitter().Emit(
		newValidatorKickedEvent(operator),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to emitv ValidatorKicked event: [%w]", err)
	}

	return precompile.MethodOutputs{true}, nil
}
