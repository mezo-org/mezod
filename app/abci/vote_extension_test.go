package abci

import (
	"fmt"
	"testing"

	"cosmossdk.io/log"

	cmtabci "github.com/cometbft/cometbft/abci/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	servermock "github.com/cosmos/cosmos-sdk/server/mock"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/mezo-org/mezod/app/abci/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

func TestVoteExtensionHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(VoteExtensionHandlerTestSuite))
}

type VoteExtensionHandlerTestSuite struct {
	suite.Suite

	logger        log.Logger
	ctx           sdk.Context
	requestHeight int64
	handler       *VoteExtensionHandler
}

func (s *VoteExtensionHandlerTestSuite) SetupTest() {
	s.logger = log.NewNopLogger()

	s.ctx = sdk.NewContext(
		servermock.NewCommitMultiStore(),
		tmproto.Header{},
		false,
		s.logger,
	)

	s.requestHeight = 100
}

func (s *VoteExtensionHandlerTestSuite) TestExtendVote() {
	tests := []struct {
		name          string
		subHandlersFn func() map[VoteExtensionPart]IVoteExtensionHandler
		expectedVE    *types.VoteExtension
		errContains   string
	}{
		{
			name:          "no sub-handlers",
			subHandlersFn: func() map[VoteExtensionPart]IVoteExtensionHandler { return nil },
			expectedVE:    nil,
			errContains:   "all sub-handlers failed to extend vote",
		},
		{
			name: "single sub-handler returning non-empty part",
			subHandlersFn: func() map[VoteExtensionPart]IVoteExtensionHandler {
				subHandler := newMockVoteExtensionHandler()

				subHandler.extendVoteHandler.On(
					"call",
					mock.Anything,
					mock.Anything,
				).Return(
					&cmtabci.ResponseExtendVote{VoteExtension: []byte("part1")},
					nil,
				)

				return map[VoteExtensionPart]IVoteExtensionHandler{
					VoteExtensionPart(1): subHandler,
				}
			},
			expectedVE: &types.VoteExtension{
				Height: s.requestHeight,
				Parts:  map[uint32][]byte{1: []byte("part1")},
			},
			errContains: "",
		},
		{
			name: "single sub-handler returning empty part",
			subHandlersFn: func() map[VoteExtensionPart]IVoteExtensionHandler {
				subHandler := newMockVoteExtensionHandler()

				subHandler.extendVoteHandler.On(
					"call",
					mock.Anything,
					mock.Anything,
				).Return(
					&cmtabci.ResponseExtendVote{VoteExtension: []byte{}},
					nil,
				)

				return map[VoteExtensionPart]IVoteExtensionHandler{
					VoteExtensionPart(1): subHandler,
				}
			},
			expectedVE: &types.VoteExtension{
				Height: s.requestHeight,
				Parts:  map[uint32][]byte{1: {}},
			},
			errContains: "",
		},
		{
			name: "single sub-handler returning nil part",
			subHandlersFn: func() map[VoteExtensionPart]IVoteExtensionHandler {
				subHandler := newMockVoteExtensionHandler()

				subHandler.extendVoteHandler.On(
					"call",
					mock.Anything,
					mock.Anything,
				).Return(
					&cmtabci.ResponseExtendVote{VoteExtension: nil},
					nil,
				)

				return map[VoteExtensionPart]IVoteExtensionHandler{
					VoteExtensionPart(1): subHandler,
				}
			},
			expectedVE: &types.VoteExtension{
				Height: s.requestHeight,
				Parts:  map[uint32][]byte{1: {}},
			},
			errContains: "",
		},
		{
			name: "single sub-handler returning error",
			subHandlersFn: func() map[VoteExtensionPart]IVoteExtensionHandler {
				subHandler := newMockVoteExtensionHandler()

				subHandler.extendVoteHandler.On(
					"call",
					mock.Anything,
					mock.Anything,
				).Return(
					nil,
					fmt.Errorf("sub-handler error"),
				)

				return map[VoteExtensionPart]IVoteExtensionHandler{
					VoteExtensionPart(1): subHandler,
				}
			},
			expectedVE:  nil,
			errContains: "all sub-handlers failed to extend vote",
		},
		{
			name: "multiple sub-handlers - all returning non-empty parts",
			subHandlersFn: func() map[VoteExtensionPart]IVoteExtensionHandler {
				subHandler1 := newMockVoteExtensionHandler()
				subHandler2 := newMockVoteExtensionHandler()

				subHandler1.extendVoteHandler.On(
					"call",
					mock.Anything,
					mock.Anything,
				).Return(
					&cmtabci.ResponseExtendVote{VoteExtension: []byte("part1")},
					nil,
				)

				subHandler2.extendVoteHandler.On(
					"call",
					mock.Anything,
					mock.Anything,
				).Return(
					&cmtabci.ResponseExtendVote{VoteExtension: []byte("part2")},
					nil,
				)

				return map[VoteExtensionPart]IVoteExtensionHandler{
					VoteExtensionPart(1): subHandler1,
					VoteExtensionPart(2): subHandler2,
				}
			},
			expectedVE: &types.VoteExtension{
				Height: s.requestHeight,
				Parts: map[uint32][]byte{
					1: []byte("part1"),
					2: []byte("part2"),
				},
			},
			errContains: "",
		},
		{
			name: "multiple sub-handlers - one returning empty part",
			subHandlersFn: func() map[VoteExtensionPart]IVoteExtensionHandler {
				subHandler1 := newMockVoteExtensionHandler()
				subHandler2 := newMockVoteExtensionHandler()

				subHandler1.extendVoteHandler.On(
					"call",
					mock.Anything,
					mock.Anything,
				).Return(
					&cmtabci.ResponseExtendVote{VoteExtension: []byte{}},
					nil,
				)

				subHandler2.extendVoteHandler.On(
					"call",
					mock.Anything,
					mock.Anything,
				).Return(
					&cmtabci.ResponseExtendVote{VoteExtension: []byte("part2")},
					nil,
				)

				return map[VoteExtensionPart]IVoteExtensionHandler{
					VoteExtensionPart(1): subHandler1,
					VoteExtensionPart(2): subHandler2,
				}
			},
			expectedVE: &types.VoteExtension{
				Height: s.requestHeight,
				Parts: map[uint32][]byte{
					1: {},
					2: []byte("part2"),
				},
			},
			errContains: "",
		},
		{
			name: "multiple sub-handlers - one returning nil part",
			subHandlersFn: func() map[VoteExtensionPart]IVoteExtensionHandler {
				subHandler1 := newMockVoteExtensionHandler()
				subHandler2 := newMockVoteExtensionHandler()

				subHandler1.extendVoteHandler.On(
					"call",
					mock.Anything,
					mock.Anything,
				).Return(
					&cmtabci.ResponseExtendVote{VoteExtension: nil},
					nil,
				)

				subHandler2.extendVoteHandler.On(
					"call",
					mock.Anything,
					mock.Anything,
				).Return(
					&cmtabci.ResponseExtendVote{VoteExtension: []byte("part2")},
					nil,
				)

				return map[VoteExtensionPart]IVoteExtensionHandler{
					VoteExtensionPart(1): subHandler1,
					VoteExtensionPart(2): subHandler2,
				}
			},
			expectedVE: &types.VoteExtension{
				Height: s.requestHeight,
				Parts: map[uint32][]byte{
					1: {},
					2: []byte("part2"),
				},
			},
			errContains: "",
		},
		{
			name: "multiple sub-handlers - one returning error",
			subHandlersFn: func() map[VoteExtensionPart]IVoteExtensionHandler {
				subHandler1 := newMockVoteExtensionHandler()
				subHandler2 := newMockVoteExtensionHandler()

				subHandler1.extendVoteHandler.On(
					"call",
					mock.Anything,
					mock.Anything,
				).Return(
					nil,
					fmt.Errorf("sub-handler 1 error"),
				)

				subHandler2.extendVoteHandler.On(
					"call",
					mock.Anything,
					mock.Anything,
				).Return(
					&cmtabci.ResponseExtendVote{VoteExtension: []byte("part2")},
					nil,
				)

				return map[VoteExtensionPart]IVoteExtensionHandler{
					VoteExtensionPart(1): subHandler1,
					VoteExtensionPart(2): subHandler2,
				}
			},
			expectedVE: &types.VoteExtension{
				Height: s.requestHeight,
				Parts:  map[uint32][]byte{2: []byte("part2")},
			},
			errContains: "",
		},
		{
			name: "multiple sub-handlers - all returning errors",
			subHandlersFn: func() map[VoteExtensionPart]IVoteExtensionHandler {
				subHandler1 := newMockVoteExtensionHandler()
				subHandler2 := newMockVoteExtensionHandler()

				subHandler1.extendVoteHandler.On(
					"call",
					mock.Anything,
					mock.Anything,
				).Return(
					nil,
					fmt.Errorf("sub-handler 1 error"),
				)

				subHandler2.extendVoteHandler.On(
					"call",
					mock.Anything,
					mock.Anything,
				).Return(
					nil,
					fmt.Errorf("sub-handler 2 error"),
				)

				return map[VoteExtensionPart]IVoteExtensionHandler{
					VoteExtensionPart(1): subHandler1,
					VoteExtensionPart(2): subHandler2,
				}
			},
			expectedVE:  nil,
			errContains: "all sub-handlers failed to extend vote",
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			s.SetupTest()

			s.handler = &VoteExtensionHandler{
				logger:      s.logger,
				subHandlers: test.subHandlersFn(),
			}

			req := &cmtabci.RequestExtendVote{
				Height: s.requestHeight,
			}

			res, err := s.handler.ExtendVoteHandler()(s.ctx, req)

			if len(test.errContains) == 0 {
				s.Require().NoError(err, "expected no error")
			} else {
				// ErrorContains checks if the error is non-nil so no need
				// for an explicit check here.
				s.Require().ErrorContains(
					err,
					test.errContains,
					"expected different error message",
				)
			}

			var actualVE *types.VoteExtension
			if res != nil {
				actualVE = new(types.VoteExtension)
				err = actualVE.Unmarshal(res.VoteExtension)
				s.Require().NoError(err)
			}

			s.Require().Equal(
				test.expectedVE,
				actualVE,
				"expected different vote extension",
			)
		})
	}
}

type mockExtendVoteHandler struct {
	mock.Mock
}

func (mevh *mockExtendVoteHandler) call(
	ctx sdk.Context,
	req *cmtabci.RequestExtendVote,
) (*cmtabci.ResponseExtendVote, error) {
	args := mevh.Called(ctx, req)

	if res := args.Get(0); res != nil {
		return res.(*cmtabci.ResponseExtendVote), args.Error(1)
	}

	return nil, args.Error(1)
}

type mockVerifyVoteExtensionHandler struct {
	mock.Mock
}

func (mvveh *mockVerifyVoteExtensionHandler) call(
	ctx sdk.Context,
	req *cmtabci.RequestVerifyVoteExtension,
) (*cmtabci.ResponseVerifyVoteExtension, error) {
	args := mvveh.Called(ctx, req)

	if res := args.Get(0); res != nil {
		return res.(*cmtabci.ResponseVerifyVoteExtension), args.Error(1)
	}

	return nil, args.Error(1)
}

type mockVoteExtensionHandler struct {
	extendVoteHandler          *mockExtendVoteHandler
	verifyVoteExtensionHandler *mockVerifyVoteExtensionHandler
}

func newMockVoteExtensionHandler() *mockVoteExtensionHandler {
	return &mockVoteExtensionHandler{
		extendVoteHandler:          &mockExtendVoteHandler{},
		verifyVoteExtensionHandler: &mockVerifyVoteExtensionHandler{},
	}
}

func (mveh *mockVoteExtensionHandler) ExtendVoteHandler() sdk.ExtendVoteHandler {
	return mveh.extendVoteHandler.call
}

func (mveh *mockVoteExtensionHandler) VerifyVoteExtensionHandler() sdk.VerifyVoteExtensionHandler {
	return mveh.verifyVoteExtensionHandler.call
}
