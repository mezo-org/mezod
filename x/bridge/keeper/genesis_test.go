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
	genesisState.TripartyPaused = true
	genesisState.TripartyBlockDelay = 10
	genesisState.TripartyPerRequestLimit = sdkmath.NewInt(123)
	genesisState.TripartyWindowLimit = sdkmath.NewInt(456)
	genesisState.TripartyRequestSequenceTip = sdkmath.NewInt(2)
	genesisState.TripartyProcessedSequenceTip = sdkmath.NewInt(1)
	genesisState.TripartyPendingRequests = []*types.TripartyBridgeRequest{
		{
			Sequence:     sdkmath.NewInt(1),
			BlockHeight:  100,
			Recipient:    "0x3333333333333333333333333333333333333333",
			Amount:       sdkmath.NewInt(50),
			CallbackData: []byte("callback-1"),
			Controller:   "0x1111111111111111111111111111111111111111",
		},
		{
			Sequence:     sdkmath.NewInt(2),
			BlockHeight:  101,
			Recipient:    "0x4444444444444444444444444444444444444444",
			Amount:       sdkmath.NewInt(75),
			CallbackData: []byte("callback-2"),
			Controller:   "0x2222222222222222222222222222222222222222",
		},
	}
	genesisState.TripartyWindowConsumed = sdkmath.NewInt(125)
	genesisState.TripartyWindowLastReset = 500
	genesisState.TripartyTotalBtcMinted = sdkmath.NewInt(200)

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
