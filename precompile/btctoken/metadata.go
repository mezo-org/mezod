package btctoken

import (
	"github.com/evmos/evmos/v12/precompile"
)

const (
	NameMethodName     = "name"
	SymbolMethodName   = "symbol"
	DecimalsMethodName = "decimals"
	Decimals           = uint8(18)
	Symbol             = "BTC"
	Name               = "BTC"
)

type (
	nameMethod     struct{}
	symbolMethod   struct{}
	decimalsMethod struct{}
)

// Name method returns the name of the BTC token.
func newNameMethod() *nameMethod {
	return &nameMethod{}
}

func (nm *nameMethod) MethodName() string {
	return NameMethodName
}

func (nm *nameMethod) MethodType() precompile.MethodType {
	return precompile.Read
}

func (nm *nameMethod) RequiredGas(_ []byte) (uint64, bool) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (nm *nameMethod) Payable() bool {
	return false
}

func (nm *nameMethod) Run(
	_ *precompile.RunContext,
	inputs precompile.MethodInputs,
) (precompile.MethodOutputs, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 0); err != nil {
		return nil, err
	}

	// TODO: Currently, we simplify the process of returning the name by not using
	// the bank keeper. We may need to revisit this approach in the future to determine
	// if returning the name via the bank keeper is necessary.
	return precompile.MethodOutputs{
		Name,
	}, nil
}

// Symbol method returns the symbol of the BTC token.
func newSymbolMethod() *symbolMethod {
	return &symbolMethod{}
}

func (sm *symbolMethod) MethodName() string {
	return SymbolMethodName
}

func (sm *symbolMethod) MethodType() precompile.MethodType {
	return precompile.Read
}

func (sm *symbolMethod) RequiredGas(_ []byte) (uint64, bool) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (sm *symbolMethod) Payable() bool {
	return false
}

func (sm *symbolMethod) Run(
	_ *precompile.RunContext,
	inputs precompile.MethodInputs,
) (precompile.MethodOutputs, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 0); err != nil {
		return nil, err
	}

	// TODO: Currently, we simplify the process of returning the symbol by not using
	// the bank keeper. We may need to revisit this approach in the future to determine
	// if returning the symbol via the bank keeper is necessary.
	return precompile.MethodOutputs{
		Symbol,
	}, nil
}

// Decimals method returns the number of decimals used to represent the BTC token.
func newDecimalsMethod() *decimalsMethod {
	return &decimalsMethod{}
}

func (dm *decimalsMethod) MethodName() string {
	return DecimalsMethodName
}

func (dm *decimalsMethod) MethodType() precompile.MethodType {
	return precompile.Read
}

func (dm *decimalsMethod) RequiredGas(_ []byte) (uint64, bool) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (dm *decimalsMethod) Payable() bool {
	return false
}

func (dm *decimalsMethod) Run(
	_ *precompile.RunContext,
	inputs precompile.MethodInputs,
) (precompile.MethodOutputs, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 0); err != nil {
		return nil, err
	}

	// TODO: Currently, we simplify the process of returning the decimals by not using
	// the bank keeper. We may need to revisit this approach in the future to determine
	// if returning the decimals via the bank keeper is necessary.
	return precompile.MethodOutputs{
		Decimals,
	}, nil
}
