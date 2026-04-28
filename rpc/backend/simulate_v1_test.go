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

// TestSimulateV1_EnvelopePassthroughPreservesAllFields: the gRPC-side
// payload encodes the full envelope shape (header fields,
// transactions, calls). The backend layer treats the bytes opaquely
// and the unmarshal must round-trip every field onto SimBlockResult.
//
// The backend.SimulateV1 unmarshal path keeps the Block as
// map[string]interface{} (asymmetric per design — see
// SimBlockResult.UnmarshalJSON), so the assertions read fields out of
// the legacy map.
func (suite *BackendTestSuite) TestSimulateV1_EnvelopePassthroughPreservesAllFields() {
	suite.SetupTest()
	_, bz := suite.buildEthereumTx()

	client := suite.backend.clientCtx.Client.(*mocks.Client)
	queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
	_, err := RegisterBlock(client, 1, bz)
	suite.Require().NoError(err)

	keeperPayload := []*evmtypes.SimBlockResult{{
		Block: map[string]interface{}{
			"number":           "0xb",
			"hash":             "0x1234abcdef",
			"parentHash":       "0xfff0",
			"logsBloom":        "0x" + commonZeroBloom(),
			"stateRoot":        "0x0000000000000000000000000000000000000000000000000000000000000000",
			"miner":            "0x0000000000000000000000000000000000000000",
			"difficulty":       "0x0",
			"extraData":        "0x",
			"gasLimit":         "0x1c9c380",
			"gasUsed":          "0x5208",
			"timestamp":        "0x65d2c2c0",
			"transactionsRoot": "0xabcd",
			"receiptsRoot":     "0xdcba",
			"size":             "0x100",
			"transactions":     []string{"0xaaaa"},
			"uncles":           []string{},
		},
		Calls: []evmtypes.SimCallResult{},
	}}
	registerSimulateV1OK(queryClient, keeperPayload)

	bn := rpctypes.BlockNumber(1)
	bnh := rpctypes.BlockNumberOrHash{BlockNumber: &bn}
	got, err := suite.backend.SimulateV1(evmtypes.SimOpts{}, &bnh)
	suite.Require().NoError(err)
	suite.Require().Len(got, 1)

	// Every header key the keeper sent must round-trip onto the
	// untyped Block map (UnmarshalJSON's documented contract).
	for _, key := range []string{
		"number", "hash", "parentHash", "logsBloom", "stateRoot",
		"miner", "difficulty", "extraData", "gasLimit", "gasUsed",
		"timestamp", "transactionsRoot", "receiptsRoot", "size",
		"transactions", "uncles",
	} {
		suite.Require().Contains(got[0].Block, key,
			"%s must round-trip through gRPC payload onto SimBlockResult.Block", key)
	}
	suite.Require().Equal("0xb", got[0].Block["number"])
	suite.Require().Equal("0x1234abcdef", got[0].Block["hash"])
}

// TestSimulateV1_PassesReturnFullTransactionsThroughOpts: the rpc
// backend marshals SimOpts verbatim into the gRPC request — defensive
// pin that ReturnFullTransactions=true survives the JSON encode and
// the keeper sees the flag.
func (suite *BackendTestSuite) TestSimulateV1_PassesReturnFullTransactionsThroughOpts() {
	suite.SetupTest()
	_, bz := suite.buildEthereumTx()

	client := suite.backend.clientCtx.Client.(*mocks.Client)
	queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
	_, err := RegisterBlock(client, 1, bz)
	suite.Require().NoError(err)

	var captured *evmtypes.SimulateV1Request
	queryClient.On(
		"SimulateV1", mock.Anything, mock.AnythingOfType("*types.SimulateV1Request"),
	).Run(func(args mock.Arguments) {
		captured = args.Get(1).(*evmtypes.SimulateV1Request)
	}).Return(&evmtypes.SimulateV1Response{Result: nil}, nil)

	bn := rpctypes.BlockNumber(1)
	bnh := rpctypes.BlockNumberOrHash{BlockNumber: &bn}
	_, err = suite.backend.SimulateV1(evmtypes.SimOpts{ReturnFullTransactions: true}, &bnh)
	suite.Require().NoError(err)
	suite.Require().NotNil(captured)

	var roundTripped evmtypes.SimOpts
	suite.Require().NoError(json.Unmarshal(captured.Opts, &roundTripped))
	suite.Require().True(roundTripped.ReturnFullTransactions,
		"opts JSON forwarded to the keeper must preserve ReturnFullTransactions=true")
}

// commonZeroBloom returns 256 zero bytes hex-encoded (the canonical
// empty-block logsBloom). Local helper to keep the test fixture
// readable and self-contained.
func commonZeroBloom() string {
	const zeroByte = "00"
	out := make([]byte, 0, 512)
	for i := 0; i < 256; i++ {
		out = append(out, zeroByte...)
	}
	return string(out)
}
