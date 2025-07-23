package app

import (
	"github.com/spf13/cast"

	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/mezo-org/mezod/ethereum/sidecar"
	srvflags "github.com/mezo-org/mezod/server/flags"
	bridgekeeper "github.com/mezo-org/mezod/x/bridge/keeper"
)

// initializeBridgeOutServer initialized bridge-out server handling requests
// for `AssetsUnlocked` entries.
func (app *Mezo) initializeBridgeOutServer(
	appOpts servertypes.AppOptions,
	bridgeKeeper *bridgekeeper.Keeper,
) {
	sidecar.RunBridgeOutServer(
		app.Logger().With("server", "bridge-out"),
		cast.ToString(appOpts.Get(srvflags.BridgeOutServerAddress)),
		bridgeKeeper,
	)
}
