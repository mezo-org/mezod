package keeper

import (
	"bytes"
	"testing"

	"cosmossdk.io/math"
	"github.com/mezo-org/mezod/x/bridge/types"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
	"github.com/stretchr/testify/require"
)

var (
	testPauser = evmtypes.HexAddressToBytes("0xAbCdEfAbCdEfAbCdEfAbCdEfAbCdEfAbCdEfAbCdEf")
	testCaller = evmtypes.HexAddressToBytes("0x1234567890AbcdEF1234567890aBcdef12345678")
)

func TestPauserManagement(t *testing.T) {
	ctx, k := mockContext()

	pauser := k.GetPauser(ctx)
	require.True(t, evmtypes.IsZeroHexAddress(evmtypes.BytesToHexAddress(pauser)))

	k.SetPauser(ctx, testPauser)

	pauser = k.GetPauser(ctx)
	require.True(t, bytes.Equal(testPauser, pauser))

	k.SetPauser(ctx, nil)

	pauser = k.GetPauser(ctx)
	require.True(t, evmtypes.IsZeroHexAddress(evmtypes.BytesToHexAddress(pauser)))
}

func TestPauseBridgeOut(t *testing.T) {
	ctx, k := mockContext()

	t.Run("should fail when no pauser is set", func(t *testing.T) {
		err := k.PauseBridgeOut(ctx, testCaller)
		require.Error(t, err)
		require.Contains(t, err.Error(), "no pauser is set")
	})

	t.Run("should fail when pauser is zero address", func(t *testing.T) {
		k.SetPauser(ctx, evmtypes.HexAddressToBytes(evmtypes.ZeroHexAddress()))
		err := k.PauseBridgeOut(ctx, testCaller)
		require.Error(t, err)
		require.Contains(t, err.Error(), "no pauser is set")
	})

	t.Run("should fail when caller is not the pauser", func(t *testing.T) {
		k.SetPauser(ctx, testPauser)
		err := k.PauseBridgeOut(ctx, testCaller)
		require.Error(t, err)
		require.Contains(t, err.Error(), "caller is not the pauser")
	})

	t.Run("should succeed when caller is the pauser", func(t *testing.T) {
		k.SetPauser(ctx, testPauser)

		btcToken := evmtypes.HexAddressToBytes(evmtypes.BTCTokenPrecompileAddress)
		k.SetOutflowLimit(ctx, btcToken, math.NewInt(1000))

		testERC20Token := evmtypes.HexAddressToBytes("0x546758f4C2EfA4f37d66fF53644170F1d27AA1A0")

		k.setERC20TokenMapping(ctx, &types.ERC20TokenMapping{
			SourceToken: "0xac7f043Cf1BF10143926CC0035dBc46999512732",
			MezoToken:   evmtypes.BytesToHexAddress(testERC20Token),
		})
		k.SetOutflowLimit(ctx, testERC20Token, math.NewInt(2000))

		err := k.PauseBridgeOut(ctx, testPauser)
		require.NoError(t, err)

		btcLimit := k.GetOutflowLimit(ctx, btcToken)
		require.True(t, btcLimit.IsZero())

		mezoLimit := k.GetOutflowLimit(ctx, testERC20Token)
		require.True(t, mezoLimit.IsZero())
	})
}
