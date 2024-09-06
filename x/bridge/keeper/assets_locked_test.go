package keeper

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"
)

func TestGetAssetsLockedSequenceTip(t *testing.T) {
	ctx, k := mockContext()

	newTip := math.NewInt(100)
	k.SetAssetsLockedSequenceTip(ctx, newTip)

	require.EqualValues(t, newTip, k.GetAssetsLockedSequenceTip(ctx))
}
