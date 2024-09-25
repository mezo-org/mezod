package abci

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	cmtabci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/proto/tendermint/crypto"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	servermock "github.com/cosmos/cosmos-sdk/server/mock"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/mezo-org/mezod/cmd/config"
	"github.com/mezo-org/mezod/x/bridge/abci/types"
	"github.com/mezo-org/mezod/x/bridge/keeper"
	bridgetypes "github.com/mezo-org/mezod/x/bridge/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

func TestProposalHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(ProposalHandlerTestSuite))
}

type ProposalHandlerTestSuite struct {
	suite.Suite

	logger        log.Logger
	ctx           sdk.Context
	requestHeight int64
	keeper        keeper.Keeper
	handler       *ProposalHandler
}

func (s *ProposalHandlerTestSuite) SetupTest() {
	// Set bech32 prefixes to make the recipient address validation in
	// AssetsLocked events possible (see AssetsLockedEvent.IsValid).
	cfg := sdk.GetConfig()
	config.SetBech32Prefixes(cfg)

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

func (s *ProposalHandlerTestSuite) TestPrepareProposal() {
	tests := []struct {
		name                      string
		valStore                  *mockValidatorStore
		voteExtensionDecomposerFn func() *mockVoteExtensionDecomposer
		voteExtensionsValidator   *mockVoteExtensionsValidator
		reqTxs                    [][]byte
		reqVoteExtensionsFn       func() []cmtabci.ExtendedVoteInfo
		expectedInjectedTxFn      func([]byte) *types.InjectedTx
		expectedChainTxs          [][]byte // regular chain txs being part of the proposal
		errContains               string
	}{
		{
			name:                      "invalid vote extensions",
			valStore:                  newMockValidatorStore(),
			voteExtensionDecomposerFn: func() *mockVoteExtensionDecomposer {
				return newMockVoteExtensionDecomposer()
			},
			voteExtensionsValidator:   newMockVoteExtensionsValidator(fmt.Errorf("invalid vote extensions")),
			reqTxs:                    txsVector(),
			reqVoteExtensionsFn:       func() []cmtabci.ExtendedVoteInfo { return nil },
			expectedInjectedTxFn:      func(_ []byte) *types.InjectedTx { return nil },
			expectedChainTxs:          nil,
			errContains:               "failed to validate vote extensions",
		},
		{
			name:     "proper canonical sequence - unanimous voting",
			valStore: newMockValidatorStore("val1Bridge", "val2Bridge", "val3Bridge", "val4Bridge"),
			voteExtensionDecomposerFn: func() *mockVoteExtensionDecomposer {
				return newMockVoteExtensionDecomposer().withReturnInputMode()
			},
			voteExtensionsValidator: newMockVoteExtensionsValidator(nil),
			reqTxs:                  txsVector("tx1", "tx2"),
			reqVoteExtensionsFn: func() []cmtabci.ExtendedVoteInfo {
				events := []bridgetypes.AssetsLockedEvent{
					mockEvent(201, "mezo12wsc0qgyfwwfj3wrlpgm9q3lmndl2m4qmm34dp", 100),
					mockEvent(202, "mezo1xqurxvvh8z2xpj6wltq0tajxm47xnv7q6rtvja", 200),
					mockEvent(203, "mezo1jcurf087xx9eqsnmr8lszralupcfln2vjz8ucg", 300),
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
			expectedInjectedTxFn: func(extendedCommitInfo []byte) *types.InjectedTx {
				return &types.InjectedTx{
					AssetsLockedEvents: []bridgetypes.AssetsLockedEvent{
						mockEvent(201, "mezo12wsc0qgyfwwfj3wrlpgm9q3lmndl2m4qmm34dp", 100),
						mockEvent(202, "mezo1xqurxvvh8z2xpj6wltq0tajxm47xnv7q6rtvja", 200),
						mockEvent(203, "mezo1jcurf087xx9eqsnmr8lszralupcfln2vjz8ucg", 300),
					},
					ExtendedCommitInfo: extendedCommitInfo,
				}
			},
			expectedChainTxs: txsVector("tx1", "tx2"),
			errContains:      "",
		},
		{
			name:     "proper canonical sequence - some votes on different sequence",
			valStore: newMockValidatorStore("val1Bridge", "val2Bridge", "val3Bridge", "val4Bridge"),
			voteExtensionDecomposerFn: func() *mockVoteExtensionDecomposer {
				return newMockVoteExtensionDecomposer().withReturnInputMode()
			},
			voteExtensionsValidator: newMockVoteExtensionsValidator(nil),
			reqTxs:                  txsVector("tx1", "tx2"),
			reqVoteExtensionsFn: func() []cmtabci.ExtendedVoteInfo {
				// Events the super-majority of validators are voting on.
				eventsSm := []bridgetypes.AssetsLockedEvent{
					mockEvent(201, "mezo12wsc0qgyfwwfj3wrlpgm9q3lmndl2m4qmm34dp", 100),
					mockEvent(202, "mezo1xqurxvvh8z2xpj6wltq0tajxm47xnv7q6rtvja", 200),
					mockEvent(203, "mezo1jcurf087xx9eqsnmr8lszralupcfln2vjz8ucg", 300),
				}

				// Events the minority of validators are voting on. They
				// differ in the sequence numbers.
				eventsNonSm := []bridgetypes.AssetsLockedEvent{
					mockEvent(204, "mezo12wsc0qgyfwwfj3wrlpgm9q3lmndl2m4qmm34dp", 100),
					mockEvent(205, "mezo1xqurxvvh8z2xpj6wltq0tajxm47xnv7q6rtvja", 200),
					mockEvent(206, "mezo1jcurf087xx9eqsnmr8lszralupcfln2vjz8ucg", 300),
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
			expectedInjectedTxFn: func(extendedCommitInfo []byte) *types.InjectedTx {
				return &types.InjectedTx{
					AssetsLockedEvents: []bridgetypes.AssetsLockedEvent{
						mockEvent(201, "mezo12wsc0qgyfwwfj3wrlpgm9q3lmndl2m4qmm34dp", 100),
						mockEvent(202, "mezo1xqurxvvh8z2xpj6wltq0tajxm47xnv7q6rtvja", 200),
						mockEvent(203, "mezo1jcurf087xx9eqsnmr8lszralupcfln2vjz8ucg", 300),
					},
					ExtendedCommitInfo: extendedCommitInfo,
				}
			},
			expectedChainTxs: txsVector("tx1", "tx2"),
			errContains:      "",
		},
		{
			name:     "proper canonical sequence - some votes on different recipients",
			valStore: newMockValidatorStore("val1Bridge", "val2Bridge", "val3Bridge", "val4Bridge"),
			voteExtensionDecomposerFn: func() *mockVoteExtensionDecomposer {
				return newMockVoteExtensionDecomposer().withReturnInputMode()
			},
			voteExtensionsValidator: newMockVoteExtensionsValidator(nil),
			reqTxs:                  txsVector("tx1", "tx2"),
			reqVoteExtensionsFn: func() []cmtabci.ExtendedVoteInfo {
				// Events the super-majority of validators are voting on.
				eventsSm := []bridgetypes.AssetsLockedEvent{
					mockEvent(201, "mezo12wsc0qgyfwwfj3wrlpgm9q3lmndl2m4qmm34dp", 100),
					mockEvent(202, "mezo1xqurxvvh8z2xpj6wltq0tajxm47xnv7q6rtvja", 200),
					mockEvent(203, "mezo1jcurf087xx9eqsnmr8lszralupcfln2vjz8ucg", 300),
				}

				// Events the minority of validators are voting on. They
				// differ in the recipient address.
				eventsNonSm := []bridgetypes.AssetsLockedEvent{
					mockEvent(201, "mezo1j0ghx6d9kmerxhgn5ahr2nahs6yfulea4te22c", 100),
					mockEvent(202, "mezo120mkxfvkx2t72quddqh92md2dp7csq4wgqux06", 200),
					mockEvent(203, "mezo1dmr6mhh352vh9wa34xs0qxtr8thkqu39pw6x2p", 300),
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
			expectedInjectedTxFn: func(extendedCommitInfo []byte) *types.InjectedTx {
				return &types.InjectedTx{
					AssetsLockedEvents: []bridgetypes.AssetsLockedEvent{
						mockEvent(201, "mezo12wsc0qgyfwwfj3wrlpgm9q3lmndl2m4qmm34dp", 100),
						mockEvent(202, "mezo1xqurxvvh8z2xpj6wltq0tajxm47xnv7q6rtvja", 200),
						mockEvent(203, "mezo1jcurf087xx9eqsnmr8lszralupcfln2vjz8ucg", 300),
					},
					ExtendedCommitInfo: extendedCommitInfo,
				}
			},
			expectedChainTxs: txsVector("tx1", "tx2"),
			errContains:      "",
		},
		{
			name:     "proper canonical sequence - some votes on different amounts",
			valStore: newMockValidatorStore("val1Bridge", "val2Bridge", "val3Bridge", "val4Bridge"),
			voteExtensionDecomposerFn: func() *mockVoteExtensionDecomposer {
				return newMockVoteExtensionDecomposer().withReturnInputMode()
			},
			voteExtensionsValidator: newMockVoteExtensionsValidator(nil),
			reqTxs:                  txsVector("tx1", "tx2"),
			reqVoteExtensionsFn: func() []cmtabci.ExtendedVoteInfo {
				// Events the super-majority of validators are voting on.
				eventsSm := []bridgetypes.AssetsLockedEvent{
					mockEvent(201, "mezo12wsc0qgyfwwfj3wrlpgm9q3lmndl2m4qmm34dp", 100),
					mockEvent(202, "mezo1xqurxvvh8z2xpj6wltq0tajxm47xnv7q6rtvja", 200),
					mockEvent(203, "mezo1jcurf087xx9eqsnmr8lszralupcfln2vjz8ucg", 300),
				}

				// Events the minority of validators are voting on. They
				// differ in the amount of locked assets.
				eventsNonSm := []bridgetypes.AssetsLockedEvent{
					mockEvent(201, "mezo12wsc0qgyfwwfj3wrlpgm9q3lmndl2m4qmm34dp", 1000),
					mockEvent(202, "mezo1xqurxvvh8z2xpj6wltq0tajxm47xnv7q6rtvja", 2000),
					mockEvent(203, "mezo1jcurf087xx9eqsnmr8lszralupcfln2vjz8ucg", 3000),
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
			expectedInjectedTxFn: func(extendedCommitInfo []byte) *types.InjectedTx {
				return &types.InjectedTx{
					AssetsLockedEvents: []bridgetypes.AssetsLockedEvent{
						mockEvent(201, "mezo12wsc0qgyfwwfj3wrlpgm9q3lmndl2m4qmm34dp", 100),
						mockEvent(202, "mezo1xqurxvvh8z2xpj6wltq0tajxm47xnv7q6rtvja", 200),
						mockEvent(203, "mezo1jcurf087xx9eqsnmr8lszralupcfln2vjz8ucg", 300),
					},
					ExtendedCommitInfo: extendedCommitInfo,
				}
			},
			expectedChainTxs: txsVector("tx1", "tx2"),
			errContains:      "",
		},
		{
			name:     "proper canonical sequence - some votes on sequence outliers",
			valStore: newMockValidatorStore("val1Bridge", "val2Bridge", "val3Bridge", "val4Bridge"),
			voteExtensionDecomposerFn: func() *mockVoteExtensionDecomposer {
				return newMockVoteExtensionDecomposer().withReturnInputMode()
			},
			voteExtensionsValidator: newMockVoteExtensionsValidator(nil),
			reqTxs:                  txsVector("tx1", "tx2"),
			reqVoteExtensionsFn: func() []cmtabci.ExtendedVoteInfo {
				// Events the super-majority of validators are voting on.
				eventsSm := []bridgetypes.AssetsLockedEvent{
					mockEvent(201, "mezo12wsc0qgyfwwfj3wrlpgm9q3lmndl2m4qmm34dp", 100),
					mockEvent(202, "mezo1xqurxvvh8z2xpj6wltq0tajxm47xnv7q6rtvja", 200),
					mockEvent(203, "mezo1jcurf087xx9eqsnmr8lszralupcfln2vjz8ucg", 300),
				}

				// Outliers are events that some of the validators are voting on
				// but, they are not supported by the super-majority.
				outlier1 := mockEvent(200, "mezo1j0ghx6d9kmerxhgn5ahr2nahs6yfulea4te22c", 90)
				outlier2 := mockEvent(204, "mezo120mkxfvkx2t72quddqh92md2dp7csq4wgqux06", 400)

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
			expectedInjectedTxFn: func(extendedCommitInfo []byte) *types.InjectedTx {
				return &types.InjectedTx{
					AssetsLockedEvents: []bridgetypes.AssetsLockedEvent{
						mockEvent(201, "mezo12wsc0qgyfwwfj3wrlpgm9q3lmndl2m4qmm34dp", 100),
						mockEvent(202, "mezo1xqurxvvh8z2xpj6wltq0tajxm47xnv7q6rtvja", 200),
						mockEvent(203, "mezo1jcurf087xx9eqsnmr8lszralupcfln2vjz8ucg", 300),
					},
					ExtendedCommitInfo: extendedCommitInfo,
				}
			},
			expectedChainTxs: txsVector("tx1", "tx2"),
			errContains:      "",
		},
		{
			name:     "canonical sequence not starting directly after the current tip",
			valStore: newMockValidatorStore("val1Bridge", "val2Bridge", "val3Bridge", "val4Bridge"),
			voteExtensionDecomposerFn: func() *mockVoteExtensionDecomposer {
				return newMockVoteExtensionDecomposer().withReturnInputMode()
			},
			voteExtensionsValidator: newMockVoteExtensionsValidator(nil),
			reqTxs:                  txsVector("tx1", "tx2"),
			reqVoteExtensionsFn: func() []cmtabci.ExtendedVoteInfo {
				events := []bridgetypes.AssetsLockedEvent{
					// First event starts at 202 while the tip is 200.
					mockEvent(202, "mezo12wsc0qgyfwwfj3wrlpgm9q3lmndl2m4qmm34dp", 100),
					mockEvent(203, "mezo1xqurxvvh8z2xpj6wltq0tajxm47xnv7q6rtvja", 200),
					mockEvent(204, "mezo1jcurf087xx9eqsnmr8lszralupcfln2vjz8ucg", 300),
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
			expectedInjectedTxFn: func(_ []byte) *types.InjectedTx { return nil },
			expectedChainTxs:     txsVector("tx1", "tx2"),
			errContains:          "",
		},
		{
			name:     "empty canonical sequence",
			valStore: newMockValidatorStore("val1Bridge", "val2Bridge", "val3Bridge", "val4Bridge"),
			voteExtensionDecomposerFn: func() *mockVoteExtensionDecomposer {
				return newMockVoteExtensionDecomposer().withReturnInputMode()
			},
			voteExtensionsValidator: newMockVoteExtensionsValidator(nil),
			reqTxs:                  txsVector("tx1", "tx2"),
			reqVoteExtensionsFn: func() []cmtabci.ExtendedVoteInfo {
				return []cmtabci.ExtendedVoteInfo{
					mockVoteExtension("val1Bridge", 100, tmproto.BlockIDFlagCommit),
					mockVoteExtension("val2Bridge", 100, tmproto.BlockIDFlagCommit),
					mockVoteExtension("val3Bridge", 100, tmproto.BlockIDFlagCommit),
					mockVoteExtension("val4Bridge", 100, tmproto.BlockIDFlagCommit),
					mockVoteExtension("val5NonBridge", 100, tmproto.BlockIDFlagCommit),
					mockVoteExtension("val6NonBridge", 100, tmproto.BlockIDFlagCommit),
					mockVoteExtension("val7NonBridge", 100, tmproto.BlockIDFlagCommit),
					mockVoteExtension("val8NonBridge", 100, tmproto.BlockIDFlagCommit),
				}
			},
			expectedInjectedTxFn: func(_ []byte) *types.InjectedTx { return nil },
			expectedChainTxs:     txsVector("tx1", "tx2"),
			errContains:          "",
		},
		{
			name:     "no canonical sequence - split brain",
			valStore: newMockValidatorStore("val1Bridge", "val2Bridge", "val3Bridge", "val4Bridge"),
			voteExtensionDecomposerFn: func() *mockVoteExtensionDecomposer {
				return newMockVoteExtensionDecomposer().withReturnInputMode()
			},
			voteExtensionsValidator: newMockVoteExtensionsValidator(nil),
			reqTxs:                  txsVector("tx1", "tx2"),
			reqVoteExtensionsFn: func() []cmtabci.ExtendedVoteInfo {
				// Set of events the first half of validators are voting on.
				events1 := []bridgetypes.AssetsLockedEvent{
					mockEvent(201, "mezo12wsc0qgyfwwfj3wrlpgm9q3lmndl2m4qmm34dp", 100),
					mockEvent(202, "mezo1xqurxvvh8z2xpj6wltq0tajxm47xnv7q6rtvja", 200),
					mockEvent(203, "mezo1jcurf087xx9eqsnmr8lszralupcfln2vjz8ucg", 300),
				}

				// Set of events the second half of validators are voting on.
				events2 := []bridgetypes.AssetsLockedEvent{
					mockEvent(201, "mezo1j0ghx6d9kmerxhgn5ahr2nahs6yfulea4te22c", 1000),
					mockEvent(202, "mezo120mkxfvkx2t72quddqh92md2dp7csq4wgqux06", 2000),
					mockEvent(203, "mezo1dmr6mhh352vh9wa34xs0qxtr8thkqu39pw6x2p", 3000),
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
			expectedInjectedTxFn: func(_ []byte) *types.InjectedTx { return nil },
			expectedChainTxs:     txsVector("tx1", "tx2"),
			errContains:          "",
		},
		{
			name:     "no canonical sequence - bridge validators not supporting",
			valStore: newMockValidatorStore("val1Bridge", "val2Bridge", "val3Bridge", "val4Bridge"),
			voteExtensionDecomposerFn: func() *mockVoteExtensionDecomposer {
				return newMockVoteExtensionDecomposer().withReturnInputMode()
			},
			voteExtensionsValidator: newMockVoteExtensionsValidator(nil),
			reqTxs:                  txsVector("tx1", "tx2"),
			reqVoteExtensionsFn: func() []cmtabci.ExtendedVoteInfo {
				events := []bridgetypes.AssetsLockedEvent{
					mockEvent(201, "mezo12wsc0qgyfwwfj3wrlpgm9q3lmndl2m4qmm34dp", 100),
					mockEvent(202, "mezo1xqurxvvh8z2xpj6wltq0tajxm47xnv7q6rtvja", 200),
					mockEvent(203, "mezo1jcurf087xx9eqsnmr8lszralupcfln2vjz8ucg", 300),
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
			expectedInjectedTxFn: func(_ []byte) *types.InjectedTx { return nil },
			expectedChainTxs:     txsVector("tx1", "tx2"),
			errContains:          "",
		},
		{
			name:     "no canonical sequence - non-bridge validators not supporting",
			valStore: newMockValidatorStore("val1Bridge", "val2Bridge", "val3Bridge", "val4Bridge"),
			voteExtensionDecomposerFn: func() *mockVoteExtensionDecomposer {
				return newMockVoteExtensionDecomposer().withReturnInputMode()
			},
			voteExtensionsValidator: newMockVoteExtensionsValidator(nil),
			reqTxs:                  txsVector("tx1", "tx2"),
			reqVoteExtensionsFn: func() []cmtabci.ExtendedVoteInfo {
				events := []bridgetypes.AssetsLockedEvent{
					mockEvent(201, "mezo12wsc0qgyfwwfj3wrlpgm9q3lmndl2m4qmm34dp", 100),
					mockEvent(202, "mezo1xqurxvvh8z2xpj6wltq0tajxm47xnv7q6rtvja", 200),
					mockEvent(203, "mezo1jcurf087xx9eqsnmr8lszralupcfln2vjz8ucg", 300),
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
			expectedInjectedTxFn: func(_ []byte) *types.InjectedTx { return nil },
			expectedChainTxs:     txsVector("tx1", "tx2"),
			errContains:          "",
		},
		{
			name:     "no canonical sequence - bridge validators not existing",
			valStore: newMockValidatorStore(), // We do not set any bridge validators.
			voteExtensionDecomposerFn: func() *mockVoteExtensionDecomposer {
				return newMockVoteExtensionDecomposer().withReturnInputMode()
			},
			voteExtensionsValidator: newMockVoteExtensionsValidator(nil),
			reqTxs:                  txsVector("tx1", "tx2"),
			reqVoteExtensionsFn: func() []cmtabci.ExtendedVoteInfo {
				events := []bridgetypes.AssetsLockedEvent{
					mockEvent(201, "mezo12wsc0qgyfwwfj3wrlpgm9q3lmndl2m4qmm34dp", 100),
					mockEvent(202, "mezo1xqurxvvh8z2xpj6wltq0tajxm47xnv7q6rtvja", 200),
					mockEvent(203, "mezo1jcurf087xx9eqsnmr8lszralupcfln2vjz8ucg", 300),
				}

				return []cmtabci.ExtendedVoteInfo{
					mockVoteExtension("val5NonBridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val6NonBridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val7NonBridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val8NonBridge", 100, tmproto.BlockIDFlagCommit, events...),
				}
			},
			expectedInjectedTxFn: func(_ []byte) *types.InjectedTx { return nil },
			expectedChainTxs:     txsVector("tx1", "tx2"),
			errContains:          "",
		},
		{
			name:     "no canonical sequence - non-bridge validators not existing",
			valStore: newMockValidatorStore("val1Bridge", "val2Bridge", "val3Bridge", "val4Bridge"),
			voteExtensionDecomposerFn: func() *mockVoteExtensionDecomposer {
				return newMockVoteExtensionDecomposer().withReturnInputMode()
			},
			voteExtensionsValidator: newMockVoteExtensionsValidator(nil),
			reqTxs:                  txsVector("tx1", "tx2"),
			reqVoteExtensionsFn: func() []cmtabci.ExtendedVoteInfo {
				events := []bridgetypes.AssetsLockedEvent{
					mockEvent(201, "mezo12wsc0qgyfwwfj3wrlpgm9q3lmndl2m4qmm34dp", 100),
					mockEvent(202, "mezo1xqurxvvh8z2xpj6wltq0tajxm47xnv7q6rtvja", 200),
					mockEvent(203, "mezo1jcurf087xx9eqsnmr8lszralupcfln2vjz8ucg", 300),
				}

				return []cmtabci.ExtendedVoteInfo{
					mockVoteExtension("val1Bridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val2Bridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val3Bridge", 100, tmproto.BlockIDFlagCommit, events...),
					mockVoteExtension("val4Bridge", 100, tmproto.BlockIDFlagCommit, events...),
				}
			},
			expectedInjectedTxFn: func(_ []byte) *types.InjectedTx { return nil },
			expectedChainTxs:     txsVector("tx1", "tx2"),
			errContains:          "",
		},
		{
			name:     "no canonical sequence - too many VEs rejected - wrong block ID flag (unknown)",
			valStore: newMockValidatorStore("val1Bridge", "val2Bridge", "val3Bridge", "val4Bridge"),
			voteExtensionDecomposerFn: func() *mockVoteExtensionDecomposer {
				return newMockVoteExtensionDecomposer().withReturnInputMode()
			},
			voteExtensionsValidator: newMockVoteExtensionsValidator(nil),
			reqTxs:                  txsVector("tx1", "tx2"),
			reqVoteExtensionsFn: func() []cmtabci.ExtendedVoteInfo {
				events := []bridgetypes.AssetsLockedEvent{
					mockEvent(201, "mezo12wsc0qgyfwwfj3wrlpgm9q3lmndl2m4qmm34dp", 100),
					mockEvent(202, "mezo1xqurxvvh8z2xpj6wltq0tajxm47xnv7q6rtvja", 200),
					mockEvent(203, "mezo1jcurf087xx9eqsnmr8lszralupcfln2vjz8ucg", 300),
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
			expectedInjectedTxFn: func(_ []byte) *types.InjectedTx { return nil },
			expectedChainTxs:     txsVector("tx1", "tx2"),
			errContains:          "",
		},
		{
			name:     "no canonical sequence - too many VEs rejected - wrong block ID flag (absent)",
			valStore: newMockValidatorStore("val1Bridge", "val2Bridge", "val3Bridge", "val4Bridge"),
			voteExtensionDecomposerFn: func() *mockVoteExtensionDecomposer {
				return newMockVoteExtensionDecomposer().withReturnInputMode()
			},
			voteExtensionsValidator: newMockVoteExtensionsValidator(nil),
			reqTxs:                  txsVector("tx1", "tx2"),
			reqVoteExtensionsFn: func() []cmtabci.ExtendedVoteInfo {
				events := []bridgetypes.AssetsLockedEvent{
					mockEvent(201, "mezo12wsc0qgyfwwfj3wrlpgm9q3lmndl2m4qmm34dp", 100),
					mockEvent(202, "mezo1xqurxvvh8z2xpj6wltq0tajxm47xnv7q6rtvja", 200),
					mockEvent(203, "mezo1jcurf087xx9eqsnmr8lszralupcfln2vjz8ucg", 300),
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
			expectedInjectedTxFn: func(_ []byte) *types.InjectedTx { return nil },
			expectedChainTxs:     txsVector("tx1", "tx2"),
			errContains:          "",
		},
		{
			name:     "no canonical sequence - too many VEs rejected - wrong block ID flag (nil)",
			valStore: newMockValidatorStore("val1Bridge", "val2Bridge", "val3Bridge", "val4Bridge"),
			voteExtensionDecomposerFn: func() *mockVoteExtensionDecomposer {
				return newMockVoteExtensionDecomposer().withReturnInputMode()
			},
			voteExtensionsValidator: newMockVoteExtensionsValidator(nil),
			reqTxs:                  txsVector("tx1", "tx2"),
			reqVoteExtensionsFn: func() []cmtabci.ExtendedVoteInfo {
				events := []bridgetypes.AssetsLockedEvent{
					mockEvent(201, "mezo12wsc0qgyfwwfj3wrlpgm9q3lmndl2m4qmm34dp", 100),
					mockEvent(202, "mezo1xqurxvvh8z2xpj6wltq0tajxm47xnv7q6rtvja", 200),
					mockEvent(203, "mezo1jcurf087xx9eqsnmr8lszralupcfln2vjz8ucg", 300),
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
			expectedInjectedTxFn: func(_ []byte) *types.InjectedTx { return nil },
			expectedChainTxs:     txsVector("tx1", "tx2"),
			errContains:          "",
		},
		{
			name:     "no canonical sequence - too many VEs rejected - empty composite",
			valStore: newMockValidatorStore("val1Bridge", "val2Bridge", "val3Bridge", "val4Bridge"),
			voteExtensionDecomposerFn: func() *mockVoteExtensionDecomposer {
				return newMockVoteExtensionDecomposer().withReturnInputMode()
			},
			voteExtensionsValidator: newMockVoteExtensionsValidator(nil),
			reqTxs:                  txsVector("tx1", "tx2"),
			reqVoteExtensionsFn: func() []cmtabci.ExtendedVoteInfo {
				events := []bridgetypes.AssetsLockedEvent{
					mockEvent(201, "mezo12wsc0qgyfwwfj3wrlpgm9q3lmndl2m4qmm34dp", 100),
					mockEvent(202, "mezo1xqurxvvh8z2xpj6wltq0tajxm47xnv7q6rtvja", 200),
					mockEvent(203, "mezo1jcurf087xx9eqsnmr8lszralupcfln2vjz8ucg", 300),
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
			expectedInjectedTxFn: func(_ []byte) *types.InjectedTx { return nil },
			expectedChainTxs:     txsVector("tx1", "tx2"),
			errContains:          "",
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
			voteExtensionsValidator: newMockVoteExtensionsValidator(nil),
			reqTxs:                  txsVector("tx1", "tx2"),
			reqVoteExtensionsFn: func() []cmtabci.ExtendedVoteInfo {
				events := []bridgetypes.AssetsLockedEvent{
					mockEvent(201, "mezo12wsc0qgyfwwfj3wrlpgm9q3lmndl2m4qmm34dp", 100),
					mockEvent(202, "mezo1xqurxvvh8z2xpj6wltq0tajxm47xnv7q6rtvja", 200),
					mockEvent(203, "mezo1jcurf087xx9eqsnmr8lszralupcfln2vjz8ucg", 300),
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
			expectedInjectedTxFn: func(_ []byte) *types.InjectedTx { return nil },
			expectedChainTxs:     txsVector("tx1", "tx2"),
			errContains:          "",
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
			voteExtensionsValidator: newMockVoteExtensionsValidator(nil),
			reqTxs:                  txsVector("tx1", "tx2"),
			reqVoteExtensionsFn: func() []cmtabci.ExtendedVoteInfo {
				events := []bridgetypes.AssetsLockedEvent{
					mockEvent(201, "mezo12wsc0qgyfwwfj3wrlpgm9q3lmndl2m4qmm34dp", 100),
					mockEvent(202, "mezo1xqurxvvh8z2xpj6wltq0tajxm47xnv7q6rtvja", 200),
					mockEvent(203, "mezo1jcurf087xx9eqsnmr8lszralupcfln2vjz8ucg", 300),
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
			expectedInjectedTxFn: func(_ []byte) *types.InjectedTx { return nil },
			expectedChainTxs:     txsVector("tx1", "tx2"),
			errContains:          "",
		},
		{
			name:     "no canonical sequence - too many VEs rejected - unmarshaling error",
			valStore: newMockValidatorStore("val1Bridge", "val2Bridge", "val3Bridge", "val4Bridge"),
			voteExtensionDecomposerFn: func() *mockVoteExtensionDecomposer {
				return newMockVoteExtensionDecomposer().withReturnInputMode()
			},
			voteExtensionsValidator: newMockVoteExtensionsValidator(nil),
			reqTxs:                  txsVector("tx1", "tx2"),
			reqVoteExtensionsFn: func() []cmtabci.ExtendedVoteInfo {
				events := []bridgetypes.AssetsLockedEvent{
					mockEvent(201, "mezo12wsc0qgyfwwfj3wrlpgm9q3lmndl2m4qmm34dp", 100),
					mockEvent(202, "mezo1xqurxvvh8z2xpj6wltq0tajxm47xnv7q6rtvja", 200),
					mockEvent(203, "mezo1jcurf087xx9eqsnmr8lszralupcfln2vjz8ucg", 300),
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
			expectedInjectedTxFn: func(_ []byte) *types.InjectedTx { return nil },
			expectedChainTxs:     txsVector("tx1", "tx2"),
			errContains:          "",
		},
		{
			name:     "no canonical sequence - too many VEs rejected - empty sequence",
			valStore: newMockValidatorStore("val1Bridge", "val2Bridge", "val3Bridge", "val4Bridge"),
			voteExtensionDecomposerFn: func() *mockVoteExtensionDecomposer {
				return newMockVoteExtensionDecomposer().withReturnInputMode()
			},
			voteExtensionsValidator: newMockVoteExtensionsValidator(nil),
			reqTxs:                  txsVector("tx1", "tx2"),
			reqVoteExtensionsFn: func() []cmtabci.ExtendedVoteInfo {
				events := []bridgetypes.AssetsLockedEvent{
					mockEvent(201, "mezo12wsc0qgyfwwfj3wrlpgm9q3lmndl2m4qmm34dp", 100),
					mockEvent(202, "mezo1xqurxvvh8z2xpj6wltq0tajxm47xnv7q6rtvja", 200),
					mockEvent(203, "mezo1jcurf087xx9eqsnmr8lszralupcfln2vjz8ucg", 300),
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
			expectedInjectedTxFn: func(_ []byte) *types.InjectedTx { return nil },
			expectedChainTxs:     txsVector("tx1", "tx2"),
			errContains:          "",
		},
		{
			name:     "no canonical sequence - too many VEs rejected - sequence exceeding the limit",
			valStore: newMockValidatorStore("val1Bridge", "val2Bridge", "val3Bridge", "val4Bridge"),
			voteExtensionDecomposerFn: func() *mockVoteExtensionDecomposer {
				return newMockVoteExtensionDecomposer().withReturnInputMode()
			},
			voteExtensionsValidator: newMockVoteExtensionsValidator(nil),
			reqTxs:                  txsVector("tx1", "tx2"),
			reqVoteExtensionsFn: func() []cmtabci.ExtendedVoteInfo {
				events := []bridgetypes.AssetsLockedEvent{
					mockEvent(201, "mezo12wsc0qgyfwwfj3wrlpgm9q3lmndl2m4qmm34dp", 100),
					mockEvent(202, "mezo1xqurxvvh8z2xpj6wltq0tajxm47xnv7q6rtvja", 200),
					mockEvent(203, "mezo1jcurf087xx9eqsnmr8lszralupcfln2vjz8ucg", 300),
				}

				val1BridgeEvents := []bridgetypes.AssetsLockedEvent{
					events[0],
					events[1],
					events[2],
					mockEvent(204, "mezo1jcurf087xx9eqsnmr8lszralupcfln2vjz8ucg", 300),
					mockEvent(205, "mezo1jcurf087xx9eqsnmr8lszralupcfln2vjz8ucg", 300),
					mockEvent(206, "mezo1jcurf087xx9eqsnmr8lszralupcfln2vjz8ucg", 300),
					mockEvent(207, "mezo1jcurf087xx9eqsnmr8lszralupcfln2vjz8ucg", 300),
					mockEvent(208, "mezo1jcurf087xx9eqsnmr8lszralupcfln2vjz8ucg", 300),
					mockEvent(209, "mezo1jcurf087xx9eqsnmr8lszralupcfln2vjz8ucg", 300),
					mockEvent(210, "mezo1jcurf087xx9eqsnmr8lszralupcfln2vjz8ucg", 300),
					mockEvent(211, "mezo1jcurf087xx9eqsnmr8lszralupcfln2vjz8ucg", 300),
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
			expectedInjectedTxFn: func(_ []byte) *types.InjectedTx { return nil },
			expectedChainTxs:     txsVector("tx1", "tx2"),
			errContains:          "",
		},
		{
			name:     "no canonical sequence - too many VEs rejected - invalid sequence",
			valStore: newMockValidatorStore("val1Bridge", "val2Bridge", "val3Bridge", "val4Bridge"),
			voteExtensionDecomposerFn: func() *mockVoteExtensionDecomposer {
				return newMockVoteExtensionDecomposer().withReturnInputMode()
			},
			voteExtensionsValidator: newMockVoteExtensionsValidator(nil),
			reqTxs:                  txsVector("tx1", "tx2"),
			reqVoteExtensionsFn: func() []cmtabci.ExtendedVoteInfo {
				events := []bridgetypes.AssetsLockedEvent{
					mockEvent(201, "mezo12wsc0qgyfwwfj3wrlpgm9q3lmndl2m4qmm34dp", 100),
					mockEvent(202, "mezo1xqurxvvh8z2xpj6wltq0tajxm47xnv7q6rtvja", 200),
					mockEvent(203, "mezo1jcurf087xx9eqsnmr8lszralupcfln2vjz8ucg", 300),
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
			expectedInjectedTxFn: func(_ []byte) *types.InjectedTx { return nil },
			expectedChainTxs:     txsVector("tx1", "tx2"),
			errContains:          "",
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			s.SetupTest()

			valStore := test.valStore
			voteExtensionDecomposer := test.voteExtensionDecomposerFn()
			voteExtensionsValidator := test.voteExtensionsValidator

			s.handler = NewProposalHandler(
				s.logger,
				valStore,
				s.keeper,
				voteExtensionDecomposer.call,
				voteExtensionsValidator.call,
			)

			extendedCommitInfo := cmtabci.ExtendedCommitInfo{
				Round: 1, // Just an arbitrary value.
				Votes: test.reqVoteExtensionsFn(),
			}

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
) bridgetypes.AssetsLockedEvent {
	return bridgetypes.AssetsLockedEvent{
		Sequence:  sdkmath.NewInt(sequence),
		Recipient: recipient,
		Amount:    sdkmath.NewInt(amount),
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
