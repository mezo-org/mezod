package types

import (
	"testing"

	sdkmath "cosmossdk.io/math"

	evmtypes "github.com/mezo-org/mezod/x/evm/types"
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
			desc: "proper genesis without triparty",
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
			desc: "negative triparty per-request limit",
			genState: func() *GenesisState {
				genState := DefaultGenesis()
				genState.SourceBtcToken = token
				genState.TripartyPerRequestLimit = sdkmath.NewInt(-1)
				return genState
			},
			valid:       false,
			errContains: "genesis triparty per-request limit cannot be negative",
		},
		{
			desc: "negative triparty window limit",
			genState: func() *GenesisState {
				genState := DefaultGenesis()
				genState.SourceBtcToken = token
				genState.TripartyWindowLimit = sdkmath.NewInt(-1)
				return genState
			},
			valid:       false,
			errContains: "genesis triparty window limit cannot be negative",
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
			desc: "negative triparty processed sequence tip",
			genState: func() *GenesisState {
				genState := DefaultGenesis()
				genState.SourceBtcToken = token
				genState.TripartyProcessedSequenceTip = sdkmath.NewInt(-1)
				return genState
			},
			valid:       false,
			errContains: "genesis triparty processed sequence tip cannot be negative",
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
			desc: "negative triparty window consumed",
			genState: func() *GenesisState {
				genState := DefaultGenesis()
				genState.SourceBtcToken = token
				genState.TripartyWindowConsumed = sdkmath.NewInt(-1)
				return genState
			},
			valid:       false,
			errContains: "genesis triparty window consumed cannot be negative",
		},
		{
			desc: "triparty controller BTC minted - empty controller",
			genState: func() *GenesisState {
				genState := DefaultGenesis()
				genState.SourceBtcToken = token
				genState.TripartyControllerBtcMinted = []*TripartyControllerBTCMinted{
					{Controller: "", Amount: sdkmath.NewInt(100)},
				}
				return genState
			},
			valid:       false,
			errContains: "triparty controller BTC minted entry 0 controller cannot be empty",
		},
		{
			desc: "triparty controller BTC minted - invalid hex controller",
			genState: func() *GenesisState {
				genState := DefaultGenesis()
				genState.SourceBtcToken = token
				genState.TripartyControllerBtcMinted = []*TripartyControllerBTCMinted{
					{Controller: "not-hex", Amount: sdkmath.NewInt(100)},
				}
				return genState
			},
			valid:       false,
			errContains: "must be a valid hex-encoded EVM address",
		},
		{
			desc: "triparty controller BTC minted - zero address controller",
			genState: func() *GenesisState {
				genState := DefaultGenesis()
				genState.SourceBtcToken = token
				genState.TripartyControllerBtcMinted = []*TripartyControllerBTCMinted{
					{Controller: "0x0000000000000000000000000000000000000000", Amount: sdkmath.NewInt(100)},
				}
				return genState
			},
			valid:       false,
			errContains: "cannot be the zero EVM address",
		},
		{
			desc: "triparty controller BTC minted - negative amount",
			genState: func() *GenesisState {
				genState := DefaultGenesis()
				genState.SourceBtcToken = token
				genState.TripartyControllerBtcMinted = []*TripartyControllerBTCMinted{
					{Controller: "0x1111111111111111111111111111111111111111", Amount: sdkmath.NewInt(-1)},
				}
				return genState
			},
			valid:       false,
			errContains: "amount must be positive",
		},
		{
			desc: "triparty controller BTC minted - zero amount",
			genState: func() *GenesisState {
				genState := DefaultGenesis()
				genState.SourceBtcToken = token
				genState.TripartyControllerBtcMinted = []*TripartyControllerBTCMinted{
					{Controller: "0x1111111111111111111111111111111111111111", Amount: sdkmath.NewInt(0)},
				}
				return genState
			},
			valid:       false,
			errContains: "amount must be positive",
		},
		{
			desc: "triparty controller BTC minted - duplicate controllers",
			genState: func() *GenesisState {
				genState := DefaultGenesis()
				genState.SourceBtcToken = token
				genState.TripartyControllerBtcMinted = []*TripartyControllerBTCMinted{
					{Controller: "0x1111111111111111111111111111111111111111", Amount: sdkmath.NewInt(100)},
					{Controller: "0x1111111111111111111111111111111111111111", Amount: sdkmath.NewInt(200)},
				}
				return genState
			},
			valid:       false,
			errContains: "has duplicate controller",
		},
		{
			desc: "triparty controller BTC minted - duplicate controllers different casing",
			genState: func() *GenesisState {
				genState := DefaultGenesis()
				genState.SourceBtcToken = token
				genState.TripartyControllerBtcMinted = []*TripartyControllerBTCMinted{
					{Controller: "0xAbCd000000000000000000000000000000000001", Amount: sdkmath.NewInt(100)},
					{Controller: "0xabcd000000000000000000000000000000000001", Amount: sdkmath.NewInt(200)},
				}
				return genState
			},
			valid:       false,
			errContains: "has duplicate controller",
		},
		{
			desc: "empty allowed triparty controller",
			genState: func() *GenesisState {
				genState := DefaultGenesis()
				genState.SourceBtcToken = token
				genState.AllowedTripartyControllers = []string{""}
				return genState
			},
			valid:       false,
			errContains: "allowed triparty controller 0 cannot be empty",
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
			desc: "zero-address allowed triparty controller",
			genState: func() *GenesisState {
				genState := DefaultGenesis()
				genState.SourceBtcToken = token
				genState.AllowedTripartyControllers = []string{evmtypes.ZeroHexAddress()}
				return genState
			},
			valid:       false,
			errContains: "allowed triparty controller 0 cannot be the zero EVM address",
		},
		{
			desc: "pending triparty request with non-positive sequence",
			genState: func() *GenesisState {
				genState := DefaultGenesis()
				genState.SourceBtcToken = token
				genState.TripartyRequestSequenceTip = sdkmath.NewInt(1)
				genState.TripartyPendingRequests = []*TripartyBridgeRequest{
					{
						Sequence:    sdkmath.NewInt(0),
						BlockHeight: 100,
						Recipient:   "0x2222222222222222222222222222222222222222",
						Amount:      sdkmath.NewInt(10),
						Controller:  "0x1111111111111111111111111111111111111111",
					},
				}
				return genState
			},
			valid:       false,
			errContains: "pending triparty request 0 sequence must be positive",
		},
		{
			desc: "pending triparty request with negative block height",
			genState: func() *GenesisState {
				genState := DefaultGenesis()
				genState.SourceBtcToken = token
				genState.TripartyRequestSequenceTip = sdkmath.NewInt(1)
				genState.TripartyPendingRequests = []*TripartyBridgeRequest{
					{
						Sequence:    sdkmath.NewInt(1),
						BlockHeight: -1,
						Recipient:   "0x2222222222222222222222222222222222222222",
						Amount:      sdkmath.NewInt(10),
						Controller:  "0x1111111111111111111111111111111111111111",
					},
				}
				return genState
			},
			valid:       false,
			errContains: "pending triparty request 0 block height cannot be negative",
		},
		{
			desc: "pending triparty request with empty recipient",
			genState: func() *GenesisState {
				genState := DefaultGenesis()
				genState.SourceBtcToken = token
				genState.TripartyRequestSequenceTip = sdkmath.NewInt(1)
				genState.TripartyPendingRequests = []*TripartyBridgeRequest{
					{
						Sequence:    sdkmath.NewInt(1),
						BlockHeight: 100,
						Recipient:   "",
						Amount:      sdkmath.NewInt(10),
						Controller:  "0x1111111111111111111111111111111111111111",
					},
				}
				return genState
			},
			valid:       false,
			errContains: "pending triparty request 0 recipient cannot be empty",
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
			desc: "pending triparty request with zero-address recipient",
			genState: func() *GenesisState {
				genState := DefaultGenesis()
				genState.SourceBtcToken = token
				genState.TripartyRequestSequenceTip = sdkmath.NewInt(1)
				genState.TripartyPendingRequests = []*TripartyBridgeRequest{
					{
						Sequence:    sdkmath.NewInt(1),
						BlockHeight: 100,
						Recipient:   evmtypes.ZeroHexAddress(),
						Amount:      sdkmath.NewInt(10),
						Controller:  "0x1111111111111111111111111111111111111111",
					},
				}
				return genState
			},
			valid:       false,
			errContains: "pending triparty request 0 recipient cannot be the zero EVM address",
		},
		{
			desc: "pending triparty request with empty controller",
			genState: func() *GenesisState {
				genState := DefaultGenesis()
				genState.SourceBtcToken = token
				genState.TripartyRequestSequenceTip = sdkmath.NewInt(1)
				genState.TripartyPendingRequests = []*TripartyBridgeRequest{
					{
						Sequence:    sdkmath.NewInt(1),
						BlockHeight: 100,
						Recipient:   "0x2222222222222222222222222222222222222222",
						Amount:      sdkmath.NewInt(10),
						Controller:  "",
					},
				}
				return genState
			},
			valid:       false,
			errContains: "pending triparty request 0 controller cannot be empty",
		},
		{
			desc: "pending triparty request with invalid controller",
			genState: func() *GenesisState {
				genState := DefaultGenesis()
				genState.SourceBtcToken = token
				genState.TripartyRequestSequenceTip = sdkmath.NewInt(1)
				genState.TripartyPendingRequests = []*TripartyBridgeRequest{
					{
						Sequence:    sdkmath.NewInt(1),
						BlockHeight: 100,
						Recipient:   "0x2222222222222222222222222222222222222222",
						Amount:      sdkmath.NewInt(10),
						Controller:  "corrupted",
					},
				}
				return genState
			},
			valid:       false,
			errContains: "pending triparty request 0 controller must be a valid hex-encoded EVM address",
		},
		{
			desc: "pending triparty request with zero controller",
			genState: func() *GenesisState {
				genState := DefaultGenesis()
				genState.SourceBtcToken = token
				genState.TripartyRequestSequenceTip = sdkmath.NewInt(1)
				genState.TripartyPendingRequests = []*TripartyBridgeRequest{
					{
						Sequence:    sdkmath.NewInt(1),
						BlockHeight: 100,
						Recipient:   "0x2222222222222222222222222222222222222222",
						Amount:      sdkmath.NewInt(10),
						Controller:  evmtypes.ZeroHexAddress(),
					},
				}
				return genState
			},
			valid:       false,
			errContains: "pending triparty request 0 controller cannot be the zero EVM address",
		},
		{
			desc: "pending triparty request with non-positive amount",
			genState: func() *GenesisState {
				genState := DefaultGenesis()
				genState.SourceBtcToken = token
				genState.TripartyRequestSequenceTip = sdkmath.NewInt(1)
				genState.TripartyPendingRequests = []*TripartyBridgeRequest{
					{
						Sequence:    sdkmath.NewInt(1),
						BlockHeight: 100,
						Recipient:   "0x2222222222222222222222222222222222222222",
						Amount:      sdkmath.NewInt(0),
						Controller:  "0x1111111111111111111111111111111111111111",
					},
				}
				return genState
			},
			valid:       false,
			errContains: "pending triparty request 0 amount must be positive",
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
			desc: "pending triparty requests with duplicate sequences",
			genState: func() *GenesisState {
				genState := DefaultGenesis()
				genState.SourceBtcToken = token
				genState.TripartyRequestSequenceTip = sdkmath.NewInt(3)
				genState.TripartyProcessedSequenceTip = sdkmath.NewInt(1)
				genState.TripartyPendingRequests = []*TripartyBridgeRequest{
					{
						Sequence:    sdkmath.NewInt(2),
						BlockHeight: 100,
						Recipient:   "0x2222222222222222222222222222222222222223",
						Amount:      sdkmath.NewInt(50),
						Controller:  "0x1111111111111111111111111111111111111111",
					},
					{
						Sequence:    sdkmath.NewInt(2),
						BlockHeight: 101,
						Recipient:   "0x2222222222222222222222222222222222222224",
						Amount:      sdkmath.NewInt(51),
						Controller:  "0x1111111111111111111111111111111111111112",
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
				genState.TripartyControllerBtcMinted = []*TripartyControllerBTCMinted{
					{
						Controller: "0x1111111111111111111111111111111111111111",
						Amount:     sdkmath.NewInt(200),
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
