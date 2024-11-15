package keeper

import (
	"bytes"
	"reflect"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"

	cryptocdc "github.com/cosmos/cosmos-sdk/crypto/codec"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/google/go-cmp/cmp"
	"github.com/mezo-org/mezod/x/poa/types"
)

func TestKick(t *testing.T) {
	ctx, poaKeeper := mockContext()

	// Generate an owner address using the mockValidator function.
	helper, _ := mockValidator()
	owner := sdk.AccAddress(helper.GetOperator())
	poaKeeper.setOwner(ctx, owner)

	validator1, _ := mockValidator()
	validator2, _ := mockValidator()

	err := poaKeeper.setParams(ctx, types.DefaultParams())
	if err != nil {
		t.Fatal(err)
	}

	// Append validator 1 as ValidatorStateJoining.
	poaKeeper.createValidator(ctx, validator1)

	// The owner cannot kick validator 1 which is still joining and not active yet.
	err = poaKeeper.Kick(ctx, owner, validator1.GetOperator())
	expectedErr := errorsmod.Wrap(
		types.ErrWrongValidatorState,
		"not an active validator",
	)
	if err.Error() != expectedErr.Error() {
		t.Errorf(
			"Kick when still joining, error should be %v, got %v",
			expectedErr.Error(),
			err.Error(),
		)
	}

	// Run validator updates so the validator 1 state switch to ValidatorStateActive.
	// There is 1 active validator now.
	_, err = poaKeeper.EndBlocker(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Try to impersonate the owner and kick validator 1.
	err = poaKeeper.Kick(ctx,
		sdk.AccAddress(validator2.GetOperator()),
		validator1.GetOperator(),
	)
	expectedErr = errorsmod.Wrapf(
		sdkerrors.ErrUnauthorized,
		"not the owner; expected %s, sender %s",
		owner.String(),
		sdk.AccAddress(validator2.GetOperator()).String(),
	)
	if err.Error() != expectedErr.Error() {
		t.Errorf(
			"Kick with wrong sender, error should be %v, got %v",
			expectedErr.Error(),
			err.Error(),
		)
	}

	// The owner can kick validator 1 as it's an active validator.
	// Note that the validator can be kicked even if this is the last
	// validator.
	err = poaKeeper.Kick(ctx, owner, validator1.GetOperator())
	if err != nil {
		t.Errorf("Kick should work, got error %v", err)
	}
	validatorState, found := poaKeeper.GetValidatorState(ctx, validator1.GetOperator())
	if !found {
		t.Errorf("Kick should not directly remove the validator")
	}
	if validatorState != types.ValidatorStateLeaving {
		t.Errorf("Kick should set the state of the validator to leaving")
	}

	// The owner cannot kick validator 1 which was already kicked and is leaving.
	err = poaKeeper.Kick(ctx, owner, validator1.GetOperator())
	expectedErr = errorsmod.Wrap(
		types.ErrWrongValidatorState,
		"not an active validator",
	)
	if err.Error() != expectedErr.Error() {
		t.Errorf(
			"Kick when already leaving, error should be %v, got %v",
			expectedErr.Error(),
			err.Error(),
		)
	}
}

func TestLeave(t *testing.T) {
	ctx, poaKeeper := mockContext()
	validator1, _ := mockValidator()
	validator2, _ := mockValidator()
	validator3, _ := mockValidator()

	err := poaKeeper.setParams(ctx, types.DefaultParams())
	if err != nil {
		t.Fatal(err)
	}

	// Append validator 1 as ValidatorStateJoining.
	poaKeeper.createValidator(ctx, validator1)
	// Run validator updates so the validator 1 state switch to ValidatorStateActive.
	_, err = poaKeeper.EndBlocker(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Validator 1 cannot leave the set as it's the only active validator.
	err = poaKeeper.Leave(ctx, sdk.AccAddress(validator1.GetOperator()))
	if err.Error() != types.ErrOnlyOneValidator.Error() {
		t.Errorf(
			"Leave with one validator, error should be %v, got %v",
			types.ErrOnlyOneValidator.Error(),
			err.Error(),
		)
	}

	// Add a validator 2 as ValidatorStateJoining.
	poaKeeper.createValidator(ctx, validator2)
	// Run validator updates so the validator 2 state switch to ValidatorStateActive.
	// There are 2 active validators now.
	_, err = poaKeeper.EndBlocker(ctx)
	if err != nil {
		if err != nil {
			t.Fatal(err)
		}
	}

	// Validator 3 is not a validator yet so, it cannot leave the validator set.
	err = poaKeeper.Leave(ctx, sdk.AccAddress(validator3.GetOperator()))
	if err.Error() != types.ErrNotValidator.Error() {
		t.Errorf(
			"Leave when not validator, error should be %v, got %v",
			types.ErrNotValidator.Error(),
			err.Error(),
		)
	}

	// Add a validator 3 as ValidatorStateJoining but don't run the EndBlocker.
	// There are still 2 active validators.
	poaKeeper.createValidator(ctx, validator3)

	// Validator 3 cannot leave the validator set if still joining and
	// was not become active yet.
	err = poaKeeper.Leave(ctx, sdk.AccAddress(validator3.GetOperator()))
	expectedErr := errorsmod.Wrap(
		types.ErrWrongValidatorState,
		"not an active validator",
	)
	if err.Error() != expectedErr.Error() {
		t.Errorf(
			"Leave when still joining, error should be %v, got %v",
			expectedErr.Error(),
			err.Error(),
		)
	}

	// Run validator updates so the validator 3 state switch to ValidatorStateActive.
	// There are 3 active validators now.
	_, err = poaKeeper.EndBlocker(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Validator 1 can leave the validator set as there are 3 active validators.
	err = poaKeeper.Leave(ctx, sdk.AccAddress(validator1.GetOperator()))
	if err != nil {
		t.Errorf("Leave should leave the validator set, got error %v", err)
	}
	validatorState, found := poaKeeper.GetValidatorState(ctx, validator1.GetOperator())
	if !found {
		t.Errorf("Leave should not directly remove the validator")
	}
	if validatorState != types.ValidatorStateLeaving {
		t.Errorf("Leave should set the state of the validator to leaving")
	}

	// Validator 1 cannot leave the validator set as it's already leaving.
	err = poaKeeper.Leave(ctx, sdk.AccAddress(validator1.GetOperator()))
	expectedErr = errorsmod.Wrap(
		types.ErrWrongValidatorState,
		"not an active validator",
	)
	if err.Error() != expectedErr.Error() {
		t.Errorf(
			"Leave when already leaving, error should be %v, got %v",
			expectedErr.Error(),
			err.Error(),
		)
	}
}

func TestGetValidator(t *testing.T) {
	ctx, poaKeeper := mockContext()
	validator1, _ := mockValidator()
	validator2, _ := mockValidator()

	poaKeeper.setValidator(ctx, validator1)

	// Should find the correct validator.
	retrievedValidator, found := poaKeeper.GetValidator(
		ctx,
		validator1.GetOperator(),
	)
	if !found {
		t.Errorf("GetValidator should find validator if it has been set")
	}

	if !cmp.Equal(validator1, retrievedValidator) {
		t.Errorf(
			"GetValidator should find %v, found %v",
			validator1,
			retrievedValidator,
		)
	}

	// Should not find an unset validator
	_, found = poaKeeper.GetValidator(ctx, validator2.GetOperator())
	if found {
		t.Errorf("GetValidator should not find validator if it has not been set")
	}
}

func TestGetValidatorByConsAddr(t *testing.T) {
	ctx, poaKeeper := mockContext()
	validator1, _ := mockValidator()
	validator2, _ := mockValidator()

	poaKeeper.setValidator(ctx, validator1)
	poaKeeper.setValidatorByConsAddr(ctx, validator1)

	// Should find the correct validator
	retrievedValidator, found := poaKeeper.GetValidatorByConsAddr(
		ctx,
		validator1.GetConsAddress(),
	)
	if !found {
		t.Errorf("GetValidatorByConsAddr should find validator if it has been set")
	}

	if !cmp.Equal(validator1, retrievedValidator) {
		t.Errorf(
			"GetValidatorByConsAddr should find %v, found %v",
			validator1,
			retrievedValidator,
		)
	}

	// Should not find an unset validator.
	_, found = poaKeeper.GetValidator(ctx, validator2.GetOperator())
	if found {
		t.Errorf("GetValidatorByConsAddr should not find validator if it has not been set")
	}

	// Should not find the validator if we call setValidatorByConsAddr without SetValidator.
	poaKeeper.setValidatorByConsAddr(ctx, validator2)
	_, found = poaKeeper.GetValidator(ctx, validator2.GetOperator())
	if found {
		t.Errorf("GetValidatorByConsAddr should not find validator if it has not been set with SetValidator")
	}
}

func TestGetValidatorState(t *testing.T) {
	ctx, poaKeeper := mockContext()
	validator1, _ := mockValidator()
	validator2, _ := mockValidator()

	poaKeeper.setValidatorState(ctx, validator1, types.ValidatorStateActive)

	// Should find the correct validator
	retrievedState, found := poaKeeper.GetValidatorState(
		ctx,
		validator1.GetOperator(),
	)
	if !found {
		t.Errorf("GetValidatorState should find validator state if it has been set")
	}

	if retrievedState != types.ValidatorStateActive {
		t.Errorf(
			"GetValidatorState should find %v, found %v",
			validator1,
			types.ValidatorStateActive,
		)
	}

	// Should not find an unset validator
	_, found = poaKeeper.GetValidatorState(ctx, validator2.GetOperator())
	if found {
		t.Errorf("GetValidatorState should not find validator if it has not been set")
	}
}

func TestGetValidatorStatePanic(t *testing.T) {
	ctx, poaKeeper := mockContext()
	validator1, _ := mockValidator()

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The function did not panic on an unknown state")
		}
	}()

	// Should panic if the state doesn't exist
	poaKeeper.setValidatorState(ctx, validator1, types.ValidatorStateUnknown)
}

func TestCreateValidator(t *testing.T) {
	ctx, poaKeeper := mockContext()
	validator, _ := mockValidator()

	poaKeeper.createValidator(ctx, validator)

	_, foundVal := poaKeeper.GetValidator(ctx, validator.GetOperator())
	_, foundConsAddr := poaKeeper.GetValidatorByConsAddr(
		ctx,
		validator.GetConsAddress(),
	)
	_, foundState := poaKeeper.GetValidatorState(ctx, validator.GetOperator())

	if !foundVal || !foundConsAddr || !foundState {
		t.Errorf(
			"CreateValidator should append the validator. Found val: %v, found consAddr: %v, found state: %v",
			foundVal,
			foundConsAddr,
			foundState,
		)
	}
}

func TestRemoveValidator(t *testing.T) {
	ctx, poaKeeper := mockContext()
	validator1, _ := mockValidator()
	validator2, _ := mockValidator()

	helper, _ := mockValidator()
	owner := sdk.AccAddress(helper.GetOperator())

	poaKeeper.setOwner(ctx, owner)

	// Set validators
	poaKeeper.createValidator(ctx, validator1)
	poaKeeper.createValidator(ctx, validator2)

	// Add privilege1 to validator1 and validator2.
	err := poaKeeper.AddPrivilege(
		ctx,
		owner,
		[]sdk.ValAddress{validator1.GetOperator(), validator2.GetOperator()},
		"privilege1",
	)
	require.NoError(t, err)

	// Add privilege2 to validator1 and validator2.
	err = poaKeeper.AddPrivilege(
		ctx,
		owner,
		[]sdk.ValAddress{validator1.GetOperator(), validator2.GetOperator()},
		"privilege2",
	)
	require.NoError(t, err)

	poaKeeper.removeValidator(ctx, validator1.GetOperator())

	// Should not find a removed validator
	_, foundVal := poaKeeper.GetValidator(ctx, validator1.GetOperator())
	_, foundConsAddr := poaKeeper.GetValidatorByConsAddr(ctx, validator1.GetConsAddress())
	_, foundState := poaKeeper.GetValidatorState(ctx, validator1.GetOperator())

	if foundVal || foundConsAddr || foundState {
		t.Errorf(
			"RemoveValidator should remove validator record. Found val: %v, found consAddr: %v, found state: %v",
			foundVal,
			foundConsAddr,
			foundState,
		)
	}

	// Should still find a non-removed validator
	_, foundVal = poaKeeper.GetValidator(ctx, validator2.GetOperator())
	_, foundConsAddr = poaKeeper.GetValidatorByConsAddr(ctx, validator2.GetConsAddress())
	_, foundState = poaKeeper.GetValidatorState(ctx, validator2.GetOperator())

	if !foundVal || !foundConsAddr || !foundState {
		t.Errorf(
			"RemoveValidator should not remove validator 2 record. Found val: %v, found consAddr: %v, found state: %v",
			foundVal,
			foundConsAddr,
			foundState,
		)
	}

	// The removed validator should be removed from both privilege sets.
	// The non-removed validator should still be in both privilege sets.
	require.Equal(
		t,
		[]sdk.ConsAddress{validator2.GetConsAddress()},
		poaKeeper.GetValidatorsConsAddrsByPrivilege(ctx, "privilege1"),
	)
	require.Equal(
		t,
		[]sdk.ConsAddress{validator2.GetConsAddress()},
		poaKeeper.GetValidatorsConsAddrsByPrivilege(ctx, "privilege2"),
	)
}

func TestGetAllValidators(t *testing.T) {
	ctx, poaKeeper := mockContext()
	validator1, _ := mockValidator()
	validator2, _ := mockValidator()
	validator3, _ := mockValidator()

	poaKeeper.setValidator(ctx, validator1)
	poaKeeper.setValidator(ctx, validator2)
	poaKeeper.setValidator(ctx, validator3)

	poaKeeper.setValidatorState(ctx, validator1, types.ValidatorStateJoining)
	poaKeeper.setValidatorState(ctx, validator2, types.ValidatorStateActive)
	poaKeeper.setValidatorState(ctx, validator3, types.ValidatorStateLeaving)

	expectedValidators := []types.Validator{validator1, validator2, validator3}
	sort.Slice(expectedValidators, func(i, j int) bool {
		return bytes.Compare(
			expectedValidators[i].GetOperator(),
			expectedValidators[j].GetOperator(),
		) == -1
	})

	retrievedValidators := poaKeeper.GetAllValidators(ctx)
	sort.Slice(retrievedValidators, func(i, j int) bool {
		return bytes.Compare(
			retrievedValidators[i].GetOperator(),
			retrievedValidators[j].GetOperator(),
		) == -1
	})

	if !reflect.DeepEqual(expectedValidators, retrievedValidators) {
		t.Errorf(
			"GetAllValidators should find %v, found %v",
			expectedValidators,
			retrievedValidators,
		)
	}
}

func TestGetActiveValidators(t *testing.T) {
	ctx, poaKeeper := mockContext()
	validator1, _ := mockValidator()
	validator2, _ := mockValidator()
	validator3, _ := mockValidator()

	poaKeeper.setValidator(ctx, validator1)
	poaKeeper.setValidator(ctx, validator2)
	poaKeeper.setValidator(ctx, validator3)

	poaKeeper.setValidatorState(ctx, validator1, types.ValidatorStateJoining)
	poaKeeper.setValidatorState(ctx, validator2, types.ValidatorStateActive)
	poaKeeper.setValidatorState(ctx, validator3, types.ValidatorStateLeaving)

	expectedValidators := []types.Validator{validator2}

	retrievedValidators := poaKeeper.GetActiveValidators(ctx)

	if !reflect.DeepEqual(expectedValidators, retrievedValidators) {
		t.Errorf(
			"GetActiveValidators should find %v, found %v",
			expectedValidators,
			retrievedValidators,
		)
	}
}

func TestGetPubKeyByConsAddr(t *testing.T) {
	validator, _ := mockValidator()
	otherValidator, _ := mockValidator()

	tests := []struct {
		name           string
		blockHeight    int64
		prepareFn      func(sdk.Context, Keeper)
		expectedPubKey cryptotypes.PubKey
		expectedErr    error
	}{
		{
			name:        "validator in store",
			blockHeight: 100,
			prepareFn: func(ctx sdk.Context, k Keeper) {
				k.setValidator(ctx, validator)
				k.setValidatorByConsAddr(ctx, validator)
			},
			expectedPubKey: validator.GetConsPubKey(),
			expectedErr:    nil,
		},
		{
			name:        "validator not in store - present in historical info",
			blockHeight: 100,
			prepareFn: func(ctx sdk.Context, k Keeper) {
				k.SetHistoricalInfo(ctx, 99, &types.HistoricalInfo{Valset: []types.Validator{otherValidator}})
				k.SetHistoricalInfo(ctx, 98, &types.HistoricalInfo{Valset: []types.Validator{otherValidator}})
				k.SetHistoricalInfo(ctx, 97, &types.HistoricalInfo{Valset: []types.Validator{otherValidator}})
				k.SetHistoricalInfo(ctx, 96, &types.HistoricalInfo{Valset: []types.Validator{otherValidator}})
				k.SetHistoricalInfo(ctx, 95, &types.HistoricalInfo{Valset: []types.Validator{otherValidator}})
				k.SetHistoricalInfo(ctx, 94, &types.HistoricalInfo{Valset: []types.Validator{otherValidator}})
				k.SetHistoricalInfo(ctx, 93, &types.HistoricalInfo{Valset: []types.Validator{otherValidator}})
				k.SetHistoricalInfo(ctx, 92, &types.HistoricalInfo{Valset: []types.Validator{otherValidator}})
				k.SetHistoricalInfo(ctx, 91, &types.HistoricalInfo{Valset: []types.Validator{otherValidator}})
				// Validator exists in the last historical info being part of the search range.
				k.SetHistoricalInfo(ctx, 90, &types.HistoricalInfo{Valset: []types.Validator{otherValidator, validator}})
				k.SetHistoricalInfo(ctx, 89, &types.HistoricalInfo{Valset: []types.Validator{otherValidator, validator}})
			},
			expectedPubKey: validator.GetConsPubKey(),
			expectedErr:    nil,
		},
		{
			name:        "validator not in store - not present in historical info",
			blockHeight: 100,
			prepareFn: func(ctx sdk.Context, k Keeper) {
				k.SetHistoricalInfo(ctx, 99, &types.HistoricalInfo{Valset: []types.Validator{otherValidator}})
				k.SetHistoricalInfo(ctx, 98, &types.HistoricalInfo{Valset: []types.Validator{otherValidator}})
				k.SetHistoricalInfo(ctx, 97, &types.HistoricalInfo{Valset: []types.Validator{otherValidator}})
				k.SetHistoricalInfo(ctx, 96, &types.HistoricalInfo{Valset: []types.Validator{otherValidator}})
				k.SetHistoricalInfo(ctx, 95, &types.HistoricalInfo{Valset: []types.Validator{otherValidator}})
				k.SetHistoricalInfo(ctx, 94, &types.HistoricalInfo{Valset: []types.Validator{otherValidator}})
				k.SetHistoricalInfo(ctx, 93, &types.HistoricalInfo{Valset: []types.Validator{otherValidator}})
				k.SetHistoricalInfo(ctx, 92, &types.HistoricalInfo{Valset: []types.Validator{otherValidator}})
				k.SetHistoricalInfo(ctx, 91, &types.HistoricalInfo{Valset: []types.Validator{otherValidator}})
				k.SetHistoricalInfo(ctx, 90, &types.HistoricalInfo{Valset: []types.Validator{otherValidator}})
				// Validator exists in the last historical info that is just beyond the search range.
				k.SetHistoricalInfo(ctx, 89, &types.HistoricalInfo{Valset: []types.Validator{otherValidator, validator}})
			},
			expectedPubKey: nil,
			expectedErr:    types.ErrNoValidatorFound,
		},
		{
			name:        "validator not in store - not present in limited historical info",
			blockHeight: 100,
			prepareFn: func(ctx sdk.Context, k Keeper) {
				// Simulate historical info pruning. The count of existing
				// historical info to check is lesser than the size of the
				// search range.
				k.SetHistoricalInfo(ctx, 99, &types.HistoricalInfo{Valset: []types.Validator{otherValidator}})
				k.SetHistoricalInfo(ctx, 98, &types.HistoricalInfo{Valset: []types.Validator{otherValidator}})
				k.SetHistoricalInfo(ctx, 97, &types.HistoricalInfo{Valset: []types.Validator{otherValidator}})
				k.SetHistoricalInfo(ctx, 96, &types.HistoricalInfo{Valset: []types.Validator{otherValidator}})
				k.SetHistoricalInfo(ctx, 95, &types.HistoricalInfo{Valset: []types.Validator{otherValidator}})
			},
			expectedPubKey: nil,
			expectedErr:    types.ErrNoValidatorFound,
		},
		{
			name: "validator not in store - not present in initial historical info",
			// Simulate the beginning of the chain. The current block
			// of the chain is lesser than the size of the search range.
			blockHeight: 3,
			prepareFn: func(ctx sdk.Context, k Keeper) {
				k.SetHistoricalInfo(ctx, 2, &types.HistoricalInfo{Valset: []types.Validator{otherValidator}})
				k.SetHistoricalInfo(ctx, 1, &types.HistoricalInfo{Valset: []types.Validator{otherValidator}})
			},
			expectedPubKey: nil,
			expectedErr:    types.ErrNoValidatorFound,
		},
		{
			name:           "validator not in store - no historical info",
			blockHeight:    100,
			prepareFn:      func(_ sdk.Context, _ Keeper) {},
			expectedPubKey: nil,
			expectedErr:    types.ErrNoValidatorFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx, poaKeeper := mockContext()

			ctx = ctx.WithBlockHeight(test.blockHeight)

			test.prepareFn(ctx, poaKeeper)

			pubKey, err := poaKeeper.GetPubKeyByConsAddr(
				ctx,
				validator.GetConsAddress(),
			)

			if !reflect.DeepEqual(test.expectedErr, err) {
				t.Errorf(
					"unexpected error:\n"+
						"expected: %v\n"+
						"actual:   %v",
					test.expectedErr,
					err,
				)
			}

			var pubKeySdk cryptotypes.PubKey

			if err == nil {
				// Retrieved key is CometBFT-proto-specific. Convert it to the
				// Cosmos-SDK-specific type for comparison.
				var convErr error
				pubKeySdk, convErr = cryptocdc.FromCmtProtoPublicKey(pubKey)
				if convErr != nil {
					t.Error(convErr)
				}
			}

			if !cmp.Equal(test.expectedPubKey, pubKeySdk) {
				t.Errorf(
					"unexpected public key:\n"+
						"expected: %v\n"+
						"actual:   %v",
					test.expectedPubKey,
					pubKeySdk,
				)
			}
		})
	}
}
