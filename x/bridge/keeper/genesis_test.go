package keeper

import (
	"context"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/mock"

	"github.com/mezo-org/mezod/x/bridge/types"
	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	ctx, k := mockContext()

	genesisState := types.DefaultGenesis()
	genesisState.SourceBtcToken = testSourceBTCToken

	accountKeeper := newMockAccountKeeper()
	accountKeeper.On(
		"GetModuleAccount",
		ctx,
		types.ModuleName,
	).Return(authtypes.NewEmptyModuleAccount(types.ModuleName))

	k.InitGenesis(ctx, *genesisState, accountKeeper)

	got := k.ExportGenesis(ctx)

	require.NotNil(t, got)
	require.EqualValues(t, genesisState, got)
	accountKeeper.AssertExpectations(t)
}

func TestGenesisTripartyState(t *testing.T) {
	ctx, k := mockContext()

	genesisState := types.DefaultGenesis()
	genesisState.SourceBtcToken = testSourceBTCToken
	genesisState.AllowedTripartyControllers = []string{
		"0x1111111111111111111111111111111111111111",
		"0x2222222222222222222222222222222222222222",
	}
	genesisState.TripartyBlockDelay = 10
	genesisState.TripartyPerRequestLimit = sdkmath.NewInt(123)
	genesisState.TripartyWindowLimit = sdkmath.NewInt(456)
	genesisState.TripartyRequestSequenceTip = sdkmath.NewInt(3)
	genesisState.TripartyProcessedSequenceTip = sdkmath.NewInt(1)
	genesisState.TripartyPendingRequests = []*types.TripartyBridgeRequest{
		{
			Sequence:     sdkmath.NewInt(2),
			BlockHeight:  100,
			Recipient:    "0x3333333333333333333333333333333333333333",
			Amount:       sdkmath.NewInt(50),
			CallbackData: []byte("callback-1"),
			Controller:   "0x1111111111111111111111111111111111111111",
		},
		{
			Sequence:     sdkmath.NewInt(3),
			BlockHeight:  101,
			Recipient:    "0x4444444444444444444444444444444444444444",
			Amount:       sdkmath.NewInt(75),
			CallbackData: []byte("callback-2"),
			Controller:   "0x2222222222222222222222222222222222222222",
		},
	}
	genesisState.TripartyWindowConsumed = sdkmath.NewInt(125)
	genesisState.TripartyWindowLastReset = 500
	genesisState.TripartyControllerBtcMinted = []*types.TripartyControllerBTCMinted{
		{
			Controller: "0x1111111111111111111111111111111111111111",
			Amount:     sdkmath.NewInt(100),
		},
		{
			Controller: "0x2222222222222222222222222222222222222222",
			Amount:     sdkmath.NewInt(100),
		},
	}

	accountKeeper := newMockAccountKeeper()
	accountKeeper.On(
		"GetModuleAccount",
		ctx,
		types.ModuleName,
	).Return(authtypes.NewEmptyModuleAccount(types.ModuleName))

	k.InitGenesis(ctx, *genesisState, accountKeeper)

	got := k.ExportGenesis(ctx)

	require.NotNil(t, got)
	require.EqualValues(t, genesisState, got)
	accountKeeper.AssertExpectations(t)
}

func TestGenesisLockdownFlags(t *testing.T) {
	tests := map[string]struct {
		bridgeInPaused  bool
		bridgeOutPaused bool
	}{
		"both flags unset":  {bridgeInPaused: false, bridgeOutPaused: false},
		"bridge-in paused":  {bridgeInPaused: true, bridgeOutPaused: false},
		"bridge-out paused": {bridgeInPaused: false, bridgeOutPaused: true},
		"both flags set":    {bridgeInPaused: true, bridgeOutPaused: true},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctx, k := mockContext()

			genesisState := types.DefaultGenesis()
			genesisState.SourceBtcToken = testSourceBTCToken
			genesisState.BridgeInPaused = test.bridgeInPaused
			genesisState.BridgeOutPaused = test.bridgeOutPaused

			accountKeeper := newMockAccountKeeper()
			accountKeeper.On(
				"GetModuleAccount",
				ctx,
				types.ModuleName,
			).Return(authtypes.NewEmptyModuleAccount(types.ModuleName))

			k.InitGenesis(ctx, *genesisState, accountKeeper)

			require.Equal(t, test.bridgeInPaused, k.IsBridgeInPaused(ctx))
			require.Equal(t, test.bridgeOutPaused, k.IsBridgeOutPaused(ctx))

			got := k.ExportGenesis(ctx)

			require.NotNil(t, got)
			require.EqualValues(t, genesisState, got)
			accountKeeper.AssertExpectations(t)
		})
	}
}

func TestGenesisBridgeOutChains(t *testing.T) {
	tests := map[string]struct {
		bridgeOutChains []uint32
		expectedSet     []uint8
		expectedExport  []uint32
	}{
		"only Ethereum": {
			bridgeOutChains: []uint32{types.TargetChainEthereum},
			expectedSet:     []uint8{types.TargetChainEthereum},
			expectedExport:  []uint32{types.TargetChainEthereum},
		},
		"both target chains": {
			bridgeOutChains: []uint32{types.TargetChainEthereum, types.TargetChainBitcoin},
			expectedSet:     []uint8{types.TargetChainEthereum, types.TargetChainBitcoin},
			expectedExport:  []uint32{types.TargetChainEthereum, types.TargetChainBitcoin},
		},
		"unsorted entries": {
			bridgeOutChains: []uint32{types.TargetChainBitcoin, types.TargetChainEthereum},
			expectedSet:     []uint8{types.TargetChainEthereum, types.TargetChainBitcoin},
			expectedExport:  []uint32{types.TargetChainEthereum, types.TargetChainBitcoin},
		},
		"chain at the byte maximum": {
			bridgeOutChains: []uint32{types.TargetChainEthereum, 255},
			expectedSet:     []uint8{types.TargetChainEthereum, 255},
			expectedExport:  []uint32{types.TargetChainEthereum, 255},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctx, k := mockContext()
			// InitGenesis runs on a chain that has no bridge-out chain set.
			clearBridgeOutChains(ctx, k)

			genesisState := types.DefaultGenesis()
			genesisState.SourceBtcToken = testSourceBTCToken
			genesisState.BridgeOutChains = test.bridgeOutChains

			accountKeeper := newMockAccountKeeper()
			accountKeeper.On(
				"GetModuleAccount",
				ctx,
				types.ModuleName,
			).Return(authtypes.NewEmptyModuleAccount(types.ModuleName))

			k.InitGenesis(ctx, *genesisState, accountKeeper)

			for _, chain := range test.expectedSet {
				require.True(t, k.IsBridgeOutChainEnabled(ctx, chain))
			}
			require.Equal(t, test.expectedSet, k.GetBridgeOutChains(ctx))

			got := k.ExportGenesis(ctx)

			require.NotNil(t, got)
			// The export is sorted ascending, so an unsorted genesis does not
			// round-trip entry for entry.
			require.Equal(t, test.expectedExport, got.BridgeOutChains)
			accountKeeper.AssertExpectations(t)
		})
	}
}

func TestGenesisWithoutBridgeOutChains(t *testing.T) {
	ctx, k := mockContext()
	clearBridgeOutChains(ctx, k)

	// A genesis exported before the bridge-out chain set existed carries no
	// entry, which leaves every chain disabled.
	genesisState := types.DefaultGenesis()
	genesisState.SourceBtcToken = testSourceBTCToken
	genesisState.BridgeOutChains = nil

	accountKeeper := newMockAccountKeeper()
	accountKeeper.On(
		"GetModuleAccount",
		ctx,
		types.ModuleName,
	).Return(authtypes.NewEmptyModuleAccount(types.ModuleName))

	k.InitGenesis(ctx, *genesisState, accountKeeper)

	require.False(t, k.IsBridgeOutChainEnabled(ctx, types.TargetChainEthereum))
	require.False(t, k.IsBridgeOutChainEnabled(ctx, types.TargetChainBitcoin))
	require.Empty(t, k.GetBridgeOutChains(ctx))
	require.Empty(t, k.ExportGenesis(ctx).BridgeOutChains)
	accountKeeper.AssertExpectations(t)
}

type mockAccountKeeper struct {
	mock.Mock
}

func newMockAccountKeeper() *mockAccountKeeper {
	return &mockAccountKeeper{}
}

func (mak *mockAccountKeeper) GetModuleAccount(
	ctx context.Context,
	moduleName string,
) sdk.ModuleAccountI {
	args := mak.Called(ctx, moduleName)

	if res := args.Get(0); res != nil {
		return res.(sdk.ModuleAccountI)
	}

	return nil
}
