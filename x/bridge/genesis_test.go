package bridge_test

import (
	"testing"

	keepertest "github.com/evmos/evmos/v12/testutil/keeper"
	"github.com/evmos/evmos/v12/x/bridge"
	"github.com/evmos/evmos/v12/x/bridge/types"
	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),
	}

	k, ctx := keepertest.BridgeKeeper(t)
	bridge.InitGenesis(ctx, *k, genesisState)
	got := bridge.ExportGenesis(ctx, *k)
	require.NotNil(t, got)
}
