//nolint:revive,stylecheck
package v0_3

import (
	store "cosmossdk.io/store/types"
	"github.com/mezo-org/mezod/app/upgrades"
	"github.com/mezo-org/mezod/utils"
	marketmaptypes "github.com/skip-mev/connect/v2/x/marketmap/types"
	oracletypes "github.com/skip-mev/connect/v2/x/oracle/types"
)

const (
	// UpgradeName defines the name of the upgrade.
	UpgradeName = "v0.3.0"
	// UpgradeInfo defines the binaries that will be used for the upgrade.
	UpgradeInfo = `'{"binaries":{"linux/amd64":"https://artifactregistry.googleapis.com/download/v1/projects/mezo-test-420708/locations/us-central1/repositories/mezo-staging-binary-public/files/mezod:v0.3.0-rc0:linux-amd64.tar.gz:download?alt=media"}}'`
	// TestnetUpgradeHeight defines the block height at which the upgrade is
	// triggered on testnet. This is Dec 17th 2024, around 11:00 AM UTC.
	// https://explorer.test.mezo.org/block/countdown/1093500
	TestnetUpgradeHeight = 1_093_500
)

var upgradeHeight = func(chainID string) int64 {
	if utils.IsTestnet(chainID) {
		return TestnetUpgradeHeight
	}

	return -1
}

var Fork = upgrades.Fork{
	UpgradeName:    UpgradeName,
	UpgradeHeight:  upgradeHeight,
	BeginForkLogic: RunForkLogic,
}

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added: []string{
			marketmaptypes.StoreKey,
			oracletypes.StoreKey,
		},
	},
}
