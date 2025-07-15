package keeper

import (
	"bytes"
	"fmt"

	"github.com/ethereum/go-ethereum/common"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/mezo-org/mezod/x/bridge/types"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
)

// GetAssetsUnlockedSequenceTip returns the current sequence tip for the
// AssetsLocked events. The tip denotes the sequence number of the last event
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

// SetAssetsUnlockedSequenceTip sets the current sequence tip for the AssetsLocked
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
	assetUnlocked *types.AssetsUnlockedEvent,
) {
	bz, err := assetUnlocked.Marshal()
	if err != nil {
		panic(err)
	}

	ctx.KVStore(k.storeKey).Set(types.GetAssetsUnlockedKey(assetUnlocked.Sequence), bz)
}

func (k Keeper) GetAssetsUnlocked(
	ctx sdk.Context,
	sequence math.Int,
) *types.AssetsUnlockedEvent {
	bz := ctx.KVStore(k.storeKey).Get(types.GetAssetsUnlockedKey(sequence))

	var assetUnlocked types.AssetsUnlockedEvent
	err := assetUnlocked.Unmarshal(bz)
	if err != nil {
		panic(err)
	}

	return &assetUnlocked
}

func (k Keeper) AssetsUnlocked(
	ctx sdk.Context,
	token []byte,
	amount math.Int,
	chain uint8,
	recipient []byte,
) (*types.AssetsUnlockedEvent, error) {
	var targetToken common.Address
	// is it the btc token?
	btcToken := common.HexToAddress(
		evmtypes.BTCTokenPrecompileAddress,
	)
	if bytes.Equal(btcToken.Bytes(), token) {
		targetToken = common.BytesToAddress(k.GetSourceBTCToken(ctx))
	} else {
		if mapping, ok := k.GetERC20TokenMapping(ctx, token); ok {
			targetToken = common.BytesToAddress(
				mapping.SourceTokenBytes(),
			)
		}
	}

	if len(targetToken) == 0 {
		return nil, fmt.Errorf("unknown token %v", token)
	}

	assetUnlocked := &types.AssetsUnlockedEvent{
		Sequence:  k.GetAssetsUnlockedSequenceTip(ctx),
		Recipient: recipient,
		Amount:    amount,
		Token:     targetToken.Hex(),
		Chain:     uint32(chain),
	}

	k.saveAssetsUnlocked(ctx, assetUnlocked)
	k.setAssetsUnlockedSequenceTip(
		ctx, assetUnlocked.Sequence.AddRaw(1),
	)

	return assetUnlocked, nil
}
