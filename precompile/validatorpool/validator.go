package validatorpool

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v12/precompile"
)

// ValidatorsMethodName is the name of the validators method. It matches the name
// of the method in the contract ABI.
const ValidatorsMethodName = "validators"

// ValidatorsMethod is the implementation of the validators method that returns
// the operator addresses of all existing validators
type ValidatorsMethod struct {
	keeper PoaKeeper
}

func NewValidatorsMethod(pk PoaKeeper) *ValidatorsMethod {
	return &ValidatorsMethod{
		keeper: pk,
	}
}

func (m *ValidatorsMethod) MethodName() string {
	return ValidatorsMethodName
}

func (m *ValidatorsMethod) MethodType() precompile.MethodType {
	return precompile.Read
}

func (m *ValidatorsMethod) RequiredGas(_ []byte) (uint64, bool) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (m *ValidatorsMethod) Payable() bool {
	return false
}

func (m *ValidatorsMethod) Run(context *precompile.RunContext, inputs precompile.MethodInputs) (precompile.MethodOutputs, error) {
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

// ValidatorMethodName is the name of the validators method. It matches the name
// of the method in the contract ABI.
const ValidatorMethodName = "validator"

// ValidatorMethod is the implementation of the validator method that returns
// the validator information for the given operator address
type ValidatorMethod struct {
	keeper PoaKeeper
}

func NewValidatorMethod(pk PoaKeeper) *ValidatorMethod {
	return &ValidatorMethod{
		keeper: pk,
	}
}

func (m *ValidatorMethod) MethodName() string {
	return ValidatorMethodName
}

func (m *ValidatorMethod) MethodType() precompile.MethodType {
	return precompile.Read
}

func (m *ValidatorMethod) RequiredGas(_ []byte) (uint64, bool) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (m *ValidatorMethod) Payable() bool {
	return false
}

func (m *ValidatorMethod) Run(context *precompile.RunContext, inputs precompile.MethodInputs) (precompile.MethodOutputs, error) {
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
