package poa

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v12/x/poa/keeper"
	"github.com/evmos/evmos/v12/x/poa/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// BeginBlocker check for infraction evidence or downtime of validators
// on every begin block
func BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock, k keeper.Keeper) {
	return
}

// EndBlocker called every block, process inflation, update validator set.
func EndBlocker(ctx sdk.Context, k keeper.Keeper) (updates []abci.ValidatorUpdate) {
	// Retrieve all validators
	validators := k.GetAllValidators(ctx)

	// Check the state of all validator
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
			k.SetValidatorState(ctx, validator, types.ValidatorStateJoined)

		case types.ValidatorStateLeaving:
			// Set the validator power to 0 and remove it from the keeper
			updates = append(updates, validator.ABCIValidatorUpdateRemove())
			k.RemoveValidator(ctx, validator.GetOperator())

		default:
			panic("A validator has a unknown state")
		}
	}

	return updates
}
