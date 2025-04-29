package erc20

import (
	"fmt"

	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/mezo-org/mezod/precompile"
)

// TotalSupplyMethodName is the name of the totalSupply method.
// It matches the name of the method in the contract ABI.
const TotalSupplyMethodName = "totalSupply"

// TotalSupplyMethod is the implementation of the totalSupply method that returns
// the total supply of the ERC20 tokens in existence.
type TotalSupplyMethod struct {
	bankKeeper bankkeeper.Keeper
	denom      string
}

func NewTotalSupplyMethod(
	bankKeeper bankkeeper.Keeper,
	denom string,
) *TotalSupplyMethod {
	return &TotalSupplyMethod{
		bankKeeper: bankKeeper,
		denom:      denom,
	}
}

func (tsm *TotalSupplyMethod) MethodName() string {
	return TotalSupplyMethodName
}

func (tsm *TotalSupplyMethod) MethodType() precompile.MethodType {
	return precompile.Read
}

func (tsm *TotalSupplyMethod) RequiredGas(_ []byte) (uint64, bool) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (tsm *TotalSupplyMethod) Payable() bool {
	return false
}

func (tsm *TotalSupplyMethod) Run(
	context *precompile.RunContext,
	inputs precompile.MethodInputs,
) (precompile.MethodOutputs, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 0); err != nil {
		return nil, err
	}

	supply := tsm.bankKeeper.GetSupply(context.SdkCtx(), tsm.denom)
	if supply.Amount.IsNil() {
		return nil, fmt.Errorf("failed to get the supply amount of the %s token", tsm.denom)
	}

	return precompile.MethodOutputs{
		precompile.TypesConverter.BigInt.FromSDK(supply.Amount),
	}, nil
}
