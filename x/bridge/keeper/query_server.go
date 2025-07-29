package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/mezo-org/mezod/x/bridge/types"
)

var _ types.QueryServer = queryServer{}

type queryServer struct {
	keeper Keeper
}

// NewQueryServer returns an implementation of the QueryServer interface
// for the provided Keeper.
func NewQueryServer(keeper Keeper) types.QueryServer {
	return &queryServer{keeper: keeper}
}

func (qs queryServer) Params(
	ctx context.Context,
	_ *types.QueryParamsRequest,
) (*types.QueryParamsResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	params := qs.keeper.GetParams(sdkCtx)

	return &types.QueryParamsResponse{
		Params: params,
	}, nil
}

// AssetsUnlockedSequenceTip returns the current assets unlocked sequence tip.
func (qs queryServer) AssetsUnlockedSequenceTip(
	ctx context.Context,
	_ *types.QueryAssetsUnlockedSequenceTipRequest,
) (*types.QueryAssetsUnlockedSequenceTipResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sequenceTip := qs.keeper.GetAssetsUnlockedSequenceTip(sdkCtx)

	return &types.QueryAssetsUnlockedSequenceTipResponse{
		SequenceTip: sequenceTip,
	}, nil
}

// AssetsUnlockedEvents returns `AssetsUnlocked` events from the requested
// sequence range. The range is inclusive for sequence start and exclusive for
// sequence end. If the start or end of the requested range is not provided,
// default unbounded values are used.
// Notice that storing `AssetsUnlocked` events begins with the unlock sequence
// of `1`. Additionally, the current assets unlocked sequence tip is always
// equal to the total number of `AssetsUnlocked` events stored so far.
func (qs queryServer) AssetsUnlockedEvents(
	ctx context.Context,
	req *types.QueryAssetsUnlockedEventsRequest,
) (*types.QueryAssetsUnlockedEventsResponse, error) {
	start, end := req.SequenceStart, req.SequenceEnd

	// If the non-nil start and end of sequence were requested, ensure start is
	// smaller than end.
	if !start.IsNil() && !end.IsNil() && start.GTE(end) {
		return nil, fmt.Errorf("sequence start is not lower than sequence end")
	}

	// If the non-nil start and end of sequence were requested, ensure they are
	// positive.
	if (!start.IsNil() && !start.IsPositive()) || (!end.IsNil() && !end.IsPositive()) {
		return nil, fmt.Errorf("invalid non-positive sequence range")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sequenceTip := qs.keeper.GetAssetsUnlockedSequenceTip(sdkCtx)
	events := []types.AssetsUnlockedEvent{}

	// If no `AssetUnlocked` events have been stored so far, return an empty
	// list.
	if sequenceTip.IsZero() {
		return &types.QueryAssetsUnlockedEventsResponse{
			Events: events,
		}, nil
	}

	// If sequence start is nil, use `1` as the sequence start. Notice that
	// the first `AssetsUnlocked` event is stored with `UnlockSequence` of `1`,
	// not `0`.
	if start.IsNil() {
		start = math.OneInt()
	}

	// If sequence end is nil, use the current sequence tip plus `1` as the
	// end is exclusive.
	if end.IsNil() {
		end = sequenceTip.AddRaw(1)
	} else {
		end = math.MinInt(end, sequenceTip.AddRaw(1))
	}

	for seq := start; seq.LT(end); seq = seq.AddRaw(1) {
		event := qs.keeper.GetAssetsUnlocked(sdkCtx, seq)
		events = append(events, *event)
	}

	return &types.QueryAssetsUnlockedEventsResponse{
		Events: events,
	}, nil
}
