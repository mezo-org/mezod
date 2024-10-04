package keeper

import (
	"bytes"
	"context"
	"slices"

	errorsmod "cosmossdk.io/errors"
	storetypes "cosmossdk.io/store/types"
	cmtprotocrypto "github.com/cometbft/cometbft/proto/tendermint/crypto"
	cryptocdc "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/mezo-org/mezod/x/poa/types"
)

// Kick forcibly removes a validator from the validator pool.
// The validator will be removed from active validators at the end of the block.
//
// The function returns an error if:
// - the sender is not the owner,
// - the validator does not exist,
// - the validator is not an active validator.
// Returns nil if the validator is successfully kicked.
//
// Upstream is responsible for setting the `sender` parameter to the actual
// actor performing the operation. If the sender address is empty, the function
// will return an error.
func (k Keeper) Kick(
	ctx sdk.Context,
	sender sdk.AccAddress,
	operator sdk.ValAddress,
) error {
	if err := k.checkOwner(ctx, sender); err != nil {
		return err
	}

	return k.setValidatorStateLeaving(ctx, operator)
}

// Leave voluntarily removes a validator from the validator pool.
// The validator will be removed from active validators at the end of the block.
//
// The function returns an error if:
// - there is only one validator,
// - the sender is not an existing validator,
// - the validator is not an active validator.
// Returns nil if the validator successfully leaves.
//
// Upstream is responsible for setting the `sender` parameter to the actual
// actor performing the operation. If the sender address is empty, the function
// will return an error.
func (k Keeper) Leave(ctx sdk.Context, sender sdk.AccAddress) error {
	// Block voluntary leaving if there is only one active validator.
	if len(k.GetActiveValidators(ctx)) == 1 {
		return types.ErrOnlyOneValidator
	}

	operator := sdk.ValAddress(sender)

	return k.setValidatorStateLeaving(ctx, operator)
}

// setValidatorStateLeaving sets the validator state to types.ValidatorStateLeaving.
// The validator will be removed from active validators at the end of the block.
//
// The function returns an error if:
// - the validator does not exist,
// - the validator is already leaving.
// Returns nil if the validator state is successfully set to types.ValidatorStateLeaving.
func (k Keeper) setValidatorStateLeaving(
	ctx sdk.Context,
	operator sdk.ValAddress,
) error {
	// Validator must be known.
	validator, found := k.GetValidator(ctx, operator)
	if !found {
		return types.ErrNotValidator
	}

	// Check if the validator is not already leaving.
	validatorState, found := k.GetValidatorState(ctx, operator)
	if !found {
		// This should never happen. All validators should have a state.
		panic("Validator state is unknown")
	}
	// Only an active validator can leave.
	if validatorState != types.ValidatorStateActive {
		return errorsmod.Wrap(
			types.ErrWrongValidatorState,
			"not an active validator",
		)
	}

	// Set the validator state to Leaving. Validator removal will be
	// finalized at the end of the block (see EndBlocker method).
	k.setValidatorState(ctx, validator, types.ValidatorStateLeaving)

	return nil
}

// GetValidator gets a validator by the operator address.
func (k Keeper) GetValidator(
	ctx sdk.Context,
	operator sdk.ValAddress,
) (types.Validator, bool) {
	store := ctx.KVStore(k.storeKey)

	value := store.Get(types.GetValidatorKey(operator))
	if len(value) == 0 {
		return types.Validator{}, false
	}

	return types.MustUnmarshalValidator(k.cdc, value), true
}

// GetValidatorByConsAddr gets a validator by the consensus address.
func (k Keeper) GetValidatorByConsAddr(
	ctx sdk.Context,
	cons sdk.ConsAddress,
) (types.Validator, bool) {
	store := ctx.KVStore(k.storeKey)

	operator := store.Get(types.GetValidatorByConsAddrKey(cons))
	if len(operator) == 0 {
		return types.Validator{}, false
	}

	return k.GetValidator(ctx, operator)
}

// GetValidatorState gets the state of a validator.
func (k Keeper) GetValidatorState(
	ctx sdk.Context,
	operator sdk.ValAddress,
) (types.ValidatorState, bool) {
	store := ctx.KVStore(k.storeKey)

	value := store.Get(types.GetValidatorStateKey(operator))
	if len(value) == 0 {
		return types.ValidatorStateUnknown, false
	}

	// A single byte represents the state.
	return types.ValidatorState(value[0]), true
}

// setValidator stores the given validator.
func (k Keeper) setValidator(ctx sdk.Context, validator types.Validator) {
	store := ctx.KVStore(k.storeKey)
	validatorBytes := types.MustMarshalValidator(k.cdc, validator)
	store.Set(types.GetValidatorKey(validator.GetOperator()), validatorBytes)
}

// setValidatorByConsAddr indexes the given validator by the consensus address.
func (k Keeper) setValidatorByConsAddr(ctx sdk.Context, validator types.Validator) {
	store := ctx.KVStore(k.storeKey)
	store.Set(
		types.GetValidatorByConsAddrKey(validator.GetConsAddress()),
		validator.GetOperator(),
	)
}

// setValidatorState sets the state of a validator.
func (k Keeper) setValidatorState(
	ctx sdk.Context,
	validator types.Validator,
	state types.ValidatorState,
) {
	if state != types.ValidatorStateJoining &&
		state != types.ValidatorStateActive &&
		state != types.ValidatorStateLeaving {
		panic("Incorrect validator state")
	}

	store := ctx.KVStore(k.storeKey)

	// The state can be encoded in a single byte.
	stateBytes := []byte{uint8(state)}
	store.Set(types.GetValidatorStateKey(validator.GetOperator()), stateBytes)
}

// createValidator appends a new validator to the validator pool with the state
// types.ValidatorStateJoining.
func (k Keeper) createValidator(ctx sdk.Context, validator types.Validator) {
	k.setValidator(ctx, validator)
	k.setValidatorByConsAddr(ctx, validator)
	k.setValidatorState(ctx, validator, types.ValidatorStateJoining)
}

// removeValidator removes a validator from the validator pool.
//
// WARNING: This function should only be called by the end blocker to ensure
// the validator is removed from the Tendermint validator state. This function
// is called by the end blocker when the validator state is leaving
func (k Keeper) removeValidator(ctx sdk.Context, operator sdk.ValAddress) {
	validator, found := k.GetValidator(ctx, operator)
	if !found {
		return
	}

	cons := validator.GetConsAddress()

	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetValidatorKey(operator))
	store.Delete(types.GetValidatorByConsAddrKey(cons))
	store.Delete(types.GetValidatorStateKey(operator))
}

// GetAllValidators gets the set of all validators registered in the module store.
// The result contains validators of all states:
// - types.ValidatorStateJoining: not yet present in the Tendermint validator set
// - types.ValidatorStateActive: already present in the Tendermint validator set
// - types.ValidatorStateLeaving: will leave the Tendermint validator set at the end of the block
func (k Keeper) GetAllValidators(ctx sdk.Context) (validators []types.Validator) {
	store := ctx.KVStore(k.storeKey)

	iterator := storetypes.KVStorePrefixIterator(store, types.ValidatorKeyPrefix)
	defer func() {
		_ = iterator.Close()
	}()

	for ; iterator.Valid(); iterator.Next() {
		validator := types.MustUnmarshalValidator(k.cdc, iterator.Value())
		validators = append(validators, validator)
	}

	return validators
}

// GetActiveValidators gets the set of all active validators that are part
// of the Tendermint consensus set. The result contains only validators with
// the state types.ValidatorStateActive.
func (k Keeper) GetActiveValidators(ctx sdk.Context) (validators []types.Validator) {
	for _, validator := range k.GetAllValidators(ctx) {
		state, found := k.GetValidatorState(ctx, validator.GetOperator())
		// Panic on no state.
		if !found {
			panic("Found a validator with no state, a validator should always have a state")
		}

		// Consider only validators with Active state. Ignore Joining and Leaving
		// validators. The former will join the Tendermint consensus set and the
		// latter will leave the Tendermint consensus set at the end of the block.
		if state == types.ValidatorStateActive {
			validators = append(validators, validator)
		}
	}

	return validators
}

// GetValidatorsConsAddrsByPrivilege returns the consensus addresses of
// all validators that are currently present in the store and have the
// given privilege. There is no guarantee that the returned validators
// are currently part of the CometBFT validator set.
//
// TODO: Temporary implementation that assumes the first half of the validators
// have the requested privilege. Change this function once the actual privilege
// system is implemented. Cover with unit tests once that happens.
func (k Keeper) GetValidatorsConsAddrsByPrivilege(
	ctx sdk.Context,
	_ string,
) []sdk.ConsAddress {
	validators := k.GetAllValidators(ctx)

	// Sort to ensure determinism.
	slices.SortFunc(validators, func(i, j types.Validator) int {
		return bytes.Compare(i.GetOperator().Bytes(), j.GetOperator().Bytes())
	})

	mid := len(validators) / 2

	consAddresses := make([]sdk.ConsAddress, 0)
	for _, validator := range validators[:mid] {
		consAddresses = append(consAddresses, validator.GetConsAddress())
	}

	return consAddresses
}

// GetPubKeyByConsAddr gets the public key of a validator by the consensus address.
func (k Keeper) GetPubKeyByConsAddr(
	ctx context.Context,
	cons sdk.ConsAddress,
) (cmtprotocrypto.PublicKey, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	validator, ok := k.GetValidatorByConsAddr(sdkCtx, cons)
	if !ok {
		return cmtprotocrypto.PublicKey{}, types.ErrNoValidatorFound
	}

	protoPubKey, err := cryptocdc.ToCmtProtoPublicKey(validator.GetConsPubKey())
	if err != nil {
		return cmtprotocrypto.PublicKey{}, err
	}

	return protoPubKey, nil
}
