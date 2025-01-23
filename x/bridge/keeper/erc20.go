package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) GetSupportedERC20Tokens(_ sdk.Context) map[string]string {
	// TODO: Implement GetSupportedERC20Tokens.
	panic("implement me")
}

func (k Keeper) setSupportedERC20Tokens(
	_ sdk.Context,
	_ map[string]string,
) {
	// TODO: Implement setSourceBTCToken.
	panic("implement me")
}
