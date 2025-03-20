package keeper

import (
	"bytes"
	"fmt"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

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

// AcceptAssetsLocked processes the given AssetsLocked events sequence by minting
// the corresponding amount of coins for each event and sending them to the
// recipient address.
//
// Requirements:
//  1. The AssetsLocked sequence must not be empty.
//  2. The AssetsLocked sequence must be valid (i.e. all events in the slice
//     pass the AssetsLockedEvent.IsValid test AND sequence numbers of events
//     form a sequence strictly increasing by 1).
//  3. The sequence number of the first event in the slice must be exactly one
//     greater than the current sequence tip held in the state.
//
// The function returns an error if any of the requirements is not met.
// Checking the mentioned requirements is crucial to ensure state consistency
// regardless of the guarantees provided by the upstream code.
//
// If all requirements are met and x/bank interactions are all successful, the
// current sequence tip in the state is updated to the sequence number of the
// last event in the slice.
func (k Keeper) AcceptAssetsLocked(
	ctx sdk.Context,
	events types.AssetsLockedEvents,
) error {
	if len(events) == 0 {
		return fmt.Errorf("empty AssetsLocked sequence")
	}

	if !events.IsValid() {
		return fmt.Errorf("invalid AssetsLocked sequence")
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

	sourceBTCToken := k.GetSourceBTCToken(ctx)

	for _, event := range events {
		recipient, err := sdk.AccAddressFromBech32(event.Recipient)
		if err != nil {
			return fmt.Errorf("failed to parse recipient address: %w", err)
		}

		if bytes.Equal(event.TokenBytes(), sourceBTCToken) {
			// in cases of BTC mint, we need to ensure that
			// the funds are not sent to a blocked address first,
			// in case they are blocked, then not minting should happen,
			// the funds will be locked in the bridge.
			if _, ok := k.blockedAddrs[recipient.String()]; ok {
				ctx.Logger().Warn(
					"BTC deposit recipient is a blocked address; "+
						"AssetsLocked event skipped",
					"eventSequence", event.Sequence,
				)

				continue
			}

			err = k.mintBTC(ctx, recipient, event.Amount)
			if err != nil {
				return fmt.Errorf(
					"failed to mint BTC for event %v: %w",
					event.Sequence,
					err,
				)
			}
		} else {
			mapping, exists := k.GetERC20TokenMapping(ctx, event.TokenBytes())
			if !exists {
				// In case of a missing mapping, we skip the event and log a
				// warning. This is because we cannot mint ERC20 tokens without
				// a valid mapping. NOTE THAT THE SKIPPED EVENT WON'T BE
				// RE-PROCESSED SO FUNDS REMAIN LOCKED ON THE SOURCE CHAIN.
				ctx.Logger().Warn(
					"ERC20 mapping for source token not found; "+
						"AssetsLocked event skipped",
					"eventSequence", event.Sequence,
				)
				continue
			}

			err = k.mintERC20(
				ctx,
				recipient,
				mapping.MezoTokenBytes(),
				event.Amount,
			)
			if err != nil {
				// In case of an error, we skip the event and log a warning.
				// Unlike the BTC minting, we do not return an error here to
				// stop the consensus engine. This is because minting occurs on
				// external contracts, and we cannot guarantee that the minting
				// will always succeed. We cannot take a risk that one
				// malfunctioning contract will halt the whole chain.
				// NOTE THAT THE SKIPPED EVENT WON'T BE RE-PROCESSED SO FUNDS
				// REMAIN LOCKED ON THE SOURCE CHAIN.
				ctx.Logger().Error(
					"Encountered error while minting ERC20; "+
						"AssetsLocked event skipped",
					"eventSequence", event.Sequence,
					"error", err,
				)
				continue
			}
		}
	}

	k.setAssetsLockedSequenceTip(ctx, events[len(events)-1].Sequence)

	return nil
}

// mintBTC mints the given amount of BTC to the recipient address, directly
// in the x/bank module.
func (k Keeper) mintBTC(
	ctx sdk.Context,
	recipient sdk.AccAddress,
	amount math.Int,
) error {
	coins := sdk.NewCoins(sdk.NewCoin(evmtypes.DefaultEVMDenom, amount))

	// Mint coins to the x/bridge module account.
	err := k.bankKeeper.MintCoins(ctx, types.ModuleName, coins)
	if err != nil {
		return fmt.Errorf("failed to mint coins: %w", err)
	}

	// Send the minted coins from x/bridge module account to the final recipient.
	err = k.bankKeeper.SendCoinsFromModuleToAccount(
		ctx,
		types.ModuleName,
		recipient,
		coins,
	)
	if err != nil {
		return fmt.Errorf("failed to send coins: %w", err)
	}

	if err := k.IncreaseBTCMinted(ctx, amount); err != nil {
		return fmt.Errorf("failed to increase btc minted: %w", err)
	}

	return nil
}

// mintERC20 mints the given amount of ERC20 token to the recipient address,
// by executing a mint(address,uint256) call on the token contract.
func (k Keeper) mintERC20(
	ctx sdk.Context,
	recipient sdk.AccAddress,
	token []byte,
	amount math.Int,
) error {
	call, err := evmtypes.NewERC20MintCall(
		authtypes.NewModuleAddress(types.ModuleName).Bytes(),
		token,
		recipient.Bytes(),
		amount.BigInt(),
	)
	if err != nil {
		return fmt.Errorf("failed to create ERC20 mint call: %w", err)
	}

	_, err = k.evmKeeper.ExecuteContractCall(ctx, call)
	if err != nil {
		return fmt.Errorf("failed to execute ERC20 mint call: %w", err)
	}

	return nil
}
