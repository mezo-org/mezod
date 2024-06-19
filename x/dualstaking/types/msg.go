package types

import (
  sdk "github.com/cosmos/cosmos-sdk/types"
  sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	errorsmod "cosmossdk.io/errors"
)

var (
	_ sdk.Msg = &MsgStakeTokens{}
	_ sdk.Msg = &MsgDelegateTokens{}
	_ sdk.Msg = &MsgStakeDelegateTokens{}
)

// Route implements the sdk.Msg interface.
func (msg *MsgStakeTokens) Route() string { return RouterKey }

// Type implements the sdk.Msg interface.
func (msg *MsgStakeTokens) Type() string { return "StakeTokens" }

// GetSigners implements the sdk.Msg interface.
func (msg *MsgStakeTokens) GetSigners() []sdk.AccAddress {
  staker, err := sdk.AccAddressFromBech32(msg.Staker)
  if err != nil {
    panic(err)
  }
  return []sdk.AccAddress{staker}
}

// GetSignBytes implements the sdk.Msg interface.
func (msg *MsgStakeTokens) GetSignBytes() []byte {
  bz := ModuleCdc.MustMarshalJSON(msg)
  return sdk.MustSortJSON(bz)
}

// ValidateBasic implements the sdk.Msg interface.
func (msg *MsgStakeTokens) ValidateBasic() error {
  if msg.Staker == "" {
    return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "staker address cannot be empty")
  }
  if !msg.Amount.IsValid() {
    return errorsmod.Wrap(sdkerrors.ErrInvalidCoins, "invalid amount")
  }
  if msg.LockDuration <= 0 {
    return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "lock duration must be positive")
  }
  return nil
}

// MsgDeleageTokens implementation

// Route implements the sdk.Msg interface.
func (msg *MsgDelegateTokens) Route() string { return RouterKey }

// Type implements the sdk.Msg interface.
func (msg *MsgDelegateTokens) Type() string { return "DelegateTokens" }

// GetSigners implements the sdk.Msg interface.
func (msg *MsgDelegateTokens) GetSigners() []sdk.AccAddress {
	delegator, err := sdk.AccAddressFromBech32(msg.Staker)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{delegator}
}

// GetSignBytes implements the sdk.Msg interface.
func (msg *MsgDelegateTokens) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements the sdk.Msg interface.
func (msg *MsgDelegateTokens) ValidateBasic() error {
	if msg.Validator == "" {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "validator address cannot be empty")
	}
	if !msg.Amount.IsValid() {
		return errorsmod.Wrap(sdkerrors.ErrInvalidCoins, "invalid amount")
	}
	return nil
}

// MsgStakeDelegateTokens implementation

// Route implements the sdk.Msg interface.
func (msg *MsgStakeDelegateTokens) Route() string { return RouterKey }

// Type implements the sdk.Msg interface.
func (msg *MsgStakeDelegateTokens) Type() string { return "StakeDelegateTokens" }

// GetSigners implements the sdk.Msg interface.
func (msg *MsgStakeDelegateTokens) GetSigners() []sdk.AccAddress {
  staker, err := sdk.AccAddressFromBech32(msg.Staker)
  if err != nil {
    panic(err)
  }
  return []sdk.AccAddress{staker}
}

// GetSignBytes implements the sdk.Msg interface.
func (msg *MsgStakeDelegateTokens) GetSignBytes() []byte {
  bz := ModuleCdc.MustMarshalJSON(msg)
  return sdk.MustSortJSON(bz)
}

// ValidateBasic implements the sdk.Msg interface.
func (msg *MsgStakeDelegateTokens) ValidateBasic() error {
  if msg.Staker == "" {
    return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "staker address cannot be empty")
  }
  if !msg.Amount.IsValid() {
    return errorsmod.Wrap(sdkerrors.ErrInvalidCoins, "invalid amount")
  }
  if msg.Validator == "" {
    return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "validator address cannot be empty")
  }
  if msg.LockDuration <= 0 {
    return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "lock duration must be positive")
  }
  return nil
}

