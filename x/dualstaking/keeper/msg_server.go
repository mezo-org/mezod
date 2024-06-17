package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"github.com/evmos/evmos/v12/x/dualstaking/types"

	"context"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	// "github.com/tharsis/evmos/x/dualstaking/types"
	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/types/errors"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ types.MsgServer = msgServer{}

type msgServer struct {
	Keeper
}

// NewMsgServer returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServer(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

func (m msgServer) StakeTokens(goCtx context.Context, msg *types.MsgStakeTokens) (*types.MsgStakeTokensResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	lockDuration := time.Duration(msg.LockDuration) * time.Second
	if lockDuration < time.Duration(types.MinLockPeriod) || lockDuration > time.Duration(types.MaxLockPeriod) {
		return nil, errorsmod.Wrap(errortypes.ErrInvalidRequest, "lock duration must be between 1 week and 4 years")
	}

	var veTokenDenom string
	switch msg.Amount.Denom {
	case "BTC":
		veTokenDenom = "veBTC"
	case "MEZO":
		veTokenDenom = "veMEZO"
	default:
		return nil, errorsmod.Wrap(errortypes.ErrInvalidRequest, "unsupported token")
	}

	staker, err := sdk.AccAddressFromBech32(msg.Staker)
	if err != nil {
		return nil, errorsmod.Wrap(errortypes.ErrInvalidAddress, "invalid staker address")
	}

	err = m.bankKeeper.SendCoinsFromAccountToModule(ctx, staker, types.ModuleName, sdk.NewCoins(*msg.Amount))
	if err != nil {
		return nil, err
	}

	stakeId := fmt.Sprintf("%s-%d", staker.String(), time.Now().UnixNano())
	position := types.StakingPosition{
		Staker:    staker.String(),
		StakeId:   stakeId,
		Denom:     veTokenDenom,
		Amount:    msg.Amount.Amount.Int64(),
		StartTime: ctx.BlockTime().Unix(),
		EndTime:   ctx.BlockTime().Unix() + msg.LockDuration,
	}
	m.Keeper.SetStakingPosition(ctx, position)

	// Mint veTokens and transfer to staker
	veAmount := sdk.NewCoin(fmt.Sprintf("ve%s", msg.Amount.Denom), msg.Amount.Amount)
	err = m.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(veAmount))
	if err != nil {
		return nil, err
	}
	err = m.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, staker, sdk.NewCoins(veAmount))
	if err != nil {
		return nil, err
	}

	return &types.MsgStakeTokensResponse{}, nil
}

func (m msgServer) DelegateTokens(goCtx context.Context, msg *types.MsgDelegateTokens) (*types.MsgDelegateTokensResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	staker, err := sdk.AccAddressFromBech32(msg.Staker)
	if err != nil {
		return nil, errorsmod.Wrap(errors.ErrInvalidAddress, "invalid staker address")
	}

	validatorAddr, err := sdk.ValAddressFromBech32(msg.Validator)
	if err != nil {
		return nil, errorsmod.Wrap(errors.ErrInvalidAddress, "invalid validator address")
	}

	// TODO: Validate denom

	delegationId := fmt.Sprintf("%s-%d", staker.String(), time.Now().UnixNano())
	position := types.DelegationPosition{
		Staker:       staker.String(),
		DelegationId: delegationId,
		Validator:    validatorAddr.String(),
		Denom:        msg.Amount.Denom,
		Amount:       msg.Amount.Amount.Int64(),
	}
	m.Keeper.SetDelegationPosition(ctx, position)

	// Delegate veTokens logic
	veAmount := sdk.NewCoin(fmt.Sprintf("ve%s", msg.Amount.Denom), msg.Amount.Amount)
	err = m.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(veAmount))
	if err != nil {
		return nil, err
	}
	err = m.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, staker, sdk.NewCoins(veAmount))
	if err != nil {
		return nil, err
	}

	// TODO: Calculate SSC amount and delegate to validator (implementation needed)

	return &types.MsgDelegateTokensResponse{}, nil
}

func (m msgServer) AutoDelegateTokens(goCtx context.Context, msg *types.MsgAutoDelegateTokens) (*types.MsgAutoDelegateTokensResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	staker, err := sdk.AccAddressFromBech32(msg.Staker)
	if err != nil {
		return nil, errorsmod.Wrap(errors.ErrInvalidAddress, "invalid staker address")
	}

	validatorAddr, err := sdk.ValAddressFromBech32(msg.Validator)
	if err != nil {
		return nil, errorsmod.Wrap(errors.ErrInvalidAddress, "invalid validator address")
	}

	// TODO: Validate denom of coins

	err = m.bankKeeper.SendCoinsFromAccountToModule(ctx, staker, types.ModuleName, sdk.NewCoins(*msg.Amount))
	if err != nil {
		return nil, err
	}

	stakeId := fmt.Sprintf("%s-%d", staker.String(), time.Now().UnixNano())
	stakingPosition := types.StakingPosition{
		Staker:    staker.String(),
		StakeId:   stakeId,
		Denom:     msg.Amount.Denom,
		Amount:    msg.Amount.Amount.Int64(),
		StartTime: ctx.BlockTime().Unix(),
		EndTime:   ctx.BlockTime().Unix() + msg.LockDuration,
	}
	m.Keeper.SetStakingPosition(ctx, stakingPosition)

	delegationId := fmt.Sprintf("%s-%d", staker.String(), time.Now().UnixNano())
	delegationPosition := types.DelegationPosition{
		Staker:       staker.String(),
		DelegationId: delegationId,
		Validator:    validatorAddr.String(),
		Denom:        msg.Amount.Denom,
		Amount:       msg.Amount.Amount.Int64(),
	}
	m.Keeper.SetDelegationPosition(ctx, delegationPosition)

	// Mint veTokens and delegate to validator
	veAmount := sdk.NewCoin(fmt.Sprintf("ve%s", msg.Amount.Denom), msg.Amount.Amount)
	err = m.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(veAmount))
	if err != nil {
		return nil, err
	}
	err = m.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, staker, sdk.NewCoins(veAmount))
	if err != nil {
		return nil, err
	}

	// TODO: Calculate SSC amount and delegate to validator (implementation needed)

	return &types.MsgAutoDelegateTokensResponse{}, nil
}

func (m msgServer) calculateSSCAmount(goCtx context.Context, amount sdk.Coin) math.Int {
	// Implement the logic to calculate SSC amount based on the amount of veToken
	return amount.Amount
}
