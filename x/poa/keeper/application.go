package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v12/x/poa/types"
)

// SubmitApplication submits a new application to become a validator.
//
// The function returns an error if:
// - the sender is not the operator of the candidate validator,
// - the max validators limit is reached,
// - the candidate is already a validator,
// - the candidate is already applying.
// Returns nil if the application is successfully submitted.
//
// Upstream is responsible for setting the `sender` parameter to the actual
// actor performing the operation. If the sender address is empty, the function
// will return an error.
func (k Keeper) SubmitApplication(
	ctx sdk.Context,
	sender sdk.AccAddress,
	validator types.Validator,
) error {
	if err := k.checkValidatorOperator(sender, validator); err != nil {
		return err
	}

	params := k.GetParams(ctx)

	// Check max validator is not reached.
	allValidators := k.GetAllValidators(ctx)
	if uint32(len(allValidators)) == params.MaxValidators {
		return types.ErrMaxValidatorsReached
	}

	// Candidate should not be a validator.
	_, found := k.GetValidator(ctx, validator.GetOperator())
	if found {
		return types.ErrAlreadyValidator
	}
	_, found = k.GetValidatorByConsAddr(ctx, validator.GetConsAddr())
	if found {
		return types.ErrAlreadyValidator
	}

	// Candidate should not be already applying.
	_, found = k.GetApplication(ctx, validator.GetOperator())
	if found {
		return types.ErrAlreadyApplying
	}
	_, found = k.GetApplicationByConsAddr(ctx, validator.GetConsAddr())
	if found {
		return types.ErrAlreadyApplying
	}

	// Create the new application
	k.appendApplication(ctx, validator)

	return nil
}

// ApproveApplication approves an application to become a validator.
// The candidate validator will become an active validator at the
// end of the block.
//
// The function returns an error if:
// - the sender is not the owner,
// - the max validators limit is reached,
// - the application does not exist.
// Returns nil if the application is successfully approved.
//
// Upstream is responsible for setting the `sender` parameter to the actual
// actor performing the operation. If the sender address is empty, the function
// will return an error.
func (k Keeper) ApproveApplication(
	ctx sdk.Context,
	sender sdk.AccAddress,
	operator sdk.ValAddress,
) error {
	if err := k.checkOwner(ctx, sender); err != nil {
		return err
	}

	validatorsCount := uint32(len(k.GetAllValidators(ctx)))
	maxValidators := k.GetParams(ctx).MaxValidators

	// Check max validator is not reached. If max validator is reached,
	// no more validators can be added.
	if validatorsCount >= maxValidators {
		return types.ErrMaxValidatorsReached
	}

	// Check the application exists
	application, found := k.GetApplication(ctx, operator)
	if !found {
		return types.ErrNoApplicationFound
	}

	k.removeApplication(ctx, operator)
	k.appendValidator(ctx, application.GetSubject())

	return nil
}

// GetApplication returns the application by the given candidate operator address.
func (k Keeper) GetApplication(
	ctx sdk.Context,
	operator sdk.ValAddress,
) (types.Vote, bool) {
	store := ctx.KVStore(k.storeKey)

	value := store.Get(types.GetApplicationKey(operator))
	if len(value) == 0 {
		return types.Vote{}, false
	}

	return types.MustUnmarshalVote(k.cdc, value), true
}

// GetApplicationByConsAddr gets an application by the given candidate consensus address.
func (k Keeper) GetApplicationByConsAddr(
	ctx sdk.Context,
	cons sdk.ConsAddress,
) (types.Vote, bool) {
	store := ctx.KVStore(k.storeKey)

	operator := store.Get(types.GetApplicationByConsAddrKey(cons))
	if len(operator) == 0 {
		return types.Vote{}, false
	}

	return k.GetApplication(ctx, operator)
}

// setApplication stores the given application.
func (k Keeper) setApplication(ctx sdk.Context, application types.Vote) {
	store := ctx.KVStore(k.storeKey)
	applicationBytes := types.MustMarshalVote(k.cdc, application)
	store.Set(
		types.GetApplicationKey(application.GetSubject().GetOperator()),
		applicationBytes,
	)
}

// setApplicationByConsAddr indexes the given application by the candidate consensus address.
func (k Keeper) setApplicationByConsAddr(
	ctx sdk.Context,
	application types.Vote,
) {
	store := ctx.KVStore(k.storeKey)
	store.Set(
		types.GetApplicationByConsAddrKey(application.GetSubject().GetConsAddr()),
		application.GetSubject().GetOperator(),
	)
}

// appendApplication appends a new application for the given candidate validator.
func (k Keeper) appendApplication(ctx sdk.Context, validator types.Validator) {
	application := types.NewVote(validator)
	k.setApplication(ctx, application)
	k.setApplicationByConsAddr(ctx, application)
}

// removeApplication removes the application by the given candidate operator address.
func (k Keeper) removeApplication(
	ctx sdk.Context,
	operator sdk.ValAddress,
) {
	application, found := k.GetApplication(ctx, operator)
	if !found {
		return
	}

	candidateCons := application.GetSubject().GetConsAddr()

	// delete the validator record
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetApplicationKey(operator))
	store.Delete(types.GetApplicationByConsAddrKey(candidateCons))
}

// GetAllApplications gets all applications.
func (k Keeper) GetAllApplications(ctx sdk.Context) (applications []types.Vote) {
	store := ctx.KVStore(k.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, types.ApplicationKeyPrefix)
	defer func() {
		_ = iterator.Close()
	}()

	for ; iterator.Valid(); iterator.Next() {
		application := types.MustUnmarshalVote(k.cdc, iterator.Value())
		applications = append(applications, application)
	}

	return applications
}
