package keeper

import (
	"errors"
	"testing"

	"cosmossdk.io/math"

	"github.com/google/go-cmp/cmp"
	"github.com/mezo-org/mezod/x/poa/types"
)

func TestValidatorByConsAddr(t *testing.T) {
	ctx, poaKeeper := mockContext()
	validator1, _ := mockValidator()
	validator2, _ := mockValidator()

	poaKeeper.setValidator(ctx, validator1)
	poaKeeper.setValidatorByConsAddr(ctx, validator1)

	// Should find a validator
	_, err := poaKeeper.ValidatorByConsAddr(
		ctx,
		validator1.GetConsAddress(),
	)
	if err != nil {
		t.Errorf("ValidatorByConsAddr should find validator if it has been set")
	}

	// Should not find an unset validator.
	_, err = poaKeeper.ValidatorByConsAddr(ctx, validator2.GetConsAddress())
	if !errors.Is(err, types.ErrNoValidatorFound) {
		t.Errorf("ValidatorByConsAddr should not find validator if it has not been set")
	}

	// Should not find the validator if we call setValidatorByConsAddr without SetValidator.
	poaKeeper.setValidatorByConsAddr(ctx, validator2)
	_, err = poaKeeper.ValidatorByConsAddr(ctx, validator2.GetConsAddress())
	if !errors.Is(err, types.ErrNoValidatorFound) {
		t.Errorf("ValidatorByConsAddr should not find validator if it has not been set with SetValidator")
	}
}

func TestTotalBondedTokens(t *testing.T) {
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

	expectedTotal := math.NewInt(2)

	retrievedTotal, err := poaKeeper.TotalBondedTokens(ctx)
	// Should not error
	if err != nil {
		t.Errorf(
			"TotalBondedTokens failed %v", err,
		)
	}

	// Should find the expected number of active vals
	if !cmp.Equal(retrievedTotal, expectedTotal) {
		t.Errorf("TotalBondedTokens got %v, want %v", retrievedTotal, expectedTotal)
	}
}
