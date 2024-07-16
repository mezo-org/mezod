package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v12/x/poa/types"
)

// GetParams returns the total set of poa parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ParamsKey)
	if len(bz) == 0 {
		return params
	}

	k.cdc.MustUnmarshal(bz, &params)
	return params
}

// UpdateParams updates the poa module's parameters.
// Upstream is responsible for setting the `sender` parameter to the actual
// actor performing the operation.
func (k Keeper) UpdateParams(
	ctx sdk.Context,
	sender sdk.AccAddress,
	params types.Params,
) error {
	if err := k.CheckOwnership(ctx, sender); err != nil {
		return err
	}

	if err := params.Validate(); err != nil {
		return errorsmod.Wrapf(err, "invalid params")
	}

	return k.setParams(ctx, params)
}

// setParams sets the poa module's parameters.
func (k Keeper) setParams(
	ctx sdk.Context,
	params types.Params,
) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.Marshal(&params)
	if err != nil {
		return err
	}

	store.Set(types.ParamsKey, bz)

	return nil
}
