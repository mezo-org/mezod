package backend

import (
	"errors"

	rpctypes "github.com/mezo-org/mezod/rpc/types"
)

func (suite *BackendTestSuite) TestSimulateV1_StubReturnsMethodNotFound() {
	_, err := suite.backend.SimulateV1(rpctypes.SimOpts{}, nil)
	suite.Require().Error(err)

	var coded *rpctypes.JSONRPCError
	suite.Require().True(errors.As(err, &coded), "error should be a *JSONRPCError, got %T", err)
	suite.Require().Equal(rpctypes.SimErrCodeMethodNotFound, coded.ErrorCode())
}

func (suite *BackendTestSuite) TestSimulateV1_StubIgnoresOptsAndBlockNr() {
	opts := rpctypes.SimOpts{
		BlockStateCalls:        []rpctypes.SimBlock{{}, {}},
		TraceTransfers:         true,
		Validation:             true,
		ReturnFullTransactions: true,
	}
	bn := rpctypes.BlockNumber(1)
	bnh := rpctypes.BlockNumberOrHash{BlockNumber: &bn}

	_, err := suite.backend.SimulateV1(opts, &bnh)
	suite.Require().Error(err)

	var coded *rpctypes.JSONRPCError
	suite.Require().True(errors.As(err, &coded))
	suite.Require().Equal(rpctypes.SimErrCodeMethodNotFound, coded.ErrorCode())
}
