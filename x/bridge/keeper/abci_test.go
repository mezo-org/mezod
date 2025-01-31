package keeper

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
	"github.com/stretchr/testify/require"
)

func TestEndBlock(t *testing.T) {
	ctx, k := mockContext()

	// first increase the BTC minted
	require.NoError(t, k.IncreaseBTCsMinted(ctx, math.NewInt(42)))

	// ... BTC burnt
	require.NoError(t, k.IncreaseBTCsBurnt(ctx, math.NewInt(21)))

	t.Run("does not panic when valid state", func(t *testing.T) {
		// return the same supply so all is fine
		k.bankKeeper.(*mockBankKeeper).
			On("GetSupply", ctx, evmtypes.DefaultEVMDenom).
			Return(sdk.NewCoin(evmtypes.DefaultEVMDenom, math.NewInt(21))).
			Times(1)

		require.NoError(t, k.EndBlock(ctx))
	})

	t.Run("panics when state is invalid", func(t *testing.T) {
		// return the same supply so all is fine
		k.bankKeeper.(*mockBankKeeper).
			On("GetSupply", ctx, evmtypes.DefaultEVMDenom).
			Return(sdk.NewCoin(evmtypes.DefaultEVMDenom, math.NewInt(37))).
			Times(1)

		require.Panics(t, func() { _ = k.EndBlock(ctx) })
	})
}
