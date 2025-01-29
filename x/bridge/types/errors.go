package types

import (
	sdkerrors "cosmossdk.io/errors"
)

var (
	ErrInvalidEVMAddress  = sdkerrors.Register(ModuleName, 1, "invalid hex-encoded EVM address")
	ErrAlreadyMapping     = sdkerrors.Register(ModuleName, 2, "given ERC20 mapping already exists")
	ErrNotMapping         = sdkerrors.Register(ModuleName, 3, "given ERC20 mapping does not exist")
	ErrMaxMappingsReached = sdkerrors.Register(ModuleName, 4, "the maximum number of ERC20 mappings has been reached")
)
