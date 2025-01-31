package keeper

import (
	"testing"

	"github.com/mezo-org/mezod/x/bridge/types"
	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	ctx, k := mockContext()

	genesisState := types.DefaultGenesis()

	k.InitGenesis(ctx, *genesisState, nil)

	got := k.ExportGenesis(ctx)

	require.NotNil(t, got)
	require.EqualValues(t, genesisState, got)
}
