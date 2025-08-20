package keeper

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/ethereum/go-ethereum/common"
	"github.com/mezo-org/mezod/x/bridge/types"
	"github.com/stretchr/testify/require"
)

func TestOutflowLimitManagement(t *testing.T) {
	ctx, keeper := mockContext()
	tokenAddr := common.HexToAddress("0x1234567890123456789012345678901234567890").Bytes()

	// Initially should return zero
	limit := keeper.GetOutflowLimit(ctx, tokenAddr)
	require.True(t, limit.IsZero(), "initial limit should be zero")

	// Set a limit
	keeper.SetOutflowLimit(ctx, tokenAddr, math.NewInt(1000000))

	// Get the limit
	limit = keeper.GetOutflowLimit(ctx, tokenAddr)
	require.Equal(t, math.NewInt(1000000), limit, "limit should match what was set")

	// Update the limit
	keeper.SetOutflowLimit(ctx, tokenAddr, math.NewInt(2000000))
	limit = keeper.GetOutflowLimit(ctx, tokenAddr)
	require.Equal(t, math.NewInt(2000000), limit, "limit should be updated")

	// Set zero limit
	keeper.SetOutflowLimit(ctx, tokenAddr, math.ZeroInt())
	limit = keeper.GetOutflowLimit(ctx, tokenAddr)
	require.True(t, limit.IsZero(), "limit should be zero")
}

func TestCurrentOutflowManagement(t *testing.T) {
	ctx, keeper := mockContext()
	tokenAddr := common.HexToAddress("0x1234567890123456789012345678901234567890").Bytes()

	// Start with zero outflow
	outflow := keeper.getCurrentOutflow(ctx, tokenAddr)
	require.True(t, outflow.IsZero())

	// Increase outflow
	keeper.increaseCurrentOutflow(ctx, tokenAddr, math.NewInt(500))

	// Check updated outflow
	outflow = keeper.getCurrentOutflow(ctx, tokenAddr)
	require.Equal(t, math.NewInt(500), outflow)

	// Increase again
	keeper.increaseCurrentOutflow(ctx, tokenAddr, math.NewInt(300))

	// Check cumulative outflow
	outflow = keeper.getCurrentOutflow(ctx, tokenAddr)
	require.Equal(t, math.NewInt(800), outflow)
}

func TestCheckOutflowLimit(t *testing.T) {
	ctx, keeper := mockContext()
	tokenAddr := common.HexToAddress("0x1234567890123456789012345678901234567890").Bytes()

	t.Run("checkOutflowLimit with no limit set", func(t *testing.T) {
		// No limit set means zero limit, so any amount should fail
		err := keeper.checkOutflowLimit(ctx, tokenAddr, math.NewInt(1))
		require.Error(t, err)
		require.ErrorIs(t, err, types.ErrOutflowLimitExceeded)
	})

	t.Run("checkOutflowLimit with limit set and no outflow", func(t *testing.T) {
		// Set a limit
		keeper.SetOutflowLimit(ctx, tokenAddr, math.NewInt(1000))

		// Amount within limit should pass
		err := keeper.checkOutflowLimit(ctx, tokenAddr, math.NewInt(500))
		require.NoError(t, err)

		// Amount equal to limit should pass
		err = keeper.checkOutflowLimit(ctx, tokenAddr, math.NewInt(1000))
		require.NoError(t, err)

		// Amount exceeding limit should fail
		err = keeper.checkOutflowLimit(ctx, tokenAddr, math.NewInt(1001))
		require.Error(t, err)
		require.ErrorIs(t, err, types.ErrOutflowLimitExceeded)
	})

	t.Run("checkOutflowLimit with limit set and existing outflow", func(t *testing.T) {
		// Set limit and current outflow
		keeper.SetOutflowLimit(ctx, tokenAddr, math.NewInt(1000))
		keeper.increaseCurrentOutflow(ctx, tokenAddr, math.NewInt(600))

		// Amount within remaining capacity should pass
		err := keeper.checkOutflowLimit(ctx, tokenAddr, math.NewInt(400))
		require.NoError(t, err)

		// Amount exceeding remaining capacity should fail
		err = keeper.checkOutflowLimit(ctx, tokenAddr, math.NewInt(401))
		require.Error(t, err)
		require.ErrorIs(t, err, types.ErrOutflowLimitExceeded)
	})
}

func TestGetOutflowCapacity(t *testing.T) {
	ctx, keeper := mockContext()
	tokenAddr := common.HexToAddress("0x1234567890123456789012345678901234567890").Bytes()

	t.Run("GetOutflowCapacity with no limit", func(t *testing.T) {
		capacity, resetHeight := keeper.GetOutflowCapacity(ctx, tokenAddr)
		require.True(t, capacity.IsZero(), "capacity should be zero when no limit set")
		require.Equal(t, uint64(OutflowResetBlocks), resetHeight, "reset height should be OutflowResetBlocks")
	})

	t.Run("GetOutflowCapacity with limit set", func(t *testing.T) {
		// Set a limit
		keeper.SetOutflowLimit(ctx, tokenAddr, math.NewInt(1000))

		capacity, resetHeight := keeper.GetOutflowCapacity(ctx, tokenAddr)
		require.Equal(t, math.NewInt(1000), capacity, "capacity should equal limit when no outflow")
		require.Equal(t, uint64(OutflowResetBlocks), resetHeight)
	})

	t.Run("GetOutflowCapacity with existing outflow", func(t *testing.T) {
		// Set limit and current outflow
		keeper.SetOutflowLimit(ctx, tokenAddr, math.NewInt(1000))
		keeper.increaseCurrentOutflow(ctx, tokenAddr, math.NewInt(300))

		capacity, resetHeight := keeper.GetOutflowCapacity(ctx, tokenAddr)
		require.Equal(t, math.NewInt(700), capacity, "capacity should be limit minus current outflow")
		require.Equal(t, uint64(OutflowResetBlocks), resetHeight)
	})

	t.Run("GetOutflowCapacity with outflow exceeding limit", func(t *testing.T) {
		// This shouldn't normally happen, but let's test the safety check
		keeper.SetOutflowLimit(ctx, tokenAddr, math.NewInt(100))
		keeper.increaseCurrentOutflow(ctx, tokenAddr, math.NewInt(200))

		capacity, _ := keeper.GetOutflowCapacity(ctx, tokenAddr)
		require.True(t, capacity.IsZero(), "capacity should be zero when outflow exceeds limit")
	})
}

func TestLastOutflowResetManagement(t *testing.T) {
	ctx, keeper := mockContext()

	// Initially should be zero
	lastReset := keeper.getLastOutflowReset(ctx)
	require.Equal(t, uint64(0), lastReset, "initial reset height should be zero")

	// Set reset height
	keeper.setLastOutflowReset(ctx, 12345)
	lastReset = keeper.getLastOutflowReset(ctx)
	require.Equal(t, uint64(12345), lastReset, "reset height should match what was set")
}

func TestResetAllOutflows(t *testing.T) {
	ctx, keeper := mockContext()
	tokenAddr1 := common.HexToAddress("0x1111111111111111111111111111111111111111").Bytes()
	tokenAddr2 := common.HexToAddress("0x2222222222222222222222222222222222222222").Bytes()

	t.Run("resetAllOutflows with existing outflows", func(t *testing.T) {
		// Set some outflows
		keeper.increaseCurrentOutflow(ctx, tokenAddr1, math.NewInt(500))
		keeper.increaseCurrentOutflow(ctx, tokenAddr2, math.NewInt(300))

		// Verify outflows are set
		outflow1 := keeper.getCurrentOutflow(ctx, tokenAddr1)
		outflow2 := keeper.getCurrentOutflow(ctx, tokenAddr2)
		require.Equal(t, math.NewInt(500), outflow1)
		require.Equal(t, math.NewInt(300), outflow2)

		// Reset all outflows
		keeper.resetAllOutflows(ctx)

		// Verify outflows are reset to zero
		outflow1 = keeper.getCurrentOutflow(ctx, tokenAddr1)
		outflow2 = keeper.getCurrentOutflow(ctx, tokenAddr2)
		require.True(t, outflow1.IsZero(), "outflow1 should be reset to zero")
		require.True(t, outflow2.IsZero(), "outflow2 should be reset to zero")
	})

	t.Run("resetAllOutflows with no outflows", func(_ *testing.T) {
		keeper.resetAllOutflows(ctx)
	})
}
