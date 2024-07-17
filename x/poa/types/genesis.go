package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewGenesisState creates a new GenesisState object.
func NewGenesisState(params Params, validators []Validator) GenesisState {
	return GenesisState{
		Params:     params,
		Validators: validators,
	}
}

// DefaultGenesisState defines the default GenesisState.
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params: DefaultParams(),
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
