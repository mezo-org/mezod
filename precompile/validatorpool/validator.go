package validatorpool

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v12/precompile"
)

// GetValidatorsMethodName is the name of the getValidators method. It matches the name
// of the method in the contract ABI.
const GetValidatorsMethodName = "getValidators"

// getValidatorsMethod is the implementation of the getValidators method that returns
// the current getValidators
type getValidatorsMethod struct {
	keeper PoaKeeper
}

func newGetValidatorsMethod(pk PoaKeeper) *getValidatorsMethod {
	return &getValidatorsMethod{
		keeper: pk,
	}
}

func (m *getValidatorsMethod) MethodName() string {
	return GetValidatorsMethodName
}

func (m *getValidatorsMethod) MethodType() precompile.MethodType {
	return precompile.Read
}

func (m *getValidatorsMethod) RequiredGas(_ []byte) (uint64, bool) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (m *getValidatorsMethod) Payable() bool {
	return false
}

func (m *getValidatorsMethod) Run(context *precompile.RunContext, inputs precompile.MethodInputs) (precompile.MethodOutputs, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 0); err != nil {
		return nil, err
	}

	validators := m.keeper.GetAllValidators(
		context.SdkCtx(),
	)

	operators := make([]common.Address, len(validators))

	for i, validator := range validators {
		operator := validator.GetOperator()
		operators[i] = precompile.TypesConverter.Address.FromSDK(types.AccAddress(operator))
	}

	return precompile.MethodOutputs{operators}, nil
}

// GetValidatorMethodName is the name of the getValidators method. It matches the name
// of the method in the contract ABI.
const GetValidatorMethodName = "getValidator"

// getValidatorMethod is the implementation of the getValidator method that returns
// the current getValidator
type getValidatorMethod struct {
	keeper PoaKeeper
}

func newGetValidatorMethod(pk PoaKeeper) *getValidatorsMethod {
	return &getValidatorsMethod{
		keeper: pk,
	}
}

func (m *getValidatorMethod) MethodName() string {
	return GetValidatorsMethodName
}

func (m *getValidatorMethod) MethodType() precompile.MethodType {
	return precompile.Read
}

func (m *getValidatorMethod) RequiredGas(_ []byte) (uint64, bool) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (m *getValidatorMethod) Payable() bool {
	return false
}

func (m *getValidatorMethod) Run(context *precompile.RunContext, inputs precompile.MethodInputs) (precompile.MethodOutputs, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 1); err != nil {
		return nil, err
	}

	operator, ok := inputs[0].(common.Address)
	if !ok {
		return nil, fmt.Errorf("operator argument must be of type common.Address")
	}

	validator, ok := m.keeper.GetValidator(
		context.SdkCtx(),
		types.ValAddress(precompile.TypesConverter.Address.ToSDK(operator)),
	)
	if !ok {
		return nil, fmt.Errorf("validator does not exist")
	}

	consPubKey := [32]byte(validator.GetConsPubKey().Bytes())

	return precompile.MethodOutputs{operator, consPubKey, validator.Description}, nil
}
