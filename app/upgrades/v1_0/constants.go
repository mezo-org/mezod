//nolint:revive,stylecheck
package v1_0

import (
	store "cosmossdk.io/store/types"
	"github.com/mezo-org/mezod/app/upgrades"
)

// UpgradeName defines the name of the upgrade.
const (
	UpgradeRC0Name = "v1.0.0"
	UpgradeRC1Name = "v1.0.0-rc1"
)

var UpgradeRC0 = upgrades.Upgrade{
	UpgradeName:          UpgradeRC0Name,
	CreateUpgradeHandler: CreateUpgradeHandlerRC0,
	StoreUpgrades:        store.StoreUpgrades{},
}

var UpgradeRC1 = upgrades.Upgrade{
	UpgradeName:          UpgradeRC1Name,
	CreateUpgradeHandler: CreateUpgradeHandlerRC1,
	StoreUpgrades:        store.StoreUpgrades{},
}