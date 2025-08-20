package assetsbridge_test

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"testing"
	"time"

	"cosmossdk.io/math"
	"github.com/cosmos/btcutil/base58"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/mezo-org/mezod/precompile"
	"github.com/mezo-org/mezod/precompile/assetsbridge"
	bridgetypes "github.com/mezo-org/mezod/x/bridge/types"
	"github.com/mezo-org/mezod/x/evm/statedb"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
	"github.com/stretchr/testify/suite"
)

// Test addresses and constants
var (
	testBTCToken     = common.HexToAddress(evmtypes.BTCTokenPrecompileAddress)
	testERC20Token   = common.HexToAddress("0x1234567890123456789012345678901234567890")
	unsupportedToken = common.HexToAddress("0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF")
	bridgeAddress    = common.HexToAddress(evmtypes.AssetsBridgePrecompileAddress)

	ethRecipient      = []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10, 0x11, 0x12, 0x13, 0x14}
	btcRecipient      = makeValidScript("1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa")
	invalidBtcAddress = []byte("invalid_btc_address")
)

// just P2PKH for testing purpose
func makeValidScript(address string) []byte {
	recipient, _, _ := base58.CheckDecode(address)
	recipient = append([]byte{0x19, 0x76, 0xa9, 0x14}, recipient...)
	return append(recipient, []byte{0x88, 0xac}...)
}

type FakeAuthzKeeper struct {
	authorizations map[string]authz.Authorization
	expirations    map[string]*time.Time
}

func NewFakeAuthzKeeper() *FakeAuthzKeeper {
	return &FakeAuthzKeeper{
		authorizations: make(map[string]authz.Authorization),
		expirations:    make(map[string]*time.Time),
	}
}

func (k *FakeAuthzKeeper) GetAuthorization(_ context.Context, grantee, granter sdk.AccAddress, msgType string) (authz.Authorization, *time.Time) {
	key := fmt.Sprintf("%s-%s-%s", grantee.String(), granter.String(), msgType)
	return k.authorizations[key], k.expirations[key]
}

func (k *FakeAuthzKeeper) SetAuthorization(granter, grantee sdk.AccAddress, msgType string, auth authz.Authorization, expiration *time.Time) {
	key := fmt.Sprintf("%s-%s-%s", grantee.String(), granter.String(), msgType)
	k.authorizations[key] = auth
	k.expirations[key] = expiration
}

func (k *FakeAuthzKeeper) DeleteGrant(_ context.Context, _, _ sdk.AccAddress, _ string) error {
	return nil
}

func (k *FakeAuthzKeeper) SaveGrant(_ context.Context, _, _ sdk.AccAddress, _ authz.Authorization, _ *time.Time) error {
	return nil
}

type ExtendedFakeBridgeKeeper struct {
	*FakeBridgeKeeper
	assetsUnlockedCalled  bool
	assetsUnlockedSuccess bool
	lastAssetsUnlocked    *bridgetypes.AssetsUnlockedEvent
	sequenceNumber        int64
}

func NewExtendedFakeBridgeKeeper(sourceBTCToken []byte) *ExtendedFakeBridgeKeeper {
	return &ExtendedFakeBridgeKeeper{
		FakeBridgeKeeper: NewFakeBridgeKeeper(sourceBTCToken),
		sequenceNumber:   1,
	}
}

func (k *ExtendedFakeBridgeKeeper) SaveAssetsUnlocked(
	_ sdk.Context,
	recipient []byte,
	token []byte,
	sender []byte,
	amount math.Int,
	chain uint8,
) (*bridgetypes.AssetsUnlockedEvent, error) {
	k.assetsUnlockedCalled = true

	if !k.assetsUnlockedSuccess {
		return nil, errors.New("AssetsUnlocked failed")
	}

	event := &bridgetypes.AssetsUnlockedEvent{
		Token:          common.BytesToAddress(token).Hex(),
		Amount:         amount,
		Chain:          uint32(chain),
		Sender:         common.BytesToAddress(sender).Hex(),
		Recipient:      recipient,
		UnlockSequence: math.NewInt(k.sequenceNumber),
	}
	k.sequenceNumber++
	k.lastAssetsUnlocked = event

	return event, nil
}

func (k *ExtendedFakeBridgeKeeper) SetAssetsUnlockedSuccess(success bool) {
	k.assetsUnlockedSuccess = success
}

func (k *ExtendedFakeBridgeKeeper) AssetsUnlockedCalled() bool {
	return k.assetsUnlockedCalled
}

func (k *ExtendedFakeBridgeKeeper) Reset() {
	k.assetsUnlockedCalled = false
	k.lastAssetsUnlocked = nil
	k.burnErr = nil
	k.minAmountByToken = make(map[string]math.Int)
}

type BridgeOutTestSuite struct {
	PrecompileTestSuite

	authzKeeper     *FakeAuthzKeeper
	extBridgeKeeper *ExtendedFakeBridgeKeeper
}

func TestBridgeOutTestSuite(t *testing.T) {
	suite.Run(t, new(BridgeOutTestSuite))
}

func (s *BridgeOutTestSuite) SetupTest() {
	s.PrecompileTestSuite.SetupTest()

	s.authzKeeper = NewFakeAuthzKeeper()
	s.extBridgeKeeper = NewExtendedFakeBridgeKeeper(testBTCToken.Bytes())

	if err := s.extBridgeKeeper.CreateERC20TokenMapping(s.ctx, testERC20Token.Bytes(), testERC20Token.Bytes()); err != nil {
		s.FailNow("couldn't create ERC20 mapping: %v", err)
	}
}

func (s *BridgeOutTestSuite) TestInstantiateAssetBridge() {
	assetsBridgePrecompile, err := assetsbridge.NewPrecompile(
		s.poaKeeper,
		s.extBridgeKeeper,
		s.authzKeeper,
		&assetsbridge.Settings{
			Observability:   true,
			BTCManagement:   true,
			ERC20Management: true,
			SequenceTipView: true,
			BridgeOut:       true,
		},
	)
	s.Require().NoError(err)

	// Test that bridgeOut method exists
	method, exists := assetsBridgePrecompile.Abi.Methods["bridgeOut"]
	s.Require().True(exists, "bridgeOut method should exist")

	// Test method properties
	s.Require().Equal("bridgeOut", method.Name)
	s.Require().Equal(4, len(method.Inputs), "bridgeOut should have 4 inputs")
	s.Require().Equal("token", method.Inputs[0].Name)
	s.Require().Equal("amount", method.Inputs[1].Name)
	s.Require().Equal("chain", method.Inputs[2].Name)
	s.Require().Equal("recipient", method.Inputs[3].Name)
}

func (s *BridgeOutTestSuite) TestBridgeOutTokenValidation() {
	testcases := []TestCase{
		{
			name: "valid ERC20 token",
			run: func() []interface{} {
				s.extBridgeKeeper.SetAssetsUnlockedSuccess(true)
				return []interface{}{testERC20Token, big.NewInt(100), uint8(0), ethRecipient}
			},
			as:        s.account1.EvmAddr,
			basicPass: true,
			output:    []interface{}{true},
		},
		{
			name: "unsupported token",
			run: func() []interface{} {
				return []interface{}{unsupportedToken, big.NewInt(100), uint8(0), ethRecipient}
			},
			as:          s.account1.EvmAddr,
			basicPass:   true,
			revert:      true,
			errContains: "unsupported token",
		},
		{
			name: "zero address token",
			run: func() []interface{} {
				return []interface{}{common.Address{}, big.NewInt(100), uint8(0), ethRecipient}
			},
			as:          s.account1.EvmAddr,
			basicPass:   true,
			revert:      true,
			errContains: "unsupported token",
		},
	}

	s.RunMethodTestCasesWithKeepers(testcases, "bridgeOut")
}

func (s *BridgeOutTestSuite) TestBridgeOutAmountValidation() {
	testcases := []TestCase{
		{
			name: "minimum bridgeable amount not set",
			run: func() []interface{} {
				s.extBridgeKeeper.SetAssetsUnlockedSuccess(true)
				return []interface{}{testERC20Token, big.NewInt(200), uint8(0), ethRecipient}
			},
			as:        s.account1.EvmAddr,
			basicPass: true,
			output:    []interface{}{true},
		},
		{
			name: "amount below minimum bridgeable amount",
			run: func() []interface{} {
				s.extBridgeKeeper.SetAssetsUnlockedSuccess(true)
				_ = s.extBridgeKeeper.SetMinBridgeOutAmount(s.ctx, testERC20Token.Bytes(), math.NewInt(200))
				return []interface{}{testERC20Token, big.NewInt(199), uint8(0), ethRecipient}
			},
			as:          s.account1.EvmAddr,
			basicPass:   true,
			revert:      true,
			errContains: "amount below minimum bridgeable amount",
			output:      nil,
		},
		{
			name: "amount equal to minimum bridgeable amount",
			run: func() []interface{} {
				s.extBridgeKeeper.SetAssetsUnlockedSuccess(true)
				_ = s.extBridgeKeeper.SetMinBridgeOutAmount(s.ctx, testERC20Token.Bytes(), math.NewInt(200))
				return []interface{}{testERC20Token, big.NewInt(200), uint8(0), ethRecipient}
			},
			as:        s.account1.EvmAddr,
			basicPass: true,
			output:    []interface{}{true},
		},
		{
			name: "amount above minimum bridgeable amount",
			run: func() []interface{} {
				s.extBridgeKeeper.SetAssetsUnlockedSuccess(true)
				_ = s.extBridgeKeeper.SetMinBridgeOutAmount(s.ctx, testERC20Token.Bytes(), math.NewInt(200))
				return []interface{}{testERC20Token, big.NewInt(201), uint8(0), ethRecipient}
			},
			as:        s.account1.EvmAddr,
			basicPass: true,
			output:    []interface{}{true},
		},
	}

	s.RunMethodTestCasesWithKeepers(testcases, "bridgeOut")
}

func (s *BridgeOutTestSuite) TestBridgeOutERC20Execution() {
	testcases := []TestCase{
		{
			name: "successful ethereum bridge",
			run: func() []interface{} {
				// Setup all prerequisites
				s.extBridgeKeeper.SetAssetsUnlockedSuccess(true)

				return []interface{}{testERC20Token, big.NewInt(100), uint8(0), ethRecipient}
			},
			as:        s.account1.EvmAddr,
			basicPass: true,
			output:    []interface{}{true},
			postCheck: func() {
				// Verify burn was called
				// Verify AssetsUnlocked was called
				s.Require().True(s.extBridgeKeeper.AssetsUnlockedCalled())
				s.Require().NotNil(s.extBridgeKeeper.lastAssetsUnlocked)
				s.Require().Equal(testERC20Token.Hex(), s.extBridgeKeeper.lastAssetsUnlocked.Token)
				s.Require().Equal(ethRecipient, s.extBridgeKeeper.lastAssetsUnlocked.Recipient)
				s.Require().Equal(uint32(0), s.extBridgeKeeper.lastAssetsUnlocked.Chain)
			},
		},
		{
			name: "burn call failure",
			run: func() []interface{} {
				s.extBridgeKeeper.SetBurnError(errors.New("failed to execute ERC20 burnFrom call"))

				return []interface{}{testERC20Token, big.NewInt(100), uint8(0), ethRecipient}
			},
			as:          s.account1.EvmAddr,
			basicPass:   true,
			revert:      true,
			errContains: "failed to execute ERC20 burnFrom call",
		},
		{
			name: "AssetsUnlocked failure after burn",
			run: func() []interface{} {
				s.extBridgeKeeper.SetAssetsUnlockedSuccess(false)

				return []interface{}{testERC20Token, big.NewInt(100), uint8(0), ethRecipient}
			},
			as:          s.account1.EvmAddr,
			basicPass:   true,
			revert:      true,
			errContains: "failed to send AssetsUnlocked to bridge",
			postCheck:   func() {},
		},
		{
			name: "amount max uint256",
			run: func() []interface{} {
				veryLargeAmount := new(big.Int)
				veryLargeAmount.SetString("115792089237316195423570985008687907853269984665640564039457584007913129639935", 10) // Max uint256

				s.extBridgeKeeper.SetAssetsUnlockedSuccess(true)

				return []interface{}{testERC20Token, veryLargeAmount, uint8(0), ethRecipient}
			},
			as:        s.account1.EvmAddr,
			basicPass: true,
			output:    []interface{}{true}, // Should still succeed
		},
	}

	s.RunMethodTestCasesWithKeepers(testcases, "bridgeOut")
}

func (s *BridgeOutTestSuite) TestBridgeOutBitcoinExecution() {
	testcases := []TestCase{
		{
			name: "successful bitcoin bridge",
			run: func() []interface{} {
				// Setup authorization
				s.authzKeeper.SetAuthorization(
					s.account1.SdkAddr,
					sdk.AccAddress(bridgeAddress.Bytes()),
					assetsbridge.SendMsgURL,
					&banktypes.SendAuthorization{
						SpendLimit: sdk.NewCoins(sdk.NewCoin(evmtypes.DefaultEVMDenom, math.NewInt(1000))),
					},
					nil,
				)
				s.extBridgeKeeper.SetAssetsUnlockedSuccess(true)

				return []interface{}{testBTCToken, big.NewInt(100), uint8(1), btcRecipient}
			},
			as:        s.account1.EvmAddr,
			basicPass: true,
			output:    []interface{}{true},
			postCheck: func() {
				// Verify AssetsUnlocked was called
				s.Require().True(s.extBridgeKeeper.AssetsUnlockedCalled())
				s.Require().NotNil(s.extBridgeKeeper.lastAssetsUnlocked)
				s.Require().Equal(testBTCToken.Hex(), s.extBridgeKeeper.lastAssetsUnlocked.Token)
				s.Require().Equal(btcRecipient, s.extBridgeKeeper.lastAssetsUnlocked.Recipient)
				s.Require().Equal(uint32(1), s.extBridgeKeeper.lastAssetsUnlocked.Chain)
			},
		},
		{
			name: "transfer failure",
			run: func() []interface{} {
				s.extBridgeKeeper.SetBurnError(errors.New("burn failed"))
				s.authzKeeper.SetAuthorization(
					s.account1.SdkAddr,
					sdk.AccAddress(bridgeAddress.Bytes()),
					assetsbridge.SendMsgURL,
					&banktypes.SendAuthorization{
						SpendLimit: sdk.NewCoins(sdk.NewCoin(evmtypes.DefaultEVMDenom, math.NewInt(1000))),
					},
					nil,
				)

				return []interface{}{testBTCToken, big.NewInt(100), uint8(1), btcRecipient}
			},
			as:          s.account1.EvmAddr,
			basicPass:   true,
			revert:      true,
			errContains: "burn failed",
		},
		{
			name: "AssetsUnlocked failure after transfer",
			run: func() []interface{} {
				s.authzKeeper.SetAuthorization(
					s.account1.SdkAddr,
					sdk.AccAddress(bridgeAddress.Bytes()),
					assetsbridge.SendMsgURL,
					&banktypes.SendAuthorization{
						SpendLimit: sdk.NewCoins(sdk.NewCoin(evmtypes.DefaultEVMDenom, math.NewInt(1000))),
					},
					nil,
				)
				s.extBridgeKeeper.SetAssetsUnlockedSuccess(false)

				return []interface{}{testBTCToken, big.NewInt(100), uint8(1), btcRecipient}
			},
			as:          s.account1.EvmAddr,
			basicPass:   true,
			revert:      true,
			errContains: "failed to send AssetsUnlocked to bridge",
			postCheck:   func() {},
		},
		{
			name: "valid P2PKH bitcoin address",
			run: func() []interface{} {
				s.authzKeeper.SetAuthorization(
					s.account1.SdkAddr,
					sdk.AccAddress(bridgeAddress.Bytes()),
					assetsbridge.SendMsgURL,
					&banktypes.SendAuthorization{
						SpendLimit: sdk.NewCoins(sdk.NewCoin(evmtypes.DefaultEVMDenom, math.NewInt(1000))),
					},
					nil,
				)
				s.extBridgeKeeper.SetAssetsUnlockedSuccess(true)

				// Valid P2PKH address (starts with '1')
				p2pkhAddress := btcRecipient

				return []interface{}{testBTCToken, big.NewInt(100), uint8(1), p2pkhAddress}
			},
			as:        s.account1.EvmAddr,
			basicPass: true,
			output:    []interface{}{true},
		},
		{
			name: "valid P2SH bitcoin address",
			run: func() []interface{} {
				s.authzKeeper.SetAuthorization(
					s.account1.SdkAddr,
					sdk.AccAddress(bridgeAddress.Bytes()),
					assetsbridge.SendMsgURL,
					&banktypes.SendAuthorization{
						SpendLimit: sdk.NewCoins(sdk.NewCoin(evmtypes.DefaultEVMDenom, math.NewInt(1000))),
					},
					nil,
				)
				s.extBridgeKeeper.SetAssetsUnlockedSuccess(true)

				p2shAddress := []byte("3J98t1WpEZ73CNmQviecrnyiWrnqRhWNLy")
				recipient, _, _ := base58.CheckDecode(string(p2shAddress))
				recipient = append([]byte{0x17, 0xa9, 0x14}, recipient...)
				recipient = append(recipient, []byte{0x87}...)

				return []interface{}{testBTCToken, big.NewInt(100), uint8(1), recipient}
			},
			as:        s.account1.EvmAddr,
			basicPass: true,
			output:    []interface{}{true},
		},
		{
			name: "invalid bitcoin address",
			run: func() []interface{} {
				s.authzKeeper.SetAuthorization(
					s.account1.SdkAddr,
					sdk.AccAddress(bridgeAddress.Bytes()),
					assetsbridge.SendMsgURL,
					&banktypes.SendAuthorization{
						SpendLimit: sdk.NewCoins(sdk.NewCoin(evmtypes.DefaultEVMDenom, math.NewInt(1000))),
					},
					nil,
				)
				s.extBridgeKeeper.SetAssetsUnlockedSuccess(true)

				bech32Address := []byte("notvalidaddress")

				return []interface{}{testBTCToken, big.NewInt(100), uint8(1), bech32Address}
			},
			as:          s.account1.EvmAddr,
			revert:      true,
			basicPass:   true,
			errContains: "couldn't get script from var-len data",
			output:      []interface{}{false},
		},
	}

	s.RunMethodTestCasesWithKeepers(testcases, "bridgeOut")
}

func (s *BridgeOutTestSuite) TestBridgeOutBitcoinAuthorization() {
	testcases := []TestCase{
		{
			name: "missing authorization",
			run: func() []interface{} {
				s.extBridgeKeeper.SetAssetsUnlockedSuccess(true)
				return []interface{}{testBTCToken, big.NewInt(100), uint8(1), btcRecipient}
			},
			as:          s.account1.EvmAddr,
			basicPass:   true,
			revert:      true,
			errContains: "authorization type does not exist",
		},
		{
			name: "valid authorization",
			run: func() []interface{} {
				s.authzKeeper.SetAuthorization(
					s.account1.SdkAddr,
					sdk.AccAddress(bridgeAddress.Bytes()),
					assetsbridge.SendMsgURL,
					&banktypes.SendAuthorization{
						SpendLimit: sdk.NewCoins(sdk.NewCoin(evmtypes.DefaultEVMDenom, math.NewInt(1000))),
					},
					nil,
				)
				s.extBridgeKeeper.SetAssetsUnlockedSuccess(true)
				return []interface{}{testBTCToken, big.NewInt(100), uint8(1), btcRecipient}
			},
			as:        s.account1.EvmAddr,
			basicPass: true,
			output:    []interface{}{true},
		},
	}

	s.RunMethodTestCasesWithKeepers(testcases, "bridgeOut")
}

// Test input validation
func (s *BridgeOutTestSuite) TestBridgeOutInputValidation() {
	testcases := []TestCase{
		{
			name: "incorrect input count - less than 4",
			run: func() []interface{} {
				return []interface{}{testERC20Token, big.NewInt(100), uint8(0)}
			},
			basicPass:   false,
			errContains: "argument count mismatch",
		},
		{
			name: "incorrect input count - more than 4",
			run: func() []interface{} {
				return []interface{}{testERC20Token, big.NewInt(100), uint8(0), ethRecipient, "extra"}
			},
			basicPass:   false,
			errContains: "argument count mismatch",
		},
		{
			name: "invalid token type - string instead of address",
			run: func() []interface{} {
				return []interface{}{"invalid", big.NewInt(100), uint8(0), ethRecipient}
			},
			basicPass:   false,
			errContains: "cannot use string as type",
		},
		{
			name: "invalid amount type - string instead of big.Int",
			run: func() []interface{} {
				return []interface{}{testERC20Token, "100", uint8(0), ethRecipient}
			},
			basicPass:   false,
			errContains: "cannot use string as type",
		},
		{
			name: "zero amount",
			run: func() []interface{} {
				s.extBridgeKeeper.SetAssetsUnlockedSuccess(true)
				return []interface{}{testERC20Token, big.NewInt(0), uint8(0), ethRecipient}
			},
			as:          s.account1.EvmAddr,
			basicPass:   true,
			revert:      true,
			errContains: "amount must be positive",
		},
		{
			name: "invalid chain type - string instead of uint8",
			run: func() []interface{} {
				return []interface{}{testERC20Token, big.NewInt(100), "ethereum", ethRecipient}
			},
			basicPass:   false,
			errContains: "cannot use string as type",
		},
		{
			name: "unsupported chain value - 2",
			run: func() []interface{} {
				return []interface{}{testERC20Token, big.NewInt(100), uint8(2), ethRecipient}
			},
			as:          s.account1.EvmAddr,
			basicPass:   true,
			revert:      true,
			errContains: "unsupported chain",
		},
		{
			name: "invalid recipient type - string instead of bytes",
			run: func() []interface{} {
				return []interface{}{testERC20Token, big.NewInt(100), uint8(0), "recipient"}
			},
			basicPass:   false,
			errContains: "cannot use string as type",
		},
		{
			name: "empty recipient",
			run: func() []interface{} {
				return []interface{}{testERC20Token, big.NewInt(100), uint8(0), []byte{}}
			},
			as:          s.account1.EvmAddr,
			basicPass:   true,
			revert:      true,
			errContains: "recipient can't be empty",
		},
		{
			name: "ethereum recipient - incorrect length",
			run: func() []interface{} {
				return []interface{}{testERC20Token, big.NewInt(100), uint8(0), []byte{0x01, 0x02, 0x03}}
			},
			as:          s.account1.EvmAddr,
			basicPass:   true,
			revert:      true,
			errContains: "invalid recipient address for Ethereum chain",
		},
		{
			name: "bitcoin recipient - invalid address",
			run: func() []interface{} {
				return []interface{}{testERC20Token, big.NewInt(100), uint8(1), invalidBtcAddress}
			},
			as:          s.account1.EvmAddr,
			basicPass:   true,
			revert:      true,
			errContains: "couldn't get script from var-len dat",
		},
		{
			name: "valid ethereum inputs",
			run: func() []interface{} {
				s.extBridgeKeeper.SetAssetsUnlockedSuccess(true)
				return []interface{}{testERC20Token, big.NewInt(100), uint8(0), ethRecipient}
			},
			as:        s.account1.EvmAddr,
			basicPass: true,
			output:    []interface{}{true},
		},
		{
			name: "valid bitcoin inputs",
			run: func() []interface{} {
				// Setup authorization
				s.authzKeeper.SetAuthorization(
					s.account1.SdkAddr,
					sdk.AccAddress(bridgeAddress.Bytes()),
					assetsbridge.SendMsgURL,
					&banktypes.SendAuthorization{
						SpendLimit: sdk.NewCoins(sdk.NewCoin(evmtypes.DefaultEVMDenom, math.NewInt(1000))),
					},
					nil,
				)
				s.extBridgeKeeper.SetAssetsUnlockedSuccess(true)
				return []interface{}{testBTCToken, big.NewInt(100), uint8(1), btcRecipient}
			},
			as:        s.account1.EvmAddr,
			basicPass: true,
		},
	}

	s.RunMethodTestCasesWithKeepers(testcases, "bridgeOut")
}

func (s *BridgeOutTestSuite) RunMethodTestCasesWithKeepers(testcases []TestCase, methodName string) {
	for _, tc := range testcases {
		s.Run(tc.name, func() {
			// Reset keepers state
			s.extBridgeKeeper.Reset()

			evm := &vm.EVM{
				StateDB: statedb.New(s.ctx, statedb.NewMockKeeper(), statedb.TxConfig{}),
			}

			assetsBridgePrecompile, err := assetsbridge.NewPrecompile(
				s.poaKeeper,
				s.extBridgeKeeper,
				s.authzKeeper,
				&assetsbridge.Settings{
					Observability:   true,
					BTCManagement:   true,
					ERC20Management: true,
					SequenceTipView: true,
					BridgeOut:       true,
				},
			)
			s.Require().NoError(err)
			s.assetsBridgePrecompile = assetsBridgePrecompile

			var methodInputs []interface{}
			if tc.run != nil {
				methodInputs = tc.run()
			}

			method := s.assetsBridgePrecompile.Abi.Methods[methodName]
			var methodInputArgs []byte
			methodInputArgs, err = method.Inputs.Pack(methodInputs...)

			if tc.basicPass {
				s.Require().NoError(err, "expected no error")
			} else {
				s.Require().Error(err, "expected error")
				s.Require().ErrorContains(err, tc.errContains, "expected different error message")
				return
			}

			vmContract := vm.NewContract(&precompile.Contract{}, nil, nil, 0)
			vmContract.Input = append(vmContract.Input, method.ID...)
			vmContract.Input = append(vmContract.Input, methodInputArgs...)
			vmContract.CallerAddress = tc.as

			output, err := s.assetsBridgePrecompile.Run(evm, vmContract, false)

			if tc.revert {
				s.Require().Error(err, "expected error")
				s.Require().ErrorContains(err, tc.errContains, "expected different error message")
				return
			}
			s.Require().NoError(err, "expected no error")

			if tc.output != nil {
				// The output is packed as bytes, we need to unpack it
				out, err := method.Outputs.Unpack(output)
				s.Require().NoError(err)
				s.Require().Equal(tc.output, out)
			}

			if tc.postCheck != nil {
				tc.postCheck()
			}
		})
	}
}
