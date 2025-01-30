package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
)

// EndBlock is use to run a given set of insert after processing a block.
// Each assert will ensure that the mezo state is in sync with its understanding
// of the state of the bridge at the time.
// In case an assertion prove false, the function will panic, leaving time
// for the node operators to investigate.
func (k *Keeper) EndBlock(ctx context.Context) error {
	asserts := []func(context.Context) error{
		k.verifyBTCSupply,
		k.verifyBridgeSequenceTip,
		k.verifyERC20Supply,
	}

	for _, f := range asserts {
		if err := f(ctx); err != nil {
			panic(fmt.Sprintf("inconsistent state between the bridge and mezo: %v", err))
		}
	}

	return nil
}

// verifyBTCSupply asserts that:
// btc_supply = total_btc_minted - total_btc_burnt.
// btc_supply being the total supply of BTC as tracked by x/bank
// and total_btc_{minted/burnt} being value tracked when the x/bridge
// is instructed to burn or mint BTC.
func (k *Keeper) verifyBTCSupply(ctx context.Context) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	var (
		totalSupply = k.bankKeeper.GetSupply(ctx, evmtypes.DefaultEVMDenom)
		totalMinted = k.GetCoinsMinted(sdkCtx, evmtypes.DefaultEVMDenom)
		totalBurnt  = k.GetCoinsBurnt(sdkCtx, evmtypes.DefaultEVMDenom)
	)

	if !totalSupply.IsEqual(totalMinted.Sub(totalBurnt)) {
		return fmt.Errorf(
			"invalid asset supply x/bank = %v, total minted = %v, total burnt = %v",
			totalSupply, totalMinted, totalBurnt,
		)
	}

	k.Logger(sdkCtx).Info("safe BTC supply state",
		"totalSupply", totalSupply.String(),
		"totalMinted", totalMinted.String(),
		"totalBurnt", totalBurnt.String(),
	)

	return nil
}

// verifyBridgeSequenceTip ...
func (k *Keeper) verifyBridgeSequenceTip(ctx context.Context) error {
	/* todo */
	return nil
}

// verifyERC20Supply ...
func (k *Keeper) verifyERC20Supply(ctx context.Context) error {
	/* todo */
	return nil
}
