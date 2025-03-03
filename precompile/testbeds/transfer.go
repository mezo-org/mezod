package testbeds

import (
	"errors"
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/holiman/uint256"
	"github.com/mezo-org/mezod/precompile"
	evm "github.com/mezo-org/mezod/x/evm/types"
)

const (
	TransferMethodName           = "transfer"
	TransferWithRevertMethodName = "transferWithRevert"
)

// SendMsgURL defines the authorization type for MsgSend
var SendMsgURL = sdk.MsgTypeURL(&banktypes.MsgSend{})

type transferMethod struct {
	bankKeeper  bankkeeper.Keeper
	authzkeeper authzkeeper.Keeper
}

func newTransferMethod(
	bankKeeper bankkeeper.Keeper,
	authzkeeper authzkeeper.Keeper,
) *transferMethod {
	return &transferMethod{
		bankKeeper:  bankKeeper,
		authzkeeper: authzkeeper,
	}
}

func (tm *transferMethod) MethodName() string {
	return TransferMethodName
}

func (tm *transferMethod) MethodType() precompile.MethodType {
	return precompile.Write
}

func (tm *transferMethod) RequiredGas(_ []byte) (uint64, bool) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (tm *transferMethod) Payable() bool {
	return false
}

func (tm *transferMethod) Run(
	context *precompile.RunContext,
	inputs precompile.MethodInputs,
) (precompile.MethodOutputs, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 2); err != nil {
		return nil, err
	}

	from := context.MsgSender()

	to, ok := inputs[0].(common.Address)
	if !ok {
		return nil, fmt.Errorf("invalid to address: %v", inputs[0])
	}

	amount, ok := inputs[1].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("invalid amount: %v", inputs[1])
	}
	if amount == nil || amount.Sign() < 0 {
		amount = big.NewInt(0)
	}

	return transfer(context, tm.bankKeeper, tm.authzkeeper, from, to, amount)
}

func transfer(context *precompile.RunContext, bankKeeper bankkeeper.Keeper, authzkeeper authzkeeper.Keeper, from, to common.Address, amount *big.Int) (precompile.MethodOutputs, error) {
	if amount.Sign() > 0 {
		sdkAmount, err := precompile.TypesConverter.BigInt.ToSDK(amount)
		if err != nil {
			return nil, fmt.Errorf("failed to convert amount: [%w]", err)
		}
		coins := sdk.Coins{{Denom: evm.DefaultEVMDenom, Amount: sdkAmount}}

		msg := banktypes.NewMsgSend(from.Bytes(), to.Bytes(), coins)

		spenderAddr := context.MsgSender()
		spender := sdk.AccAddress(spenderAddr.Bytes())

		if spender.Equals(sdk.AccAddress(from.Bytes())) {
			// owner is spender
			msgSrv := bankkeeper.NewMsgServerImpl(bankKeeper)
			_, err = msgSrv.Send(context.SdkCtx(), msg)
		} else {
			authorization, _ := authzkeeper.GetAuthorization(context.SdkCtx(), spender.Bytes(), from.Bytes(), SendMsgURL)
			if authorization == nil {
				return nil, fmt.Errorf("%s authorization type does not exist or is expired for address %s", SendMsgURL, spender)
			}

			_, ok := authorization.(*banktypes.SendAuthorization)
			if !ok {
				return nil, fmt.Errorf(
					"expected authorization to be a %T", banktypes.SendAuthorization{},
				)
			}

			_, err = authzkeeper.DispatchActions(context.SdkCtx(), spender, []sdk.Msg{msg})
		}

		if err != nil {
			return nil, err
		}
	}

	// Emit Transfer event.
	err := context.EventEmitter().Emit(
		NewTransferEvent(from, to, amount),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to emit transfer event: [%w]", err)
	}

	balanceDelta, overflow := uint256.FromBig(amount)
	if overflow {
		return nil, fmt.Errorf("conversion from big.Int to uint256.Int overflowed: %v", amount)
	}

	j := context.Journal()
	// update our from and to balance by setting properly the state
	// in the state DB
	j.SubBalance(from, balanceDelta, tracing.BalanceChangeTransfer)
	j.AddBalance(to, balanceDelta, tracing.BalanceChangeTransfer)

	return precompile.MethodOutputs{true}, nil
}

// TransferEventName is the name of the Transfer event. It matches the name
// of the event in the contract ABI.
const TransferEventName = "Transfer"

// transferEvent is the implementation of the Transfer event that contains
// the following arguments:
// - from (indexed): the address from which the tokens are transferred,
// - to (indexed): the address to which the tokens are transferred,
// - value (non-indexed): the amount of tokens transferred.
type TransferEvent struct {
	from, to common.Address
	value    *big.Int
}

func NewTransferEvent(from, to common.Address, value *big.Int) *TransferEvent {
	return &TransferEvent{
		from:  from,
		to:    to,
		value: value,
	}
}

func (te *TransferEvent) EventName() string {
	return TransferEventName
}

func (te *TransferEvent) Arguments() []*precompile.EventArgument {
	return []*precompile.EventArgument{
		{
			Indexed: true,
			Value:   te.from,
		},
		{
			Indexed: true,
			Value:   te.to,
		},
		{
			Indexed: false,
			Value:   te.value,
		},
	}
}

type transferWithRevertMethod struct {
	bankKeeper  bankkeeper.Keeper
	authzkeeper authzkeeper.Keeper
}

func newTransferWithRevertMethod(
	bankKeeper bankkeeper.Keeper,
	authzkeeper authzkeeper.Keeper,
) *transferWithRevertMethod {
	return &transferWithRevertMethod{
		bankKeeper:  bankKeeper,
		authzkeeper: authzkeeper,
	}
}

func (tm *transferWithRevertMethod) MethodName() string {
	return TransferWithRevertMethodName
}

func (tm *transferWithRevertMethod) MethodType() precompile.MethodType {
	return precompile.Write
}

func (tm *transferWithRevertMethod) RequiredGas(_ []byte) (uint64, bool) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (tm *transferWithRevertMethod) Payable() bool {
	return false
}

func (tm *transferWithRevertMethod) Run(
	context *precompile.RunContext,
	inputs precompile.MethodInputs,
) (precompile.MethodOutputs, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 2); err != nil {
		return nil, err
	}

	from := context.MsgSender()

	to, ok := inputs[0].(common.Address)
	if !ok {
		return nil, fmt.Errorf("invalid to address: %v", inputs[0])
	}

	amount, ok := inputs[1].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("invalid amount: %v", inputs[1])
	}
	if amount == nil || amount.Sign() < 0 {
		amount = big.NewInt(0)
	}

	return transfer2(context, tm.bankKeeper, tm.authzkeeper, from, to, amount)
}

func transfer2(context *precompile.RunContext, bankKeeper bankkeeper.Keeper, authzkeeper authzkeeper.Keeper, from, to common.Address, amount *big.Int) (precompile.MethodOutputs, error) {
	if amount.Sign() > 0 {
		sdkAmount, err := precompile.TypesConverter.BigInt.ToSDK(amount)
		if err != nil {
			return nil, fmt.Errorf("failed to convert amount: [%w]", err)
		}
		coins := sdk.Coins{{Denom: evm.DefaultEVMDenom, Amount: sdkAmount}}

		msg := banktypes.NewMsgSend(from.Bytes(), to.Bytes(), coins)

		spenderAddr := context.MsgSender()
		spender := sdk.AccAddress(spenderAddr.Bytes())

		if spender.Equals(sdk.AccAddress(from.Bytes())) {
			// owner is spender
			msgSrv := bankkeeper.NewMsgServerImpl(bankKeeper)
			_, err = msgSrv.Send(context.SdkCtx(), msg)
		} else {
			authorization, _ := authzkeeper.GetAuthorization(context.SdkCtx(), spender.Bytes(), from.Bytes(), SendMsgURL)
			if authorization == nil {
				return nil, fmt.Errorf("%s authorization type does not exist or is expired for address %s", SendMsgURL, spender)
			}

			_, ok := authorization.(*banktypes.SendAuthorization)
			if !ok {
				return nil, fmt.Errorf(
					"expected authorization to be a %T", banktypes.SendAuthorization{},
				)
			}

			_, err = authzkeeper.DispatchActions(context.SdkCtx(), spender, []sdk.Msg{msg})
		}

		if err != nil {
			return nil, err
		}
	}

	fmt.Printf("\n\n\n\n\nTRANSFER WITH REVERT STRIPPED ERC20!!!!!!!!\n\n\n\n")

	// Emit Transfer event.
	err := context.EventEmitter().Emit(
		NewTransferEvent(from, to, amount),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to emit transfer event: [%w]", err)
	}

	return precompile.MethodOutputs{}, errors.New("unexpected error")
}
