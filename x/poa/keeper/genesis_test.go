package keeper

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	cryptocdc "github.com/cosmos/cosmos-sdk/crypto/codec"
	//nolint:staticcheck
	"github.com/cosmos/cosmos-sdk/types/bech32/legacybech32"

	"github.com/google/go-cmp/cmp"
	"github.com/mezo-org/mezod/x/poa/types"
)

func TestValidateGenesis(t *testing.T) {
	// Generate an owner address using the mockValidator function.
	helper, _ := mockValidator()
	owner := sdk.AccAddress(helper.GetOperator())

	// Valid genesis.
	validator, _ := mockValidator()
	validGenesis := types.NewGenesisState(
		types.DefaultParams(),
		owner,
		[]types.Validator{validator},
	)
	if types.ValidateGenesis(validGenesis) != nil {
		t.Errorf("the genesis state %v should be valid", validGenesis)
	}

	// A genesis with two validators with the same operator address is invalid.
	validatorSameOp1, _ := mockValidator()
	validatorSameOp2, _ := mockValidator()
	validatorSameOp2.OperatorBech32 = validatorSameOp1.OperatorBech32
	invalidGenesis := types.NewGenesisState(
		types.DefaultParams(),
		owner,
		[]types.Validator{validatorSameOp1, validatorSameOp2},
	)
	if types.ValidateGenesis(invalidGenesis) == nil {
		t.Errorf("the genesis state %v should not be valid", invalidGenesis)
	}

	// A genesis with two validators with the same consensus pubkey is invalid.
	validatorSameCons1, _ := mockValidator()
	validatorSameCons2, _ := mockValidator()
	validatorSameCons2.ConsPubKeyBech32 = validatorSameCons1.ConsPubKeyBech32
	invalidGenesis = types.NewGenesisState(
		types.DefaultParams(),
		owner,
		[]types.Validator{validatorSameCons1, validatorSameCons2},
	)
	if types.ValidateGenesis(invalidGenesis) == nil {
		t.Errorf("the genesis state %v should not be valid", invalidGenesis)
	}

	// Default genesis state.
	genesisState := types.DefaultGenesisState()
	expectedErr := fmt.Errorf("invalid owner address : empty address string is not allowed")
	if err := types.ValidateGenesis(*genesisState); err.Error() != expectedErr.Error() {
		t.Errorf(
			"the default genesis state should be invalid due to a missing owner, error should be %v, got %v",
			expectedErr.Error(),
			err.Error(),
		)
	}
}

func TestValidateGenesisEmergencyTeam(t *testing.T) {
	// Generate an owner address using the mockValidator function.
	helper, _ := mockValidator()
	owner := sdk.AccAddress(helper.GetOperator())

	// Generate an emergency team address using the mockValidator function.
	helper, _ = mockValidator()
	emergencyTeam := sdk.AccAddress(helper.GetOperator())

	validator, _ := mockValidator()

	// An empty emergency team is valid. The role is optional.
	genesis := types.NewGenesisState(
		types.DefaultParams(),
		owner,
		[]types.Validator{validator},
	)
	if genesis.EmergencyTeam != "" {
		t.Errorf("the default genesis state emergency team should be empty, got %v", genesis.EmergencyTeam)
	}
	if types.ValidateGenesis(genesis) != nil {
		t.Errorf("the genesis state %v should be valid", genesis)
	}

	// A properly encoded emergency team is valid.
	genesis.EmergencyTeam = emergencyTeam.String()
	if types.ValidateGenesis(genesis) != nil {
		t.Errorf("the genesis state %v should be valid", genesis)
	}

	// A malformed emergency team is invalid.
	genesis.EmergencyTeam = "not-a-bech32-address"
	if types.ValidateGenesis(genesis) == nil {
		t.Errorf("the genesis state %v should not be valid", genesis)
	}
}

func TestInitAndExportGenesisEmergencyTeam(t *testing.T) {
	// Generate an owner address using the mockValidator function.
	helper, _ := mockValidator()
	owner := sdk.AccAddress(helper.GetOperator())

	// Generate an emergency team address using the mockValidator function.
	helper, _ = mockValidator()
	emergencyTeam := sdk.AccAddress(helper.GetOperator())

	validator, _ := mockValidator()

	testGenesis := types.NewGenesisState(
		types.DefaultParams(),
		owner,
		[]types.Validator{validator},
	)
	testGenesis.EmergencyTeam = emergencyTeam.String()

	ctx, poaKeeper := mockContext()
	poaKeeper.InitGenesis(ctx, testGenesis)

	currentEmergencyTeam := poaKeeper.GetEmergencyTeam(ctx)
	if !currentEmergencyTeam.Equals(emergencyTeam) {
		t.Errorf(
			"InitGenesis should set the emergency team to %v, got %v",
			emergencyTeam,
			currentEmergencyTeam,
		)
	}

	exportedGenesis := poaKeeper.ExportGenesis(ctx)
	if exportedGenesis.EmergencyTeam != emergencyTeam.String() {
		t.Errorf(
			"exported genesis emergency team should be %v, not %v",
			emergencyTeam.String(),
			exportedGenesis.EmergencyTeam,
		)
	}
}

func TestInitAndExportGenesisWithoutEmergencyTeam(t *testing.T) {
	// Generate an owner address using the mockValidator function.
	helper, _ := mockValidator()
	owner := sdk.AccAddress(helper.GetOperator())

	validator, _ := mockValidator()

	testGenesis := types.NewGenesisState(
		types.DefaultParams(),
		owner,
		[]types.Validator{validator},
	)

	ctx, poaKeeper := mockContext()
	poaKeeper.InitGenesis(ctx, testGenesis)

	if !poaKeeper.GetEmergencyTeam(ctx).Empty() {
		t.Errorf("InitGenesis without an emergency team should leave the role unset")
	}

	exportedGenesis := poaKeeper.ExportGenesis(ctx)
	if exportedGenesis.EmergencyTeam != "" {
		t.Errorf(
			"exported genesis emergency team should be empty, not %v",
			exportedGenesis.EmergencyTeam,
		)
	}
}

func TestInitGenesis(t *testing.T) {
	// Generate an owner address using the mockValidator function.
	helper, _ := mockValidator()
	owner := sdk.AccAddress(helper.GetOperator())

	ctx, poaKeeper := mockContext()
	validator, consPubKey := mockValidator()

	testGenesis := types.NewGenesisState(
		types.DefaultParams(),
		owner,
		[]types.Validator{validator},
	)

	validatorUpdates := poaKeeper.InitGenesis(ctx, testGenesis)

	if len(validatorUpdates) != 1 {
		t.Errorf("should get exactly one validator update")
	}

	power := validatorUpdates[0].Power
	if power != 1 {
		t.Errorf("power should be 1, got %v", power)
	}

	// Correct public key
	pubKey, err := cryptocdc.FromCmtProtoPublicKey(validatorUpdates[0].PubKey)
	if err != nil {
		t.Errorf("incorrect public key: %v", err)
	}

	//nolint:staticcheck
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
	err := poaKeeper.setParams(ctx, types.DefaultParams())
	if err != nil {
		t.Fatal(err)
	}

	exportedGenesis := poaKeeper.ExportGenesis(ctx)

	if !cmp.Equal(exportedGenesis.Params, types.DefaultParams()) {
		t.Errorf(
			"exported genesis params shoud be: %v, not %v",
			types.DefaultParams(),
			exportedGenesis.Params,
		)
	}

	if !cmp.Equal(exportedGenesis.Validators, []types.Validator{validator}) {
		t.Errorf(
			"exported genesis validators should be: %v, not %v",
			[]types.Validator{validator},
			exportedGenesis.Validators,
		)
	}
}
