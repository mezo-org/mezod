package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v12/x/dualstaking/types"
)

// GetParams get all parameters as types.Params
func (k Keeper) GetParams(_ sdk.Context) types.Params {
	return types.NewParams()
}

// SetParams set the params
func (k Keeper) SetParams(_ sdk.Context, _ types.Params) {
}
