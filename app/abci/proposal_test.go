package abci

import (
	"fmt"
	"strings"
	"testing"

	"cosmossdk.io/log"

	cmtabci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	servermock "github.com/cosmos/cosmos-sdk/server/mock"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/mezo-org/mezod/app/abci/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

func TestProposalHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(ProposalHandlerTestSuite))
}

type ProposalHandlerTestSuite struct {
	suite.Suite

	logger     log.Logger
	ctx        sdk.Context
	maxTxBytes int64
	handler    *ProposalHandler
}

func (s *ProposalHandlerTestSuite) SetupTest() {
	s.logger = log.NewNopLogger()

	s.ctx = sdk.NewContext(
		servermock.NewCommitMultiStore(),
		cmtproto.Header{},
		false,
		s.logger,
	)

	// Enable vote extensions at height 100.
	s.ctx = s.ctx.WithConsensusParams(cmtproto.ConsensusParams{
		Abci: &cmtproto.ABCIParams{
			VoteExtensionsEnableHeight: 100,
		},
	})

	s.maxTxBytes = 50
}

func (s *ProposalHandlerTestSuite) TestPrepareProposal() {
	txsVector := func(txs ...string) [][]byte {
		res := make([][]byte, 0)

		for _, tx := range txs {
			res = append(res, []byte(tx))
		}

		return res
	}

	tests := []struct {
		name               string
		subHandlersFn      func(*cmtabci.RequestPrepareProposal) map[VoteExtensionPart]IProposalHandler
		reqHeight          int64
		reqTxs             [][]byte
		expectedInjectedTx *types.InjectedTx
		expectedChainTxs   [][]byte // regular chain txs being part of the proposal
		errContains        string
	}{
		{
			name: "vote extensions not enabled",
			subHandlersFn: func(
				_ *cmtabci.RequestPrepareProposal,
			) map[VoteExtensionPart]IProposalHandler {
				return nil
			},
			// Vote extensions become enabled at height 100. However, the proposal
			// handler looks at the vote extensions from the previous block, so
			// from the handler's perspective, they are not enabled yet. The proposal
			// handler becomes fully functional at height 101.
			reqHeight:          100,
			reqTxs:             txsVector("tx1", "tx2"),
			expectedInjectedTx: nil,
			expectedChainTxs:   txsVector("tx1", "tx2"),
			errContains:        "",
		},
		{
			name: "no sub-handlers",
			subHandlersFn: func(
				_ *cmtabci.RequestPrepareProposal,
			) map[VoteExtensionPart]IProposalHandler {
				return nil
			},
			reqHeight:          101,
			reqTxs:             txsVector("tx1", "tx2"),
			expectedInjectedTx: nil,
			expectedChainTxs:   nil,
			errContains:        "all sub-handlers failed to prepare proposal",
		},
		{
			name: "single sub-handler injecting non-empty pseudo-tx",
			subHandlersFn: func(
				req *cmtabci.RequestPrepareProposal,
			) map[VoteExtensionPart]IProposalHandler {
				subHandler := newMockProposalHandler()

				subHandler.prepareProposalHandler.On(
					"call",
					mock.Anything,
					req,
				).Return(
					&cmtabci.ResponsePrepareProposal{
						Txs: append(
							[][]byte{[]byte("pseudoTx1")},
							req.Txs...,
						),
					},
					nil,
				)

				return map[VoteExtensionPart]IProposalHandler{
					VoteExtensionPart(1): subHandler,
				}
			},
			reqHeight: 101,
			reqTxs:    txsVector("tx1", "tx2"),
			expectedInjectedTx: &types.InjectedTx{
				Parts: map[uint32][]byte{1: []byte("pseudoTx1")},
			},
			expectedChainTxs: txsVector("tx1", "tx2"),
			errContains:      "",
		},
		{
			name: "single sub-handler injecting empty pseudo-tx",
			subHandlersFn: func(
				req *cmtabci.RequestPrepareProposal,
			) map[VoteExtensionPart]IProposalHandler {
				subHandler := newMockProposalHandler()

				subHandler.prepareProposalHandler.On(
					"call",
					mock.Anything,
					req,
				).Return(
					&cmtabci.ResponsePrepareProposal{
						Txs: append([][]byte{{}}, req.Txs...),
					},
					nil,
				)

				return map[VoteExtensionPart]IProposalHandler{
					VoteExtensionPart(1): subHandler,
				}
			},
			reqHeight: 101,
			reqTxs:    txsVector("tx1", "tx2"),
			expectedInjectedTx: &types.InjectedTx{
				Parts: map[uint32][]byte{1: {}},
			},
			expectedChainTxs: txsVector("tx1", "tx2"),
			errContains:      "",
		},
		{
			name: "single sub-handler injecting nil pseudo-tx",
			subHandlersFn: func(
				req *cmtabci.RequestPrepareProposal,
			) map[VoteExtensionPart]IProposalHandler {
				subHandler := newMockProposalHandler()

				subHandler.prepareProposalHandler.On(
					"call",
					mock.Anything,
					req,
				).Return(
					&cmtabci.ResponsePrepareProposal{
						Txs: append([][]byte{nil}, req.Txs...),
					},
					nil,
				)

				return map[VoteExtensionPart]IProposalHandler{
					VoteExtensionPart(1): subHandler,
				}
			},
			reqHeight: 101,
			reqTxs:    txsVector("tx1", "tx2"),
			expectedInjectedTx: &types.InjectedTx{
				Parts: map[uint32][]byte{1: {}},
			},
			expectedChainTxs: txsVector("tx1", "tx2"),
			errContains:      "",
		},
		{
			name: "single sub-handler returning error",
			subHandlersFn: func(
				req *cmtabci.RequestPrepareProposal,
			) map[VoteExtensionPart]IProposalHandler {
				subHandler := newMockProposalHandler()

				subHandler.prepareProposalHandler.On(
					"call",
					mock.Anything,
					req,
				).Return(
					nil,
					fmt.Errorf("sub-handler error"),
				)

				return map[VoteExtensionPart]IProposalHandler{
					VoteExtensionPart(1): subHandler,
				}
			},
			reqHeight:          101,
			reqTxs:             txsVector("tx1", "tx2"),
			expectedInjectedTx: nil,
			expectedChainTxs:   nil,
			errContains:        "all sub-handlers failed to prepare proposal",
		},
		{
			name: "multiple sub-handlers - all injecting non-empty pseudo-txs",
			subHandlersFn: func(
				req *cmtabci.RequestPrepareProposal,
			) map[VoteExtensionPart]IProposalHandler {
				subHandler1 := newMockProposalHandler()
				subHandler2 := newMockProposalHandler()

				subHandler1.prepareProposalHandler.On(
					"call",
					mock.Anything,
					req,
				).Return(
					&cmtabci.ResponsePrepareProposal{
						Txs: append(
							[][]byte{[]byte("pseudoTx1")},
							req.Txs...,
						),
					},
					nil,
				)

				subHandler2.prepareProposalHandler.On(
					"call",
					mock.Anything,
					req,
				).Return(
					&cmtabci.ResponsePrepareProposal{
						Txs: append(
							[][]byte{[]byte("pseudoTx2")},
							req.Txs...,
						),
					},
					nil,
				)

				return map[VoteExtensionPart]IProposalHandler{
					VoteExtensionPart(1): subHandler1,
					VoteExtensionPart(2): subHandler2,
				}
			},
			reqHeight: 101,
			reqTxs:    txsVector("tx1", "tx2"),
			expectedInjectedTx: &types.InjectedTx{
				Parts: map[uint32][]byte{
					1: []byte("pseudoTx1"),
					2: []byte("pseudoTx2"),
				},
			},
			expectedChainTxs: txsVector("tx1", "tx2"),
			errContains:      "",
		},
		{
			name: "multiple sub-handlers - one injecting empty pseudo-tx",
			subHandlersFn: func(
				req *cmtabci.RequestPrepareProposal,
			) map[VoteExtensionPart]IProposalHandler {
				subHandler1 := newMockProposalHandler()
				subHandler2 := newMockProposalHandler()

				subHandler1.prepareProposalHandler.On(
					"call",
					mock.Anything,
					req,
				).Return(
					&cmtabci.ResponsePrepareProposal{
						Txs: append([][]byte{{}}, req.Txs...),
					},
					nil,
				)

				subHandler2.prepareProposalHandler.On(
					"call",
					mock.Anything,
					req,
				).Return(
					&cmtabci.ResponsePrepareProposal{
						Txs: append(
							[][]byte{[]byte("pseudoTx2")},
							req.Txs...,
						),
					},
					nil,
				)

				return map[VoteExtensionPart]IProposalHandler{
					VoteExtensionPart(1): subHandler1,
					VoteExtensionPart(2): subHandler2,
				}
			},
			reqHeight: 101,
			reqTxs:    txsVector("tx1", "tx2"),
			expectedInjectedTx: &types.InjectedTx{
				Parts: map[uint32][]byte{
					1: {},
					2: []byte("pseudoTx2"),
				},
			},
			expectedChainTxs: txsVector("tx1", "tx2"),
			errContains:      "",
		},
		{
			name: "multiple sub-handlers - one injecting nil pseudo-tx",
			subHandlersFn: func(
				req *cmtabci.RequestPrepareProposal,
			) map[VoteExtensionPart]IProposalHandler {
				subHandler1 := newMockProposalHandler()
				subHandler2 := newMockProposalHandler()

				subHandler1.prepareProposalHandler.On(
					"call",
					mock.Anything,
					req,
				).Return(
					&cmtabci.ResponsePrepareProposal{
						Txs: append([][]byte{nil}, req.Txs...),
					},
					nil,
				)

				subHandler2.prepareProposalHandler.On(
					"call",
					mock.Anything,
					req,
				).Return(
					&cmtabci.ResponsePrepareProposal{
						Txs: append(
							[][]byte{[]byte("pseudoTx2")},
							req.Txs...,
						),
					},
					nil,
				)

				return map[VoteExtensionPart]IProposalHandler{
					VoteExtensionPart(1): subHandler1,
					VoteExtensionPart(2): subHandler2,
				}
			},
			reqHeight: 101,
			reqTxs:    txsVector("tx1", "tx2"),
			expectedInjectedTx: &types.InjectedTx{
				Parts: map[uint32][]byte{
					1: {},
					2: []byte("pseudoTx2"),
				},
			},
			expectedChainTxs: txsVector("tx1", "tx2"),
			errContains:      "",
		},
		{
			name: "multiple sub-handlers - one returning error",
			subHandlersFn: func(
				req *cmtabci.RequestPrepareProposal,
			) map[VoteExtensionPart]IProposalHandler {
				subHandler1 := newMockProposalHandler()
				subHandler2 := newMockProposalHandler()

				subHandler1.prepareProposalHandler.On(
					"call",
					mock.Anything,
					req,
				).Return(
					nil,
					fmt.Errorf("sub-handler 1 error"),
				)

				subHandler2.prepareProposalHandler.On(
					"call",
					mock.Anything,
					req,
				).Return(
					&cmtabci.ResponsePrepareProposal{
						Txs: append(
							[][]byte{[]byte("pseudoTx2")},
							req.Txs...,
						),
					},
					nil,
				)

				return map[VoteExtensionPart]IProposalHandler{
					VoteExtensionPart(1): subHandler1,
					VoteExtensionPart(2): subHandler2,
				}
			},
			reqHeight: 101,
			reqTxs:    txsVector("tx1", "tx2"),
			expectedInjectedTx: &types.InjectedTx{
				Parts: map[uint32][]byte{
					2: []byte("pseudoTx2"),
				},
			},
			expectedChainTxs: txsVector("tx1", "tx2"),
			errContains:      "",
		},
		{
			name: "multiple sub-handlers - all returning errors",
			subHandlersFn: func(
				req *cmtabci.RequestPrepareProposal,
			) map[VoteExtensionPart]IProposalHandler {
				subHandler1 := newMockProposalHandler()
				subHandler2 := newMockProposalHandler()

				subHandler1.prepareProposalHandler.On(
					"call",
					mock.Anything,
					req,
				).Return(
					nil,
					fmt.Errorf("sub-handler 1 error"),
				)

				subHandler2.prepareProposalHandler.On(
					"call",
					mock.Anything,
					req,
				).Return(
					nil,
					fmt.Errorf("sub-handler 2 error"),
				)

				return map[VoteExtensionPart]IProposalHandler{
					VoteExtensionPart(1): subHandler1,
					VoteExtensionPart(2): subHandler2,
				}
			},
			reqHeight:          101,
			reqTxs:             txsVector("tx1", "tx2"),
			expectedInjectedTx: nil,
			expectedChainTxs:   nil,
			errContains:        "all sub-handlers failed to prepare proposal",
		},
		{
			name: "max tx bytes exceeded by chain txs",
			subHandlersFn: func(
				req *cmtabci.RequestPrepareProposal,
			) map[VoteExtensionPart]IProposalHandler {
				subHandler := newMockProposalHandler()

				subHandler.prepareProposalHandler.On(
					"call",
					mock.Anything,
					req,
				).Return(
					&cmtabci.ResponsePrepareProposal{
						Txs: append(
							[][]byte{[]byte("pseudoTx1")},
							req.Txs...,
						),
					},
					nil,
				)

				return map[VoteExtensionPart]IProposalHandler{
					VoteExtensionPart(1): subHandler,
				}
			},
			reqHeight: 101,
			reqTxs:    txsVector("tx1", "tx2", strings.Repeat("bigTx", 10)),
			expectedInjectedTx: &types.InjectedTx{
				Parts: map[uint32][]byte{1: []byte("pseudoTx1")},
			},
			expectedChainTxs: txsVector("tx1", "tx2"),
			errContains:      "",
		},
		{
			name: "max tx bytes exceeded by injected pseudo-tx",
			subHandlersFn: func(
				req *cmtabci.RequestPrepareProposal,
			) map[VoteExtensionPart]IProposalHandler {
				subHandler := newMockProposalHandler()

				subHandler.prepareProposalHandler.On(
					"call",
					mock.Anything,
					req,
				).Return(
					&cmtabci.ResponsePrepareProposal{
						Txs: append(
							[][]byte{[]byte(strings.Repeat("bigPseudoTx1", 10))},
							req.Txs...,
						),
					},
					nil,
				)

				return map[VoteExtensionPart]IProposalHandler{
					VoteExtensionPart(1): subHandler,
				}
			},
			reqHeight:          101,
			reqTxs:             txsVector("tx1", "tx2"),
			expectedInjectedTx: nil,
			expectedChainTxs:   txsVector(),
			errContains:        "",
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			s.SetupTest()

			req := &cmtabci.RequestPrepareProposal{
				// Fill only the fields that are used in the tests.
				Height:     test.reqHeight,
				Txs:        test.reqTxs,
				MaxTxBytes: s.maxTxBytes,
			}

			subHandlers := test.subHandlersFn(req)

			s.handler = &ProposalHandler{
				logger:      s.logger,
				subHandlers: subHandlers,
			}

			res, err := s.handler.PrepareProposalHandler()(s.ctx, req)

			for _, subHandler := range subHandlers {
				subHandler.(*mockProposalHandler).prepareProposalHandler.AssertCalled(
					s.T(),
					"call",
					s.ctx,
					req,
				)
			}

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

			var actualInjectedTx *types.InjectedTx
			var actualChainTxs [][]byte

			if res != nil {
				if injectedTxBytes := extractInjectedTx(req, res); len(injectedTxBytes) > 0 {
					actualInjectedTx = new(types.InjectedTx)
					err = actualInjectedTx.Unmarshal(injectedTxBytes)
					s.Require().NoError(err)

					actualChainTxs = res.Txs[1:]
				} else {
					actualChainTxs = res.Txs
				}
			}

			s.Require().Equal(
				test.expectedInjectedTx,
				actualInjectedTx,
				"expected different injected pseudo-tx",
			)

			s.Require().Equal(
				test.expectedChainTxs,
				actualChainTxs,
				"expected different chain transactions",
			)
		})
	}
}

type mockPrepareProposalHandler struct {
	mock.Mock
}

func (mpph *mockPrepareProposalHandler) call(
	ctx sdk.Context,
	req *cmtabci.RequestPrepareProposal,
) (*cmtabci.ResponsePrepareProposal, error) {
	args := mpph.Called(ctx, req)

	if res := args.Get(0); res != nil {
		return res.(*cmtabci.ResponsePrepareProposal), args.Error(1)
	}

	return nil, args.Error(1)
}

type mockProcessProposalHandler struct {
	mock.Mock
}

func (mpph *mockProcessProposalHandler) call(
	ctx sdk.Context,
	req *cmtabci.RequestProcessProposal,
) (*cmtabci.ResponseProcessProposal, error) {
	args := mpph.Called(ctx, req)

	if res := args.Get(0); res != nil {
		return res.(*cmtabci.ResponseProcessProposal), args.Error(1)
	}

	return nil, args.Error(1)
}

type mockProposalHandler struct {
	prepareProposalHandler *mockPrepareProposalHandler
	processProposalHandler *mockProcessProposalHandler
}

func newMockProposalHandler() *mockProposalHandler {
	return &mockProposalHandler{
		prepareProposalHandler: &mockPrepareProposalHandler{},
		processProposalHandler: &mockProcessProposalHandler{},
	}
}

func (mph *mockProposalHandler) PrepareProposalHandler() sdk.PrepareProposalHandler {
	return mph.prepareProposalHandler.call
}

func (mph *mockProposalHandler) ProcessProposalHandler() sdk.ProcessProposalHandler {
	return mph.processProposalHandler.call
}
