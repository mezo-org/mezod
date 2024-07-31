package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	storetypes "cosmossdk.io/store/types"

	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v12/x/poa/types"
)

// Keeper of the poa store
type Keeper struct {
	storeKey          storetypes.StoreKey
	cdc               codec.BinaryCodec
	historicalEntries uint32
}

// NewKeeper creates a poa keeper
func NewKeeper(
	storeKey storetypes.StoreKey,
	cdc codec.BinaryCodec,
) Keeper {
	return Keeper{
		storeKey:          storeKey,
		cdc:               cdc,
		historicalEntries: types.DefaultHistoricalEntries,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// checkOwner checks if the sender is the validator pool owner.
// Returns an error if the sender is not the owner or either of the compared
// addresses is empty. Returns nil otherwise.
func (k Keeper) checkOwner(ctx sdk.Context, sender sdk.AccAddress) error {
	return checkAccount(sender, k.GetOwner(ctx), "owner")
}

// checkCandidateOwner checks if the sender is the candidate validator pool owner.
// Returns an error if the sender is not the candidate owner or either of the
// compared addresses is empty. Returns nil otherwise.
func (k Keeper) checkCandidateOwner(ctx sdk.Context, sender sdk.AccAddress) error {
	candidateOwner := k.GetCandidateOwner(ctx)

	// Fail fast with an explicit error if the ownership transfer is not initialized.
	if candidateOwner.Empty() {
		return types.ErrOwnershipTransferNotInitialized
	}

	return checkAccount(sender, candidateOwner, "candidate owner")
}

// checkValidatorOperator checks if the sender is the operator of the given validator.
// Returns an error if the sender is not the operator or either of the
// compared addresses is empty. Returns nil otherwise.
func (k Keeper) checkValidatorOperator(
	sender sdk.AccAddress,
	validator types.Validator,
) error {
	// validator.GetOperator() returns a ValAddress while sender is an AccAddress.
	// We need to convert the operator address to an AccAddress in order to
	// obtain the same Bech32 prefix during string conversion as the sender.
	return checkAccount(
		sender,
		sdk.AccAddress(validator.GetOperator()),
		"validator operator",
	)
}

// checkAccount checks if the sender is the expected account.
// Returns an error if the sender is not the expected account or either of the
// compared addresses is empty. Returns nil otherwise.
func checkAccount(
	sender sdk.AccAddress,
	expected sdk.AccAddress,
	accName string,
) error {
	if sender.Empty() {
		return errorsmod.Wrap(
			sdkerrors.ErrInvalidAddress,
			"sender address is empty",
		)
	}

	if expected.Empty() {
		return errorsmod.Wrapf(
			sdkerrors.ErrInvalidAddress,
			"%s address is empty",
			accName,
		)
	}

	expectedStr := expected.String()
	senderStr := sender.String()

	if expectedStr != senderStr {
		return errorsmod.Wrapf(
			sdkerrors.ErrUnauthorized,
			"not the %s; expected %s, sender %s",
			accName,
			expectedStr,
			senderStr,
		)
	}

	return nil
}
