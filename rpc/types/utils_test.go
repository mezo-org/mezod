package types_test

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/require"

	rpctypes "github.com/mezo-org/mezod/rpc/types"
)

func signedSetCodeTx(t *testing.T, chainID *big.Int) (*ethtypes.Transaction, []ethtypes.SetCodeAuthorization) {
	t.Helper()

	priv, err := crypto.GenerateKey()
	require.NoError(t, err)

	to := common.HexToAddress("0x1111111111111111111111111111111111111111")

	auth, err := ethtypes.SignSetCode(priv, ethtypes.SetCodeAuthorization{
		ChainID: *uint256.MustFromBig(chainID),
		Address: to,
		Nonce:   1,
	})
	require.NoError(t, err)

	auths := []ethtypes.SetCodeAuthorization{auth}

	inner := &ethtypes.SetCodeTx{
		ChainID:   uint256.MustFromBig(chainID),
		Nonce:     0,
		GasTipCap: uint256.NewInt(1),
		GasFeeCap: uint256.NewInt(10),
		Gas:       100_000,
		To:        to,
		Value:     uint256.NewInt(0),
		AuthList:  auths,
	}

	signer := ethtypes.NewPragueSigner(chainID)
	tx := ethtypes.MustSignNewTx(priv, signer, inner)
	return tx, auths
}

func TestNewRPCTransaction_SetCodeTx(t *testing.T) {
	chainID := big.NewInt(31611)
	tx, auths := signedSetCodeTx(t, chainID)

	t.Run("no base fee, no block hash", func(t *testing.T) {
		rpcTx, err := rpctypes.NewRPCTransaction(
			tx,
			common.Hash{},
			0,
			0,
			nil,
			chainID,
		)
		require.NoError(t, err)

		require.Equal(t, uint64(ethtypes.SetCodeTxType), uint64(rpcTx.Type))
		require.NotNil(t, rpcTx.Accesses)
		require.Equal(t, ethtypes.AccessList{}, *rpcTx.Accesses)
		require.Equal(t, auths, rpcTx.AuthorizationList)
		require.NotNil(t, rpcTx.ChainID)
		require.Equal(t, chainID, rpcTx.ChainID.ToInt())
		require.NotNil(t, rpcTx.GasFeeCap)
		require.NotNil(t, rpcTx.GasTipCap)
		// With no base fee / block hash, the effective gas price falls back to
		// GasFeeCap.
		require.Equal(t, tx.GasFeeCap(), rpcTx.GasPrice.ToInt())
	})

	t.Run("with base fee and block hash", func(t *testing.T) {
		baseFee := big.NewInt(3)
		blockHash := common.HexToHash("0xabc")

		rpcTx, err := rpctypes.NewRPCTransaction(
			tx,
			blockHash,
			1,
			0,
			baseFee,
			chainID,
		)
		require.NoError(t, err)

		require.Equal(t, uint64(ethtypes.SetCodeTxType), uint64(rpcTx.Type))
		require.NotNil(t, rpcTx.AuthorizationList)
		require.Len(t, rpcTx.AuthorizationList, len(auths))

		// price = min(tipCap+baseFee, feeCap)
		want := new(big.Int).Add(tx.GasTipCap(), baseFee)
		if want.Cmp(tx.GasFeeCap()) > 0 {
			want = tx.GasFeeCap()
		}
		require.Equal(t, want, rpcTx.GasPrice.ToInt())
	})

	t.Run("JSON marshals authorizationList", func(t *testing.T) {
		rpcTx, err := rpctypes.NewRPCTransaction(
			tx,
			common.Hash{},
			0,
			0,
			nil,
			chainID,
		)
		require.NoError(t, err)

		raw, err := json.Marshal(rpcTx)
		require.NoError(t, err)

		var decoded map[string]json.RawMessage
		require.NoError(t, json.Unmarshal(raw, &decoded))

		alRaw, ok := decoded["authorizationList"]
		require.True(t, ok, "authorizationList missing from JSON output")

		var alDecoded []ethtypes.SetCodeAuthorization
		require.NoError(t, json.Unmarshal(alRaw, &alDecoded))
		require.Equal(t, auths, alDecoded)

		// Empty access list should marshal as [] not null, mirroring how geth
		// surfaces it to clients.
		accRaw, ok := decoded["accessList"]
		require.True(t, ok, "accessList missing from JSON output")
		require.JSONEq(t, "[]", string(accRaw))
	})
}

func TestNewRPCTransaction_DynamicFeeTx(t *testing.T) {
	chainID := big.NewInt(31611)
	priv, err := crypto.GenerateKey()
	require.NoError(t, err)

	to := common.HexToAddress("0x1111111111111111111111111111111111111111")
	signer := ethtypes.NewPragueSigner(chainID)
	tipCap := big.NewInt(1)
	feeCap := big.NewInt(10)
	tx := ethtypes.MustSignNewTx(priv, signer, &ethtypes.DynamicFeeTx{
		ChainID:   chainID,
		Nonce:     0,
		GasTipCap: tipCap,
		GasFeeCap: feeCap,
		Gas:       21_000,
		To:        &to,
		Value:     big.NewInt(0),
	})

	t.Run("no base fee, no block hash", func(t *testing.T) {
		rpcTx, err := rpctypes.NewRPCTransaction(tx, common.Hash{}, 0, 0, nil, chainID)
		require.NoError(t, err)

		require.Equal(t, uint64(ethtypes.DynamicFeeTxType), uint64(rpcTx.Type))
		require.Nil(t, rpcTx.AuthorizationList)
		require.NotNil(t, rpcTx.Accesses)
		require.Equal(t, chainID, rpcTx.ChainID.ToInt())
		require.Equal(t, feeCap, rpcTx.GasFeeCap.ToInt())
		require.Equal(t, tipCap, rpcTx.GasTipCap.ToInt())
		// Effective price falls back to GasFeeCap when not mined.
		require.Equal(t, feeCap, rpcTx.GasPrice.ToInt())

		raw, err := json.Marshal(rpcTx)
		require.NoError(t, err)
		require.NotContains(t, string(raw), "authorizationList")
	})

	t.Run("with base fee and block hash", func(t *testing.T) {
		baseFee := big.NewInt(3)
		blockHash := common.HexToHash("0xabc")

		rpcTx, err := rpctypes.NewRPCTransaction(tx, blockHash, 1, 0, baseFee, chainID)
		require.NoError(t, err)

		// price = min(tipCap + baseFee, feeCap) = min(4, 10) = 4.
		require.Equal(t, big.NewInt(4), rpcTx.GasPrice.ToInt())
		require.Equal(t, feeCap, rpcTx.GasFeeCap.ToInt())
		require.Equal(t, tipCap, rpcTx.GasTipCap.ToInt())
	})

	t.Run("tip+baseFee exceeds feeCap", func(t *testing.T) {
		// baseFee 100 + tipCap 1 = 101 > feeCap 10 → price capped at feeCap.
		baseFee := big.NewInt(100)
		blockHash := common.HexToHash("0xabc")

		rpcTx, err := rpctypes.NewRPCTransaction(tx, blockHash, 1, 0, baseFee, chainID)
		require.NoError(t, err)
		require.Equal(t, feeCap, rpcTx.GasPrice.ToInt())
	})
}
