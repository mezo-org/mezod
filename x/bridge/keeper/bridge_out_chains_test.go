package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/mezo-org/mezod/x/bridge/types"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
	"github.com/stretchr/testify/require"
)

// clearBridgeOutChains empties the bridge-out chain set that mockContext
// seeds. A network that never ran the seeding step has an empty set.
func clearBridgeOutChains(ctx sdk.Context, k Keeper) {
	for _, chain := range k.GetBridgeOutChains(ctx) {
		k.DisableBridgeOutChain(ctx, chain)
	}
}

func TestBridgeOutChainEnableAndDisable(t *testing.T) {
	ctx, k := mockContext()
	clearBridgeOutChains(ctx, k)

	require.False(t, k.IsBridgeOutChainEnabled(ctx, types.TargetChainEthereum))
	require.False(t, k.IsBridgeOutChainEnabled(ctx, types.TargetChainBitcoin))

	k.EnableBridgeOutChain(ctx, types.TargetChainBitcoin)
	require.True(t, k.IsBridgeOutChainEnabled(ctx, types.TargetChainBitcoin))

	// Each chain is independent from the other ones.
	require.False(t, k.IsBridgeOutChainEnabled(ctx, types.TargetChainEthereum))

	k.EnableBridgeOutChain(ctx, types.TargetChainEthereum)
	require.True(t, k.IsBridgeOutChainEnabled(ctx, types.TargetChainEthereum))

	k.DisableBridgeOutChain(ctx, types.TargetChainBitcoin)
	require.False(t, k.IsBridgeOutChainEnabled(ctx, types.TargetChainBitcoin))
	require.True(t, k.IsBridgeOutChainEnabled(ctx, types.TargetChainEthereum))
}

func TestBridgeOutChainEnableIsIdempotent(t *testing.T) {
	ctx, k := mockContext()
	clearBridgeOutChains(ctx, k)

	k.EnableBridgeOutChain(ctx, types.TargetChainBitcoin)
	k.EnableBridgeOutChain(ctx, types.TargetChainBitcoin)

	require.True(t, k.IsBridgeOutChainEnabled(ctx, types.TargetChainBitcoin))
	require.Equal(t, []uint8{types.TargetChainBitcoin}, k.GetBridgeOutChains(ctx))
}

func TestBridgeOutChainDisableIsIdempotent(t *testing.T) {
	ctx, k := mockContext()
	clearBridgeOutChains(ctx, k)

	k.EnableBridgeOutChain(ctx, types.TargetChainEthereum)

	// Disabling a chain that is not in the set leaves the set untouched.
	k.DisableBridgeOutChain(ctx, types.TargetChainBitcoin)
	require.Equal(t, []uint8{types.TargetChainEthereum}, k.GetBridgeOutChains(ctx))

	k.DisableBridgeOutChain(ctx, types.TargetChainEthereum)
	k.DisableBridgeOutChain(ctx, types.TargetChainEthereum)

	require.False(t, k.IsBridgeOutChainEnabled(ctx, types.TargetChainEthereum))
	require.Empty(t, k.GetBridgeOutChains(ctx))
}

func TestGetBridgeOutChainsIsEmptyForAnUnseededSet(t *testing.T) {
	ctx, k := mockContext()
	clearBridgeOutChains(ctx, k)

	require.Empty(t, k.GetBridgeOutChains(ctx))
}

func TestGetBridgeOutChainsIsSortedAscending(t *testing.T) {
	ctx, k := mockContext()
	clearBridgeOutChains(ctx, k)

	// The insertion order must not leak into the listing.
	for _, chain := range []uint8{200, 1, 255, 0, 17} {
		k.EnableBridgeOutChain(ctx, chain)
	}

	require.Equal(t, []uint8{0, 1, 17, 200, 255}, k.GetBridgeOutChains(ctx))
}

func TestBridgeOutChainAtTheByteRangeBounds(t *testing.T) {
	ctx, k := mockContext()
	clearBridgeOutChains(ctx, k)

	k.EnableBridgeOutChain(ctx, 0)
	k.EnableBridgeOutChain(ctx, 255)

	require.True(t, k.IsBridgeOutChainEnabled(ctx, 0))
	require.True(t, k.IsBridgeOutChainEnabled(ctx, 255))
	require.Equal(t, []uint8{0, 255}, k.GetBridgeOutChains(ctx))

	k.DisableBridgeOutChain(ctx, 255)

	require.False(t, k.IsBridgeOutChainEnabled(ctx, 255))
	require.Equal(t, []uint8{0}, k.GetBridgeOutChains(ctx))
}

func TestGetBridgeOutChainsIgnoresOtherStoreEntries(t *testing.T) {
	ctx, k := mockContext()
	clearBridgeOutChains(ctx, k)

	// Entries below and above the bridge-out chain key prefix. The listing
	// must not pick them up.
	k.SetBridgeInPaused(ctx, true)
	k.AllowTripartyController(
		ctx,
		evmtypes.HexAddressToBytes("0x1111111111111111111111111111111111111111"),
		true,
	)

	require.Empty(t, k.GetBridgeOutChains(ctx))

	k.EnableBridgeOutChain(ctx, types.TargetChainBitcoin)

	require.Equal(t, []uint8{types.TargetChainBitcoin}, k.GetBridgeOutChains(ctx))
}
