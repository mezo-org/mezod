package keeper

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/mezo-org/mezod/x/bridge/types"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
)

// GetCoinsMinted return the amount of a given coin minted.
func (k Keeper) GetCoinsMinted(ctx sdk.Context, denom string) (coins sdk.Coin) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.CoinMintedKey(denom))
	if len(bz) == 0 {
		return k.applyStorageMigration(ctx, denom)
	}

	k.cdc.MustUnmarshal(bz, &coins)

	return coins
}

// IncreaseCoinsMinted increase the total amount of coin minted.
func (k Keeper) IncreaseCoinsMinted(ctx sdk.Context, newMint sdk.Coin) error {
	store := ctx.KVStore(k.storeKey)

	coinsMinted := sdk.NewCoin(newMint.Denom, math.NewInt(0))
	bz := store.Get(types.CoinMintedKey(newMint.Denom))
	if len(bz) > 0 {
		k.cdc.MustUnmarshal(bz, &coinsMinted)
	}

	coinsMinted.Amount = coinsMinted.Amount.Add(newMint.Amount)

	bz, err := k.cdc.Marshal(&coinsMinted)
	if err != nil {
		return err
	}

	store.Set(types.CoinMintedKey(coinsMinted.Denom), bz)

	return nil
}

// GetCoinsBurnt returns the total amount of a given coin burnt.
func (k Keeper) GetCoinsBurnt(ctx sdk.Context, denom string) (coins sdk.Coin) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.CoinBurntKey(denom))
	if len(bz) == 0 {
		return sdk.NewCoin(denom, math.NewInt(0))
	}

	k.cdc.MustUnmarshal(bz, &coins)
	return coins
}

// IncreaseCoinsBurnt increments the total amount of coin burnt.
func (k Keeper) IncreaseCoinsBurnt(ctx sdk.Context, newBurn sdk.Coin) error {
	store := ctx.KVStore(k.storeKey)

	coinsBurnt := sdk.NewCoin(newBurn.Denom, math.NewInt(0))
	bz := store.Get(types.CoinBurntKey(newBurn.Denom))
	if len(bz) > 0 {
		k.cdc.MustUnmarshal(bz, &coinsBurnt)
	}

	coinsBurnt.Amount = coinsBurnt.Amount.Sub(newBurn.Amount)

	bz, err := k.cdc.Marshal(&coinsBurnt)
	if err != nil {
		return err
	}

	store.Set(types.CoinBurntKey(coinsBurnt.Denom), bz)

	return nil
}

// applyStorageMigration is used for migrations purposes. Every coins (BTC) or ERC20 token
// should have they amount burn and minted tracked at all time, however this was
// introduce only in a later upgrade. This will be call as soon as the first EndBlock call
// for the execution of the first block after the upgrade, if the keeper didn't have a
// storage slot for this coin specifically, and then initialise it with the current total
// supply known by the x/bank module.
func (k Keeper) applyStorageMigration(ctx sdk.Context, denomOrAddress string) (coins sdk.Coin) {
	// only applies for BTC for now.
	if denomOrAddress == evmtypes.DefaultEVMDenom {
		// get the current total supply from the x/bank
		supply := k.bankKeeper.GetSupply(ctx, evmtypes.DefaultEVMDenom)
		// if the supply is != 0, likely after an upgrade
		if !supply.IsZero() {
			// then we upgrade the store with the current
			// know supply
			if err := k.IncreaseCoinsMinted(ctx, supply); err != nil {
				panic(fmt.Sprintf("unable to migrate storage on the 1st block of an upgrade"))
			}
		}
	}

	// later handled ERC20 if migration is needed
	return sdk.NewCoin(denomOrAddress, math.NewInt(0))
}
