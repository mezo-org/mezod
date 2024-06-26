package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	ErrNoValidatorFound      = sdkerrors.Register(ModuleName, 1, "no validator found")
	ErrInvalidValidator      = sdkerrors.Register(ModuleName, 2, "invalid validator")
	ErrInvalidQuorumValue    = sdkerrors.Register(ModuleName, 3, "quorum should be a percentage")
	ErrInvalidVoterPoolSize  = sdkerrors.Register(ModuleName, 4, "the voter pool size is incorrect")
	ErrAlreadyApplying       = sdkerrors.Register(ModuleName, 5, "the candidate is already applying to become a validator")
	ErrAlreadyValidator      = sdkerrors.Register(ModuleName, 6, "the candidate is already a validator")
	ErrMaxValidatorsReached  = sdkerrors.Register(ModuleName, 7, "the maximum number of validators has been reached")
	ErrInvalidVoteMsg        = sdkerrors.Register(ModuleName, 8, "the vote message is invalid")
	ErrVoterNotValidator     = sdkerrors.Register(ModuleName, 9, "the voter is not a validator")
	ErrNoApplicationFound    = sdkerrors.Register(ModuleName, 10, "no application found")
	ErrAlreadyVoted          = sdkerrors.Register(ModuleName, 11, "the validator already voted")
	ErrInvalidKickProposal   = sdkerrors.Register(ModuleName, 12, "the kick proposal is invalid")
	ErrNotValidator          = sdkerrors.Register(ModuleName, 13, "the candidate is not a validator")
	ErrValidatorLeaving      = sdkerrors.Register(ModuleName, 14, "the candidate is leaving the validator set")
	ErrProposerNotValidator  = sdkerrors.Register(ModuleName, 15, "the proposer is not a validator")
	ErrAlreadyInKickProposal = sdkerrors.Register(ModuleName, 16, "the candidate is already in a kick proposal")
	ErrNoKickProposalFound   = sdkerrors.Register(ModuleName, 17, "no kick proposal found")
	ErrVoterIsCandidate      = sdkerrors.Register(ModuleName, 18, "the voter cannot be the candidate")
	ErrProposerIsCandidate   = sdkerrors.Register(ModuleName, 19, "the proposer cannot be the candidate")
	ErrOnlyOneValidator      = sdkerrors.Register(ModuleName, 20, "there is only one validator in the validator set")
)
