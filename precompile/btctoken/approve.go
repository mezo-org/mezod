package btctoken

import (
	"fmt"

	"math/big"

	"errors"
	"time"

	sdkmath "cosmossdk.io/math"
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
	ApproveMethodName = "approve"
	// TODO: revisit and decide what the default expiration should be
	ApprovalExpiration = time.Hour * 24 * 365 // 1 year
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
//  1. no authorization, amount negative -> return error
//  2. no authorization, amount positive -> create a new authorization
//  3. authorization exists, amount 0 or negative -> delete authorization
//  4. authorization exists, amount positive -> update authorization
//  5. no authorizaiton, amount 0 -> no-op but still emit Approval event
func (am *approveMethod) Run(
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

	amount, ok := inputs[1].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("invalid amount: %v", inputs[1])
	}

	granter := context.MsgSender()

	authorization, expiration := am.authzkeeper.GetAuthorization(context.SdkCtx(), spender.Bytes(), granter.Bytes(), SendMsgURL)
	if authorization == nil {
		fmt.Printf("authorization to %s for address %s does not exist or is expired", SendMsgURL, spender)
	}

	var err error

	switch {
	case authorization == nil && amount != nil && amount.Sign() < 0:
		// scenario 1: no authorization, amount negative -> error
		err = errors.New("cannot approve negative values")
	case authorization == nil && amount != nil && amount.Sign() > 0:
		// scenario 2: no authorization, amount positive -> create a new authorization
		err = am.createAuthorization(context.SdkCtx(), spender, granter, amount)
	case authorization != nil && amount != nil && amount.Sign() <= 0:
		// scenario 3: authorization exists, amount 0 or negative -> remove from spend limit and delete authorization if no spend limit left
		err = am.removeSpendLimitOrDeleteAuthorization(context.SdkCtx(), spender, granter, authorization, expiration)
	case authorization != nil && amount != nil && amount.Sign() > 0:
		// scenario 4: authorization exists, amount positive -> update authorization
		sendAuthz, ok := authorization.(*banktypes.SendAuthorization)
		if !ok {
			return nil, fmt.Errorf("unknown authorization type")
		}

		err = am.updateAuthorization(context.SdkCtx(), spender, granter, amount, sendAuthz, expiration)
	}
	
	if err != nil {
		return nil, err
	}
	
	// scenario 5: no authorizaiton, amount 0 -> no-op but emit Approval event
	// TODO: emit Approval event

	return precompile.MethodOutputs{true}, nil
}

func (am approveMethod) createAuthorization(ctx sdk.Context, grantee, granter common.Address, amount *big.Int) error {
	if amount.BitLen() > sdkmath.MaxBitLen {
		return fmt.Errorf("amount %s causes integer overflow", amount)
	}

	coins := sdk.Coins{{Denom: evm.DefaultEVMDenom, Amount: sdkmath.NewIntFromBigInt(amount)}}
	
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

func (am approveMethod) updateAuthorization(ctx sdk.Context, grantee, granter common.Address, amount *big.Int, authorization *banktypes.SendAuthorization, expiration *time.Time) error {
	authorization.SpendLimit[0] = sdk.Coin{Denom: evm.DefaultEVMDenom, Amount: sdkmath.NewIntFromBigInt(amount)}
	if err := authorization.ValidateBasic(); err != nil {
		return err
	}

	return am.authzkeeper.SaveGrant(ctx, grantee.Bytes(), granter.Bytes(), authorization, expiration)
}
