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

// GetValidatorsMethod is the implementation of the getValidators method that returns
// the current getValidators
type GetValidatorsMethod struct {
	keeper PoaKeeper
}

func NewGetValidatorsMethod(pk PoaKeeper) *GetValidatorsMethod {
	return &GetValidatorsMethod{
		keeper: pk,
	}
}

func (m *GetValidatorsMethod) MethodName() string {
	return GetValidatorsMethodName
}

func (m *GetValidatorsMethod) MethodType() precompile.MethodType {
	return precompile.Read
}

func (m *GetValidatorsMethod) RequiredGas(_ []byte) (uint64, bool) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (m *GetValidatorsMethod) Payable() bool {
	return false
}

func (m *GetValidatorsMethod) Run(context *precompile.RunContext, inputs precompile.MethodInputs) (precompile.MethodOutputs, error) {
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

// GetValidatorMethod is the implementation of the getValidator method that returns
// the current getValidator
type GetValidatorMethod struct {
	keeper PoaKeeper
}

func NewGetValidatorMethod(pk PoaKeeper) *GetValidatorMethod {
	return &GetValidatorMethod{
		keeper: pk,
	}
}

func (m *GetValidatorMethod) MethodName() string {
	return GetValidatorMethodName
}

func (m *GetValidatorMethod) MethodType() precompile.MethodType {
	return precompile.Read
}

func (m *GetValidatorMethod) RequiredGas(_ []byte) (uint64, bool) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (m *GetValidatorMethod) Payable() bool {
	return false
}

func (m *GetValidatorMethod) Run(context *precompile.RunContext, inputs precompile.MethodInputs) (precompile.MethodOutputs, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 1); err != nil {
		return nil, err
	}

	operator, ok := inputs[0].(common.Address)
	if !ok {
		return nil, fmt.Errorf("operator argument must be of type common.Address")
	}

	validator, found := m.keeper.GetValidator(
		context.SdkCtx(),
		types.ValAddress(precompile.TypesConverter.Address.ToSDK(operator)),
	)
	if !found {
		return nil, fmt.Errorf("validator does not exist")
	}

	var consPubKey [32]byte
	copy(consPubKey[:], validator.GetConsPubKey().Bytes())

	return precompile.MethodOutputs{operator, consPubKey, validator.Description}, nil
}
