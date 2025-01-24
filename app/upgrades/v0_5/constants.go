//nolint:revive,stylecheck
package v0_5

import (
	"github.com/mezo-org/mezod/app/upgrades"
	"github.com/mezo-org/mezod/utils"
)

const (
	// UpgradeName defines the name of the upgrade.
	UpgradeName = "v0.5.0"
	// TestnetUpgradeHeight defines the block height at which the upgrade is
	// triggered on testnet. This is Jan 30th 2025, around 12:00 PM UTC.
	// https://explorer.test.mezo.org/block/countdown/2120000
	TestnetUpgradeHeight = 2_120_000
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
