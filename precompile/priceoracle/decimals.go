package priceoracle

import "github.com/mezo-org/mezod/precompile"

// Decimals denotes the decimal places of the precision used to represent the price.
// E.g. if decimals is 18, the price is represented with the 10^18 precision.
const Decimals = uint8(18)

// DecimalsMethodName is the name of the decimals method. It matches the name
// of the method in the contract ABI.
const DecimalsMethodName = "decimals"

// DecimalsMethod is the implementation of the decimals method.
type DecimalsMethod struct {}

func newDecimalsMethod() *DecimalsMethod {
	return &DecimalsMethod{}
}

func (m *DecimalsMethod) MethodName() string {
	return DecimalsMethodName
}

func (m *DecimalsMethod) MethodType() precompile.MethodType {
	return precompile.Read
}

func (m *DecimalsMethod) RequiredGas(_ []byte) (uint64, bool) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (m *DecimalsMethod) Payable() bool {
	return false
}

func (m *DecimalsMethod) Run(
	_ *precompile.RunContext,
	inputs precompile.MethodInputs,
) (precompile.MethodOutputs, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 0); err != nil {
		return nil, err
	}

	return precompile.MethodOutputs{Decimals}, nil
}
