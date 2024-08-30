package keeper

import (
	"testing"

	"github.com/mezo-org/mezod/x/bridge/types"
	"github.com/stretchr/testify/require"
)

func TestGetParams(t *testing.T) {
	ctx, k := mockContext()
	params := types.DefaultParams()

	err := k.SetParams(ctx, params)
	if err != nil {
		t.Fatal(err)
	}

	require.EqualValues(t, params, k.GetParams(ctx))
}
