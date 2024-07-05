package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
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
func (k Keeper) UpdateParams(
	ctx sdk.Context,
	sender sdk.AccAddress,
	params types.Params,
) error {
	if k.authority.String() != sender.String() {
		return errorsmod.Wrapf(
			govtypes.ErrInvalidSigner,
			"invalid authority; expected %s, got %s",
			k.authority.String(),
			sender.String(),
		)
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
