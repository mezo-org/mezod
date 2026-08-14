//nolint:revive,stylecheck
package v13_0_test

import (
	"bytes"
	"slices"
	"testing"
	"time"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/mezo-org/mezod/app"
	v13_0 "github.com/mezo-org/mezod/app/upgrades/v13_0"
	"github.com/mezo-org/mezod/crypto/ethsecp256k1"
	"github.com/mezo-org/mezod/testutil"
	bridgetypes "github.com/mezo-org/mezod/x/bridge/types"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
	"github.com/stretchr/testify/require"
)

var testPauser = sdk.AccAddress(
	evmtypes.HexAddressToBytes("0x40C7b9612B394212394Ea860caCd0E176CA4ae5b"),
)

// Raw keys of the retired pause state. The types package no longer declares
// them; the migration inlines them the same way.
var (
	retiredPauserKey         = []byte{0x93}
	retiredTripartyPausedKey = []byte{0xA1}
)

func setupApp(t *testing.T) (*app.Mezo, sdk.Context) {
	t.Helper()

	privCons, err := ethsecp256k1.GenerateKey()
	require.NoError(t, err)
	consAddress := sdk.ConsAddress(privCons.PubKey().Address())

	mezoApp := app.Setup(false, nil)
	header := testutil.NewHeader(
		1, time.Now().UTC(), "mezo_31612-1", consAddress, nil, nil,
	)

	return mezoApp, mezoApp.BaseApp.NewContextLegacy(false, header)
}

// bridgeStore returns the raw x/bridge store. The migrations operate on keys
// that no longer have keeper accessors outside the migration path.
func bridgeStore(mezoApp *app.Mezo, ctx sdk.Context) storetypes.KVStore {
	return ctx.KVStore(mezoApp.GetKey(bridgetypes.StoreKey))
}

func precompileVersion(
	t *testing.T,
	mezoApp *app.Mezo,
	ctx sdk.Context,
	precompileAddress string,
) uint32 {
	t.Helper()

	params := mezoApp.EvmKeeper.GetParams(ctx)

	index := slices.IndexFunc(
		params.PrecompilesVersions,
		func(versionInfo *evmtypes.PrecompileVersionInfo) bool {
			return bytes.Equal(
				evmtypes.HexAddressToBytes(versionInfo.PrecompileAddress),
				evmtypes.HexAddressToBytes(precompileAddress),
			)
		},
	)
	require.NotEqual(t, -1, index)

	return params.PrecompilesVersions[index].Version
}

func setPrecompileVersion(
	t *testing.T,
	mezoApp *app.Mezo,
	ctx sdk.Context,
	precompileAddress string,
	version uint32,
) {
	t.Helper()

	params := mezoApp.EvmKeeper.GetParams(ctx)

	index := slices.IndexFunc(
		params.PrecompilesVersions,
		func(versionInfo *evmtypes.PrecompileVersionInfo) bool {
			return bytes.Equal(
				evmtypes.HexAddressToBytes(versionInfo.PrecompileAddress),
				evmtypes.HexAddressToBytes(precompileAddress),
			)
		},
	)
	require.NotEqual(t, -1, index)

	params.PrecompilesVersions[index].Version = version
	require.NoError(t, mezoApp.EvmKeeper.SetParams(ctx, params))
}

func TestUpdateMaintenancePrecompileVersion(t *testing.T) {
	mezoApp, ctx := setupApp(t)

	setPrecompileVersion(t, mezoApp, ctx, evmtypes.MaintenancePrecompileAddress, 5)

	require.NoError(t, v13_0.UpdateMaintenancePrecompileVersion(ctx, mezoApp.EvmKeeper))

	require.EqualValues(
		t,
		6,
		precompileVersion(t, mezoApp, ctx, evmtypes.MaintenancePrecompileAddress),
	)
}

func TestUpdateAssetsBridgePrecompileVersion(t *testing.T) {
	mezoApp, ctx := setupApp(t)

	setPrecompileVersion(t, mezoApp, ctx, evmtypes.AssetsBridgePrecompileAddress, 5)

	require.NoError(t, v13_0.UpdateAssetsBridgePrecompileVersion(ctx, mezoApp.EvmKeeper))

	require.EqualValues(
		t,
		6,
		precompileVersion(t, mezoApp, ctx, evmtypes.AssetsBridgePrecompileAddress),
	)

	// The other precompile versions must stay untouched.
	require.EqualValues(
		t,
		evmtypes.BTCTokenPrecompileLatestVersion,
		precompileVersion(t, mezoApp, ctx, evmtypes.BTCTokenPrecompileAddress),
	)
}

func TestMigrateEmergencyTeam(t *testing.T) {
	t.Run("pauser is set", func(t *testing.T) {
		mezoApp, ctx := setupApp(t)

		bridgeStore(mezoApp, ctx).Set(retiredPauserKey, testPauser.Bytes())
		bridgeStore(mezoApp, ctx).Set(retiredTripartyPausedKey, []byte{0x01})

		require.NoError(t, v13_0.MigrateEmergencyTeam(
			ctx,
			mezoApp.PoaKeeper,
			mezoApp.BridgeKeeper,
		))

		// The pauser becomes the emergency team so the pause capability stays
		// continuous through the upgrade.
		require.Equal(t, testPauser, mezoApp.PoaKeeper.GetEmergencyTeam(ctx))
		require.False(t, bridgeStore(mezoApp, ctx).Has(retiredPauserKey))

		// The retired triparty paused flag is dropped, not carried over to the
		// lockdown flags.
		require.False(t, bridgeStore(mezoApp, ctx).Has(retiredTripartyPausedKey))
		require.False(t, mezoApp.BridgeKeeper.IsBridgeInPaused(ctx))
		require.False(t, mezoApp.BridgeKeeper.IsBridgeOutPaused(ctx))
	})

	t.Run("pauser is the zero address", func(t *testing.T) {
		mezoApp, ctx := setupApp(t)

		bridgeStore(mezoApp, ctx).Set(retiredPauserKey, make([]byte, 20))

		require.NoError(t, v13_0.MigrateEmergencyTeam(
			ctx,
			mezoApp.PoaKeeper,
			mezoApp.BridgeKeeper,
		))

		require.True(t, mezoApp.PoaKeeper.GetEmergencyTeam(ctx).Empty())
		require.False(t, bridgeStore(mezoApp, ctx).Has(retiredPauserKey))
	})

	t.Run("pauser is absent", func(t *testing.T) {
		mezoApp, ctx := setupApp(t)

		require.False(t, bridgeStore(mezoApp, ctx).Has(retiredPauserKey))

		require.NoError(t, v13_0.MigrateEmergencyTeam(
			ctx,
			mezoApp.PoaKeeper,
			mezoApp.BridgeKeeper,
		))

		require.True(t, mezoApp.PoaKeeper.GetEmergencyTeam(ctx).Empty())
	})
}
