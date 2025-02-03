package keeper

import (
	"context"
	"testing"

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

