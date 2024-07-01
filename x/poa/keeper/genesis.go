package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v12/x/poa/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// InitGenesis initialize default parameters
// and the keeper's address to pubkey map
func (k Keeper) InitGenesis(ctx sdk.Context, data types.GenesisState) (res []abci.ValidatorUpdate) {
	err := k.setParams(ctx, data.Params)
	if err != nil {
		panic(errorsmod.Wrapf(err, "error setting params"))
	}

	// Set validators in the storage
	for _, validator := range data.Validators {
		k.setValidator(ctx, validator)
		k.setValidatorByConsAddr(ctx, validator)
		k.setValidatorState(ctx, validator, types.ValidatorStateJoined)
		res = append(res, validator.ABCIValidatorUpdateAppend())
	}

	return res
}

// ExportGenesis writes the current store values
// to a genesis file, which can be imported again
// with InitGenesis
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	return &types.GenesisState{
		Params:     k.GetParams(ctx),
		Validators: k.GetAllValidators(ctx),
	}
}
