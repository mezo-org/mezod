package btctoken

import (
	"fmt"
	"math/big"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v12/precompile"
	evm "github.com/evmos/evmos/v12/x/evm/types"
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
type approveMethod struct {
	bankKeeper  bankkeeper.Keeper
	authzkeeper authzkeeper.Keeper
}

func newApproveMethod(
	bankKeeper bankkeeper.Keeper,
	authzkeeper authzkeeper.Keeper,
) *approveMethod {
	return &approveMethod{
		bankKeeper:  bankKeeper,
		authzkeeper: authzkeeper,
	}
}

func (am *approveMethod) MethodName() string {
	return ApproveMethodName
}

func (am *approveMethod) MethodType() precompile.MethodType {
	return precompile.Write
}

func (am *approveMethod) RequiredGas(_ []byte) (uint64, bool) {
	// Fallback to the default gas calculation.
	return 0, false
}

func (am *approveMethod) Payable() bool {
	return false
}

// Approve sets the given amount as the allowance of the spender address over
// BTC coin. It returns a boolean value when the operation succeeded.
//
// The Approve method handles the following cases:
// 1. no authorization, amount 0 -> return error
// 2. no authorization, amount positive -> create a new authorization
// 3. authorization exists, amount 0 -> delete authorization
// 4. authorization exists, amount positive -> update authorization
func (am *approveMethod) Run(
	context *precompile.RunContext,
	inputs precompile.MethodInputs,
) (precompile.MethodOutputs, error) {
	logger := context.SdkCtx().Logger()
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
	if amount == nil || amount.Sign() < 0 {
		amount = big.NewInt(0)
	}

	granter := context.MsgSender()

	authorization, expiration := am.authzkeeper.GetAuthorization(context.SdkCtx(), spender.Bytes(), granter.Bytes(), SendMsgURL)

	var err error

	if authorization == nil {
		logger.Debug("authorization to %s for address %s does not exist or is expired", SendMsgURL, spender)
		if amount.Sign() == 0 {
			// no authorization, amount 0 -> error
			err = fmt.Errorf("no existing approvals, cannot approve 0")
		} else {
			// no authorization, amount positive -> create a new authorization
			err = am.createAuthorization(context.SdkCtx(), spender, granter, amount)
		}
	} else {
		if amount.Sign() == 0 {
			// authorization exists, amount 0 -> remove from spend limit and delete authorization if no spend limit left
			err = am.removeSpendLimitOrDeleteAuthorization(context.SdkCtx(), spender, granter, authorization, expiration)
		} else {
			// authorization exists, amount positive -> update authorization
			err = am.updateAuthorization(context.SdkCtx(), spender, granter, amount, authorization, expiration)
		}
	}

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

func (am approveMethod) createAuthorization(ctx sdk.Context, grantee, granter common.Address, amount *big.Int) error {
	sdkAmount, err := precompile.TypesConverter.BigInt.ToSDK(amount)
	if err != nil {
		return fmt.Errorf("failed to convert amount: [%w]", err)
	}

	coins := sdk.Coins{{Denom: evm.DefaultEVMDenom, Amount: sdkAmount}}

	expiration := ctx.BlockTime().Add(ApprovalExpiration)

	authorization := banktypes.NewSendAuthorization(coins)
	if err := authorization.ValidateBasic(); err != nil {
		return err
	}

	return am.authzkeeper.SaveGrant(ctx, grantee.Bytes(), granter.Bytes(), authorization, &expiration)
}

// Removes the spend limit for BTC and update the grant or delete the
// authorization if no spend limit in another denomination is set.
func (am approveMethod) removeSpendLimitOrDeleteAuthorization(ctx sdk.Context, grantee, granter common.Address, authorization authz.Authorization, expiration *time.Time) error {
	sendAuthz, ok := authorization.(*banktypes.SendAuthorization)
	if !ok {
		return fmt.Errorf("unknown authorization type")
	}

	found, denomCoins := sendAuthz.SpendLimit.Find(evm.DefaultEVMDenom)
	if !found {
		return fmt.Errorf("allowance for token %s does not exist", evm.DefaultEVMDenom)
	}

	newSpendLimit, hasNeg := sendAuthz.SpendLimit.SafeSub(denomCoins)
	if hasNeg {
		return fmt.Errorf("subtracted value cannot be greater than existing allowance for denom %s: %s > %s", evm.DefaultEVMDenom, denomCoins, sendAuthz.SpendLimit)
	}

	if newSpendLimit.IsZero() {
		return am.authzkeeper.DeleteGrant(ctx, grantee.Bytes(), granter.Bytes(), SendMsgURL)
	}

	sendAuthz.SpendLimit = newSpendLimit
	return am.authzkeeper.SaveGrant(ctx, grantee.Bytes(), granter.Bytes(), sendAuthz, expiration)
}

func (am approveMethod) updateAuthorization(ctx sdk.Context, grantee, granter common.Address, amount *big.Int, authorization authz.Authorization, expiration *time.Time) error {
	sendAuthz, ok := authorization.(*banktypes.SendAuthorization)
	if !ok {
		return fmt.Errorf("unknown authorization type")
	}

	sdkAmount, err := precompile.TypesConverter.BigInt.ToSDK(amount)
	if err != nil {
		return fmt.Errorf("failed to convert amount: [%w]", err)
	}

	sendAuthz.SpendLimit[0] = sdk.Coin{Denom: evm.DefaultEVMDenom, Amount: sdkAmount}
	if err := sendAuthz.ValidateBasic(); err != nil {
		return err
	}

	return am.authzkeeper.SaveGrant(ctx, grantee.Bytes(), granter.Bytes(), sendAuthz, expiration)
}

func isZeroAddress(address common.Address) bool {
	return address == common.Address{}
}
