package backend

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/mezo-org/mezod/rpc/backend/mocks"
	rpctypes "github.com/mezo-org/mezod/rpc/types"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
)

// registerSimulateV1SimError wires the gRPC mock so the keeper returns
// a successful response carrying a structured *SimError on the
// response's `error` field. The backend must return it verbatim so
// geth's JSON-RPC server emits the spec-reserved code via the error
// interface methods.
func registerSimulateV1SimError(
	queryClient *mocks.EVMQueryClient,
	code int,
	message string,
) {
	queryClient.On(
		"SimulateV1", mock.Anything, mock.AnythingOfType("*types.SimulateV1Request"),
	).Return(&evmtypes.SimulateV1Response{
		Error: &evmtypes.SimError{Code: int32(code), Message: message}, //nolint:gosec // spec-reserved codes fit in int32
	}, nil)
}

// registerSimulateV1OK wires the gRPC mock so the keeper returns a
// successful response with a serialized []*SimBlockResult payload.
func registerSimulateV1OK(queryClient *mocks.EVMQueryClient, results []*evmtypes.SimBlockResult) {
	bz, _ := json.Marshal(results)
	queryClient.On(
		"SimulateV1", mock.Anything, mock.AnythingOfType("*types.SimulateV1Request"),
	).Return(&evmtypes.SimulateV1Response{Result: bz}, nil)
}

// TestSimulateV1_EmptyResultsReturnEmptySlice verifies that a keeper
// response with no Result bytes surfaces as an empty slice on the RPC
// side, not as "not yet implemented" (which is reserved for the
// Unimplemented gRPC status path).
func (suite *BackendTestSuite) TestSimulateV1_EmptyResultsReturnEmptySlice() {
	_, bz := suite.buildEthereumTx()
	suite.SetupTest()

	client := suite.backend.clientCtx.Client.(*mocks.Client)
	queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
	_, err := RegisterBlock(client, 1, bz)
	suite.Require().NoError(err)
	registerSimulateV1OK(queryClient, nil)

	bn := rpctypes.BlockNumber(1)
	bnh := rpctypes.BlockNumberOrHash{BlockNumber: &bn}
	got, err := suite.backend.SimulateV1(evmtypes.SimOpts{}, &bnh)
	suite.Require().NoError(err)
	suite.Require().Empty(got)
}

func (suite *BackendTestSuite) TestSimulateV1_BubblesSimError() {
	testCases := []struct {
		name string
		code int
	}{
		{"move-precompile self reference → -38022", evmtypes.SimErrCodeMovePrecompileSelfReference},
		{"move-precompile duplicate dest → -38023", evmtypes.SimErrCodeMovePrecompileDuplicateDest},
		{"state and stateDiff both set → -32602", evmtypes.SimErrCodeInvalidParams},
		{"sanitize-chain block number invalid → -38020", evmtypes.SimErrCodeBlockNumberInvalid},
		{"sanitize-chain client limit exceeded → -38026", evmtypes.SimErrCodeClientLimitExceeded},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			_, bz := suite.buildEthereumTx()

			client := suite.backend.clientCtx.Client.(*mocks.Client)
			queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
			_, err := RegisterBlock(client, 1, bz)
			suite.Require().NoError(err)
			registerSimulateV1SimError(queryClient, tc.code, "boom")

			bn := rpctypes.BlockNumber(1)
			bnh := rpctypes.BlockNumberOrHash{BlockNumber: &bn}
			_, err = suite.backend.SimulateV1(evmtypes.SimOpts{}, &bnh)
			suite.Require().Error(err)

			var simErr *evmtypes.SimError
			suite.Require().True(errors.As(err, &simErr))
			suite.Require().Equal(tc.code, simErr.ErrorCode())
			suite.Require().Equal("boom", simErr.Error())
		})
	}
}

// TestSimulateV1_PopulatesBaseBlockHash: the rpc backend forwards the
// canonical CometBFT base block hash — the same hash
// eth_getBlockByNumber surfaces — on the gRPC request so the keeper
// can use it as the first simulated block's parentHash.
func (suite *BackendTestSuite) TestSimulateV1_PopulatesBaseBlockHash() {
	suite.SetupTest()
	_, bz := suite.buildEthereumTx()

	client := suite.backend.clientCtx.Client.(*mocks.Client)
	queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
	resBlock, err := RegisterBlock(client, 1, bz)
	suite.Require().NoError(err)

	wantBytes := resBlock.Block.Hash().Bytes()

	var captured *evmtypes.SimulateV1Request
	queryClient.On(
		"SimulateV1", mock.Anything, mock.AnythingOfType("*types.SimulateV1Request"),
	).Run(func(args mock.Arguments) {
		captured = args.Get(1).(*evmtypes.SimulateV1Request)
	}).Return(&evmtypes.SimulateV1Response{Result: nil}, nil)

	bn := rpctypes.BlockNumber(1)
	bnh := rpctypes.BlockNumberOrHash{BlockNumber: &bn}
	_, err = suite.backend.SimulateV1(evmtypes.SimOpts{}, &bnh)
	suite.Require().NoError(err)
	suite.Require().NotNil(captured)
	suite.Require().Equal(wantBytes, captured.BaseBlockHash)
}

// gRPC DeadlineExceeded must surface as a structured -32016 SimError.
func (suite *BackendTestSuite) TestSimulateV1_GRPCDeadlineExceededTranslatesToSimError() {
	suite.SetupTest()
	_, bz := suite.buildEthereumTx()

	client := suite.backend.clientCtx.Client.(*mocks.Client)
	queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
	_, err := RegisterBlock(client, 1, bz)
	suite.Require().NoError(err)

	queryClient.On(
		"SimulateV1", mock.Anything, mock.AnythingOfType("*types.SimulateV1Request"),
	).Return((*evmtypes.SimulateV1Response)(nil), status.Error(codes.DeadlineExceeded, "deadline exceeded"))

	bn := rpctypes.BlockNumber(1)
	bnh := rpctypes.BlockNumberOrHash{BlockNumber: &bn}
	_, err = suite.backend.SimulateV1(evmtypes.SimOpts{}, &bnh)
	suite.Require().Error(err)

	var simErr *evmtypes.SimError
	suite.Require().True(errors.As(err, &simErr))
	suite.Require().Equal(evmtypes.SimErrCodeTimeout, simErr.ErrorCode())
}

// Local context.DeadlineExceeded must surface as a structured -32016 SimError.
func (suite *BackendTestSuite) TestSimulateV1_LocalDeadlineExceededTranslatesToSimError() {
	suite.SetupTest()
	_, bz := suite.buildEthereumTx()

	client := suite.backend.clientCtx.Client.(*mocks.Client)
	queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
	_, err := RegisterBlock(client, 1, bz)
	suite.Require().NoError(err)

	queryClient.On(
		"SimulateV1", mock.Anything, mock.AnythingOfType("*types.SimulateV1Request"),
	).Return((*evmtypes.SimulateV1Response)(nil), context.DeadlineExceeded)

	bn := rpctypes.BlockNumber(1)
	bnh := rpctypes.BlockNumberOrHash{BlockNumber: &bn}
	_, err = suite.backend.SimulateV1(evmtypes.SimOpts{}, &bnh)
	suite.Require().Error(err)

	var simErr *evmtypes.SimError
	suite.Require().True(errors.As(err, &simErr))
	suite.Require().Equal(evmtypes.SimErrCodeTimeout, simErr.ErrorCode())
}

func (suite *BackendTestSuite) TestSimulateV1_UnmarshalsResults() {
	suite.SetupTest()
	_, bz := suite.buildEthereumTx()

	client := suite.backend.clientCtx.Client.(*mocks.Client)
	queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
	_, err := RegisterBlock(client, 1, bz)
	suite.Require().NoError(err)

	expected := []*evmtypes.SimBlockResult{
		{Block: map[string]interface{}{"number": "0x1"}, Calls: []evmtypes.SimCallResult{}},
	}
	registerSimulateV1OK(queryClient, expected)

	bn := rpctypes.BlockNumber(1)
	bnh := rpctypes.BlockNumberOrHash{BlockNumber: &bn}
	got, err := suite.backend.SimulateV1(evmtypes.SimOpts{}, &bnh)
	suite.Require().NoError(err)
	suite.Require().Len(got, 1)
	suite.Require().Equal("0x1", got[0].Block["number"])
}
