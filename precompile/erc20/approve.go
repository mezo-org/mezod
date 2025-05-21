package erc20

import (
	"errors"
	"fmt"
	"math/big"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/mezo-org/mezod/precompile"
)

// ApproveMethodName is the name of the approve method that should match the name
// in the contract ABI.
const (
	ApproveMethodName  = "approve"
	ApprovalExpiration = time.Hour * 24 * 365 * time.Duration(100) // 100 years
)

// SendMsgURL defines the authorization type for MsgSend
var SendMsgURL = sdk.MsgTypeURL(&banktypes.MsgSend{})

// Sets a `value` amount of tokens as the allowance of `spender` over the
// caller's tokens.
type ApproveMethod struct {
	bankKeeper  bankkeeper.Keeper
	authzkeeper authzkeeper.Keeper
	denom       string
}

func NewApproveMethod(
	bankKeeper bankkeeper.Keeper,
	authzkeeper authzkeeper.Keeper,
	denom string,
) *ApproveMethod {
	return &ApproveMethod{
		bankKeeper:  bankKeeper,
		authzkeeper: authzkeeper,
		denom:       denom,
	}
}

func (am *ApproveMethod) MethodName() string {
	return ApproveMethodName
}

func (am *ApproveMethod) MethodType() precompile.MethodType {
	return precompile.Write
}

func (am *ApproveMethod) RequiredGas(_ []byte) (uint64, bool) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (am *ApproveMethod) Payable() bool {
	return false
}

// Approve sets the given amount as the allowance of the spender address over
// the ERC20 token. It returns a boolean value when the operation succeeded.
//
// The Approve method handles the following cases:
// 1. no authorization, amount 0 -> return error
// 2. no authorization, amount positive -> create a new authorization
// 3. authorization exists, amount 0 -> delete authorization
// 4. authorization exists, amount positive -> update authorization
func (am *ApproveMethod) Run(
	context *precompile.RunContext,
	inputs precompile.MethodInputs,
) (precompile.MethodOutputs, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 2); err != nil {
		return nil, err
	}

	spender, ok := inputs[0].(common.Address)
	if !ok {
		return nil, fmt.Errorf("invalid spender address: %v", inputs[0])
	}

	if isZeroAddress(spender) {
		return nil, fmt.Errorf("spender address cannot be empty")
	}

	amount, ok := inputs[1].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("invalid amount: %v", inputs[1])
	}

	if amount == nil {
		return nil, errors.New("amount is required")
	}

	if amount.Sign() < 0 {
		return nil, errors.New("amount cannot be negative")
	}

	granter := context.MsgSender()

	authorization, expiration := am.authzkeeper.GetAuthorization(context.SdkCtx(), spender.Bytes(), granter.Bytes(), SendMsgURL)

	err := handleAuthorization(am.denom, authorization, spender, amount, context, granter, expiration, am.authzkeeper)
	if err != nil {
		return nil, err
	}

	err = context.EventEmitter().Emit(
		NewApprovalEvent(
			granter,
			spender,
			amount,
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to emit approval event: [%w]", err)
	}

	return precompile.MethodOutputs{true}, nil
}

// no authorization, amount 0 -> noop
// no authorization, amount positive -> create a new authorization
// authorization exists, amount 0 -> delete authorization
// authorization exists, amount positive -> update authorization
func handleAuthorization(
	denom string,
	authorization authz.Authorization,
	spender common.Address,
	amount *big.Int,
	context *precompile.RunContext,
	granter common.Address,
	expiration *time.Time,
	authzkeeper authzkeeper.Keeper,
) error {
	var err error

	if authorization == nil {
		if amount.Sign() == 0 {
			err = nil // this is not an error to comply with the ERC20 std behavior.
		} else {
			err = createAuthorization(context.SdkCtx(), denom, spender, granter, amount, authzkeeper)
		}
	} else {
		// updateAuthorization updates or deletes the authorization, depending on the amount.
		err = updateAuthorization(context.SdkCtx(), denom, spender, granter, amount, authorization, expiration, authzkeeper)
	}
	return err
}

func createAuthorization(
	ctx sdk.Context,
	denom string,
	grantee, granter common.Address,
	amount *big.Int,
	authzkeeper authzkeeper.Keeper,
) error {
	sdkAmount, err := precompile.TypesConverter.BigInt.ToSDK(amount)
	if err != nil {
		return fmt.Errorf("failed to convert amount: [%w]", err)
	}

	// Single-coin spend limit is always sorted so no need to sort explicitly.
	// The isSorted invariant is maintained for future updates.
	coins := sdk.Coins{{Denom: denom, Amount: sdkAmount}}

	expiration := ctx.BlockTime().Add(ApprovalExpiration)

	authorization := banktypes.NewSendAuthorization(coins, nil)
	if err := authorization.ValidateBasic(); err != nil {
		return err
	}

	return authzkeeper.SaveGrant(ctx, grantee.Bytes(), granter.Bytes(), authorization, &expiration)
}

func updateAuthorization(
	ctx sdk.Context,
	denom string,
	grantee, granter common.Address,
	amount *big.Int,
	authorization authz.Authorization,
	expiration *time.Time,
	authzkeeper authzkeeper.Keeper,
) error {
	sendAuthz, ok := authorization.(*banktypes.SendAuthorization)
	if !ok {
		return fmt.Errorf("unknown authorization type")
	}

	// Caller ensures targetAmount is >= 0.
	targetAmount, err := precompile.TypesConverter.BigInt.ToSDK(amount)
	if err != nil {
		return fmt.Errorf("failed to convert amount: [%w]", err)
	}

	// Sort the spend limit to ensure the isSorted invariant is maintained for all
	// below operations made on the spend limit. This is not strictly required if
	// the authorization is managed only by this precompile function
	// (the spend limit is sorted upon update) but may save us in case the
	// authorization is ever modified somewhere else.
	sendAuthz.SpendLimit = sendAuthz.SpendLimit.Sort()

	currentAmount := sendAuthz.SpendLimit.AmountOfNoDenomValidation(denom)
	deltaAmount := targetAmount.Sub(currentAmount)

	switch deltaAmount.Sign() {
	case 1:
		deltaCoins := sdk.Coins{{Denom: denom, Amount: deltaAmount}}
		sendAuthz.SpendLimit = sendAuthz.SpendLimit.Add(deltaCoins...)
	case -1:
		deltaCoins := sdk.Coins{{Denom: denom, Amount: deltaAmount.Neg()}}
		sendAuthz.SpendLimit = sendAuthz.SpendLimit.Sub(deltaCoins...)
	default:
		// No change so do nothing.
		return nil
	}

	// Sort the updated spend limit to ensure the isSorted invariant is maintained for future updates.
	sendAuthz.SpendLimit = sendAuthz.SpendLimit.Sort()

	if sendAuthz.SpendLimit.IsZero() {
		// If the spend limit is zero, short-circuit and delete the authorization.
		// This is necessary as a send authorization with zero spend limit is invalid
		// and will not pass the ValidateBasic check.
		return authzkeeper.DeleteGrant(ctx, grantee.Bytes(), granter.Bytes(), SendMsgURL)
	}

	if err := sendAuthz.ValidateBasic(); err != nil {
		return fmt.Errorf("failed to validate new spend authorization: [%w]", err)
	}

	return authzkeeper.SaveGrant(ctx, grantee.Bytes(), granter.Bytes(), sendAuthz, expiration)
}

func isZeroAddress(address common.Address) bool {
	return address == common.Address{}
}

// ApprovalEvent is the implementation of the Approval event that contains
// the following arguments:
// - owner (indexed): the address of ERC20 owner,
// - to (indexed): the address of spender,
// - amount (non-indexed): the amount of tokens approved.
type ApprovalEvent struct {
	from, to common.Address
	amount   *big.Int
}

// ApprovalEventName is the name of the Approval event. It matches the name
// of the event in the contract ABI.
const ApprovalEventName = "Approval"

func NewApprovalEvent(from, to common.Address, amount *big.Int) *ApprovalEvent {
	return &ApprovalEvent{
		from:   from,
		to:     to,
		amount: amount,
	}
}

func (ae *ApprovalEvent) EventName() string {
	return ApprovalEventName
}

func (ae *ApprovalEvent) Arguments() []*precompile.EventArgument {
	return []*precompile.EventArgument{
		{
			Indexed: true,
			Value:   ae.from,
		},
		{
			Indexed: true,
			Value:   ae.to,
		},
		{
			Indexed: false,
			Value:   ae.amount,
		},
	}
}
