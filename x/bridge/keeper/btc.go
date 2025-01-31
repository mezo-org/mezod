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

// GetBTCsMinted return the amount of a given coin minted.
func (k Keeper) GetBTCsMinted(ctx sdk.Context) math.Int {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.BTCMintedKey)
	if len(bz) == 0 {
		return k.applyBTCStorageMigration(ctx)
	}

	var amount types.BalanceStorage
	k.cdc.MustUnmarshal(bz, &amount)

	return amount.Amount
}

// IncreaseBTCsMinted increase the total amount of coin minted.
func (k Keeper) IncreaseBTCsMinted(ctx sdk.Context, amount math.Int) error {
	store := ctx.KVStore(k.storeKey)

	coinsMinted := types.BalanceStorage{Amount: math.NewInt(0)}
	bz := store.Get(types.BTCMintedKey)
	if len(bz) > 0 {
		k.cdc.MustUnmarshal(bz, &coinsMinted)
	}

	coinsMinted.Amount = coinsMinted.Amount.Add(amount)

	bz, err := k.cdc.Marshal(&coinsMinted)
	if err != nil {
		return err
	}

	store.Set(types.BTCMintedKey, bz)

	return nil
}

// GetBTCsBurnt returns the total amount of a given coin burnt.
func (k Keeper) GetBTCsBurnt(ctx sdk.Context) math.Int {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.BTCBurntKey)
	if len(bz) == 0 {
		return math.NewInt(0)
	}

	var amount types.BalanceStorage
	k.cdc.MustUnmarshal(bz, &amount)

	return amount.Amount
}

// IncreaseBTCsBurnt increments the total amount of coin burnt.
func (k Keeper) IncreaseBTCsBurnt(ctx sdk.Context, amount math.Int) error {
	store := ctx.KVStore(k.storeKey)

	coinsBurnt := types.BalanceStorage{Amount: math.NewInt(0)}
	bz := store.Get(types.BTCBurntKey)
	if len(bz) > 0 {
		k.cdc.MustUnmarshal(bz, &coinsBurnt)
	}

	coinsBurnt.Amount = coinsBurnt.Amount.Add(amount)

	bz, err := k.cdc.Marshal(&coinsBurnt)
	if err != nil {
		return err
	}

	store.Set(types.BTCBurntKey, bz)

	return nil
}

// applyBTCStorageMigration is used for migrations purposes. Every coins (BTC) or ERC20 token
// should have they amount burn and minted tracked at all time, however this was
// introduce only in a later upgrade. This will be call as soon as the first EndBlock call
// for the execution of the first block after the upgrade, if the keeper didn't have a
// storage slot for this coin specifically, and then initialise it with the current total
// supply known by the x/bank module.
func (k Keeper) applyBTCStorageMigration(ctx sdk.Context) math.Int {
	supply := k.bankKeeper.GetSupply(ctx, evmtypes.DefaultEVMDenom)
	// if the supply is != 0, likely after an upgrade
	if !supply.IsZero() {
		// then we upgrade the store with the current
		// know supply
		if err := k.IncreaseBTCsMinted(ctx, supply.Amount); err != nil {
			panic(fmt.Sprintf("unable to migrate storage on the 1st block of an upgrade"))
		}
	}

	return supply.Amount
}
