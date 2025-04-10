package erc20

import (
	"fmt"

	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/ethereum/go-ethereum/common"
	"github.com/mezo-org/mezod/precompile"
)

// BalanceOfMethodName is the name of the balanceOf method. It matches the name
// of the method in the contract ABI.
const BalanceOfMethodName = "balanceOf"

// BalanceOfMethod is the implementation of the balanceOf method that returns
// the balance of the ERC20 token for the given account.
//
// The method has the following input arguments:
// - account: the EVM address of the account for which the balance is returned.
//
// The method returns the ERC20 balance of the account (in the token's precision) if
// everything goes well. Otherwise, it returns nil and an error.
type BalanceOfMethod struct {
	bankKeeper bankkeeper.Keeper
	denom      string
}

func NewBalanceOfMethod(
	bankKeeper bankkeeper.Keeper,
	denom string,
) *BalanceOfMethod {
	return &BalanceOfMethod{
		bankKeeper: bankKeeper,
		denom:      denom,
	}
}

func (bom *BalanceOfMethod) MethodName() string {
	return BalanceOfMethodName
}

func (bom *BalanceOfMethod) MethodType() precompile.MethodType {
	return precompile.Read
}

func (bom *BalanceOfMethod) RequiredGas(_ []byte) (uint64, bool) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (bom *BalanceOfMethod) Payable() bool {
	return false
}

func (bom *BalanceOfMethod) Run(
	context *precompile.RunContext,
	inputs precompile.MethodInputs,
) (precompile.MethodOutputs, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 1); err != nil {
		return nil, err
	}

	account, ok := inputs[0].(common.Address)
	if !ok {
		return nil, fmt.Errorf("account argument must be common.Address")
	}

	balance := bom.bankKeeper.GetBalance(
		context.SdkCtx(),
		precompile.TypesConverter.Address.ToSDK(account),
		bom.denom,
	)

	return precompile.MethodOutputs{
		precompile.TypesConverter.BigInt.FromSDK(balance.Amount),
	}, nil
}
