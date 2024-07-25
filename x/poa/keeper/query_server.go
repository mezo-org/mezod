package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v12/x/poa/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ types.QueryServer = queryServer{}

type queryServer struct {
	keeper Keeper
}

// NewQueryServer returns an implementation of the QueryServer interface
// for the provided Keeper.
func NewQueryServer(keeper Keeper) types.QueryServer {
	return &queryServer{keeper: keeper}
}

func (qs queryServer) Params(
	ctx context.Context,
	_ *types.QueryParamsRequest,
) (*types.QueryParamsResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	params := qs.keeper.GetParams(sdkCtx)

	return &types.QueryParamsResponse{
		Params: params,
	}, nil
}

func (qs queryServer) Validators(
	ctx context.Context,
	_ *types.QueryValidatorsRequest,
) (*types.QueryValidatorsResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	validators := qs.keeper.GetAllValidators(sdkCtx)

	return &types.QueryValidatorsResponse{
		Validators: validators,
	}, nil
}

func (qs queryServer) Validator(
	ctx context.Context,
	request *types.QueryValidatorRequest,
) (*types.QueryValidatorResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if len(request.Operator) == 0 {
		return nil, status.Error(codes.InvalidArgument, "validator address cannot be empty")
	}

	// The request is expected to carry a bech32 encoded address of an operator.
	// We use that operator address to fetch the corresponding validator
	// instance from the store.
	operator, err := sdk.ValAddressFromBech32(request.Operator)
	if err != nil {
		return nil, err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	validator, found := qs.keeper.GetValidator(sdkCtx, operator)
	if !found {
		return nil, types.ErrNoValidatorFound
	}

	return &types.QueryValidatorResponse{
		Validator: validator,
	}, nil
}

func (qs queryServer) Applications(
	ctx context.Context,
	_ *types.QueryApplicationsRequest,
) (*types.QueryApplicationsResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	applications := qs.keeper.GetAllApplications(sdkCtx)

	return &types.QueryApplicationsResponse{
		Applications: applications,
	}, nil
}
