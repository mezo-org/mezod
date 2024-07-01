package keeper

import (
	cryptocdc "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/types/bech32/legacybech32"
	"testing"

	"github.com/evmos/evmos/v12/x/poa/types"
	"github.com/google/go-cmp/cmp"
)

func TestValidateGenesis(t *testing.T) {
	validator, _ := mockValidator()

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
	genesisState := types.DefaultGenesisState()
	if types.ValidateGenesis(*genesisState) != nil {
		t.Errorf("The default genesis state should be valid")
	}
}

func TestInitGenesis(t *testing.T) {
	ctx, poaKeeper := mockContext()
	validator, consPubKey := mockValidator()

	// Test genesis data
	testGenesis := types.NewGenesisState(types.DefaultParams(), []types.Validator{validator})

	// InitGenesis
	validatorUpdates := poaKeeper.InitGenesis(ctx, testGenesis)

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
	pubKey, err := cryptocdc.FromTmProtoPublicKey(validatorUpdates[0].PubKey)
	if err != nil {
		t.Errorf("Incorrect public key: %v", err)
	}
	pubKeyString := legacybech32.MustMarshalPubKey(legacybech32.ConsPK, pubKey)
	if pubKeyString != consPubKey {
		t.Errorf("validator PubKey should be %v, got %v", consPubKey, pubKeyString)
	}
}

func TestExportGenesis(t *testing.T) {
	ctx, poaKeeper := mockContext()
	validator, _ := mockValidator()

	// Manually set values in keeper
	poaKeeper.setValidator(ctx, validator)
	poaKeeper.setParams(ctx, types.DefaultParams())

	exportedGenesis := poaKeeper.ExportGenesis(ctx)

	if !cmp.Equal(exportedGenesis.Params, types.DefaultParams()) {
		t.Errorf("Exported genesis param shoud be: %v, not %v", types.DefaultParams(), exportedGenesis.Params)
	}

	if !cmp.Equal(exportedGenesis.Validators, []types.Validator{validator}) {
		t.Errorf("Exported genesis validators shoud be: %v, not %v", []types.Validator{validator}, exportedGenesis.Validators)
	}
}
