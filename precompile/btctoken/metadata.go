package btctoken

import (
	"fmt"

	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/evmos/evmos/v12/precompile"
	evm "github.com/evmos/evmos/v12/x/evm/types"
)

const (
	NameMethodName     = "name"
	SymbolMethodName   = "symbol"
	DecimalsMethodName = "decimals"
)

type nameMethod struct {
	bankKeeper bankkeeper.Keeper
}
type symbolMethod struct {
	bankKeeper bankkeeper.Keeper
}

type decimalsMethod struct {
	bankKeeper bankkeeper.Keeper
}

// Name method returns the name of the BTC token.
func newNameMethod(bankKeeper bankkeeper.Keeper) *nameMethod {
	return &nameMethod{
		bankKeeper: bankKeeper,
	}
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
	context *precompile.RunContext,
	inputs precompile.MethodInputs,
) (precompile.MethodOutputs, error) {
	metadata, found := nm.bankKeeper.GetDenomMetaData(context.SdkCtx(), evm.DefaultEVMDenom)
	if !found {
		return nil, fmt.Errorf("metadata name not found")
	}

	return precompile.MethodOutputs{
		metadata.Name,
	}, nil
}

// Symbol method returns the symbol of the BTC token.
func newSymbolMethod(bankKeeper bankkeeper.Keeper) *symbolMethod {
	return &symbolMethod{
		bankKeeper: bankKeeper,
	}
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
	context *precompile.RunContext,
	inputs precompile.MethodInputs,
) (precompile.MethodOutputs, error) {
	metadata, found := sm.bankKeeper.GetDenomMetaData(context.SdkCtx(), evm.DefaultEVMDenom)
	if !found {
		return nil, fmt.Errorf("metadata symbol not found")
	}

	return precompile.MethodOutputs{
		metadata.Symbol,
	}, nil
}

// Decimals method returns the number of decimals used to represent the BTC token.
func newDecimalsMethod(bankKeeper bankkeeper.Keeper) *decimalsMethod {
	return &decimalsMethod{
		bankKeeper: bankKeeper,
	}
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
	context *precompile.RunContext,
	inputs precompile.MethodInputs,
) (precompile.MethodOutputs, error) {
	metadata, found := dm.bankKeeper.GetDenomMetaData(context.SdkCtx(), evm.DefaultEVMDenom)
	if !found {
		return nil, fmt.Errorf("metadata decimals not found")
	}

	return precompile.MethodOutputs{
		metadata.DenomUnits[0].Exponent,
	}, nil
}
