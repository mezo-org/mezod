package types

import (
	sdkerrors "cosmossdk.io/errors"
)

var (
	ErrInvalidEVMAddress               = sdkerrors.Register(ModuleName, 1, "invalid hex-encoded EVM address")
	ErrZeroEVMAddress                  = sdkerrors.Register(ModuleName, 2, "zero EVM address")
	ErrAlreadyMapping                  = sdkerrors.Register(ModuleName, 3, "given ERC20 mapping already exists")
	ErrNotMapping                      = sdkerrors.Register(ModuleName, 4, "given ERC20 mapping does not exist")
	ErrMaxMappingsReached              = sdkerrors.Register(ModuleName, 5, "the maximum number of ERC20 mappings has been reached")
	ErrTokenNotContract                = sdkerrors.Register(ModuleName, 6, "token address is not a contract")
	ErrOutflowLimitExceeded            = sdkerrors.Register(ModuleName, 7, "outflow limit exceeded")
	ErrTripartyWindowLimitExceeded     = sdkerrors.Register(ModuleName, 8, "triparty window limit exceeded")
	ErrTripartyPerRequestLimitExceeded = sdkerrors.Register(ModuleName, 9, "triparty per-request limit exceeded")
	ErrTripartyControllerNotAllowed    = sdkerrors.Register(ModuleName, 10, "controller is not an allowed triparty controller")
	ErrTripartyPaused                  = sdkerrors.Register(ModuleName, 11, "triparty bridging is paused")
	ErrTripartyCallbackDataTooLarge    = sdkerrors.Register(ModuleName, 12, "triparty callback data exceeds maximum length")
	ErrTripartyAmountNotPositive       = sdkerrors.Register(ModuleName, 13, "triparty amount must be positive")
)
