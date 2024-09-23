package abci

import (
	"cosmossdk.io/log"
	"fmt"
	"strings"
	"testing"
	"time"

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
			name: "single sub-handler injecting non-empty pseudo-tx as only tx in proposal",
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
			reqTxs:    txsVector(),
			expectedInjectedTx: &types.InjectedTx{
				Parts: map[uint32][]byte{1: []byte("pseudoTx1")},
			},
			expectedChainTxs: txsVector(),
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
			name: "max tx bytes exceeded - chain txs",
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
			name: "max tx bytes exceeded - injected pseudo-tx",
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

func (s *ProposalHandlerTestSuite) TestProcessProposal() {
	marshalInjectedTx := func(injectedTx types.InjectedTx) []byte {
		injectedTxBytes, err := injectedTx.Marshal()
		s.Require().NoError(err)
		return injectedTxBytes
	}

	extractInjectedTxPart := func(
		injectedTxBytes []byte,
		part VoteExtensionPart,
	) []byte {
		var injectedTx types.InjectedTx
		err := injectedTx.Unmarshal(injectedTxBytes)
		s.Require().NoError(err)
		return injectedTx.Parts[uint32(part)]
	}

	tests := []struct {
		name              string
		subHandlersFn     func() map[VoteExtensionPart]IProposalHandler
		reqHeight         int64
		reqTxsFn		  func() [][]byte
		expectedRes       *cmtabci.ResponseProcessProposal
		subHandlersCalled map[VoteExtensionPart]bool // true if expected to be called, false otherwise
		errContains       string
	}{
		{
			name: "vote extensions not enabled",
			subHandlersFn: func() map[VoteExtensionPart]IProposalHandler { return nil },
			// Vote extensions become enabled at height 100. However, the proposal
			// handler looks at the vote extensions from the previous block, so
			// from the handler's perspective, they are not enabled yet. The proposal
			// handler becomes fully functional at height 101.
			reqHeight: 100,
			reqTxsFn:  func() [][]byte { return txsVector("tx1", "tx2") },
			expectedRes: &cmtabci.ResponseProcessProposal{
				Status: cmtabci.ResponseProcessProposal_ACCEPT,
			},
			subHandlersCalled: nil,
			errContains:       "",
		},
		{
			name: "no txs in the request",
			subHandlersFn: func() map[VoteExtensionPart]IProposalHandler { return nil },
			reqHeight: 101,
			reqTxsFn:  func() [][]byte { return txsVector() },
			expectedRes: &cmtabci.ResponseProcessProposal{
				Status: cmtabci.ResponseProcessProposal_ACCEPT,
			},
			subHandlersCalled: nil,
			errContains:       "",
		},
		{
			name:              "non-unmarshalable injected tx",
			subHandlersFn:     func() map[VoteExtensionPart]IProposalHandler { return nil },
			reqHeight:         101,
			reqTxsFn:          func() [][]byte { return txsVector("corrupted") },
			expectedRes:       nil,
			subHandlersCalled: nil,
			errContains:       "failed to unmarshal injected tx",
		},
		{
			name:              "empty-parts injected tx",
			subHandlersFn:     func() map[VoteExtensionPart]IProposalHandler { return nil },
			reqHeight:         101,
			reqTxsFn:          func() [][]byte {
				return append(
					[][]byte{marshalInjectedTx(
						types.InjectedTx{Parts: map[uint32][]byte{}}),
					},
					txsVector("tx1", "tx2")...,
				)
			},
			expectedRes:       nil,
			subHandlersCalled: nil,
			errContains:       "injected tx has no parts",
		},
		{
			name:              "nil-parts injected tx",
			subHandlersFn:     func() map[VoteExtensionPart]IProposalHandler { return nil },
			reqHeight:         101,
			reqTxsFn:          func() [][]byte {
				return append(
					[][]byte{marshalInjectedTx(
						types.InjectedTx{Parts: nil}),
					},
					txsVector("tx1", "tx2")...,
				)
			},
			expectedRes:       nil,
			subHandlersCalled: nil,
			errContains:       "injected tx has no parts",
		},

		// TODO: More test cases.
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			s.SetupTest()

			now := time.Now()

			req := &cmtabci.RequestProcessProposal{
				Txs:                test.reqTxsFn(),
				ProposedLastCommit: cmtabci.CommitInfo{},
				Misbehavior:        []cmtabci.Misbehavior{},
				Hash:               []byte("hash"),
				Height:             test.reqHeight,
				Time:               now,
				NextValidatorsHash: []byte("nextValidatorsHash"),
				ProposerAddress:    []byte("proposerAddress"),
			}

			subHandlers := test.subHandlersFn()

			s.handler = &ProposalHandler{
				logger:      s.logger,
				subHandlers: subHandlers,
			}

			res, err := s.handler.ProcessProposalHandler()(s.ctx, req)

			// Make sure sub-handlers calls were as expected.
			for part, expected := range test.subHandlersCalled {
				subHandler, ok := subHandlers[part]
				s.Require().True(ok)

				fn := subHandler.(*mockProposalHandler).processProposalHandler

				if expected {
					subTxs := req.Txs
					if len(req.Txs) > 0 {
						subTxs = append(
							[][]byte{
								extractInjectedTxPart(
									req.Txs[0],
									part,
								),
							}, req.Txs[1:]...,
						)
					}

					fn.AssertCalled(
						s.T(),
						"call",
						s.ctx,
						&cmtabci.RequestProcessProposal{
							Txs:                subTxs,
							ProposedLastCommit: cmtabci.CommitInfo{},
							Misbehavior:        []cmtabci.Misbehavior{},
							Hash:               []byte("hash"),
							Height:             test.reqHeight,
							Time:               now,
							NextValidatorsHash: []byte("nextValidatorsHash"),
							ProposerAddress:    []byte("proposerAddress"),
						},
					)
				} else {
					fn.AssertNotCalled(
						s.T(),
						"call",
						mock.Anything,
						mock.Anything,
					)
				}
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

			s.Require().Equal(
				test.expectedRes,
				res,
				"expected different response",
			)
		})
	}
}

func txsVector(txs ...string) [][]byte {
	res := make([][]byte, 0)

	for _, tx := range txs {
		res = append(res, []byte(tx))
	}

	return res
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
