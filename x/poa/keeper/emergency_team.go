package keeper

import (
	"bytes"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/mezo-org/mezod/x/poa/types"
)

// GetEmergencyTeam returns the emergency team address. The returned address is
// empty if the role is not granted.
func (k Keeper) GetEmergencyTeam(ctx sdk.Context) sdk.AccAddress {
	return ctx.KVStore(k.storeKey).Get(types.EmergencyTeamKey)
}

// SetEmergencyTeam grants the emergency team role to the given address. Passing
// an empty or zero address revokes the role. The grant is a single step, unlike
// the ownership transfer, because the role must be quick to rotate.
//
// The function returns an error if the sender is not the current owner.
// Returns nil if the role is set successfully.
//
// Upstream is responsible for setting the `sender` parameter to the actual
// actor performing the operation. If the sender address is empty, the function
// will return an error.
func (k Keeper) SetEmergencyTeam(
	ctx sdk.Context,
	sender sdk.AccAddress,
	emergencyTeam sdk.AccAddress,
) error {
	if err := k.checkOwner(ctx, sender); err != nil {
		return err
	}

	k.SetEmergencyTeamUnchecked(ctx, emergencyTeam)

	return nil
}

// SetEmergencyTeamUnchecked sets the emergency team address without any
// authorization check. It is reserved for genesis and upgrade handlers.
func (k Keeper) SetEmergencyTeamUnchecked(
	ctx sdk.Context,
	emergencyTeam sdk.AccAddress,
) {
	store := ctx.KVStore(k.storeKey)

	// The maintenance precompile passes the 20-byte zero address to revoke the
	// role, as Solidity has no empty address.
	if emergencyTeam.Empty() || bytes.Equal(emergencyTeam, make([]byte, 20)) {
		store.Delete(types.EmergencyTeamKey)
		return
	}

	store.Set(types.EmergencyTeamKey, emergencyTeam)
}

// CheckOwnerOrEmergencyTeam checks if the sender is the validator pool owner or
// the emergency team. This is the gate for the emergency actions both parties
// can perform.
// Returns an error if the sender is neither of them or either of the compared
// addresses is empty. Returns nil otherwise.
func (k Keeper) CheckOwnerOrEmergencyTeam(
	ctx sdk.Context,
	sender sdk.AccAddress,
) error {
	if err := k.checkOwner(ctx, sender); err == nil {
		return nil
	}

	return checkAccount(sender, k.GetEmergencyTeam(ctx), "emergency team")
}
