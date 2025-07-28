package keeper

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/mezo-org/mezod/x/bridge/types"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
)

// GetAssetsUnlockedSequenceTip returns the current sequence tip for the
// AssetsUnlocked events. The tip denotes the sequence number of the last event
// processed by the x/bridge module.
func (k Keeper) GetAssetsUnlockedSequenceTip(ctx sdk.Context) math.Int {
	bz := ctx.KVStore(k.storeKey).Get(types.AssetsUnlockedSequenceTipKey)

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

// SetAssetsUnlockedSequenceTip sets the current sequence tip for the AssetsUnlocked
// events. The tip denotes the sequence number of the last event processed by
// the x/bridge module.
func (k Keeper) setAssetsUnlockedSequenceTip(
	ctx sdk.Context,
	sequenceTip math.Int,
) {
	bz, err := sequenceTip.Marshal()
	if err != nil {
		panic(err)
	}

	ctx.KVStore(k.storeKey).Set(types.AssetsUnlockedSequenceTipKey, bz)
}

func (k Keeper) saveAssetsUnlocked(
	ctx sdk.Context,
	assetsUnlocked *types.AssetsUnlockedEvent,
) {
	bz, err := assetsUnlocked.Marshal()
	if err != nil {
		panic(err)
	}

	ctx.KVStore(k.storeKey).Set(types.GetAssetsUnlockedKey(assetsUnlocked.UnlockSequence), bz)
}

func (k Keeper) GetAssetsUnlocked(
	ctx sdk.Context,
	sequence math.Int,
) *types.AssetsUnlockedEvent {
	bz := ctx.KVStore(k.storeKey).Get(types.GetAssetsUnlockedKey(sequence))

	var assetsUnlocked types.AssetsUnlockedEvent
	err := assetsUnlocked.Unmarshal(bz)
	if err != nil {
		panic(err)
	}

	return &assetsUnlocked
}

func (k Keeper) SaveAssetsUnlocked(
	ctx sdk.Context,
	token []byte,
	amount math.Int,
	chain uint8,
	recipient []byte,
) (*types.AssetsUnlockedEvent, error) {
	var targetToken string
	// is it the btc token?
	btcToken := evmtypes.HexAddressToBytes(
		evmtypes.BTCTokenPrecompileAddress,
	)
	if bytes.Equal(btcToken, token) {
		targetToken = evmtypes.BytesToHexAddress(k.GetSourceBTCToken(ctx))
	} else {
		if mapping, ok := k.GetERC20TokenMappingFromMezoToken(ctx, token); ok {
			targetToken = evmtypes.BytesToHexAddress(
				mapping.SourceTokenBytes(),
			)
		}
	}

	if len(targetToken) == 0 {
		return nil, fmt.Errorf("unknown token %v", hex.EncodeToString(token))
	}

	// calculate the next unlock sequence
	nextSequence := k.GetAssetsUnlockedSequenceTip(ctx).AddRaw(1)

	// save it
	k.setAssetsUnlockedSequenceTip(
		ctx, nextSequence,
	)

	// then save the event
	assetsUnlocked := &types.AssetsUnlockedEvent{
		UnlockSequence: nextSequence,
		Recipient:      recipient,
		Amount:         amount,
		Token:          targetToken,
		Chain:          uint32(chain),
	}
	k.saveAssetsUnlocked(ctx, assetsUnlocked)

	return assetsUnlocked, nil
}

func (k Keeper) BurnBTC(
	ctx sdk.Context,
	fromAddr []byte,
	amount math.Int,
) error {
	coins := sdk.NewCoins(sdk.NewCoin(evmtypes.DefaultEVMDenom, amount))

	err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, fromAddr, types.ModuleName, coins)
	if err != nil {
		return err
	}

	err = k.bankKeeper.BurnCoins(ctx, types.ModuleName, coins)
	if err != nil {
		return err
	}

	err = k.IncreaseBTCBurnt(ctx, amount)
	if err != nil {
		return err
	}

	return nil
}

func (k Keeper) BurnERC20(
	ctx sdk.Context,
	token []byte,
	fromAddr []byte,
	amount *big.Int,
) error {
	call, err := evmtypes.NewERC20BurnFromCall(
		authtypes.NewModuleAddress(types.ModuleName).Bytes(),
		token,
		fromAddr,
		amount,
	)
	if err != nil {
		return fmt.Errorf("failed to create ERC20 burnFrom call: %w", err)
	}

	_, err = k.evmKeeper.ExecuteContractCall(ctx, call)
	if err != nil {
		return fmt.Errorf("failed to execute ERC20 burnFrom call: %w", err)
	}

	return nil
}
