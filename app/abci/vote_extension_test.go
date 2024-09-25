package abci

import (
	"fmt"
	"testing"
	"time"

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
		name                   string
		subHandlersFn          func() map[VoteExtensionPart]IVoteExtensionHandler
		reqInjectedTx          *types.InjectedTx
		reqChainTxs            [][]byte
		expectedSubHandlersTxs map[VoteExtensionPart][][]byte
		expectedVE             *types.VoteExtension
		errContains            string
	}{
		{
			name:                   "no sub-handlers",
			subHandlersFn:          func() map[VoteExtensionPart]IVoteExtensionHandler { return nil },
			reqInjectedTx:          nil,
			reqChainTxs:            txsVector(),
			expectedSubHandlersTxs: nil,
			expectedVE:             nil,
			errContains:            "all sub-handlers failed to extend vote",
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
			reqInjectedTx: nil,
			reqChainTxs:   txsVector(),
			expectedSubHandlersTxs: map[VoteExtensionPart][][]byte{
				VoteExtensionPart(1): append([][]byte{nil}, txsVector()...),
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
			reqInjectedTx: nil,
			reqChainTxs:   txsVector(),
			expectedSubHandlersTxs: map[VoteExtensionPart][][]byte{
				VoteExtensionPart(1): append([][]byte{nil}, txsVector()...),
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
			reqInjectedTx: nil,
			reqChainTxs:   txsVector(),
			expectedSubHandlersTxs: map[VoteExtensionPart][][]byte{
				VoteExtensionPart(1): append([][]byte{nil}, txsVector()...),
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
			reqInjectedTx: nil,
			reqChainTxs:   txsVector(),
			expectedSubHandlersTxs: map[VoteExtensionPart][][]byte{
				VoteExtensionPart(1): append([][]byte{nil}, txsVector()...),
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
			reqInjectedTx: nil,
			reqChainTxs:   txsVector(),
			expectedSubHandlersTxs: map[VoteExtensionPart][][]byte{
				VoteExtensionPart(1): append([][]byte{nil}, txsVector()...),
				VoteExtensionPart(2): append([][]byte{nil}, txsVector()...),
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
			reqInjectedTx: nil,
			reqChainTxs:   txsVector(),
			expectedSubHandlersTxs: map[VoteExtensionPart][][]byte{
				VoteExtensionPart(1): append([][]byte{nil}, txsVector()...),
				VoteExtensionPart(2): append([][]byte{nil}, txsVector()...),
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
			reqInjectedTx: nil,
			reqChainTxs:   txsVector(),
			expectedSubHandlersTxs: map[VoteExtensionPart][][]byte{
				VoteExtensionPart(1): append([][]byte{nil}, txsVector()...),
				VoteExtensionPart(2): append([][]byte{nil}, txsVector()...),
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
			reqInjectedTx: nil,
			reqChainTxs:   txsVector(),
			expectedSubHandlersTxs: map[VoteExtensionPart][][]byte{
				VoteExtensionPart(1): append([][]byte{nil}, txsVector()...),
				VoteExtensionPart(2): append([][]byte{nil}, txsVector()...),
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
			reqInjectedTx: nil,
			reqChainTxs:   txsVector(),
			expectedSubHandlersTxs: map[VoteExtensionPart][][]byte{
				VoteExtensionPart(1): append([][]byte{nil}, txsVector()...),
				VoteExtensionPart(2): append([][]byte{nil}, txsVector()...),
			},
			expectedVE:  nil,
			errContains: "all sub-handlers failed to extend vote",
		},
		{
			name: "injected tx present - holding sub-handler part",
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
			reqInjectedTx: &types.InjectedTx{
				Parts: map[uint32][]byte{1: []byte("pseudoTx1")},
			},
			reqChainTxs: txsVector("tx1", "tx2"),
			expectedSubHandlersTxs: map[VoteExtensionPart][][]byte{
				VoteExtensionPart(1): append([][]byte{[]byte("pseudoTx1")}, txsVector("tx1", "tx2")...),
			},
			expectedVE: &types.VoteExtension{
				Height: s.requestHeight,
				Parts:  map[uint32][]byte{1: []byte("part1")},
			},
			errContains: "",
		},
		{
			name: "injected tx present - not holding sub-handler part",
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
			reqInjectedTx: &types.InjectedTx{
				Parts: map[uint32][]byte{},
			},
			reqChainTxs: txsVector("tx1", "tx2"),
			expectedSubHandlersTxs: map[VoteExtensionPart][][]byte{
				VoteExtensionPart(1): append([][]byte{nil}, txsVector("tx1", "tx2")...),
			},
			expectedVE: &types.VoteExtension{
				Height: s.requestHeight,
				Parts:  map[uint32][]byte{1: []byte("part1")},
			},
			errContains: "",
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			s.SetupTest()

			subHandlers := test.subHandlersFn()

			s.handler = &VoteExtensionHandler{
				logger:      s.logger,
				subHandlers: subHandlers,
			}

			now := time.Now()

			txs := txsVector()
			if injectedTx := test.reqInjectedTx; injectedTx != nil {
				injectedTxBytes, err := injectedTx.Marshal()
				s.Require().NoError(err)
				txs = append(txs, injectedTxBytes)
			}
			txs = append(txs, test.reqChainTxs...)

			req := &cmtabci.RequestExtendVote{
				Hash:               []byte("hash"),
				Height:             s.requestHeight,
				Time:               now,
				Txs:                txs,
				ProposedLastCommit: cmtabci.CommitInfo{},
				Misbehavior:        []cmtabci.Misbehavior{},
				NextValidatorsHash: []byte("nextValidatorsHash"),
				ProposerAddress:    []byte("proposerAddress"),
			}

			res, err := s.handler.ExtendVoteHandler()(s.ctx, req)

			for part, subHandler := range subHandlers {
				subHandler.(*mockVoteExtensionHandler).extendVoteHandler.AssertCalled(
					s.T(),
					"call",
					s.ctx,
					&cmtabci.RequestExtendVote{
						Hash:               req.Hash,
						Height:             req.Height,
						Time:               req.Time,
						Txs:                test.expectedSubHandlersTxs[part],
						ProposedLastCommit: req.ProposedLastCommit,
						Misbehavior:        req.Misbehavior,
						NextValidatorsHash: req.NextValidatorsHash,
						ProposerAddress:    req.ProposerAddress,
					},
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

func (s *VoteExtensionHandlerTestSuite) TestVerifyVoteExtension() {
	marshalVE := func(ve types.VoteExtension) []byte {
		veBytes, err := ve.Marshal()
		s.Require().NoError(err)
		return veBytes
	}

	extractVEPart := func(veBytes []byte, part VoteExtensionPart) []byte {
		var ve types.VoteExtension
		err := ve.Unmarshal(veBytes)
		s.Require().NoError(err)
		return ve.Parts[uint32(part)]
	}

	tests := []struct {
		name              string
		subHandlersFn     func() map[VoteExtensionPart]IVoteExtensionHandler
		voteExtensionFn   func() []byte
		expectedRes       *cmtabci.ResponseVerifyVoteExtension
		subHandlersCalled map[VoteExtensionPart]bool // true if expected to be called, false otherwise
		errContains       string
	}{
		{
			name:            "empty vote extension",
			subHandlersFn:   func() map[VoteExtensionPart]IVoteExtensionHandler { return nil },
			voteExtensionFn: func() []byte { return []byte{} },
			expectedRes: &cmtabci.ResponseVerifyVoteExtension{
				Status: cmtabci.ResponseVerifyVoteExtension_ACCEPT,
			},
			subHandlersCalled: nil,
			errContains:       "",
		},
		{
			name:            "nil vote extension",
			subHandlersFn:   func() map[VoteExtensionPart]IVoteExtensionHandler { return nil },
			voteExtensionFn: func() []byte { return nil },
			expectedRes: &cmtabci.ResponseVerifyVoteExtension{
				Status: cmtabci.ResponseVerifyVoteExtension_ACCEPT,
			},
			subHandlersCalled: nil,
			errContains:       "",
		},
		{
			name:              "non-unmarshalable vote extension",
			subHandlersFn:     func() map[VoteExtensionPart]IVoteExtensionHandler { return nil },
			voteExtensionFn:   func() []byte { return []byte("corrupted") },
			expectedRes:       nil,
			subHandlersCalled: nil,
			errContains:       "failed to unmarshal vote extension",
		},
		{
			name:          "wrong-height vote extension",
			subHandlersFn: func() map[VoteExtensionPart]IVoteExtensionHandler { return nil },
			voteExtensionFn: func() []byte {
				return marshalVE(
					types.VoteExtension{
						Height: s.requestHeight + 1,
						Parts:  map[uint32][]byte{1: []byte("part1")},
					},
				)
			},
			expectedRes:       nil,
			subHandlersCalled: nil,
			errContains:       "vote extension height does not match request height",
		},
		{
			name:          "empty-parts vote extension",
			subHandlersFn: func() map[VoteExtensionPart]IVoteExtensionHandler { return nil },
			voteExtensionFn: func() []byte {
				return marshalVE(
					types.VoteExtension{
						Height: s.requestHeight,
						Parts:  map[uint32][]byte{},
					},
				)
			},
			expectedRes:       nil,
			subHandlersCalled: nil,
			errContains:       "vote extension has no parts",
		},
		{
			name:          "nil-parts vote extension",
			subHandlersFn: func() map[VoteExtensionPart]IVoteExtensionHandler { return nil },
			voteExtensionFn: func() []byte {
				return marshalVE(
					types.VoteExtension{
						Height: s.requestHeight,
						Parts:  nil,
					},
				)
			},
			expectedRes:       nil,
			subHandlersCalled: nil,
			errContains:       "vote extension has no parts",
		},
		{
			name:          "no sub-handlers",
			subHandlersFn: func() map[VoteExtensionPart]IVoteExtensionHandler { return nil },
			voteExtensionFn: func() []byte {
				return marshalVE(
					types.VoteExtension{
						Height: s.requestHeight,
						Parts:  map[uint32][]byte{1: []byte("part1")},
					},
				)
			},
			expectedRes:       nil,
			subHandlersCalled: nil,
			errContains:       "unknown vote extension part",
		},
		{
			name: "single sub-handler - unknown vote extension part",
			subHandlersFn: func() map[VoteExtensionPart]IVoteExtensionHandler {
				subHandler := newMockVoteExtensionHandler()

				subHandler.verifyVoteExtensionHandler.On(
					"call",
					mock.Anything,
					mock.Anything,
				).Return(
					&cmtabci.ResponseVerifyVoteExtension{
						Status: cmtabci.ResponseVerifyVoteExtension_ACCEPT,
					}, nil,
				)

				return map[VoteExtensionPart]IVoteExtensionHandler{
					VoteExtensionPart(1): subHandler,
				}
			},
			voteExtensionFn: func() []byte {
				return marshalVE(
					types.VoteExtension{
						Height: s.requestHeight,
						Parts:  map[uint32][]byte{2: []byte("part2")},
					},
				)
			},
			expectedRes: nil,
			subHandlersCalled: map[VoteExtensionPart]bool{
				VoteExtensionPart(1): false,
			},
			errContains: "unknown vote extension part",
		},
		{
			name: "single sub-handler returning error",
			subHandlersFn: func() map[VoteExtensionPart]IVoteExtensionHandler {
				subHandler := newMockVoteExtensionHandler()

				subHandler.verifyVoteExtensionHandler.On(
					"call",
					mock.Anything,
					mock.Anything,
				).Return(
					nil, fmt.Errorf("sub-handler error"),
				)

				return map[VoteExtensionPart]IVoteExtensionHandler{
					VoteExtensionPart(1): subHandler,
				}
			},
			voteExtensionFn: func() []byte {
				return marshalVE(
					types.VoteExtension{
						Height: s.requestHeight,
						Parts:  map[uint32][]byte{1: []byte("part1")},
					},
				)
			},
			expectedRes: nil,
			subHandlersCalled: map[VoteExtensionPart]bool{
				VoteExtensionPart(1): true,
			},
			errContains: "sub-handler failed to verify vote extension part bridge",
		},
		{
			name: "single sub-handler rejecting vote extension",
			subHandlersFn: func() map[VoteExtensionPart]IVoteExtensionHandler {
				subHandler := newMockVoteExtensionHandler()

				subHandler.verifyVoteExtensionHandler.On(
					"call",
					mock.Anything,
					mock.Anything,
				).Return(
					&cmtabci.ResponseVerifyVoteExtension{
						Status: cmtabci.ResponseVerifyVoteExtension_REJECT,
					}, nil,
				)

				return map[VoteExtensionPart]IVoteExtensionHandler{
					VoteExtensionPart(1): subHandler,
				}
			},
			voteExtensionFn: func() []byte {
				return marshalVE(
					types.VoteExtension{
						Height: s.requestHeight,
						Parts:  map[uint32][]byte{1: []byte("part1")},
					},
				)
			},
			expectedRes: nil,
			subHandlersCalled: map[VoteExtensionPart]bool{
				VoteExtensionPart(1): true,
			},
			errContains: "sub-handler rejected vote extension part bridge",
		},
		{
			name: "single sub-handler accepting vote extension",
			subHandlersFn: func() map[VoteExtensionPart]IVoteExtensionHandler {
				subHandler := newMockVoteExtensionHandler()

				subHandler.verifyVoteExtensionHandler.On(
					"call",
					mock.Anything,
					mock.Anything,
				).Return(
					&cmtabci.ResponseVerifyVoteExtension{
						Status: cmtabci.ResponseVerifyVoteExtension_ACCEPT,
					}, nil,
				)

				return map[VoteExtensionPart]IVoteExtensionHandler{
					VoteExtensionPart(1): subHandler,
				}
			},
			voteExtensionFn: func() []byte {
				return marshalVE(
					types.VoteExtension{
						Height: s.requestHeight,
						Parts:  map[uint32][]byte{1: []byte("part1")},
					},
				)
			},
			expectedRes: &cmtabci.ResponseVerifyVoteExtension{
				Status: cmtabci.ResponseVerifyVoteExtension_ACCEPT,
			},
			subHandlersCalled: map[VoteExtensionPart]bool{
				VoteExtensionPart(1): true,
			},
			errContains: "",
		},
		{
			name: "multiple sub-handlers - unknown vote extension part",
			subHandlersFn: func() map[VoteExtensionPart]IVoteExtensionHandler {
				subHandler1 := newMockVoteExtensionHandler()
				subHandler2 := newMockVoteExtensionHandler()

				subHandler1.verifyVoteExtensionHandler.On(
					"call",
					mock.Anything,
					mock.Anything,
				).Return(
					&cmtabci.ResponseVerifyVoteExtension{
						Status: cmtabci.ResponseVerifyVoteExtension_ACCEPT,
					}, nil,
				)

				subHandler2.verifyVoteExtensionHandler.On(
					"call",
					mock.Anything,
					mock.Anything,
				).Return(
					&cmtabci.ResponseVerifyVoteExtension{
						Status: cmtabci.ResponseVerifyVoteExtension_ACCEPT,
					}, nil,
				)

				return map[VoteExtensionPart]IVoteExtensionHandler{
					VoteExtensionPart(1): subHandler1,
					VoteExtensionPart(2): subHandler2,
				}
			},
			voteExtensionFn: func() []byte {
				return marshalVE(
					types.VoteExtension{
						Height: s.requestHeight,
						Parts: map[uint32][]byte{
							3: []byte("part3"),
						},
					},
				)
			},
			expectedRes: nil,
			subHandlersCalled: map[VoteExtensionPart]bool{
				VoteExtensionPart(1): false,
				VoteExtensionPart(2): false,
			},
			errContains: "unknown vote extension part",
		},
		{
			name: "multiple sub-handlers - one returning error",
			subHandlersFn: func() map[VoteExtensionPart]IVoteExtensionHandler {
				subHandler1 := newMockVoteExtensionHandler()
				subHandler2 := newMockVoteExtensionHandler()

				subHandler1.verifyVoteExtensionHandler.On(
					"call",
					mock.Anything,
					mock.Anything,
				).Return(
					&cmtabci.ResponseVerifyVoteExtension{
						Status: cmtabci.ResponseVerifyVoteExtension_ACCEPT,
					}, nil,
				)

				subHandler2.verifyVoteExtensionHandler.On(
					"call",
					mock.Anything,
					mock.Anything,
				).Return(
					nil, fmt.Errorf("sub-handler 2 error"),
				)

				return map[VoteExtensionPart]IVoteExtensionHandler{
					VoteExtensionPart(1): subHandler1,
					VoteExtensionPart(2): subHandler2,
				}
			},
			voteExtensionFn: func() []byte {
				return marshalVE(
					types.VoteExtension{
						Height: s.requestHeight,
						Parts: map[uint32][]byte{
							1: []byte("part1"),
							2: []byte("part2"),
						},
					},
				)
			},
			expectedRes: nil,
			subHandlersCalled: map[VoteExtensionPart]bool{
				// Sub-handler 2 should be always called but whether
				// sub-handler 1 is called depends on the order of the
				// iteration over the map. If sub-handler 2 is called first,
				// error will be returned before calling sub-handler 1. That
				// said, we cannot determine whether sub-handler 1 will be
				// called or not.
				VoteExtensionPart(2): true,
			},
			errContains: "sub-handler failed to verify vote extension part",
		},
		{
			name: "multiple sub-handlers - one rejecting vote extension",
			subHandlersFn: func() map[VoteExtensionPart]IVoteExtensionHandler {
				subHandler1 := newMockVoteExtensionHandler()
				subHandler2 := newMockVoteExtensionHandler()

				subHandler1.verifyVoteExtensionHandler.On(
					"call",
					mock.Anything,
					mock.Anything,
				).Return(
					&cmtabci.ResponseVerifyVoteExtension{
						Status: cmtabci.ResponseVerifyVoteExtension_ACCEPT,
					}, nil,
				)

				subHandler2.verifyVoteExtensionHandler.On(
					"call",
					mock.Anything,
					mock.Anything,
				).Return(
					&cmtabci.ResponseVerifyVoteExtension{
						Status: cmtabci.ResponseVerifyVoteExtension_REJECT,
					}, nil,
				)

				return map[VoteExtensionPart]IVoteExtensionHandler{
					VoteExtensionPart(1): subHandler1,
					VoteExtensionPart(2): subHandler2,
				}
			},
			voteExtensionFn: func() []byte {
				return marshalVE(
					types.VoteExtension{
						Height: s.requestHeight,
						Parts: map[uint32][]byte{
							1: []byte("part1"),
							2: []byte("part2"),
						},
					},
				)
			},
			expectedRes: nil,
			subHandlersCalled: map[VoteExtensionPart]bool{
				// Sub-handler 2 should be always called but whether
				// sub-handler 1 is called depends on the order of the
				// iteration over the map. If sub-handler 2 is called first,
				// error will be returned before calling sub-handler 1. That
				// said, we cannot determine whether sub-handler 1 will be
				// called or not.
				VoteExtensionPart(2): true,
			},
			errContains: "sub-handler rejected vote extension part",
		},
		{
			name: "multiple sub-handlers accepting vote extension",
			subHandlersFn: func() map[VoteExtensionPart]IVoteExtensionHandler {
				subHandler1 := newMockVoteExtensionHandler()
				subHandler2 := newMockVoteExtensionHandler()

				subHandler1.verifyVoteExtensionHandler.On(
					"call",
					mock.Anything,
					mock.Anything,
				).Return(
					&cmtabci.ResponseVerifyVoteExtension{
						Status: cmtabci.ResponseVerifyVoteExtension_ACCEPT,
					}, nil,
				)

				subHandler2.verifyVoteExtensionHandler.On(
					"call",
					mock.Anything,
					mock.Anything,
				).Return(
					&cmtabci.ResponseVerifyVoteExtension{
						Status: cmtabci.ResponseVerifyVoteExtension_ACCEPT,
					}, nil,
				)

				return map[VoteExtensionPart]IVoteExtensionHandler{
					VoteExtensionPart(1): subHandler1,
					VoteExtensionPart(2): subHandler2,
				}
			},
			voteExtensionFn: func() []byte {
				return marshalVE(
					types.VoteExtension{
						Height: s.requestHeight,
						Parts: map[uint32][]byte{
							1: []byte("part1"),
							2: []byte("part2"),
						},
					},
				)
			},
			expectedRes: &cmtabci.ResponseVerifyVoteExtension{
				Status: cmtabci.ResponseVerifyVoteExtension_ACCEPT,
			},
			subHandlersCalled: map[VoteExtensionPart]bool{
				VoteExtensionPart(1): true,
				VoteExtensionPart(2): true,
			},
			errContains: "",
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			s.SetupTest()

			subHandlers := test.subHandlersFn()

			s.handler = &VoteExtensionHandler{
				logger:      s.logger,
				subHandlers: subHandlers,
			}

			req := &cmtabci.RequestVerifyVoteExtension{
				Hash:             []byte("hash"),
				ValidatorAddress: []byte("validatorAddress"),
				Height:           s.requestHeight,
				VoteExtension:    test.voteExtensionFn(),
			}

			res, err := s.handler.VerifyVoteExtensionHandler()(s.ctx, req)

			// Make sure sub-handlers calls were as expected.
			for part, expected := range test.subHandlersCalled {
				subHandler, ok := subHandlers[part]
				s.Require().True(ok)

				fn := subHandler.(*mockVoteExtensionHandler).verifyVoteExtensionHandler

				if expected {
					fn.AssertCalled(
						s.T(),
						"call",
						s.ctx,
						&cmtabci.RequestVerifyVoteExtension{
							Hash:             req.Hash,
							ValidatorAddress: req.ValidatorAddress,
							Height:           req.Height,
							VoteExtension:    extractVEPart(req.VoteExtension, part),
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
