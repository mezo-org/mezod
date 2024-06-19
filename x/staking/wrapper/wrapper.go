package wrapper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

type WrappedMsgServer struct {
	ms types.MsgServer
}

// Returns a pointer to a new WrappedMsgServer
func NewWrappedMsgServer(_ms types.MsgServer) *WrappedMsgServer {
	return &WrappedMsgServer{
		ms: _ms,
	}
}

// CreateValidator defines a method for creating a new validator.
func (w *WrappedMsgServer) CreateValidator(ctx context.Context, req *types.MsgCreateValidator) (*types.MsgCreateValidatorResponse, error) {
	return w.ms.CreateValidator(ctx, req)
}

// EditValidator defines a method for editing an existing validator.
func (w *WrappedMsgServer) EditValidator(ctx context.Context, req *types.MsgEditValidator) (*types.MsgEditValidatorResponse, error) {
	return w.ms.EditValidator(ctx, req)
}

// Delegate defines a method for performing a delegation of coins
// from a delegator to a validator.
func (w *WrappedMsgServer) Delegate(ctx context.Context, req *types.MsgDelegate) (*types.MsgDelegateResponse, error) {
	return w.ms.Delegate(ctx, req)
}

// BeginRedelegate defines a method for performing a redelegation
// of coins from a delegator and source validator to a destination validator.
func (w *WrappedMsgServer) BeginRedelegate(ctx context.Context, req *types.MsgBeginRedelegate) (*types.MsgBeginRedelegateResponse, error) {
	return w.ms.BeginRedelegate(ctx, req)
}

// Undelegate defines a method for performing an undelegation from a
// delegate and a validator.
func (w *WrappedMsgServer) Undelegate(ctx context.Context, req *types.MsgUndelegate) (*types.MsgUndelegateResponse, error) {
	return w.ms.Undelegate(ctx, req)
}

// CancelUnbondingDelegation defines a method for performing canceling the unbonding delegation
// and delegate back to previous validator.
//
// Since: cosmos-sdk 0.46
func (w *WrappedMsgServer) CancelUnbondingDelegation(ctx context.Context, req *types.MsgCancelUnbondingDelegation) (*types.MsgCancelUnbondingDelegationResponse, error) {
	return w.ms.CancelUnbondingDelegation(ctx, req)
}
