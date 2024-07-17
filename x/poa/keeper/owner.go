package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v12/x/poa/types"
)

// GetOwner returns the validator pool owner address.
func (k Keeper) GetOwner(ctx sdk.Context) sdk.AccAddress {
	return ctx.KVStore(k.storeKey).Get(types.OwnerKey)
}

// GetCandidateOwner returns the candidate validator pool owner address.
func (k Keeper) GetCandidateOwner(ctx sdk.Context) sdk.AccAddress {
	return ctx.KVStore(k.storeKey).Get(types.CandidateOwnerKey)
}

// TransferOwnership initializes the 2-step validator pool ownership transfer
// process. The new owner is set as a candidate owner and must accept the
// ownership to be promoted to the actual owner.
//
// The function returns an error if the sender is not the current owner.
// Returns nil if the ownership transfer is initialized successfully.
//
// Upstream is responsible for setting the `sender` parameter to the actual
// actor performing the operation. If the sender address is empty, the function
// will return an error.
func (k Keeper) TransferOwnership(
	ctx sdk.Context,
	sender sdk.AccAddress,
	newOwner sdk.AccAddress,
) error {
	if err := k.checkOwner(ctx, sender); err != nil {
		return err
	}

	k.setCandidateOwner(ctx, newOwner)

	return nil
}

// AcceptOwnership finalizes the 2-step validator pool ownership transfer process.
// The candidate owner is promoted to the actual owner.
//
// The function returns an error if the sender is not the current candidate owner
// or the ownership transfer process is not initialized. Returns nil if the
// ownership transfer is finalized successfully.
//
// Upstream is responsible for setting the `sender` parameter to the actual
// actor performing the operation. If the sender address is empty, the function
// will return an error.
func (k Keeper) AcceptOwnership(
	ctx sdk.Context,
	sender sdk.AccAddress,
) error {
	if err := k.checkCandidateOwner(ctx, sender); err != nil {
		return err
	}

	k.setOwner(ctx, k.GetCandidateOwner(ctx))
	k.deleteCandidateOwner(ctx)

	return nil
}

// setCandidateOwner sets the candidate validator pool owner address.
func (k Keeper) setCandidateOwner(
	ctx sdk.Context,
	candidateOwner sdk.AccAddress,
) {
	ctx.KVStore(k.storeKey).Set(types.CandidateOwnerKey, candidateOwner)
}

// deleteCandidateOwner deletes the candidate validator pool owner address.
func (k Keeper) deleteCandidateOwner(
	ctx sdk.Context,
) {
	ctx.KVStore(k.storeKey).Delete(types.CandidateOwnerKey)
}

// setOwner sets the validator pool owner address.
func (k Keeper) setOwner(ctx sdk.Context, owner sdk.AccAddress) {
	ctx.KVStore(k.storeKey).Set(types.OwnerKey, owner)
}
