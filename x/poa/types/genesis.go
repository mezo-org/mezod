package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewGenesisState creates a new GenesisState object
func NewGenesisState(params Params, validators []Validator) GenesisState {
	return GenesisState{
		Params:     params,
		Validators: validators,
	}
}

// DefaultGenesisState - default GenesisState used by Cosmos Hub
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params: DefaultParams(),
	}
}

// ValidateGenesis validates the poa genesis parameters
func ValidateGenesis(data GenesisState) error {
	if _, err := sdk.AccAddressFromBech32(data.Owner); err != nil {
		return fmt.Errorf("invalid owner address %s: %w", data.Owner, err)
	}

	if err := validateGenesisStateValidators(data.Validators); err != nil {
		return fmt.Errorf("failed to validate genesis validators: %w", err)
	}

	return data.Params.Validate()
}

// Validate the validator set in genesis
func validateGenesisStateValidators(validators []Validator) (err error) {
	addrMap := make(map[string]bool, len(validators))

	for i := 0; i < len(validators); i++ {
		val := validators[i]
		strKey := string(val.GetConsPubKey().Bytes())

		if _, ok := addrMap[strKey]; ok {
			return fmt.Errorf(
				"duplicate validator in genesis state: moniker %v, address %v",
				val.Description.Moniker,
				val.GetConsAddr(),
			)
		}

		addrMap[strKey] = true
	}
	return
}
