package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/mezo-org/mezod/x/bridge/types"
)

// SetBridgeOutPaused sets or removes the bridge-out paused flag. Pausing does
// not change the outflow limits.
func (k Keeper) SetBridgeOutPaused(ctx sdk.Context, isPaused bool) {
	store := ctx.KVStore(k.storeKey)

	if isPaused {
		store.Set(types.BridgeOutPausedKey, []byte{0x01})
	} else {
		store.Delete(types.BridgeOutPausedKey)
	}
}

// IsBridgeOutPaused checks if bridging out is paused.
func (k Keeper) IsBridgeOutPaused(ctx sdk.Context) bool {
	return ctx.KVStore(k.storeKey).Has(types.BridgeOutPausedKey)
}

// SetBridgeInPaused sets or removes the bridge-in paused flag. For now the flag
// stops the triparty inbound path only; the validator-observed path has no
// pause switch yet.
func (k Keeper) SetBridgeInPaused(ctx sdk.Context, isPaused bool) {
	store := ctx.KVStore(k.storeKey)

	if isPaused {
		store.Set(types.BridgeInPausedKey, []byte{0x01})
	} else {
		store.Delete(types.BridgeInPausedKey)
	}
}

// IsBridgeInPaused checks if bridging in is paused.
func (k Keeper) IsBridgeInPaused(ctx sdk.Context) bool {
	return ctx.KVStore(k.storeKey).Has(types.BridgeInPausedKey)
}
