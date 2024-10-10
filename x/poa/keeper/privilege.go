package keeper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	storetypes "cosmossdk.io/store/types"

	sdkerrors "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/mezo-org/mezod/x/poa/types"
)

// AddPrivilege adds the privilege to a set of operators.
//
// The function returns an error if any of the following occurs:
// - The sender is not the current owner.
// - No privilege is provided.
// - The operators list is empty.
// - The operators list contains an address that is not a validator.
// - The operators list contains an address that already has the privilege.
// - The operators list contains duplicate addresses.
//
// Upstream is responsible for setting the `sender` parameter to the actual
// actor performing the operation. If the sender address is empty, the function
// will return an error.
func (k Keeper) AddPrivilege(
	ctx sdk.Context,
	sender sdk.AccAddress,
	operators []sdk.ValAddress,
	privilege string,
) error {
	if err := k.checkOwner(ctx, sender); err != nil {
		return err
	}

	if len(operators) == 0 {
		return fmt.Errorf("no operators provided")
	}

	if len(privilege) == 0 {
		return fmt.Errorf("no privilege provided")
	}

	consAddrs := k.GetValidatorsConsAddrsByPrivilege(ctx, privilege)

	for _, operator := range operators {
		validator, found := k.GetValidator(ctx, operator)
		if !found {
			return sdkerrors.Wrapf(
				types.ErrNotValidator,
				"operator: %s",
				operator.String(),
			)
		}

		validatorConsAddr := validator.GetConsAddress()

		existingPrivilege := slices.ContainsFunc(
			consAddrs,
			func(consAddr sdk.ConsAddress) bool {
				return consAddr.Equals(validatorConsAddr)
			},
		)
		if existingPrivilege {
			return sdkerrors.Wrapf(
				types.ErrExistingPrivilege,
				"operator: %s, privilege: %s",
				operator.String(),
				privilege,
			)
		}

		consAddrs = append(consAddrs, validatorConsAddr)
	}

	// Update the store.
	k.setValidatorsConsAddrsByPrivilege(ctx, privilege, consAddrs)

	return nil
}

// RemovePrivilege removes the privilege from a set of operators.
//
// The function returns an error if any of the following occurs:
// - The sender is not the current owner.
// - No privilege is provided.
// - The operators list is empty.
// - The operators list contains an address that is not a validator.
// - The operators list contains an address that doesn't have the privilege.
// - The operators list contains duplicate addresses.
//
// Upstream is responsible for setting the `sender` parameter to the actual
// actor performing the operation. If the sender address is empty, the function
// will return an error.
func (k Keeper) RemovePrivilege(
	ctx sdk.Context,
	sender sdk.AccAddress,
	operators []sdk.ValAddress,
	privilege string,
) error {
	if err := k.checkOwner(ctx, sender); err != nil {
		return err
	}

	if len(operators) == 0 {
		return fmt.Errorf("no operators provided")
	}

	if len(privilege) == 0 {
		return fmt.Errorf("no privilege provided")
	}

	consAddrs := k.GetValidatorsConsAddrsByPrivilege(ctx, privilege)

	for _, operator := range operators {
		validator, found := k.GetValidator(ctx, operator)
		if !found {
			return sdkerrors.Wrapf(
				types.ErrNotValidator,
				"operator: %s",
				operator.String(),
			)
		}

		validatorConsAddr := validator.GetConsAddress()

		index := slices.IndexFunc(
			consAddrs,
			func(consAddr sdk.ConsAddress) bool {
				return consAddr.Equals(validatorConsAddr)
			},
		)
		if index < 0 {
			return sdkerrors.Wrapf(
				types.ErrMissingPrivilege,
				"operator: %s, privilege: %s",
				operator.String(),
				privilege,
			)
		}

		// Remove the consensus address from the slice preserving order.
		consAddrs = append(consAddrs[:index], consAddrs[index+1:]...)
	}

	// Update the store.
	k.setValidatorsConsAddrsByPrivilege(ctx, privilege, consAddrs)

	return nil
}

// GetValidatorsOperatorsByPrivilege returns the operator addresses of
// all validators that are currently present in the store and have the
// given privilege. There is no guarantee that the returned validators
// are currently part of the CometBFT validator set.
func (k Keeper) GetValidatorsOperatorsByPrivilege(
	ctx sdk.Context,
	privilege string,
) []sdk.ValAddress {
	consAddrs := k.GetValidatorsConsAddrsByPrivilege(ctx, privilege)
	if len(consAddrs) == 0 {
		return nil
	}

	operators := make([]sdk.ValAddress, len(consAddrs))
	for i, consAddr := range consAddrs {
		validator, found := k.GetValidatorByConsAddr(ctx, consAddr)
		if !found {
			// If the given consensus address has the privilege, then the
			// corresponding validator should exist.
			panic("consensus address has privilege but corresponding validator not found")
		}

		operators[i] = validator.GetOperator()
	}

	return operators
}

// setValidatorsConsAddrsByPrivilege sets the consensus addresses of validators
// that have the given privilege. Overwrites any existing consensus addresses.
func (k Keeper) setValidatorsConsAddrsByPrivilege(
	ctx sdk.Context,
	privilege string,
	consAddrs []sdk.ConsAddress,
) {
	// Ensure predictable ordering of the consensus addresses.
	slices.SortFunc(consAddrs, func(a, b sdk.ConsAddress) int {
		return bytes.Compare(a.Bytes(), b.Bytes())
	})

	// Ensure that there are no duplicates.
	consAddrs = slices.CompactFunc(consAddrs, func(a, b sdk.ConsAddress) bool {
		return a.Equals(b)
	})

	consAddrsBytes, err := json.Marshal(consAddrs)
	if err != nil {
		// Should always be able to marshal the value.
		panic(err)
	}

	store := ctx.KVStore(k.storeKey)

	store.Set(types.GetValidatorsByPrivilegeKey(privilege), consAddrsBytes)
}

// GetValidatorsConsAddrsByPrivilege returns the consensus addresses of
// all validators that are currently present in the store and have the
// given privilege. There is no guarantee that the returned validators
// are currently part of the CometBFT validator set.
func (k Keeper) GetValidatorsConsAddrsByPrivilege(
	ctx sdk.Context,
	privilege string,
) []sdk.ConsAddress {
	store := ctx.KVStore(k.storeKey)

	value := store.Get(types.GetValidatorsByPrivilegeKey(privilege))
	if len(value) == 0 {
		return nil
	}

	var consAddrs []sdk.ConsAddress
	if err := json.Unmarshal(value, &consAddrs); err != nil {
		// Should always be able to unmarshal the value.
		panic(err)
	}

	return consAddrs
}

// getAllPrivileges returns all distinct privileges that are currently present
// in the store.
func (k Keeper) getAllPrivileges(ctx sdk.Context) []string {
	store := ctx.KVStore(k.storeKey)

	iterator := storetypes.KVStorePrefixIterator(
		store,
		types.ValidatorsByPrivilegeKeyPrefix,
	)
	defer func() {
		_ = iterator.Close()
	}()

	privileges := make([]string, 0)
	for ; iterator.Valid(); iterator.Next() {
		privilege := strings.TrimPrefix(
			string(iterator.Key()),
			string(types.ValidatorsByPrivilegeKeyPrefix),
		)
		privileges = append(privileges, privilege)
	}

	return privileges
}

// removeAllPrivileges removes all privileges from a validator.
func (k Keeper) removeAllPrivileges(
	ctx sdk.Context,
	validatorConsAddr sdk.ConsAddress,
) {
	privileges := k.getAllPrivileges(ctx)

	for _, privilege := range privileges {
		consAddrs := k.GetValidatorsConsAddrsByPrivilege(ctx, privilege)
		if len(consAddrs) == 0 {
			// The privilege does not have any validators, we can skip.
			continue
		}

		index := slices.IndexFunc(
			consAddrs,
			func(consAddr sdk.ConsAddress) bool {
				return consAddr.Equals(validatorConsAddr)
			},
		)
		if index < 0 {
			// The validator does not have the given privilege, we can skip.
			continue
		}

		// Remove the consensus address from the slice preserving order.
		consAddrs = append(consAddrs[:index], consAddrs[index+1:]...)

		// Update the store.
		k.setValidatorsConsAddrsByPrivilege(ctx, privilege, consAddrs)
	}
}
