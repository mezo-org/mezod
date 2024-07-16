package types

import (
	errorsmod "cosmossdk.io/errors"
)

var ErrInvalidSigner = errorsmod.Register(ModuleName, 1, "invalid signer")
