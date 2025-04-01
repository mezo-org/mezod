//nolint:revive,stylecheck
package v1_0

import (
	store "cosmossdk.io/store/types"
	"github.com/mezo-org/mezod/app/upgrades"
)

const (
	// UpgradeName defines the name of the upgrade.
	UpgradeName = "v1.0.0"
	// UpgradeInfo defines the binaries that will be used for the upgrade.
	UpgradeInfo = `'{"binaries":{"linux/amd64":"https://artifactregistry.googleapis.com/download/v1/projects/mezo-test-420708/locations/us-central1/repositories/mezo-staging-binary-public/files/mezod:v1.0.0-rc0:linux-amd64.tar.gz:download?alt=media"}}'`
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades:        store.StoreUpgrades{},
}
