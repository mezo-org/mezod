package types_test

import (
	"crypto/ecdsa"
	"errors"
	"math/big"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/require"

	utiltx "github.com/mezo-org/mezod/testutil/tx"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
)

var (
	mezoChainID    = sdkmath.NewInt(31612)
	bigMezoChainID = mezoChainID.BigInt()
)

func newTestAuth(t *testing.T, chainID *big.Int, addr common.Address, nonce uint64) ethtypes.SetCodeAuthorization {
	t.Helper()
	return ethtypes.SetCodeAuthorization{
		ChainID: *uint256.MustFromBig(chainID),
		Address: addr,
		Nonce:   nonce,
		V:       1,
		R:       *uint256.NewInt(7),
		S:       *uint256.NewInt(11),
	}
}

func newSignedTestAuth(
	t *testing.T,
	chainID *big.Int,
	addr common.Address,
	nonce uint64,
	priv *ecdsa.PrivateKey,
) ethtypes.SetCodeAuthorization {
	t.Helper()

	auth := ethtypes.SetCodeAuthorization{
		ChainID: *uint256.MustFromBig(chainID),
		Address: addr,
		Nonce:   nonce,
	}
	signed, err := ethtypes.SignSetCode(priv, auth)
	require.NoError(t, err)
	return signed
}

func TestNewSetCodeTx(t *testing.T) {
	to := utiltx.GenerateAddress()
	auth := newTestAuth(t, big.NewInt(31611), to, 1)

	tx := ethtypes.NewTx(&ethtypes.SetCodeTx{
		ChainID:   uint256.MustFromBig(bigMezoChainID),
		Nonce:     1,
		GasTipCap: uint256.NewInt(1),
		GasFeeCap: uint256.NewInt(1),
		Gas:       100,
		To:        to,
		Value:     uint256.NewInt(1),
		Data:      []byte("data"),
		AuthList:  []ethtypes.SetCodeAuthorization{auth},
	})

	out, err := evmtypes.NewSetCodeTx(tx)
	require.NoError(t, err)
	require.NotNil(t, out)
	require.Equal(t, uint8(ethtypes.SetCodeTxType), out.TxType())
}

func TestSetCodeTxAsEthereumData(t *testing.T) {
	to := utiltx.GenerateAddress()
	priv, err := crypto.GenerateKey()
	require.NoError(t, err)

	auth := newSignedTestAuth(t, big.NewInt(31611), to, 1, priv)

	original := &ethtypes.SetCodeTx{
		ChainID:    uint256.MustFromBig(bigMezoChainID),
		Nonce:      7,
		GasTipCap:  uint256.NewInt(2),
		GasFeeCap:  uint256.NewInt(5),
		Gas:        21000,
		To:         to,
		Value:      uint256.NewInt(123),
		Data:       []byte("payload"),
		AccessList: ethtypes.AccessList{{Address: to, StorageKeys: []common.Hash{{}}}},
		AuthList:   []ethtypes.SetCodeAuthorization{auth},
	}

	tx := ethtypes.NewTx(original)

	cosmosTx, err := evmtypes.NewSetCodeTx(tx)
	require.NoError(t, err)

	out := cosmosTx.AsEthereumData()
	resTx := ethtypes.NewTx(out)

	require.Equal(t, original.Nonce, resTx.Nonce())
	require.Equal(t, original.Data, resTx.Data())
	require.Equal(t, original.Gas, resTx.Gas())
	require.Equal(t, original.Value.ToBig(), resTx.Value())
	require.Equal(t, original.AccessList, resTx.AccessList())
	require.Equal(t, &original.To, resTx.To())
	require.Equal(t, original.AuthList, resTx.SetCodeAuthorizations())
	require.Equal(t, bigMezoChainID, resTx.ChainId())
}

func TestSetCodeTxCopy(t *testing.T) {
	t.Run("zero value", func(t *testing.T) {
		tx := &evmtypes.SetCodeTx{}
		cp := tx.Copy().(*evmtypes.SetCodeTx)
		require.NotSame(t, tx, cp)
		require.Nil(t, cp.AuthList)
	})

	t.Run("empty (non-nil) AuthList preserved as empty", func(t *testing.T) {
		tx := &evmtypes.SetCodeTx{
			AuthList: evmtypes.AuthorizationList{},
		}
		cp := tx.Copy().(*evmtypes.SetCodeTx)
		require.NotNil(t, cp.AuthList)
		require.Len(t, cp.AuthList, 0)
	})

	t.Run("populated tx — mutating original does not affect copy", func(t *testing.T) {
		auth := evmtypes.SetCodeAuthorization{
			Address: common.HexToAddress("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa").Hex(),
			V:       []byte{0x01},
			R:       []byte{0x02},
			S:       []byte{0x03},
		}

		tx := &evmtypes.SetCodeTx{
			Data:     []byte{0xde, 0xad},
			V:        []byte{0x10},
			R:        []byte{0x20},
			S:        []byte{0x30},
			AuthList: evmtypes.AuthorizationList{auth},
		}

		cp := tx.Copy().(*evmtypes.SetCodeTx)

		tx.Data[0] = 0
		tx.V[0] = 0
		tx.R[0] = 0
		tx.S[0] = 0
		tx.AuthList[0].V[0] = 0
		tx.AuthList[0].R[0] = 0
		tx.AuthList[0].S[0] = 0

		require.Equal(t, byte(0xde), cp.Data[0])
		require.Equal(t, byte(0x10), cp.V[0])
		require.Equal(t, byte(0x20), cp.R[0])
		require.Equal(t, byte(0x30), cp.S[0])
		require.Equal(t, byte(0x01), cp.AuthList[0].V[0])
		require.Equal(t, byte(0x02), cp.AuthList[0].R[0])
		require.Equal(t, byte(0x03), cp.AuthList[0].S[0])
	})
}

func TestSetCodeTxValidate(t *testing.T) {
	hexAddr := utiltx.GenerateAddress().Hex()
	innerChainID := sdkmath.NewInt(31611)

	validAuth := func() evmtypes.SetCodeAuthorization {
		return evmtypes.SetCodeAuthorization{
			ChainID: &innerChainID,
			Address: hexAddr,
			Nonce:   1,
			V:       []byte{1},
			R:       big.NewInt(7).Bytes(),
			S:       big.NewInt(11).Bytes(),
		}
	}

	overflowBytes := new(big.Int).Exp(big.NewInt(2), big.NewInt(256), nil).Bytes()

	zero := sdkmath.ZeroInt()
	minusOne := sdkmath.NewInt(-1)

	cases := []struct {
		name     string
		mutate   func(tx *evmtypes.SetCodeTx)
		wantErr  bool
		errMatch string
	}{
		{
			name:     "happy path with one auth",
			mutate:   nil,
			wantErr:  false,
			errMatch: "",
		},
		{
			name: "happy path with five auths",
			mutate: func(tx *evmtypes.SetCodeTx) {
				tx.AuthList = evmtypes.AuthorizationList{
					validAuth(), validAuth(), validAuth(), validAuth(), validAuth(),
				}
			},
		},
		{
			name: "to is empty rejects with ErrSetCodeMissingTo",
			mutate: func(tx *evmtypes.SetCodeTx) {
				tx.To = ""
			},
			wantErr:  true,
			errMatch: evmtypes.ErrSetCodeMissingTo.Error(),
		},
		{
			name: "to is zero address accepted",
			mutate: func(tx *evmtypes.SetCodeTx) {
				tx.To = (common.Address{}).Hex()
			},
		},
		{
			name: "to is malformed-but-non-empty rejected by ValidateAddress",
			mutate: func(tx *evmtypes.SetCodeTx) {
				tx.To = "not-a-hex-address"
			},
			wantErr:  true,
			errMatch: "invalid to address",
		},
		{
			name: "nil GasTipCap rejected",
			mutate: func(tx *evmtypes.SetCodeTx) {
				tx.GasTipCap = nil
			},
			wantErr:  true,
			errMatch: "gas tip cap cannot nil",
		},
		{
			name: "nil GasFeeCap rejected",
			mutate: func(tx *evmtypes.SetCodeTx) {
				tx.GasFeeCap = nil
			},
			wantErr:  true,
			errMatch: "gas fee cap cannot nil",
		},
		{
			name: "nil ChainID rejected",
			mutate: func(tx *evmtypes.SetCodeTx) {
				tx.ChainID = nil
			},
			wantErr:  true,
			errMatch: "chain ID must be present",
		},
		{
			name: "nil Amount accepted (treated as zero, mirrors DynamicFeeTx)",
			mutate: func(tx *evmtypes.SetCodeTx) {
				tx.Amount = nil
			},
		},
		{
			name: "outer R out of int256 bounds rejected",
			mutate: func(tx *evmtypes.SetCodeTx) {
				tx.R = overflowBytes
			},
			wantErr:  true,
			errMatch: "V, R or S out of bound",
		},
		{
			name: "outer S out of int256 bounds rejected",
			mutate: func(tx *evmtypes.SetCodeTx) {
				tx.S = overflowBytes
			},
			wantErr:  true,
			errMatch: "V, R or S out of bound",
		},
		{
			name: "outer V out of int256 bounds rejected",
			mutate: func(tx *evmtypes.SetCodeTx) {
				tx.V = overflowBytes
			},
			wantErr:  true,
			errMatch: "V, R or S out of bound",
		},
		{
			name: "auth list empty rejects with ErrSetCodeEmptyAuthList",
			mutate: func(tx *evmtypes.SetCodeTx) {
				tx.AuthList = nil
			},
			wantErr:  true,
			errMatch: evmtypes.ErrSetCodeEmptyAuthList.Error(),
		},
		{
			name: "negative gas tip cap",
			mutate: func(tx *evmtypes.SetCodeTx) {
				tx.GasTipCap = &minusOne
			},
			wantErr:  true,
			errMatch: "gas tip cap cannot be negative",
		},
		{
			name: "negative gas fee cap",
			mutate: func(tx *evmtypes.SetCodeTx) {
				tx.GasFeeCap = &minusOne
				tx.GasTipCap = &zero
			},
			wantErr:  true,
			errMatch: "gas fee cap cannot be negative",
		},
		{
			name: "feecap < tipcap",
			mutate: func(tx *evmtypes.SetCodeTx) {
				high := sdkmath.NewInt(1000)
				low := sdkmath.NewInt(1)
				tx.GasTipCap = &high
				tx.GasFeeCap = &low
			},
			wantErr:  true,
			errMatch: "max priority fee per gas higher than max fee per gas",
		},
		{
			name: "negative amount",
			mutate: func(tx *evmtypes.SetCodeTx) {
				tx.Amount = &minusOne
			},
			wantErr:  true,
			errMatch: "amount cannot be negative",
		},
		{
			name: "non-Mezo chain id (1) rejected",
			mutate: func(tx *evmtypes.SetCodeTx) {
				cid := sdkmath.NewInt(1)
				tx.ChainID = &cid
			},
			wantErr:  true,
			errMatch: "chain ID must be 31611 or 31612",
		},
		{
			name: "non-Mezo chain id (31610) rejected",
			mutate: func(tx *evmtypes.SetCodeTx) {
				cid := sdkmath.NewInt(31610)
				tx.ChainID = &cid
			},
			wantErr:  true,
			errMatch: "chain ID must be 31611 or 31612",
		},
		{
			name: "non-Mezo chain id (31613) rejected",
			mutate: func(tx *evmtypes.SetCodeTx) {
				cid := sdkmath.NewInt(31613)
				tx.ChainID = &cid
			},
			wantErr:  true,
			errMatch: "chain ID must be 31611 or 31612",
		},
		{
			name: "auth with malformed address",
			mutate: func(tx *evmtypes.SetCodeTx) {
				bad := validAuth()
				bad.Address = "not-a-hex-address"
				tx.AuthList = evmtypes.AuthorizationList{bad}
			},
			wantErr:  true,
			errMatch: "authorization[0]",
		},
		{
			name: "auth with R out of int256 bounds rejected",
			mutate: func(tx *evmtypes.SetCodeTx) {
				bad := validAuth()
				bad.R = overflowBytes
				tx.AuthList = evmtypes.AuthorizationList{bad}
			},
			wantErr:  true,
			errMatch: "authorization[0] V, R or S out of bound",
		},
		{
			name: "auth with S out of int256 bounds rejected",
			mutate: func(tx *evmtypes.SetCodeTx) {
				bad := validAuth()
				bad.S = overflowBytes
				tx.AuthList = evmtypes.AuthorizationList{bad}
			},
			wantErr:  true,
			errMatch: "authorization[0] V, R or S out of bound",
		},
		{
			name: "auth with V out of int256 bounds rejected",
			mutate: func(tx *evmtypes.SetCodeTx) {
				bad := validAuth()
				bad.V = overflowBytes
				tx.AuthList = evmtypes.AuthorizationList{bad}
			},
			wantErr:  true,
			errMatch: "authorization[0] V, R or S out of bound",
		},
		{
			name: "auth with nil chain id",
			mutate: func(tx *evmtypes.SetCodeTx) {
				bad := validAuth()
				bad.ChainID = nil
				tx.AuthList = evmtypes.AuthorizationList{bad}
			},
			wantErr:  true,
			errMatch: "chain ID cannot be nil",
		},
		{
			name: "auth with zero chain id accepted (any-chain sentinel)",
			mutate: func(tx *evmtypes.SetCodeTx) {
				zeroAuth := validAuth()
				z := sdkmath.ZeroInt()
				zeroAuth.ChainID = &z
				tx.AuthList = evmtypes.AuthorizationList{zeroAuth}
			},
		},
		{
			name: "auth with non-Mezo, non-zero chain id accepted (apply-time skip per EIP-7702)",
			mutate: func(tx *evmtypes.SetCodeTx) {
				ok := validAuth()
				cid := sdkmath.NewInt(99999)
				ok.ChainID = &cid
				tx.AuthList = evmtypes.AuthorizationList{ok}
			},
		},
		{
			name: "auth with V[0] == 0 accepted",
			mutate: func(tx *evmtypes.SetCodeTx) {
				ok := validAuth()
				ok.V = []byte{0x00}
				tx.AuthList = evmtypes.AuthorizationList{ok}
			},
		},
		{
			name: "auth with V[0] > 1 accepted (canonical check deferred to keeper)",
			mutate: func(tx *evmtypes.SetCodeTx) {
				ok := validAuth()
				ok.V = []byte{0x02}
				tx.AuthList = evmtypes.AuthorizationList{ok}
			},
		},
		{
			name: "auth with R == 0 accepted (canonical check deferred to keeper)",
			mutate: func(tx *evmtypes.SetCodeTx) {
				ok := validAuth()
				ok.R = []byte{}
				tx.AuthList = evmtypes.AuthorizationList{ok}
			},
		},
		{
			name: "auth with S == 0 accepted (canonical check deferred to keeper)",
			mutate: func(tx *evmtypes.SetCodeTx) {
				ok := validAuth()
				ok.S = []byte{}
				tx.AuthList = evmtypes.AuthorizationList{ok}
			},
		},
	}

	build := func() evmtypes.SetCodeTx {
		amount := sdkmath.NewInt(1)
		return evmtypes.SetCodeTx{
			ChainID:   &mezoChainID,
			GasTipCap: &mezoChainID,
			GasFeeCap: &mezoChainID,
			Amount:    &amount,
			To:        hexAddr,
			AuthList:  evmtypes.AuthorizationList{validAuth()},
		}
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tx := build()
			if tc.mutate != nil {
				tc.mutate(&tx)
			}
			err := tx.Validate()
			if tc.wantErr {
				require.Error(t, err)
				if tc.errMatch != "" {
					require.Contains(t, err.Error(), tc.errMatch)
				}
				return
			}
			require.NoError(t, err)
		})
	}

	t.Run("auth R/S overflow wraps ErrInvalidSigner", func(t *testing.T) {
		tx := build()
		bad := validAuth()
		bad.R = overflowBytes
		tx.AuthList = evmtypes.AuthorizationList{bad}
		err := tx.Validate()
		require.Error(t, err)
		require.True(t, errors.Is(err, evmtypes.ErrInvalidSigner),
			"expected ErrInvalidSigner, got %v", err)
		require.False(t, errors.Is(err, evmtypes.ErrInvalidAmount),
			"R/S overflow must not surface as ErrInvalidAmount")
	})

	t.Run("outer V/R/S out-of-bound wraps ErrInvalidSigner", func(t *testing.T) {
		tx := build()
		tx.R = overflowBytes
		err := tx.Validate()
		require.Error(t, err)
		require.True(t, errors.Is(err, evmtypes.ErrInvalidSigner),
			"expected ErrInvalidSigner, got %v", err)
	})
}

func TestSetCodeTxGetAuthorizationList(t *testing.T) {
	to := utiltx.GenerateAddress()
	priv, err := crypto.GenerateKey()
	require.NoError(t, err)

	auths := []ethtypes.SetCodeAuthorization{
		newSignedTestAuth(t, big.NewInt(31611), to, 1, priv),
		newSignedTestAuth(t, big.NewInt(31611), to, 2, priv),
	}

	tx := ethtypes.NewTx(&ethtypes.SetCodeTx{
		ChainID:   uint256.MustFromBig(bigMezoChainID),
		GasTipCap: uint256.NewInt(1),
		GasFeeCap: uint256.NewInt(1),
		Gas:       100,
		To:        to,
		Value:     uint256.NewInt(0),
		AuthList:  auths,
	})

	cosmosTx, err := evmtypes.NewSetCodeTx(tx)
	require.NoError(t, err)

	got := cosmosTx.GetAuthorizationList()
	require.Equal(t, auths, got)
}

func TestSetCodeTxGetters(t *testing.T) {
	addr := utiltx.GenerateAddress()
	hexAddr := addr.Hex()

	tx := &evmtypes.SetCodeTx{
		ChainID:   &mezoChainID,
		GasTipCap: &mezoChainID,
		GasFeeCap: &mezoChainID,
		GasLimit:  21000,
		To:        hexAddr,
		Amount:    &mezoChainID,
		Nonce:     7,
		Data:      []byte("payload"),
	}

	require.Equal(t, big.NewInt(31612), tx.GetChainID())
	require.Nil(t, tx.GetAccessList())
	require.Equal(t, []byte("payload"), tx.GetData())
	require.Equal(t, uint64(21000), tx.GetGas())
	require.Equal(t, big.NewInt(31612), tx.GetGasFeeCap())
	require.Equal(t, big.NewInt(31612), tx.GetGasTipCap())
	require.Equal(t, big.NewInt(31612), tx.GetGasPrice())
	require.Equal(t, big.NewInt(31612), tx.GetValue())
	require.Equal(t, uint64(7), tx.GetNonce())
	require.Equal(t, &addr, tx.GetTo())

	// Empty To field returns nil
	tx.To = ""
	require.Nil(t, tx.GetTo())
}

func TestSetCodeTxFeeCost(t *testing.T) {
	tx := &evmtypes.SetCodeTx{
		GasFeeCap: &mezoChainID,
		GasTipCap: &mezoChainID,
		GasLimit:  2,
		Amount:    &mezoChainID,
	}

	expFee := new(big.Int).Mul(big.NewInt(31612), big.NewInt(2))
	require.Equal(t, expFee, tx.Fee())

	expCost := new(big.Int).Add(expFee, big.NewInt(31612))
	require.Equal(t, expCost, tx.Cost())

	baseFee := big.NewInt(0)
	require.Equal(t, big.NewInt(31612), tx.EffectiveGasPrice(baseFee))
	require.Equal(t, expFee, tx.EffectiveFee(baseFee))
	require.Equal(t, expCost, tx.EffectiveCost(baseFee))
}

// TestSetCodeTxAnyRoundTrip marshals a packed Any to bytes and unmarshals it
// back so the round-trip exercises proto encoding of V/R/S, AuthList, and the
// ChainID customtype, plus the codec interface registration. Without the byte
// round-trip the cached *SetCodeTx pointer would short-circuit decoding.
func TestSetCodeTxAnyRoundTrip(t *testing.T) {
	registry := types.NewInterfaceRegistry()
	evmtypes.RegisterInterfaces(registry)

	innerChainID := sdkmath.NewInt(31611)
	auth := evmtypes.SetCodeAuthorization{
		ChainID: &innerChainID,
		Address: utiltx.GenerateAddress().Hex(),
		Nonce:   1,
		V:       []byte{1},
		R:       big.NewInt(7).Bytes(),
		S:       big.NewInt(11).Bytes(),
	}

	original := &evmtypes.SetCodeTx{
		ChainID:   &mezoChainID,
		GasTipCap: &mezoChainID,
		GasFeeCap: &mezoChainID,
		Amount:    &mezoChainID,
		GasLimit:  21000,
		Nonce:     1,
		To:        utiltx.GenerateAddress().Hex(),
		Data:      []byte("payload"),
		AuthList:  evmtypes.AuthorizationList{auth},
		V:         []byte{1},
		R:         []byte{2},
		S:         []byte{3},
	}

	packed, err := evmtypes.PackTxData(original)
	require.NoError(t, err)

	bz, err := proto.Marshal(packed)
	require.NoError(t, err)

	var fresh types.Any
	require.NoError(t, proto.Unmarshal(bz, &fresh))
	require.Empty(t, fresh.GetCachedValue(),
		"a freshly-unmarshaled Any must not carry a cached value")

	var out evmtypes.TxData
	require.NoError(t, registry.UnpackAny(&fresh, &out))

	require.Equal(t, original, out)
}
