package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/evmos/evmos/v12/x/dualstaking/types"
)

type Keeper struct {
	storeKey      storetypes.StoreKey
	cdc           codec.BinaryCodec
	accountKeeper types.AccountKeeper
	bankKeeper    types.BankKeeper
}

func NewKeeper(cdc codec.BinaryCodec, key storetypes.StoreKey, accountKeeper types.AccountKeeper, bankKeeper types.BankKeeper) Keeper {
	return Keeper{
		storeKey:      key,
		cdc:           cdc,
		accountKeeper: accountKeeper,
		bankKeeper:    bankKeeper,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) SetStake(ctx sdk.Context, position types.Stake) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyStakingPositionPrefix)
	key := types.GetStakeKey(position.Staker, position.StakeId)
	bz := k.cdc.MustMarshal(&position)
	store.Set(key, bz)
}

func (k Keeper) GetStake(ctx sdk.Context, staker string, stakeId string) *types.Stake {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyStakingPositionPrefix)
	key := types.GetStakeKey(staker, stakeId)
	bz := store.Get(key)
	if bz == nil {
		return nil
	}

	var position types.Stake
	k.cdc.MustUnmarshal(bz, &position)
	return &position
}

func (k Keeper) SetDelegation(ctx sdk.Context, position types.Delegation) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyDelegationPositionPrefix)
	key := types.GetDelegationKey(position.Staker, position.DelegationId)
	bz := k.cdc.MustMarshal(&position)
	store.Set(key, bz)
}

func (k Keeper) GetDelegation(ctx sdk.Context, staker string, delegationId string) *types.Delegation {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyDelegationPositionPrefix)
	key := types.GetDelegationKey(staker, delegationId)
	bz := store.Get(key)
	if bz == nil {
		return nil
	}

	var position types.Delegation
	k.cdc.MustUnmarshal(bz, &position)
	return &position
}

// TODO: add delete positions for staking and delegating
