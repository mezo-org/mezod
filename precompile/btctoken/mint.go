package btctoken

import (
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v12/precompile"
	evm "github.com/evmos/evmos/v12/x/evm/types"
)

// TODO: This implementation is a playground for now.
// It should be replaced with the actual implementation of the mint method.

// MintMethodName is the name of the mint method. It matches the name
// of the method in the contract ABI.
const MintMethodName = "mint"

// mintMethod is the implementation of the mint method that mints BTC tokens
// to the recipient address.
//
// The method has the following input arguments:
// - recipient: the address to which the tokens are minted,
// - amount: the amount of BTC tokens minted (in 1e18 precision).
//
// The method returns true if the minting was successful.
// Otherwise, it returns nil and an error.
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

func (mm *mintMethod) RequiredGas(_ []byte) (uint64, bool) {
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

	recipient, ok := inputs[0].(common.Address)
	if !ok {
		return nil, fmt.Errorf("recipient argument must be common.Address")
	}

	amount, ok := inputs[1].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("amount argument must be *big.Int")
	}

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
		NewTransferEvent(
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
