package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
)

// EndBlock is used to run a given set of assertions after processing a block.
// Each assertion will ensure that the mezo state is in sync with its
// understanding of the state of the bridge at the time.
// In case an assertion prove false, the function will panic, leaving time
// for the node operators to investigate.
func (k *Keeper) EndBlock(ctx context.Context) error {
	params := k.GetParams(sdk.UnwrapSDKContext(ctx))
	var asserts []func(context.Context) error

	if params.BtcSupplyAssertionEnabled {
		asserts = append(asserts, k.verifyBTCSupply)
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
		totalMinted = k.GetBTCMinted(sdkCtx)
		totalBurnt  = k.GetBTCBurnt(sdkCtx)
	)

	if !totalSupply.Amount.Equal(totalMinted.Sub(totalBurnt)) {
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
