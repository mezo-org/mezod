package keeper

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
	"github.com/stretchr/testify/require"
)

func TestGetSourceBTCToken(t *testing.T) {
	ctx, k := mockContext()
	expectedToken := testSourceBTCToken

	k.SetSourceBTCToken(ctx, evmtypes.HexAddressToBytes(expectedToken))
	actualToken := k.GetSourceBTCToken(ctx)
	require.Equal(t, expectedToken, evmtypes.BytesToHexAddress(actualToken))
}

func TestBTCMint(t *testing.T) {
	ctx, k := mockContext()

	t.Run("test migration when there's no existing storage", func(t *testing.T) {
		// first get the storage before it's ever set
		k.bankKeeper.(*mockBankKeeper).
			On("GetSupply", ctx, evmtypes.DefaultEVMDenom).
			Return(sdk.NewCoin(evmtypes.DefaultEVMDenom, math.NewInt(0))).
			Times(1)

		amount := k.GetBTCMinted(ctx)
		require.Equal(t, amount, math.NewInt(0))
	})

	t.Run("can increase successfully", func(t *testing.T) {
		// then try to increase, and get it again
		require.NoError(t, k.IncreaseBTCMinted(ctx, math.NewInt(10)))
	})

	t.Run("can get successfully", func(t *testing.T) {
		// now get the value
		require.Equal(t, k.GetBTCMinted(ctx), math.NewInt(10))
	})
}

func TestBTCBurnt(t *testing.T) {
	ctx, k := mockContext()

	t.Run("test before the storage is initialized", func(t *testing.T) {
		amount := k.GetBTCBurnt(ctx)
		require.Equal(t, amount, math.NewInt(0))
	})

	t.Run("can increase successfully", func(t *testing.T) {
		// then try to increase, and get it again
		require.NotPanics(t, func() { require.NoError(t, k.IncreaseBTCBurnt(ctx, math.NewInt(10))) })
	})

	t.Run("can get successfully", func(t *testing.T) {
		// now get the value
		require.Equal(t, k.GetBTCBurnt(ctx), math.NewInt(10))
	})
}
