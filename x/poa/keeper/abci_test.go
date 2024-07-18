package keeper

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/evmos/evmos/v12/x/poa/types"
)

func TestEndBlocker(t *testing.T) {
	ctx, poaKeeper := mockContext()
	validator1, _ := mockValidator()
	validator2, _ := mockValidator()
	validator3, _ := mockValidator()
	validator4, _ := mockValidator()
	validator5, _ := mockValidator()

	// Set the validator in the keeper
	poaKeeper.appendValidator(ctx, validator1)
	poaKeeper.appendValidator(ctx, validator2)
	poaKeeper.appendValidator(ctx, validator3)
	poaKeeper.appendValidator(ctx, validator4)
	poaKeeper.appendValidator(ctx, validator5)

	// Simulate validator 2 as if it is already in the validator set
	poaKeeper.setValidatorState(ctx, validator2, types.ValidatorStateActive)

	// Simulate validator 4 and 5 as if those are leaving the validator set
	poaKeeper.setValidatorState(ctx, validator4, types.ValidatorStateLeaving)
	poaKeeper.setValidatorState(ctx, validator5, types.ValidatorStateLeaving)

	updates := poaKeeper.EndBlocker(ctx)

	// There should be 4 updates
	if len(updates) != 4 {
		t.Errorf("EndBlocker should perform 4 updates, found %v updates", len(updates))
	}

	// Check the updates
	val1Update := validator1.ABCIValidatorUpdateAppend()
	val3Update := validator3.ABCIValidatorUpdateAppend()
	val4Update := validator4.ABCIValidatorUpdateRemove()
	val5Update := validator5.ABCIValidatorUpdateRemove()
	for _, update := range updates {
		// Check if the update has the correct power
		switch {
		case cmp.Equal(update.GetPubKey(), val1Update.GetPubKey()):
			if update.GetPower() != 1 {
				t.Errorf("Validator 1 should join")
			}
		case cmp.Equal(update.GetPubKey(), val3Update.GetPubKey()):
			if update.GetPower() != 1 {
				t.Errorf("Validator 3 should join")
			}
		case cmp.Equal(update.GetPubKey(), val4Update.GetPubKey()):
			if update.GetPower() != 0 {
				t.Errorf("Validator 4 should leave")
			}
		case cmp.Equal(update.GetPubKey(), val5Update.GetPubKey()):
			if update.GetPower() != 0 {
				t.Errorf("Validator 5 should leave")
			}
		default:
			t.Errorf("EndBlocker returns an unknown update: %v", update)
		}
	}

	// Check remaining validators in the keeper
	_, found1 := poaKeeper.GetValidator(ctx, validator1.GetOperator())
	_, found2 := poaKeeper.GetValidator(ctx, validator2.GetOperator())
	_, found3 := poaKeeper.GetValidator(ctx, validator3.GetOperator())
	_, found4 := poaKeeper.GetValidator(ctx, validator4.GetOperator())
	_, found5 := poaKeeper.GetValidator(ctx, validator5.GetOperator())

	if !found1 || !found2 || !found3 {
		t.Errorf("EndBlocker should leave validator 1, 2 and 3 in the set: %v, %v, %v", found1, found2, found3)
	}
	if found4 || found5 {
		t.Errorf("EndBlocker should remove validator 4 and 5 from the set: %v, %v", found4, found5)
	}
}
