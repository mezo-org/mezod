package keeper

import (
	"github.com/evmos/evmos/v12/x/dualstaking/types"
)

var _ types.QueryServer = Keeper{}
