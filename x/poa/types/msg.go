package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// verify interface at compile time
var _ sdk.Msg = &MsgUpdateParams{}
var _ sdk.Msg = &MsgSubmitApplication{}
var _ sdk.Msg = &MsgVote{}
var _ sdk.Msg = &MsgProposeKick{}
var _ sdk.Msg = &MsgLeaveValidatorSet{}

func NewMsgUpdateParams(authority string, params Params) MsgUpdateParams {
	return MsgUpdateParams{
		Authority: authority,
		Params:    params,
	}
}

func (m MsgUpdateParams) GetSigners() []sdk.AccAddress {
	addr := sdk.MustAccAddressFromBech32(m.Authority)
	return []sdk.AccAddress{addr}
}

func (m MsgUpdateParams) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return errorsmod.Wrap(err, "invalid authority address")
	}

	return m.Params.Validate()
}

func NewMsgSubmitApplication(candidate Validator) MsgSubmitApplication {
	return MsgSubmitApplication{
		Candidate: candidate,
	}
}

func (msg MsgSubmitApplication) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.AccAddress(msg.Candidate.GetOperator())}
}

// ValidateBasic validity check for the AnteHandler
func (msg MsgSubmitApplication) ValidateBasic() error {
	return msg.Candidate.CheckValid()
}

func NewMsgProposeKick(candidate sdk.ValAddress, proposer sdk.ValAddress) MsgProposeKick {
	return MsgProposeKick{
		CandidateAddr: candidate,
		ProposerAddr:  proposer,
	}
}

func (msg MsgProposeKick) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.AccAddress(msg.ProposerAddr)}
}

// ValidateBasic validity check for the AnteHandler
func (msg MsgProposeKick) ValidateBasic() error {
	if sdk.ValAddress(msg.ProposerAddr).Empty() || sdk.ValAddress(msg.CandidateAddr).Empty() {
		return sdkerrors.Wrap(ErrInvalidKickProposal, "missing address")
	}
	return nil
}

func NewMsgVote(voteType uint32, voter sdk.ValAddress, candidate sdk.ValAddress, approve bool) MsgVote {
	return MsgVote{
		VoteType:      voteType,
		VoterAddr:     voter,
		CandidateAddr: candidate,
		Approve:       approve,
	}
}

const (
	VoteTypeApplication  uint32 = iota
	VoteTypeKickProposal uint32 = iota
)

func (msg MsgVote) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.AccAddress(msg.VoterAddr)}
}

// ValidateBasic validity check for the AnteHandler
func (msg MsgVote) ValidateBasic() error {
	if sdk.ValAddress(msg.VoterAddr).Empty() || sdk.ValAddress(msg.CandidateAddr).Empty() {
		return sdkerrors.Wrap(ErrInvalidVoteMsg, "missing address")
	}
	if msg.VoteType != VoteTypeApplication && msg.VoteType != VoteTypeKickProposal {
		return sdkerrors.Wrap(ErrInvalidVoteMsg, "vote type incorrect")
	}

	return nil
}

func NewMsgLeaveValidatorSet(validatorAddr sdk.ValAddress) MsgLeaveValidatorSet {
	return MsgLeaveValidatorSet{
		ValidatorAddr: validatorAddr,
	}
}

func (msg MsgLeaveValidatorSet) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.AccAddress(msg.ValidatorAddr)}
}

// ValidateBasic validity check for the AnteHandler
func (msg MsgLeaveValidatorSet) ValidateBasic() error {
	if sdk.ValAddress(msg.ValidatorAddr).Empty() {
		return sdkerrors.Wrap(ErrInvalidValidator, "missing address")
	}

	return nil
}
