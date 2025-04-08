package erc20

import (
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
	evmkeeper "github.com/mezo-org/mezod/x/evm/keeper"
)

const (
	TransferMethodName     = "transfer"
	TransferFromMethodName = "transferFrom"
)

type TransferMethod struct {
	bankKeeper  bankkeeper.Keeper
	authzkeeper authzkeeper.Keeper
	evmKeeper   evmkeeper.Keeper
	denom       string
}

func NewTransferMethod(
	bankKeeper bankkeeper.Keeper,
	authzkeeper authzkeeper.Keeper,
	evmKeeper evmkeeper.Keeper,
	denom string,
) *TransferMethod {
	return &TransferMethod{
		bankKeeper:  bankKeeper,
		authzkeeper: authzkeeper,
		evmKeeper:   evmKeeper,
		denom:       denom,
	}
}

func (tm *TransferMethod) MethodName() string {
	return TransferMethodName
}

func (tm *TransferMethod) MethodType() precompile.MethodType {
	return precompile.Write
}

func (tm *TransferMethod) RequiredGas(_ []byte) (uint64, bool) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (tm *TransferMethod) Payable() bool {
	return false
}

func (tm *TransferMethod) Run(
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

	return transfer(context, tm.bankKeeper, tm.authzkeeper, tm.evmKeeper, tm.denom, from, to, amount)
}

type TransferFromMethod struct {
	bankKeeper  bankkeeper.Keeper
	authzkeeper authzkeeper.Keeper
	evmKeeper   evmkeeper.Keeper
	denom       string
}

func NewTransferFromMethod(
	bankKeeper bankkeeper.Keeper,
	authzkeeper authzkeeper.Keeper,
	evmKeeper evmkeeper.Keeper,
	denom string,
) *TransferFromMethod {
	return &TransferFromMethod{
		bankKeeper:  bankKeeper,
		authzkeeper: authzkeeper,
		evmKeeper:   evmKeeper,
		denom:       denom,
	}
}

func (tfm *TransferFromMethod) MethodName() string {
	return TransferFromMethodName
}

func (tfm *TransferFromMethod) MethodType() precompile.MethodType {
	return precompile.Write
}

func (tfm *TransferFromMethod) RequiredGas(_ []byte) (uint64, bool) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (tfm *TransferFromMethod) Payable() bool {
	return false
}

func (tfm *TransferFromMethod) Run(
	context *precompile.RunContext,
	inputs precompile.MethodInputs,
) (precompile.MethodOutputs, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 3); err != nil {
		return nil, err
	}

	from, ok := inputs[0].(common.Address)
	if !ok {
		return nil, fmt.Errorf("invalid from address: %v", inputs[0])
	}

	to, ok := inputs[1].(common.Address)
	if !ok {
		return nil, fmt.Errorf("invalid to address: %v", inputs[1])
	}

	amount, ok := inputs[2].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("invalid amount: %v", inputs[2])
	}
	if amount == nil || amount.Sign() < 0 {
		amount = big.NewInt(0)
	}

	return transfer(context, tfm.bankKeeper, tfm.authzkeeper, tfm.evmKeeper, tfm.denom, from, to, amount)
}

func transfer(
	context *precompile.RunContext,
	bankKeeper bankkeeper.Keeper,
	authzkeeper authzkeeper.Keeper,
	evmKeeper evmkeeper.Keeper,
	denom string,
	from, to common.Address,
	amount *big.Int,
) (precompile.MethodOutputs, error) {
	if amount.Sign() > 0 {
		sdkAmount, err := precompile.TypesConverter.BigInt.ToSDK(amount)
		if err != nil {
			return nil, fmt.Errorf("failed to convert amount: [%w]", err)
		}
		coins := sdk.Coins{{Denom: denom, Amount: sdkAmount}}

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

	if denom == evmKeeper.GetParams(context.SdkCtx()).EvmDenom {
		// If this precompile is tied to the native gas token of the EVM layer,
		// we MUST use the journal to propagate balance changes back to the
		// EVM stateDB. This is absolutely CRITICAL to properly sync state
		// changes between the Cosmos SDK and EVM layers.
		//
		// On the other hand, this MUST NOT be done if the precompile is tied to
		// a non-gas ERC20 token to ensure its transfers do not affect the
		// balances of the EVM native gas token for the from/to addresses.
		balanceDelta, overflow := uint256.FromBig(amount)
		if overflow {
			return nil, fmt.Errorf("conversion from big.Int to uint256.Int overflowed: %v", amount)
		}

		journal := context.Journal()
		journal.SubBalance(from, balanceDelta, tracing.BalanceChangeTransfer)
		journal.AddBalance(to, balanceDelta, tracing.BalanceChangeTransfer)
	}

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
