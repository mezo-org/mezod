package types

import (
	"fmt"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v12/x/poa/keeper"
	tmtypes "github.com/tendermint/tendermint/types"
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
	if err := validateGenesisStateValidators(data.Validators); err != nil {
		return err
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
			return fmt.Errorf("duplicate validator in genesis state: moniker %v, address %v", val.Description.Moniker, val.GetConsAddr())
		}

		addrMap[strKey] = true
	}
	return
}

// ExportGenesisValidators exports the existing validators for genesis purposes.
func ExportGenesisValidators(
	ctx sdk.Context,
	keeper keeper.Keeper,
) ([]tmtypes.GenesisValidator, error) {
	tmValidators := make([]tmtypes.GenesisValidator, 0)

	for _, validator := range keeper.GetAllValidators(ctx) {
		state, found := keeper.GetValidatorState(ctx, validator.GetOperator())
		// Panic on no state
		if !found {
			panic("found a validator with no state; a validator should always have a state")
		}

		// Ignore candidate validators and validators that are leaving.
		// The exported state should contain validators that can continue
		// their work after the state is reloaded.
		if state != ValidatorStateJoined {
			continue
		}

		pubKey := validator.GetConsPubKey()

		tmPubKey, err := cryptocodec.ToTmPubKeyInterface(pubKey)
		if err != nil {
			return nil, err
		}

		tmValidators = append(
			tmValidators, tmtypes.GenesisValidator{
				Address: sdk.ConsAddress(tmPubKey.Address()).Bytes(),
				PubKey:  tmPubKey,
				Power:   1, // all existing validators have a power of 1
				Name:    validator.Description.Moniker,
			},
		)
	}

	return tmValidators, nil
}
