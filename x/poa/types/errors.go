package types

import (
	errorsmod "cosmossdk.io/errors"
)

var (
	ErrNoValidatorFound      = errorsmod.Register(ModuleName, 1, "no validator found")
	ErrInvalidValidator      = errorsmod.Register(ModuleName, 2, "invalid validator")
	ErrInvalidQuorumValue    = errorsmod.Register(ModuleName, 3, "quorum should be a percentage")
	ErrInvalidVoterPoolSize  = errorsmod.Register(ModuleName, 4, "the voter pool size is incorrect")
	ErrAlreadyApplying       = errorsmod.Register(ModuleName, 5, "the candidate is already applying to become a validator")
	ErrAlreadyValidator      = errorsmod.Register(ModuleName, 6, "the candidate is already a validator")
	ErrMaxValidatorsReached  = errorsmod.Register(ModuleName, 7, "the maximum number of validators has been reached")
	ErrInvalidVoteMsg        = errorsmod.Register(ModuleName, 8, "the vote message is invalid")
	ErrVoterNotValidator     = errorsmod.Register(ModuleName, 9, "the voter is not a validator")
	ErrNoApplicationFound    = errorsmod.Register(ModuleName, 10, "no application found")
	ErrAlreadyVoted          = errorsmod.Register(ModuleName, 11, "the validator already voted")
	ErrInvalidKickProposal   = errorsmod.Register(ModuleName, 12, "the kick proposal is invalid")
	ErrNotValidator          = errorsmod.Register(ModuleName, 13, "the candidate is not a validator")
	ErrValidatorLeaving      = errorsmod.Register(ModuleName, 14, "the candidate is leaving the validator set")
	ErrProposerNotValidator  = errorsmod.Register(ModuleName, 15, "the proposer is not a validator")
	ErrAlreadyInKickProposal = errorsmod.Register(ModuleName, 16, "the candidate is already in a kick proposal")
	ErrNoKickProposalFound   = errorsmod.Register(ModuleName, 17, "no kick proposal found")
	ErrVoterIsCandidate      = errorsmod.Register(ModuleName, 18, "the voter cannot be the candidate")
	ErrProposerIsCandidate   = errorsmod.Register(ModuleName, 19, "the proposer cannot be the candidate")
	ErrOnlyOneValidator      = errorsmod.Register(ModuleName, 20, "there is only one validator in the validator set")
	ErrInvalidSigner         = errorsmod.Register(ModuleName, 21, "invalid signer")
)
