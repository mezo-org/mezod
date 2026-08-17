package keeper

import (
	"testing"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// zeroAccAddress is the 20-byte zero address. The maintenance precompile passes
// it to revoke the Emergency Team role.
var zeroAccAddress = sdk.AccAddress(make([]byte, 20))

func TestGetEmergencyTeam(t *testing.T) {
	ctx, poaKeeper := mockContext()

	emergencyTeam := poaKeeper.GetEmergencyTeam(ctx)
	if !emergencyTeam.Empty() {
		t.Errorf("GetEmergencyTeam should be empty when unset, got %v", emergencyTeam)
	}
}

func TestSetEmergencyTeam(t *testing.T) {
	ctx, poaKeeper := mockContext()

	// Generate an owner address using the mockValidator function.
	helper, _ := mockValidator()
	owner := sdk.AccAddress(helper.GetOperator())
	poaKeeper.setOwner(ctx, owner)

	// Generate an emergency team address using the mockValidator function.
	helper, _ = mockValidator()
	emergencyTeam := sdk.AccAddress(helper.GetOperator())

	// Make sure the function fails if the sender is not the owner.
	err := poaKeeper.SetEmergencyTeam(ctx, emergencyTeam, emergencyTeam)
	expectedErr := errorsmod.Wrapf(
		sdkerrors.ErrUnauthorized,
		"not the owner; expected %s, sender %s",
		owner.String(),
		emergencyTeam.String(),
	)
	if err.Error() != expectedErr.Error() {
		t.Errorf(
			"SetEmergencyTeam with wrong sender, error should be %v, got %v",
			expectedErr.Error(),
			err.Error(),
		)
	}
	if !poaKeeper.GetEmergencyTeam(ctx).Empty() {
		t.Errorf("SetEmergencyTeam with wrong sender should not set the emergency team")
	}

	err = poaKeeper.SetEmergencyTeam(ctx, owner, emergencyTeam)
	if err != nil {
		t.Errorf("SetEmergencyTeam should pass, got error %v", err)
	}
	currentEmergencyTeam := poaKeeper.GetEmergencyTeam(ctx)
	if !currentEmergencyTeam.Equals(emergencyTeam) {
		t.Errorf(
			"SetEmergencyTeam should properly set the emergency team, expected %v, got %v",
			emergencyTeam,
			currentEmergencyTeam,
		)
	}

	// Generate another emergency team address to make sure the role can be
	// rotated in a single call.
	helper, _ = mockValidator()
	newEmergencyTeam := sdk.AccAddress(helper.GetOperator())

	err = poaKeeper.SetEmergencyTeam(ctx, owner, newEmergencyTeam)
	if err != nil {
		t.Errorf("SetEmergencyTeam should pass, got error %v", err)
	}
	currentEmergencyTeam = poaKeeper.GetEmergencyTeam(ctx)
	if !currentEmergencyTeam.Equals(newEmergencyTeam) {
		t.Errorf(
			"SetEmergencyTeam should properly rotate the emergency team, expected %v, got %v",
			newEmergencyTeam,
			currentEmergencyTeam,
		)
	}
}

func TestSetEmergencyTeamRevoke(t *testing.T) {
	// Generate an owner address using the mockValidator function.
	helper, _ := mockValidator()
	owner := sdk.AccAddress(helper.GetOperator())

	// Generate an emergency team address using the mockValidator function.
	helper, _ = mockValidator()
	emergencyTeam := sdk.AccAddress(helper.GetOperator())

	tests := map[string]sdk.AccAddress{
		"zero address":  zeroAccAddress,
		"empty address": {},
	}

	for name, revokeAddress := range tests {
		t.Run(name, func(t *testing.T) {
			ctx, poaKeeper := mockContext()
			poaKeeper.setOwner(ctx, owner)

			err := poaKeeper.SetEmergencyTeam(ctx, owner, emergencyTeam)
			if err != nil {
				t.Errorf("SetEmergencyTeam should pass, got error %v", err)
			}

			err = poaKeeper.SetEmergencyTeam(ctx, owner, revokeAddress)
			if err != nil {
				t.Errorf("SetEmergencyTeam should pass, got error %v", err)
			}

			currentEmergencyTeam := poaKeeper.GetEmergencyTeam(ctx)
			if !currentEmergencyTeam.Empty() {
				t.Errorf(
					"SetEmergencyTeam with the %s should revoke the role, got %v",
					name,
					currentEmergencyTeam,
				)
			}
		})
	}
}

func TestSetEmergencyTeamUnchecked(t *testing.T) {
	ctx, poaKeeper := mockContext()

	// Generate an emergency team address using the mockValidator function.
	helper, _ := mockValidator()
	emergencyTeam := sdk.AccAddress(helper.GetOperator())

	// The unchecked setter is reserved for genesis and upgrade handlers so it
	// must work without an owner set in the store.
	poaKeeper.SetEmergencyTeamUnchecked(ctx, emergencyTeam)

	currentEmergencyTeam := poaKeeper.GetEmergencyTeam(ctx)
	if !currentEmergencyTeam.Equals(emergencyTeam) {
		t.Errorf(
			"SetEmergencyTeamUnchecked should properly set the emergency team, expected %v, got %v",
			emergencyTeam,
			currentEmergencyTeam,
		)
	}

	poaKeeper.SetEmergencyTeamUnchecked(ctx, zeroAccAddress)

	currentEmergencyTeam = poaKeeper.GetEmergencyTeam(ctx)
	if !currentEmergencyTeam.Empty() {
		t.Errorf(
			"SetEmergencyTeamUnchecked with the zero address should revoke the role, got %v",
			currentEmergencyTeam,
		)
	}
}

func TestCheckOwnerOrEmergencyTeam(t *testing.T) {
	ctx, poaKeeper := mockContext()

	// Generate an owner address using the mockValidator function.
	helper, _ := mockValidator()
	owner := sdk.AccAddress(helper.GetOperator())
	poaKeeper.setOwner(ctx, owner)

	// Generate an emergency team address using the mockValidator function.
	helper, _ = mockValidator()
	emergencyTeam := sdk.AccAddress(helper.GetOperator())

	// Generate a third party address using the mockValidator function.
	helper, _ = mockValidator()
	thirdParty := sdk.AccAddress(helper.GetOperator())

	// The owner passes the check even if the emergency team is not set.
	if err := poaKeeper.CheckOwnerOrEmergencyTeam(ctx, owner); err != nil {
		t.Errorf("CheckOwnerOrEmergencyTeam for the owner should pass, got error %v", err)
	}

	// The emergency team does not pass the check before the role is granted.
	err := poaKeeper.CheckOwnerOrEmergencyTeam(ctx, emergencyTeam)
	expectedErr := errorsmod.Wrap(
		sdkerrors.ErrInvalidAddress,
		"emergency team address is empty",
	)
	if err.Error() != expectedErr.Error() {
		t.Errorf(
			"CheckOwnerOrEmergencyTeam without the emergency team set, error should be %v, got %v",
			expectedErr.Error(),
			err.Error(),
		)
	}

	if err := poaKeeper.SetEmergencyTeam(ctx, owner, emergencyTeam); err != nil {
		t.Errorf("SetEmergencyTeam should pass, got error %v", err)
	}

	// Both the owner and the emergency team pass the check.
	if err := poaKeeper.CheckOwnerOrEmergencyTeam(ctx, owner); err != nil {
		t.Errorf("CheckOwnerOrEmergencyTeam for the owner should pass, got error %v", err)
	}
	if err := poaKeeper.CheckOwnerOrEmergencyTeam(ctx, emergencyTeam); err != nil {
		t.Errorf("CheckOwnerOrEmergencyTeam for the emergency team should pass, got error %v", err)
	}

	// A third party does not pass the check.
	err = poaKeeper.CheckOwnerOrEmergencyTeam(ctx, thirdParty)
	expectedErr = errorsmod.Wrapf(
		sdkerrors.ErrUnauthorized,
		"not the emergency team; expected %s, sender %s",
		emergencyTeam.String(),
		thirdParty.String(),
	)
	if err.Error() != expectedErr.Error() {
		t.Errorf(
			"CheckOwnerOrEmergencyTeam for a third party, error should be %v, got %v",
			expectedErr.Error(),
			err.Error(),
		)
	}

	// An empty sender does not pass the check.
	err = poaKeeper.CheckOwnerOrEmergencyTeam(ctx, sdk.AccAddress{})
	expectedErr = errorsmod.Wrap(
		sdkerrors.ErrInvalidAddress,
		"sender address is empty",
	)
	if err.Error() != expectedErr.Error() {
		t.Errorf(
			"CheckOwnerOrEmergencyTeam for an empty sender, error should be %v, got %v",
			expectedErr.Error(),
			err.Error(),
		)
	}

	// The emergency team does not pass the check after the role is revoked.
	if err := poaKeeper.SetEmergencyTeam(ctx, owner, zeroAccAddress); err != nil {
		t.Errorf("SetEmergencyTeam should pass, got error %v", err)
	}
	err = poaKeeper.CheckOwnerOrEmergencyTeam(ctx, emergencyTeam)
	expectedErr = errorsmod.Wrap(
		sdkerrors.ErrInvalidAddress,
		"emergency team address is empty",
	)
	if err.Error() != expectedErr.Error() {
		t.Errorf(
			"CheckOwnerOrEmergencyTeam after the role is revoked, error should be %v, got %v",
			expectedErr.Error(),
			err.Error(),
		)
	}
}
