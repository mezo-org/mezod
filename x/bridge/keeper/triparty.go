package keeper

import (
	"fmt"

	sdkerrors "cosmossdk.io/errors"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/mezo-org/mezod/x/bridge/types"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
)

// TripartyWindowResetBlocks is the number of blocks after which the
// triparty minting window is reset.
const TripartyWindowResetBlocks = 25000

// TripartyBatch is the maximum number of triparty bridge requests
// processed per block. While triparty mints are expected to be rare,
// capping the batch size provides defense in depth to ensure stable
// block times.
const TripartyBatch = 5

// MinTripartyAmount is the minimum amount for a triparty bridge request
// (0.01 BTC in 18-decimal precision). This hard-coded floor prevents
// a compromised controller from spamming the chain with many small requests.
var MinTripartyAmount = math.NewInt(10_000_000_000_000_000)

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

// getAllAllowedTripartyControllers returns all allowed triparty controllers.
func (k Keeper) getAllAllowedTripartyControllers(ctx sdk.Context) []string {
	store := ctx.KVStore(k.storeKey)

	iterator := storetypes.KVStorePrefixIterator(
		store, types.TripartyControllerKeyPrefix,
	)
	defer func() {
		_ = iterator.Close()
	}()

	var out []string

	for ; iterator.Valid(); iterator.Next() {
		controller := iterator.Key()[len(types.TripartyControllerKeyPrefix):]
		out = append(out, evmtypes.BytesToHexAddress(controller))
	}

	return out
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
// If not set, it returns the default value of 1. The return type is
// int64 to match block heights (int64 in Cosmos SDK), so the delay
// can be used directly in block height arithmetic without casting.
func (k Keeper) GetTripartyBlockDelay(ctx sdk.Context) int64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.TripartyBlockDelayKey)
	if len(bz) == 0 {
		return 1
	}
	// The stored value is int64, as accepted by SetTripartyBlockDelay.
	return int64(sdk.BigEndianToUint64(bz)) //nolint:gosec
}

// SetTripartyBlockDelay sets the triparty block delay. The delay is
// int64 to match block heights in the Cosmos SDK. The delay must be
// at least 1; otherwise an error is returned.
func (k Keeper) SetTripartyBlockDelay(ctx sdk.Context, delay int64) error {
	if delay < 1 {
		return fmt.Errorf("delay must not be less than 1")
	}
	store := ctx.KVStore(k.storeKey)
	store.Set(types.TripartyBlockDelayKey, sdk.Uint64ToBigEndian(uint64(delay))) //nolint:gosec
	return nil
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

// GetTripartyWindowLimit returns the triparty request window limit.
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

// SetTripartyWindowLimit sets the triparty request window limit.
func (k Keeper) SetTripartyWindowLimit(ctx sdk.Context, limit math.Int) {
	store := ctx.KVStore(k.storeKey)

	bz, err := limit.Marshal()
	if err != nil {
		panic(err)
	}

	store.Set(types.TripartyWindowLimitKey, bz)
}

// GetTripartyRequestSequenceTip returns the last assigned triparty request
// sequence number. Returns 0 if not set.
func (k Keeper) GetTripartyRequestSequenceTip(ctx sdk.Context) math.Int {
	bz := ctx.KVStore(k.storeKey).Get(types.TripartyRequestSequenceTipKey)
	if len(bz) == 0 {
		return math.ZeroInt()
	}

	tip := math.ZeroInt()
	if err := tip.Unmarshal(bz); err != nil {
		panic(err)
	}

	return tip
}

// incrementTripartyRequestSequenceTip advances the triparty request
// sequence tip by one and returns the new value. The returned value is
// used as the unique identifier (sequence number) for a new incoming
// triparty bridge request.
func (k Keeper) incrementTripartyRequestSequenceTip(ctx sdk.Context) math.Int {
	tip := k.GetTripartyRequestSequenceTip(ctx).AddRaw(1)

	bz, err := tip.Marshal()
	if err != nil {
		panic(err)
	}

	ctx.KVStore(k.storeKey).Set(types.TripartyRequestSequenceTipKey, bz)

	return tip
}

// setTripartyRequestSequenceTip sets the last assigned triparty request
// sequence number.
func (k Keeper) setTripartyRequestSequenceTip(ctx sdk.Context, tip math.Int) {
	bz, err := tip.Marshal()
	if err != nil {
		panic(err)
	}

	ctx.KVStore(k.storeKey).Set(types.TripartyRequestSequenceTipKey, bz)
}

// GetTripartyProcessedSequenceTip returns the last processed triparty
// request sequence number. Returns 0 if not set.
func (k Keeper) GetTripartyProcessedSequenceTip(ctx sdk.Context) math.Int {
	bz := ctx.KVStore(k.storeKey).Get(types.TripartyProcessedSequenceTipKey)
	if len(bz) == 0 {
		return math.ZeroInt()
	}

	tip := math.ZeroInt()
	if err := tip.Unmarshal(bz); err != nil {
		panic(err)
	}

	return tip
}

// setTripartyProcessedSequenceTip sets the last processed triparty
// request sequence number.
func (k Keeper) setTripartyProcessedSequenceTip(ctx sdk.Context, tip math.Int) {
	bz, err := tip.Marshal()
	if err != nil {
		panic(err)
	}

	ctx.KVStore(k.storeKey).Set(types.TripartyProcessedSequenceTipKey, bz)
}

// validateTripartyBridgeRequest validates a triparty bridge request. It
// checks all inputs and state-dependent conditions:
//   - the recipient is a valid, non-zero hex address,
//   - the recipient is not a blocked address (e.g. module account),
//   - the recipient is not a custom precompile address,
//   - the controller is a valid hex address,
//   - the controller is an allowed triparty controller,
//   - the callback data does not exceed the maximum length,
//   - the amount is positive,
//   - the amount is at least the minimum triparty amount,
//   - the amount does not exceed the per-request limit.
//
// This function is called both at request creation time and at processing
// time in the PreBlocker. The PreBlocker repeats these checks because
// conditions may change between request creation and processing (e.g. a
// recipient may become blocked or a controller may be deauthorized).
func (k Keeper) validateTripartyBridgeRequest(
	ctx sdk.Context,
	recipient string,
	amount math.Int,
	callbackData []byte,
	controller string,
) error {
	if !evmtypes.IsHexAddress(recipient) {
		return sdkerrors.Wrap(types.ErrInvalidEVMAddress, "invalid recipient")
	}
	if evmtypes.IsZeroHexAddress(recipient) {
		return sdkerrors.Wrap(types.ErrZeroEVMAddress, "zero recipient")
	}

	recipientBytes := evmtypes.HexAddressToBytes(recipient)
	recipientAddr := sdk.AccAddress(recipientBytes)
	if _, blocked := k.blockedAddrs[recipientAddr.String()]; blocked {
		return types.ErrTripartyRecipientBlocked
	}

	if k.evmKeeper.IsCustomPrecompileAddress(recipient) {
		return types.ErrTripartyRecipientIsPrecompile
	}

	if !evmtypes.IsHexAddress(controller) {
		return sdkerrors.Wrap(types.ErrInvalidEVMAddress, "invalid controller")
	}

	if !k.IsAllowedTripartyController(ctx, evmtypes.HexAddressToBytes(controller)) {
		return types.ErrTripartyControllerNotAllowed
	}

	if len(callbackData) > types.MaxTripartyCallbackDataLength {
		return types.ErrTripartyCallbackDataTooLarge
	}

	if !amount.IsPositive() {
		return types.ErrTripartyAmountNotPositive
	}

	if amount.LT(MinTripartyAmount) {
		return types.ErrTripartyAmountBelowMinimum
	}

	perRequestLimit := k.GetTripartyPerRequestLimit(ctx)
	if perRequestLimit.IsPositive() && amount.GT(perRequestLimit) {
		return types.ErrTripartyPerRequestLimitExceeded
	}

	return nil
}

// CreateTripartyBridgeRequest creates a new pending triparty bridge request,
// assigns it the next sequence number, records the current block height,
// stores it in state, and returns the assigned requestId. It returns an
// error if validation fails.
func (k Keeper) CreateTripartyBridgeRequest(
	ctx sdk.Context,
	recipient string,
	amount math.Int,
	callbackData []byte,
	controller string,
) (math.Int, error) {
	if k.IsTripartyPaused(ctx) {
		return math.Int{}, types.ErrTripartyPaused
	}

	if err := k.validateTripartyBridgeRequest(
		ctx,
		recipient,
		amount,
		callbackData,
		controller,
	); err != nil {
		return math.Int{}, err
	}

	if err := k.checkTripartyCapacity(ctx, amount); err != nil {
		return math.Int{}, err
	}

	seq := k.incrementTripartyRequestSequenceTip(ctx)

	req := &types.TripartyBridgeRequest{
		Sequence:     seq,
		BlockHeight:  ctx.BlockHeight(),
		Recipient:    recipient,
		Amount:       amount,
		CallbackData: callbackData,
		Controller:   controller,
	}

	k.saveTripartyBridgeRequest(ctx, req)
	k.increaseTripartyWindowConsumed(ctx, amount)

	return seq, nil
}

func (k Keeper) saveTripartyBridgeRequest(
	ctx sdk.Context,
	req *types.TripartyBridgeRequest,
) {
	bz, err := req.Marshal()
	if err != nil {
		panic(err)
	}

	ctx.KVStore(k.storeKey).Set(
		types.GetTripartyBridgeRequestKey(req.Sequence),
		bz,
	)
}

// getTripartyBridgeRequest returns a pending triparty bridge request by its
// sequence number. Returns nil and false if the request does not exist.
func (k Keeper) getTripartyBridgeRequest(
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

// getAllPendingTripartyBridgeRequests returns all pending triparty requests.
func (k Keeper) getAllPendingTripartyBridgeRequests(
	ctx sdk.Context,
) []*types.TripartyBridgeRequest {
	store := ctx.KVStore(k.storeKey)

	iterator := storetypes.KVStorePrefixIterator(
		store, types.TripartyRequestKeyPrefix,
	)
	defer func() {
		_ = iterator.Close()
	}()

	var out []*types.TripartyBridgeRequest

	for ; iterator.Valid(); iterator.Next() {
		req := &types.TripartyBridgeRequest{}
		if err := req.Unmarshal(iterator.Value()); err != nil {
			panic(err)
		}

		out = append(out, req)
	}

	return out
}

// deleteTripartyBridgeRequest removes a pending triparty bridge request
// from state. It enforces sequential deletion: if a request with a lower
// sequence number exists, the deletion is rejected to prevent gaps.
func (k Keeper) deleteTripartyBridgeRequest(ctx sdk.Context, sequence math.Int) error {
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

// getTripartyWindowConsumed returns the current triparty window consumed
// aggregate. Returns zero if not set.
func (k Keeper) getTripartyWindowConsumed(ctx sdk.Context) math.Int {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.TripartyWindowConsumedKey)
	if len(bz) == 0 {
		return math.ZeroInt()
	}

	consumed := math.ZeroInt()
	if err := consumed.Unmarshal(bz); err != nil {
		panic(err)
	}

	return consumed
}

// increaseTripartyWindowConsumed adds the given amount to the current
// triparty window consumed aggregate.
func (k Keeper) increaseTripartyWindowConsumed(ctx sdk.Context, amount math.Int) {
	consumed := k.getTripartyWindowConsumed(ctx).Add(amount)

	store := ctx.KVStore(k.storeKey)

	bz, err := consumed.Marshal()
	if err != nil {
		panic(err)
	}

	store.Set(types.TripartyWindowConsumedKey, bz)
}

// resetTripartyWindowConsumed clears the current triparty window consumed
// aggregate to zero and records the current block height as the last
// reset point.
func (k Keeper) resetTripartyWindowConsumed(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.TripartyWindowConsumedKey)
	store.Set(
		types.TripartyWindowLastResetKey,
		// block height can't be negative so int64->uint64 conversion is safe
		sdk.Uint64ToBigEndian(uint64(ctx.BlockHeight())), //nolint:gosec
	)
}

// setTripartyWindowLastReset sets the block height of the last triparty
// window reset.
func (k Keeper) setTripartyWindowLastReset(ctx sdk.Context, height uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.TripartyWindowLastResetKey, sdk.Uint64ToBigEndian(height))
}

// getTripartyWindowLastReset returns the block height at which the
// triparty minting window was last reset. Returns 0 if not set.
func (k Keeper) getTripartyWindowLastReset(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	return sdk.BigEndianToUint64(store.Get(types.TripartyWindowLastResetKey))
}

// GetTripartyCapacity returns the remaining triparty request window capacity
// within the current window and the block height at which the window
// resets.
func (k Keeper) GetTripartyCapacity(ctx sdk.Context) (capacity math.Int, resetHeight uint64) {
	limit := k.GetTripartyWindowLimit(ctx)
	lastReset := k.getTripartyWindowLastReset(ctx)
	consumed := k.getTripartyWindowConsumed(ctx)

	capacity = limit.Sub(consumed)
	if capacity.IsNegative() {
		capacity = math.ZeroInt()
	}

	resetHeight = lastReset + TripartyWindowResetBlocks

	return capacity, resetHeight
}

// checkTripartyCapacity returns an error if the given amount exceeds the
// remaining triparty request window capacity within the current window.
func (k Keeper) checkTripartyCapacity(ctx sdk.Context, amount math.Int) error {
	capacity, _ := k.GetTripartyCapacity(ctx)

	if amount.GT(capacity) {
		return types.ErrTripartyWindowLimitExceeded
	}

	return nil
}

// GetTripartyTotalBTCMinted returns the total BTC minted via triparty
// bridging, derived by summing all per-controller provenance counters.
func (k Keeper) GetTripartyTotalBTCMinted(ctx sdk.Context) math.Int {
	total := math.ZeroInt()
	for _, entry := range k.getAllTripartyControllerBTCMinted(ctx) {
		total = total.Add(entry.Amount)
	}
	return total
}

// GetTripartyControllerBTCMinted returns the total BTC minted via triparty
// bridging by the given controller. Returns zero if not set.
func (k Keeper) GetTripartyControllerBTCMinted(
	ctx sdk.Context,
	controller string,
) math.Int {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.GetTripartyControllerBTCMintedKey(
		evmtypes.HexAddressToBytes(controller),
	))
	if len(bz) == 0 {
		return math.ZeroInt()
	}

	total := math.ZeroInt()
	if err := total.Unmarshal(bz); err != nil {
		panic(err)
	}

	return total
}

// increaseTripartyControllerBTCMinted adds the given amount to the BTC
// minted via triparty bridging by the given controller.
func (k Keeper) increaseTripartyControllerBTCMinted(
	ctx sdk.Context,
	controller string,
	amount math.Int,
) {
	total := k.GetTripartyControllerBTCMinted(ctx, controller).Add(amount)

	controllerBytes := evmtypes.HexAddressToBytes(controller)
	store := ctx.KVStore(k.storeKey)

	bz, err := total.Marshal()
	if err != nil {
		panic(err)
	}

	store.Set(types.GetTripartyControllerBTCMintedKey(controllerBytes), bz)
}

// getAllTripartyControllerBTCMinted returns the per-controller BTC minted
// amounts for all controllers that have minted through the triparty path.
func (k Keeper) getAllTripartyControllerBTCMinted(
	ctx sdk.Context,
) []*types.TripartyControllerBTCMinted {
	store := ctx.KVStore(k.storeKey)

	iterator := storetypes.KVStorePrefixIterator(
		store,
		types.TripartyControllerBTCMintedKeyPrefix,
	)
	defer func() {
		_ = iterator.Close()
	}()

	var out []*types.TripartyControllerBTCMinted

	for ; iterator.Valid(); iterator.Next() {
		controller := iterator.Key()[len(types.TripartyControllerBTCMintedKeyPrefix):]

		var amount math.Int
		err := amount.Unmarshal(iterator.Value())
		if err != nil {
			panic(err)
		}

		out = append(
			out,
			&types.TripartyControllerBTCMinted{
				Controller: evmtypes.BytesToHexAddress(controller),
				Amount:     amount,
			},
		)
	}

	return out
}

// ProcessTripartyBridgeRequests reads pending triparty bridge requests
// from state and processes up to TripartyBatch mature requests by
// minting BTC and issuing callbacks to controllers.
//
// Requests are processed in strictly increasing sequence order. Processing
// stops at the first immature request to preserve ordering guarantees.
// Invalid requests (blocked recipient, deauthorized controller, limit
// exceeded) are skipped and deleted.
//
// A callback failure is logged but does not prevent the mint from
// completing or block subsequent requests. A mintBTC failure is fatal
// and returns an error that will cause a consensus failure.
func (k Keeper) ProcessTripartyBridgeRequests(ctx sdk.Context) error {
	if k.IsTripartyPaused(ctx) {
		k.Logger(ctx).Info("triparty bridging is paused; skipping processing")
		return nil
	}

	blockDelay := k.GetTripartyBlockDelay(ctx)

	seq := k.GetTripartyProcessedSequenceTip(ctx).AddRaw(1)

	for range TripartyBatch {
		req, found := k.getTripartyBridgeRequest(ctx, seq)
		if !found {
			break
		}

		// Stop at the first immature request. No request can be processed
		// ahead of an earlier one that is not yet mature.
		if ctx.BlockHeight() < req.BlockHeight+blockDelay {
			k.Logger(ctx).Info(
				"triparty request not yet mature; stopping processing",
				"sequence", req.Sequence,
				"requestHeight", req.BlockHeight,
				"currentHeight", ctx.BlockHeight(),
				"blockDelay", blockDelay,
			)

			break
		}

		// Defense in depth: re-validate the request. Conditions may have
		// changed between request creation and processing (e.g. a
		// recipient may have become blocked, a controller deauthorized,
		// or the per-request limit lowered).
		if err := k.validateTripartyBridgeRequest(ctx, req.Recipient, req.Amount, req.CallbackData, req.Controller); err != nil {
			k.Logger(ctx).Warn(
				"triparty request failed validation; "+
					"request skipped",
				"sequence", req.Sequence,
				"error", err,
			)
		} else {
			// Mint BTC. A failure here is a system error (x/bank failure)
			// and causes a consensus failure, same as the AssetsLocked
			// path.
			recipientBytes := evmtypes.HexAddressToBytes(req.Recipient)
			recipientAddr := sdk.AccAddress(recipientBytes)
			if err := k.mintBTC(ctx, recipientAddr, req.Amount); err != nil {
				return fmt.Errorf(
					"failed to mint BTC for triparty request %s: %w",
					req.Sequence, err,
				)
			}

			// Update the per-controller provenance counter.
			k.increaseTripartyControllerBTCMinted(ctx, req.Controller, req.Amount)

			// Issue the EVM callback to the controller. A callback
			// failure is logged but must not prevent the mint from
			// completing or block subsequent requests.
			k.issueTripartyCallback(ctx, req)

			k.Logger(ctx).Info(
				"triparty bridge request processed",
				"sequence", req.Sequence,
				"recipient", req.Recipient,
				"amount", req.Amount,
				"controller", req.Controller,
			)
		}

		// Delete the request from state regardless of whether it was
		// processed or skipped. A failure here is fatal because leaving
		// a processed request would cause double-minting on the next
		// block.
		if err := k.deleteTripartyBridgeRequest(ctx, req.Sequence); err != nil {
			return fmt.Errorf(
				"failed to delete triparty request %s: %w",
				req.Sequence, err,
			)
		}

		k.setTripartyProcessedSequenceTip(ctx, seq)
		seq = seq.AddRaw(1)
	}

	return nil
}

// issueTripartyCallback issues an EVM callback to the controller that
// submitted a triparty bridge request. Failures are logged but do not
// cause errors — the BTC has already been minted and cannot be rolled
// back without risking a supply invariant violation.
func (k Keeper) issueTripartyCallback(
	ctx sdk.Context,
	req *types.TripartyBridgeRequest,
) {
	controllerBytes := evmtypes.HexAddressToBytes(req.Controller)

	recipientBytes := evmtypes.HexAddressToBytes(req.Recipient)

	call, err := evmtypes.NewTripartyCallbackCall(
		authtypes.NewModuleAddress(types.ModuleName).Bytes(),
		controllerBytes,
		req.Sequence.BigInt(),
		recipientBytes,
		req.Amount.BigInt(),
		req.CallbackData,
	)
	if err != nil {
		k.Logger(ctx).Warn(
			"failed to create triparty callback call; mint completed but callback skipped",
			"sequence", req.Sequence,
			"error", err,
		)

		return
	}

	_, _, err = k.evmKeeper.ExecuteContractCall(ctx, call)
	if err != nil {
		k.Logger(ctx).Warn(
			"triparty callback failed; mint completed but callback skipped",
			"sequence", req.Sequence,
			"controller", req.Controller,
			"error", err,
		)
	}
}
