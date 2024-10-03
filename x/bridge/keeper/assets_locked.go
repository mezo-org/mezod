package keeper

import (
	"fmt"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/mezo-org/mezod/x/bridge/types"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
)

// GetAssetsLockedSequenceTip returns the current sequence tip for the
// AssetsLocked events. The tip denotes the sequence number of the last event
// processed by the x/bridge module.
func (k Keeper) GetAssetsLockedSequenceTip(ctx sdk.Context) math.Int {
	bz := ctx.KVStore(k.storeKey).Get(types.AssetsLockedSequenceTipKey)

	var sequenceTip math.Int
	err := sequenceTip.Unmarshal(bz)
	if err != nil {
		panic(err)
	}

	if sequenceTip.IsNil() {
		sequenceTip = math.ZeroInt()
	}

	return sequenceTip
}

// SetAssetsLockedSequenceTip sets the current sequence tip for the AssetsLocked
// events. The tip denotes the sequence number of the last event processed by
// the x/bridge module.
func (k Keeper) setAssetsLockedSequenceTip(
	ctx sdk.Context,
	sequenceTip math.Int,
) {
	bz, err := sequenceTip.Marshal()
	if err != nil {
		panic(err)
	}

	ctx.KVStore(k.storeKey).Set(types.AssetsLockedSequenceTipKey, bz)
}

// AcceptAssetsLocked processes the given AssetsLocked events by minting the
// corresponding amount of coins and sending them to the respective recipients.
// A must-have precondition for this function is that the sequence number
// of the first event in the slice is exactly one greater than the current
// sequence tip held in the state. The function returns an error if this
// precondition is not met. If the processing is successful, the current
// sequence tip in the state is updated to the sequence number of the last
// event in the slice.
func (k Keeper) AcceptAssetsLocked(
	ctx sdk.Context,
	events []types.AssetsLockedEvent,
) error {
	if len(events) == 0 {
		return nil
	}

	currentSequenceTip := k.GetAssetsLockedSequenceTip(ctx)
	expectedSequenceStart := currentSequenceTip.AddRaw(1)
	if sequenceStart := events[0].Sequence; !expectedSequenceStart.Equal(sequenceStart) {
		return fmt.Errorf(
			"unexpected AssetsLocked sequence start; expected %s, got %s",
			expectedSequenceStart,
			sequenceStart,
		)
	}

	toMint := math.ZeroInt()
	for _, event := range events {
		toMint = toMint.Add(event.Amount)
	}

	err := k.bankKeeper.MintCoins(
		ctx,
		types.ModuleName,
		sdk.NewCoins(sdk.NewCoin(evmtypes.DefaultEVMDenom, toMint)),
	)
	if err != nil {
		return fmt.Errorf("failed to mint coins: %w", err)
	}

	for _, event := range events {
		recipient, err := sdk.AccAddressFromBech32(event.Recipient)
		if err != nil {
			return fmt.Errorf("failed to parse recipient address: %w", err)
		}

		err = k.bankKeeper.SendCoinsFromModuleToAccount(
			ctx,
			types.ModuleName,
			recipient,
			sdk.NewCoins(sdk.NewCoin(evmtypes.DefaultEVMDenom, event.Amount)),
		)
		if err != nil {
			return fmt.Errorf("failed to send coins: %w", err)
		}
	}

	k.setAssetsLockedSequenceTip(ctx, events[len(events)-1].Sequence)

	// TODO: Revisit this in the context of bridging events observability.
	//  From state's perspective, it's enough to update the sequence tip
	//  based on processed events to avoid double-bridging. Storing all
	//  processed events in the state is redundant, increases state management
	//  complexity, and negatively impacts the blockchain size in the long run.
	//  A sane alternative is using an opt-in EVM tx indexer (kv_indexer.go)
	//  to capture processed AssetsLocked events (they are part of the injected
	//  pseudo-tx and are available in the indexer) and expose them through
	//  a custom JSON-RPC API namespace (e.g. mezo_assetsLocked).

	return nil
}
