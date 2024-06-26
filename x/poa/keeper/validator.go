package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v12/x/poa/types"
)

// Get a validator
func (k Keeper) GetValidator(ctx sdk.Context, addr sdk.ValAddress) (validator types.Validator, found bool) {
	store := ctx.KVStore(k.storeKey)

	// Search the value
	value := store.Get(types.GetValidatorKey(addr))
	if value == nil {
		return validator, false
	}

	// Return the value
	validator = types.MustUnmarshalValidator(k.cdc, value)
	return validator, true
}

// Get a validator by consensus address
func (k Keeper) GetValidatorByConsAddr(ctx sdk.Context, consAddr sdk.ConsAddress) (validator types.Validator, found bool) {
	store := ctx.KVStore(k.storeKey)

	opAddr := store.Get(types.GetValidatorByConsAddrKey(consAddr))
	if opAddr == nil {
		return validator, false
	}

	return k.GetValidator(ctx, opAddr)
}

// Get a validator state
func (k Keeper) GetValidatorState(ctx sdk.Context, addr sdk.ValAddress) (state uint16, found bool) {
	store := ctx.KVStore(k.storeKey)

	// Search the value
	value := store.Get(types.GetValidatorStateKey(addr))
	if value == nil {
		return state, false
	}

	// Return the value
	state = uint16(value[0]) // A single byte represents the state
	return state, true
}

// Set validator details
func (k Keeper) SetValidator(ctx sdk.Context, validator types.Validator) {
	store := ctx.KVStore(k.storeKey)
	bz := types.MustMarshalValidator(k.cdc, validator)
	store.Set(types.GetValidatorKey(validator.OperatorAddress), bz)
}

// Set validator consensus address
func (k Keeper) SetValidatorByConsAddr(ctx sdk.Context, validator types.Validator) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetValidatorByConsAddrKey(validator.GetConsAddr()), validator.OperatorAddress)
}

// Set validator state
func (k Keeper) SetValidatorState(ctx sdk.Context, validator types.Validator, state uint16) {
	if state != types.ValidatorStateJoining && state != types.ValidatorStateJoined && state != types.ValidatorStateLeaving {
		panic("Incorrect validator state")
	}

	store := ctx.KVStore(k.storeKey)
	bz := []byte{byte(state)} // The state can be encoded in a single byte
	store.Set(types.GetValidatorStateKey(validator.OperatorAddress), bz)
}

// Append a validator and set its state to joining
func (k Keeper) AppendValidator(ctx sdk.Context, validator types.Validator) {
	k.SetValidator(ctx, validator)
	k.SetValidatorByConsAddr(ctx, validator)
	k.SetValidatorState(ctx, validator, types.ValidatorStateJoining)
}

// Remove the validator
// !!! This function should only be called by the end blocker to ensure the validator is removed from the Tendermint validator state
// !!! This function is called by the end blocker when the validator state is leaving
func (k Keeper) RemoveValidator(ctx sdk.Context, address sdk.ValAddress) {
	validator, found := k.GetValidator(ctx, address)
	if !found {
		return
	}

	consAddr := validator.GetConsAddr()

	// delete the validator record
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetValidatorKey(address))
	store.Delete(types.GetValidatorByConsAddrKey(consAddr))
	store.Delete(types.GetValidatorStateKey(address))
}

// Get the set of all validators
func (k Keeper) GetAllValidators(ctx sdk.Context) (validators []types.Validator) {
	store := ctx.KVStore(k.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, types.ValidatorsKey)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		validator := types.MustUnmarshalValidator(k.cdc, iterator.Value())
		validators = append(validators, validator)
	}

	return validators
}
