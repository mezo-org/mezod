package types_test

import (
	"math/big"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/require"

	"github.com/mezo-org/mezod/x/evm/types"
)

func TestNewAuthorizationList_NilAndEmpty(t *testing.T) {
	require.Nil(t, types.NewAuthorizationList(nil))

	empty := types.NewAuthorizationList([]ethtypes.SetCodeAuthorization{})
	require.NotNil(t, empty)
	require.Len(t, empty, 0)
}

func TestAuthorizationList_ToEthAuthorizationList_Nil(t *testing.T) {
	var al types.AuthorizationList
	require.Nil(t, al.ToEthAuthorizationList())
}

func TestSetCodeAuthorization_RoundTrip(t *testing.T) {
	addr := common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")

	// Pick a value near the high end of uint256 (well above int256 bounds, but
	// the round-trip path itself does no bounds checking — that is Validate's
	// job).
	largeR := uint256.MustFromHex(
		"0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
	)

	tests := []struct {
		name string
		auth ethtypes.SetCodeAuthorization
	}{
		{
			"zero values",
			ethtypes.SetCodeAuthorization{
				ChainID: uint256.Int{},
				Address: common.Address{},
				Nonce:   0,
				V:       0,
				R:       uint256.Int{},
				S:       uint256.Int{},
			},
		},
		{
			"v=0 with non-zero values",
			ethtypes.SetCodeAuthorization{
				ChainID: *uint256.NewInt(31611),
				Address: addr,
				Nonce:   42,
				V:       0,
				R:       *uint256.NewInt(7),
				S:       *uint256.NewInt(11),
			},
		},
		{
			"v=1 with large R",
			ethtypes.SetCodeAuthorization{
				ChainID: *uint256.NewInt(31612),
				Address: addr,
				Nonce:   1<<63 - 1,
				V:       1,
				R:       *largeR,
				S:       *uint256.NewInt(1),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			proto := types.NewSetCodeAuthorization(tc.auth)
			back := proto.ToEthAuthorization()

			require.Equal(t, tc.auth.ChainID, back.ChainID)
			require.Equal(t, tc.auth.Address, back.Address)
			require.Equal(t, tc.auth.Nonce, back.Nonce)
			require.Equal(t, tc.auth.V, back.V)
			require.Equal(t, tc.auth.R, back.R)
			require.Equal(t, tc.auth.S, back.S)
		})
	}
}

func TestAuthorizationList_RoundTrip(t *testing.T) {
	addr := common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")

	in := []ethtypes.SetCodeAuthorization{
		{
			ChainID: *uint256.NewInt(31611),
			Address: addr,
			Nonce:   1,
			V:       0,
			R:       *uint256.NewInt(2),
			S:       *uint256.NewInt(3),
		},
		{
			ChainID: uint256.Int{},
			Address: common.Address{},
			Nonce:   0,
			V:       1,
			R:       uint256.Int{},
			S:       uint256.Int{},
		},
		{
			ChainID: *uint256.NewInt(31612),
			Address: addr,
			Nonce:   99,
			V:       1,
			R:       *uint256.NewInt(0xff),
			S:       *uint256.NewInt(0xee),
		},
	}

	proto := types.NewAuthorizationList(in)
	require.Len(t, proto, len(in))

	back := proto.ToEthAuthorizationList()
	require.Equal(t, in, back)
}

func TestSetCodeAuthorization_GetChainID(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		a := types.SetCodeAuthorization{}
		require.Nil(t, a.GetChainID())
	})

	t.Run("non-nil", func(t *testing.T) {
		ci := sdkmath.NewInt(31611)
		a := types.SetCodeAuthorization{ChainID: &ci}
		require.Equal(t, big.NewInt(31611), a.GetChainID())
	})
}

func TestSetCodeAuthorization_EmptyByteFieldsRoundTripToZero(t *testing.T) {
	// A proto-decoded authorization can land with nil V/R/S byte slices (an
	// uninitialized SetCodeAuthorization) — converting to eth and back must
	// not panic and must yield zeros.
	proto := types.SetCodeAuthorization{
		Address: common.Address{}.Hex(),
	}
	eth := proto.ToEthAuthorization()
	require.Equal(t, uint8(0), eth.V)
	require.Equal(t, uint256.Int{}, eth.R)
	require.Equal(t, uint256.Int{}, eth.S)
	require.Equal(t, uint256.Int{}, eth.ChainID)
}
