package types

import (
	"math/big"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/require"
)

func TestTxData_chainID(t *testing.T) {
	chainID := sdkmath.NewInt(1)

	testCases := []struct {
		msg        string
		data       TxData
		expChainID *big.Int
	}{
		{
			"access list tx", &AccessListTx{Accesses: AccessList{}, ChainID: &chainID}, big.NewInt(1),
		},
		{
			"access list tx, nil chain ID", &AccessListTx{Accesses: AccessList{}}, nil,
		},
		{
			"legacy tx, derived", &LegacyTx{}, nil,
		},
	}

	for _, tc := range testCases {
		chainID := tc.data.GetChainID()
		require.Equal(t, chainID, tc.expChainID, tc.msg)
	}
}

func TestTxData_DeriveChainID(t *testing.T) {
	bitLen64, ok := new(big.Int).SetString("0x8000000000000000", 0)
	require.True(t, ok)

	bitLen80, ok := new(big.Int).SetString("0x80000000000000000000", 0)
	require.True(t, ok)

	expBitLen80, ok := new(big.Int).SetString("302231454903657293676526", 0)
	require.True(t, ok)

	testCases := []struct {
		msg        string
		data       TxData
		expChainID *big.Int
	}{
		{
			"v = -1", &LegacyTx{V: big.NewInt(-1).Bytes()}, nil,
		},
		{
			"v = 0", &LegacyTx{V: big.NewInt(0).Bytes()}, nil,
		},
		{
			"v = 1", &LegacyTx{V: big.NewInt(1).Bytes()}, nil,
		},
		{
			"v = 27", &LegacyTx{V: big.NewInt(27).Bytes()}, new(big.Int),
		},
		{
			"v = 28", &LegacyTx{V: big.NewInt(28).Bytes()}, new(big.Int),
		},
		{
			"Ethereum mainnet", &LegacyTx{V: big.NewInt(37).Bytes()}, big.NewInt(1),
		},
		{
			"chain ID 31611", &LegacyTx{V: big.NewInt(63257).Bytes()}, big.NewInt(31611),
		},
		{
			"bit len 64", &LegacyTx{V: bitLen64.Bytes()}, big.NewInt(4611686018427387886),
		},
		{
			"bit len 80", &LegacyTx{V: bitLen80.Bytes()}, expBitLen80,
		},
		{
			"v = nil ", &LegacyTx{V: nil}, nil,
		},
	}

	for _, tc := range testCases {
		v, _, _ := tc.data.GetRawSignatureValues()

		chainID := DeriveChainID(v)
		require.Equal(t, tc.expChainID, chainID, tc.msg)
	}
}

func TestNewTxDataFromTx_SupportedTypes(t *testing.T) {
	to := common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")

	testCases := []struct {
		name   string
		txType uint8
		tx     *ethtypes.Transaction
	}{
		{
			"legacy tx (type 0)",
			ethtypes.LegacyTxType,
			ethtypes.NewTx(&ethtypes.LegacyTx{
				Nonce:    0,
				GasPrice: big.NewInt(1),
				Gas:      21000,
				To:       &to,
				Value:    big.NewInt(0),
			}),
		},
		{
			"access list tx (type 1)",
			ethtypes.AccessListTxType,
			ethtypes.NewTx(&ethtypes.AccessListTx{
				ChainID:  big.NewInt(1),
				Nonce:    0,
				GasPrice: big.NewInt(1),
				Gas:      21000,
				To:       &to,
				Value:    big.NewInt(0),
			}),
		},
		{
			"dynamic fee tx (type 2)",
			ethtypes.DynamicFeeTxType,
			ethtypes.NewTx(&ethtypes.DynamicFeeTx{
				ChainID:   big.NewInt(1),
				Nonce:     0,
				GasTipCap: big.NewInt(1),
				GasFeeCap: big.NewInt(1),
				Gas:       21000,
				To:        &to,
				Value:     big.NewInt(0),
			}),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.txType, tc.tx.Type())
			txData, err := NewTxDataFromTx(tc.tx)
			require.NoError(t, err)
			require.NotNil(t, txData)
		})
	}
}

func TestNewTxDataFromTx_RejectBlobTx(t *testing.T) {
	to := common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")

	blobTx := ethtypes.NewTx(&ethtypes.BlobTx{
		ChainID:    uint256.NewInt(1),
		Nonce:      0,
		GasTipCap:  uint256.NewInt(1),
		GasFeeCap:  uint256.NewInt(1),
		Gas:        21000,
		To:         to,
		Value:      uint256.NewInt(0),
		BlobFeeCap: uint256.NewInt(1),
		BlobHashes: []common.Hash{{}},
	})
	require.Equal(t, uint8(ethtypes.BlobTxType), blobTx.Type())

	txData, err := NewTxDataFromTx(blobTx)
	require.Nil(t, txData)
	require.Error(t, err)
	require.ErrorContains(t, err, "transaction type not supported")
	require.ErrorIs(t, err, ErrTxTypeNotSupported)
}
