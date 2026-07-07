//nolint:revive,stylecheck
package v10_0

import (
	store "cosmossdk.io/store/types"
	"github.com/mezo-org/mezod/app/upgrades"
)

// UpgradeName defines the name of the upgrade.
const UpgradeName = "v10.0.0"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades:        store.StoreUpgrades{},
}
