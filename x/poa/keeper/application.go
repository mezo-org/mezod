package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v12/x/poa/types"
)

// SubmitApplication submits a new application to become a validator.
func (k Keeper) SubmitApplication(
	ctx sdk.Context,
	candidate types.Validator,
) error {
	params := k.GetParams(ctx)

	// Check max validator is not reached
	allValidators := k.GetAllValidators(ctx)
	if uint32(len(allValidators)) == params.MaxValidators {
		return types.ErrMaxValidatorsReached
	}
	// Candidate should not be a validator
	_, found := k.GetValidator(ctx, candidate.GetOperator())
	if found {
		return types.ErrAlreadyValidator
	}
	_, found = k.GetValidatorByConsAddr(ctx, candidate.GetConsAddr())
	if found {
		return types.ErrAlreadyValidator
	}

	// If quorum is 0 the application is immediately approved
	if params.Quorum == 0 {
		// The validator is directly appended in the validator set
		k.appendValidator(ctx, candidate)

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeAppendValidator,
				sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
				sdk.NewAttribute(types.AttributeKeyCandidate, candidate.GetOperator().String()),
			),
		)
	} else {
		// If quorum is more than 0, we create an application vote

		// Candidate should not be already applying
		_, found = k.GetApplication(ctx, candidate.GetOperator())
		if found {
			return types.ErrAlreadyApplying
		}
		_, found = k.GetApplicationByConsAddr(ctx, candidate.GetConsAddr())
		if found {
			return types.ErrAlreadyApplying
		}

		// Create the new application
		k.appendApplication(ctx, candidate)

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeSubmitApplication,
				sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
				sdk.NewAttribute(types.AttributeKeyCandidate, candidate.GetOperator().String()),
			),
		)
	}

	return nil
}

// VoteApplication votes for an application to become a validator.
func (k Keeper) VoteApplication(
	ctx sdk.Context,
	voterAddr sdk.ValAddress,
	candidateAddr sdk.ValAddress,
	approve bool,
) error {
	params := k.GetParams(ctx)

	// Check max validator is not reached. If max validator is reached, no application can be voted
	allValidators := k.GetAllValidators(ctx)
	validatorCount := len(allValidators)
	if uint32(validatorCount) == params.MaxValidators {
		return types.ErrMaxValidatorsReached
	}

	// The voter must be a validator
	_, found := k.GetValidator(ctx, voterAddr)
	if !found {
		return types.ErrVoterNotValidator
	}

	// Check the application exists
	application, found := k.GetApplication(ctx, candidateAddr)
	if !found {
		return types.ErrNoApplicationFound
	}

	// Check if already voted and vote
	alreadyVoted := application.AddVote(voterAddr, approve)
	if alreadyVoted {
		return types.ErrAlreadyVoted
	}

	// Emit the vote event
	if approve {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeApproveApplication,
				sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
				sdk.NewAttribute(types.AttributeKeyVoter, voterAddr.String()),
				sdk.NewAttribute(types.AttributeKeyCandidate, candidateAddr.String()),
			),
		)
	} else {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeRejectApplication,
				sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
				sdk.NewAttribute(types.AttributeKeyVoter, voterAddr.String()),
				sdk.NewAttribute(types.AttributeKeyCandidate, candidateAddr.String()),
			),
		)
	}

	// Check if the quorum has been reached
	reached, approved, err := application.CheckQuorum(uint64(validatorCount), uint64(params.Quorum))
	if err != nil {
		return err
	}

	if reached {
		if approved {
			// Candidate is appended to the validator set
			k.removeApplication(ctx, candidateAddr)
			k.appendValidator(ctx, application.GetSubject())

			// Emit approved event
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeAppendValidator,
					sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
					sdk.NewAttribute(types.AttributeKeyCandidate, candidateAddr.String()),
				),
			)
		} else {
			// Candidate is rejected from joining the validator set
			k.removeApplication(ctx, candidateAddr)

			// Emit rejected event
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeRejectValidator,
					sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
					sdk.NewAttribute(types.AttributeKeyCandidate, candidateAddr.String()),
				),
			)
		}
	} else {
		// Quorum has not been reached yet, update the vote
		k.setApplication(ctx, application)
	}

	return nil
}

// Get an application
func (k Keeper) GetApplication(ctx sdk.Context, addr sdk.ValAddress) (application types.Vote, found bool) {
	store := ctx.KVStore(k.storeKey)

	// Search the value
	value := store.Get(types.GetApplicationKey(addr))
	if value == nil {
		return application, false
	}

	// Return the value
	application = types.MustUnmarshalVote(k.cdc, value)
	return application, true
}

// Get an application by consensus address
func (k Keeper) GetApplicationByConsAddr(ctx sdk.Context, consAddr sdk.ConsAddress) (application types.Vote, found bool) {
	store := ctx.KVStore(k.storeKey)

	opAddr := store.Get(types.GetApplicationByConsAddrKey(consAddr))
	if opAddr == nil {
		return application, false
	}

	return k.GetApplication(ctx, opAddr)
}

// Set application details
func (k Keeper) setApplication(ctx sdk.Context, application types.Vote) {
	store := ctx.KVStore(k.storeKey)
	bz := types.MustMarshalVote(k.cdc, application)
	store.Set(types.GetApplicationKey(application.GetSubject().GetOperator()), bz)
}

// Set application consensus address
func (k Keeper) setApplicationByConsAddr(ctx sdk.Context, application types.Vote) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetApplicationByConsAddrKey(application.GetSubject().GetConsAddr()), application.GetSubject().GetOperator())
}

// Append a new application with a new vote
func (k Keeper) appendApplication(ctx sdk.Context, candidate types.Validator) {
	applicationNewVote := types.NewVote(candidate)
	k.setApplication(ctx, applicationNewVote)
	k.setApplicationByConsAddr(ctx, applicationNewVote)
}

// Remove the application
func (k Keeper) removeApplication(ctx sdk.Context, address sdk.ValAddress) {
	application, found := k.GetApplication(ctx, address)
	if !found {
		return
	}

	consAddr := application.GetSubject().GetConsAddr()

	// delete the validator record
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetApplicationKey(address))
	store.Delete(types.GetApplicationByConsAddrKey(consAddr))
}

// Get the set of all applications
func (k Keeper) GetAllApplications(ctx sdk.Context) (applications []types.Vote) {
	store := ctx.KVStore(k.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, types.ApplicationKeyPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		application := types.MustUnmarshalVote(k.cdc, iterator.Value())
		applications = append(applications, application)
	}

	return applications
}
