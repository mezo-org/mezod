package abci

import (
	"fmt"
	"testing"
	"time"

	"cosmossdk.io/log"

	cmtabci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	servermock "github.com/cosmos/cosmos-sdk/server/mock"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/mezo-org/mezod/app/abci/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

func TestPreBlockHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(PreBlockHandlerTestSuite))
}

type PreBlockHandlerTestSuite struct {
	suite.Suite

	logger  log.Logger
	ctx     sdk.Context
	handler *PreBlockHandler
}

func (s *PreBlockHandlerTestSuite) SetupTest() {
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
}

func (s *PreBlockHandlerTestSuite) TestPreBlocker() {
	tests := []struct {
		name                   string
		mmFn                   func(sdk.Context) *mockModuleManager
		subHandlersFn          func() map[VoteExtensionPart]IPreBlockHandler
		reqHeight              int64
		reqInjectedTx          *types.InjectedTx
		reqChainTxs            [][]byte
		expectedSubHandlersTxs map[VoteExtensionPart][][]byte
		expectedRes            *sdk.ResponsePreBlock
		errContains            string
	}{
		{
			name: "module manager returning error",
			mmFn: func(ctx sdk.Context) *mockModuleManager {
				return newMockModuleManager(ctx, fmt.Errorf("module manager error"))
			},
			subHandlersFn: func() map[VoteExtensionPart]IPreBlockHandler {
				return nil
			},
			reqHeight:              101,
			reqInjectedTx:          nil,
			reqChainTxs:            txsVector(),
			expectedSubHandlersTxs: nil,
			expectedRes:            nil,
			errContains:            "module manager error",
		},
		{
			name: "vote extensions not enabled",
			mmFn: func(ctx sdk.Context) *mockModuleManager {
				return newMockModuleManager(ctx)
			},
			subHandlersFn: func() map[VoteExtensionPart]IPreBlockHandler {
				return nil
			},
			reqHeight:              100,
			reqInjectedTx:          nil,
			reqChainTxs:            txsVector(),
			expectedSubHandlersTxs: nil,
			expectedRes:            &sdk.ResponsePreBlock{},
			errContains:            "",
		},
		{
			name: "no sub-handlers",
			mmFn: func(ctx sdk.Context) *mockModuleManager {
				return newMockModuleManager(ctx)
			},
			subHandlersFn: func() map[VoteExtensionPart]IPreBlockHandler {
				return nil
			},
			reqHeight:              101,
			reqInjectedTx:          nil,
			reqChainTxs:            txsVector(),
			expectedSubHandlersTxs: nil,
			expectedRes:            &sdk.ResponsePreBlock{},
			errContains:            "",
		},
		{
			name: "single sub-handler executing with success",
			mmFn: func(ctx sdk.Context) *mockModuleManager {
				return newMockModuleManager(ctx)
			},
			subHandlersFn: func() map[VoteExtensionPart]IPreBlockHandler {
				subHandler := newMockPreBlockHandler()

				subHandler.preBlocker.On(
					"call",
					mock.Anything,
					mock.Anything,
				).Return(&sdk.ResponsePreBlock{}, nil)

				return map[VoteExtensionPart]IPreBlockHandler{
					VoteExtensionPart(1): subHandler,
				}
			},
			reqHeight:     101,
			reqInjectedTx: nil,
			reqChainTxs:   txsVector(),
			expectedSubHandlersTxs: map[VoteExtensionPart][][]byte{
				VoteExtensionPart(1): append([][]byte{nil}, txsVector()...),
			},
			expectedRes: &sdk.ResponsePreBlock{},
			errContains: "",
		},
		{
			name: "single sub-handler returning error",
			mmFn: func(ctx sdk.Context) *mockModuleManager {
				return newMockModuleManager(ctx)
			},
			subHandlersFn: func() map[VoteExtensionPart]IPreBlockHandler {
				subHandler := newMockPreBlockHandler()

				subHandler.preBlocker.On(
					"call",
					mock.Anything,
					mock.Anything,
				).Return(nil, fmt.Errorf("sub-handler error"))

				return map[VoteExtensionPart]IPreBlockHandler{
					VoteExtensionPart(1): subHandler,
				}
			},
			reqHeight:     101,
			reqInjectedTx: nil,
			reqChainTxs:   txsVector(),
			expectedSubHandlersTxs: map[VoteExtensionPart][][]byte{
				VoteExtensionPart(1): append([][]byte{nil}, txsVector()...),
			},
			expectedRes: nil,
			errContains: "sub-handler error",
		},
		{
			name: "multiple sub-handlers - all executing with success",
			mmFn: func(ctx sdk.Context) *mockModuleManager {
				return newMockModuleManager(ctx)
			},
			subHandlersFn: func() map[VoteExtensionPart]IPreBlockHandler {
				subHandler1 := newMockPreBlockHandler()
				subHandler2 := newMockPreBlockHandler()

				subHandler1.preBlocker.On(
					"call",
					mock.Anything,
					mock.Anything,
				).Return(&sdk.ResponsePreBlock{}, nil)

				subHandler2.preBlocker.On(
					"call",
					mock.Anything,
					mock.Anything,
				).Return(&sdk.ResponsePreBlock{}, nil)

				return map[VoteExtensionPart]IPreBlockHandler{
					VoteExtensionPart(1): subHandler1,
					VoteExtensionPart(2): subHandler2,
				}
			},
			reqHeight:     101,
			reqInjectedTx: nil,
			reqChainTxs:   txsVector(),
			expectedSubHandlersTxs: map[VoteExtensionPart][][]byte{
				VoteExtensionPart(1): append([][]byte{nil}, txsVector()...),
				VoteExtensionPart(2): append([][]byte{nil}, txsVector()...),
			},
			expectedRes: &sdk.ResponsePreBlock{},
			errContains: "",
		},
		{
			name: "multiple sub-handlers - one returning error",
			mmFn: func(ctx sdk.Context) *mockModuleManager {
				return newMockModuleManager(ctx)
			},
			subHandlersFn: func() map[VoteExtensionPart]IPreBlockHandler {
				subHandler1 := newMockPreBlockHandler()
				subHandler2 := newMockPreBlockHandler()

				subHandler1.preBlocker.On(
					"call",
					mock.Anything,
					mock.Anything,
				).Return(nil, fmt.Errorf("sub-handler 1 error"))

				subHandler2.preBlocker.On(
					"call",
					mock.Anything,
					mock.Anything,
				).Return(&sdk.ResponsePreBlock{}, nil)

				return map[VoteExtensionPart]IPreBlockHandler{
					VoteExtensionPart(1): subHandler1,
					VoteExtensionPart(2): subHandler2,
				}
			},
			reqHeight:     101,
			reqInjectedTx: nil,
			reqChainTxs:   txsVector(),
			expectedSubHandlersTxs: map[VoteExtensionPart][][]byte{
				VoteExtensionPart(1): append([][]byte{nil}, txsVector()...),
				// We do not expect sub-handler 2 to be called.
			},
			expectedRes: nil,
			errContains: "sub-handler 1 error",
		},
		{
			name: "injected tx present - holding sub-handler part",
			mmFn: func(ctx sdk.Context) *mockModuleManager {
				return newMockModuleManager(ctx)
			},
			subHandlersFn: func() map[VoteExtensionPart]IPreBlockHandler {
				subHandler := newMockPreBlockHandler()

				subHandler.preBlocker.On(
					"call",
					mock.Anything,
					mock.Anything,
				).Return(&sdk.ResponsePreBlock{}, nil)

				return map[VoteExtensionPart]IPreBlockHandler{
					VoteExtensionPart(1): subHandler,
				}
			},
			reqHeight: 101,
			reqInjectedTx: &types.InjectedTx{
				Parts: map[uint32][]byte{1: []byte("pseudoTx1")},
			},
			reqChainTxs: txsVector("tx1", "tx2"),
			expectedSubHandlersTxs: map[VoteExtensionPart][][]byte{
				VoteExtensionPart(1): append([][]byte{[]byte("pseudoTx1")}, txsVector("tx1", "tx2")...),
			},
			expectedRes: &sdk.ResponsePreBlock{},
			errContains: "",
		},
		{
			name: "injected tx present - not holding sub-handler part",
			mmFn: func(ctx sdk.Context) *mockModuleManager {
				return newMockModuleManager(ctx)
			},
			subHandlersFn: func() map[VoteExtensionPart]IPreBlockHandler {
				subHandler := newMockPreBlockHandler()

				subHandler.preBlocker.On(
					"call",
					mock.Anything,
					mock.Anything,
				).Return(&sdk.ResponsePreBlock{}, nil)

				return map[VoteExtensionPart]IPreBlockHandler{
					VoteExtensionPart(1): subHandler,
				}
			},
			reqHeight: 101,
			reqInjectedTx: &types.InjectedTx{
				Parts: map[uint32][]byte{},
			},
			reqChainTxs: txsVector("tx1", "tx2"),
			expectedSubHandlersTxs: map[VoteExtensionPart][][]byte{
				VoteExtensionPart(1): append([][]byte{nil}, txsVector("tx1", "tx2")...),
			},
			expectedRes: &sdk.ResponsePreBlock{},
			errContains: "",
		},
		{
			name: "injected tx not present but regular txs are",
			mmFn: func(ctx sdk.Context) *mockModuleManager {
				return newMockModuleManager(ctx)
			},
			subHandlersFn: func() map[VoteExtensionPart]IPreBlockHandler {
				subHandler := newMockPreBlockHandler()

				subHandler.preBlocker.On(
					"call",
					mock.Anything,
					mock.Anything,
				).Return(&sdk.ResponsePreBlock{}, nil)

				return map[VoteExtensionPart]IPreBlockHandler{
					VoteExtensionPart(1): subHandler,
				}
			},
			reqHeight:     101,
			reqInjectedTx: nil,
			reqChainTxs:   txsVector("tx1", "tx2"),
			expectedSubHandlersTxs: map[VoteExtensionPart][][]byte{
				VoteExtensionPart(1): append([][]byte{nil}, txsVector("tx1", "tx2")...),
			},
			expectedRes: &sdk.ResponsePreBlock{},
			errContains: "",
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			s.SetupTest()

			now := time.Now()

			txs := txsVector()
			if injectedTx := test.reqInjectedTx; injectedTx != nil {
				injectedTxBytes, err := injectedTx.Marshal()
				s.Require().NoError(err)
				txs = append(txs, injectedTxBytes)
			}
			txs = append(txs, test.reqChainTxs...)

			req := &cmtabci.RequestFinalizeBlock{
				Txs:                txs,
				DecidedLastCommit:  cmtabci.CommitInfo{},
				Misbehavior:        []cmtabci.Misbehavior{},
				Hash:               []byte("hash"),
				Height:             test.reqHeight,
				Time:               now,
				NextValidatorsHash: []byte("nextValidatorsHash"),
				ProposerAddress:    []byte("proposerAddress"),
			}

			subHandlers := test.subHandlersFn()

			s.handler = &PreBlockHandler{
				logger:      s.logger,
				subHandlers: subHandlers,
			}

			mm := test.mmFn(s.ctx)

			res, err := s.handler.PreBlocker(mm)(s.ctx, req)

			for part, subHandler := range subHandlers {
				txs, ok := test.expectedSubHandlersTxs[part]
				if !ok {
					// If the test case does not specify the expected txs for
					// the sub-handler, the sub-handler is not expected to be
					// called.
					subHandler.(*mockPreBlockHandler).preBlocker.AssertNumberOfCalls(
						s.T(),
						"call",
						0,
					)
					continue
				}

				subHandler.(*mockPreBlockHandler).preBlocker.AssertCalled(
					s.T(),
					"call",
					s.ctx,
					&cmtabci.RequestFinalizeBlock{
						Txs:                txs,
						DecidedLastCommit:  req.DecidedLastCommit,
						Misbehavior:        req.Misbehavior,
						Hash:               req.Hash,
						Height:             req.Height,
						Time:               req.Time,
						NextValidatorsHash: req.NextValidatorsHash,
						ProposerAddress:    req.ProposerAddress,
					},
				)
			}

			mm.AssertExpectations(s.T())
			// Regardless of the scenario, the module manager's PreBlock method
			// should be always called once.
			mm.AssertNumberOfCalls(s.T(), "PreBlock", 1)

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

type mockModuleManager struct {
	mock.Mock
}

func newMockModuleManager(ctx sdk.Context, err ...error) *mockModuleManager {
	mm := &mockModuleManager{}

	if len(err) == 1 {
		mm.On("PreBlock", ctx).Return(
			nil,
			err[0],
		)
	} else {
		mm.On("PreBlock", ctx).Return(
			&sdk.ResponsePreBlock{},
			nil,
		)
	}

	return mm
}

func (mmm *mockModuleManager) PreBlock(ctx sdk.Context) (*sdk.ResponsePreBlock, error) {
	args := mmm.Called(ctx)

	if res := args.Get(0); res != nil {
		return res.(*sdk.ResponsePreBlock), args.Error(1)
	}

	return nil, args.Error(1)
}

type mockPreBlocker struct {
	mock.Mock
}

func (mpb *mockPreBlocker) call(
	ctx sdk.Context,
	req *cmtabci.RequestFinalizeBlock,
) (*sdk.ResponsePreBlock, error) {
	args := mpb.Called(ctx, req)

	if res := args.Get(0); res != nil {
		return res.(*sdk.ResponsePreBlock), args.Error(1)
	}

	return nil, args.Error(1)
}

type mockPreBlockHandler struct {
	preBlocker *mockPreBlocker
}

func newMockPreBlockHandler() *mockPreBlockHandler {
	return &mockPreBlockHandler{
		preBlocker: &mockPreBlocker{},
	}
}

func (mpbh *mockPreBlockHandler) PreBlocker() sdk.PreBlocker {
	return mpbh.preBlocker.call
}
