package poa_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	tmtypes "github.com/tendermint/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v12/x/poa"
	"github.com/evmos/evmos/v12/x/poa/types"
)

func TestValidateGenesis(t *testing.T) {
	validator, _ := poa.MockValidator()

	// Valid genesis
	validGenesis := types.NewGenesisState(types.DefaultParams(), []types.Validator{validator})
	if types.ValidateGenesis(validGenesis) != nil {
		t.Errorf("The genesis state %v should be valid", validGenesis)
	}

	// A genesis with two validators with the same consensus pukey is invalid
	invalidGenesis := types.NewGenesisState(types.DefaultParams(), []types.Validator{validator, validator})
	if types.ValidateGenesis(invalidGenesis) == nil {
		t.Errorf("The genesis state %v should not be valid", invalidGenesis)
	}

	// Default genesis state
	if types.ValidateGenesis(types.DefaultGenesisState()) != nil {
		t.Errorf("The default genesis state should be valid")
	}
}

func TestInitGenesis(t *testing.T) {
	ctx, poaKeeper := poa.MockContext()
	validator, consPubKey := poa.MockValidator()

	// Test genesis data
	testGenesis := types.NewGenesisState(types.DefaultParams(), []types.Validator{validator})

	// InitGenesis
	validatorUpdates := poa.InitGenesis(ctx, poaKeeper, testGenesis)

	// Only one update
	if len(validatorUpdates) != 1 {
		t.Errorf("Should get exactly one validator update")
	}

	// No weight
	power := validatorUpdates[0].Power
	if power != 1 {
		t.Errorf("power should be 1, got %v", power)
	}

	// Correct public key
	pubKey, err := tmtypes.PB2TM.PubKey(validatorUpdates[0].PubKey)
	if err != nil {
		t.Errorf("Incorrect public key: %v", err)
	}
	pubKeyString := sdk.MustBech32ifyPubKey(sdk.Bech32PubKeyTypeConsPub, pubKey)
	if pubKeyString != consPubKey {
		t.Errorf("validator PubKey should be %v, got %v", consPubKey, pubKeyString)
	}
}

func TestExportGenesis(t *testing.T) {
	ctx, poaKeeper := poa.MockContext()
	validator, _ := poa.MockValidator()

	// Manually set values in keeper
	poaKeeper.SetValidator(ctx, validator)
	poaKeeper.SetParams(ctx, types.DefaultParams())

	exportedGenesis := poa.ExportGenesis(ctx, poaKeeper)

	if !cmp.Equal(exportedGenesis.Params, types.DefaultParams()) {
		t.Errorf("Exported genesis param shoud be: %v, not %v", types.DefaultParams(), exportedGenesis.Params)
	}

	if !cmp.Equal(exportedGenesis.Validators, []types.Validator{validator}) {
		t.Errorf("Exported genesis validators shoud be: %v, not %v", []types.Validator{validator}, exportedGenesis.Validators)
	}
}
