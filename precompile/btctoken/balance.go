package btctoken

import (
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v12/precompile"
	evm "github.com/evmos/evmos/v12/x/evm/types"
)

// BalanceOfMethodName is the name of the balanceOf method. It matches the name
// of the method in the contract ABI.
const BalanceOfMethodName = "balanceOf"

// balanceOfMethod is the implementation of the balanceOf method that returns
// the balance of the BTC token for the given account.
//
// The method has the following input arguments:
// - account: the EVM address of the account for which the balance is returned.
//
// The method returns the BTC balance of the account (in 1e18 precision) if
// everything goes well. Otherwise, it returns nil and an error.
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

func (bom *balanceOfMethod) RequiredGas(_ []byte) (uint64, bool) {
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
	if err := precompile.ValidateMethodInputsCount(inputs, 1); err != nil {
		return nil, err
	}

	account := inputs[0].(common.Address)

	balance := bom.bankKeeper.GetBalance(
		context.SdkCtx(),
		precompile.TypesConverter.Address.ToSDK(account),
		// TODO: This is normally taken from EVM module's parameters.
		//       Let's make a shortcut for now.
		evm.DefaultEVMDenom,
	)

	return precompile.MethodOutputs{
		precompile.TypesConverter.BigInt.FromSDK(balance.Amount),
	}, nil
}
