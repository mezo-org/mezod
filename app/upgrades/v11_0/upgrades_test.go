//nolint:revive,stylecheck
package v11_0_test

import (
	"testing"
	"time"

	"github.com/cometbft/cometbft/crypto/tmhash"
	"github.com/mezo-org/mezod/app"
	v11_0 "github.com/mezo-org/mezod/app/upgrades/v11_0"
	"github.com/mezo-org/mezod/testutil"
	"github.com/mezo-org/mezod/utils"
	"github.com/stretchr/testify/require"
)

func TestSetPragueTime(t *testing.T) {
	mezo := app.Setup(false, nil)

	header := testutil.NewHeader(
		1,
		time.Unix(1_800_000_000, 0).UTC(),
		utils.MainnetChainID+"-1",
		nil,
		tmhash.Sum([]byte("app")),
		tmhash.Sum([]byte("validators")),
	)
	ctx := mezo.NewContextLegacy(false, header)

	// A chain that started before the PragueTime field existed stores
	// nil for the field. Reproduce that state explicitly.
	params := mezo.EvmKeeper.GetParams(ctx)
	params.ChainConfig.PragueTime = nil
	require.NoError(t, mezo.EvmKeeper.SetParams(ctx, params))
	require.Nil(t, mezo.EvmKeeper.GetParams(ctx).ChainConfig.PragueTime)

	require.NoError(t, v11_0.SetPragueTimeForTest(ctx, mezo.EvmKeeper))

	updated := mezo.EvmKeeper.GetParams(ctx).ChainConfig.PragueTime
	require.NotNil(t, updated)
	require.Equal(t, uint64(1_800_000_000), updated.Uint64())
	require.NoError(t, mezo.EvmKeeper.GetParams(ctx).ChainConfig.Validate())
}
