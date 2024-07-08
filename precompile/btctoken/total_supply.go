package btctoken

import (
	"fmt"

	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/evmos/evmos/v12/precompile"
	evm "github.com/evmos/evmos/v12/x/evm/types"
)

// TotalSupplyMethodName is the name of the totalSupply method. It matches the name
// of the method in the contract ABI.
const TotalSupplyMethodName = "totalSupply"

// totalSupplyMethod is the implementation of the totalSupply method that returns
// the total supply of the BTC tokens in existence.
type totalSupplyMethod struct {
	bankKeeper bankkeeper.Keeper
}

func newTotalSupplyMethod(
	bankKeeper bankkeeper.Keeper,
) *totalSupplyMethod {
	return &totalSupplyMethod{
		bankKeeper: bankKeeper,
	}
}

func (tsm *totalSupplyMethod) MethodName() string {
	return TotalSupplyMethodName
}

func (tsm *totalSupplyMethod) MethodType() precompile.MethodType {
	return precompile.Read
}

func (tsm *totalSupplyMethod) RequiredGas(_ []byte) (uint64, bool) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (tsm *totalSupplyMethod) Payable() bool {
	return false
}

func (tsm *totalSupplyMethod) Run(
	context *precompile.RunContext,
	inputs precompile.MethodInputs,
) (precompile.MethodOutputs, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 0); err != nil {
		return nil, err
	}

	supply := tsm.bankKeeper.GetSupply(context.SdkCtx(), evm.DefaultEVMDenom)
	if supply.Amount.IsNil() {
		return nil, fmt.Errorf("failed to get the supply amount of the BTC token")
	}

	return precompile.MethodOutputs{
		precompile.TypesConverter.BigInt.FromSDK(supply.Amount),
	}, nil
}
