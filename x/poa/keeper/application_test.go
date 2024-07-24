package keeper

import (
	"testing"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/evmos/evmos/v12/x/poa/types"
	"github.com/google/go-cmp/cmp"
)

func TestSubmitApplication(t *testing.T) {
	ctx, poaKeeper := mockContext()
	validator1, _ := mockValidator()
	validator2, _ := mockValidator()
	validator3, _ := mockValidator()

	// Max validators is 2.
	err := poaKeeper.setParams(ctx, types.NewParams(2))
	if err != nil {
		t.Fatal(err)
	}

	// Validator 1 is in the validator set from the beginning.
	poaKeeper.appendValidator(ctx, validator1)

	// Try to impersonate the operator of validator 2 and submit an application.
	err = poaKeeper.SubmitApplication(
		ctx,
		sdk.AccAddress(validator1.GetOperator()),
		validator2,
	)
	expectedErr := errorsmod.Wrapf(
		sdkerrors.ErrUnauthorized,
		"not the validator operator; expected %s, sender %s",
		sdk.AccAddress(validator2.GetOperator()).String(),
		sdk.AccAddress(validator1.GetOperator()).String(),
	)
	if err.Error() != expectedErr.Error() {
		t.Errorf(
			"SubmitApplication with wrong sender, error should be %v, got %v",
			expectedErr.Error(),
			err.Error(),
		)
	}

	// The application for validator 2 is submitted correctly
	err = poaKeeper.SubmitApplication(
		ctx,
		sdk.AccAddress(validator2.GetOperator()),
		validator2,
	)
	if err != nil {
		t.Errorf(
			"SubmitApplication should submit an application, got error %v",
			err,
		)
	}
	_, found := poaKeeper.GetApplication(ctx, validator2.GetOperator())
	if !found {
		t.Errorf("SubmitApplication should submit an application, the application has not been found")
	}
	_, found = poaKeeper.GetApplicationByConsAddr(
		ctx,
		validator2.GetConsAddress(),
	)
	if !found {
		t.Errorf("SubmitApplication should submit an application, the application has not been found by cons addr")
	}

	// A new application with the same validator cannot be created
	// (validator 2 just submitted an application).
	err = poaKeeper.SubmitApplication(
		ctx,
		sdk.AccAddress(validator2.GetOperator()),
		validator2,
	)
	if err.Error() != types.ErrAlreadyApplying.Error() {
		t.Errorf(
			"SubmitApplication with duplicate, error should be %v, got %v",
			types.ErrAlreadyApplying.Error(),
			err.Error(),
		)
	}

	// A new application cannot be created if the validator already exist
	// (validator 1 is in the validator set from the beginning).
	err = poaKeeper.SubmitApplication(
		ctx,
		sdk.AccAddress(validator1.GetOperator()),
		validator1,
	)
	if err.Error() != types.ErrAlreadyValidator.Error() {
		t.Errorf(
			"SubmitApplication with duplicate, error should be %v, got %v",
			types.ErrAlreadyValidator.Error(),
			err.Error(),
		)
	}

	// A new application cannot be created if the validator already exist,
	// even if they use a different consensus public key.
	// (validator 1 is in the validator set from the beginning).
	_, newPubKey := mockValidator()
	validator1Copy := validator1
	validator1Copy.ConsPubKeyBech32 = newPubKey
	err = poaKeeper.SubmitApplication(
		ctx,
		sdk.AccAddress(validator1.GetOperator()),
		validator1Copy,
	)
	if err.Error() != types.ErrAlreadyValidator.Error() {
		t.Errorf(
			"SubmitApplication with duplicate, error should be %v, got %v",
			types.ErrAlreadyValidator.Error(),
			err.Error(),
		)
	}

	// A new application cannot be created if the validator does not exist,
	// but they use a consensus public key that is already in use
	// (validator 1 is in the validator set from the beginning).
	validator3Copy := validator3
	validator3Copy.ConsPubKeyBech32 = validator1.GetConsPubKeyBech32()
	err = poaKeeper.SubmitApplication(
		ctx,
		sdk.AccAddress(validator3Copy.GetOperator()),
		validator3Copy,
	)
	if err.Error() != types.ErrAlreadyValidator.Error() {
		t.Errorf(
			"SubmitApplication with duplicate, error should be %v, got %v",
			types.ErrAlreadyValidator.Error(),
			err.Error(),
		)
	}

	// Test max validators condition.
	err = poaKeeper.setParams(ctx, types.NewParams(1))
	if err != nil {
		t.Fatal(err)
	}
	err = poaKeeper.SubmitApplication(
		ctx,
		sdk.AccAddress(validator3.GetOperator()),
		validator3,
	)
	if err.Error() != types.ErrMaxValidatorsReached.Error() {
		t.Errorf(
			"SubmitApplication with max validators reached, error should be %v, got %v",
			types.ErrMaxValidatorsReached.Error(),
			err.Error(),
		)
	}
}

func TestApproveApplication(t *testing.T) {
	ctx, poaKeeper := mockContext()

	// Max validators is 2.
	err := poaKeeper.setParams(ctx, types.NewParams(2))
	if err != nil {
		t.Fatal(err)
	}

	// Generate an owner address using the mockValidator function.
	helper, _ := mockValidator()
	owner := sdk.AccAddress(helper.GetOperator())
	poaKeeper.setOwner(ctx, owner)

	// Generate two validators submitting applications.
	validator1, _ := mockValidator()
	validator2, _ := mockValidator()

	err = poaKeeper.SubmitApplication(
		ctx,
		sdk.AccAddress(validator1.GetOperator()),
		validator1,
	)
	if err != nil {
		t.Fatal(err)
	}

	err = poaKeeper.SubmitApplication(
		ctx,
		sdk.AccAddress(validator2.GetOperator()),
		validator2,
	)
	if err != nil {
		t.Fatal(err)
	}

	// Try to impersonate the owner and approve the application of validator 1.
	err = poaKeeper.ApproveApplication(
		ctx,
		sdk.AccAddress(validator1.GetOperator()),
		validator1.GetOperator(),
	)
	expectedErr := errorsmod.Wrapf(
		sdkerrors.ErrUnauthorized,
		"not the owner; expected %s, sender %s",
		owner.String(),
		sdk.AccAddress(validator1.GetOperator()).String(),
	)
	if err.Error() != expectedErr.Error() {
		t.Errorf(
			"ApproveApplication with wrong sender, error should be %v, got %v",
			expectedErr.Error(),
			err.Error(),
		)
	}

	// Approve the application of validator 1.
	err = poaKeeper.ApproveApplication(
		ctx,
		owner,
		validator1.GetOperator(),
	)
	if err != nil {
		t.Errorf("ApproveApplication should pass, got error %v", err)
	}
	// Make sure the application has been removed.
	_, found := poaKeeper.GetApplication(ctx, validator1.GetOperator())
	if found {
		t.Errorf("ApproveApplication should remove the application from the set")
	}
	_, found = poaKeeper.GetApplicationByConsAddr(ctx, validator1.GetConsAddress())
	if found {
		t.Errorf("ApproveApplication should remove the application from the index by consensus address")
	}
	// Make sure the validator has been added.
	_, found = poaKeeper.GetValidator(ctx, validator1.GetOperator())
	if !found {
		t.Errorf("ApproveApplication should add the candidate to the set")
	}
	_, found = poaKeeper.GetValidatorByConsAddr(ctx, validator1.GetConsAddress())
	if !found {
		t.Errorf("ApproveApplication should add the candidate to the index by consensus address")
	}

	// Approve the validator 1 application again and make sure it fails
	// due to a non-existing application.
	err = poaKeeper.ApproveApplication(
		ctx,
		owner,
		validator1.GetOperator(),
	)
	if err.Error() != types.ErrNoApplicationFound.Error() {
		t.Errorf(
			"ApproveApplication with non-existing application, error should be %v, got %v",
			types.ErrNoApplicationFound.Error(),
			err.Error(),
		)
	}

	// Test max validators condition. Try to approve the application of validator 2.
	err = poaKeeper.setParams(ctx, types.NewParams(1))
	if err != nil {
		t.Fatal(err)
	}
	err = poaKeeper.ApproveApplication(
		ctx,
		owner,
		validator2.GetOperator(),
	)
	if err.Error() != types.ErrMaxValidatorsReached.Error() {
		t.Errorf(
			"ApproveApplication with max validators reached, error should be %v, got %v",
			types.ErrMaxValidatorsReached.Error(),
			err.Error(),
		)
	}
}

func TestGetApplication(t *testing.T) {
	ctx, poaKeeper := mockContext()
	validator1, _ := mockValidator()
	validator2, _ := mockValidator()
	application := types.NewApplication(validator1)

	poaKeeper.setApplication(ctx, application)

	// Should find the correct application
	retrievedApplication, found := poaKeeper.GetApplication(
		ctx,
		validator1.GetOperator(),
	)
	if !found {
		t.Errorf("GetApplication should find application if it has been set")
	}

	if !cmp.Equal(
		application.GetValidator(),
		retrievedApplication.GetValidator(),
	) {
		t.Errorf(
			"GetApplication should find %v, found %v",
			application.GetValidator(),
			retrievedApplication.GetValidator(),
		)
	}

	// Should not find an unset application
	_, found = poaKeeper.GetApplication(ctx, validator2.GetOperator())
	if found {
		t.Errorf("GetApplication should not find application if it has not been set")
	}
}

func TestGetApplicationByConsAddr(t *testing.T) {
	ctx, poaKeeper := mockContext()
	validator1, _ := mockValidator()
	validator2, _ := mockValidator()
	application1 := types.NewApplication(validator1)
	application2 := types.NewApplication(validator2)

	poaKeeper.setApplication(ctx, application1)
	poaKeeper.setApplicationByConsAddr(ctx, application1)

	// Should find the correct application
	retrievedApplication, found := poaKeeper.GetApplicationByConsAddr(
		ctx,
		application1.GetValidator().GetConsAddress(),
	)
	if !found {
		t.Errorf("GetApplicationByConsAddr should find application if it has been set")
	}

	if !cmp.Equal(
		application1.GetValidator(),
		retrievedApplication.GetValidator(),
	) {
		t.Errorf(
			"GetApplicationByConsAddr should find %v, found %v",
			application1.GetValidator(),
			retrievedApplication.GetValidator(),
		)
	}

	// Should not find an unset application
	_, found = poaKeeper.GetApplication(ctx, validator2.GetOperator())
	if found {
		t.Errorf("GetApplicationByConsAddr should not find application if it has not been set")
	}

	// Should not find the application if we call SetApplicationByConsAddr without SetApplication
	poaKeeper.setApplicationByConsAddr(ctx, application2)
	_, found = poaKeeper.GetApplication(ctx, validator2.GetOperator())
	if found {
		t.Errorf("GetApplicationByConsAddr should not find application if it has not been set with SetApplication")
	}
}

func TestAppendApplication(t *testing.T) {
	ctx, poaKeeper := mockContext()
	validator, _ := mockValidator()

	poaKeeper.appendApplication(ctx, validator)

	_, foundApplication := poaKeeper.GetApplication(
		ctx,
		validator.GetOperator(),
	)
	_, foundConsAddr := poaKeeper.GetApplicationByConsAddr(
		ctx,
		validator.GetConsAddress(),
	)

	if !foundApplication || !foundConsAddr {
		t.Errorf(
			"AppendValidator should append the application. Found val: %v, found consAddr: %v",
			foundApplication,
			foundConsAddr,
		)
	}
}

func TestRemoveApplication(t *testing.T) {
	ctx, poaKeeper := mockContext()
	validator, _ := mockValidator()

	// Append  and remove application
	poaKeeper.appendApplication(ctx, validator)
	poaKeeper.removeApplication(ctx, validator.GetOperator())

	// Should not find a removed validator
	_, foundApplication := poaKeeper.GetApplication(
		ctx,
		validator.GetOperator(),
	)
	_, foundConsAddr := poaKeeper.GetApplicationByConsAddr(
		ctx,
		validator.GetConsAddress(),
	)

	if foundApplication || foundConsAddr {
		t.Errorf(
			"RemoveApplication should remove application record. Found val: %v, found consAddr: %v",
			foundApplication,
			foundConsAddr,
		)
	}
}

func TestGetAllApplications(t *testing.T) {
	ctx, poaKeeper := mockContext()
	validator1, _ := mockValidator()
	validator2, _ := mockValidator()
	application1 := types.NewApplication(validator1)
	application2 := types.NewApplication(validator2)

	poaKeeper.setApplication(ctx, application1)
	poaKeeper.setApplication(ctx, application2)

	retrievedApplications := poaKeeper.GetAllApplications(ctx)
	if len(retrievedApplications) != 2 {
		t.Errorf(
			"GetAllApplications should find %v applications, found %v",
			2,
			len(retrievedApplications),
		)
	}
}
