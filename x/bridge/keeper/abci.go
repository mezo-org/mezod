package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
)

// EndBlock update block gas wanted.
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

	return nil
}

func (k *Keeper) verifyBridgeSequenceTip(ctx context.Context) error {
	/* todo */
	return nil
}

func (k *Keeper) verifyERC20Supply(ctx context.Context) error {
	/* todo */
	return nil
}
