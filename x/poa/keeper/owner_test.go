package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/evmos/evmos/v12/x/poa/types"
	"testing"
)

func TestTransferOwnership(t *testing.T) {
	ctx, poaKeeper := mockContext()

	// Generate an owner address using the mockValidator function.
	helper, _ := mockValidator()
	owner := sdk.AccAddress(helper.GetOperator())
	poaKeeper.setOwner(ctx, owner)

	// Generate a candidate owner address using the mockValidator function.
	helper, _ = mockValidator()
	candidateOwner := sdk.AccAddress(helper.GetOperator())

	// Make sure the function fails if the sender is not the current owner.
	err := poaKeeper.TransferOwnership(ctx, candidateOwner, candidateOwner)
	expectedErr := errorsmod.Wrapf(
		sdkerrors.ErrUnauthorized,
		"not the owner; expected %s, sender %s",
		owner.String(),
		candidateOwner.String(),
	)
	if err.Error() != expectedErr.Error() {
		t.Errorf(
			"TransferOwnership with wrong sender, error should be %v, got %v",
			expectedErr.Error(),
			err.Error(),
		)
	}

	err = poaKeeper.TransferOwnership(ctx, owner, candidateOwner)
	if err != nil {
		t.Errorf("TransferOwnership should pass, got error %v", err)
	}
	currentOwner := poaKeeper.GetOwner(ctx)
	if !currentOwner.Equals(owner) {
		t.Errorf("TransferOwnership should not change the owner imediately")
	}
	currentCandidateOwner := poaKeeper.GetCandidateOwner(ctx)
	if !currentCandidateOwner.Equals(candidateOwner) {
		t.Errorf("TransferOwnership should properly set the candidate owner")
	}
}

func TestAcceptOwnership(t *testing.T) {
	ctx, poaKeeper := mockContext()

	// Generate an owner address using the mockValidator function.
	helper, _ := mockValidator()
	owner := sdk.AccAddress(helper.GetOperator())
	poaKeeper.setOwner(ctx, owner)

	// Generate a candidate owner address using the mockValidator function.
	helper, _ = mockValidator()
	candidateOwner := sdk.AccAddress(helper.GetOperator())

	// AcceptOwnership should fail if the ownership transfer was not initialized.
	err := poaKeeper.AcceptOwnership(ctx, candidateOwner)
	if err.Error() != types.ErrOwnershipTransferNotInitialized.Error() {
		t.Errorf(
			"AcceptOwnership when not initalized, error should be %v, got %v",
			types.ErrOwnershipTransferNotInitialized.Error(),
			err.Error(),
		)
	}

	// Initialize the ownership transfer.
	err = poaKeeper.TransferOwnership(ctx, owner, candidateOwner)
	if err != nil {
		t.Errorf("TransferOwnership should pass, got error %v", err)
	}

	// Make sure only the candidate owner can accept the ownership.
	err = poaKeeper.AcceptOwnership(ctx, owner)
	expectedErr := errorsmod.Wrapf(
		sdkerrors.ErrUnauthorized,
		"not the candidate owner; expected %s, sender %s",
		candidateOwner.String(),
		owner.String(),
	)
	if err.Error() != expectedErr.Error() {
		t.Errorf(
			"AcceptOwnership with wrong sender, error should be %v, got %v",
			expectedErr.Error(),
			err.Error(),
		)
	}

	// Finalize the ownership transfer.
	err = poaKeeper.AcceptOwnership(ctx, candidateOwner)
	if err != nil {
		t.Errorf("AcceptOwnership should pass, got error %v", err)
	}
	currentOwner := poaKeeper.GetOwner(ctx)
	if !currentOwner.Equals(candidateOwner) {
		t.Errorf("AcceptOwnership should properly change the owner")
	}
	currentCandidateOwner := poaKeeper.GetCandidateOwner(ctx)
	if !currentCandidateOwner.Empty() {
		t.Errorf("TransferOwnership should reset the candidate owner")
	}
}