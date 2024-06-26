package keeper_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/evmos/evmos/v12/x/poa"
	"github.com/evmos/evmos/v12/x/poa/types"
)

func TestGetValidator(t *testing.T) {
	ctx, poaKeeper := poa.MockContext()
	validator1, _ := poa.MockValidator()
	validator2, _ := poa.MockValidator()

	poaKeeper.SetValidator(ctx, validator1)

	// Should find the correct validator
	retrievedValidator, found := poaKeeper.GetValidator(ctx, validator1.GetOperator())
	if !found {
		t.Errorf("GetValidator should find validator if it has been set")
	}

	if !cmp.Equal(validator1, retrievedValidator) {
		t.Errorf("GetValidator should find %v, found %v", validator1, retrievedValidator)
	}

	// Should not find a unset validator
	_, found = poaKeeper.GetValidator(ctx, validator2.GetOperator())
	if found {
		t.Errorf("GetValidator should not find validator if it has not been set")
	}
}

func TestGetValidatorByConsAddr(t *testing.T) {
	ctx, poaKeeper := poa.MockContext()
	validator1, _ := poa.MockValidator()
	validator2, _ := poa.MockValidator()

	poaKeeper.SetValidator(ctx, validator1)
	poaKeeper.SetValidatorByConsAddr(ctx, validator1)

	// Should find the correct validator
	retrievedValidator, found := poaKeeper.GetValidatorByConsAddr(ctx, validator1.GetConsAddr())
	if !found {
		t.Errorf("GetValidatorByConsAddr should find validator if it has been set")
	}

	if !cmp.Equal(validator1, retrievedValidator) {
		t.Errorf("GetValidatorByConsAddr should find %v, found %v", validator1, retrievedValidator)
	}

	// Should not find a unset validator
	_, found = poaKeeper.GetValidator(ctx, validator2.GetOperator())
	if found {
		t.Errorf("GetValidatorByConsAddr should not find validator if it has not been set")
	}

	// Should not find the validator if we call SetValidatorByConsAddr without SetValidator
	poaKeeper.SetValidatorByConsAddr(ctx, validator2)
	_, found = poaKeeper.GetValidator(ctx, validator2.GetOperator())
	if found {
		t.Errorf("GetValidatorByConsAddr should not find validator if it has not been set with SetValidator")
	}
}

func TestGetValidatorState(t *testing.T) {
	ctx, poaKeeper := poa.MockContext()
	validator1, _ := poa.MockValidator()
	validator2, _ := poa.MockValidator()

	poaKeeper.SetValidatorState(ctx, validator1, types.ValidatorStateJoined)

	// Should find the correct validator
	retrievedState, found := poaKeeper.GetValidatorState(ctx, validator1.GetOperator())
	if !found {
		t.Errorf("GetValidatorState should find validator state if it has been set")
	}

	if !cmp.Equal(types.ValidatorStateJoined, retrievedState) {
		t.Errorf("GetValidatorState should find %v, found %v", validator1, types.ValidatorStateJoined)
	}

	// Should not find a unset validator
	_, found = poaKeeper.GetValidatorState(ctx, validator2.GetOperator())
	if found {
		t.Errorf("GetValidator should not find validator if it has not been set")
	}
}

func TestGetValidatorStatePanic(t *testing.T) {
	ctx, poaKeeper := poa.MockContext()
	validator1, _ := poa.MockValidator()

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The function did not panic on an unknown state")
		}
	}()

	// Should panic if the state doesn't exist
	poaKeeper.SetValidatorState(ctx, validator1, 1000)
}

func TestAppendValidator(t *testing.T) {
	ctx, poaKeeper := poa.MockContext()
	validator, _ := poa.MockValidator()

	poaKeeper.AppendValidator(ctx, validator)

	_, foundVal := poaKeeper.GetValidator(ctx, validator.GetOperator())
	_, foundConsAddr := poaKeeper.GetValidatorByConsAddr(ctx, validator.GetConsAddr())
	_, foundState := poaKeeper.GetValidatorState(ctx, validator.GetOperator())

	if !foundVal || !foundConsAddr || !foundState {
		t.Errorf("AppendValidator should append the validator. Found val: %v, found consAddr: %v, found state: %v", foundVal, foundConsAddr, foundState)
	}
}

func TestRemoveValidator(t *testing.T) {
	ctx, poaKeeper := poa.MockContext()
	validator1, _ := poa.MockValidator()
	validator2, _ := poa.MockValidator()

	// Set validators
	poaKeeper.AppendValidator(ctx, validator1)
	poaKeeper.AppendValidator(ctx, validator2)

	poaKeeper.RemoveValidator(ctx, validator1.GetOperator())

	// Should not find a removed validator
	_, foundVal := poaKeeper.GetValidator(ctx, validator1.GetOperator())
	_, foundConsAddr := poaKeeper.GetValidatorByConsAddr(ctx, validator1.GetConsAddr())
	_, foundState := poaKeeper.GetValidatorState(ctx, validator1.GetOperator())

	if foundVal || foundConsAddr || foundState {
		t.Errorf("RemoveValidator should remove validator record. Found val: %v, found consAddr: %v, found state: %v", foundVal, foundConsAddr, foundState)
	}

	// Should still find a non removed validator
	_, foundVal = poaKeeper.GetValidator(ctx, validator2.GetOperator())
	_, foundConsAddr = poaKeeper.GetValidatorByConsAddr(ctx, validator2.GetConsAddr())
	_, foundState = poaKeeper.GetValidatorState(ctx, validator2.GetOperator())

	if !foundVal || !foundConsAddr || !foundState {
		t.Errorf("RemoveValidator should not remove validator 2 record. Found val: %v, found consAddr: %v, found state: %v", foundVal, foundConsAddr, foundState)
	}
}

func TestGetAllValidators(t *testing.T) {
	ctx, poaKeeper := poa.MockContext()
	validator1, _ := poa.MockValidator()
	validator2, _ := poa.MockValidator()

	poaKeeper.SetValidator(ctx, validator1)
	poaKeeper.SetValidator(ctx, validator2)

	retrievedValidators := poaKeeper.GetAllValidators(ctx)
	if len(retrievedValidators) != 2 {
		t.Errorf("GetAllValidators should find %v validators, found %v", 2, len(retrievedValidators))
	}
}
