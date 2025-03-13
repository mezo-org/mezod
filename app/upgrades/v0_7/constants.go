//nolint:revive,stylecheck
package v0_7

import (
	"github.com/mezo-org/mezod/app/upgrades"
	"github.com/mezo-org/mezod/utils"
)

const (
	// UpgradeName defines the name of the upgrade.
	UpgradeName = "v0.7.0"
	// TestnetUpgradeHeight defines the block height at which the upgrade is
	// triggered on testnet. No expected date as we are executing this fork on halted chain.
	// https://explorer.test.mezo.org/block/countdown/3078794
	TestnetUpgradeHeight = 3_078_794
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
