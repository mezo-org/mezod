package keeper

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/mezo-org/mezod/x/bridge/types"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
	"github.com/stretchr/testify/require"
)

func TestBridgeOutPaused(t *testing.T) {
	ctx, k := mockContext()

	require.False(t, k.IsBridgeOutPaused(ctx))

	k.SetBridgeOutPaused(ctx, true)
	require.True(t, k.IsBridgeOutPaused(ctx))

	// The bridge-out flag is independent from the bridge-in flag.
	require.False(t, k.IsBridgeInPaused(ctx))

	k.SetBridgeOutPaused(ctx, false)
	require.False(t, k.IsBridgeOutPaused(ctx))
}

func TestBridgeInPaused(t *testing.T) {
	ctx, k := mockContext()

	require.False(t, k.IsBridgeInPaused(ctx))

	k.SetBridgeInPaused(ctx, true)
	require.True(t, k.IsBridgeInPaused(ctx))

	// The bridge-in flag is independent from the bridge-out flag.
	require.False(t, k.IsBridgeOutPaused(ctx))

	k.SetBridgeInPaused(ctx, false)
	require.False(t, k.IsBridgeInPaused(ctx))
}

func TestBridgeOutPausedKeepsOutflowLimits(t *testing.T) {
	ctx, k := mockContext()

	btcToken := evmtypes.HexAddressToBytes(evmtypes.BTCTokenPrecompileAddress)
	k.SetOutflowLimit(ctx, btcToken, math.NewInt(1000))

	erc20Token := evmtypes.HexAddressToBytes("0x546758f4C2EfA4f37d66fF53644170F1d27AA1A0")
	k.setERC20TokenMapping(ctx, &types.ERC20TokenMapping{
		SourceToken: "0xac7f043Cf1BF10143926CC0035dBc46999512732",
		MezoToken:   evmtypes.BytesToHexAddress(erc20Token),
	})
	k.SetOutflowLimit(ctx, erc20Token, math.NewInt(2000))

	k.SetBridgeOutPaused(ctx, true)

	// Pausing must not destroy the outflow limit configuration.
	require.Equal(t, math.NewInt(1000), k.GetOutflowLimit(ctx, btcToken))
	require.Equal(t, math.NewInt(2000), k.GetOutflowLimit(ctx, erc20Token))

	k.SetBridgeOutPaused(ctx, false)

	require.Equal(t, math.NewInt(1000), k.GetOutflowLimit(ctx, btcToken))
	require.Equal(t, math.NewInt(2000), k.GetOutflowLimit(ctx, erc20Token))
}
