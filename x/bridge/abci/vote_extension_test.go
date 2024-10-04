package abci

import (
	"context"
	"fmt"
	"testing"

	"github.com/mezo-org/mezod/cmd/config"

	sdkmath "cosmossdk.io/math"
	bridgetypes "github.com/mezo-org/mezod/x/bridge/types"

	"cosmossdk.io/log"

	cmtabci "github.com/cometbft/cometbft/abci/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	servermock "github.com/cosmos/cosmos-sdk/server/mock"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/mezo-org/mezod/x/bridge/abci/types"
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
	bridgeKeeper  *mockBridgeKeeper
	handler       *VoteExtensionHandler
}

func (s *VoteExtensionHandlerTestSuite) SetupTest() {
	// Set bech32 prefixes to make the recipient address validation in
	// AssetsLocked events possible (see AssetsLockedEvent.IsValid).
	cfg := sdk.GetConfig()
	config.SetBech32Prefixes(cfg)

	s.logger = log.NewNopLogger()

	s.ctx = sdk.NewContext(
		servermock.NewCommitMultiStore(),
		tmproto.Header{},
		false,
		s.logger,
	)

	s.requestHeight = 100

	s.bridgeKeeper = newMockBridgeKeeper()

	s.bridgeKeeper.On(
		"GetAssetsLockedSequenceTip",
		s.ctx,
	).Return(sdkmath.NewInt(200))
}

func (s *VoteExtensionHandlerTestSuite) TestExtendVote() {
	marshalInjectedTx := func(injectedTx types.InjectedTx) []byte {
		injectedTxBytes, err := injectedTx.Marshal()
		s.Require().NoError(err)
		return injectedTxBytes
	}

	// Default sequence boundaries where there is no injected tx in the proposal
	// and sequence tip is fetched from state.
	defaultSequenceStart := sdkmath.NewInt(201)
	defaultSequenceEnd := sdkmath.NewInt(211)

	tests := []struct {
		name        string
		sidecarFn   func() EthereumSidecarClient
		reqTxs      [][]byte
		expectedVE  *types.VoteExtension
		errContains string

	}{
		{
			name: "sidecar returning error",
			sidecarFn: func() EthereumSidecarClient {
				sidecar := newMockEthereumSidecarClient()

				sidecar.On(
					"GetAssetsLockedEvents",
					s.ctx,
					defaultSequenceStart,
					defaultSequenceEnd,
				).Return(nil, fmt.Errorf("sidecar error"))

				return sidecar
			},
			reqTxs:      txsVector(),
			expectedVE:  nil,
			errContains: "failed to fetch AssetsLocked events from the sidecar",
		},
		{
			name: "sidecar returning empty slice",
			sidecarFn: func() EthereumSidecarClient {
				sidecar := newMockEthereumSidecarClient()

				sidecar.On(
					"GetAssetsLockedEvents",
					s.ctx,
					defaultSequenceStart,
					defaultSequenceEnd,
				).Return([]bridgetypes.AssetsLockedEvent{}, nil)

				return sidecar
			},
			reqTxs: txsVector(),
			expectedVE: &types.VoteExtension{
				AssetsLockedEvents: nil,
			},
			errContains: "",
		},
		{
			name: "sidecar returning nil slice",
			sidecarFn: func() EthereumSidecarClient {
				sidecar := newMockEthereumSidecarClient()

				sidecar.On(
					"GetAssetsLockedEvents",
					s.ctx,
					defaultSequenceStart,
					defaultSequenceEnd,
				).Return(nil, nil)

				return sidecar
			},
			reqTxs: txsVector(),
			expectedVE: &types.VoteExtension{
				AssetsLockedEvents: nil,
			},
			errContains: "",
		},
		{
			name: "sidecar returning single event",
			sidecarFn: func() EthereumSidecarClient {
				sidecar := newMockEthereumSidecarClient()

				sidecar.On(
					"GetAssetsLockedEvents",
					s.ctx,
					defaultSequenceStart,
					defaultSequenceEnd,
				).Return([]bridgetypes.AssetsLockedEvent{
					mockEvent(201, recipient1, 1000),
				}, nil)

				return sidecar
			},
			reqTxs: txsVector(),
			expectedVE: &types.VoteExtension{
				AssetsLockedEvents: []bridgetypes.AssetsLockedEvent{
					mockEvent(201, recipient1, 1000),
				},
			},
			errContains: "",
		},
		{
			name: "sidecar returning improper sequence - strictly decreasing",
			sidecarFn: func() EthereumSidecarClient {
				sidecar := newMockEthereumSidecarClient()

				sidecar.On(
					"GetAssetsLockedEvents",
					s.ctx,
					defaultSequenceStart,
					defaultSequenceEnd,
				).Return([]bridgetypes.AssetsLockedEvent{
					mockEvent(203, recipient1, 1000),
					mockEvent(202, recipient1, 1000),
					mockEvent(201, recipient1, 1000),
				}, nil)

				return sidecar
			},
			reqTxs:      txsVector(),
			expectedVE:  nil,
			errContains: "events list is not valid",
		},
		{
			name: "sidecar returning improper sequence - increasing (non-strictly)",
			sidecarFn: func() EthereumSidecarClient {
				sidecar := newMockEthereumSidecarClient()

				sidecar.On(
					"GetAssetsLockedEvents",
					s.ctx,
					defaultSequenceStart,
					defaultSequenceEnd,
				).Return([]bridgetypes.AssetsLockedEvent{
					mockEvent(201, recipient1, 1000),
					mockEvent(201, recipient1, 1000),
					mockEvent(202, recipient1, 1000),
					mockEvent(203, recipient1, 1000),
				}, nil)

				return sidecar
			},
			reqTxs:      txsVector(),
			expectedVE:  nil,
			errContains: "events list is not valid",
		},
		{
			name: "sidecar returning improper sequence - decreasing (non-strictly)",
			sidecarFn: func() EthereumSidecarClient {
				sidecar := newMockEthereumSidecarClient()

				sidecar.On(
					"GetAssetsLockedEvents",
					s.ctx,
					defaultSequenceStart,
					defaultSequenceEnd,
				).Return([]bridgetypes.AssetsLockedEvent{
					mockEvent(203, recipient1, 1000),
					mockEvent(203, recipient1, 1000),
					mockEvent(202, recipient1, 1000),
					mockEvent(201, recipient1, 1000),
				}, nil)

				return sidecar
			},
			reqTxs:      txsVector(),
			expectedVE:  nil,
			errContains: "events list is not valid",
		},
		{
			name: "sidecar returning improper sequence - gap",
			sidecarFn: func() EthereumSidecarClient {
				sidecar := newMockEthereumSidecarClient()

				sidecar.On(
					"GetAssetsLockedEvents",
					s.ctx,
					defaultSequenceStart,
					defaultSequenceEnd,
				).Return([]bridgetypes.AssetsLockedEvent{
					mockEvent(201, recipient1, 1000),
					mockEvent(202, recipient1, 1000),
					mockEvent(204, recipient1, 1000),
				}, nil)

				return sidecar
			},
			reqTxs:      txsVector(),
			expectedVE:  nil,
			errContains: "events list is not valid",
		},
		{
			name: "sidecar returning improper sequence - duplicate",
			sidecarFn: func() EthereumSidecarClient {
				sidecar := newMockEthereumSidecarClient()

				sidecar.On(
					"GetAssetsLockedEvents",
					s.ctx,
					defaultSequenceStart,
					defaultSequenceEnd,
				).Return([]bridgetypes.AssetsLockedEvent{
					mockEvent(201, recipient1, 1000),
					mockEvent(202, recipient1, 1000),
					mockEvent(203, recipient1, 1000),
					mockEvent(201, recipient1, 1000),
				}, nil)

				return sidecar
			},
			reqTxs:      txsVector(),
			expectedVE:  nil,
			errContains: "events list is not valid",
		},
		{
			name: "sidecar returning more events than the limit",
			sidecarFn: func() EthereumSidecarClient {
				sidecar := newMockEthereumSidecarClient()

				sidecar.On(
					"GetAssetsLockedEvents",
					s.ctx,
					defaultSequenceStart,
					defaultSequenceEnd,
				).Return([]bridgetypes.AssetsLockedEvent{
					mockEvent(201, recipient1, 1000),
					mockEvent(202, recipient1, 1000),
					mockEvent(203, recipient1, 1000),
					mockEvent(204, recipient1, 1000),
					mockEvent(205, recipient1, 1000),
					mockEvent(206, recipient1, 1000),
					mockEvent(207, recipient1, 1000),
					mockEvent(208, recipient1, 1000),
					mockEvent(209, recipient1, 1000),
					mockEvent(210, recipient1, 1000),
					mockEvent(211, recipient1, 1000),
				}, nil)

				return sidecar
			},
			reqTxs:      txsVector(),
			expectedVE:  nil,
			errContains: "number of events exceeds the limit",
		},
		{
			name: "sidecar returning events within the limit",
			sidecarFn: func() EthereumSidecarClient {
				sidecar := newMockEthereumSidecarClient()

				sidecar.On(
					"GetAssetsLockedEvents",
					s.ctx,
					defaultSequenceStart,
					defaultSequenceEnd,
				).Return([]bridgetypes.AssetsLockedEvent{
					mockEvent(201, recipient1, 1000),
					mockEvent(202, recipient1, 1000),
					mockEvent(203, recipient1, 1000),
					mockEvent(204, recipient1, 1000),
					mockEvent(205, recipient1, 1000),
					mockEvent(206, recipient1, 1000),
					mockEvent(207, recipient1, 1000),
					mockEvent(208, recipient1, 1000),
					mockEvent(209, recipient1, 1000),
					mockEvent(210, recipient1, 1000),
				}, nil)

				return sidecar
			},
			reqTxs: txsVector(),
			expectedVE: &types.VoteExtension{
				AssetsLockedEvents: []bridgetypes.AssetsLockedEvent{
					mockEvent(201, recipient1, 1000),
					mockEvent(202, recipient1, 1000),
					mockEvent(203, recipient1, 1000),
					mockEvent(204, recipient1, 1000),
					mockEvent(205, recipient1, 1000),
					mockEvent(206, recipient1, 1000),
					mockEvent(207, recipient1, 1000),
					mockEvent(208, recipient1, 1000),
					mockEvent(209, recipient1, 1000),
					mockEvent(210, recipient1, 1000),
				},
			},
			errContains: "",
		},
		{
			name: "injected tx present - contains non-empty sequence",
			sidecarFn: func() EthereumSidecarClient {
				sidecar := newMockEthereumSidecarClient()

				// The sidecar will be called with the sequence boundaries
				// determined by the injected tx.
				sequenceStart := sdkmath.NewInt(206)
				sequenceEnd := sdkmath.NewInt(216)

				sidecar.On(
					"GetAssetsLockedEvents",
					s.ctx,
					sequenceStart,
					sequenceEnd,
				).Return([]bridgetypes.AssetsLockedEvent{
					mockEvent(206, recipient1, 1000),
				}, nil)

				return sidecar
			},
			reqTxs: append(
				[][]byte{
					marshalInjectedTx(
						types.InjectedTx{
							AssetsLockedEvents: []bridgetypes.AssetsLockedEvent{
								// The sequence tip from state is 200 but
								// the injected tx contains events that will
								// move the sequence tip to 205
								mockEvent(201, recipient1, 1000),
								mockEvent(202, recipient1, 1000),
								mockEvent(203, recipient1, 1000),
								mockEvent(204, recipient1, 1000),
								mockEvent(205, recipient1, 1000),
							},
							ExtendedCommitInfo: []byte("extendedCommitInfo"),
						},
					),
				}, txsVector("tx1", "tx2")...,
			),
			expectedVE: &types.VoteExtension{
				AssetsLockedEvents: []bridgetypes.AssetsLockedEvent{
					mockEvent(206, recipient1, 1000),
				},
			},
			errContains: "",
		},
		{
			name: "injected tx present - contains empty sequence",
			sidecarFn: func() EthereumSidecarClient {
				return newMockEthereumSidecarClient()
			},
			reqTxs: append(
				[][]byte{
					marshalInjectedTx(
						types.InjectedTx{
							AssetsLockedEvents: nil,
							ExtendedCommitInfo: []byte("extendedCommitInfo"),
						},
					),
				}, txsVector("tx1", "tx2")...,
			),
			expectedVE:  nil,
			errContains: "no AssetsLocked events in the injected tx",
		},
		{
			name: "injected tx present - empty itself",
			sidecarFn: func() EthereumSidecarClient {
				sidecar := newMockEthereumSidecarClient()

				// The injected tx contains no events so the sidecar will be
				// called with the default sequence boundaries determined by
				// the sequence tip from state.
				sidecar.On(
					"GetAssetsLockedEvents",
					s.ctx,
					defaultSequenceStart,
					defaultSequenceEnd,
				).Return([]bridgetypes.AssetsLockedEvent{
					mockEvent(201, recipient1, 1000),
				}, nil)

				return sidecar
			},
			reqTxs: append(
				[][]byte{nil}, txsVector("tx1", "tx2")...,
			),
			expectedVE: &types.VoteExtension{
				AssetsLockedEvents: []bridgetypes.AssetsLockedEvent{
					mockEvent(201, recipient1, 1000),
				},
			},
			errContains: "",
		},
		{
			name: "injected tx present - not an actual injected tx",
			sidecarFn: func() EthereumSidecarClient {
				return newMockEthereumSidecarClient()
			},
			reqTxs:      txsVector("tx1", "tx2"),
			expectedVE:  nil,
			errContains: "failed to unmarshal injected tx",
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			s.SetupTest()

			sidecar := test.sidecarFn()

			s.handler = NewVoteExtensionHandler(
				s.logger,
				sidecar,
				s.bridgeKeeper,
			)

			req := &cmtabci.RequestExtendVote{
				Height: s.requestHeight,
				Txs:    test.reqTxs,
			}

			res, err := s.handler.ExtendVoteHandler()(s.ctx, req)

			sidecar.(*mockEthereumSidecarClient).AssertExpectations(s.T())

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

	tests := []struct {
		name            string
		voteExtensionFn func() []byte
		expectedRes     *cmtabci.ResponseVerifyVoteExtension
		errContains     string
	}{
		{
			name:            "empty vote extension",
			voteExtensionFn: func() []byte { return []byte{} },
			expectedRes: &cmtabci.ResponseVerifyVoteExtension{
				Status: cmtabci.ResponseVerifyVoteExtension_ACCEPT,
			},
			errContains: "",
		},
		{
			name:            "nil vote extension",
			voteExtensionFn: func() []byte { return nil },
			expectedRes: &cmtabci.ResponseVerifyVoteExtension{
				Status: cmtabci.ResponseVerifyVoteExtension_ACCEPT,
			},
			errContains: "",
		},
		{
			name:            "non-unmarshalable vote extension",
			voteExtensionFn: func() []byte { return []byte("corrupted") },
			expectedRes: &cmtabci.ResponseVerifyVoteExtension{
				Status: cmtabci.ResponseVerifyVoteExtension_REJECT,
			},
			errContains: "failed to unmarshal vote extension",
		},
		{
			name: "empty events slice",
			voteExtensionFn: func() []byte {
				return marshalVE(types.VoteExtension{
					AssetsLockedEvents: []bridgetypes.AssetsLockedEvent{},
				})
			},
			expectedRes: &cmtabci.ResponseVerifyVoteExtension{
				Status: cmtabci.ResponseVerifyVoteExtension_ACCEPT,
			},
			errContains: "",
		},
		{
			name: "nil events slice",
			voteExtensionFn: func() []byte {
				return marshalVE(types.VoteExtension{
					AssetsLockedEvents: nil,
				})
			},
			expectedRes: &cmtabci.ResponseVerifyVoteExtension{
				Status: cmtabci.ResponseVerifyVoteExtension_ACCEPT,
			},
			errContains: "",
		},
		{
			name: "single-event slice",
			voteExtensionFn: func() []byte {
				return marshalVE(types.VoteExtension{
					AssetsLockedEvents: []bridgetypes.AssetsLockedEvent{
						mockEvent(201, recipient1, 1000),
					},
				})
			},
			expectedRes: &cmtabci.ResponseVerifyVoteExtension{
				Status: cmtabci.ResponseVerifyVoteExtension_ACCEPT,
			},
			errContains: "",
		},
		{
			name: "events slice forming improper sequence - strictly decreasing",
			voteExtensionFn: func() []byte {
				return marshalVE(types.VoteExtension{
					AssetsLockedEvents: []bridgetypes.AssetsLockedEvent{
						mockEvent(203, recipient1, 1000),
						mockEvent(202, recipient1, 1000),
						mockEvent(201, recipient1, 1000),
					},
				})
			},
			expectedRes: &cmtabci.ResponseVerifyVoteExtension{
				Status: cmtabci.ResponseVerifyVoteExtension_REJECT,
			},
			errContains: "events list is not valid",
		},
		{
			name: "events slice forming improper sequence - increasing (non-strictly)",
			voteExtensionFn: func() []byte {
				return marshalVE(types.VoteExtension{
					AssetsLockedEvents: []bridgetypes.AssetsLockedEvent{
						mockEvent(201, recipient1, 1000),
						mockEvent(201, recipient1, 1000),
						mockEvent(202, recipient1, 1000),
						mockEvent(203, recipient1, 1000),
					},
				})
			},
			expectedRes: &cmtabci.ResponseVerifyVoteExtension{
				Status: cmtabci.ResponseVerifyVoteExtension_REJECT,
			},
			errContains: "events list is not valid",
		},
		{
			name: "events slice forming improper sequence - decreasing (non-strictly)",
			voteExtensionFn: func() []byte {
				return marshalVE(types.VoteExtension{
					AssetsLockedEvents: []bridgetypes.AssetsLockedEvent{
						mockEvent(203, recipient1, 1000),
						mockEvent(203, recipient1, 1000),
						mockEvent(202, recipient1, 1000),
						mockEvent(201, recipient1, 1000),
					},
				})
			},
			expectedRes: &cmtabci.ResponseVerifyVoteExtension{
				Status: cmtabci.ResponseVerifyVoteExtension_REJECT,
			},
			errContains: "events list is not valid",
		},
		{
			name: "events slice forming improper sequence - gap",
			voteExtensionFn: func() []byte {
				return marshalVE(types.VoteExtension{
					AssetsLockedEvents: []bridgetypes.AssetsLockedEvent{
						mockEvent(201, recipient1, 1000),
						mockEvent(202, recipient1, 1000),
						mockEvent(204, recipient1, 1000),
					},
				})
			},
			expectedRes: &cmtabci.ResponseVerifyVoteExtension{
				Status: cmtabci.ResponseVerifyVoteExtension_REJECT,
			},
			errContains: "events list is not valid",
		},
		{
			name: "events slice forming improper sequence - duplicate",
			voteExtensionFn: func() []byte {
				return marshalVE(types.VoteExtension{
					AssetsLockedEvents: []bridgetypes.AssetsLockedEvent{
						mockEvent(201, recipient1, 1000),
						mockEvent(202, recipient1, 1000),
						mockEvent(203, recipient1, 1000),
						mockEvent(201, recipient1, 1000),
					},
				})
			},
			expectedRes: &cmtabci.ResponseVerifyVoteExtension{
				Status: cmtabci.ResponseVerifyVoteExtension_REJECT,
			},
			errContains: "events list is not valid",
		},
		{
			name: "events slice exceeding the limit",
			voteExtensionFn: func() []byte {
				return marshalVE(types.VoteExtension{
					AssetsLockedEvents: []bridgetypes.AssetsLockedEvent{
						mockEvent(201, recipient1, 1000),
						mockEvent(202, recipient1, 1000),
						mockEvent(203, recipient1, 1000),
						mockEvent(204, recipient1, 1000),
						mockEvent(205, recipient1, 1000),
						mockEvent(206, recipient1, 1000),
						mockEvent(207, recipient1, 1000),
						mockEvent(208, recipient1, 1000),
						mockEvent(209, recipient1, 1000),
						mockEvent(210, recipient1, 1000),
						mockEvent(211, recipient1, 1000),
					},
				})
			},
			expectedRes: &cmtabci.ResponseVerifyVoteExtension{
				Status: cmtabci.ResponseVerifyVoteExtension_REJECT,
			},
			errContains: "number of events exceeds the limit",
		},
		{
			name: "events slice within the limit",
			voteExtensionFn: func() []byte {
				return marshalVE(types.VoteExtension{
					AssetsLockedEvents: []bridgetypes.AssetsLockedEvent{
						mockEvent(201, recipient1, 1000),
						mockEvent(202, recipient1, 1000),
						mockEvent(203, recipient1, 1000),
						mockEvent(204, recipient1, 1000),
						mockEvent(205, recipient1, 1000),
						mockEvent(206, recipient1, 1000),
						mockEvent(207, recipient1, 1000),
						mockEvent(208, recipient1, 1000),
						mockEvent(209, recipient1, 1000),
						mockEvent(210, recipient1, 1000),
					},
				})
			},
			expectedRes: &cmtabci.ResponseVerifyVoteExtension{
				Status: cmtabci.ResponseVerifyVoteExtension_ACCEPT,
			},
			errContains: "",
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			s.SetupTest()

			s.handler = NewVoteExtensionHandler(
				s.logger,
				newMockEthereumSidecarClient(),
				s.bridgeKeeper,
			)

			req := &cmtabci.RequestVerifyVoteExtension{
				Hash:             []byte("hash"),
				ValidatorAddress: []byte("validatorAddress"),
				Height:           s.requestHeight,
				VoteExtension:    test.voteExtensionFn(),
			}

			res, err := s.handler.VerifyVoteExtensionHandler()(s.ctx, req)

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

type mockEthereumSidecarClient struct {
	mock.Mock
}

func newMockEthereumSidecarClient() *mockEthereumSidecarClient {
	return &mockEthereumSidecarClient{}
}

func (mesc *mockEthereumSidecarClient) GetAssetsLockedEvents(
	ctx context.Context,
	sequenceStart sdkmath.Int,
	sequenceEnd sdkmath.Int,
) ([]bridgetypes.AssetsLockedEvent, error) {
	args := mesc.Called(ctx, sequenceStart, sequenceEnd)

	if res := args.Get(0); res != nil {
		return res.([]bridgetypes.AssetsLockedEvent), args.Error(1)
	}

	return nil, args.Error(1)
}

type mockBridgeKeeper struct {
	mock.Mock
}

func newMockBridgeKeeper() *mockBridgeKeeper {
	return &mockBridgeKeeper{}
}

func (mbk *mockBridgeKeeper) GetAssetsLockedSequenceTip(ctx sdk.Context) sdkmath.Int {
	args := mbk.Called(ctx)
	return args.Get(0).(sdkmath.Int)
}

func (mbk *mockBridgeKeeper) AcceptAssetsLocked(
	ctx sdk.Context,
	events bridgetypes.AssetsLockedEvents,
) error {
	args := mbk.Called(ctx, events)
	return args.Error(0)
}
