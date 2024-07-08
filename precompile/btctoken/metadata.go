package btctoken

import (
	"fmt"

	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/evmos/evmos/v12/precompile"
	evm "github.com/evmos/evmos/v12/x/evm/types"
)

const NameMethodName = "name"

type nameMethod struct {
	bankKeeper bankkeeper.Keeper
}

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
		return nil, fmt.Errorf("metadata not found")
	}

	return precompile.MethodOutputs{
		metadata.Name,
	}, nil
}

///////////////////////////

const SymbolMethodName = "symbol"

type symbolMethod struct {
}

const DecimalsMethodName = "decimals"

type decimalsMethod struct {
}
