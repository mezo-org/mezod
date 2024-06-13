package btctoken

import (
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v12/precompile"
	evm "github.com/evmos/evmos/v12/x/evm/types"
)

const BalanceOfMethodName = "balanceOf"

type balanceOfMethod struct {
	bankKeeper bankkeeper.Keeper
}

func newBalanceOfMethod(
	bankKeeper bankkeeper.Keeper,
) *balanceOfMethod {
	return &balanceOfMethod{
		bankKeeper: bankKeeper,
	}
}

func (bom *balanceOfMethod) MethodName() string {
	return BalanceOfMethodName
}

func (bom *balanceOfMethod) MethodType() precompile.MethodType {
	return precompile.Read
}

func (bom *balanceOfMethod) RequiredGas(methodInputArgs []byte) (uint64, bool) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (bom *balanceOfMethod) Payable() bool {
	return false
}

func (bom *balanceOfMethod) Run(
	context *precompile.RunContext,
	inputs precompile.MethodInputs,
) (precompile.MethodOutputs, error) {
	if err := precompile.ValidateMethodInputsLength(inputs, 1); err != nil {
		return nil, err
	}

	account := inputs[0].(common.Address)

	balance := bom.bankKeeper.GetBalance(
		context.SdkCtx(),
		precompile.AddressConverter{}.ToSDK(account),
		// TODO: This is normally taken from EVM module's parameters.
		//       Let's make a shortcut for now.
		evm.DefaultEVMDenom,
	)

	return precompile.MethodOutputs{
		precompile.BigIntConverter{}.FromSDK(balance.Amount),
	}, nil
}




