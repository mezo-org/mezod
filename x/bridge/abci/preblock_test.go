package abci

import (
	"fmt"
	"testing"

	"cosmossdk.io/math"

	"github.com/mezo-org/mezod/x/bridge/abci/types"
	bridgetypes "github.com/mezo-org/mezod/x/bridge/types"

	"cosmossdk.io/log"
	cmtabci "github.com/cometbft/cometbft/abci/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	servermock "github.com/cosmos/cosmos-sdk/server/mock"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/mezo-org/mezod/cmd/config"
	"github.com/stretchr/testify/suite"
)

func TestPreBlockHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(PreBlockHandlerTestSuite))
}

type PreBlockHandlerTestSuite struct {
	suite.Suite

	logger        log.Logger
	ctx           sdk.Context
	requestHeight int64
	handler       *PreBlockHandler
}

func (s *PreBlockHandlerTestSuite) SetupTest() {
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
}

func (s *PreBlockHandlerTestSuite) TestPreBlocker() {
	tests := []struct {
		name           string
		bridgeKeeperFn func() *mockBridgeKeeper
		reqTxs         [][]byte
		expectedRes    *sdk.ResponsePreBlock
		errContains    string
	}{
		{
			name:           "no txs in the request",
			bridgeKeeperFn: newMockBridgeKeeper,
			reqTxs:         txsVector(),
			expectedRes:    nil,
			errContains:    "empty transaction vector in the block",
		},
		{
			name:           "empty injected tx",
			bridgeKeeperFn: newMockBridgeKeeper,
			reqTxs:         txsVector(""),
			expectedRes:    &sdk.ResponsePreBlock{},
			errContains:    "",
		},
		{
			name:           "non-unmarshalable injected tx",
			bridgeKeeperFn: newMockBridgeKeeper,
			reqTxs:         txsVector("corrupted"),
			expectedRes:    nil,
			errContains:    "failed to unmarshal injected tx",
		},
		{
			name: "keeper not accepting events",
			bridgeKeeperFn: func() *mockBridgeKeeper {
				bridgeKeeper := newMockBridgeKeeper()

				bridgeKeeper.On(
					"AcceptAssetsLocked",
					s.ctx,
					bridgetypes.AssetsLockedEvents{
						mockEvent(1, recipient1, 100, token),
						mockEvent(2, recipient2, 200, token),
					},
				).Return(fmt.Errorf("keeper error"))

				return bridgeKeeper
			},
			reqTxs: [][]byte{
				marshalInjectedTx(
					types.InjectedTx{
						AssetsLockedEvents: bridgetypes.AssetsLockedEvents{
							mockEvent(1, recipient1, 100, token),
							mockEvent(2, recipient2, 200, token),
						},
						ExtendedCommitInfo: []byte("extendedCommitInfo"),
					},
				),
			},
			expectedRes: nil,
			errContains: "cannot accept AssetsLocked events",
		},
		{
			name: "keeper accepting events",
			bridgeKeeperFn: func() *mockBridgeKeeper {
				bridgeKeeper := newMockBridgeKeeper()

				bridgeKeeper.On(
					"AcceptAssetsLocked",
					s.ctx,
					bridgetypes.AssetsLockedEvents{
						mockEvent(1, recipient1, 100, token),
						mockEvent(2, recipient2, 200, token),
					},
				).Return(nil)

				bridgeKeeper.On(
					"GetAssetsLockedSequenceTip",
					s.ctx,
				).Return(math.NewInt(2))

				return bridgeKeeper
			},
			reqTxs: [][]byte{
				marshalInjectedTx(
					types.InjectedTx{
						AssetsLockedEvents: bridgetypes.AssetsLockedEvents{
							mockEvent(1, recipient1, 100, token),
							mockEvent(2, recipient2, 200, token),
						},
						ExtendedCommitInfo: []byte("extendedCommitInfo"),
					},
				),
			},
			expectedRes: &sdk.ResponsePreBlock{},
			errContains: "",
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			s.SetupTest()

			bridgeKeeper := test.bridgeKeeperFn()

			s.handler = NewPreBlockHandler(
				s.logger,
				bridgeKeeper,
			)

			req := &cmtabci.RequestFinalizeBlock{
				// Fill only the fields that are relevant for the test.
				Height:          s.requestHeight,
				Txs:             test.reqTxs,
				ProposerAddress: []byte("proposerAddress"),
			}

			res, err := s.handler.PreBlocker()(s.ctx, req)

			bridgeKeeper.AssertExpectations(s.T())

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
