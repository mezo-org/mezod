package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"github.com/evmos/evmos/v12/x/dualstaking/types"

	"context"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
)

var _ types.MsgServer = &Keeper{}

func (k Keeper) StakeTokens(goCtx context.Context, msg *types.MsgStakeTokens) (*types.MsgStakeTokensResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	lockDuration := time.Duration(msg.LockDuration) * time.Second
	if lockDuration < time.Duration(types.MinLockPeriod) || lockDuration > time.Duration(types.MaxLockPeriod) {
		return nil, errorsmod.Wrap(errors.ErrInvalidRequest, "lock duration must be between 1 week and 4 years")
	}

	veTokenDenom, err := convertToVeDenom(msg.Amount.Denom)
	if err != nil {
		return nil, err
	}

	staker, err := sdk.AccAddressFromBech32(msg.Staker)
	if err != nil {
		return nil, errorsmod.Wrap(errors.ErrInvalidAddress, "invalid staker address")
	}

	err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, staker, types.ModuleName, sdk.NewCoins(*msg.Amount))
	if err != nil {
		return nil, err
	}

	stakeId := fmt.Sprintf("%s-%d", staker.String(), time.Now().UnixNano())
	position := types.Stake{
		Staker:    staker.String(),
		StakeId:   stakeId,
		Denom:     veTokenDenom,
		Amount:    msg.Amount.Amount.String(),
		StartTime: uint64(ctx.BlockTime().Unix()),
		EndTime:   uint64(ctx.BlockTime().Unix()) + uint64(msg.LockDuration),
	}
	k.SetStake(ctx, position)

	// Mint veTokens and transfer to staker
	veAmount := sdk.NewCoin(fmt.Sprintf("ve%s", msg.Amount.Denom), msg.Amount.Amount)
	err = k.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(veAmount))
	if err != nil {
		return nil, err
	}
	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, staker, sdk.NewCoins(veAmount))
	if err != nil {
		return nil, err
	}

	return &types.MsgStakeTokensResponse{}, nil
}

func (k Keeper) DelegateTokens(goCtx context.Context, msg *types.MsgDelegateTokens) (*types.MsgDelegateTokensResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	staker, err := sdk.AccAddressFromBech32(msg.Staker)
	if err != nil {
		return nil, errorsmod.Wrap(errors.ErrInvalidAddress, "invalid staker address")
	}

	validatorAddr, err := sdk.ValAddressFromBech32(msg.Validator)
	if err != nil {
		return nil, errorsmod.Wrap(errors.ErrInvalidAddress, "invalid validator address")
	}

	veTokenDenom, err := convertToVeDenom(msg.Amount.Denom)
	if err != nil {
		return nil, err
	}

	delegationId := fmt.Sprintf("%s-%d", staker.String(), time.Now().UnixNano())
	position := types.Delegation{
		Staker:       staker.String(),
		DelegationId: delegationId,
		Validator:    validatorAddr.String(),
		Denom:        veTokenDenom,
		Amount:       msg.Amount.Amount.String(),
	}
	k.SetDelegation(ctx, position)

	// Delegate veTokens logic
	veAmount := sdk.NewCoin(fmt.Sprintf("ve%s", msg.Amount.Denom), msg.Amount.Amount)
	err = k.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(veAmount))
	if err != nil {
		return nil, err
	}
	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, staker, sdk.NewCoins(veAmount))
	if err != nil {
		return nil, err
	}

	// TODO: Calculate SSC amount and delegate to validator (implementation needed)

	return &types.MsgDelegateTokensResponse{}, nil
}

func (k Keeper) StakeDelegateTokens(goCtx context.Context, msg *types.MsgStakeDelegateTokens) (*types.MsgStakeDelegateTokensResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	staker, err := sdk.AccAddressFromBech32(msg.Staker)
	if err != nil {
		return nil, errorsmod.Wrap(errors.ErrInvalidAddress, "invalid staker address")
	}

	validatorAddr, err := sdk.ValAddressFromBech32(msg.Validator)
	if err != nil {
		return nil, errorsmod.Wrap(errors.ErrInvalidAddress, "invalid validator address")
	}

	veTokenDenom, err := convertToVeDenom(msg.Amount.Denom)
	if err != nil {
		return nil, err
	}

	err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, staker, types.ModuleName, sdk.NewCoins(*msg.Amount))
	if err != nil {
		return nil, err
	}

	stakeId := fmt.Sprintf("%s-%d", staker.String(), time.Now().UnixNano())
	stake := types.Stake{
		Staker:    staker.String(),
		StakeId:   stakeId,
		Denom:     veTokenDenom,
		Amount:    msg.Amount.Amount.String(),
		StartTime: uint64(ctx.BlockTime().Unix()),
		EndTime:   uint64(ctx.BlockTime().Unix() + msg.LockDuration),
	}
	k.SetStake(ctx, stake)

	delegationId := fmt.Sprintf("%s-%d", staker.String(), time.Now().UnixNano())
	delegation := types.Delegation{
		Staker:       staker.String(),
		DelegationId: delegationId,
		Validator:    validatorAddr.String(),
		Denom:        msg.Amount.Denom,
		Amount:       msg.Amount.Amount.String(),
	}
	k.SetDelegation(ctx, delegation)

	// Mint veTokens and delegate to validator
	veAmount := sdk.NewCoin(fmt.Sprintf("ve%s", msg.Amount.Denom), msg.Amount.Amount)
	err = k.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(veAmount))
	if err != nil {
		return nil, err
	}
	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, staker, sdk.NewCoins(veAmount))
	if err != nil {
		return nil, err
	}

	// TODO: Calculate SSC amount and delegate to validator (implementation needed)

	return &types.MsgStakeDelegateTokensResponse{}, nil
}

func (k Keeper) calculateSSCAmount(goCtx context.Context, amount sdk.Coin) math.Int {
	// Implement the logic to calculate SSC amount based on the amount of veToken
	return amount.Amount
}

// ConvertToVeDenom converts a staking denom to its VE counterpart
func convertToVeDenom(denom string) (string, error) {
	switch denom {
	case types.BTCDenom:
		return types.VeBTCDenom, nil
	case types.MEZODenom:
		return types.VeMEZODenom, nil
	default:
		return "", errorsmod.Wrap(errors.ErrInvalidRequest, "unsupported token")
	}
}
