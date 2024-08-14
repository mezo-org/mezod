package keeper

import (
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/mezo-org/mezod/x/poa/types"
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

	if err := k.checkMaxValidators(ctx); err != nil {
		return err
	}

	// Candidate should not be a validator.
	_, found := k.GetValidator(ctx, validator.GetOperator())
	if found {
		return types.ErrAlreadyValidator
	}
	_, found = k.GetValidatorByConsAddr(ctx, validator.GetConsAddress())
	if found {
		return types.ErrAlreadyValidator
	}

	// Candidate should not be already applying.
	_, found = k.GetApplication(ctx, validator.GetOperator())
	if found {
		return types.ErrAlreadyApplying
	}
	_, found = k.GetApplicationByConsAddr(ctx, validator.GetConsAddress())
	if found {
		return types.ErrAlreadyApplying
	}

	// Create the new application
	k.createApplication(ctx, validator)

	return nil
}

// ApproveApplication approves an application submitted by a validator.
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

	if err := k.checkMaxValidators(ctx); err != nil {
		return err
	}

	// Check the application exists
	application, found := k.GetApplication(ctx, operator)
	if !found {
		return types.ErrNoApplicationFound
	}

	k.removeApplication(ctx, operator)
	k.createValidator(ctx, application.GetValidator())

	return nil
}

// checkMaxValidators checks if the maximum number of validators is reached.
func (k Keeper) checkMaxValidators(ctx sdk.Context) error {
	validatorsCount := uint32(len(k.GetAllValidators(ctx)))
	maxValidators := k.GetParams(ctx).MaxValidators

	if validatorsCount >= maxValidators {
		return types.ErrMaxValidatorsReached
	}

	return nil
}

// GetApplication returns the application by the given candidate operator address.
func (k Keeper) GetApplication(
	ctx sdk.Context,
	operator sdk.ValAddress,
) (types.Application, bool) {
	store := ctx.KVStore(k.storeKey)

	value := store.Get(types.GetApplicationKey(operator))
	if len(value) == 0 {
		return types.Application{}, false
	}

	return types.MustUnmarshalApplication(k.cdc, value), true
}

// GetApplicationByConsAddr gets an application by the given candidate consensus address.
func (k Keeper) GetApplicationByConsAddr(
	ctx sdk.Context,
	cons sdk.ConsAddress,
) (types.Application, bool) {
	store := ctx.KVStore(k.storeKey)

	operator := store.Get(types.GetApplicationByConsAddrKey(cons))
	if len(operator) == 0 {
		return types.Application{}, false
	}

	return k.GetApplication(ctx, operator)
}

// setApplication stores the given application.
func (k Keeper) setApplication(ctx sdk.Context, application types.Application) {
	store := ctx.KVStore(k.storeKey)
	applicationBytes := types.MustMarshalApplication(k.cdc, application)
	store.Set(
		types.GetApplicationKey(application.GetValidator().GetOperator()),
		applicationBytes,
	)
}

// setApplicationByConsAddr indexes the given application by the candidate consensus address.
func (k Keeper) setApplicationByConsAddr(
	ctx sdk.Context,
	application types.Application,
) {
	store := ctx.KVStore(k.storeKey)
	store.Set(
		types.GetApplicationByConsAddrKey(application.GetValidator().GetConsAddress()),
		application.GetValidator().GetOperator(),
	)
}

// createApplication creates a new application for the given candidate validator.
func (k Keeper) createApplication(ctx sdk.Context, validator types.Validator) {
	application := types.NewApplication(validator)
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

	validatorCons := application.GetValidator().GetConsAddress()

	// delete the validator record
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetApplicationKey(operator))
	store.Delete(types.GetApplicationByConsAddrKey(validatorCons))
}

// GetAllApplications gets all applications.
func (k Keeper) GetAllApplications(ctx sdk.Context) (applications []types.Application) {
	store := ctx.KVStore(k.storeKey)

	iterator := storetypes.KVStorePrefixIterator(store, types.ApplicationKeyPrefix)
	defer func() {
		_ = iterator.Close()
	}()

	for ; iterator.Valid(); iterator.Next() {
		application := types.MustUnmarshalApplication(k.cdc, iterator.Value())
		applications = append(applications, application)
	}

	return applications
}
