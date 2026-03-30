package keeper

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"
)

func TestTripartyBlockDelayManagement(t *testing.T) {
	ctx, keeper := mockContext()

	// Initially should return default value of 1
	delay := keeper.GetTripartyBlockDelay(ctx)
	require.Equal(t, uint64(1), delay, "initial delay should be 1")

	// Set a delay
	keeper.SetTripartyBlockDelay(ctx, 100)

	// Get the delay
	delay = keeper.GetTripartyBlockDelay(ctx)
	require.Equal(t, uint64(100), delay, "delay should match what was set")

	// Update the delay
	keeper.SetTripartyBlockDelay(ctx, 200)
	delay = keeper.GetTripartyBlockDelay(ctx)
	require.Equal(t, uint64(200), delay, "delay should be updated")

	// Set back to minimum
	keeper.SetTripartyBlockDelay(ctx, 1)
	delay = keeper.GetTripartyBlockDelay(ctx)
	require.Equal(t, uint64(1), delay, "delay should be 1")
}

func TestTripartyPerRequestLimitManagement(t *testing.T) {
	ctx, keeper := mockContext()

	// Initially should return zero
	limit := keeper.GetTripartyPerRequestLimit(ctx)
	require.True(t, limit.IsZero(), "initial limit should be zero")

	// Set a limit
	keeper.SetTripartyPerRequestLimit(ctx, math.NewInt(1000000))

	// Get the limit
	limit = keeper.GetTripartyPerRequestLimit(ctx)
	require.Equal(t, math.NewInt(1000000), limit, "limit should match what was set")

	// Update the limit
	keeper.SetTripartyPerRequestLimit(ctx, math.NewInt(2000000))
	limit = keeper.GetTripartyPerRequestLimit(ctx)
	require.Equal(t, math.NewInt(2000000), limit, "limit should be updated")

	// Set zero limit
	keeper.SetTripartyPerRequestLimit(ctx, math.ZeroInt())
	limit = keeper.GetTripartyPerRequestLimit(ctx)
	require.True(t, limit.IsZero(), "limit should be zero")
}

func TestTripartyWindowLimitManagement(t *testing.T) {
	ctx, keeper := mockContext()

	// Initially should return zero
	limit := keeper.GetTripartyWindowLimit(ctx)
	require.True(t, limit.IsZero(), "initial limit should be zero")

	// Set a limit
	keeper.SetTripartyWindowLimit(ctx, math.NewInt(5000000))

	// Get the limit
	limit = keeper.GetTripartyWindowLimit(ctx)
	require.Equal(t, math.NewInt(5000000), limit, "limit should match what was set")

	// Update the limit
	keeper.SetTripartyWindowLimit(ctx, math.NewInt(10000000))
	limit = keeper.GetTripartyWindowLimit(ctx)
	require.Equal(t, math.NewInt(10000000), limit, "limit should be updated")

	// Set zero limit
	keeper.SetTripartyWindowLimit(ctx, math.ZeroInt())
	limit = keeper.GetTripartyWindowLimit(ctx)
	require.True(t, limit.IsZero(), "limit should be zero")
}
