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

func (k Keeper) SetStakingPosition(ctx sdk.Context, position types.StakingPosition) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyStakingPositionPrefix)
	key := types.GetStakingPositionKey(position.Staker, position.StakeId)
	bz := k.cdc.MustMarshal(&position)
	store.Set(key, bz)
}

func (k Keeper) GetStakingPosition(ctx sdk.Context, staker string, stakeId string) *types.StakingPosition {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyStakingPositionPrefix)
	key := types.GetStakingPositionKey(staker, stakeId)
	bz := store.Get(key)
	if bz == nil {
		return nil
	}

	var position types.StakingPosition
	k.cdc.MustUnmarshal(bz, &position)
	return &position
}

func (k Keeper) SetDelegationPosition(ctx sdk.Context, position types.DelegationPosition) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyDelegationPositionPrefix)
	key := types.GetDelegationPositionKey(position.Staker, position.DelegationId)
	bz := k.cdc.MustMarshal(&position)
	store.Set(key, bz)
}

func (k Keeper) GetDelegationPosition(ctx sdk.Context, staker string, delegationId string) *types.DelegationPosition {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyDelegationPositionPrefix)
	key := types.GetDelegationPositionKey(staker, delegationId)
	bz := store.Get(key)
	if bz == nil {
		return nil
	}

	var position types.DelegationPosition
	k.cdc.MustUnmarshal(bz, &position)
	return &position
}

// TODO: add delete positions for staking and delegating
