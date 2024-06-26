package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v12/x/poa/types"
)

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
func (k Keeper) SetApplication(ctx sdk.Context, application types.Vote) {
	store := ctx.KVStore(k.storeKey)
	bz := types.MustMarshalVote(k.cdc, application)
	store.Set(types.GetApplicationKey(application.GetSubject().GetOperator()), bz)
}

// Set application consensus address
func (k Keeper) SetApplicationByConsAddr(ctx sdk.Context, application types.Vote) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetApplicationByConsAddrKey(application.GetSubject().GetConsAddr()), application.GetSubject().GetOperator())
}

// Append a new application with a new vote
func (k Keeper) AppendApplication(ctx sdk.Context, candidate types.Validator) {
	applicationNewVote := types.NewVote(candidate)
	k.SetApplication(ctx, applicationNewVote)
	k.SetApplicationByConsAddr(ctx, applicationNewVote)
}

// Remove the application
func (k Keeper) RemoveApplication(ctx sdk.Context, address sdk.ValAddress) {
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

// Get the set of all application
func (k Keeper) GetAllApplications(ctx sdk.Context) (applications []types.Vote) {
	store := ctx.KVStore(k.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, types.ApplicationPoolKey)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		application := types.MustUnmarshalVote(k.cdc, iterator.Value())
		applications = append(applications, application)
	}

	return applications
}
