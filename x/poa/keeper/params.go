package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v12/x/poa/types"
)

// MaxValidators - Maximum number of validators
func (k Keeper) MaxValidators(ctx sdk.Context) (res uint16) {
	k.paramspace.Get(ctx, types.KeyMaxValidators, &res)
	return
}

// Quorum - Quorum percentage to add or remove validators
func (k Keeper) Quorum(ctx sdk.Context) (res uint16) {
	k.paramspace.Get(ctx, types.KeyQuorum, &res)
	return
}

// GetParams returns the total set of poa parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramspace.GetParamSet(ctx, &params)
	return params
}

// SetParams sets the poa parameters to the param space.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramspace.SetParamSet(ctx, &params)
}
