package mocks

import (
	context "context"

	oracletypes "github.com/skip-mev/connect/v2/x/oracle/types"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

type OracleQueryClient struct {
	mock.Mock
}

func (oqc *OracleQueryClient) GetAllCurrencyPairs(
	ctx context.Context,
	req *oracletypes.GetAllCurrencyPairsRequest,
	opts ...grpc.CallOption,
) (*oracletypes.GetAllCurrencyPairsResponse, error) {
	args := oqc.Called(ctx, req, opts)

	if res := args.Get(0); res != nil {
		return res.(*oracletypes.GetAllCurrencyPairsResponse), args.Error(1)
	}

	return nil, args.Error(1)
}

func (oqc *OracleQueryClient) GetCurrencyPairMapping(
	ctx context.Context,
	req *oracletypes.GetCurrencyPairMappingRequest,
	opts ...grpc.CallOption,
) (*oracletypes.GetCurrencyPairMappingResponse, error) {
	args := oqc.Called(ctx, req, opts)

	if res := args.Get(0); res != nil {
		return res.(*oracletypes.GetCurrencyPairMappingResponse), args.Error(1)
	}

	return nil, args.Error(1)
}

func (oqc *OracleQueryClient) GetPrice(
	ctx context.Context,
	req *oracletypes.GetPriceRequest,
	opts ...grpc.CallOption,
) (*oracletypes.GetPriceResponse, error) {
	args := oqc.Called(ctx, req, opts)

	if res := args.Get(0); res != nil {
		return res.(*oracletypes.GetPriceResponse), args.Error(1)
	}

	return nil, args.Error(1)
}

func (oqc *OracleQueryClient) GetPrices(
	ctx context.Context,
	in *oracletypes.GetPricesRequest,
	opts ...grpc.CallOption,
) (*oracletypes.GetPricesResponse, error) {
	args := oqc.Called(ctx, in, opts)

	if res := args.Get(0); res != nil {
		return res.(*oracletypes.GetPricesResponse), args.Error(1)
	}

	return nil, args.Error(1)
}

func NewOracleQueryClient(t mockConstructorTestingTNewQueryClient) *OracleQueryClient {
	mock := &OracleQueryClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
