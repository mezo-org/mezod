//nolint:revive,stylecheck
package v0_5

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/mezo-org/mezod/app/upgrades"
)

func RunForkLogic(_ sdk.Context, _ *upgrades.Keepers) {
	panic("v0.5.0 fork height reached; upgrade version to continue")
}
