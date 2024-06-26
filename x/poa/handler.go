package poa

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/evmos/evmos/v12/x/poa/keeper"
	"github.com/evmos/evmos/v12/x/poa/types"
)

// NewHandler creates an sdk.Handler for all the poa type messages
func NewHandler(k keeper.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		switch msg := msg.(type) {
		case types.MsgSubmitApplication:
			return handleMsgSubmitApplication(ctx, k, msg)
		case types.MsgVote:
			return handleMsgVote(ctx, k, msg)
		case types.MsgProposeKick:
			return handleMsgProposeKick(ctx, k, msg)
		case types.MsgLeaveValidatorSet:
			return handleMsgLeaveValidatorSet(ctx, k, msg)
		default:
			errMsg := fmt.Sprintf("unrecognized %s message type: %T", types.ModuleName, msg)
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, errMsg)
		}
	}
}

// handleMsgSubmitApplication create a new application to become a validator
func handleMsgSubmitApplication(ctx sdk.Context, k keeper.Keeper, msg types.MsgSubmitApplication) (*sdk.Result, error) {
	// Check max validator is not reached
	allValidators := k.GetAllValidators(ctx)
	maxValidator := k.MaxValidators(ctx)
	if uint16(len(allValidators)) == maxValidator {
		return nil, types.ErrMaxValidatorsReached
	}
	// Candidate should not be a validator
	_, found := k.GetValidator(ctx, msg.Candidate.GetOperator())
	if found {
		return nil, types.ErrAlreadyValidator
	}
	_, found = k.GetValidatorByConsAddr(ctx, msg.Candidate.GetConsAddr())
	if found {
		return nil, types.ErrAlreadyValidator
	}

	// If quorum is 0 the application is immediately approved
	if k.Quorum(ctx) == 0 {
		// The validator is directly appended in the validator set
		k.AppendValidator(ctx, msg.Candidate)

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeAppendValidator,
				sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
				sdk.NewAttribute(types.AttributeKeyCandidate, msg.Candidate.GetOperator().String()),
			),
		)
	} else {
		// If quorum is more than 0, we create a application vote

		// Candidate should not be already applying
		_, found = k.GetApplication(ctx, msg.Candidate.GetOperator())
		if found {
			return nil, types.ErrAlreadyApplying
		}
		_, found = k.GetApplicationByConsAddr(ctx, msg.Candidate.GetConsAddr())
		if found {
			return nil, types.ErrAlreadyApplying
		}

		// Create the new application
		k.AppendApplication(ctx, msg.Candidate)

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeSubmitApplication,
				sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
				sdk.NewAttribute(types.AttributeKeyCandidate, msg.Candidate.GetOperator().String()),
			),
		)
	}

	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}

// handleMsgProposeKick creates a new vote in the kick proposal pools if all conditions are met
func handleMsgProposeKick(ctx sdk.Context, k keeper.Keeper, msg types.MsgProposeKick) (*sdk.Result, error) {
	// The candidate of the kick proposal can't be the proposer
	if msg.ProposerAddr.Equals(msg.CandidateAddr) {
		return nil, types.ErrProposerIsCandidate
	}

	// The proposer must be a validator
	_, found := k.GetValidator(ctx, msg.ProposerAddr)
	if !found {
		return nil, types.ErrProposerNotValidator
	}

	// Candidate should be a validator
	candidate, found := k.GetValidator(ctx, msg.CandidateAddr)
	if !found {
		return nil, types.ErrNotValidator
	}
	// Can't create a kick proposal if the candidate is leaving the validator set
	valState, found := k.GetValidatorState(ctx, msg.CandidateAddr)
	if !found {
		panic("A validator has no state")
	}
	if valState == types.ValidatorStateLeaving {
		return nil, types.ErrValidatorLeaving
	}

	// If quorum is 0 the candidate is immediatelly kicked from the validator set
	if k.Quorum(ctx) == 0 {
		// We set the validator state to leaving, the End Blocker will update the keeper
		k.SetValidatorState(ctx, candidate, types.ValidatorStateLeaving)

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeKickValidator,
				sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
				sdk.NewAttribute(types.AttributeKeyValidator, msg.CandidateAddr.String()),
			),
		)
	} else {
		// If quorum is more than 0, we create a kick proposal vote

		// Candidate should not be already in a kick proposal
		_, found = k.GetKickProposal(ctx, msg.CandidateAddr)
		if found {
			return nil, types.ErrAlreadyInKickProposal
		}

		// Create the new application
		k.AppendKickProposal(ctx, candidate)

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeProposeKick,
				sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
				sdk.NewAttribute(types.AttributeKeyValidator, msg.CandidateAddr.String()),
				sdk.NewAttribute(types.AttributeKeyProposer, msg.ProposerAddr.String()),
			),
		)
	}

	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}

// handleMsgVote handles a vote performed by a validator
func handleMsgVote(ctx sdk.Context, k keeper.Keeper, msg types.MsgVote) (*sdk.Result, error) {
	switch msg.VoteType {
	case types.VoteTypeApplication:
		return handleMsgVoteApplication(ctx, k, msg)
	case types.VoteTypeKickProposal:
		return handleMsgVoteTypeKickProposal(ctx, k, msg)
	default:
		return nil, types.ErrInvalidVoteMsg
	}
}

func handleMsgVoteApplication(ctx sdk.Context, k keeper.Keeper, msg types.MsgVote) (*sdk.Result, error) {
	// Check max validator is not reached. If max validator is reached, not application can be voted
	allValidators := k.GetAllValidators(ctx)
	validatorCount := len(allValidators)
	maxValidator := k.MaxValidators(ctx)
	if uint16(validatorCount) == maxValidator {
		return nil, types.ErrMaxValidatorsReached
	}

	// The voter must be a validator
	_, found := k.GetValidator(ctx, msg.VoterAddr)
	if !found {
		return nil, types.ErrVoterNotValidator
	}

	// Check the application exist
	application, found := k.GetApplication(ctx, msg.CandidateAddr)
	if !found {
		return nil, types.ErrNoApplicationFound
	}

	// Check if already voted and vote
	alreadyVoted := application.AddVote(msg.VoterAddr, msg.Approve)
	if alreadyVoted {
		return nil, types.ErrAlreadyVoted
	}

	// Emit the vote event
	if msg.Approve {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeApproveApplication,
				sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
				sdk.NewAttribute(types.AttributeKeyVoter, msg.VoterAddr.String()),
				sdk.NewAttribute(types.AttributeKeyCandidate, msg.CandidateAddr.String()),
			),
		)
	} else {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeRejectApplication,
				sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
				sdk.NewAttribute(types.AttributeKeyVoter, msg.VoterAddr.String()),
				sdk.NewAttribute(types.AttributeKeyCandidate, msg.CandidateAddr.String()),
			),
		)
	}

	// Check if the quorum has been reached
	reached, approved, err := application.CheckQuorum(uint64(validatorCount), uint64(k.Quorum(ctx)))
	if err != nil {
		return nil, err
	}

	if reached {
		if approved {
			// Candidate is appended to the validator set
			k.RemoveApplication(ctx, msg.CandidateAddr)
			k.AppendValidator(ctx, application.GetSubject())

			// Emit approved event
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeAppendValidator,
					sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
					sdk.NewAttribute(types.AttributeKeyCandidate, msg.CandidateAddr.String()),
				),
			)
		} else {
			// Candidate is rejected from joining the validator set
			k.RemoveApplication(ctx, msg.CandidateAddr)

			// Emit rejected event
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeRejectValidator,
					sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
					sdk.NewAttribute(types.AttributeKeyCandidate, msg.CandidateAddr.String()),
				),
			)
		}
	} else {
		// Quorum has not been reached yet, update the vote
		k.SetApplication(ctx, application)
	}

	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}

func handleMsgVoteTypeKickProposal(ctx sdk.Context, k keeper.Keeper, msg types.MsgVote) (*sdk.Result, error) {
	// The candidate of the kick proposal can't be the voter
	if msg.VoterAddr.Equals(msg.CandidateAddr) {
		return nil, types.ErrVoterIsCandidate
	}

	// The voter must be a validator
	_, found := k.GetValidator(ctx, msg.VoterAddr)
	if !found {
		return nil, types.ErrVoterNotValidator
	}

	// Check the kick proposal exist
	kickProposal, found := k.GetKickProposal(ctx, msg.CandidateAddr)
	if !found {
		return nil, types.ErrNoKickProposalFound
	}

	// Check if already voted and vote
	alreadyVoted := kickProposal.AddVote(msg.VoterAddr, msg.Approve)
	if alreadyVoted {
		return nil, types.ErrAlreadyVoted
	}

	// Emit the vote event
	if msg.Approve {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeApproveKickProposal,
				sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
				sdk.NewAttribute(types.AttributeKeyVoter, msg.VoterAddr.String()),
				sdk.NewAttribute(types.AttributeKeyValidator, msg.CandidateAddr.String()),
			),
		)
	} else {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeRejectKickProposal,
				sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
				sdk.NewAttribute(types.AttributeKeyVoter, msg.VoterAddr.String()),
				sdk.NewAttribute(types.AttributeKeyValidator, msg.CandidateAddr.String()),
			),
		)
	}

	// Get validator count
	allValidators := k.GetAllValidators(ctx)
	validatorCount := len(allValidators)

	// Check if the quorum has been reached
	// We decrement validator count, the candidate of the kick proposal cannot vote
	reached, approved, err := kickProposal.CheckQuorum(uint64(validatorCount)-1, uint64(k.Quorum(ctx)))
	if err != nil {
		return nil, err
	}

	if reached {
		if approved {
			// The validator leave the validator set
			// The state is set to leave, End Blocker will remove definitely the validator
			k.RemoveKickProposal(ctx, msg.CandidateAddr)
			k.SetValidatorState(ctx, kickProposal.GetSubject(), types.ValidatorStateLeaving)

			// Emit approved event
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeKickValidator,
					sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
					sdk.NewAttribute(types.AttributeKeyValidator, msg.CandidateAddr.String()),
				),
			)
		} else {
			// Kick proposal rejected, validator is not removed
			k.RemoveKickProposal(ctx, msg.CandidateAddr)

			// Emit rejected event
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeKeepValidator,
					sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
					sdk.NewAttribute(types.AttributeKeyValidator, msg.CandidateAddr.String()),
				),
			)
		}
	} else {
		// Quorum has not been reached yet, update the vote
		k.SetKickProposal(ctx, kickProposal)
	}

	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}

func handleMsgLeaveValidatorSet(ctx sdk.Context, k keeper.Keeper, msg types.MsgLeaveValidatorSet) (*sdk.Result, error) {
	// Sender must be a validator
	validator, found := k.GetValidator(ctx, msg.ValidatorAddr)
	if !found {
		return nil, types.ErrNotValidator
	}

	// Get validator count
	allValidators := k.GetAllValidators(ctx)
	validatorCount := len(allValidators)
	if validatorCount == 1 {
		return nil, types.ErrOnlyOneValidator
	}

	// If a kick proposal exist for this validator, remove it
	_, found = k.GetKickProposal(ctx, msg.ValidatorAddr)
	if found {
		k.RemoveKickProposal(ctx, msg.ValidatorAddr)
	}

	// Set the state of the validator to leaving, End Blocker will remove the validator from the keeper
	k.SetValidatorState(ctx, validator, types.ValidatorStateLeaving)

	// Emit approved event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeLeaveValidatorSet,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(types.AttributeKeyValidator, msg.ValidatorAddr.String()),
		),
	)

	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}
