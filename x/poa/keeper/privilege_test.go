package keeper

import (
	"bytes"
	"slices"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/mezo-org/mezod/x/poa/types"
	"github.com/stretchr/testify/require"
)

const privilege = "testPrivilege"

func TestAddPrivilege(t *testing.T) {
	helper, _ := mockValidator()
	owner := sdk.AccAddress(helper.GetOperator())

	helper, _ = mockValidator()
	thirdParty := sdk.AccAddress(helper.GetOperator())

	validator1, _ := mockValidator()
	validator2, _ := mockValidator()
	validator3, _ := mockValidator()
	nonValidator, _ := mockValidator()

	tests := []struct {
		name        string
		prepareFn   func(ctx sdk.Context, poaKeeper Keeper)
		sender      sdk.AccAddress
		operators   []sdk.ValAddress
		privilege   string
		postCheckFn func(ctx sdk.Context, poaKeeper Keeper)
		errContains string
	}{
		{
			name:        "no sender",
			prepareFn:   nil,
			sender:      sdk.AccAddress{},
			operators:   nil,
			privilege:   "",
			postCheckFn: nil,
			errContains: "sender address is empty",
		},
		{
			name:        "sender is not owner",
			prepareFn:   nil,
			sender:      thirdParty,
			operators:   nil,
			privilege:   "",
			postCheckFn: nil,
			errContains: "not the owner",
		},
		{
			name:        "nil operators",
			prepareFn:   nil,
			sender:      owner,
			operators:   nil,
			privilege:   "",
			postCheckFn: nil,
			errContains: "no operators provided",
		},
		{
			name:        "empty operators",
			prepareFn:   nil,
			sender:      owner,
			operators:   []sdk.ValAddress{},
			privilege:   "",
			postCheckFn: nil,
			errContains: "no operators provided",
		},
		{
			name:        "empty privilege",
			prepareFn:   nil,
			sender:      owner,
			operators:   []sdk.ValAddress{validator1.GetOperator()},
			privilege:   "",
			postCheckFn: nil,
			errContains: "no privilege provided",
		},
		{
			name:      "non-validator operator",
			prepareFn: nil,
			sender:    owner,
			operators: []sdk.ValAddress{
				validator1.GetOperator(),
				nonValidator.GetOperator(),
			},
			privilege:   privilege,
			postCheckFn: nil,
			errContains: types.ErrNotValidator.Error(),
		},
		{
			name:      "duplicate operator",
			prepareFn: nil,
			sender:    owner,
			operators: []sdk.ValAddress{
				validator1.GetOperator(),
				validator1.GetOperator(),
			},
			privilege:   privilege,
			postCheckFn: nil,
			errContains: types.ErrExistingPrivilege.Error(),
		},
		{
			name: "same privilege twice",
			prepareFn: func(ctx sdk.Context, poaKeeper Keeper) {
				err := poaKeeper.AddPrivilege(
					ctx,
					owner,
					[]sdk.ValAddress{validator2.GetOperator()},
					privilege,
				)
				require.NoError(t, err, "expected no error")
			},
			sender: owner,
			operators: []sdk.ValAddress{
				validator1.GetOperator(),
				validator2.GetOperator(),
			},
			privilege:   privilege,
			postCheckFn: nil,
			errContains: types.ErrExistingPrivilege.Error(),
		},
		{
			name:      "happy path - no prior privileges exist",
			prepareFn: nil,
			sender:    owner,
			operators: []sdk.ValAddress{
				validator1.GetOperator(),
				validator2.GetOperator(),
				validator3.GetOperator(),
			},
			privilege: privilege,
			postCheckFn: func(ctx sdk.Context, poaKeeper Keeper) {
				expectedValidators := []types.Validator{validator1, validator2, validator3}

				// The privilege's set consists of consensus addresses
				// sorted in ascending order lexicographically. Sort the
				// validators slice to match the expected order during comparison.
				slices.SortFunc(expectedValidators, func(a, b types.Validator) int {
					return bytes.Compare(a.GetConsAddress(), b.GetConsAddress())
				})

				expectedOperators := make([]sdk.ValAddress, len(expectedValidators))
				expectedConsAddrs := make([]sdk.ConsAddress, len(expectedValidators))
				for i, validator := range expectedValidators {
					expectedOperators[i] = validator.GetOperator()
					expectedConsAddrs[i] = validator.GetConsAddress()
				}

				require.Equal(
					t,
					expectedOperators,
					poaKeeper.GetValidatorsOperatorsByPrivilege(ctx, privilege),
					"expected operators mismatch",
				)

				require.Equal(
					t,
					expectedConsAddrs,
					poaKeeper.GetValidatorsConsAddrsByPrivilege(ctx, privilege),
					"expected consensus addresses mismatch",
				)
			},
			errContains: "",
		},
		{
			name: "happy path - prior privileges exist",
			prepareFn: func(ctx sdk.Context, poaKeeper Keeper) {
				err := poaKeeper.AddPrivilege(
					ctx,
					owner,
					[]sdk.ValAddress{validator2.GetOperator()},
					privilege,
				)
				require.NoError(t, err, "expected no error")
			},
			sender: owner,
			operators: []sdk.ValAddress{
				validator1.GetOperator(),
				validator3.GetOperator(),
			},
			privilege: privilege,
			postCheckFn: func(ctx sdk.Context, poaKeeper Keeper) {
				expectedValidators := []types.Validator{validator1, validator2, validator3}

				// The privilege's set consists of consensus addresses
				// sorted in ascending order lexicographically. Sort the
				// validators slice to match the expected order during comparison.
				slices.SortFunc(expectedValidators, func(a, b types.Validator) int {
					return bytes.Compare(a.GetConsAddress(), b.GetConsAddress())
				})

				expectedOperators := make([]sdk.ValAddress, len(expectedValidators))
				expectedConsAddrs := make([]sdk.ConsAddress, len(expectedValidators))
				for i, validator := range expectedValidators {
					expectedOperators[i] = validator.GetOperator()
					expectedConsAddrs[i] = validator.GetConsAddress()
				}

				require.Equal(
					t,
					expectedOperators,
					poaKeeper.GetValidatorsOperatorsByPrivilege(ctx, privilege),
					"expected operators mismatch",
				)

				require.Equal(
					t,
					expectedConsAddrs,
					poaKeeper.GetValidatorsConsAddrsByPrivilege(ctx, privilege),
					"expected consensus addresses mismatch",
				)
			},
			errContains: "",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx, poaKeeper := mockContext()

			poaKeeper.setOwner(ctx, owner)

			poaKeeper.setValidator(ctx, validator1)
			poaKeeper.setValidator(ctx, validator2)
			poaKeeper.setValidator(ctx, validator3)

			poaKeeper.setValidatorByConsAddr(ctx, validator1)
			poaKeeper.setValidatorByConsAddr(ctx, validator2)
			poaKeeper.setValidatorByConsAddr(ctx, validator3)

			if test.prepareFn != nil {
				test.prepareFn(ctx, poaKeeper)
			}

			err := poaKeeper.AddPrivilege(
				ctx,
				test.sender,
				test.operators,
				test.privilege,
			)

			if len(test.errContains) == 0 {
				require.NoError(t, err, "expected no error")
			} else {
				// ErrorContains checks if the error is non-nil so no need
				// for an explicit check here.
				require.ErrorContains(
					t,
					err,
					test.errContains,
					"expected different error message",
				)
			}

			if test.postCheckFn != nil {
				test.postCheckFn(ctx, poaKeeper)
			}
		})
	}
}

func TestRemovePrivilege(t *testing.T) {
	helper, _ := mockValidator()
	owner := sdk.AccAddress(helper.GetOperator())

	helper, _ = mockValidator()
	thirdParty := sdk.AccAddress(helper.GetOperator())

	validator1, _ := mockValidator()
	validator2, _ := mockValidator()
	validator3, _ := mockValidator()
	nonValidator, _ := mockValidator()

	tests := []struct {
		name        string
		prepareFn   func(ctx sdk.Context, poaKeeper Keeper)
		sender      sdk.AccAddress
		operators   []sdk.ValAddress
		privilege   string
		postCheckFn func(ctx sdk.Context, poaKeeper Keeper)
		errContains string
	}{
		{
			name:        "no sender",
			prepareFn:   nil,
			sender:      sdk.AccAddress{},
			operators:   nil,
			privilege:   "",
			postCheckFn: nil,
			errContains: "sender address is empty",
		},
		{
			name:        "sender is not owner",
			prepareFn:   nil,
			sender:      thirdParty,
			operators:   nil,
			privilege:   "",
			postCheckFn: nil,
			errContains: "not the owner",
		},
		{
			name:        "nil operators",
			prepareFn:   nil,
			sender:      owner,
			operators:   nil,
			privilege:   "",
			postCheckFn: nil,
			errContains: "no operators provided",
		},
		{
			name:        "empty operators",
			prepareFn:   nil,
			sender:      owner,
			operators:   []sdk.ValAddress{},
			privilege:   "",
			postCheckFn: nil,
			errContains: "no operators provided",
		},
		{
			name:        "empty privilege",
			prepareFn:   nil,
			sender:      owner,
			operators:   []sdk.ValAddress{validator1.GetOperator()},
			privilege:   "",
			postCheckFn: nil,
			errContains: "no privilege provided",
		},
		{
			name:      "non-validator operator",
			prepareFn: nil,
			sender:    owner,
			operators: []sdk.ValAddress{
				validator1.GetOperator(),
				nonValidator.GetOperator(),
			},
			privilege:   privilege,
			postCheckFn: nil,
			errContains: types.ErrNotValidator.Error(),
		},
		{
			name:      "duplicate operator",
			prepareFn: nil,
			sender:    owner,
			operators: []sdk.ValAddress{
				validator1.GetOperator(),
				validator1.GetOperator(),
			},
			privilege:   privilege,
			postCheckFn: nil,
			errContains: types.ErrMissingPrivilege.Error(),
		},
		{
			name: "same privilege twice",
			prepareFn: func(ctx sdk.Context, poaKeeper Keeper) {
				err := poaKeeper.RemovePrivilege(
					ctx,
					owner,
					[]sdk.ValAddress{validator2.GetOperator()},
					privilege,
				)
				require.NoError(t, err, "expected no error")
			},
			sender: owner,
			operators: []sdk.ValAddress{
				validator1.GetOperator(),
				validator2.GetOperator(),
			},
			privilege:   privilege,
			postCheckFn: nil,
			errContains: types.ErrMissingPrivilege.Error(),
		},
		{
			name:      "happy path - no privileges left",
			prepareFn: nil,
			sender:    owner,
			operators: []sdk.ValAddress{
				validator1.GetOperator(),
				validator2.GetOperator(),
				validator3.GetOperator(),
			},
			privilege: privilege,
			postCheckFn: func(ctx sdk.Context, poaKeeper Keeper) {
				require.Equal(
					t,
					0,
					len(poaKeeper.GetValidatorsOperatorsByPrivilege(ctx, privilege)),
					"expected operators mismatch",
				)

				require.Equal(
					t,
					0,
					len(poaKeeper.GetValidatorsConsAddrsByPrivilege(ctx, privilege)),
					"expected consensus addresses mismatch",
				)
			},
			errContains: "",
		},
		{
			name:      "happy path - some privileges left",
			prepareFn: nil,
			sender:    owner,
			operators: []sdk.ValAddress{
				validator2.GetOperator(),
			},
			privilege: privilege,
			postCheckFn: func(ctx sdk.Context, poaKeeper Keeper) {
				expectedValidators := []types.Validator{validator1, validator3}

				// The privilege's set consists of consensus addresses
				// sorted in ascending order lexicographically. Sort the
				// validators slice to match the expected order during comparison.
				slices.SortFunc(expectedValidators, func(a, b types.Validator) int {
					return bytes.Compare(a.GetConsAddress(), b.GetConsAddress())
				})

				expectedOperators := make([]sdk.ValAddress, len(expectedValidators))
				expectedConsAddrs := make([]sdk.ConsAddress, len(expectedValidators))
				for i, validator := range expectedValidators {
					expectedOperators[i] = validator.GetOperator()
					expectedConsAddrs[i] = validator.GetConsAddress()
				}

				require.Equal(
					t,
					expectedOperators,
					poaKeeper.GetValidatorsOperatorsByPrivilege(ctx, privilege),
					"expected operators mismatch",
				)

				require.Equal(
					t,
					expectedConsAddrs,
					poaKeeper.GetValidatorsConsAddrsByPrivilege(ctx, privilege),
					"expected consensus addresses mismatch",
				)
			},
			errContains: "",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx, poaKeeper := mockContext()

			poaKeeper.setOwner(ctx, owner)

			poaKeeper.setValidator(ctx, validator1)
			poaKeeper.setValidator(ctx, validator2)
			poaKeeper.setValidator(ctx, validator3)

			poaKeeper.setValidatorByConsAddr(ctx, validator1)
			poaKeeper.setValidatorByConsAddr(ctx, validator2)
			poaKeeper.setValidatorByConsAddr(ctx, validator3)

			err := poaKeeper.AddPrivilege(
				ctx,
				owner,
				[]sdk.ValAddress{
					validator1.GetOperator(),
					validator2.GetOperator(),
					validator3.GetOperator(),
				},
				privilege,
			)
			require.NoError(t, err, "expected no error")

			if test.prepareFn != nil {
				test.prepareFn(ctx, poaKeeper)
			}

			err = poaKeeper.RemovePrivilege(
				ctx,
				test.sender,
				test.operators,
				test.privilege,
			)

			if len(test.errContains) == 0 {
				require.NoError(t, err, "expected no error")
			} else {
				// ErrorContains checks if the error is non-nil so no need
				// for an explicit check here.
				require.ErrorContains(
					t,
					err,
					test.errContains,
					"expected different error message",
				)
			}

			if test.postCheckFn != nil {
				test.postCheckFn(ctx, poaKeeper)
			}
		})
	}
}

func TestGetValidatorsOperatorsByPrivilege(t *testing.T) {
	ctx, poaKeeper := mockContext()

	helper, _ := mockValidator()
	owner := sdk.AccAddress(helper.GetOperator())

	validator1, _ := mockValidator()
	validator2, _ := mockValidator()
	validator3, _ := mockValidator()

	poaKeeper.setOwner(ctx, owner)

	poaKeeper.setValidator(ctx, validator1)
	poaKeeper.setValidator(ctx, validator2)
	poaKeeper.setValidator(ctx, validator3)

	poaKeeper.setValidatorByConsAddr(ctx, validator1)
	poaKeeper.setValidatorByConsAddr(ctx, validator2)
	poaKeeper.setValidatorByConsAddr(ctx, validator3)

	err := poaKeeper.AddPrivilege(
		ctx,
		owner,
		[]sdk.ValAddress{
			validator1.GetOperator(),
			validator2.GetOperator(),
			validator3.GetOperator(),
		},
		privilege,
	)
	require.NoError(t, err, "expected no error")

	expectedValidators := []types.Validator{validator1, validator2, validator3}

	// The privilege's set consists of consensus addresses
	// sorted in ascending order lexicographically. Sort the
	// validators slice to match the expected order during comparison.
	slices.SortFunc(expectedValidators, func(a, b types.Validator) int {
		return bytes.Compare(a.GetConsAddress(), b.GetConsAddress())
	})

	expectedOperators := make([]sdk.ValAddress, len(expectedValidators))
	for i, validator := range expectedValidators {
		expectedOperators[i] = validator.GetOperator()
	}

	require.Equal(
		t,
		expectedOperators,
		poaKeeper.GetValidatorsOperatorsByPrivilege(ctx, privilege),
		"expected operators mismatch",
	)
}

func TestGetValidatorsConsAddrsByPrivilege(t *testing.T) {
	ctx, poaKeeper := mockContext()

	helper, _ := mockValidator()
	owner := sdk.AccAddress(helper.GetOperator())

	validator1, _ := mockValidator()
	validator2, _ := mockValidator()
	validator3, _ := mockValidator()

	poaKeeper.setOwner(ctx, owner)

	poaKeeper.setValidator(ctx, validator1)
	poaKeeper.setValidator(ctx, validator2)
	poaKeeper.setValidator(ctx, validator3)

	poaKeeper.setValidatorByConsAddr(ctx, validator1)
	poaKeeper.setValidatorByConsAddr(ctx, validator2)
	poaKeeper.setValidatorByConsAddr(ctx, validator3)

	err := poaKeeper.AddPrivilege(
		ctx,
		owner,
		[]sdk.ValAddress{
			validator1.GetOperator(),
			validator2.GetOperator(),
			validator3.GetOperator(),
		},
		privilege,
	)
	require.NoError(t, err, "expected no error")

	expectedValidators := []types.Validator{validator1, validator2, validator3}

	// The privilege's set consists of consensus addresses
	// sorted in ascending order lexicographically. Sort the
	// validators slice to match the expected order during comparison.
	slices.SortFunc(expectedValidators, func(a, b types.Validator) int {
		return bytes.Compare(a.GetConsAddress(), b.GetConsAddress())
	})

	expectedConsAddrs := make([]sdk.ConsAddress, len(expectedValidators))
	for i, validator := range expectedValidators {
		expectedConsAddrs[i] = validator.GetConsAddress()
	}

	require.Equal(
		t,
		expectedConsAddrs,
		poaKeeper.GetValidatorsConsAddrsByPrivilege(ctx, privilege),
		"expected consensus addresses mismatch",
	)
}
