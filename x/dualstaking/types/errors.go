package types

import (
	errorsmod "cosmossdk.io/errors"
)

// x/dualstaking module sentinel errors
var (
	ErrStakeNotFound = errorsmod.Register(ModuleName, 2, "stake not found")
	ErrDelegationNotFound = errorsmod.Register(ModuleName, 3, "delegation not found")
)
