package keeper

import (
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/mezo-org/mezod/x/bridge/types"
)

// OutflowResetBlocks is the number of blocks after which the outflow limit is reset.
const OutflowResetBlocks = 25000

// SetOutflowLimit sets the maximum outflow limit for a specific token.
func (k Keeper) SetOutflowLimit(
	ctx sdk.Context,
	token []byte,
	limit math.Int,
) {
	store := ctx.KVStore(k.storeKey)

	bz, err := limit.Marshal()
	if err != nil {
		panic(err)
	}

	store.Set(types.GetOutflowLimitKey(token), bz)
}

// GetOutflowLimit returns the current outflow limit for a specific token.
func (k Keeper) GetOutflowLimit(
	ctx sdk.Context,
	token []byte,
) math.Int {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.GetOutflowLimitKey(token))
	if len(bz) == 0 {
		return math.ZeroInt()
	}

	limit := math.ZeroInt()
	if err := limit.Unmarshal(bz); err != nil {
		panic(err)
	}

	return limit
}

// GetCurrentOutflow returns the current outflow amount for a specific token.
func (k Keeper) GetCurrentOutflow(
	ctx sdk.Context,
	token []byte,
) math.Int {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.GetCurrentOutflowKey(token))
	if len(bz) == 0 {
		return math.ZeroInt()
	}

	outflow := math.ZeroInt()
	if err := outflow.Unmarshal(bz); err != nil {
		panic(err)
	}

	return outflow
}

// increaseCurrentOutflow adds the specified amount to the current outflow for a token.
func (k Keeper) increaseCurrentOutflow(
	ctx sdk.Context,
	token []byte,
	amount math.Int,
) error {
	currentOutflow := k.GetCurrentOutflow(ctx, token)
	newOutflow := currentOutflow.Add(amount)

	store := ctx.KVStore(k.storeKey)

	bz, err := newOutflow.Marshal()
	if err != nil {
		panic(err)
	}

	store.Set(types.GetCurrentOutflowKey(token), bz)

	return nil
}

// checkOutflowLimit verifies if adding the amount would exceed the outflow limit for a token.
func (k Keeper) checkOutflowLimit(
	ctx sdk.Context,
	token []byte,
	amount math.Int,
) error {
	capacity, _ := k.GetOutflowCapacity(ctx, token)

	if amount.GT(capacity) {
		return types.ErrOutflowLimitExceeded
	}

	return nil
}

// getLastOutflowReset returns the block height of the last outflow reset.
func (k Keeper) getLastOutflowReset(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	return sdk.BigEndianToUint64(store.Get(types.LastOutflowResetKey))
}

// setLastOutflowReset sets the block height of the last outflow reset.
func (k Keeper) setLastOutflowReset(ctx sdk.Context, height uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.LastOutflowResetKey, sdk.Uint64ToBigEndian(height))
}

// resetAllOutflows clears all current outflow amounts.
func (k Keeper) resetAllOutflows(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)

	iterator := storetypes.KVStorePrefixIterator(store, types.CurrentOutflowKeyPrefix)
	defer func() {
		_ = iterator.Close()
	}()

	keys := make([][]byte, 0)
	for ; iterator.Valid(); iterator.Next() {
		keys = append(keys, iterator.Key())
	}

	// separate the deletion from the iteration
	for _, key := range keys {
		store.Delete(key)
	}
}

// GetOutflowCapacity returns the outflow capacity for a specific token
// and the capacity reset block.
func (k Keeper) GetOutflowCapacity(
	ctx sdk.Context,
	token []byte,
) (capacity math.Int, resetHeight uint64) {
	limit := k.GetOutflowLimit(ctx, token)
	current := k.GetCurrentOutflow(ctx, token)
	lastReset := k.getLastOutflowReset(ctx)

	// Calculate outflow capacity (limit - current)
	capacity = limit.Sub(current)
	// Should never happen, but just in case.
	if capacity.IsNegative() {
		capacity = math.ZeroInt()
	}

	// Calculate capacity reset block (last reset + reset blocks)
	resetHeight = lastReset + OutflowResetBlocks

	return capacity, resetHeight
}
