package types

import (
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/params"

	"github.com/stretchr/testify/require"
)

func TestParamsValidate(t *testing.T) {
	extraEips := []int64{2929, 1884, 1344}
	testCases := []struct {
		name     string
		params   Params
		expError bool
	}{
		{"default", DefaultParams(), false},
		{
			"valid",
			NewParams("ara", false, true, true, DefaultChainConfig(), extraEips),
			false,
		},
		{
			"empty",
			Params{},
			true,
		},
		{
			"invalid evm denom",
			Params{
				EvmDenom: "@!#!@$!@5^32",
			},
			true,
		},
		{
			"invalid eip",
			Params{
				EvmDenom:  "stake",
				ExtraEIPs: []int64{1},
			},
			true,
		},
	}

	for _, tc := range testCases {
		err := tc.params.Validate()

		if tc.expError {
			require.Error(t, err, tc.name)
		} else {
			require.NoError(t, err, tc.name)
		}
	}
}

func TestParamsValidateTxLockdownAddresses(t *testing.T) {
	// txLockdownAddresses returns a list of count distinct hex-encoded EVM
	// addresses. The first address is 0x00...01, so the zero address is never
	// part of the list.
	txLockdownAddresses := func(count int) []string {
		addresses := make([]string, count)
		for i := range addresses {
			addresses[i] = fmt.Sprintf("0x%040x", i+1)
		}
		return addresses
	}

	// paramsWithTxLockdown returns the default parameters with the transaction
	// lockdown allowlists replaced by the given lists.
	paramsWithTxLockdown := func(senders, targets []string) Params {
		params := DefaultParams()
		params.TxLockdownEnabled = true
		params.TxLockdownSenders = senders
		params.TxLockdownTargets = targets
		return params
	}

	testCases := []struct {
		name     string
		params   Params
		expError bool
	}{
		{
			"empty allowlists",
			paramsWithTxLockdown(nil, nil),
			false,
		},
		{
			"valid allowlists",
			paramsWithTxLockdown(
				txLockdownAddresses(2),
				txLockdownAddresses(3),
			),
			false,
		},
		{
			"lowercase and checksum forms of distinct addresses",
			paramsWithTxLockdown(
				[]string{
					"0x1111111111111111111111111111111111111111",
					"0x2222222222222222222222222222222222222222",
				},
				nil,
			),
			false,
		},
		{
			"empty sender entry",
			paramsWithTxLockdown([]string{""}, nil),
			true,
		},
		{
			"empty target entry",
			paramsWithTxLockdown(nil, []string{""}),
			true,
		},
		{
			"invalid hex sender",
			paramsWithTxLockdown([]string{"0xnothexadecimal"}, nil),
			true,
		},
		{
			"invalid hex target",
			paramsWithTxLockdown(nil, []string{"0xnothexadecimal"}),
			true,
		},
		{
			"sender is too short",
			paramsWithTxLockdown([]string{"0x1111"}, nil),
			true,
		},
		{
			"zero address sender",
			paramsWithTxLockdown(
				[]string{"0x0000000000000000000000000000000000000000"},
				nil,
			),
			true,
		},
		{
			"zero address target",
			paramsWithTxLockdown(
				nil,
				[]string{"0x0000000000000000000000000000000000000000"},
			),
			true,
		},
		{
			"duplicate sender",
			paramsWithTxLockdown(
				[]string{
					"0x1111111111111111111111111111111111111111",
					"0x1111111111111111111111111111111111111111",
				},
				nil,
			),
			true,
		},
		{
			"duplicate sender in a different case",
			paramsWithTxLockdown(
				[]string{
					"0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					"0xAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
				},
				nil,
			),
			true,
		},
		{
			"duplicate target in a different case",
			paramsWithTxLockdown(
				nil,
				[]string{
					"0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					"0xAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
				},
			),
			true,
		},
		{
			"senders at the size cap",
			paramsWithTxLockdown(txLockdownAddresses(256), nil),
			false,
		},
		{
			"targets at the size cap",
			paramsWithTxLockdown(nil, txLockdownAddresses(256)),
			false,
		},
		{
			"senders above the size cap",
			paramsWithTxLockdown(txLockdownAddresses(257), nil),
			true,
		},
		{
			"targets above the size cap",
			paramsWithTxLockdown(nil, txLockdownAddresses(257)),
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.params.Validate()

			if tc.expError {
				require.Error(t, err, tc.name)
			} else {
				require.NoError(t, err, tc.name)
			}
		})
	}
}

func TestParamsEIPs(t *testing.T) {
	extraEips := []int64{2929, 1884, 1344}
	params := NewParams("ara", false, true, true, DefaultChainConfig(), extraEips)
	actual := params.EIPs()

	require.Equal(t, []int{2929, 1884, 1344}, actual)
}

func TestParamsValidatePriv(t *testing.T) {
	require.Error(t, validateEVMDenom(false))
	require.NoError(t, validateEVMDenom("inj"))
	require.Error(t, validateBool(""))
	require.NoError(t, validateBool(true))
	require.Error(t, validateEIPs(""))
	require.NoError(t, validateEIPs([]int64{1884}))
}

func TestValidateChainConfig(t *testing.T) {
	testCases := []struct {
		name     string
		i        interface{}
		expError bool
	}{
		{
			"invalid chain config type",
			"string",
			true,
		},
		{
			"valid chain config type",
			DefaultChainConfig(),
			false,
		},
	}
	for _, tc := range testCases {
		err := validateChainConfig(tc.i)

		if tc.expError {
			require.Error(t, err, tc.name)
		} else {
			require.NoError(t, err, tc.name)
		}
	}
}

func TestIsLondon(t *testing.T) {
	testCases := []struct {
		name   string
		height int64
		result bool
	}{
		{
			"Before london block",
			5,
			false,
		},
		{
			"After london block",
			12_965_001,
			true,
		},
		{
			"london block",
			12_965_000,
			true,
		},
	}

	for _, tc := range testCases {
		ethConfig := params.MainnetChainConfig
		require.Equal(t, IsLondon(ethConfig, tc.height), tc.result)
	}
}
