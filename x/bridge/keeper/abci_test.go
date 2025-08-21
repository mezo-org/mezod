package keeper

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
	"github.com/stretchr/testify/require"
)

func TestEndBlock(t *testing.T) {
	ctx, k := mockContext()

	// first increase the BTC minted
	require.NoError(t, k.IncreaseBTCMinted(ctx, math.NewInt(42)))

	// ... BTC burnt
	require.NotPanics(t, func() { require.NoError(t, k.IncreaseBTCBurnt(ctx, math.NewInt(21))) })

	t.Run("does not panic when valid state", func(t *testing.T) {
		// return the same supply so all is fine
		k.bankKeeper.(*mockBankKeeper).
			On("GetSupply", ctx, evmtypes.DefaultEVMDenom).
			Return(sdk.NewCoin(evmtypes.DefaultEVMDenom, math.NewInt(21))).
			Times(1)

		require.NoError(t, k.EndBlock(ctx))
	})

	t.Run("panics when state is invalid", func(t *testing.T) {
		// return a different supply so not fine
		k.bankKeeper.(*mockBankKeeper).
			On("GetSupply", ctx, evmtypes.DefaultEVMDenom).
			Return(sdk.NewCoin(evmtypes.DefaultEVMDenom, math.NewInt(37))).
			Times(1)

		require.Panics(t, func() { _ = k.EndBlock(ctx) })
	})
}

func TestHandleOutflowReset(t *testing.T) {
	ctx, keeper := mockContext()
	tokenAddr1 := common.HexToAddress("0x1111111111111111111111111111111111111111").Bytes()
	tokenAddr2 := common.HexToAddress("0x2222222222222222222222222222222222222222").Bytes()

	t.Run("no reset when blocks threshold not reached", func(t *testing.T) {
		// Set some outflows
		keeper.increaseCurrentOutflow(ctx, tokenAddr1, math.NewInt(500))
		keeper.increaseCurrentOutflow(ctx, tokenAddr2, math.NewInt(300))

		// Set last reset to current block (0)
		//nolint:gosec
		keeper.setLastOutflowReset(ctx, uint64(ctx.BlockHeight()))

		// Call handleOutflowReset - should not reset since threshold not reached
		keeper.handleOutflowReset(ctx)

		// Verify outflows are still present
		outflow1 := keeper.getCurrentOutflow(ctx, tokenAddr1)
		outflow2 := keeper.getCurrentOutflow(ctx, tokenAddr2)
		require.Equal(t, math.NewInt(500), outflow1, "outflow1 should not be reset")
		require.Equal(t, math.NewInt(300), outflow2, "outflow2 should not be reset")
	})

	t.Run("reset when blocks threshold reached", func(t *testing.T) {
		// Create context at block height that triggers reset
		ctxAtResetBlock := ctx.WithBlockHeight(int64(OutflowResetBlocks))

		// Set some outflows
		keeper.increaseCurrentOutflow(ctxAtResetBlock, tokenAddr1, math.NewInt(500))
		keeper.increaseCurrentOutflow(ctxAtResetBlock, tokenAddr2, math.NewInt(300))

		// Set last reset to 0
		keeper.setLastOutflowReset(ctxAtResetBlock, 0)

		// Call handleOutflowReset - should reset since threshold reached
		keeper.handleOutflowReset(ctxAtResetBlock)

		// Verify outflows are reset to zero
		outflow1 := keeper.getCurrentOutflow(ctxAtResetBlock, tokenAddr1)
		outflow2 := keeper.getCurrentOutflow(ctxAtResetBlock, tokenAddr2)
		require.True(t, outflow1.IsZero(), "outflow1 should be reset to zero")
		require.True(t, outflow2.IsZero(), "outflow2 should be reset to zero")

		// Verify last reset height was updated
		lastReset := keeper.getLastOutflowReset(ctxAtResetBlock)
		require.Equal(t, uint64(OutflowResetBlocks), lastReset, "last reset height should be updated")
	})

	t.Run("reset when blocks threshold exceeded", func(t *testing.T) {
		// Create context at block height that exceeds reset threshold
		ctxBeyondReset := ctx.WithBlockHeight(int64(OutflowResetBlocks + 1000))

		// Set some outflows
		keeper.increaseCurrentOutflow(ctxBeyondReset, tokenAddr1, math.NewInt(1000))

		// Set last reset to 0
		keeper.setLastOutflowReset(ctxBeyondReset, 0)

		// Call handleOutflowReset - should reset since threshold exceeded
		keeper.handleOutflowReset(ctxBeyondReset)

		// Verify outflow is reset
		outflow := keeper.getCurrentOutflow(ctxBeyondReset, tokenAddr1)
		require.True(t, outflow.IsZero(), "outflow should be reset to zero")

		// Verify last reset height was updated
		lastReset := keeper.getLastOutflowReset(ctxBeyondReset)
		require.Equal(t, uint64(OutflowResetBlocks+1000), lastReset, "last reset height should be updated")
	})

	t.Run("reset with no existing outflows", func(t *testing.T) {
		// Create clean context
		ctxForReset := ctx.WithBlockHeight(int64(OutflowResetBlocks))

		// Set last reset to 0
		keeper.setLastOutflowReset(ctxForReset, 0)

		// Call handleOutflowReset - should not panic even with no outflows
		keeper.handleOutflowReset(ctxForReset)

		// Verify last reset height was updated
		lastReset := keeper.getLastOutflowReset(ctxForReset)
		require.Equal(t, uint64(OutflowResetBlocks), lastReset, "last reset height should be updated")
	})

	t.Run("multiple consecutive resets", func(t *testing.T) {
		// First reset at block OutflowResetBlocks
		ctxFirstReset := ctx.WithBlockHeight(int64(OutflowResetBlocks))
		keeper.setLastOutflowReset(ctxFirstReset, 0)
		keeper.increaseCurrentOutflow(ctxFirstReset, tokenAddr1, math.NewInt(100))

		keeper.handleOutflowReset(ctxFirstReset)
		require.True(t, keeper.getCurrentOutflow(ctxFirstReset, tokenAddr1).IsZero())

		// Add new outflow after first reset
		keeper.increaseCurrentOutflow(ctxFirstReset, tokenAddr1, math.NewInt(200))

		// Second reset at block 2 * OutflowResetBlocks
		ctxSecondReset := ctx.WithBlockHeight(int64(2 * OutflowResetBlocks))
		keeper.handleOutflowReset(ctxSecondReset)

		// Verify second reset
		require.True(t, keeper.getCurrentOutflow(ctxSecondReset, tokenAddr1).IsZero())
		lastReset := keeper.getLastOutflowReset(ctxSecondReset)
		require.Equal(t, uint64(2*OutflowResetBlocks), lastReset)
	})
}
