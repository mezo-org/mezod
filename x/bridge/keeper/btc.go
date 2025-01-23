package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetSourceBTCToken returns the BTC token address on the source chain.
// AssetsLocked events carrying this token address are directly mapped to the
// Mezo native denomination - BTC.
func (k Keeper) GetSourceBTCToken(_ sdk.Context) string {
	// TODO: Implement GetSourceBTCToken.
	panic("implement me")
}

func (k Keeper) setSourceBTCToken(_ sdk.Context, _ string) {
	// TODO: Implement setSourceBTCToken.
	panic("implement me")
}
