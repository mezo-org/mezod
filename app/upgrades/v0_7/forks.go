//nolint:revive,stylecheck
package v0_7

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/mezo-org/mezod/app/upgrades"
)

func RunForkLogic(ctx sdk.Context, keepers *upgrades.Keepers) {
	panic("v0.7.0 fork height reached; upgrade version to continue")
}
