package keeper

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/mezo-org/mezod/x/bridge/types"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
)

// GetSourceBTCToken returns the BTC token address on the source chain.
// AssetsLocked events carrying this token address are directly mapped to the
// Mezo native denomination - BTC.
func (k Keeper) GetSourceBTCToken(ctx sdk.Context) []byte {
	return ctx.KVStore(k.storeKey).Get(types.SourceBTCTokenKey)
}

func (k Keeper) setSourceBTCToken(ctx sdk.Context, sourceBTCToken []byte) {
	ctx.KVStore(k.storeKey).Set(types.SourceBTCTokenKey, sourceBTCToken)
}

// GetBTCMinted return the amount of a given BTC minted.
func (k Keeper) GetBTCMinted(ctx sdk.Context) math.Int {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.BTCMintedKey)
	if len(bz) == 0 {
		return k.applyBTCStorageMigration(ctx)
	}

	var amount math.Int
	if err := amount.Unmarshal(bz); err != nil {
		panic(err)
	}

	return amount
}

// IncreaseBTCMinted increase the total amount of BTC minted.
func (k Keeper) IncreaseBTCMinted(ctx sdk.Context, amount math.Int) error {
	store := ctx.KVStore(k.storeKey)
	btcMinted := math.NewInt(0)
	bz := store.Get(types.BTCMintedKey)
	if len(bz) > 0 {
		if err := btcMinted.Unmarshal(bz); err != nil {
			panic(err)
		}
	}

	btcMinted = btcMinted.Add(amount)

	bz, err := btcMinted.Marshal()
	if err != nil {
		panic(err)
	}

	store.Set(types.BTCMintedKey, bz)

	return nil
}

// GetBTCBurnt returns the total amount of a given BTC burnt.
func (k Keeper) GetBTCBurnt(ctx sdk.Context) math.Int {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.BTCBurntKey)
	if len(bz) == 0 {
		return math.NewInt(0)
	}

	var amount math.Int
	if err := amount.Unmarshal(bz); err != nil {
		panic(err)
	}

	return amount
}

// IncreaseBTCBurnt increments the total amount of BTC burnt.
func (k Keeper) IncreaseBTCBurnt(ctx sdk.Context, amount math.Int) error {
	store := ctx.KVStore(k.storeKey)
	btcBurnt := math.NewInt(0)
	bz := store.Get(types.BTCBurntKey)
	if len(bz) > 0 {
		if err := btcBurnt.Unmarshal(bz); err != nil {
			panic(err)
		}
	}

	btcBurnt = btcBurnt.Add(amount)

	bz, err := btcBurnt.Marshal()
	if err != nil {
		panic(err)
	}

	store.Set(types.BTCBurntKey, bz)

	return nil
}

// applyBTCStorageMigration is used for migrations purposes. Every BTCs (BTC) or ERC20 token
// should have they amount burn and minted tracked at all time, however this was
// introduce only in a later upgrade. This will be call as soon as the first EndBlock call
// for the execution of the first block after the upgrade, if the keeper didn't have a
// storage slot for this BTC specifically, and then initialize it with the current total
// supply known by the x/bank module.
func (k Keeper) applyBTCStorageMigration(ctx sdk.Context) math.Int {
	supply := k.bankKeeper.GetSupply(ctx, evmtypes.DefaultEVMDenom)
	// if the supply is != 0, likely after an upgrade
	if !supply.IsZero() {
		// then we upgrade the store with the current
		// know supply
		if err := k.IncreaseBTCMinted(ctx, supply.Amount); err != nil {
			panic(fmt.Sprintf("unable to migrate storage on the 1st block of an upgrade: %v", err))
		}
	}

	return supply.Amount
}
