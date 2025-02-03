package abci

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/mezo-org/mezod/cmd/config"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	cmtabci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/proto/tendermint/crypto"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	servermock "github.com/cosmos/cosmos-sdk/server/mock"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/mezo-org/mezod/x/bridge/abci/types"
	bridgetypes "github.com/mezo-org/mezod/x/bridge/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

const (
	recipient1 = "mezo12wsc0qgyfwwfj3wrlpgm9q3lmndl2m4qmm34dp"
	recipient2 = "mezo1xqurxvvh8z2xpj6wltq0tajxm47xnv7q6rtvja"
	recipient3 = "mezo1jcurf087xx9eqsnmr8lszralupcfln2vjz8ucg"
	recipient4 = "mezo1j0ghx6d9kmerxhgn5ahr2nahs6yfulea4te22c"
	recipient5 = "mezo120mkxfvkx2t72quddqh92md2dp7csq4wgqux06"
	recipient6 = "mezo1dmr6mhh352vh9wa34xs0qxtr8thkqu39pw6x2p"
	//nolint:gosec
	token = "0x517f2982701695D4E52f1ECFBEf3ba31Df470161"
)

func TestProposalHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(ProposalHandlerTestSuite))
}

type ProposalHandlerTestSuite struct {
	suite.Suite

	logger        log.Logger
	ctx           sdk.Context
	requestHeight int64
	bridgeKeeper  *mockBridgeKeeper
	handler       *ProposalHandler
}

func (s *ProposalHandlerTestSuite) SetupTest() {
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

func (s *ProposalHandlerTestSuite) TestPrepareProposal() {
	tests := []struct {
		name                    string
		voteExtensionsValidator *mockVoteExtensionsValidator
		assetsLockedExtractorFn func(cmtabci.ExtendedCommitInfo) *mockAssetsLockedExtractor
		reqTxs                  [][]byte
		reqVoteExtensionsFn     func() []cmtabci.ExtendedVoteInfo
		expectedInjectedTxFn    func([]byte) *types.InjectedTx
		expectedChainTxs        [][]byte // regular chain txs being part of the proposal
		errContains             string
	}{
		{
			name:                    "invalid vote extensions",
			voteExtensionsValidator: newMockVoteExtensionsValidator(fmt.Errorf("invalid vote extensions")),
			assetsLockedExtractorFn: func(_ cmtabci.ExtendedCommitInfo) *mockAssetsLockedExtractor {
				return newMockAssetsLockedExtractor()
			},
			reqTxs:               txsVector(),
			reqVoteExtensionsFn:  func() []cmtabci.ExtendedVoteInfo { return nil },
			expectedInjectedTxFn: func(_ []byte) *types.InjectedTx { return nil },
			expectedChainTxs:     nil,
			errContains:          "failed to validate vote extensions",
		},
		{
			name:                    "empty canonical sequence",
			voteExtensionsValidator: newMockVoteExtensionsValidator(nil),
			assetsLockedExtractorFn: func(extCommitInfo cmtabci.ExtendedCommitInfo) *mockAssetsLockedExtractor {
				extractor := newMockAssetsLockedExtractor()

				var voteExtension types.VoteExtension
				err := voteExtension.Unmarshal(extCommitInfo.Votes[0].VoteExtension)
				s.Require().NoError(err)

				// Return first vote's AssetsLocked events as the canonical sequence.
				extractor.On(
					"CanonicalEvents",
					s.ctx,
					extCommitInfo,
					s.requestHeight,
				).Return(voteExtension.AssetsLockedEvents, nil)

				return extractor
			},
			reqTxs: txsVector("tx1", "tx2"),
			reqVoteExtensionsFn: func() []cmtabci.ExtendedVoteInfo {
				return []cmtabci.ExtendedVoteInfo{
					// The mock extractor returns the canonical sequence based
					// on the first vote so use a single vote to keep the test simple.
					mockVoteExtension("val1Bridge", 100, tmproto.BlockIDFlagCommit),
				}
			},
			expectedInjectedTxFn: func(_ []byte) *types.InjectedTx { return nil },
			expectedChainTxs:     txsVector("tx1", "tx2"),
			errContains:          "",
		},
		{
			name:                    "canonical sequence not starting directly after the current tip",
			voteExtensionsValidator: newMockVoteExtensionsValidator(nil),
			assetsLockedExtractorFn: func(extCommitInfo cmtabci.ExtendedCommitInfo) *mockAssetsLockedExtractor {
				extractor := newMockAssetsLockedExtractor()

				var voteExtension types.VoteExtension
				err := voteExtension.Unmarshal(extCommitInfo.Votes[0].VoteExtension)
				s.Require().NoError(err)

				// Return first vote's AssetsLocked events as the canonical sequence.
				extractor.On(
					"CanonicalEvents",
					s.ctx,
					extCommitInfo,
					s.requestHeight,
				).Return(voteExtension.AssetsLockedEvents, nil)

				return extractor
			},
			reqTxs: txsVector("tx1", "tx2"),
			reqVoteExtensionsFn: func() []cmtabci.ExtendedVoteInfo {
				events := []bridgetypes.AssetsLockedEvent{
					// First event starts at 202 while the tip is 200.
					mockEvent(202, recipient1, 100, token),
					mockEvent(203, recipient2, 200, token),
					mockEvent(204, recipient3, 300, token),
				}

				return []cmtabci.ExtendedVoteInfo{
					// The mock extractor returns the canonical sequence based
					// on the first vote so use a single vote to keep the test simple.
					mockVoteExtension("val1Bridge", 100, tmproto.BlockIDFlagCommit, events...),
				}
			},
			expectedInjectedTxFn: func(_ []byte) *types.InjectedTx { return nil },
			expectedChainTxs:     txsVector("tx1", "tx2"),
			errContains:          "",
		},
		{
			name:                    "proper canonical sequence",
			voteExtensionsValidator: newMockVoteExtensionsValidator(nil),
			assetsLockedExtractorFn: func(extCommitInfo cmtabci.ExtendedCommitInfo) *mockAssetsLockedExtractor {
				extractor := newMockAssetsLockedExtractor()

				var voteExtension types.VoteExtension
				err := voteExtension.Unmarshal(extCommitInfo.Votes[0].VoteExtension)
				s.Require().NoError(err)

				// Return first vote's AssetsLocked events as the canonical sequence.
				extractor.On(
					"CanonicalEvents",
					s.ctx,
					extCommitInfo,
					s.requestHeight,
				).Return(voteExtension.AssetsLockedEvents, nil)

				return extractor
			},
			reqTxs: txsVector("tx1", "tx2"),
			reqVoteExtensionsFn: func() []cmtabci.ExtendedVoteInfo {
				events := []bridgetypes.AssetsLockedEvent{
					mockEvent(201, recipient1, 100, token),
					mockEvent(202, recipient2, 200, token),
					mockEvent(203, recipient3, 300, token),
				}

				return []cmtabci.ExtendedVoteInfo{
					// The mock extractor returns the canonical sequence based
					// on the first vote so use a single vote to keep the test simple.
					mockVoteExtension("val1Bridge", 100, tmproto.BlockIDFlagCommit, events...),
				}
			},
			expectedInjectedTxFn: func(extCommitInfo []byte) *types.InjectedTx {
				return &types.InjectedTx{
					AssetsLockedEvents: []bridgetypes.AssetsLockedEvent{
						mockEvent(201, recipient1, 100, token),
						mockEvent(202, recipient2, 200, token),
						mockEvent(203, recipient3, 300, token),
					},
					ExtendedCommitInfo: extCommitInfo,
				}
			},
			expectedChainTxs: txsVector("tx1", "tx2"),
			errContains:      "",
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			s.SetupTest()

			valStore := newMockValidatorStore()
			voteExtensionsValidator := test.voteExtensionsValidator

			s.handler = NewProposalHandler(
				s.logger,
				valStore,
				s.bridgeKeeper,
				// No need for the vote extension decomposer in this test.
				// The decomposer is used to construct the default assetsLockedExtractor
				// but, we are overriding it with a test one.
				nil,
				voteExtensionsValidator.call,
			)

			extendedCommitInfo := cmtabci.ExtendedCommitInfo{
				Round: 1, // Just an arbitrary value.
				Votes: test.reqVoteExtensionsFn(),
			}

			// Override the default assetsLockedExtractor with a test one.
			extractor := test.assetsLockedExtractorFn(extendedCommitInfo)
			s.handler.assetsLockedExtractor = extractor

			req := &cmtabci.RequestPrepareProposal{
				// Fill only the fields that are relevant for the test.
				Height:          s.requestHeight,
				Txs:             test.reqTxs,
				LocalLastCommit: extendedCommitInfo,
			}

			res, err := s.handler.PrepareProposalHandler()(s.ctx, req)

			voteExtensionsValidator.AssertCalled(
				s.T(),
				"call",
				s.ctx,
				valStore,
				s.requestHeight,
				s.ctx.ChainID(),
				extendedCommitInfo,
			)

			extractor.AssertExpectations(s.T())

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

			extendedCommitInfoBytes, err := extendedCommitInfo.Marshal()
			s.Require().NoError(err)

			s.Require().Equal(
				test.expectedInjectedTxFn(extendedCommitInfoBytes),
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
	marshalExtCommitInfo := func(extCommitInfo cmtabci.ExtendedCommitInfo) []byte {
		extCommitInfoBytes, err := extCommitInfo.Marshal()
		s.Require().NoError(err)
		return extCommitInfoBytes
	}

	tests := []struct {
		name                    string
		voteExtensionsValidator *mockVoteExtensionsValidator
		assetsLockedExtractorFn func(cmtabci.ExtendedCommitInfo) *mockAssetsLockedExtractor
		reqTxsFn                func() [][]byte
		expectedRes             *cmtabci.ResponseProcessProposal
		errContains             string
	}{
		{
			name:                    "no txs in the request",
			voteExtensionsValidator: newMockVoteExtensionsValidator(nil),
			assetsLockedExtractorFn: func(_ cmtabci.ExtendedCommitInfo) *mockAssetsLockedExtractor {
				return newMockAssetsLockedExtractor()
			},
			reqTxsFn: func() [][]byte { return txsVector() },
			expectedRes: &cmtabci.ResponseProcessProposal{
				Status: cmtabci.ResponseProcessProposal_REJECT,
			},
			errContains: "empty transaction vector in the proposal",
		},
		{
			name:                    "empty injected tx",
			voteExtensionsValidator: newMockVoteExtensionsValidator(nil),
			assetsLockedExtractorFn: func(_ cmtabci.ExtendedCommitInfo) *mockAssetsLockedExtractor {
				return newMockAssetsLockedExtractor()
			},
			reqTxsFn: func() [][]byte { return txsVector("") },
			expectedRes: &cmtabci.ResponseProcessProposal{
				Status: cmtabci.ResponseProcessProposal_ACCEPT,
			},
			errContains: "",
		},
		{
			name:                    "non-unmarshalable injected tx",
			voteExtensionsValidator: newMockVoteExtensionsValidator(nil),
			assetsLockedExtractorFn: func(_ cmtabci.ExtendedCommitInfo) *mockAssetsLockedExtractor {
				return newMockAssetsLockedExtractor()
			},
			reqTxsFn: func() [][]byte { return txsVector("corrupted") },
			expectedRes: &cmtabci.ResponseProcessProposal{
				Status: cmtabci.ResponseProcessProposal_REJECT,
			},
			errContains: "failed to unmarshal injected tx",
		},
		{
			name:                    "injected tx with empty sequence",
			voteExtensionsValidator: newMockVoteExtensionsValidator(nil),
			assetsLockedExtractorFn: func(_ cmtabci.ExtendedCommitInfo) *mockAssetsLockedExtractor {
				return newMockAssetsLockedExtractor()
			},
			reqTxsFn: func() [][]byte {
				return append(
					[][]byte{
						marshalInjectedTx(
							types.InjectedTx{
								AssetsLockedEvents: nil,
								// Just an arbitrary value to produce a non-empty
								// marshaled injected tx.
								ExtendedCommitInfo: []byte("extendedCommitInfo"),
							},
						),
					},
					txsVector("tx1", "tx2")...,
				)
			},
			expectedRes: &cmtabci.ResponseProcessProposal{
				Status: cmtabci.ResponseProcessProposal_REJECT,
			},
			errContains: "injected tx does not contain AssetsLocked events",
		},
		{
			name:                    "non-unmarshalable extended commit info",
			voteExtensionsValidator: newMockVoteExtensionsValidator(nil),
			assetsLockedExtractorFn: func(_ cmtabci.ExtendedCommitInfo) *mockAssetsLockedExtractor {
				return newMockAssetsLockedExtractor()
			},
			reqTxsFn: func() [][]byte {
				return append(
					[][]byte{
						marshalInjectedTx(
							types.InjectedTx{
								AssetsLockedEvents: []bridgetypes.AssetsLockedEvent{
									mockEvent(201, recipient1, 100, token),
								},
								ExtendedCommitInfo: []byte("corrupted"),
							},
						),
					},
					txsVector("tx1", "tx2")...,
				)
			},
			expectedRes: &cmtabci.ResponseProcessProposal{
				Status: cmtabci.ResponseProcessProposal_REJECT,
			},
			errContains: "failed to unmarshal commit info from injected tx",
		},
		{
			name:                    "invalid vote extensions",
			voteExtensionsValidator: newMockVoteExtensionsValidator(fmt.Errorf("invalid vote extensions")),
			assetsLockedExtractorFn: func(_ cmtabci.ExtendedCommitInfo) *mockAssetsLockedExtractor {
				return newMockAssetsLockedExtractor()
			},
			reqTxsFn: func() [][]byte {
				return append(
					[][]byte{
						marshalInjectedTx(
							types.InjectedTx{
								AssetsLockedEvents: []bridgetypes.AssetsLockedEvent{
									mockEvent(201, recipient1, 100, token),
								},
								ExtendedCommitInfo: marshalExtCommitInfo(
									cmtabci.ExtendedCommitInfo{
										Round: 1,
										Votes: []cmtabci.ExtendedVoteInfo{},
									},
								),
							},
						),
					},
					txsVector("tx1", "tx2")...,
				)
			},
			expectedRes: &cmtabci.ResponseProcessProposal{
				Status: cmtabci.ResponseProcessProposal_REJECT,
			},
			errContains: "failed to validate vote extensions from injected tx",
		},
		{
			name:                    "re-create canonical sequence error",
			voteExtensionsValidator: newMockVoteExtensionsValidator(nil),
			assetsLockedExtractorFn: func(extCommitInfo cmtabci.ExtendedCommitInfo) *mockAssetsLockedExtractor {
				extractor := newMockAssetsLockedExtractor()

				extractor.On(
					"CanonicalEvents",
					s.ctx,
					extCommitInfo,
					s.requestHeight,
				).Return(nil, fmt.Errorf("unexpected error"))

				return extractor
			},
			reqTxsFn: func() [][]byte {
				return append(
					[][]byte{
						marshalInjectedTx(
							types.InjectedTx{
								AssetsLockedEvents: []bridgetypes.AssetsLockedEvent{
									mockEvent(201, recipient1, 100, token),
								},
								ExtendedCommitInfo: marshalExtCommitInfo(
									cmtabci.ExtendedCommitInfo{
										Round: 1,
										Votes: []cmtabci.ExtendedVoteInfo{},
									},
								),
							},
						),
					},
					txsVector("tx1", "tx2")...,
				)
			},
			expectedRes: &cmtabci.ResponseProcessProposal{
				Status: cmtabci.ResponseProcessProposal_REJECT,
			},
			errContains: "failed to recreate canonical AssetsLocked events",
		},
		{
			name:                    "re-created canonical sequence not matching injected tx sequence",
			voteExtensionsValidator: newMockVoteExtensionsValidator(nil),
			assetsLockedExtractorFn: func(extCommitInfo cmtabci.ExtendedCommitInfo) *mockAssetsLockedExtractor {
				extractor := newMockAssetsLockedExtractor()

				var voteExtension types.VoteExtension
				err := voteExtension.Unmarshal(extCommitInfo.Votes[0].VoteExtension)
				s.Require().NoError(err)

				// Return first vote's AssetsLocked events as the canonical sequence.
				extractor.On(
					"CanonicalEvents",
					s.ctx,
					extCommitInfo,
					s.requestHeight,
				).Return(voteExtension.AssetsLockedEvents, nil)

				return extractor
			},
			reqTxsFn: func() [][]byte {
				return append(
					[][]byte{
						marshalInjectedTx(
							types.InjectedTx{
								AssetsLockedEvents: []bridgetypes.AssetsLockedEvent{
									mockEvent(201, recipient1, 100, token),
								},
								ExtendedCommitInfo: marshalExtCommitInfo(
									cmtabci.ExtendedCommitInfo{
										Round: 1,
										Votes: []cmtabci.ExtendedVoteInfo{
											// The mock extractor returns the canonical sequence based
											// on the first vote so use a single vote to keep the test simple.
											mockVoteExtension(
												"val1Bridge",
												100,
												tmproto.BlockIDFlagCommit,
												mockEvent(
													201,
													recipient1,
													1000, // Different amount.
													token,
												),
											),
										},
									},
								),
							},
						),
					},
					txsVector("tx1", "tx2")...,
				)
			},
			expectedRes: &cmtabci.ResponseProcessProposal{
				Status: cmtabci.ResponseProcessProposal_REJECT,
			},
			errContains: "recreated canonical AssetsLocked events do not match events from injected tx",
		},
		{
			name:                    "injected tx sequence not starting directly after the current tip",
			voteExtensionsValidator: newMockVoteExtensionsValidator(nil),
			assetsLockedExtractorFn: func(extCommitInfo cmtabci.ExtendedCommitInfo) *mockAssetsLockedExtractor {
				extractor := newMockAssetsLockedExtractor()

				var voteExtension types.VoteExtension
				err := voteExtension.Unmarshal(extCommitInfo.Votes[0].VoteExtension)
				s.Require().NoError(err)

				// Return first vote's AssetsLocked events as the canonical sequence.
				extractor.On(
					"CanonicalEvents",
					s.ctx,
					extCommitInfo,
					s.requestHeight,
				).Return(voteExtension.AssetsLockedEvents, nil)

				return extractor
			},
			reqTxsFn: func() [][]byte {
				// The injected tx sequence starts at 202 while the tip is 200.
				// It should be 201 to be valid.
				events := []bridgetypes.AssetsLockedEvent{
					mockEvent(202, recipient1, 100, token),
					mockEvent(203, recipient2, 200, token),
					mockEvent(204, recipient3, 300, token),
				}

				return append(
					[][]byte{
						marshalInjectedTx(
							types.InjectedTx{
								AssetsLockedEvents: events,
								ExtendedCommitInfo: marshalExtCommitInfo(
									cmtabci.ExtendedCommitInfo{
										Round: 1,
										Votes: []cmtabci.ExtendedVoteInfo{
											// The mock extractor returns the canonical sequence based
											// on the first vote so use a single vote to keep the test simple.
											mockVoteExtension(
												"val1Bridge",
												102,
												tmproto.BlockIDFlagCommit,
												events...,
											),
										},
									},
								),
							},
						),
					},
					txsVector("tx1", "tx2")...,
				)
			},
			expectedRes: &cmtabci.ResponseProcessProposal{
				Status: cmtabci.ResponseProcessProposal_REJECT,
			},
			errContains: "AssetsLocked events from injected tx do not start after the current sequence tip",
		},
		{
			name:                    "proper injected tx sequence",
			voteExtensionsValidator: newMockVoteExtensionsValidator(nil),
			assetsLockedExtractorFn: func(extCommitInfo cmtabci.ExtendedCommitInfo) *mockAssetsLockedExtractor {
				extractor := newMockAssetsLockedExtractor()

				var voteExtension types.VoteExtension
				err := voteExtension.Unmarshal(extCommitInfo.Votes[0].VoteExtension)
				s.Require().NoError(err)

				// Return first vote's AssetsLocked events as the canonical sequence.
				extractor.On(
					"CanonicalEvents",
					s.ctx,
					extCommitInfo,
					s.requestHeight,
				).Return(voteExtension.AssetsLockedEvents, nil)

				return extractor
			},
			reqTxsFn: func() [][]byte {
				events := []bridgetypes.AssetsLockedEvent{
					mockEvent(201, recipient1, 100, token),
					mockEvent(202, recipient2, 200, token),
					mockEvent(203, recipient3, 300, token),
				}

				return append(
					[][]byte{
						marshalInjectedTx(
							types.InjectedTx{
								AssetsLockedEvents: events,
								ExtendedCommitInfo: marshalExtCommitInfo(
									cmtabci.ExtendedCommitInfo{
										Round: 1,
										Votes: []cmtabci.ExtendedVoteInfo{
											// The mock extractor returns the canonical sequence based
											// on the first vote so use a single vote to keep the test simple.
											mockVoteExtension(
												"val1Bridge",
												102,
												tmproto.BlockIDFlagCommit,
												events...,
											),
										},
									},
								),
							},
						),
					},
					txsVector("tx1", "tx2")...,
				)
			},
			expectedRes: &cmtabci.ResponseProcessProposal{
				Status: cmtabci.ResponseProcessProposal_ACCEPT,
			},
			errContains: "",
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			s.SetupTest()

			valStore := newMockValidatorStore()
			voteExtensionsValidator := test.voteExtensionsValidator
			reqTxs := test.reqTxsFn()

			extCommitInfo, extCommitInfoOk := cmtabci.ExtendedCommitInfo{}, false
			if len(reqTxs) > 0 && len(reqTxs[0]) > 0 {
				injectedTx := types.InjectedTx{}
				if unmErr1 := injectedTx.Unmarshal(reqTxs[0]); unmErr1 == nil {
					if unmErr2 := extCommitInfo.Unmarshal(
						injectedTx.ExtendedCommitInfo,
					); unmErr2 == nil {
						extCommitInfoOk = true
					}
				}
			}

			s.handler = NewProposalHandler(
				s.logger,
				valStore,
				s.bridgeKeeper,
				// No need for the vote extension decomposer in this test.
				// The decomposer is used to construct the default assetsLockedExtractor
				// but, we are overriding it with a test one.
				nil,
				voteExtensionsValidator.call,
			)

			// Override the default assetsLockedExtractor with a test one.
			extractor := test.assetsLockedExtractorFn(extCommitInfo)
			s.handler.assetsLockedExtractor = extractor

			req := &cmtabci.RequestProcessProposal{
				// Fill only the fields that are relevant for the test.
				Height:          s.requestHeight,
				Txs:             reqTxs,
				ProposerAddress: []byte("proposerAddress"),
			}

			res, err := s.handler.ProcessProposalHandler()(s.ctx, req)

			// Make sure VE validation occurred if and only if there was
			// an unmarshalable injected tx holding unmarshalable extended
			// commit info.
			if extCommitInfoOk {
				voteExtensionsValidator.AssertCalled(
					s.T(),
					"call",
					s.ctx,
					valStore,
					s.requestHeight,
					s.ctx.ChainID(),
					extCommitInfo,
				)
			}

			extractor.AssertExpectations(s.T())

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

func TestAssetsLockedExtractorTestSuite(t *testing.T) {
	suite.Run(t, new(AssetsLockedExtractorTestSuite))
}

type AssetsLockedExtractorTestSuite struct {
	suite.Suite

	logger    log.Logger
	ctx       sdk.Context
	height    int64
	extractor *assetsLockedExtractor
}

func (s *AssetsLockedExtractorTestSuite) SetupTest() {
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

	s.height = 100
}

func (s *AssetsLockedExtractorTestSuite) TestCanonicalEvents() {
	tests := []struct {
		name                      string
		valStore                  *mockValidatorStore
		voteExtensionDecomposerFn func() *mockVoteExtensionDecomposer
		reqVoteExtensionsFn       func() []cmtabci.ExtendedVoteInfo
		expectedEvents            bridgetypes.AssetsLockedEvents
		errContains               string
	}{
		{
			name:     "proper canonical sequence - unanimous voting",
			valStore: newMockValidatorStore("val1Bridge", "val2Bridge", "val3Bridge", "val4Bridge"),
			voteExtensionDecomposerFn: func() *mockVoteExtensionDecomposer {
				return newMockVoteExtensionDecomposer().withReturnInputMode()
			},
			reqVoteExtensionsFn: func() []cmtabci.ExtendedVoteInfo {
				events := []bridgetypes.AssetsLockedEvent{
					mockEvent(201, recipient1, 100, token),
					mockEvent(202, recipient2, 200, token),
					mockEvent(203, recipient3, 300, token),
				}

				return []cmtabci.ExtendedVoteInfo{
					mockVoteExtension("val1Bridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val2Bridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val3Bridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val4Bridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val5NonBridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val6NonBridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val7NonBridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val8NonBridge", 100, tmproto.BlockIDFlagCommit, events...),
				}
			},
			expectedEvents: bridgetypes.AssetsLockedEvents{
				mockEvent(201, recipient1, 100, token),
				mockEvent(202, recipient2, 200, token),
				mockEvent(203, recipient3, 300, token),
			},
			errContains: "",
		},
		{
			name:     "proper canonical sequence - some votes on different sequence",
			valStore: newMockValidatorStore("val1Bridge", "val2Bridge", "val3Bridge", "val4Bridge"),
			voteExtensionDecomposerFn: func() *mockVoteExtensionDecomposer {
				return newMockVoteExtensionDecomposer().withReturnInputMode()
			},
			reqVoteExtensionsFn: func() []cmtabci.ExtendedVoteInfo {
				// Events the super-majority of validators are voting on.
				eventsSm := []bridgetypes.AssetsLockedEvent{
					mockEvent(201, recipient1, 100, token),
					mockEvent(202, recipient2, 200, token),
					mockEvent(203, recipient3, 300, token),
				}

				// Events the minority of validators are voting on. They
				// differ in the sequence numbers.
				eventsNonSm := []bridgetypes.AssetsLockedEvent{
					mockEvent(204, recipient1, 100, token),
					mockEvent(205, recipient2, 200, token),
					mockEvent(206, recipient3, 300, token),
				}

				return []cmtabci.ExtendedVoteInfo{
					mockVoteExtension("val1Bridge", 100, tmproto.BlockIDFlagCommit, eventsSm...),
					mockVoteExtension("val2Bridge", 100, tmproto.BlockIDFlagCommit, eventsSm...),
					mockVoteExtension("val3Bridge", 100, tmproto.BlockIDFlagCommit, eventsSm...),
					mockVoteExtension("val4Bridge", 100, tmproto.BlockIDFlagCommit, eventsNonSm...),
					mockVoteExtension("val5NonBridge", 100, tmproto.BlockIDFlagCommit, eventsSm...),
					mockVoteExtension("val6NonBridge", 100, tmproto.BlockIDFlagCommit, eventsSm...),
					mockVoteExtension("val7NonBridge", 100, tmproto.BlockIDFlagCommit, eventsSm...),
					mockVoteExtension("val8NonBridge", 100, tmproto.BlockIDFlagCommit, eventsNonSm...),
				}
			},
			expectedEvents: bridgetypes.AssetsLockedEvents{
				mockEvent(201, recipient1, 100, token),
				mockEvent(202, recipient2, 200, token),
				mockEvent(203, recipient3, 300, token),
			},

			errContains: "",
		},
		{
			name:     "proper canonical sequence - some votes on different recipients",
			valStore: newMockValidatorStore("val1Bridge", "val2Bridge", "val3Bridge", "val4Bridge"),
			voteExtensionDecomposerFn: func() *mockVoteExtensionDecomposer {
				return newMockVoteExtensionDecomposer().withReturnInputMode()
			},
			reqVoteExtensionsFn: func() []cmtabci.ExtendedVoteInfo {
				// Events the super-majority of validators are voting on.
				eventsSm := []bridgetypes.AssetsLockedEvent{
					mockEvent(201, recipient1, 100, token),
					mockEvent(202, recipient2, 200, token),
					mockEvent(203, recipient3, 300, token),
				}

				// Events the minority of validators are voting on. They
				// differ in the recipient address.
				eventsNonSm := []bridgetypes.AssetsLockedEvent{
					mockEvent(201, recipient4, 100, token),
					mockEvent(202, recipient5, 200, token),
					mockEvent(203, recipient6, 300, token),
				}

				return []cmtabci.ExtendedVoteInfo{
					mockVoteExtension("val1Bridge", 100, tmproto.BlockIDFlagCommit, eventsSm...),
					mockVoteExtension("val2Bridge", 100, tmproto.BlockIDFlagCommit, eventsSm...),
					mockVoteExtension("val3Bridge", 100, tmproto.BlockIDFlagCommit, eventsSm...),
					mockVoteExtension("val4Bridge", 100, tmproto.BlockIDFlagCommit, eventsNonSm...),
					mockVoteExtension("val5NonBridge", 100, tmproto.BlockIDFlagCommit, eventsSm...),
					mockVoteExtension("val6NonBridge", 100, tmproto.BlockIDFlagCommit, eventsSm...),
					mockVoteExtension("val7NonBridge", 100, tmproto.BlockIDFlagCommit, eventsSm...),
					mockVoteExtension("val8NonBridge", 100, tmproto.BlockIDFlagCommit, eventsNonSm...),
				}
			},
			expectedEvents: bridgetypes.AssetsLockedEvents{
				mockEvent(201, recipient1, 100, token),
				mockEvent(202, recipient2, 200, token),
				mockEvent(203, recipient3, 300, token),
			},
			errContains: "",
		},
		{
			name:     "proper canonical sequence - some votes on different amounts",
			valStore: newMockValidatorStore("val1Bridge", "val2Bridge", "val3Bridge", "val4Bridge"),
			voteExtensionDecomposerFn: func() *mockVoteExtensionDecomposer {
				return newMockVoteExtensionDecomposer().withReturnInputMode()
			},
			reqVoteExtensionsFn: func() []cmtabci.ExtendedVoteInfo {
				// Events the super-majority of validators are voting on.
				eventsSm := []bridgetypes.AssetsLockedEvent{
					mockEvent(201, recipient1, 100, token),
					mockEvent(202, recipient2, 200, token),
					mockEvent(203, recipient3, 300, token),
				}

				// Events the minority of validators are voting on. They
				// differ in the amount of locked assets.
				eventsNonSm := []bridgetypes.AssetsLockedEvent{
					mockEvent(201, recipient1, 1000, token),
					mockEvent(202, recipient2, 2000, token),
					mockEvent(203, recipient3, 3000, token),
				}

				return []cmtabci.ExtendedVoteInfo{
					mockVoteExtension("val1Bridge", 100, tmproto.BlockIDFlagCommit, eventsSm...),
					mockVoteExtension("val2Bridge", 100, tmproto.BlockIDFlagCommit, eventsSm...),
					mockVoteExtension("val3Bridge", 100, tmproto.BlockIDFlagCommit, eventsSm...),
					mockVoteExtension("val4Bridge", 100, tmproto.BlockIDFlagCommit, eventsNonSm...),
					mockVoteExtension("val5NonBridge", 100, tmproto.BlockIDFlagCommit, eventsSm...),
					mockVoteExtension("val6NonBridge", 100, tmproto.BlockIDFlagCommit, eventsSm...),
					mockVoteExtension("val7NonBridge", 100, tmproto.BlockIDFlagCommit, eventsSm...),
					mockVoteExtension("val8NonBridge", 100, tmproto.BlockIDFlagCommit, eventsNonSm...),
				}
			},
			expectedEvents: bridgetypes.AssetsLockedEvents{
				mockEvent(201, recipient1, 100, token),
				mockEvent(202, recipient2, 200, token),
				mockEvent(203, recipient3, 300, token),
			},
			errContains: "",
		},
		{
			name:     "proper canonical sequence - some votes on different tokens",
			valStore: newMockValidatorStore("val1Bridge", "val2Bridge", "val3Bridge", "val4Bridge"),
			voteExtensionDecomposerFn: func() *mockVoteExtensionDecomposer {
				return newMockVoteExtensionDecomposer().withReturnInputMode()
			},
			reqVoteExtensionsFn: func() []cmtabci.ExtendedVoteInfo {
				// Events the super-majority of validators are voting on.
				eventsSm := []bridgetypes.AssetsLockedEvent{
					mockEvent(201, recipient1, 100, token),
					mockEvent(202, recipient2, 200, token),
					mockEvent(203, recipient3, 300, token),
				}

				//nolint:gosec
				otherToken := "0x9395499f006821Fc5E22979fafEecD9f5C70E173"

				// Events the minority of validators are voting on. They
				// differ in the token of locked assets.
				eventsNonSm := []bridgetypes.AssetsLockedEvent{
					mockEvent(201, recipient1, 100, otherToken),
					mockEvent(202, recipient2, 200, otherToken),
					mockEvent(203, recipient3, 300, otherToken),
				}

				return []cmtabci.ExtendedVoteInfo{
					mockVoteExtension("val1Bridge", 100, tmproto.BlockIDFlagCommit, eventsSm...),
					mockVoteExtension("val2Bridge", 100, tmproto.BlockIDFlagCommit, eventsSm...),
					mockVoteExtension("val3Bridge", 100, tmproto.BlockIDFlagCommit, eventsSm...),
					mockVoteExtension("val4Bridge", 100, tmproto.BlockIDFlagCommit, eventsNonSm...),
					mockVoteExtension("val5NonBridge", 100, tmproto.BlockIDFlagCommit, eventsSm...),
					mockVoteExtension("val6NonBridge", 100, tmproto.BlockIDFlagCommit, eventsSm...),
					mockVoteExtension("val7NonBridge", 100, tmproto.BlockIDFlagCommit, eventsSm...),
					mockVoteExtension("val8NonBridge", 100, tmproto.BlockIDFlagCommit, eventsNonSm...),
				}
			},
			expectedEvents: bridgetypes.AssetsLockedEvents{
				mockEvent(201, recipient1, 100, token),
				mockEvent(202, recipient2, 200, token),
				mockEvent(203, recipient3, 300, token),
			},
			errContains: "",
		},
		{
			name:     "proper canonical sequence - some votes on sequence outliers",
			valStore: newMockValidatorStore("val1Bridge", "val2Bridge", "val3Bridge", "val4Bridge"),
			voteExtensionDecomposerFn: func() *mockVoteExtensionDecomposer {
				return newMockVoteExtensionDecomposer().withReturnInputMode()
			},
			reqVoteExtensionsFn: func() []cmtabci.ExtendedVoteInfo {
				// Events the super-majority of validators are voting on.
				eventsSm := []bridgetypes.AssetsLockedEvent{
					mockEvent(201, recipient1, 100, token),
					mockEvent(202, recipient2, 200, token),
					mockEvent(203, recipient3, 300, token),
				}

				// Outliers are events that some of the validators are voting on
				// but, they are not supported by the super-majority.
				outlier1 := mockEvent(200, recipient4, 90, token)
				outlier2 := mockEvent(204, recipient5, 400, token)

				return []cmtabci.ExtendedVoteInfo{
					mockVoteExtension("val1Bridge", 100, tmproto.BlockIDFlagCommit, append([]bridgetypes.AssetsLockedEvent{outlier1}, eventsSm...)...),
					mockVoteExtension("val2Bridge", 100, tmproto.BlockIDFlagCommit, append([]bridgetypes.AssetsLockedEvent{outlier1}, eventsSm...)...),
					mockVoteExtension("val3Bridge", 100, tmproto.BlockIDFlagCommit, append(eventsSm, outlier2)...),
					mockVoteExtension("val4Bridge", 100, tmproto.BlockIDFlagCommit, append(eventsSm, outlier2)...),
					mockVoteExtension("val5NonBridge", 100, tmproto.BlockIDFlagCommit, append([]bridgetypes.AssetsLockedEvent{outlier1}, eventsSm...)...),
					mockVoteExtension("val6NonBridge", 100, tmproto.BlockIDFlagCommit, append([]bridgetypes.AssetsLockedEvent{outlier1}, eventsSm...)...),
					mockVoteExtension("val7NonBridge", 100, tmproto.BlockIDFlagCommit, append(eventsSm, outlier2)...),
					mockVoteExtension("val8NonBridge", 100, tmproto.BlockIDFlagCommit, append(eventsSm, outlier2)...),
				}
			},
			expectedEvents: bridgetypes.AssetsLockedEvents{
				mockEvent(201, recipient1, 100, token),
				mockEvent(202, recipient2, 200, token),
				mockEvent(203, recipient3, 300, token),
			},
			errContains: "",
		},
		{
			name:     "no canonical sequence - split brain",
			valStore: newMockValidatorStore("val1Bridge", "val2Bridge", "val3Bridge", "val4Bridge"),
			voteExtensionDecomposerFn: func() *mockVoteExtensionDecomposer {
				return newMockVoteExtensionDecomposer().withReturnInputMode()
			},
			reqVoteExtensionsFn: func() []cmtabci.ExtendedVoteInfo {
				// Set of events the first half of validators are voting on.
				events1 := []bridgetypes.AssetsLockedEvent{
					mockEvent(201, recipient1, 100, token),
					mockEvent(202, recipient2, 200, token),
					mockEvent(203, recipient3, 300, token),
				}

				// Set of events the second half of validators are voting on.
				events2 := []bridgetypes.AssetsLockedEvent{
					mockEvent(201, recipient4, 1000, token),
					mockEvent(202, recipient5, 2000, token),
					mockEvent(203, recipient6, 3000, token),
				}

				return []cmtabci.ExtendedVoteInfo{
					mockVoteExtension("val1Bridge", 100, tmproto.BlockIDFlagCommit, events1...),
					mockVoteExtension("val2Bridge", 100, tmproto.BlockIDFlagCommit, events1...),
					mockVoteExtension("val3Bridge", 100, tmproto.BlockIDFlagCommit, events2...),
					mockVoteExtension("val4Bridge", 100, tmproto.BlockIDFlagCommit, events2...),
					mockVoteExtension("val5NonBridge", 100, tmproto.BlockIDFlagCommit, events1...),
					mockVoteExtension("val6NonBridge", 100, tmproto.BlockIDFlagCommit, events1...),
					mockVoteExtension("val7NonBridge", 100, tmproto.BlockIDFlagCommit, events2...),
					mockVoteExtension("val8NonBridge", 100, tmproto.BlockIDFlagCommit, events2...),
				}
			},
			expectedEvents: bridgetypes.AssetsLockedEvents{},
			errContains:    "",
		},
		{
			name:     "no canonical sequence - bridge validators not supporting",
			valStore: newMockValidatorStore("val1Bridge", "val2Bridge", "val3Bridge", "val4Bridge"),
			voteExtensionDecomposerFn: func() *mockVoteExtensionDecomposer {
				return newMockVoteExtensionDecomposer().withReturnInputMode()
			},
			reqVoteExtensionsFn: func() []cmtabci.ExtendedVoteInfo {
				events := []bridgetypes.AssetsLockedEvent{
					mockEvent(201, recipient1, 100, token),
					mockEvent(202, recipient2, 200, token),
					mockEvent(203, recipient3, 300, token),
				}

				// val3Bridge and val4Bridge are not supporting the canonical
				// sequence.
				return []cmtabci.ExtendedVoteInfo{
					mockVoteExtension("val1Bridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val2Bridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val3Bridge", 100, tmproto.BlockIDFlagCommit),
					mockVoteExtension("val4Bridge", 100, tmproto.BlockIDFlagCommit),
					mockVoteExtension("val5NonBridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val6NonBridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val7NonBridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val8NonBridge", 100, tmproto.BlockIDFlagCommit, events...),
				}
			},
			expectedEvents: bridgetypes.AssetsLockedEvents{},
			errContains:    "",
		},
		{
			name:     "no canonical sequence - non-bridge validators not supporting",
			valStore: newMockValidatorStore("val1Bridge", "val2Bridge", "val3Bridge", "val4Bridge"),
			voteExtensionDecomposerFn: func() *mockVoteExtensionDecomposer {
				return newMockVoteExtensionDecomposer().withReturnInputMode()
			},
			reqVoteExtensionsFn: func() []cmtabci.ExtendedVoteInfo {
				events := []bridgetypes.AssetsLockedEvent{
					mockEvent(201, recipient1, 100, token),
					mockEvent(202, recipient2, 200, token),
					mockEvent(203, recipient3, 300, token),
				}

				// val7NonBridge and val8NonBridge are not supporting the canonical
				// sequence.
				return []cmtabci.ExtendedVoteInfo{
					mockVoteExtension("val1Bridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val2Bridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val3Bridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val4Bridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val5NonBridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val6NonBridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val7NonBridge", 100, tmproto.BlockIDFlagCommit),
					mockVoteExtension("val8NonBridge", 100, tmproto.BlockIDFlagCommit),
				}
			},
			expectedEvents: bridgetypes.AssetsLockedEvents{},
			errContains:    "",
		},
		{
			name:     "no canonical sequence - bridge validators not existing",
			valStore: newMockValidatorStore(), // We do not set any bridge validators.
			voteExtensionDecomposerFn: func() *mockVoteExtensionDecomposer {
				return newMockVoteExtensionDecomposer().withReturnInputMode()
			},
			reqVoteExtensionsFn: func() []cmtabci.ExtendedVoteInfo {
				events := []bridgetypes.AssetsLockedEvent{
					mockEvent(201, recipient1, 100, token),
					mockEvent(202, recipient2, 200, token),
					mockEvent(203, recipient3, 300, token),
				}

				return []cmtabci.ExtendedVoteInfo{
					mockVoteExtension("val5NonBridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val6NonBridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val7NonBridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val8NonBridge", 100, tmproto.BlockIDFlagCommit, events...),
				}
			},
			expectedEvents: bridgetypes.AssetsLockedEvents{},
			errContains:    "",
		},
		{
			name:     "no canonical sequence - non-bridge validators not existing",
			valStore: newMockValidatorStore("val1Bridge", "val2Bridge", "val3Bridge", "val4Bridge"),
			voteExtensionDecomposerFn: func() *mockVoteExtensionDecomposer {
				return newMockVoteExtensionDecomposer().withReturnInputMode()
			},
			reqVoteExtensionsFn: func() []cmtabci.ExtendedVoteInfo {
				events := []bridgetypes.AssetsLockedEvent{
					mockEvent(201, recipient1, 100, token),
					mockEvent(202, recipient2, 200, token),
					mockEvent(203, recipient3, 300, token),
				}

				return []cmtabci.ExtendedVoteInfo{
					mockVoteExtension("val1Bridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val2Bridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val3Bridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val4Bridge", 100, tmproto.BlockIDFlagCommit, events...),
				}
			},
			expectedEvents: bridgetypes.AssetsLockedEvents{},
			errContains:    "",
		},
		{
			name:     "no canonical sequence - too many VEs rejected - wrong block ID flag (unknown)",
			valStore: newMockValidatorStore("val1Bridge", "val2Bridge", "val3Bridge", "val4Bridge"),
			voteExtensionDecomposerFn: func() *mockVoteExtensionDecomposer {
				return newMockVoteExtensionDecomposer().withReturnInputMode()
			},
			reqVoteExtensionsFn: func() []cmtabci.ExtendedVoteInfo {
				events := []bridgetypes.AssetsLockedEvent{
					mockEvent(201, recipient1, 100, token),
					mockEvent(202, recipient2, 200, token),
					mockEvent(203, recipient3, 300, token),
				}

				// Canonical sequence is not produced because val1Bridge's vote
				// is rejected due to the wrong flag.
				return []cmtabci.ExtendedVoteInfo{
					mockVoteExtension("val1Bridge", 100, tmproto.BlockIDFlagUnknown, events...),
					mockVoteExtension("val2Bridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val3Bridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val4Bridge", 100, tmproto.BlockIDFlagCommit),
					mockVoteExtension("val5NonBridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val6NonBridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val7NonBridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val8NonBridge", 100, tmproto.BlockIDFlagCommit, events...),
				}
			},
			expectedEvents: bridgetypes.AssetsLockedEvents{},
			errContains:    "",
		},
		{
			name:     "no canonical sequence - too many VEs rejected - wrong block ID flag (absent)",
			valStore: newMockValidatorStore("val1Bridge", "val2Bridge", "val3Bridge", "val4Bridge"),
			voteExtensionDecomposerFn: func() *mockVoteExtensionDecomposer {
				return newMockVoteExtensionDecomposer().withReturnInputMode()
			},
			reqVoteExtensionsFn: func() []cmtabci.ExtendedVoteInfo {
				events := []bridgetypes.AssetsLockedEvent{
					mockEvent(201, recipient1, 100, token),
					mockEvent(202, recipient2, 200, token),
					mockEvent(203, recipient3, 300, token),
				}

				// Canonical sequence is not produced because val1Bridge's vote
				// is rejected due to the wrong flag.
				return []cmtabci.ExtendedVoteInfo{
					mockVoteExtension("val1Bridge", 100, tmproto.BlockIDFlagAbsent, events...),
					mockVoteExtension("val2Bridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val3Bridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val4Bridge", 100, tmproto.BlockIDFlagCommit),
					mockVoteExtension("val5NonBridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val6NonBridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val7NonBridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val8NonBridge", 100, tmproto.BlockIDFlagCommit, events...),
				}
			},
			expectedEvents: bridgetypes.AssetsLockedEvents{},
			errContains:    "",
		},
		{
			name:     "no canonical sequence - too many VEs rejected - wrong block ID flag (nil)",
			valStore: newMockValidatorStore("val1Bridge", "val2Bridge", "val3Bridge", "val4Bridge"),
			voteExtensionDecomposerFn: func() *mockVoteExtensionDecomposer {
				return newMockVoteExtensionDecomposer().withReturnInputMode()
			},
			reqVoteExtensionsFn: func() []cmtabci.ExtendedVoteInfo {
				events := []bridgetypes.AssetsLockedEvent{
					mockEvent(201, recipient1, 100, token),
					mockEvent(202, recipient2, 200, token),
					mockEvent(203, recipient3, 300, token),
				}

				// Canonical sequence is not produced because val1Bridge's vote
				// is rejected due to the wrong flag.
				return []cmtabci.ExtendedVoteInfo{
					mockVoteExtension("val1Bridge", 100, tmproto.BlockIDFlagNil, events...),
					mockVoteExtension("val2Bridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val3Bridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val4Bridge", 100, tmproto.BlockIDFlagCommit),
					mockVoteExtension("val5NonBridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val6NonBridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val7NonBridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val8NonBridge", 100, tmproto.BlockIDFlagCommit, events...),
				}
			},
			expectedEvents: bridgetypes.AssetsLockedEvents{},
			errContains:    "",
		},
		{
			name:     "no canonical sequence - too many VEs rejected - empty composite",
			valStore: newMockValidatorStore("val1Bridge", "val2Bridge", "val3Bridge", "val4Bridge"),
			voteExtensionDecomposerFn: func() *mockVoteExtensionDecomposer {
				return newMockVoteExtensionDecomposer().withReturnInputMode()
			},
			reqVoteExtensionsFn: func() []cmtabci.ExtendedVoteInfo {
				events := []bridgetypes.AssetsLockedEvent{
					mockEvent(201, recipient1, 100, token),
					mockEvent(202, recipient2, 200, token),
					mockEvent(203, recipient3, 300, token),
				}

				// Canonical sequence is not produced because val1Bridge's vote
				// is rejected due to being empty.
				return []cmtabci.ExtendedVoteInfo{
					{
						Validator: cmtabci.Validator{
							Address: []byte("val1Bridge"),
							Power:   100,
						},
						VoteExtension: nil,
						BlockIdFlag:   tmproto.BlockIDFlagCommit,
					},
					mockVoteExtension("val2Bridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val3Bridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val4Bridge", 100, tmproto.BlockIDFlagCommit),
					mockVoteExtension("val5NonBridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val6NonBridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val7NonBridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val8NonBridge", 100, tmproto.BlockIDFlagCommit, events...),
				}
			},
			expectedEvents: bridgetypes.AssetsLockedEvents{},
			errContains:    "",
		},
		{
			name:     "no canonical sequence - too many VEs rejected - decomposition error",
			valStore: newMockValidatorStore("val1Bridge", "val2Bridge", "val3Bridge", "val4Bridge"),
			voteExtensionDecomposerFn: func() *mockVoteExtensionDecomposer {
				decomposer := newMockVoteExtensionDecomposer()

				// Return error while decomposing the first vote extension -
				// the one from val1Bridge.
				decomposer.On("call", mock.Anything).Return(
					nil,
					fmt.Errorf("decomposition error"),
				).Once()

				return decomposer.withReturnInputMode()
			},
			reqVoteExtensionsFn: func() []cmtabci.ExtendedVoteInfo {
				events := []bridgetypes.AssetsLockedEvent{
					mockEvent(201, recipient1, 100, token),
					mockEvent(202, recipient2, 200, token),
					mockEvent(203, recipient3, 300, token),
				}

				// Canonical sequence is not produced because val1Bridge's vote
				// is rejected due decomposition error.
				return []cmtabci.ExtendedVoteInfo{
					mockVoteExtension("val1Bridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val2Bridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val3Bridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val4Bridge", 100, tmproto.BlockIDFlagCommit),
					mockVoteExtension("val5NonBridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val6NonBridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val7NonBridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val8NonBridge", 100, tmproto.BlockIDFlagCommit, events...),
				}
			},
			expectedEvents: bridgetypes.AssetsLockedEvents{},
			errContains:    "",
		},
		{
			name:     "no canonical sequence - too many VEs rejected - empty bridge-specific part",
			valStore: newMockValidatorStore("val1Bridge", "val2Bridge", "val3Bridge", "val4Bridge"),
			voteExtensionDecomposerFn: func() *mockVoteExtensionDecomposer {
				decomposer := newMockVoteExtensionDecomposer()

				// Return nil slice while decomposing the first vote extension -
				// the one from val1Bridge.
				decomposer.On("call", mock.Anything).Return(
					nil,
					nil,
				).Once()

				return decomposer.withReturnInputMode()
			},
			reqVoteExtensionsFn: func() []cmtabci.ExtendedVoteInfo {
				events := []bridgetypes.AssetsLockedEvent{
					mockEvent(201, recipient1, 100, token),
					mockEvent(202, recipient2, 200, token),
					mockEvent(203, recipient3, 300, token),
				}

				// Canonical sequence is not produced because val1Bridge's vote
				// is rejected due having an empty bridge-specific part.
				return []cmtabci.ExtendedVoteInfo{
					mockVoteExtension("val1Bridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val2Bridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val3Bridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val4Bridge", 100, tmproto.BlockIDFlagCommit),
					mockVoteExtension("val5NonBridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val6NonBridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val7NonBridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val8NonBridge", 100, tmproto.BlockIDFlagCommit, events...),
				}
			},
			expectedEvents: bridgetypes.AssetsLockedEvents{},
			errContains:    "",
		},
		{
			name:     "no canonical sequence - too many VEs rejected - unmarshaling error",
			valStore: newMockValidatorStore("val1Bridge", "val2Bridge", "val3Bridge", "val4Bridge"),
			voteExtensionDecomposerFn: func() *mockVoteExtensionDecomposer {
				return newMockVoteExtensionDecomposer().withReturnInputMode()
			},
			reqVoteExtensionsFn: func() []cmtabci.ExtendedVoteInfo {
				events := []bridgetypes.AssetsLockedEvent{
					mockEvent(201, recipient1, 100, token),
					mockEvent(202, recipient2, 200, token),
					mockEvent(203, recipient3, 300, token),
				}

				// Canonical sequence is not produced because val1Bridge's vote
				// is rejected due to unmarshaling error.
				return []cmtabci.ExtendedVoteInfo{
					{
						Validator: cmtabci.Validator{
							Address: []byte("val1Bridge"),
							Power:   100,
						},
						VoteExtension: []byte("corrupted"),
						BlockIdFlag:   tmproto.BlockIDFlagCommit,
					},
					mockVoteExtension("val2Bridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val3Bridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val4Bridge", 100, tmproto.BlockIDFlagCommit),
					mockVoteExtension("val5NonBridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val6NonBridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val7NonBridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val8NonBridge", 100, tmproto.BlockIDFlagCommit, events...),
				}
			},
			expectedEvents: bridgetypes.AssetsLockedEvents{},
			errContains:    "",
		},
		{
			name:     "no canonical sequence - too many VEs rejected - empty sequence",
			valStore: newMockValidatorStore("val1Bridge", "val2Bridge", "val3Bridge", "val4Bridge"),
			voteExtensionDecomposerFn: func() *mockVoteExtensionDecomposer {
				return newMockVoteExtensionDecomposer().withReturnInputMode()
			},
			reqVoteExtensionsFn: func() []cmtabci.ExtendedVoteInfo {
				events := []bridgetypes.AssetsLockedEvent{
					mockEvent(201, recipient1, 100, token),
					mockEvent(202, recipient2, 200, token),
					mockEvent(203, recipient3, 300, token),
				}

				// Canonical sequence is not produced because val1Bridge's vote
				// is rejected due to a vote extension without events.
				return []cmtabci.ExtendedVoteInfo{
					mockVoteExtension("val1Bridge", 100, tmproto.BlockIDFlagCommit),
					mockVoteExtension("val2Bridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val3Bridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val4Bridge", 100, tmproto.BlockIDFlagCommit),
					mockVoteExtension("val5NonBridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val6NonBridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val7NonBridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val8NonBridge", 100, tmproto.BlockIDFlagCommit, events...),
				}
			},
			expectedEvents: bridgetypes.AssetsLockedEvents{},
			errContains:    "",
		},
		{
			name:     "no canonical sequence - too many VEs rejected - sequence exceeding the limit",
			valStore: newMockValidatorStore("val1Bridge", "val2Bridge", "val3Bridge", "val4Bridge"),
			voteExtensionDecomposerFn: func() *mockVoteExtensionDecomposer {
				return newMockVoteExtensionDecomposer().withReturnInputMode()
			},
			reqVoteExtensionsFn: func() []cmtabci.ExtendedVoteInfo {
				events := []bridgetypes.AssetsLockedEvent{
					mockEvent(201, recipient1, 100, token),
					mockEvent(202, recipient2, 200, token),
					mockEvent(203, recipient3, 300, token),
				}

				val1BridgeEvents := []bridgetypes.AssetsLockedEvent{
					events[0],
					events[1],
					events[2],
					mockEvent(204, recipient3, 300, token),
					mockEvent(205, recipient3, 300, token),
					mockEvent(206, recipient3, 300, token),
					mockEvent(207, recipient3, 300, token),
					mockEvent(208, recipient3, 300, token),
					mockEvent(209, recipient3, 300, token),
					mockEvent(210, recipient3, 300, token),
					mockEvent(211, recipient3, 300, token),
				}

				// Canonical sequence is not produced because val1Bridge's vote
				// is rejected due to a vote extension with to many events
				// (see AssetsLockedEventsLimit).
				return []cmtabci.ExtendedVoteInfo{
					mockVoteExtension("val1Bridge", 100, tmproto.BlockIDFlagCommit, val1BridgeEvents...),
					mockVoteExtension("val2Bridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val3Bridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val4Bridge", 100, tmproto.BlockIDFlagCommit),
					mockVoteExtension("val5NonBridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val6NonBridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val7NonBridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val8NonBridge", 100, tmproto.BlockIDFlagCommit, events...),
				}
			},
			expectedEvents: bridgetypes.AssetsLockedEvents{},
			errContains:    "",
		},
		{
			name:     "no canonical sequence - too many VEs rejected - invalid sequence",
			valStore: newMockValidatorStore("val1Bridge", "val2Bridge", "val3Bridge", "val4Bridge"),
			voteExtensionDecomposerFn: func() *mockVoteExtensionDecomposer {
				return newMockVoteExtensionDecomposer().withReturnInputMode()
			},
			reqVoteExtensionsFn: func() []cmtabci.ExtendedVoteInfo {
				events := []bridgetypes.AssetsLockedEvent{
					mockEvent(201, recipient1, 100, token),
					mockEvent(202, recipient2, 200, token),
					mockEvent(203, recipient3, 300, token),
				}

				val1BridgeEvents := []bridgetypes.AssetsLockedEvent{
					events[2],
					events[1],
					events[0],
				}

				// Canonical sequence is not produced because val1Bridge's vote
				// is rejected due to a vote extension with to an invalid
				// events sequence.
				return []cmtabci.ExtendedVoteInfo{
					mockVoteExtension("val1Bridge", 100, tmproto.BlockIDFlagCommit, val1BridgeEvents...),
					mockVoteExtension("val2Bridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val3Bridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val4Bridge", 100, tmproto.BlockIDFlagCommit),
					mockVoteExtension("val5NonBridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val6NonBridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val7NonBridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val8NonBridge", 100, tmproto.BlockIDFlagCommit, events...),
				}
			},
			expectedEvents: bridgetypes.AssetsLockedEvents{},
			errContains:    "",
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			s.SetupTest()

			valStore := test.valStore
			voteExtensionDecomposer := test.voteExtensionDecomposerFn()

			s.extractor = newAssetsLockedExtractor(
				s.logger,
				valStore,
				voteExtensionDecomposer.call,
			)

			extendedCommitInfo := cmtabci.ExtendedCommitInfo{
				Round: 1, // Just an arbitrary value.
				Votes: test.reqVoteExtensionsFn(),
			}

			events, err := s.extractor.CanonicalEvents(
				s.ctx,
				extendedCommitInfo,
				s.height,
			)

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
				test.expectedEvents,
				events,
				"expected different events",
			)
		})
	}
}

//nolint:all
func mockVoteExtension(
	valAddress string,
	power int64,
	blockIdFlag tmproto.BlockIDFlag,
	events ...bridgetypes.AssetsLockedEvent,
) cmtabci.ExtendedVoteInfo {
	voteExtension := &types.VoteExtension{AssetsLockedEvents: events}
	voteExtensionBytes, err := voteExtension.Marshal()
	if err != nil {
		panic(err)
	}

	return cmtabci.ExtendedVoteInfo{
		Validator: cmtabci.Validator{
			Address: []byte(valAddress),
			Power:   power,
		},
		VoteExtension:      voteExtensionBytes,
		ExtensionSignature: nil, // not relevant for the test
		BlockIdFlag:        blockIdFlag,
	}
}

func mockEvent(
	sequence int64,
	recipient string,
	amount int64,
	token string,
) bridgetypes.AssetsLockedEvent {
	return bridgetypes.AssetsLockedEvent{
		Sequence:  sdkmath.NewInt(sequence),
		Recipient: recipient,
		Amount:    sdkmath.NewInt(amount),
		Token:     token,
	}
}

func txsVector(txs ...string) [][]byte {
	res := make([][]byte, 0)

	for _, tx := range txs {
		res = append(res, []byte(tx))
	}

	return res
}

func extractInjectedTx(
	req *cmtabci.RequestPrepareProposal,
	res *cmtabci.ResponsePrepareProposal,
) []byte {
	safeFirstTx := func(txs [][]byte) []byte {
		if len(txs) == 0 {
			return nil
		}

		return txs[0]
	}

	firstReqTx := safeFirstTx(req.Txs)
	firstResTx := safeFirstTx(res.Txs)

	if len(firstResTx) == 0 || bytes.Equal(firstResTx, firstReqTx) {
		return nil
	}

	return firstResTx
}

type mockValidatorStore struct {
	mock.Mock
}

func newMockValidatorStore(bridgeVals ...string) *mockValidatorStore {
	valStore := &mockValidatorStore{}

	consAddresses := make([]sdk.ConsAddress, len(bridgeVals))
	for i, val := range bridgeVals {
		consAddresses[i] = sdk.ConsAddress(val)
	}

	valStore.On(
		"GetValidatorsConsAddrsByPrivilege",
		mock.Anything,
		bridgetypes.ValidatorPrivilege,
	).Return(consAddresses)

	return valStore
}

func (mvs *mockValidatorStore) GetPubKeyByConsAddr(
	ctx context.Context,
	address sdk.ConsAddress,
) (crypto.PublicKey, error) {
	args := mvs.Called(ctx, address)

	if res := args.Get(0); res != nil {
		return res.(crypto.PublicKey), args.Error(1)
	}

	return crypto.PublicKey{}, args.Error(1)
}

func (mvs *mockValidatorStore) GetValidatorsConsAddrsByPrivilege(
	ctx sdk.Context,
	privilege string,
) []sdk.ConsAddress {
	args := mvs.Called(ctx, privilege)

	return args.Get(0).([]sdk.ConsAddress)
}

type mockVoteExtensionsValidator struct {
	mock.Mock
}

func newMockVoteExtensionsValidator(result error) *mockVoteExtensionsValidator {
	validator := &mockVoteExtensionsValidator{}

	validator.On(
		"call",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(result)

	return validator
}

func (mvev *mockVoteExtensionsValidator) call(
	ctx sdk.Context,
	valStore baseapp.ValidatorStore,
	height int64,
	chainID string,
	extCommit cmtabci.ExtendedCommitInfo,
) error {
	args := mvev.Called(ctx, valStore, height, chainID, extCommit)

	return args.Error(0)
}

type mockVoteExtensionDecomposer struct {
	mock.Mock
}

func newMockVoteExtensionDecomposer() *mockVoteExtensionDecomposer {
	return &mockVoteExtensionDecomposer{}
}

func (mved *mockVoteExtensionDecomposer) withReturnInputMode() *mockVoteExtensionDecomposer {
	mved.On("call", mock.Anything).Return(
		func(input []byte) []byte { return input },
		nil,
	)

	return mved
}

func (mved *mockVoteExtensionDecomposer) call(compositeVoteExtensionBytes []byte) (
	voteExtensionBytes []byte,
	err error,
) {
	args := mved.Called(compositeVoteExtensionBytes)

	if res := args.Get(0); res != nil {
		if resFn, ok := res.(func([]byte) []byte); ok {
			return resFn(compositeVoteExtensionBytes), args.Error(1)
		}

		return res.([]byte), args.Error(1)
	}

	return nil, args.Error(1)
}

type mockAssetsLockedExtractor struct {
	mock.Mock
}

func newMockAssetsLockedExtractor() *mockAssetsLockedExtractor {
	return &mockAssetsLockedExtractor{}
}

func (male *mockAssetsLockedExtractor) CanonicalEvents(
	ctx sdk.Context,
	extendedCommitInfo cmtabci.ExtendedCommitInfo,
	height int64,
) (bridgetypes.AssetsLockedEvents, error) {
	args := male.Called(ctx, extendedCommitInfo, height)

	if res := args.Get(0); res != nil {
		return res.([]bridgetypes.AssetsLockedEvent), args.Error(1)
	}

	return nil, args.Error(1)
}

func marshalInjectedTx(injectedTx types.InjectedTx) []byte {
	injectedTxBytes, err := injectedTx.Marshal()
	if err != nil {
		panic(err)
	}

	return injectedTxBytes
}
