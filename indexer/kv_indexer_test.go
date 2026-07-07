package indexer_test

import (
	"math/big"
	"testing"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/simapp/params"
	abci "github.com/cometbft/cometbft/abci/types"
	tmtypes "github.com/cometbft/cometbft/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/holiman/uint256"
	"github.com/mezo-org/mezod/app"
	apptypes "github.com/mezo-org/mezod/app/abci/types"
	"github.com/mezo-org/mezod/crypto/ethsecp256k1"
	evmenc "github.com/mezo-org/mezod/encoding"
	"github.com/mezo-org/mezod/indexer"
	utiltx "github.com/mezo-org/mezod/testutil/tx"
	"github.com/mezo-org/mezod/utils"
	bridgeabcitypes "github.com/mezo-org/mezod/x/bridge/abci/types"
	bridgetypes "github.com/mezo-org/mezod/x/bridge/types"
	"github.com/mezo-org/mezod/x/evm/types"
	"github.com/stretchr/testify/require"
)

func TestKVIndexer(t *testing.T) {
	priv, err := ethsecp256k1.GenerateKey()
	require.NoError(t, err)
	from := common.BytesToAddress(priv.PubKey().Address().Bytes())
	signer := utiltx.NewSigner(priv)
	ethSigner := ethtypes.LatestSignerForChainID(nil)

	to := common.BigToAddress(big.NewInt(1))
	ethTxParams := types.EvmTxArgs{
		Nonce:    0,
		To:       &to,
		Amount:   big.NewInt(1000),
		GasLimit: 21000,
	}
	tx := types.NewTx(&ethTxParams)
	tx.From = from.Hex()
	require.NoError(t, tx.Sign(ethSigner, signer))
	txHash := tx.AsTransaction().Hash()

	encodingConfig := MakeEncodingConfig()
	clientCtx := client.Context{}.WithTxConfig(encodingConfig.TxConfig).WithCodec(encodingConfig.Codec)

	// build cosmos-sdk wrapper tx
	tmTx, err := tx.BuildTx(clientCtx.TxConfig.NewTxBuilder(), utils.BaseDenom)
	require.NoError(t, err)
	txBz, err := clientCtx.TxConfig.TxEncoder()(tmTx)
	require.NoError(t, err)

	// build an invalid wrapper tx
	builder := clientCtx.TxConfig.NewTxBuilder()
	require.NoError(t, builder.SetMsgs(tx))
	tmTx2 := builder.GetTx()
	txBz2, err := clientCtx.TxConfig.TxEncoder()(tmTx2)
	require.NoError(t, err)

	// build pseudo-transaction
	pseudoTx := buildPseudoTx(t)
	pseudoTxHash := common.BytesToHash(pseudoTx.Hash())

	testCases := []struct {
		name        string
		txHash      common.Hash
		block       *tmtypes.Block
		blockResult []*abci.ExecTxResult
		expSuccess  bool
	}{
		{
			"success, pseudo-transaction",
			pseudoTxHash,
			&tmtypes.Block{Header: tmtypes.Header{Height: 1}, Data: tmtypes.Data{Txs: []tmtypes.Tx{*pseudoTx}}},
			[]*abci.ExecTxResult{}, // not needed
			true,
		},
		{
			"success, format 1",
			txHash,
			&tmtypes.Block{Header: tmtypes.Header{Height: 1}, Data: tmtypes.Data{Txs: []tmtypes.Tx{txBz}}},
			[]*abci.ExecTxResult{
				{
					Code: 0,
					Events: []abci.Event{
						{Type: types.EventTypeEthereumTx, Attributes: []abci.EventAttribute{
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
			true,
		},
		{
			"success, format 2",
			txHash,
			&tmtypes.Block{Header: tmtypes.Header{Height: 1}, Data: tmtypes.Data{Txs: []tmtypes.Tx{txBz}}},
			[]*abci.ExecTxResult{
				{
					Code: 0,
					Events: []abci.Event{
						{Type: types.EventTypeEthereumTx, Attributes: []abci.EventAttribute{
							{Key: "ethereumTxHash", Value: txHash.Hex()},
							{Key: "txIndex", Value: "0"},
						}},
						{Type: types.EventTypeEthereumTx, Attributes: []abci.EventAttribute{
							{Key: "amount", Value: "1000"},
							{Key: "txGasUsed", Value: "21000"},
							{Key: "txHash", Value: "14A84ED06282645EFBF080E0B7ED80D8D8D6A36337668A12B5F229F81CDD3F57"},
							{Key: "recipient", Value: "0x775b87ef5D82ca211811C1a02CE0fE0CA3a455d7"},
						}},
					},
				},
			},
			true,
		},
		{
			"success, exceed block gas limit",
			txHash,
			&tmtypes.Block{Header: tmtypes.Header{Height: 1}, Data: tmtypes.Data{Txs: []tmtypes.Tx{txBz}}},
			[]*abci.ExecTxResult{
				{
					Code:   11,
					Log:    "out of gas in location: block gas meter; gasWanted: 21000",
					Events: []abci.Event{},
				},
			},
			true,
		},
		{
			"fail, failed eth tx",
			txHash,
			&tmtypes.Block{Header: tmtypes.Header{Height: 1}, Data: tmtypes.Data{Txs: []tmtypes.Tx{txBz}}},
			[]*abci.ExecTxResult{
				{
					Code:   15,
					Log:    "nonce mismatch",
					Events: []abci.Event{},
				},
			},
			false,
		},
		{
			"fail, invalid events",
			txHash,
			&tmtypes.Block{Header: tmtypes.Header{Height: 1}, Data: tmtypes.Data{Txs: []tmtypes.Tx{txBz}}},
			[]*abci.ExecTxResult{
				{
					Code:   0,
					Events: []abci.Event{},
				},
			},
			false,
		},
		{
			"fail, not eth tx",
			txHash,
			&tmtypes.Block{Header: tmtypes.Header{Height: 1}, Data: tmtypes.Data{Txs: []tmtypes.Tx{txBz2}}},
			[]*abci.ExecTxResult{
				{
					Code:   0,
					Events: []abci.Event{},
				},
			},
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db := dbm.NewMemDB()
			idxer := indexer.NewKVIndexer(db, log.NewNopLogger(), clientCtx)

			err = idxer.IndexBlock(tc.block, tc.blockResult)
			require.NoError(t, err)
			if !tc.expSuccess {
				first, err := idxer.FirstIndexedBlock()
				require.NoError(t, err)
				require.Equal(t, int64(-1), first)

				last, err := idxer.LastIndexedBlock()
				require.NoError(t, err)
				require.Equal(t, int64(-1), last)
			} else {
				first, err := idxer.FirstIndexedBlock()
				require.NoError(t, err)
				require.Equal(t, tc.block.Header.Height, first)

				last, err := idxer.LastIndexedBlock()
				require.NoError(t, err)
				require.Equal(t, tc.block.Header.Height, last)
				res1, err := idxer.GetByTxHash(tc.txHash)
				require.NoError(t, err)
				require.NotNil(t, res1)
				res2, err := idxer.GetByBlockAndIndex(1, 0)
				require.NoError(t, err)
				require.Equal(t, res1, res2)
			}
		})
	}
}

// MakeEncodingConfig creates the EncodingConfig
func MakeEncodingConfig() params.EncodingConfig {
	return evmenc.MakeConfig(app.ModuleBasics)
}

func buildPseudoTx(t *testing.T) *tmtypes.Tx {
	var bridgeTx bridgeabcitypes.InjectedTx
	bridgeTx.AssetsLockedEvents = []bridgetypes.AssetsLockedEvent{
		{
			Sequence:  sdkmath.NewInt(1),
			Recipient: "mezo1wengafav9m5yht926qmx4gr3d3rhxk50a5rzk8",
			Amount:    sdkmath.NewInt(1000000),
			Token:     common.HexToAddress("0x7d738d48b5c30f224aB86DaedE96CD95AB4854d9").Hex(),
		},
	}

	parts, err := bridgeTx.Marshal()
	if err != nil {
		t.Fatal(err)
	}

	var blockTx apptypes.InjectedTx
	blockTx.Parts = map[uint32][]byte{1: parts}

	tx, err := blockTx.Marshal()
	if err != nil {
		t.Fatal(err)
	}

	result := tmtypes.Tx(tx)
	return &result
}

// buildSignedSetCodeTx builds a signed Mezo SetCodeTx envelope (type 0x04)
// and returns its encoded bytes, its hash, and the client.Context whose
// TxConfig encoded it. The same context must be threaded into the indexer
// so codec round-trips line up.
//
// `to` and `target` are sentinel addresses; the auth tuple's V/R/S are
// bounds-only (canonical recovery is deferred to the keeper).
func buildSignedSetCodeTx(t *testing.T, chainID *big.Int, gas uint64) (tmtypes.Tx, common.Hash, client.Context) {
	t.Helper()
	priv, err := ethsecp256k1.GenerateKey()
	require.NoError(t, err)
	from := common.BytesToAddress(priv.PubKey().Address().Bytes())
	keyringSigner := utiltx.NewSigner(priv)
	ethSigner := ethtypes.LatestSignerForChainID(chainID)

	to := common.BigToAddress(big.NewInt(1))
	target := common.BigToAddress(big.NewInt(2))

	auth := ethtypes.SetCodeAuthorization{
		ChainID: *uint256.MustFromBig(chainID),
		Address: target,
		Nonce:   0,
		V:       1,
		R:       *uint256.NewInt(7),
		S:       *uint256.NewInt(11),
	}

	gethTx := ethtypes.NewTx(&ethtypes.SetCodeTx{
		ChainID:   uint256.MustFromBig(chainID),
		Nonce:     0,
		GasTipCap: uint256.NewInt(1),
		GasFeeCap: uint256.NewInt(1),
		Gas:       gas,
		To:        to,
		Value:     uint256.NewInt(0),
		Data:      []byte{},
		AuthList:  []ethtypes.SetCodeAuthorization{auth},
	})

	msg := &types.MsgEthereumTx{}
	require.NoError(t, msg.FromEthereumTx(gethTx))
	msg.From = from.Hex()
	require.NoError(t, msg.Sign(ethSigner, keyringSigner))

	encodingConfig := MakeEncodingConfig()
	clientCtx := client.Context{}.
		WithTxConfig(encodingConfig.TxConfig).
		WithCodec(encodingConfig.Codec)

	tmTx, err := msg.BuildTx(clientCtx.TxConfig.NewTxBuilder(), utils.BaseDenom)
	require.NoError(t, err)
	txBz, err := clientCtx.TxConfig.TxEncoder()(tmTx)
	require.NoError(t, err)

	return txBz, msg.AsTransaction().Hash(), clientCtx
}

// TestKVIndexerSetCodeTx exercises the custom KV indexer's MsgEthereumTx
// handling for an EIP-7702 SetCodeTx envelope (type 0x04). The table-driven
// TestKVIndexer above only covers a LegacyTx (types.NewTx has no SetCodeTx
// case); this test pins that the SetCodeTx path round-trips the same indexed
// fields (tx hash, height, eth tx index, gas-used, status) as any other tx
// type, so a regression that conditionalized the indexer on tx type would
// be caught.
//
// The test drives indexer.IndexBlock with a block containing a single signed
// SetCodeTx + a synthetic ExecTxResult emitting the standard
// `EventTypeEthereumTx` event the indexer parses.
func TestKVIndexerSetCodeTx(t *testing.T) {
	// Mezo SetCodeTx.Validate() restricts ChainID to {31611, 31612}.
	chainID := big.NewInt(31611)
	txBz, txHash, clientCtx := buildSignedSetCodeTx(t, chainID, 100_000)

	// The indexer treats the tx at block index 0 as a potential bridge
	// pseudo-tx and short-circuits to ParsePseudoTransaction. To exercise
	// the MsgEthereumTx codepath the SetCodeTx must sit at index >= 1; we
	// front-pad with a junk byte slice that ParsePseudoTransaction will
	// reject (no Parts[1]) but the indexer will then fall through to —
	// where TxDecoder fails and the entry is silently skipped (this
	// matches "fail, not eth tx" semantics in the legacy test). The
	// SetCodeTx then lands at index 1 and indexes through the eth path.
	junk := tmtypes.Tx([]byte{0x00})

	const blockHeight = int64(7)
	const wantTxIndex = uint32(1)
	const wantEthTxIndex = int32(0)
	const wantGasUsed = uint64(50_000)

	block := &tmtypes.Block{
		Header: tmtypes.Header{Height: blockHeight},
		Data:   tmtypes.Data{Txs: []tmtypes.Tx{junk, txBz}},
	}
	blockResult := []*abci.ExecTxResult{
		{Code: 0, Events: []abci.Event{}}, // for the junk tx at index 0
		{
			// rpctypes.ParseTxResult uses result.GasUsed as the
			// authoritative single-tx GasUsed (it unconditionally
			// overrides any per-event txGasUsed when len(p.Txs)==1).
			// Set it to the SetCodeTx's expected gas-used; the event
			// attribute below is set to the same value to keep the
			// shape identical to the format-1 emitter on the live
			// path.
			Code:    0,
			GasUsed: int64(wantGasUsed), //nolint:gosec
			Events: []abci.Event{
				{Type: types.EventTypeEthereumTx, Attributes: []abci.EventAttribute{
					{Key: "ethereumTxHash", Value: txHash.Hex()},
					{Key: "txIndex", Value: "0"},
					{Key: "amount", Value: "0"},
					{Key: "txGasUsed", Value: "50000"},
					{Key: "txHash", Value: ""},
					{Key: "recipient", Value: common.BigToAddress(big.NewInt(1)).Hex()},
				}},
			},
		},
	}

	db := dbm.NewMemDB()
	idxer := indexer.NewKVIndexer(db, log.NewNopLogger(), clientCtx)
	require.NoError(t, idxer.IndexBlock(block, blockResult))

	first, err := idxer.FirstIndexedBlock()
	require.NoError(t, err)
	require.Equal(t, blockHeight, first)

	last, err := idxer.LastIndexedBlock()
	require.NoError(t, err)
	require.Equal(t, blockHeight, last)

	// Hash lookup — proves the indexer keyed off the geth SetCodeTx hash,
	// not some legacy/dynamic-fee-only path.
	resByHash, err := idxer.GetByTxHash(txHash)
	require.NoError(t, err)
	require.NotNil(t, resByHash)

	// Block+index lookup — proves the secondary (block, eth-index) key
	// was written too; the legacy test pins the same invariant for
	// non-7702 txs.
	resByIndex, err := idxer.GetByBlockAndIndex(blockHeight, wantEthTxIndex)
	require.NoError(t, err)
	require.Equal(t, resByHash, resByIndex)

	// Field round-trip: the review-9 pin list is hash + height + index +
	// gas-used + status. txHash is implied by GetByTxHash succeeding; the
	// rest go field by field.
	require.Equal(t, blockHeight, resByHash.Height,
		"SetCodeTx indexed Height must round-trip")
	require.Equal(t, wantTxIndex, resByHash.TxIndex,
		"SetCodeTx indexed TxIndex (cosmos-level block tx index) must round-trip")
	require.Equal(t, wantEthTxIndex, resByHash.EthTxIndex,
		"SetCodeTx indexed EthTxIndex must round-trip")
	require.Equal(t, wantGasUsed, resByHash.GasUsed,
		"SetCodeTx indexed GasUsed must round-trip from the EventTypeEthereumTx event")
	require.False(t, resByHash.Failed,
		"SetCodeTx with Code=0 must index as success")
	require.Equal(t, wantGasUsed, resByHash.CumulativeGasUsed,
		"single-tx CumulativeGasUsed must equal GasUsed for the SetCodeTx")
}

// TestKVIndexerSetCodeTxExceedsBlockGasLimit pins the "exceeds block gas
// limit" branch (Code=11) for SetCodeTx: the indexer must still record the
// tx, with Failed=true and GasUsed set to the tx's gas limit (since events
// are not emitted in that branch). Mirrors the legacy test's "success,
// exceed block gas limit" case for SetCodeTx.
func TestKVIndexerSetCodeTxExceedsBlockGasLimit(t *testing.T) {
	const gasLimit = uint64(100_000)
	chainID := big.NewInt(31611)
	txBz, txHash, clientCtx := buildSignedSetCodeTx(t, chainID, gasLimit)

	junk := tmtypes.Tx([]byte{0x00})

	block := &tmtypes.Block{
		Header: tmtypes.Header{Height: 11},
		Data:   tmtypes.Data{Txs: []tmtypes.Tx{junk, txBz}},
	}
	blockResult := []*abci.ExecTxResult{
		{Code: 0, Events: []abci.Event{}},
		// The indexer's TxSuccessOrExceedsBlockGasLimit guard requires
		// the exact ExceedBlockGasLimitError prefix substring; "Code:
		// 11" alone is not enough.
		{Code: 11, Log: "out of gas in location: block gas meter; gasWanted: 100000", Events: []abci.Event{}},
	}

	db := dbm.NewMemDB()
	idxer := indexer.NewKVIndexer(db, log.NewNopLogger(), clientCtx)
	require.NoError(t, idxer.IndexBlock(block, blockResult))

	res, err := idxer.GetByTxHash(txHash)
	require.NoError(t, err)
	require.NotNil(t, res)

	require.Equal(t, int64(11), res.Height)
	require.Equal(t, uint32(1), res.TxIndex)
	require.Equal(t, int32(0), res.EthTxIndex)
	require.Equal(t, gasLimit, res.GasUsed,
		"on Code=11 the indexer must charge the tx's gas limit")
	require.True(t, res.Failed,
		"Code=11 (exceed block gas limit) must surface as Failed=true")
}
