package types

import (
	errorsmod "cosmossdk.io/errors"
)

var (
	ErrNoValidatorFound                = errorsmod.Register(ModuleName, 1, "no validator found")
	ErrInvalidValidator                = errorsmod.Register(ModuleName, 2, "invalid validator")
	ErrAlreadyApplying                 = errorsmod.Register(ModuleName, 3, "the candidate is already applying to become a validator")
	ErrAlreadyValidator                = errorsmod.Register(ModuleName, 4, "the candidate is already a validator")
	ErrMaxValidatorsReached            = errorsmod.Register(ModuleName, 5, "the maximum number of validators has been reached")
	ErrNoApplicationFound              = errorsmod.Register(ModuleName, 6, "no application found")
	ErrNotValidator                    = errorsmod.Register(ModuleName, 7, "not a validator")
	ErrWrongValidatorState             = errorsmod.Register(ModuleName, 8, "wrong validator state")
	ErrOnlyOneValidator                = errorsmod.Register(ModuleName, 9, "there is only one validator in the validator set")
	ErrOwnershipTransferNotInitialized = errorsmod.Register(ModuleName, 10, "pool ownership transfer not initialized")
)
