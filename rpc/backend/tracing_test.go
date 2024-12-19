package backend

import (
	"fmt"

	sdkmath "cosmossdk.io/math"

	"cosmossdk.io/log"
	abci "github.com/cometbft/cometbft/abci/types"
	tmrpctypes "github.com/cometbft/cometbft/rpc/core/types"
	"github.com/cometbft/cometbft/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/crypto"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/mezo-org/mezod/crypto/ethsecp256k1"
	"github.com/mezo-org/mezod/indexer"
	"github.com/mezo-org/mezod/rpc/backend/mocks"
	bridgetypes "github.com/mezo-org/mezod/x/bridge/types"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
)

func (suite *BackendTestSuite) TestTraceTransaction() {
	msgEthereumTx, _ := suite.buildEthereumTx()
	msgEthereumTx2, _ := suite.buildEthereumTx()

	txHash := msgEthereumTx.AsTransaction().Hash()
	txHash2 := msgEthereumTx2.AsTransaction().Hash()

	priv, _ := ethsecp256k1.GenerateKey()
	from := common.BytesToAddress(priv.PubKey().Address().Bytes())

	queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
	RegisterParamsWithoutHeader(queryClient, 1)

	armor := crypto.EncryptArmorPrivKey(priv, "", "eth_secp256k1")
	_ = suite.backend.clientCtx.Keyring.ImportPrivKey("test_key", armor, "")

	ethSigner := ethtypes.LatestSigner(suite.backend.ChainConfig())

	txEncoder := suite.backend.clientCtx.TxConfig.TxEncoder()

	msgEthereumTx.From = from.String()
	_ = msgEthereumTx.Sign(ethSigner, suite.signer)

	tx, _ := msgEthereumTx.BuildTx(suite.backend.clientCtx.TxConfig.NewTxBuilder(), evmtypes.DefaultEVMDenom)
	txBz, _ := txEncoder(tx)

	msgEthereumTx2.From = from.String()
	_ = msgEthereumTx2.Sign(ethSigner, suite.signer)

	tx2, _ := msgEthereumTx.BuildTx(suite.backend.clientCtx.TxConfig.NewTxBuilder(), evmtypes.DefaultEVMDenom)
	txBz2, _ := txEncoder(tx2)

	// Prepare test data for pseudo-transaction.
	event := bridgetypes.AssetsLockedEvent{
		Sequence:  sdkmath.NewInt(1),
		Recipient: "mezo1wengafav9m5yht926qmx4gr3d3rhxk50a5rzk8",
		Amount:    sdkmath.NewInt(1000000),
	}
	pseudoTx, err := buildPseudoTx(event)
	suite.Require().NoError(err)
	pseudoTxBlock := &types.Block{Header: types.Header{Height: 1}, Data: types.Data{Txs: []types.Tx{*pseudoTx}}}
	pseudoTxTrace, err := buildPseudoTxTrace(event)
	suite.Require().NoError(err)
	pseudoTxHash := common.BytesToHash(pseudoTx.Hash())

	testCases := []struct {
		name          string
		registerMock  func()
		txHash        common.Hash
		block         *types.Block
		responseBlock []*abci.ExecTxResult
		expResult     interface{}
		expPass       bool
	}{
		{
			"fail - tx not found",
			func() {},
			txHash,
			&types.Block{Header: types.Header{Height: 1}, Data: types.Data{Txs: []types.Tx{}}},
			[]*abci.ExecTxResult{
				{
					Code: 0,
					Events: []abci.Event{
						{Type: evmtypes.EventTypeEthereumTx, Attributes: []abci.EventAttribute{
							{Key: "ethereumTxHash", Value: txHash.Hex()},
							{Key: "txIndex", Value: "0"},
							{Key: "amount", Value: "1000"},
							{Key: "txGasUsed", Value: "21000"},
							{Key: "txHash", Value: ""},
							{Key: "recipient", Value: "0x775b87ef5D82ca211811C1a02CE0fE0CA3a455d7"},
						}},
					},
				},
			},
			nil,
			false,
		},
		{
			"fail - block not found",
			func() {
				// var header metadata.MD
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				RegisterBlockError(client, 1)
			},
			txHash,
			&types.Block{Header: types.Header{Height: 1}, Data: types.Data{Txs: []types.Tx{txBz}}},
			[]*abci.ExecTxResult{
				{
					Code: 0,
					Events: []abci.Event{
						{Type: evmtypes.EventTypeEthereumTx, Attributes: []abci.EventAttribute{
							{Key: "ethereumTxHash", Value: txHash.Hex()},
							{Key: "txIndex", Value: "0"},
							{Key: "amount", Value: "1000"},
							{Key: "txGasUsed", Value: "21000"},
							{Key: "txHash", Value: ""},
							{Key: "recipient", Value: "0x775b87ef5D82ca211811C1a02CE0fE0CA3a455d7"},
						}},
					},
				},
			},
			map[string]interface{}{"test": "hello"},
			false,
		},
		{
			"pass - transaction found in a block with multiple transactions",
			func() {
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				_, err := RegisterBlockMultipleTxs(client, 1, []types.Tx{txBz, txBz2})
				suite.Require().NoError(err)
				RegisterTraceTransactionWithPredecessors(queryClient, msgEthereumTx, []*evmtypes.MsgEthereumTx{msgEthereumTx})
			},
			txHash,
			&types.Block{Header: types.Header{Height: 1, ChainID: ChainID}, Data: types.Data{Txs: []types.Tx{txBz, txBz2}}},
			[]*abci.ExecTxResult{
				{
					Code: 0,
					Events: []abci.Event{
						{Type: evmtypes.EventTypeEthereumTx, Attributes: []abci.EventAttribute{
							{Key: "ethereumTxHash", Value: txHash.Hex()},
							{Key: "txIndex", Value: "0"},
							{Key: "amount", Value: "1000"},
							{Key: "txGasUsed", Value: "21000"},
							{Key: "txHash", Value: ""},
							{Key: "recipient", Value: "0x775b87ef5D82ca211811C1a02CE0fE0CA3a455d7"},
						}},
					},
				},
				{
					Code: 0,
					Events: []abci.Event{
						{Type: evmtypes.EventTypeEthereumTx, Attributes: []abci.EventAttribute{
							{Key: "ethereumTxHash", Value: txHash2.Hex()},
							{Key: "txIndex", Value: "1"},
							{Key: "amount", Value: "1000"},
							{Key: "txGasUsed", Value: "21000"},
							{Key: "txHash", Value: ""},
							{Key: "recipient", Value: "0x775b87ef5D82ca211811C1a02CE0fE0CA3a455d7"},
						}},
					},
				},
			},
			map[string]interface{}{"test": "hello"},
			true,
		},
		{
			"pass - transaction found",
			func() {
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				_, err := RegisterBlock(client, 1, txBz)
				suite.Require().NoError(err)
				RegisterTraceTransaction(queryClient, msgEthereumTx)
			},
			txHash,
			&types.Block{Header: types.Header{Height: 1}, Data: types.Data{Txs: []types.Tx{txBz}}},
			[]*abci.ExecTxResult{
				{
					Code: 0,
					Events: []abci.Event{
						{Type: evmtypes.EventTypeEthereumTx, Attributes: []abci.EventAttribute{
							{Key: "ethereumTxHash", Value: txHash.Hex()},
							{Key: "txIndex", Value: "0"},
							{Key: "amount", Value: "1000"},
							{Key: "txGasUsed", Value: "21000"},
							{Key: "txHash", Value: ""},
							{Key: "recipient", Value: "0x775b87ef5D82ca211811C1a02CE0fE0CA3a455d7"},
						}},
					},
				},
			},
			map[string]interface{}{"test": "hello"},
			true,
		},
		{
			"pass - pseudo-transaction found",
			func() {
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				_, err := RegisterBlock(client, 1, *pseudoTx)
				suite.Require().NoError(err)
			},
			pseudoTxHash,
			pseudoTxBlock,
			[]*abci.ExecTxResult{},
			pseudoTxTrace,
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("case %s", tc.name), func() {
			suite.SetupTest() // reset test and queries
			tc.registerMock()

			db := dbm.NewMemDB()
			suite.backend.indexer = indexer.NewKVIndexer(db, log.NewNopLogger(), suite.backend.clientCtx)

			err := suite.backend.indexer.IndexBlock(tc.block, tc.responseBlock)
			suite.Require().NoError(err)
			txResult, err := suite.backend.TraceTransaction(tc.txHash, nil)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(tc.expResult, txResult)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *BackendTestSuite) TestTraceBlock() {
	msgEthTx, bz := suite.buildEthereumTx()
	emptyBlock := types.MakeBlock(1, []types.Tx{}, nil, nil)
	emptyBlock.ChainID = ChainID
	filledBlock := types.MakeBlock(1, []types.Tx{bz}, nil, nil)
	filledBlock.ChainID = ChainID
	resBlockEmpty := tmrpctypes.ResultBlock{Block: emptyBlock, BlockID: emptyBlock.LastBlockID}
	resBlockFilled := tmrpctypes.ResultBlock{Block: filledBlock, BlockID: filledBlock.LastBlockID}

	// Prepare test data for pseudo-transaction.
	event := bridgetypes.AssetsLockedEvent{
		Sequence:  sdkmath.NewInt(1),
		Recipient: "mezo1wengafav9m5yht926qmx4gr3d3rhxk50a5rzk8",
		Amount:    sdkmath.NewInt(1000000),
	}
	pseudoTx, err := buildPseudoTx(event)
	suite.Require().NoError(err)
	pseudoTxBlock := types.MakeBlock(1, []types.Tx{*pseudoTx}, nil, nil)
	pseudoTxBlock.ChainID = ChainID
	resBlockPseudoTx := tmrpctypes.ResultBlock{Block: pseudoTxBlock, BlockID: pseudoTxBlock.LastBlockID}

	testCases := []struct {
		name            string
		registerMock    func()
		expTraceResults []*evmtypes.TxTraceResult
		resBlock        *tmrpctypes.ResultBlock
		config          *evmtypes.TraceConfig
		expPass         bool
	}{
		{
			"pass - no transaction returning empty array",
			func() {},
			[]*evmtypes.TxTraceResult{},
			&resBlockEmpty,
			&evmtypes.TraceConfig{},
			true,
		},
		{
			"fail - cannot unmarshal data",
			func() {
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
				RegisterTraceBlock(queryClient, []*evmtypes.MsgEthereumTx{msgEthTx}, []byte(`{"test": "hello"}`))
			},
			[]*evmtypes.TxTraceResult{},
			&resBlockFilled,
			&evmtypes.TraceConfig{},
			false,
		},
		{
			"pass - pseudo-transaction",
			func() {
				db := dbm.NewMemDB()
				suite.backend.indexer = indexer.NewKVIndexer(db, log.NewNopLogger(), suite.backend.clientCtx)
				err := suite.backend.indexer.IndexBlock(pseudoTxBlock, []*abci.ExecTxResult{})
				suite.Require().NoError(err)
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
				RegisterTraceBlock(queryClient, nil, []byte(`[]`))
			},
			[]*evmtypes.TxTraceResult{
				{
					Result: map[string]interface{}{
						"failed":      false,
						"gas":         0,
						"returnValue": "",
						"structLogs":  []interface{}{},
					},
				},
			},
			&resBlockPseudoTx,
			&evmtypes.TraceConfig{},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("case %s", tc.name), func() {
			suite.SetupTest() // reset test and queries
			tc.registerMock()

			traceResults, err := suite.backend.TraceBlock(1, tc.config, tc.resBlock)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(tc.expTraceResults, traceResults)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}
