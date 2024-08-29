package abci

import (
	"context"
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/mezo-org/mezod/x/bridge/keeper"
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
	keeper        keeper.Keeper
	handler       *VoteExtensionHandler
}

func (s *VoteExtensionHandlerTestSuite) SetupTest() {
	s.logger = log.NewNopLogger()

	multiStore := servermock.NewCommitMultiStore()

	s.ctx = sdk.NewContext(
		multiStore,
		tmproto.Header{},
		false,
		s.logger,
	)

	s.requestHeight = 100

	storeKey := storetypes.NewKVStoreKey(bridgetypes.StoreKey)

	// Only the first argument is relevant for the mock multi store.
	multiStore.MountStoreWithDB(storeKey, 0, nil)

	s.keeper = keeper.NewKeeper(
		codec.NewProtoCodec(codectypes.NewInterfaceRegistry()),
		storeKey,
	)

	s.keeper.SetAssetsLockedSequenceTip(s.ctx, sdkmath.NewInt(200))
}

func (s *VoteExtensionHandlerTestSuite) TestExtendVote() {
	mockEvent := func(sequence int64) bridgetypes.AssetsLockedEvent {
		return bridgetypes.AssetsLockedEvent{
			Sequence:  sdkmath.NewInt(sequence),
			Recipient: "recipient",
			Amount:    sdkmath.ZeroInt(),
		}
	}

	tests := []struct {
		name        string
		sidecarFn   func() EthereumSidecarClient
		expectedVE  *types.VoteExtension
		errContains string
	}{
		{
			name: "sidecar returning error",
			sidecarFn: func() EthereumSidecarClient {
				sidecar := newMockEthereumSidecarClient()

				sidecar.On(
					"GetAssetsLockedEvents",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil, fmt.Errorf("sidecar error"))

				return sidecar
			},
			expectedVE:  nil,
			errContains: "failed to fetch AssetsLocked events from the sidecar",
		},
		{
			name: "sidecar returning empty slice",
			sidecarFn: func() EthereumSidecarClient {
				sidecar := newMockEthereumSidecarClient()

				sidecar.On(
					"GetAssetsLockedEvents",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return([]bridgetypes.AssetsLockedEvent{}, nil)

				return sidecar
			},
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
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil, nil)

				return sidecar
			},
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
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return([]bridgetypes.AssetsLockedEvent{
					mockEvent(1),
				}, nil)

				return sidecar
			},
			expectedVE: &types.VoteExtension{
				AssetsLockedEvents: []bridgetypes.AssetsLockedEvent{
					mockEvent(1),
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
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return([]bridgetypes.AssetsLockedEvent{
					mockEvent(3),
					mockEvent(2),
					mockEvent(1),
				}, nil)

				return sidecar
			},
			expectedVE:  nil,
			errContains: "events do not form a proper sequence",
		},
		{
			name: "sidecar returning improper sequence - increasing (non-strictly)",
			sidecarFn: func() EthereumSidecarClient {
				sidecar := newMockEthereumSidecarClient()

				sidecar.On(
					"GetAssetsLockedEvents",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return([]bridgetypes.AssetsLockedEvent{
					mockEvent(1),
					mockEvent(1),
					mockEvent(2),
					mockEvent(3),
				}, nil)

				return sidecar
			},
			expectedVE:  nil,
			errContains: "events do not form a proper sequence",
		},
		{
			name: "sidecar returning improper sequence - decreasing (non-strictly)",
			sidecarFn: func() EthereumSidecarClient {
				sidecar := newMockEthereumSidecarClient()

				sidecar.On(
					"GetAssetsLockedEvents",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return([]bridgetypes.AssetsLockedEvent{
					mockEvent(3),
					mockEvent(3),
					mockEvent(2),
					mockEvent(1),
				}, nil)

				return sidecar
			},
			expectedVE:  nil,
			errContains: "events do not form a proper sequence",
		},
		{
			name: "sidecar returning improper sequence - gap",
			sidecarFn: func() EthereumSidecarClient {
				sidecar := newMockEthereumSidecarClient()

				sidecar.On(
					"GetAssetsLockedEvents",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return([]bridgetypes.AssetsLockedEvent{
					mockEvent(1),
					mockEvent(2),
					mockEvent(4),
				}, nil)

				return sidecar
			},
			expectedVE:  nil,
			errContains: "events do not form a proper sequence",
		},
		{
			name: "sidecar returning improper sequence - duplicate",
			sidecarFn: func() EthereumSidecarClient {
				sidecar := newMockEthereumSidecarClient()

				sidecar.On(
					"GetAssetsLockedEvents",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return([]bridgetypes.AssetsLockedEvent{
					mockEvent(1),
					mockEvent(2),
					mockEvent(3),
					mockEvent(1),
				}, nil)

				return sidecar
			},
			expectedVE:  nil,
			errContains: "events do not form a proper sequence",
		},
		{
			name: "sidecar returning more events than the limit",
			sidecarFn: func() EthereumSidecarClient {
				sidecar := newMockEthereumSidecarClient()

				sidecar.On(
					"GetAssetsLockedEvents",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return([]bridgetypes.AssetsLockedEvent{
					mockEvent(1),
					mockEvent(2),
					mockEvent(3),
					mockEvent(4),
					mockEvent(5),
					mockEvent(6),
					mockEvent(7),
					mockEvent(8),
					mockEvent(9),
					mockEvent(10),
					mockEvent(11),
					mockEvent(12),
				}, nil)

				return sidecar
			},
			expectedVE: &types.VoteExtension{
				AssetsLockedEvents: []bridgetypes.AssetsLockedEvent{
					mockEvent(1),
					mockEvent(2),
					mockEvent(3),
					mockEvent(4),
					mockEvent(5),
					mockEvent(6),
					mockEvent(7),
					mockEvent(8),
					mockEvent(9),
					mockEvent(10),
				},
			},
			errContains: "",
		},
		{
			name: "sidecar returning events within the limit",
			sidecarFn: func() EthereumSidecarClient {
				sidecar := newMockEthereumSidecarClient()

				sidecar.On(
					"GetAssetsLockedEvents",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return([]bridgetypes.AssetsLockedEvent{
					mockEvent(1),
					mockEvent(2),
					mockEvent(3),
					mockEvent(4),
					mockEvent(5),
					mockEvent(6),
					mockEvent(7),
					mockEvent(8),
					mockEvent(9),
					mockEvent(10),
				}, nil)

				return sidecar
			},
			expectedVE: &types.VoteExtension{
				AssetsLockedEvents: []bridgetypes.AssetsLockedEvent{
					mockEvent(1),
					mockEvent(2),
					mockEvent(3),
					mockEvent(4),
					mockEvent(5),
					mockEvent(6),
					mockEvent(7),
					mockEvent(8),
					mockEvent(9),
					mockEvent(10),
				},
			},
			errContains: "",
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			s.SetupTest()

			sidecar := test.sidecarFn()

			s.handler = NewVoteExtensionHandler(
				s.logger,
				sidecar,
				s.keeper,
			)

			req := &cmtabci.RequestExtendVote{
				Height: s.requestHeight,
			}

			res, err := s.handler.ExtendVoteHandler()(s.ctx, req)

			sequenceStart := sdkmath.NewInt(201)
			sequenceEnd := sdkmath.NewInt(211)

			sidecar.(*mockEthereumSidecarClient).AssertCalled(
				s.T(),
				"GetAssetsLockedEvents",
				s.ctx,
				&sequenceStart,
				&sequenceEnd,
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

	mockEvent := func(sequence int64) bridgetypes.AssetsLockedEvent {
		return bridgetypes.AssetsLockedEvent{
			Sequence:  sdkmath.NewInt(sequence),
			Recipient: "recipient",
			Amount:    sdkmath.ZeroInt(),
		}
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
			expectedRes:     nil,
			errContains:     "failed to unmarshal vote extension",
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
						mockEvent(1),
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
						mockEvent(3),
						mockEvent(2),
						mockEvent(1),
					},
				})
			},
			expectedRes: nil,
			errContains: "events do not form a proper sequence",
		},
		{
			name: "events slice forming improper sequence - increasing (non-strictly)",
			voteExtensionFn: func() []byte {
				return marshalVE(types.VoteExtension{
					AssetsLockedEvents: []bridgetypes.AssetsLockedEvent{
						mockEvent(1),
						mockEvent(1),
						mockEvent(2),
						mockEvent(3),
					},
				})
			},
			expectedRes: nil,
			errContains: "events do not form a proper sequence",
		},
		{
			name: "events slice forming improper sequence - decreasing (non-strictly)",
			voteExtensionFn: func() []byte {
				return marshalVE(types.VoteExtension{
					AssetsLockedEvents: []bridgetypes.AssetsLockedEvent{
						mockEvent(3),
						mockEvent(3),
						mockEvent(2),
						mockEvent(1),
					},
				})
			},
			expectedRes: nil,
			errContains: "events do not form a proper sequence",
		},
		{
			name: "events slice forming improper sequence - gap",
			voteExtensionFn: func() []byte {
				return marshalVE(types.VoteExtension{
					AssetsLockedEvents: []bridgetypes.AssetsLockedEvent{
						mockEvent(1),
						mockEvent(2),
						mockEvent(4),
					},
				})
			},
			expectedRes: nil,
			errContains: "events do not form a proper sequence",
		},
		{
			name: "events slice forming improper sequence - duplicate",
			voteExtensionFn: func() []byte {
				return marshalVE(types.VoteExtension{
					AssetsLockedEvents: []bridgetypes.AssetsLockedEvent{
						mockEvent(1),
						mockEvent(2),
						mockEvent(3),
						mockEvent(1),
					},
				})
			},
			expectedRes: nil,
			errContains: "events do not form a proper sequence",
		},
		{
			name: "events slice exceeding the limit",
			voteExtensionFn: func() []byte {
				return marshalVE(types.VoteExtension{
					AssetsLockedEvents: []bridgetypes.AssetsLockedEvent{
						mockEvent(1),
						mockEvent(2),
						mockEvent(3),
						mockEvent(4),
						mockEvent(5),
						mockEvent(6),
						mockEvent(7),
						mockEvent(8),
						mockEvent(9),
						mockEvent(10),
						mockEvent(11),
						mockEvent(12),
					},
				})
			},
			expectedRes: nil,
			errContains: "number of events exceeds the limit",
		},
		{
			name: "events slice within the limit",
			voteExtensionFn: func() []byte {
				return marshalVE(types.VoteExtension{
					AssetsLockedEvents: []bridgetypes.AssetsLockedEvent{
						mockEvent(1),
						mockEvent(2),
						mockEvent(3),
						mockEvent(4),
						mockEvent(5),
						mockEvent(6),
						mockEvent(7),
						mockEvent(8),
						mockEvent(9),
						mockEvent(10),
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
				s.keeper,
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
	sequenceStart *sdkmath.Int,
	sequenceEnd *sdkmath.Int,
) ([]bridgetypes.AssetsLockedEvent, error) {
	args := mesc.Called(ctx, sequenceStart, sequenceEnd)

	if res := args.Get(0); res != nil {
		return res.([]bridgetypes.AssetsLockedEvent), args.Error(1)
	}

	return nil, args.Error(1)
}
