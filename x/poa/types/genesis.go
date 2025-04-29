package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewGenesisState creates a new GenesisState object.
func NewGenesisState(
	params Params,
	owner sdk.AccAddress,
	validators []Validator,
) GenesisState {
	return GenesisState{
		Params:     params,
		Owner:      owner.String(),
		Validators: validators,
	}
}

// DefaultGenesisState defines the default GenesisState.
//
// WARNING: The default genesis state has an empty owner address hence
// it is invalid (ValidateGenesis will fail). A proper owner must be set at
// later stages, before running the network. This is done on purpose to avoid
// using a random owner that cannot be controlled.
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params: DefaultParams(),
		Owner:  "",
	}
}

// ValidateGenesis validates the poa genesis parameters.
func ValidateGenesis(data GenesisState) error {
	if _, err := sdk.AccAddressFromBech32(data.Owner); err != nil {
		return fmt.Errorf("invalid owner address %s: %w", data.Owner, err)
	}

	if err := validateGenesisStateValidators(data.Validators); err != nil {
		return fmt.Errorf("failed to validate genesis validators: %w", err)
	}

	if err := validateGenesisStatePrivilegeAssignments(data.PrivilegeAssignments); err != nil {
		return fmt.Errorf("failed to validate genesis privilege assignments: %w", err)
	}

	return data.Params.Validate()
}

// validateGenesisStateValidators validates the validator set in genesis.
func validateGenesisStateValidators(validators []Validator) (err error) {
	operators := make(map[string]bool, len(validators))
	consPubKeys := make(map[string]bool, len(validators))

	for _, validator := range validators {
		operator := validator.GetOperatorBech32()
		consPubKey := validator.GetConsPubKeyBech32()

		if _, ok := operators[operator]; ok {
			return fmt.Errorf(
				"duplicate validator in genesis state: moniker %v, operator %v",
				validator.Description.Moniker,
				operator,
			)
		}
		if _, ok := consPubKeys[consPubKey]; ok {
			return fmt.Errorf(
				"duplicate validator in genesis state: moniker %v, consensus pubkey %v",
				validator.Description.Moniker,
				consPubKey,
			)
		}

		operators[operator] = true
		consPubKeys[consPubKey] = true
	}
	return
}

// validateGenesisStatePrivilegeAssignments validates privilege assignments in genesis.
//
// Ensures that:
// - The operator and privilege are non-empty strings.
// - There are no duplicated assignments.
func validateGenesisStatePrivilegeAssignments(assignments []ValidatorPrivilegeAssignment) error {
	if len(assignments) == 0 {
		return nil
	}

	seen := make(map[string]bool)

	for _, assignment := range assignments {
		operator := assignment.GetOperatorBech32()
		privilege := assignment.GetPrivilege()

		if len(operator) == 0 || len(privilege) == 0 {
			return fmt.Errorf(
				"invalid privilege assignment in genesis state: operator %v, privilege %v",
				operator,
				privilege,
			)
		}

		key := fmt.Sprintf("%s:%s", operator, privilege)

		if seen[key] {
			return fmt.Errorf(
				"duplicate privilege assignment in genesis state: validator %v, privilege %v",
				operator,
				privilege,
			)
		}

		seen[key] = true
	}

	return nil
}
