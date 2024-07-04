package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v12/x/poa/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// EndBlocker called every block, process inflation, update validator set.
func (k Keeper) EndBlocker(ctx sdk.Context) (updates []abci.ValidatorUpdate) {
	// Retrieve all validators
	validators := k.GetAllValidators(ctx)

	// Check the state of all validators
	for _, validator := range validators {
		validatorState, found := k.GetValidatorState(ctx, validator.GetOperator())

		// Panic on no state
		if !found {
			panic("Found a validator with no state, a validator should always have a state")
		}

		// Check the state
		switch validatorState {
		case types.ValidatorStateJoined:
			// No update if the validator has already joined the validator state

		case types.ValidatorStateJoining:
			// Return the new validator in the updates and set its state to joined
			updates = append(updates, validator.ABCIValidatorUpdateAppend())
			k.setValidatorState(ctx, validator, types.ValidatorStateJoined)

		case types.ValidatorStateLeaving:
			// Set the validator power to 0 and remove it from the keeper
			updates = append(updates, validator.ABCIValidatorUpdateRemove())
			k.removeValidator(ctx, validator.GetOperator())

		default:
			panic("A validator has an unknown state")
		}
	}

	return updates
}
