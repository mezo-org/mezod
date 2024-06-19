package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/evmos/evmos/v12/x/dualstaking/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) Stake(goCtx context.Context, req *types.QueryStakeRequest) (*types.QueryStakeResponse, error) {
	if req == nil {
		return nil, errors.ErrInvalidRequest
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	store := ctx.KVStore(k.storeKey)
	stakingStore := prefix.NewStore(store, types.KeyStakingPositionPrefix)

	bz := stakingStore.Get([]byte(req.StakeId))
	if bz == nil {
		return nil, types.ErrStakeNotFound
	}

	var stake types.Stake
	k.cdc.MustUnmarshal(bz, &stake)

	return &types.QueryStakeResponse{Stake: &stake}, nil
}

func (k Keeper) Delegation(goCtx context.Context, req *types.QueryDelegationRequest) (*types.QueryDelegationResponse, error) {
	if req == nil {
		return nil, errors.ErrInvalidRequest
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	store := ctx.KVStore(k.storeKey)
	delegationStore := prefix.NewStore(store, types.KeyDelegationPositionPrefix)

	bz := delegationStore.Get([]byte(req.DelegationId))
	if bz == nil {
		return nil, types.ErrDelegationNotFound
	}

	var delegation types.Delegation
	k.cdc.MustUnmarshal(bz, &delegation)

	return &types.QueryDelegationResponse{Delegation: &delegation}, nil
}

func (k Keeper) StakesByStaker(goCtx context.Context, req *types.QueryStakesByStakerRequest) (*types.QueryStakesByStakerResponse, error) {
	if req == nil {
		return nil, errors.ErrInvalidRequest
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	store := ctx.KVStore(k.storeKey)
	stakingStore := prefix.NewStore(store, types.KeyStakingPositionPrefix)

	iterator := sdk.KVStorePrefixIterator(stakingStore, []byte(req.Staker))
	defer iterator.Close()

	var stakes []*types.Stake
	for ; iterator.Valid(); iterator.Next() {
		var stake types.Stake
		k.cdc.MustUnmarshal(iterator.Value(), &stake)
		stakes = append(stakes, &stake)
	}

	return &types.QueryStakesByStakerResponse{Stakes: stakes}, nil
}

func (k Keeper) DelegationsByStaker(goCtx context.Context, req *types.QueryDelegationsByStakerRequest) (*types.QueryDelegationsByStakerResponse, error) {
	if req == nil {
		return nil, errors.ErrInvalidRequest
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	store := ctx.KVStore(k.storeKey)
	delegationStore := prefix.NewStore(store, types.KeyDelegationPositionPrefix)

	iterator := sdk.KVStorePrefixIterator(delegationStore, []byte(req.Staker))
	defer iterator.Close()

	var delegations []*types.Delegation
	for ; iterator.Valid(); iterator.Next() {
		var delegation types.Delegation
		k.cdc.MustUnmarshal(iterator.Value(), &delegation)
		delegations = append(delegations, &delegation)
	}

	return &types.QueryDelegationsByStakerResponse{Delegations: delegations}, nil
}
