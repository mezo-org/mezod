package dualstaking_test

import (
	"testing"

	keepertest "github.com/evmos/evmos/v12/testutil/keeper"
	"github.com/evmos/evmos/v12/x/dualstaking"
	"github.com/evmos/evmos/v12/x/dualstaking/types"
	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
	}

	k, ctx := keepertest.DualstakingKeeper(t)
	dualstaking.InitGenesis(ctx, *k, genesisState)
	got := dualstaking.ExportGenesis(ctx, *k)
	require.NotNil(t, got)
}
