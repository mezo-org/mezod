package keeper

import (
	errorsmod "cosmossdk.io/errors"
	abci "github.com/cometbft/cometbft/abci/types"
	tmtypes "github.com/cometbft/cometbft/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/mezo-org/mezod/x/poa/types"
)

// InitGenesis initialize default parameters
// and the keeper's address to pubkey map
func (k Keeper) InitGenesis(ctx sdk.Context, data types.GenesisState) (res []abci.ValidatorUpdate) {
	err := k.setParams(ctx, data.Params)
	if err != nil {
		panic(errorsmod.Wrapf(err, "error setting params"))
	}

	k.setOwner(ctx, sdk.MustAccAddressFromBech32(data.Owner))

	for _, validator := range data.Validators {
		k.setValidator(ctx, validator)
		k.setValidatorByConsAddr(ctx, validator)
		k.setValidatorState(ctx, validator, types.ValidatorStateActive)
		res = append(res, validator.ABCIValidatorUpdateAppend())
	}

	for _, assignment := range data.PrivilegeAssignments {
		privilege := assignment.GetPrivilege()
		operatorBech32 := assignment.GetOperatorBech32()

		operator, err := sdk.ValAddressFromBech32(operatorBech32)
		if err != nil {
			panic(errorsmod.Wrapf(
				err,
				"error converting operator %s to sdk.ValAddress",
				operatorBech32,
			))
		}

		err = k.AddPrivilege(ctx, k.GetOwner(ctx), []sdk.ValAddress{operator}, privilege)
		if err != nil {
			panic(errorsmod.Wrapf(
				err,
				"error adding privilege %s for operator %s",
				privilege,
				operatorBech32,
			))
		}
	}

	return res
}

// ExportGenesis writes the current store values
// to a genesis file, which can be imported again
// with InitGenesis
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	privilegeAssignments := make([]types.ValidatorPrivilegeAssignment, 0)
	for _, privilege := range k.getAllPrivileges(ctx) {
		operators := k.GetValidatorsOperatorsByPrivilege(ctx, privilege)
		for _, operator := range operators {
			privilegeAssignments = append(
				privilegeAssignments,
				types.ValidatorPrivilegeAssignment{
					OperatorBech32: operator.String(),
					Privilege:      privilege,
				},
			)
		}
	}

	return &types.GenesisState{
		Params:               k.GetParams(ctx),
		Owner:                k.GetOwner(ctx).String(),
		Validators:           k.GetAllValidators(ctx),
		PrivilegeAssignments: privilegeAssignments,
	}
}

// ExportGenesisValidators exports the existing validators for genesis purposes.
func (k Keeper) ExportGenesisValidators(
	ctx sdk.Context,
) ([]tmtypes.GenesisValidator, error) {
	tmValidators := make([]tmtypes.GenesisValidator, 0)

	for _, validator := range k.GetAllValidators(ctx) {
		state, found := k.GetValidatorState(ctx, validator.GetOperator())
		// Panic on no state
		if !found {
			panic("found a validator with no state; a validator should always have a state")
		}

		// Ignore candidate validators and validators that are leaving.
		// The exported state should contain validators that can continue
		// their work after the state is reloaded.
		if state != types.ValidatorStateActive {
			continue
		}

		pubKey := validator.GetConsPubKey()

		tmPubKey, err := cryptocodec.ToCmtPubKeyInterface(pubKey)
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
