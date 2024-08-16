package keeper

import (
	"context"

	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/mezo-org/mezod/x/poa/types"
)

// BeginBlocker will persist the current header and validator set as a
// historical entry and prune the oldest entry based on the HistoricalEntries
// parameter.
func (k Keeper) BeginBlocker(ctx context.Context) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	k.TrackHistoricalInfo(sdkCtx)
	return nil
}

// EndBlocker called every block, update validator set.
func (k Keeper) EndBlocker(ctx context.Context) (
	updates []abci.ValidatorUpdate,
	err error,
) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Retrieve all validators
	validators := k.GetAllValidators(sdkCtx)

	// Check the state of all validators
	for _, validator := range validators {
		validatorState, found := k.GetValidatorState(sdkCtx, validator.GetOperator())

		// Panic on no state
		if !found {
			panic("Found a validator with no state, a validator should always have a state")
		}

		// Check the state
		switch validatorState {
		case types.ValidatorStateActive:
			// No update if the validator has already joined the validator state

		case types.ValidatorStateJoining:
			// Return the new validator in the updates and set its state to joined
			updates = append(updates, validator.ABCIValidatorUpdateAppend())
			k.setValidatorState(sdkCtx, validator, types.ValidatorStateActive)

		case types.ValidatorStateLeaving:
			// Set the validator power to 0 and remove it from the keeper
			updates = append(updates, validator.ABCIValidatorUpdateRemove())
			k.removeValidator(sdkCtx, validator.GetOperator())

		default:
			panic("A validator has an unknown state")
		}
	}

	return updates, nil
}
