package keeper

import (
	"context"
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v12/x/poa/types"
)

var _ types.MsgServer = msgServer{}

type msgServer struct {
	keeper Keeper
}

// NewMsgServer returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServer(keeper Keeper) types.MsgServer {
	return &msgServer{keeper: keeper}
}

func (ms msgServer) UpdateParams(
	ctx context.Context,
	msg *types.MsgUpdateParams,
) (*types.MsgUpdateParamsResponse, error) {
	authority, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "invalid authority address")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	if err := ms.keeper.UpdateParams(sdkCtx, authority, msg.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}

func (ms msgServer) SubmitApplication(
	ctx context.Context,
	msg *types.MsgSubmitApplication,
) (*types.MsgSubmitApplicationResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	err := ms.keeper.SubmitApplication(sdkCtx, msg.Candidate)
	if err != nil {
		return nil, err
	}

	return &types.MsgSubmitApplicationResponse{}, nil
}

func (ms msgServer) Vote(
	ctx context.Context,
	msg *types.MsgVote,
) (*types.MsgVoteResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	voterAddr := sdk.ValAddress(msg.VoterAddr)
	candidateAddr := sdk.ValAddress(msg.CandidateAddr)

	switch msg.VoteType {
	case types.VoteTypeApplication:
		err := ms.keeper.VoteApplication(sdkCtx, voterAddr, candidateAddr, msg.Approve)
		if err != nil {
			return nil, err
		}
	case types.VoteTypeKickProposal:
		err := ms.keeper.VoteKickProposal(sdkCtx, voterAddr, candidateAddr, msg.Approve)
		if err != nil {
			return nil, err
		}
	default:
		return nil, types.ErrInvalidVoteMsg
	}

	return &types.MsgVoteResponse{}, nil
}

func (ms msgServer) ProposeKick(
	ctx context.Context,
	msg *types.MsgProposeKick,
) (*types.MsgProposeKickResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	candidateAddr := sdk.ValAddress(msg.CandidateAddr)
	proposerAddr := sdk.ValAddress(msg.ProposerAddr)

	err := ms.keeper.ProposeKick(sdkCtx, candidateAddr, proposerAddr)
	if err != nil {
		return nil, err
	}

	return &types.MsgProposeKickResponse{}, nil
}

func (ms msgServer) LeaveValidatorSet(
	ctx context.Context,
	msg *types.MsgLeaveValidatorSet,
) (*types.MsgLeaveValidatorSetResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	validatorAddr := sdk.ValAddress(msg.ValidatorAddr)

	err := ms.keeper.LeaveValidatorSet(sdkCtx, validatorAddr)
	if err != nil {
		return nil, err
	}

	return &types.MsgLeaveValidatorSetResponse{}, nil
}


