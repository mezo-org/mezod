package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/evmos/evmos/v12/x/dualstaking/types"
)

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
