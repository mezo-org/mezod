package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v12/x/poa/types"
)

// ProposeKick handles a proposal to kick a validator from the validator set.
func (k Keeper) ProposeKick(
	ctx sdk.Context,
	candidateAddr sdk.ValAddress,
	proposerAddr sdk.ValAddress,
) error {
	params := k.GetParams(ctx)

	// The candidate of the kick proposal can't be the proposer
	if proposerAddr.Equals(candidateAddr) {
		return types.ErrProposerIsCandidate
	}

	// The proposer must be a validator
	_, found := k.GetValidator(ctx, proposerAddr)
	if !found {
		return types.ErrProposerNotValidator
	}

	// Candidate should be a validator
	candidate, found := k.GetValidator(ctx, candidateAddr)
	if !found {
		return types.ErrNotValidator
	}
	// Can't create a kick proposal if the candidate is leaving the validator set
	valState, found := k.GetValidatorState(ctx, candidateAddr)
	if !found {
		panic("A validator has no state")
	}
	if valState == types.ValidatorStateLeaving {
		return types.ErrValidatorLeaving
	}

	// If quorum is 0 the candidate is immediately kicked from the validator set
	if params.Quorum == 0 {
		// We set the validator state to leaving, the End Blocker will update the keeper
		k.setValidatorState(ctx, candidate, types.ValidatorStateLeaving)

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeKickValidator,
				sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
				sdk.NewAttribute(types.AttributeKeyValidator, candidateAddr.String()),
			),
		)
	} else {
		// If quorum is more than 0, we create a kick proposal vote

		// Candidate should not be already in a kick proposal
		_, found = k.GetKickProposal(ctx, candidateAddr)
		if found {
			return types.ErrAlreadyInKickProposal
		}

		// Create the new application
		k.appendKickProposal(ctx, candidate)

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeProposeKick,
				sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
				sdk.NewAttribute(types.AttributeKeyValidator, candidateAddr.String()),
				sdk.NewAttribute(types.AttributeKeyProposer, proposerAddr.String()),
			),
		)
	}

	return nil
}

// VoteKickProposal handles a vote on a kick proposal.
func (k Keeper) VoteKickProposal(
	ctx sdk.Context,
	voterAddr sdk.ValAddress,
	candidateAddr sdk.ValAddress,
	approve bool,
) error {
	params := k.GetParams(ctx)

	// The candidate of the kick proposal can't be the voter
	if voterAddr.Equals(candidateAddr) {
		return types.ErrVoterIsCandidate
	}

	// The voter must be a validator
	_, found := k.GetValidator(ctx, voterAddr)
	if !found {
		return types.ErrVoterNotValidator
	}

	// Check the kick proposal exist
	kickProposal, found := k.GetKickProposal(ctx, candidateAddr)
	if !found {
		return types.ErrNoKickProposalFound
	}

	// Check if already voted and vote
	alreadyVoted := kickProposal.AddVote(voterAddr, approve)
	if alreadyVoted {
		return types.ErrAlreadyVoted
	}

	// Emit the vote event
	if approve {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeApproveKickProposal,
				sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
				sdk.NewAttribute(types.AttributeKeyVoter, voterAddr.String()),
				sdk.NewAttribute(types.AttributeKeyValidator, candidateAddr.String()),
			),
		)
	} else {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeRejectKickProposal,
				sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
				sdk.NewAttribute(types.AttributeKeyVoter, voterAddr.String()),
				sdk.NewAttribute(types.AttributeKeyValidator, candidateAddr.String()),
			),
		)
	}

	// Get validator count
	allValidators := k.GetAllValidators(ctx)
	validatorCount := len(allValidators)

	// Check if the quorum has been reached
	// We decrement validator count, the candidate of the kick proposal cannot vote
	reached, approved, err := kickProposal.CheckQuorum(uint64(validatorCount)-1, uint64(params.Quorum))
	if err != nil {
		return err
	}

	if reached {
		if approved {
			// The validator leave the validator set
			// The state is set to leave, End Blocker will remove definitely the validator
			k.removeKickProposal(ctx, candidateAddr)
			k.setValidatorState(ctx, kickProposal.GetSubject(), types.ValidatorStateLeaving)

			// Emit approved event
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeKickValidator,
					sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
					sdk.NewAttribute(types.AttributeKeyValidator, candidateAddr.String()),
				),
			)
		} else {
			// Kick proposal rejected, validator is not removed
			k.removeKickProposal(ctx, candidateAddr)

			// Emit rejected event
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeKeepValidator,
					sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
					sdk.NewAttribute(types.AttributeKeyValidator, candidateAddr.String()),
				),
			)
		}
	} else {
		// Quorum has not been reached yet, update the vote
		k.setKickProposal(ctx, kickProposal)
	}

	return nil
}

// Get a kick proposal
func (k Keeper) GetKickProposal(ctx sdk.Context, addr sdk.ValAddress) (kickProposal types.Vote, found bool) {
	store := ctx.KVStore(k.storeKey)

	// Search the value
	value := store.Get(types.GetKickProposalKey(addr))
	if value == nil {
		return kickProposal, false
	}

	// Return the value
	kickProposal = types.MustUnmarshalVote(k.cdc, value)
	return kickProposal, true
}

// Set kick proposal details
func (k Keeper) setKickProposal(ctx sdk.Context, kickProposal types.Vote) {
	store := ctx.KVStore(k.storeKey)
	bz := types.MustMarshalVote(k.cdc, kickProposal)
	store.Set(types.GetKickProposalKey(kickProposal.GetSubject().GetOperator()), bz)
}

// Append a new kick proposal with a new vote
func (k Keeper) appendKickProposal(ctx sdk.Context, candidate types.Validator) {
	kickProposalNewVote := types.NewVote(candidate)
	k.setKickProposal(ctx, kickProposalNewVote)
}

// Remove the kick proposal
func (k Keeper) removeKickProposal(ctx sdk.Context, address sdk.ValAddress) {
	_, found := k.GetKickProposal(ctx, address)
	if !found {
		return
	}

	// Delete the kick proposal record
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetKickProposalKey(address))
}

// Get the set of all kick proposals
func (k Keeper) GetAllKickProposals(ctx sdk.Context) (kickProposals []types.Vote) {
	store := ctx.KVStore(k.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, types.KickProposalPoolKey)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		kickProposal := types.MustUnmarshalVote(k.cdc, iterator.Value())
		kickProposals = append(kickProposals, kickProposal)
	}

	return kickProposals
}
