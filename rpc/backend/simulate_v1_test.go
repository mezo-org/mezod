package backend

import (
	"encoding/json"
	"errors"

	"github.com/stretchr/testify/mock"

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
