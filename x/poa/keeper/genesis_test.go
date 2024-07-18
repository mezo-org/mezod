package keeper

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"testing"

	cryptocdc "github.com/cosmos/cosmos-sdk/crypto/codec"
	//nolint:staticcheck
	"github.com/cosmos/cosmos-sdk/types/bech32/legacybech32"

	"github.com/evmos/evmos/v12/x/poa/types"
	"github.com/google/go-cmp/cmp"
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
	pubKey, err := cryptocdc.FromTmProtoPublicKey(validatorUpdates[0].PubKey)
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
