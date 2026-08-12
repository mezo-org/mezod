//nolint:revive,stylecheck
package v13_0

import (
	store "cosmossdk.io/store/types"
	"github.com/mezo-org/mezod/app/upgrades"
)

// UpgradeName defines the name of the upgrade.
const UpgradeName = "v13.0.0"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades:        store.StoreUpgrades{},
}
