package keeper

import (
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/evmos/evmos/v12/x/poa/types"
)

// NewQuerier creates a new querier for poa clients.
func NewQuerier(k Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		switch path[0] {
		case types.QueryValidators:
			return queryValidators(ctx, k)

		case types.QueryValidator:
			return queryValidator(ctx, req, k)

		case types.QueryParams:
			return queryParams(ctx, k)

		case types.QueryApplications:
			return queryApplications(ctx, k)

		case types.QueryKickProposals:
			return queryKickProposals(ctx, k)

		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, "unknown poa query endpoint")
		}
	}
}

func queryValidators(ctx sdk.Context, k Keeper) ([]byte, error) {
	// Get all the validators
	validators := k.GetAllValidators(ctx)

	res, err := codec.MarshalJSONIndent(types.ModuleCdc, validators)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}

func queryValidator(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, error) {
	var params types.QueryValidatorParams

	// Unmarschal request
	err := types.ModuleCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	validator, found := k.GetValidator(ctx, params.ValidatorAddr)
	if !found {
		return nil, types.ErrNoValidatorFound
	}

	res, err := codec.MarshalJSONIndent(types.ModuleCdc, validator)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}

func queryParams(ctx sdk.Context, k Keeper) ([]byte, error) {
	params := k.GetParams(ctx)

	res, err := codec.MarshalJSONIndent(types.ModuleCdc, params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}

func queryApplications(ctx sdk.Context, k Keeper) ([]byte, error) {
	// Get all the applications
	applications := k.GetAllApplications(ctx)

	res, err := codec.MarshalJSONIndent(types.ModuleCdc, applications)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}

func queryKickProposals(ctx sdk.Context, k Keeper) ([]byte, error) {
	// Get all the kick proposals
	kickProposals := k.GetAllKickProposals(ctx)

	res, err := codec.MarshalJSONIndent(types.ModuleCdc, kickProposals)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}
