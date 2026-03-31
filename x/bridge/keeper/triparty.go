package keeper

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/mezo-org/mezod/x/bridge/types"
)

// TripartyWindowResetBlocks is the number of blocks after which the
// triparty minting window is reset.
const TripartyWindowResetBlocks = 25000

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
	if len(bz) == 0 {
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

// GetTripartySequenceTip returns the last assigned triparty request
// sequence number. Returns 0 if not set.
func (k Keeper) GetTripartySequenceTip(ctx sdk.Context) math.Int {
	bz := ctx.KVStore(k.storeKey).Get(types.TripartySequenceTipKey)

	var tip math.Int
	if err := tip.Unmarshal(bz); err != nil {
		panic(err)
	}

	if tip.IsNil() {
		tip = math.ZeroInt()
	}

	return tip
}

// incrementTripartySequenceTip advances the triparty sequence tip by one
// and returns the new value.
func (k Keeper) incrementTripartySequenceTip(ctx sdk.Context) math.Int {
	tip := k.GetTripartySequenceTip(ctx).AddRaw(1)

	bz, err := tip.Marshal()
	if err != nil {
		panic(err)
	}

	ctx.KVStore(k.storeKey).Set(types.TripartySequenceTipKey, bz)

	return tip
}

// CreateTripartyBridgeRequest creates a new pending triparty bridge request,
// assigns it the next sequence number, records the current block height,
// stores it in state, and returns the assigned requestId.
func (k Keeper) CreateTripartyBridgeRequest(
	ctx sdk.Context,
	recipient string,
	amount math.Int,
	callbackData []byte,
	controller string,
) math.Int {
	seq := k.incrementTripartySequenceTip(ctx)

	req := &types.TripartyBridgeRequest{
		Sequence:     seq,
		BlockHeight:  ctx.BlockHeight(),
		Recipient:    recipient,
		Amount:       amount,
		CallbackData: callbackData,
		Controller:   controller,
	}

	bz, err := req.Marshal()
	if err != nil {
		panic(err)
	}

	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetTripartyBridgeRequestKey(seq), bz)

	return seq
}

// GetTripartyBridgeRequest returns a pending triparty bridge request by its
// sequence number. Returns nil and false if the request does not exist.
func (k Keeper) GetTripartyBridgeRequest(
	ctx sdk.Context,
	sequence math.Int,
) (*types.TripartyBridgeRequest, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetTripartyBridgeRequestKey(sequence))
	if len(bz) == 0 {
		return nil, false
	}

	req := &types.TripartyBridgeRequest{}
	if err := req.Unmarshal(bz); err != nil {
		panic(err)
	}

	return req, true
}

// DeleteTripartyBridgeRequest removes a pending triparty bridge request
// from state. It enforces sequential deletion: if a request with a lower
// sequence number exists, the deletion is rejected to prevent gaps.
func (k Keeper) DeleteTripartyBridgeRequest(ctx sdk.Context, sequence math.Int) error {
	store := ctx.KVStore(k.storeKey)

	prev := sequence.SubRaw(1)
	if prev.IsPositive() && store.Has(types.GetTripartyBridgeRequestKey(prev)) {
		return fmt.Errorf(
			"cannot delete triparty request %s: previous request %s still pending",
			sequence, prev,
		)
	}

	store.Delete(types.GetTripartyBridgeRequestKey(sequence))

	return nil
}

// GetPendingTripartyBridgeRequests returns up to `limit` pending triparty
// bridge requests starting from the given sequence number, in strictly
// increasing sequence order.
func (k Keeper) GetPendingTripartyBridgeRequests(
	ctx sdk.Context,
	startSequence math.Int,
	limit int,
) []*types.TripartyBridgeRequest {
	store := ctx.KVStore(k.storeKey)
	var requests []*types.TripartyBridgeRequest

	seq := startSequence
	for i := 0; i < limit; i++ {
		bz := store.Get(types.GetTripartyBridgeRequestKey(seq))
		if len(bz) == 0 {
			break
		}

		req := &types.TripartyBridgeRequest{}
		if err := req.Unmarshal(bz); err != nil {
			panic(err)
		}

		requests = append(requests, req)
		seq = seq.AddRaw(1)
	}

	return requests
}

// GetTripartyWindowMinted returns the current triparty window minted
// aggregate. Returns zero if not set.
func (k Keeper) GetTripartyWindowMinted(ctx sdk.Context) math.Int {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.TripartyWindowMintedKey)
	if len(bz) == 0 {
		return math.ZeroInt()
	}

	minted := math.ZeroInt()
	if err := minted.Unmarshal(bz); err != nil {
		panic(err)
	}

	return minted
}

// IncreaseTripartyWindowMinted adds the given amount to the current
// triparty window minted aggregate.
func (k Keeper) IncreaseTripartyWindowMinted(ctx sdk.Context, amount math.Int) {
	minted := k.GetTripartyWindowMinted(ctx).Add(amount)

	store := ctx.KVStore(k.storeKey)

	bz, err := minted.Marshal()
	if err != nil {
		panic(err)
	}

	store.Set(types.TripartyWindowMintedKey, bz)
}

// ResetTripartyWindowMinted clears the current triparty window minted
// aggregate to zero and records the current block height as the last
// reset point.
func (k Keeper) ResetTripartyWindowMinted(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.TripartyWindowMintedKey)
	store.Set(
		types.TripartyWindowLastResetKey,
		// block height can't be negative so int64->uint64 conversion is safe
		sdk.Uint64ToBigEndian(uint64(ctx.BlockHeight())), //nolint:gosec
	)
}

// GetTripartyWindowLastReset returns the block height at which the
// triparty minting window was last reset. Returns 0 if not set.
func (k Keeper) GetTripartyWindowLastReset(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	return sdk.BigEndianToUint64(store.Get(types.TripartyWindowLastResetKey))
}

// GetTripartyCapacity returns the remaining triparty minting capacity
// within the current window and the block height at which the window
// resets.
func (k Keeper) GetTripartyCapacity(ctx sdk.Context) (capacity math.Int, resetHeight uint64) {
	limit := k.GetTripartyWindowLimit(ctx)
	minted := k.GetTripartyWindowMinted(ctx)
	lastReset := k.GetTripartyWindowLastReset(ctx)

	capacity = limit.Sub(minted)
	if capacity.IsNegative() {
		capacity = math.ZeroInt()
	}

	resetHeight = lastReset + TripartyWindowResetBlocks

	return capacity, resetHeight
}

// CheckTripartyCapacity returns an error if the given amount exceeds the
// remaining triparty minting capacity within the current window.
func (k Keeper) CheckTripartyCapacity(ctx sdk.Context, amount math.Int) error {
	capacity, _ := k.GetTripartyCapacity(ctx)

	if amount.GT(capacity) {
		return types.ErrTripartyWindowLimitExceeded
	}

	return nil
}

// GetTripartyTotalBTCMinted returns the total BTC minted via triparty
// bridging. This is an informational provenance counter. Returns zero
// if not set.
func (k Keeper) GetTripartyTotalBTCMinted(ctx sdk.Context) math.Int {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.TripartyTotalBTCMintedKey)
	if len(bz) == 0 {
		return math.ZeroInt()
	}

	total := math.ZeroInt()
	if err := total.Unmarshal(bz); err != nil {
		panic(err)
	}

	return total
}

// IncreaseTripartyTotalBTCMinted adds the given amount to the total BTC
// minted via triparty bridging provenance counter.
func (k Keeper) IncreaseTripartyTotalBTCMinted(ctx sdk.Context, amount math.Int) {
	total := k.GetTripartyTotalBTCMinted(ctx).Add(amount)

	store := ctx.KVStore(k.storeKey)

	bz, err := total.Marshal()
	if err != nil {
		panic(err)
	}

	store.Set(types.TripartyTotalBTCMintedKey, bz)
}
