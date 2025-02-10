//nolint:revive,stylecheck
package v0_6

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/mezo-org/mezod/app/upgrades"
)

func RunForkLogic(_ sdk.Context, _ *upgrades.Keepers) {
	panic("v0.6.0 fork height reached; upgrade version to continue")
}
