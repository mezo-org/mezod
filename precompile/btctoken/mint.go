package btctoken

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v12/precompile"
	evm "github.com/evmos/evmos/v12/x/evm/types"
	"math/big"
)

const MintMethodName = "mint"

// TODO: This implementation is a playground for now. It should be replaced with
//       the actual implementation of the mint method. The actual implementation
//       should be controlled by the bridge account that will have
//       the mint authority.
type mintMethod struct {
	bankKeeper bankkeeper.Keeper
}

func newMintMethod(
	bankKeeper bankkeeper.Keeper,
) *mintMethod {
	return &mintMethod{
		bankKeeper: bankKeeper,
	}
}

func (mm *mintMethod) MethodName() string {
	return MintMethodName
}

func (mm *mintMethod) MethodType() precompile.MethodType {
	return precompile.Write
}

func (mm *mintMethod) RequiredGas(methodInputArgs []byte) (uint64, bool) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (mm *mintMethod) Payable() bool {
	return false
}

func (mm *mintMethod) Run(
	context *precompile.RunContext,
	inputs precompile.MethodInputs,
) (precompile.MethodOutputs, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 2); err != nil {
		return nil, err
	}

	recipient := inputs[0].(common.Address)
	amount := inputs[1].(*big.Int)

	sdkAmount, err := precompile.TypesConverter.BigInt.ToSDK(amount)
	if err != nil {
		return nil, fmt.Errorf("failed to convert amount: [%w]", err)
	}

	coins := sdk.NewCoins(
		sdk.NewCoin(
			// TODO: This is normally taken from EVM module's parameters.
			//       Let's make a shortcut for now.
			evm.DefaultEVMDenom,
			sdkAmount,
		),
	)

	err = mm.bankKeeper.MintCoins(
		context.SdkCtx(),
		evm.ModuleName,
		coins,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to mint coins: [%w]", err)
	}

	err = mm.bankKeeper.SendCoinsFromModuleToAccount(
		context.SdkCtx(),
		evm.ModuleName,
		precompile.TypesConverter.Address.ToSDK(recipient),
		coins,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to send minted coins to recipient: [%w]", err)
	}

	err = context.EventEmitter().Emit(
		newTransferEvent(
			common.BigToAddress(big.NewInt(0)),
			recipient,
			amount,
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to emit transfer event: [%w]", err)
	}

	return precompile.MethodOutputs{true}, nil
}

