package keeper

import (
	"bytes"
	"sort"
	"testing"

	"github.com/mezo-org/mezod/x/poa/types"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/require"
)

func TestTrackHistoricalInfo(t *testing.T) {
	ctx, poaKeeper := mockContext()
	validator1, _ := mockValidator()
	validator2, _ := mockValidator()

	// Set historical entries in params to 5.
	poaKeeper.historicalEntries = 5

	// Add initial validators.
	poaKeeper.createValidator(ctx, validator1)
	poaKeeper.createValidator(ctx, validator2)

	// Refresh the validator set.
	poaKeeper.EndBlocker(ctx)

	// Make sure the validator set is correct.
	activeValSet := poaKeeper.GetActiveValidators(ctx)
	require.Len(t, activeValSet, 2)

	// Set historical info at 5, 4 which should be pruned and check that it
	// has been stored.
	h4 := tmproto.Header{
		ChainID: "HelloChain",
		Height:  4,
	}
	h5 := tmproto.Header{
		ChainID: "HelloChain",
		Height:  5,
	}
	hi4 := types.NewHistoricalInfo(h4, activeValSet)
	hi5 := types.NewHistoricalInfo(h5, activeValSet)
	poaKeeper.SetHistoricalInfo(ctx, 4, &hi4)
	poaKeeper.SetHistoricalInfo(ctx, 5, &hi5)

	recv, found := poaKeeper.GetHistoricalInfo(ctx, 4)
	require.True(t, found)
	require.Equal(t, hi4, recv)
	recv, found = poaKeeper.GetHistoricalInfo(ctx, 5)
	require.True(t, found)
	require.Equal(t, hi5, recv)

	// Add a new validator.
	validator3, _ := mockValidator()
	poaKeeper.createValidator(ctx, validator3)

	// Refresh the validator set.
	poaKeeper.EndBlocker(ctx)

	// Make sure the validator set is correct.
	activeValSet = poaKeeper.GetActiveValidators(ctx)
	require.Len(t, activeValSet, 3)
	// Sort the validator set in the same way that historical info does.
	sort.SliceStable(activeValSet, func(i, j int) bool {
		return bytes.Compare(
			activeValSet[i].GetOperator(),
			activeValSet[j].GetOperator(),
		) == -1
	})

	// Set Header for BeginBlock context.
	header := tmproto.Header{
		ChainID: "HelloChain",
		Height:  10,
	}
	ctx = ctx.WithBlockHeader(header)

	poaKeeper.TrackHistoricalInfo(ctx)

	// Check HistoricalInfo at height 10 is persisted.
	expected := types.HistoricalInfo{
		Header: header,
		Valset: activeValSet,
	}
	recv, found = poaKeeper.GetHistoricalInfo(ctx, 10)
	require.True(t, found, "GetHistoricalInfo failed after BeginBlock")
	require.Equal(t, expected, recv, "GetHistoricalInfo returned unexpected result")

	// Check HistoricalInfo at height 5, 4 is pruned.
	recv, found = poaKeeper.GetHistoricalInfo(ctx, 4)
	require.False(t, found, "GetHistoricalInfo did not prune earlier height")
	require.Equal(t, types.HistoricalInfo{}, recv, "GetHistoricalInfo at height 4 is not empty after prune")
	recv, found = poaKeeper.GetHistoricalInfo(ctx, 5)
	require.False(t, found, "GetHistoricalInfo did not prune first prune height")
	require.Equal(t, types.HistoricalInfo{}, recv, "GetHistoricalInfo at height 5 is not empty after prune")
}
