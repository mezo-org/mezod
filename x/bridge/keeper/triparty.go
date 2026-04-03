package keeper

import (
	"fmt"

	sdkerrors "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/mezo-org/mezod/x/bridge/types"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
)

// TripartyWindowResetBlocks is the number of blocks after which the
// triparty minting window is reset.
const TripartyWindowResetBlocks = 25000

// MaxTripartyCallbackDataLength is the maximum allowed length of
// callbackData in a triparty bridge request (10 × 32-byte ABI words).
const MaxTripartyCallbackDataLength = 320

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

	if len(callbackData) > MaxTripartyCallbackDataLength {
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

	if err := k.validateTripartyBridgeRequest(ctx, recipient, amount, callbackData, controller); err != nil {
		return math.Int{}, err
	}

	if err := k.CheckTripartyCapacity(ctx, amount); err != nil {
		return math.Int{}, fmt.Errorf("triparty capacity check error: [%w]", err)
	}

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

	ctx.KVStore(k.storeKey).Set(types.GetTripartyBridgeRequestKey(seq), bz)
	k.IncreaseTripartyWindowMinted(ctx, amount)

	return seq, nil
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

// getTripartyProcessedSequenceTip returns the last processed triparty
// request sequence number. Returns 0 if not set.
func (k Keeper) getTripartyProcessedSequenceTip(ctx sdk.Context) math.Int {
	bz := ctx.KVStore(k.storeKey).Get(types.TripartyProcessedSequenceTipKey)

	var tip math.Int
	if err := tip.Unmarshal(bz); err != nil {
		panic(err)
	}

	if tip.IsNil() {
		tip = math.ZeroInt()
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

	seq := k.getTripartyProcessedSequenceTip(ctx).AddRaw(1)

	for range TripartyBatch {
		req, found := k.GetTripartyBridgeRequest(ctx, seq)
		if !found {
			break
		}

		// Stop at the first immature request. No request can be processed
		// ahead of an earlier one that is not yet mature.
		if ctx.BlockHeight() < req.BlockHeight+int64(blockDelay) { //nolint:gosec
			k.Logger(ctx).Info(
				"triparty request not yet mature; stopping batch",
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

			// Update the provenance counter.
			k.IncreaseTripartyTotalBTCMinted(ctx, req.Amount)

			// Issue the EVM callback to the controller. A callback
			// failure is logged but must not prevent the mint from
			// completing or block subsequent requests.
			controllerBytes := evmtypes.HexAddressToBytes(req.Controller)
			k.issueTripartyCallback(ctx, req, controllerBytes)

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
		if err := k.DeleteTripartyBridgeRequest(ctx, req.Sequence); err != nil {
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
	controllerBytes []byte,
) {
	callbackData := req.CallbackData
	if callbackData == nil {
		callbackData = []byte{}
	}

	recipientBytes := evmtypes.HexAddressToBytes(req.Recipient)

	call, err := evmtypes.NewTripartyCallbackCall(
		authtypes.NewModuleAddress(types.ModuleName).Bytes(),
		controllerBytes,
		req.Sequence.BigInt(),
		recipientBytes,
		req.Amount.BigInt(),
		callbackData,
	)
	if err != nil {
		k.Logger(ctx).Warn(
			"failed to create triparty callback call; callback skipped",
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
