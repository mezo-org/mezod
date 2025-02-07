package keeper

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/mezo-org/mezod/x/bridge/types"
)

// GetSourceBTCToken returns the BTC token address on the source chain.
// AssetsLocked events carrying this token address are directly mapped to the
// Mezo native denomination - BTC.
func (k Keeper) GetSourceBTCToken(ctx sdk.Context) []byte {
	return ctx.KVStore(k.storeKey).Get(types.SourceBTCTokenKey)
}

// SetSourceBTCToken sets the BTC token address on the source chain.
func (k Keeper) SetSourceBTCToken(ctx sdk.Context, sourceBTCToken []byte) {
	ctx.KVStore(k.storeKey).Set(types.SourceBTCTokenKey, sourceBTCToken)
}

// GetBTCMinted return the amount of a given BTC minted.
func (k Keeper) GetBTCMinted(ctx sdk.Context) math.Int {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.BTCMintedKey)
	if len(bz) == 0 {
		return math.NewInt(0)
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
