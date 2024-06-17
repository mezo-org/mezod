package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/evmos/evmos/v12/x/dualstaking/types"
)

var _ types.QueryServer = queryServer{}

type queryServer struct {
	Keeper
}

// NewQueryServer returns an implementation of the QueryServer interface
// for the provided Keeper.
func NewQueryServer(keeper Keeper) types.QueryServer {
	return &queryServer{Keeper: keeper}
}

func (q queryServer) StakingPosition(goCtx context.Context, req *types.QueryStakingPositionRequest) (*types.QueryStakingPositionResponse, error) {
	if req == nil {
		return nil, errortypes.ErrInvalidRequest
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	store := ctx.KVStore(q.storeKey)
	stakingStore := prefix.NewStore(store, types.KeyStakingPositionPrefix)

	bz := stakingStore.Get([]byte(req.StakeId))
	if bz == nil {
		return nil, types.ErrStakeNotFound
	}

	var stakingPosition types.StakingPosition
	q.cdc.MustUnmarshal(bz, &stakingPosition)

	return &types.QueryStakingPositionResponse{StakingPosition: &stakingPosition}, nil
}

func (q queryServer) DelegationPosition(goCtx context.Context, req *types.QueryDelegationPositionRequest) (*types.QueryDelegationPositionResponse, error) {
	if req == nil {
		return nil, errortypes.ErrInvalidRequest
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	store := ctx.KVStore(q.storeKey)
	delegationStore := prefix.NewStore(store, types.KeyDelegationPositionPrefix)

	bz := delegationStore.Get([]byte(req.DelegationId))
	if bz == nil {
		return nil, types.ErrDelegationNotFound
	}

	var delegationPosition types.DelegationPosition
	q.cdc.MustUnmarshal(bz, &delegationPosition)

	return &types.QueryDelegationPositionResponse{DelegationPosition: &delegationPosition}, nil
}

func (q queryServer) StakingPositionsByStaker(goCtx context.Context, req *types.QueryStakingPositionsByStakerRequest) (*types.QueryStakingPositionsByStakerResponse, error) {
	if req == nil {
		return nil, errortypes.ErrInvalidRequest
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	store := ctx.KVStore(q.storeKey)
	stakingStore := prefix.NewStore(store, types.KeyStakingPositionPrefix)

	iterator := sdk.KVStorePrefixIterator(stakingStore, []byte(req.Staker))
	defer iterator.Close()

	var stakingPositions []*types.StakingPosition
	for ; iterator.Valid(); iterator.Next() {
		var stakingPosition types.StakingPosition
		q.cdc.MustUnmarshal(iterator.Value(), &stakingPosition)
		stakingPositions = append(stakingPositions, &stakingPosition)
	}

	return &types.QueryStakingPositionsByStakerResponse{StakingPositions: stakingPositions}, nil
}

func (q queryServer) DelegationPositionsByStaker(goCtx context.Context, req *types.QueryDelegationPositionsByStakerRequest) (*types.QueryDelegationPositionsByStakerResponse, error) {
	if req == nil {
		return nil, errortypes.ErrInvalidRequest
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	store := ctx.KVStore(q.storeKey)
	delegationStore := prefix.NewStore(store, types.KeyDelegationPositionPrefix)

	iterator := sdk.KVStorePrefixIterator(delegationStore, []byte(req.Staker))
	defer iterator.Close()

	var delegationPositions []*types.DelegationPosition
	for ; iterator.Valid(); iterator.Next() {
		var delegationPosition types.DelegationPosition
		q.cdc.MustUnmarshal(iterator.Value(), &delegationPosition)
		delegationPositions = append(delegationPositions, &delegationPosition)
	}

	return &types.QueryDelegationPositionsByStakerResponse{DelegationPositions: delegationPositions}, nil
}
