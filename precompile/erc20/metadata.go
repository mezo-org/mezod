package erc20

import (
	"github.com/mezo-org/mezod/precompile"
)

const (
	NameMethodName     = "name"
	SymbolMethodName   = "symbol"
	DecimalsMethodName = "decimals"
)

type (
	NameMethod struct {
		name string
	}
	SymbolMethod struct {
		symbol string
	}
	DecimalsMethod struct {
		decimals uint8
	}
)

// Name method returns the name of the token.
func NewNameMethod(name string) *NameMethod {
	return &NameMethod{name: name}
}

func (nm *NameMethod) MethodName() string {
	return NameMethodName
}

func (nm *NameMethod) MethodType() precompile.MethodType {
	return precompile.Read
}

func (nm *NameMethod) RequiredGas(_ []byte) (uint64, bool) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (nm *NameMethod) Payable() bool {
	return false
}

func (nm *NameMethod) Run(
	_ *precompile.RunContext,
	inputs precompile.MethodInputs,
) (precompile.MethodOutputs, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 0); err != nil {
		return nil, err
	}

	// Return stored name
	return precompile.MethodOutputs{
		nm.name,
	}, nil
}

// Symbol method returns the symbol of the token.
func NewSymbolMethod(symbol string) *SymbolMethod {
	return &SymbolMethod{symbol: symbol}
}

func (sm *SymbolMethod) MethodName() string {
	return SymbolMethodName
}

func (sm *SymbolMethod) MethodType() precompile.MethodType {
	return precompile.Read
}

func (sm *SymbolMethod) RequiredGas(_ []byte) (uint64, bool) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (sm *SymbolMethod) Payable() bool {
	return false
}

func (sm *SymbolMethod) Run(
	_ *precompile.RunContext,
	inputs precompile.MethodInputs,
) (precompile.MethodOutputs, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 0); err != nil {
		return nil, err
	}

	// Return stored symbol
	return precompile.MethodOutputs{
		sm.symbol,
	}, nil
}

// Decimals method returns the number of decimals used to represent the token.
func NewDecimalsMethod(decimals uint8) *DecimalsMethod {
	return &DecimalsMethod{decimals: decimals}
}

func (dm *DecimalsMethod) MethodName() string {
	return DecimalsMethodName
}

func (dm *DecimalsMethod) MethodType() precompile.MethodType {
	return precompile.Read
}

func (dm *DecimalsMethod) RequiredGas(_ []byte) (uint64, bool) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (dm *DecimalsMethod) Payable() bool {
	return false
}

func (dm *DecimalsMethod) Run(
	_ *precompile.RunContext,
	inputs precompile.MethodInputs,
) (precompile.MethodOutputs, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 0); err != nil {
		return nil, err
	}

	// Return stored decimals
	return precompile.MethodOutputs{
		dm.decimals,
	}, nil
}
