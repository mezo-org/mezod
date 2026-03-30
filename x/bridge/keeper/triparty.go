package keeper

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/mezo-org/mezod/x/bridge/types"
)

// IsAllowedTripartyController checks if the given address is an allowed
// triparty controller.
func (k Keeper) IsAllowedTripartyController(ctx sdk.Context, controller []byte) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.GetTripartyControllerKey(controller))
}

// AllowTripartyController sets or removes the given address as an allowed
// triparty controller.
func (k Keeper) AllowTripartyController(ctx sdk.Context, controller []byte, isAllowed bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetTripartyControllerKey(controller)

	if isAllowed {
		store.Set(key, []byte{0x01})
	} else {
		store.Delete(key)
	}
}

// IsTripartyPaused checks if triparty bridging is paused.
func (k Keeper) IsTripartyPaused(ctx sdk.Context) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.TripartyPausedKey)
}

// SetTripartyPaused sets or removes the triparty paused flag.
func (k Keeper) SetTripartyPaused(ctx sdk.Context, isPaused bool) {
	store := ctx.KVStore(k.storeKey)

	if isPaused {
		store.Set(types.TripartyPausedKey, []byte{0x01})
	} else {
		store.Delete(types.TripartyPausedKey)
	}
}

// GetTripartyBlockDelay returns the configured triparty block delay.
// If not set, it returns the default value of 1.
func (k Keeper) GetTripartyBlockDelay(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.TripartyBlockDelayKey)
	if bz == nil {
		return 1
	}
	return sdk.BigEndianToUint64(bz)
}

// SetTripartyBlockDelay sets the triparty block delay.
func (k Keeper) SetTripartyBlockDelay(ctx sdk.Context, delay uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.TripartyBlockDelayKey, sdk.Uint64ToBigEndian(delay))
}

// GetTripartyPerRequestLimit returns the triparty per-request limit.
// Returns zero if not set.
func (k Keeper) GetTripartyPerRequestLimit(ctx sdk.Context) math.Int {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.TripartyPerRequestLimitKey)
	if len(bz) == 0 {
		return math.ZeroInt()
	}

	limit := math.ZeroInt()
	if err := limit.Unmarshal(bz); err != nil {
		panic(err)
	}

	return limit
}

// SetTripartyPerRequestLimit sets the triparty per-request limit.
func (k Keeper) SetTripartyPerRequestLimit(ctx sdk.Context, limit math.Int) {
	store := ctx.KVStore(k.storeKey)

	bz, err := limit.Marshal()
	if err != nil {
		panic(err)
	}

	store.Set(types.TripartyPerRequestLimitKey, bz)
}

// GetTripartyWindowLimit returns the triparty window limit.
// Returns zero if not set.
func (k Keeper) GetTripartyWindowLimit(ctx sdk.Context) math.Int {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.TripartyWindowLimitKey)
	if len(bz) == 0 {
		return math.ZeroInt()
	}

	limit := math.ZeroInt()
	if err := limit.Unmarshal(bz); err != nil {
		panic(err)
	}

	return limit
}

// SetTripartyWindowLimit sets the triparty window limit.
func (k Keeper) SetTripartyWindowLimit(ctx sdk.Context, limit math.Int) {
	store := ctx.KVStore(k.storeKey)

	bz, err := limit.Marshal()
	if err != nil {
		panic(err)
	}

	store.Set(types.TripartyWindowLimitKey, bz)
}
