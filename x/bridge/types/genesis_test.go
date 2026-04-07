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
		{
			desc: "triparty block delay less than one",
			genState: func() *GenesisState {
				genState := DefaultGenesis()
				genState.SourceBtcToken = token
				genState.TripartyBlockDelay = 0
				return genState
			},
			valid:       false,
			errContains: "genesis triparty block delay cannot be less than 1",
		},
		{
			desc: "negative triparty request sequence tip",
			genState: func() *GenesisState {
				genState := DefaultGenesis()
				genState.SourceBtcToken = token
				genState.TripartyRequestSequenceTip = sdkmath.NewInt(-1)
				return genState
			},
			valid:       false,
			errContains: "genesis triparty request sequence tip cannot be negative",
		},
		{
			desc: "invalid allowed triparty controller",
			genState: func() *GenesisState {
				genState := DefaultGenesis()
				genState.SourceBtcToken = token
				genState.AllowedTripartyControllers = []string{"bad-controller"}
				return genState
			},
			valid:       false,
			errContains: "allowed triparty controller 0 must be a valid hex-encoded EVM address",
		},
		{
			desc: "pending triparty request with invalid recipient",
			genState: func() *GenesisState {
				genState := DefaultGenesis()
				genState.SourceBtcToken = token
				genState.TripartyPendingRequests = []*TripartyBridgeRequest{
					{
						Sequence:    sdkmath.NewInt(1),
						BlockHeight: 100,
						Recipient:   "bad-recipient",
						Amount:      sdkmath.NewInt(10),
						Controller:  "0x1111111111111111111111111111111111111111",
					},
				}
				return genState
			},
			valid:       false,
			errContains: "pending triparty request 0 recipient must be a valid hex-encoded EVM address",
		},
		{
			desc: "pending triparty request with callback data too large",
			genState: func() *GenesisState {
				genState := DefaultGenesis()
				genState.SourceBtcToken = token
				genState.TripartyRequestSequenceTip = sdkmath.NewInt(1)
				genState.TripartyPendingRequests = []*TripartyBridgeRequest{
					{
						Sequence:     sdkmath.NewInt(1),
						BlockHeight:  100,
						Recipient:    "0x2222222222222222222222222222222222222222",
						Amount:       sdkmath.NewInt(10),
						CallbackData: make([]byte, MaxTripartyCallbackDataLength+1),
						Controller:   "0x1111111111111111111111111111111111111111",
					},
				}
				return genState
			},
			valid:       false,
			errContains: "pending triparty request 0 callback data exceeds maximum length",
		},
		{
			desc: "triparty request sequence tip less than processed sequence tip",
			genState: func() *GenesisState {
				genState := DefaultGenesis()
				genState.SourceBtcToken = token
				genState.TripartyRequestSequenceTip = sdkmath.NewInt(1)
				genState.TripartyProcessedSequenceTip = sdkmath.NewInt(2)
				return genState
			},
			valid:       false,
			errContains: "genesis triparty request sequence tip cannot be less than processed sequence tip",
		},
		{
			desc: "pending triparty request sequence not above processed sequence tip",
			genState: func() *GenesisState {
				genState := DefaultGenesis()
				genState.SourceBtcToken = token
				genState.TripartyRequestSequenceTip = sdkmath.NewInt(2)
				genState.TripartyProcessedSequenceTip = sdkmath.NewInt(1)
				genState.TripartyPendingRequests = []*TripartyBridgeRequest{
					{
						Sequence:    sdkmath.NewInt(1),
						BlockHeight: 100,
						Recipient:   "0x2222222222222222222222222222222222222222",
						Amount:      sdkmath.NewInt(50),
						Controller:  "0x1111111111111111111111111111111111111111",
					},
				}
				return genState
			},
			valid:       false,
			errContains: "pending triparty request 0 sequence must be greater than processed sequence tip",
		},
		{
			desc: "pending triparty request sequence above request sequence tip",
			genState: func() *GenesisState {
				genState := DefaultGenesis()
				genState.SourceBtcToken = token
				genState.TripartyRequestSequenceTip = sdkmath.NewInt(1)
				genState.TripartyPendingRequests = []*TripartyBridgeRequest{
					{
						Sequence:    sdkmath.NewInt(2),
						BlockHeight: 100,
						Recipient:   "0x2222222222222222222222222222222222222222",
						Amount:      sdkmath.NewInt(50),
						Controller:  "0x1111111111111111111111111111111111111111",
					},
				}
				return genState
			},
			valid:       false,
			errContains: "pending triparty request 0 sequence cannot be greater than request sequence tip",
		},
		{
			desc: "pending triparty requests with sequence gap",
			genState: func() *GenesisState {
				genState := DefaultGenesis()
				genState.SourceBtcToken = token
				genState.TripartyRequestSequenceTip = sdkmath.NewInt(3)
				genState.TripartyProcessedSequenceTip = sdkmath.NewInt(1)
				genState.TripartyPendingRequests = []*TripartyBridgeRequest{
					{
						Sequence:    sdkmath.NewInt(3),
						BlockHeight: 100,
						Recipient:   "0x2222222222222222222222222222222222222222",
						Amount:      sdkmath.NewInt(50),
						Controller:  "0x1111111111111111111111111111111111111111",
					},
				}
				return genState
			},
			valid:       false,
			errContains: "pending triparty requests must form a gapless range between processed and request sequence tips",
		},
		{
			desc: "proper genesis with triparty state",
			genState: func() *GenesisState {
				genState := DefaultGenesis()
				genState.SourceBtcToken = token
				genState.AllowedTripartyControllers = []string{
					"0x1111111111111111111111111111111111111111",
				}
				genState.TripartyPaused = true
				genState.TripartyBlockDelay = 10
				genState.TripartyPerRequestLimit = sdkmath.NewInt(123)
				genState.TripartyWindowLimit = sdkmath.NewInt(456)
				genState.TripartyRequestSequenceTip = sdkmath.NewInt(2)
				genState.TripartyProcessedSequenceTip = sdkmath.NewInt(1)
				genState.TripartyPendingRequests = []*TripartyBridgeRequest{
					{
						Sequence:    sdkmath.NewInt(2),
						BlockHeight: 100,
						Recipient:   "0x2222222222222222222222222222222222222222",
						Amount:      sdkmath.NewInt(50),
						Controller:  "0x1111111111111111111111111111111111111111",
					},
				}
				genState.TripartyWindowConsumed = sdkmath.NewInt(50)
				genState.TripartyWindowLastReset = 500
				genState.TripartyTotalBtcMinted = sdkmath.NewInt(200)
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
