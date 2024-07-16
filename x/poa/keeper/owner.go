package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/evmos/evmos/v12/x/poa/types"
)

// GetOwner returns the validator pool owner address.
func (k Keeper) GetOwner(ctx sdk.Context) sdk.AccAddress {
	return ctx.KVStore(k.storeKey).Get(types.OwnerKey)
}

// TransferOwnership transfers the validator pool ownership to a new address.
// Upstream is responsible for setting the `sender` parameter to the actual
// actor performing the operation.
func (k Keeper) TransferOwnership(
	ctx sdk.Context,
	sender sdk.AccAddress,
	newOwner sdk.AccAddress,
) error {
	if err := k.CheckOwnership(ctx, sender); err != nil {
		return err
	}

	k.setOwner(ctx, newOwner)

	return nil
}

// setOwner sets the validator pool owner address.
func (k Keeper) setOwner(ctx sdk.Context, owner sdk.AccAddress) {
	ctx.KVStore(k.storeKey).Set(types.OwnerKey, owner)
}

// CheckOwnership checks if the sender is the validator pool owner.
// Returns an error if the sender is not the owner. Returns nil otherwise.
func (k Keeper) CheckOwnership(ctx sdk.Context, sender sdk.AccAddress) error {
	ownerStr := k.GetOwner(ctx).String()
	senderStr := sender.String()

	if ownerStr != senderStr {
		return errorsmod.Wrapf(
			sdkerrors.ErrorInvalidSigner,
			"not the validator pool owner; expected %s, got %s",
			ownerStr,
			senderStr,
		)
	}

	return nil
}
