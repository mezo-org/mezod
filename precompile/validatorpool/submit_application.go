package validatorpool

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v12/precompile"
	poatypes "github.com/evmos/evmos/v12/x/poa/types"
)

// SubmitApplicationMethodName is the name of the submitApplication method. It matches the name
// of the method in the contract ABI.
const SubmitApplicationMethodName = "submitApplication"

// submitApplicationMethod is the implementation of the submitApplication method that registers
// a validator candidates application as pending

// The method has the following input arguments:
// - consPubKey: the consensus public key of the validator used to vote on blocks
// - operator: the EVM address identifying the validator.
type submitApplicationMethod struct {
	keeper PoaKeeper
}

func newSubmitApplicationMethod(pk PoaKeeper) *submitApplicationMethod {
	return &submitApplicationMethod{
		keeper: pk,
	}
}

func (sam *submitApplicationMethod) MethodName() string {
	return SubmitApplicationMethodName
}

func (sam *submitApplicationMethod) MethodType() precompile.MethodType {
	return precompile.Write
}

func (sam *submitApplicationMethod) RequiredGas(_ []byte) (uint64, bool) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (sam *submitApplicationMethod) Payable() bool {
	return false
}

func (sam *submitApplicationMethod) Run(context *precompile.RunContext, inputs precompile.MethodInputs) (precompile.MethodOutputs, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 2); err != nil {
		return nil, err
	}

	consPubKey, ok := inputs[0].(common.Address)
	if !ok {
		return nil, fmt.Errorf("consPubKey argument must be common.Address")
	}

	operator, ok := inputs[1].(common.Address)
	if !ok {
		return nil, fmt.Errorf("operator argument must be common.Address")
	}

	validator := poatypes.Validator{
		OperatorAddress: types.ValAddress(precompile.TypesConverter.Address.ToSDK(operator)),
		ConsensusPubkey: consPubKey.String(),
	}

	err := sam.keeper.SubmitApplication(
		context.SdkCtx(),
		precompile.TypesConverter.Address.ToSDK(context.MsgSender()),
		validator,
	)
	if err != nil {
		return nil, err
	}

	// emit event
	err = context.EventEmitter().Emit(
		newApplicationSubmittedEvent(
			operator,
			consPubKey,
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to emit ApplicationSubmitted event: [%w]", err)
	}

	return precompile.MethodOutputs{true}, nil
}
