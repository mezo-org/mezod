package keeper

import (
	"fmt"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/mezo-org/mezod/x/bridge/types"
)

type Keeper struct {
	cdc        codec.Codec
	storeKey   storetypes.StoreKey
	bankKeeper types.BankKeeper
	evmKeeper  types.EvmKeeper
}

func NewKeeper(
	cdc codec.Codec,
	storeKey storetypes.StoreKey,
	bankKeeper types.BankKeeper,
	evmKeeper types.EvmKeeper,
) Keeper {
	return Keeper{
		cdc:        cdc,
		storeKey:   storeKey,
		bankKeeper: bankKeeper,
		evmKeeper:  evmKeeper,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) GetCoinsMinted(ctx sdk.Context, denom string) (coins sdk.Coin) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.CoinMintedKey(denom))
	if len(bz) == 0 {
		return sdk.NewCoin(denom, math.NewInt(0))
	}

	k.cdc.MustUnmarshal(bz, &coins)
	return coins
}

func (k Keeper) IncreaseCoinsMinted(ctx sdk.Context, newMint sdk.Coin) error {
	store := ctx.KVStore(k.storeKey)

	coinsMinted := sdk.NewCoin(newMint.Denom, math.NewInt(0))
	bz := store.Get(types.CoinMintedKey(newMint.Denom))
	if len(bz) > 0 {
		k.cdc.MustUnmarshal(bz, &coinsMinted)
	}

	bz, err := k.cdc.Marshal(&coinsMinted)
	if err != nil {
		return err
	}

	store.Set(types.CoinMintedKey(coinsMinted.Denom), bz)

	return nil
}

func (k Keeper) GetCoinsBurnt(ctx sdk.Context, denom string) (coins sdk.Coin) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.CoinBurntKey(denom))
	if len(bz) == 0 {
		return sdk.NewCoin(denom, math.NewInt(0))
	}

	k.cdc.MustUnmarshal(bz, &coins)
	return coins
}

func (k Keeper) IncreaseCoinsBurnt(ctx sdk.Context, newBurn sdk.Coin) error {
	store := ctx.KVStore(k.storeKey)

	coinsBurnt := sdk.NewCoin(newBurn.Denom, math.NewInt(0))
	bz := store.Get(types.CoinBurntKey(newBurn.Denom))
	if len(bz) > 0 {
		k.cdc.MustUnmarshal(bz, &coinsBurnt)
	}

	bz, err := k.cdc.Marshal(&coinsBurnt)
	if err != nil {
		return err
	}

	store.Set(types.CoinBurntKey(coinsBurnt.Denom), bz)

	return nil
}
