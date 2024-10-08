package keeper

import sdk "github.com/cosmos/cosmos-sdk/types"

func (k Keeper) AddPrivilege(
	ctx sdk.Context,
	sender sdk.AccAddress,
	operators []sdk.ValAddress,
	privilege string,
) error {
	panic("implement me")
}
