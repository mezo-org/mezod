package types

import (
	"testing"

	sdkmath "cosmossdk.io/math"

	"github.com/stretchr/testify/require"
)

func TestGenesisState_Validate(t *testing.T) {
	for _, tc := range []struct {
		desc        string
		genState    func() *GenesisState
		valid       bool
		errContains string
	}{
		{
			desc: "negative assets locked sequence tip",
			genState: func() *GenesisState {
				genState := DefaultGenesis()
				genState.AssetsLockedSequenceTip = sdkmath.NewInt(-1)
				return genState
			},
			valid:       false,
			errContains: "genesis assets locked sequence tip cannot be negative",
		},
		{
			desc: "negative assets unlocked sequence tip",
			genState: func() *GenesisState {
				genState := DefaultGenesis()
				genState.AssetsUnlockedSequenceTip = sdkmath.NewInt(-1)
				return genState
			},
			valid:       false,
			errContains: "genesis assets unlocked sequence tip cannot be negative",
		},
		{
			desc:        "missing source btc token",
			genState:    DefaultGenesis,
			valid:       false,
			errContains: "source btc token cannot be empty",
		},
		{
			desc: "invalid source btc token",
			genState: func() *GenesisState {
				genState := DefaultGenesis()
				genState.SourceBtcToken = "corrupted"
				return genState
			},
			valid:       false,
			errContains: "source btc token must be a valid hex-encoded EVM address",
		},
		{
			desc: "mapping with empty source token",
			genState: func() *GenesisState {
				genState := DefaultGenesis()
				genState.SourceBtcToken = token
				genState.Erc20TokensMappings = []*ERC20TokenMapping{
					{
						SourceToken: "",
						MezoToken:   "0x4992eeF73616587B3463B0f9dEb92fF1087D39D2",
					},
				}
				return genState
			},
			valid:       false,
			errContains: "source token of ERC20 mapping 0 cannot be empty",
		},
		{
			desc: "mapping with invalid source token",
			genState: func() *GenesisState {
				genState := DefaultGenesis()
				genState.SourceBtcToken = token
				genState.Erc20TokensMappings = []*ERC20TokenMapping{
					{
						SourceToken: "corrupted",
						MezoToken:   "0x4992eeF73616587B3463B0f9dEb92fF1087D39D2",
					},
				}
				return genState
			},
			valid:       false,
			errContains: "source token of ERC20 mapping 0 must be a valid hex-encoded EVM address",
		},
		{
			desc: "mapping with empty mezo token",
			genState: func() *GenesisState {
				genState := DefaultGenesis()
				genState.SourceBtcToken = token
				genState.Erc20TokensMappings = []*ERC20TokenMapping{
					{
						SourceToken: "0x4992eeF73616587B3463B0f9dEb92fF1087D39D2",
						MezoToken:   "",
					},
				}
				return genState
			},
			valid:       false,
			errContains: "mezo token of ERC20 mapping 0 cannot be empty",
		},
		{
			desc: "mapping with invalid mezo token",
			genState: func() *GenesisState {
				genState := DefaultGenesis()
				genState.SourceBtcToken = token
				genState.Erc20TokensMappings = []*ERC20TokenMapping{
					{
						SourceToken: "0x4992eeF73616587B3463B0f9dEb92fF1087D39D2",
						MezoToken:   "corrupted",
					},
				}
				return genState
			},
			valid:       false,
			errContains: "mezo token of ERC20 mapping 0 must be a valid hex-encoded EVM address",
		},
		{
			desc: "proper genesis",
			genState: func() *GenesisState {
				genState := DefaultGenesis()
				genState.SourceBtcToken = token
				genState.Erc20TokensMappings = []*ERC20TokenMapping{
					{
						SourceToken: "0xe8bC8EA7A06151276a89579dFDa03a9a59d5F085",
						MezoToken:   "0x4992eeF73616587B3463B0f9dEb92fF1087D39D2",
					},
				}
				return genState
			},
			valid: true,
		},
	} {
		t.Run(
			tc.desc, func(t *testing.T) {
				err := tc.genState().Validate()
				if tc.valid {
					require.NoError(t, err)
				} else {
					require.ErrorContains(t, err, tc.errContains)
				}
			},
		)
	}
}
