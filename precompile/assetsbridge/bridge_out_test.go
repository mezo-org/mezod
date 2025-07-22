package assetsbridge_test

import (
	"bytes"
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
	recipient = append([]byte{0x76, 0xa9, 0x14}, recipient...)
	return append(recipient, []byte{0x88, 0xac}...)
}

type FakeBankKeeper struct {
	balances map[string]sdk.Coins
}

func NewFakeBankKeeper() *FakeBankKeeper {
	return &FakeBankKeeper{
		balances: make(map[string]sdk.Coins),
	}
}

func (k *FakeBankKeeper) AllBalances(_ context.Context, req *banktypes.QueryAllBalancesRequest) (*banktypes.QueryAllBalancesResponse, error) {
	addr := req.Address
	if coins, ok := k.balances[addr]; ok {
		return &banktypes.QueryAllBalancesResponse{
			Balances: coins,
		}, nil
	}
	return &banktypes.QueryAllBalancesResponse{
		Balances: sdk.Coins{},
	}, nil
}

func (k *FakeBankKeeper) SetBalance(addr string, coins sdk.Coins) {
	k.balances[addr] = coins
}

type FakeEvmKeeper struct {
	contracts       map[string]bool
	callResponses   map[string]*evmtypes.MsgEthereumTxResponse
	shouldRevert    bool
	revertMessage   string
	balances        map[string]*big.Int
	allowances      map[string]map[string]*big.Int
	burnFromCalled  bool
	burnFromSuccess bool
	burnFromError   error
	lastBurnAmount  *big.Int
	lastBurnFrom    common.Address
	callHandler     func(evmtypes.ContractCall) (*evmtypes.MsgEthereumTxResponse, error)
}

func NewFakeEvmKeeper() *FakeEvmKeeper {
	return &FakeEvmKeeper{
		contracts:       make(map[string]bool),
		callResponses:   make(map[string]*evmtypes.MsgEthereumTxResponse),
		balances:        make(map[string]*big.Int),
		allowances:      make(map[string]map[string]*big.Int),
		burnFromSuccess: true,
	}
}

func (k *FakeEvmKeeper) ExecuteContractCall(_ sdk.Context, call evmtypes.ContractCall) (*evmtypes.MsgEthereumTxResponse, error) {
	if k.callHandler != nil {
		return k.callHandler(call)
	}

	// Handle different call types based on the method signature
	data := call.Data()

	// ERC20 balanceOf(address) - 0x70a08231
	if len(data) >= 4 && bytes.Equal(data[:4], []byte{0x70, 0xa0, 0x82, 0x31}) {
		if k.shouldRevert {
			return &evmtypes.MsgEthereumTxResponse{
				VmError: k.revertMessage,
			}, nil
		}
		if len(data) < 36 {
			return nil, errors.New("invalid balanceOf call data")
		}
		addr := common.BytesToAddress(data[16:36])
		balance := k.GetBalance(addr)

		// Pack the balance as 32 bytes
		ret := common.LeftPadBytes(balance.Bytes(), 32)
		return &evmtypes.MsgEthereumTxResponse{
			Ret: ret,
		}, nil
	}

	// ERC20 allowance(address,address) - 0xdd62ed3e
	if len(data) >= 4 && bytes.Equal(data[:4], []byte{0xdd, 0x62, 0xed, 0x3e}) {
		if k.shouldRevert {
			return &evmtypes.MsgEthereumTxResponse{
				VmError: k.revertMessage,
			}, nil
		}
		if len(data) < 68 {
			return nil, errors.New("invalid allowance call data")
		}
		owner := common.BytesToAddress(data[16:36])
		spender := common.BytesToAddress(data[48:68])
		allowance := k.GetAllowance(owner, spender)

		// Pack the allowance as 32 bytes
		ret := common.LeftPadBytes(allowance.Bytes(), 32)
		return &evmtypes.MsgEthereumTxResponse{
			Ret: ret,
		}, nil
	}

	// ERC20 burnFrom(address,uint256) - 0x79cc6790
	if len(data) >= 4 && bytes.Equal(data[:4], []byte{0x79, 0xcc, 0x67, 0x90}) {
		if k.shouldRevert {
			return &evmtypes.MsgEthereumTxResponse{
				VmError: k.revertMessage,
			}, nil
		}
		if len(data) < 68 {
			return nil, errors.New("invalid burnFrom call data")
		}
		k.burnFromCalled = true
		k.lastBurnFrom = common.BytesToAddress(data[16:36])
		k.lastBurnAmount = new(big.Int).SetBytes(data[36:68])

		if k.burnFromError != nil {
			return nil, k.burnFromError
		}
		if !k.burnFromSuccess {
			return nil, errors.New("burnFrom failed")
		}

		return &evmtypes.MsgEthereumTxResponse{}, nil
	}

	return nil, errors.New("unknown contract call")
}

func (k *FakeEvmKeeper) IsContract(_ sdk.Context, address []byte) bool {
	return k.contracts[common.BytesToAddress(address).Hex()]
}

func (k *FakeEvmKeeper) SetContract(address common.Address, isContract bool) {
	k.contracts[address.Hex()] = isContract
}

func (k *FakeEvmKeeper) SetBalance(addr common.Address, balance *big.Int) {
	k.balances[addr.Hex()] = balance
}

func (k *FakeEvmKeeper) GetBalance(addr common.Address) *big.Int {
	if balance, ok := k.balances[addr.Hex()]; ok {
		return balance
	}
	return big.NewInt(0)
}

func (k *FakeEvmKeeper) SetAllowance(owner, spender common.Address, allowance *big.Int) {
	if k.allowances[owner.Hex()] == nil {
		k.allowances[owner.Hex()] = make(map[string]*big.Int)
	}
	k.allowances[owner.Hex()][spender.Hex()] = allowance
}

func (k *FakeEvmKeeper) GetAllowance(owner, spender common.Address) *big.Int {
	if ownerAllowances, ok := k.allowances[owner.Hex()]; ok {
		if allowance, ok := ownerAllowances[spender.Hex()]; ok {
			return allowance
		}
	}
	return big.NewInt(0)
}

func (k *FakeEvmKeeper) SetShouldRevert(shouldRevert bool, message string) {
	k.shouldRevert = shouldRevert
	k.revertMessage = message
}

func (k *FakeEvmKeeper) SetBurnFromSuccess(success bool) {
	k.burnFromSuccess = success
}

func (k *FakeEvmKeeper) SetBurnFromError(err error) {
	k.burnFromError = err
}

func (k *FakeEvmKeeper) BurnFromCalled() bool {
	return k.burnFromCalled
}

func (k *FakeEvmKeeper) SetCallHandler(handler func(evmtypes.ContractCall) (*evmtypes.MsgEthereumTxResponse, error)) {
	k.callHandler = handler
}

func (k *FakeEvmKeeper) Reset() {
	k.burnFromCalled = false
	k.lastBurnAmount = nil
	k.lastBurnFrom = common.Address{}
	k.burnFromError = nil
	k.callHandler = nil
	k.shouldRevert = false
	k.revertMessage = ""
}

type FakeAuthzKeeper struct {
	authorizations   map[string]authz.Authorization
	expirations      map[string]*time.Time
	dispatchCalled   bool
	dispatchSuccess  bool
	lastDispatchMsgs []sdk.Msg
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

func (k *FakeAuthzKeeper) DispatchActions(_ context.Context, _ sdk.AccAddress, msgs []sdk.Msg) ([][]byte, error) {
	k.dispatchCalled = true
	k.lastDispatchMsgs = msgs

	if !k.dispatchSuccess {
		return nil, errors.New("dispatch failed")
	}

	results := make([][]byte, len(msgs))
	for i := range msgs {
		results[i] = []byte("success")
	}
	return results, nil
}

func (k *FakeAuthzKeeper) SetAuthorization(granter, grantee sdk.AccAddress, msgType string, auth authz.Authorization, expiration *time.Time) {
	key := fmt.Sprintf("%s-%s-%s", grantee.String(), granter.String(), msgType)
	k.authorizations[key] = auth
	k.expirations[key] = expiration
}

func (k *FakeAuthzKeeper) SetDispatchSuccess(success bool) {
	k.dispatchSuccess = success
}

func (k *FakeAuthzKeeper) DispatchCalled() bool {
	return k.dispatchCalled
}

func (k *FakeAuthzKeeper) Reset() {
	k.dispatchCalled = false
	k.lastDispatchMsgs = nil
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

func (k *ExtendedFakeBridgeKeeper) AssetsUnlocked(
	_ sdk.Context,
	token []byte,
	amount math.Int,
	chain uint8,
	recipient []byte,
) (*bridgetypes.AssetsUnlockedEvent, error) {
	k.assetsUnlockedCalled = true

	if !k.assetsUnlockedSuccess {
		return nil, errors.New("AssetsUnlocked failed")
	}

	event := &bridgetypes.AssetsUnlockedEvent{
		Token:          common.BytesToAddress(token).Hex(),
		Amount:         amount,
		Chain:          uint32(chain),
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
}

type BridgeOutTestSuite struct {
	PrecompileTestSuite

	bankKeeper      *FakeBankKeeper
	evmKeeper       *FakeEvmKeeper
	authzKeeper     *FakeAuthzKeeper
	extBridgeKeeper *ExtendedFakeBridgeKeeper
}

func TestBridgeOutTestSuite(t *testing.T) {
	suite.Run(t, new(BridgeOutTestSuite))
}

func (s *BridgeOutTestSuite) SetupTest() {
	s.PrecompileTestSuite.SetupTest()

	s.bankKeeper = NewFakeBankKeeper()
	s.evmKeeper = NewFakeEvmKeeper()
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
		s.bankKeeper,
		s.evmKeeper,
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
				// Token is already registered in SetupTest
				s.evmKeeper.SetBalance(s.account1.EvmAddr, big.NewInt(1000))
				s.evmKeeper.SetAllowance(s.account1.EvmAddr, bridgeAddress, big.NewInt(1000))
				s.evmKeeper.SetBurnFromSuccess(true)
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

func (s *BridgeOutTestSuite) TestBridgeOutERC20Execution() {
	testcases := []TestCase{
		{
			name: "successful ethereum bridge",
			run: func() []interface{} {
				// Setup all prerequisites
				s.evmKeeper.SetBalance(s.account1.EvmAddr, big.NewInt(1000))
				s.evmKeeper.SetAllowance(s.account1.EvmAddr, bridgeAddress, big.NewInt(1000))
				s.evmKeeper.SetBurnFromSuccess(true)
				s.extBridgeKeeper.SetAssetsUnlockedSuccess(true)

				return []interface{}{testERC20Token, big.NewInt(100), uint8(0), ethRecipient}
			},
			as:        s.account1.EvmAddr,
			basicPass: true,
			output:    []interface{}{true},
			postCheck: func() {
				// Verify burn was called
				s.Require().True(s.evmKeeper.BurnFromCalled())
				s.Require().Equal(big.NewInt(100), s.evmKeeper.lastBurnAmount)
				s.Require().Equal(s.account1.EvmAddr, s.evmKeeper.lastBurnFrom)
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
				s.evmKeeper.SetBalance(s.account1.EvmAddr, big.NewInt(1000))
				s.evmKeeper.SetAllowance(s.account1.EvmAddr, bridgeAddress, big.NewInt(1000))
				s.evmKeeper.SetBurnFromSuccess(false)

				return []interface{}{testERC20Token, big.NewInt(100), uint8(0), ethRecipient}
			},
			as:          s.account1.EvmAddr,
			basicPass:   true,
			revert:      true,
			errContains: "failed to execute ERC20 burnFrom call",
		},
		{
			name: "burn call revert",
			run: func() []interface{} {
				s.evmKeeper.SetBalance(s.account1.EvmAddr, big.NewInt(1000))
				s.evmKeeper.SetAllowance(s.account1.EvmAddr, bridgeAddress, big.NewInt(1000))
				s.evmKeeper.SetBurnFromSuccess(false)
				s.evmKeeper.SetBurnFromError(fmt.Errorf("failed to execute ERC20 burnFrom call: burn reverted"))

				return []interface{}{testERC20Token, big.NewInt(100), uint8(0), ethRecipient}
			},
			as:          s.account1.EvmAddr,
			basicPass:   true,
			revert:      true,
			errContains: "failed to execute ERC20 burnFrom call: burn reverted",
		},
		{
			name: "AssetsUnlocked failure after burn",
			run: func() []interface{} {
				s.evmKeeper.SetBalance(s.account1.EvmAddr, big.NewInt(1000))
				s.evmKeeper.SetAllowance(s.account1.EvmAddr, bridgeAddress, big.NewInt(1000))
				s.evmKeeper.SetBurnFromSuccess(true)
				s.extBridgeKeeper.SetAssetsUnlockedSuccess(false)

				return []interface{}{testERC20Token, big.NewInt(100), uint8(0), ethRecipient}
			},
			as:          s.account1.EvmAddr,
			basicPass:   true,
			revert:      true,
			errContains: "failed to send AssetsUnlocked to bridge",
			postCheck: func() {
				// Verify burn was still called (critical: funds may be lost)
				s.Require().True(s.evmKeeper.BurnFromCalled())
			},
		},
		{
			name: "amount max uint256",
			run: func() []interface{} {
				veryLargeAmount := new(big.Int)
				veryLargeAmount.SetString("115792089237316195423570985008687907853269984665640564039457584007913129639935", 10) // Max uint256

				s.evmKeeper.SetBalance(s.account1.EvmAddr, veryLargeAmount)
				s.evmKeeper.SetAllowance(s.account1.EvmAddr, bridgeAddress, veryLargeAmount)
				s.evmKeeper.SetBurnFromSuccess(true)
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
				// Setup ERC20 allowance first
				s.evmKeeper.SetBalance(s.account1.EvmAddr, big.NewInt(1000))
				s.evmKeeper.SetAllowance(s.account1.EvmAddr, bridgeAddress, big.NewInt(1000))

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
				s.authzKeeper.SetDispatchSuccess(true)
				s.extBridgeKeeper.SetAssetsUnlockedSuccess(true)

				return []interface{}{testBTCToken, big.NewInt(100), uint8(1), btcRecipient}
			},
			as:        s.account1.EvmAddr,
			basicPass: true,
			output:    []interface{}{true},
			postCheck: func() {
				// Verify dispatch was called
				s.Require().True(s.authzKeeper.DispatchCalled())
				s.Require().Len(s.authzKeeper.lastDispatchMsgs, 1)

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
				// Setup ERC20 allowance first
				s.evmKeeper.SetBalance(s.account1.EvmAddr, big.NewInt(1000))
				s.evmKeeper.SetAllowance(s.account1.EvmAddr, bridgeAddress, big.NewInt(1000))

				s.authzKeeper.SetAuthorization(
					s.account1.SdkAddr,
					sdk.AccAddress(bridgeAddress.Bytes()),
					assetsbridge.SendMsgURL,
					&banktypes.SendAuthorization{
						SpendLimit: sdk.NewCoins(sdk.NewCoin(evmtypes.DefaultEVMDenom, math.NewInt(1000))),
					},
					nil,
				)
				s.authzKeeper.SetDispatchSuccess(false)

				return []interface{}{testBTCToken, big.NewInt(100), uint8(1), btcRecipient}
			},
			as:          s.account1.EvmAddr,
			basicPass:   true,
			revert:      true,
			errContains: "dispatch failed",
		},
		{
			name: "AssetsUnlocked failure after transfer",
			run: func() []interface{} {
				// Setup ERC20 allowance first
				s.evmKeeper.SetBalance(s.account1.EvmAddr, big.NewInt(1000))
				s.evmKeeper.SetAllowance(s.account1.EvmAddr, bridgeAddress, big.NewInt(1000))

				s.authzKeeper.SetAuthorization(
					s.account1.SdkAddr,
					sdk.AccAddress(bridgeAddress.Bytes()),
					assetsbridge.SendMsgURL,
					&banktypes.SendAuthorization{
						SpendLimit: sdk.NewCoins(sdk.NewCoin(evmtypes.DefaultEVMDenom, math.NewInt(1000))),
					},
					nil,
				)
				s.authzKeeper.SetDispatchSuccess(true)
				s.extBridgeKeeper.SetAssetsUnlockedSuccess(false)

				return []interface{}{testBTCToken, big.NewInt(100), uint8(1), btcRecipient}
			},
			as:          s.account1.EvmAddr,
			basicPass:   true,
			revert:      true,
			errContains: "failed to send AssetsUnlocked to bridge",
			postCheck: func() {
				// Verify transfer was still called (critical: funds may be lost)
				s.Require().True(s.authzKeeper.DispatchCalled())
			},
		},
		{
			name: "valid P2PKH bitcoin address",
			run: func() []interface{} {
				// Setup ERC20 allowance first
				s.evmKeeper.SetBalance(s.account1.EvmAddr, big.NewInt(1000))
				s.evmKeeper.SetAllowance(s.account1.EvmAddr, bridgeAddress, big.NewInt(1000))

				s.authzKeeper.SetAuthorization(
					s.account1.SdkAddr,
					sdk.AccAddress(bridgeAddress.Bytes()),
					assetsbridge.SendMsgURL,
					&banktypes.SendAuthorization{
						SpendLimit: sdk.NewCoins(sdk.NewCoin(evmtypes.DefaultEVMDenom, math.NewInt(1000))),
					},
					nil,
				)
				s.authzKeeper.SetDispatchSuccess(true)
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
				// Setup ERC20 allowance first
				s.evmKeeper.SetBalance(s.account1.EvmAddr, big.NewInt(1000))
				s.evmKeeper.SetAllowance(s.account1.EvmAddr, bridgeAddress, big.NewInt(1000))

				s.authzKeeper.SetAuthorization(
					s.account1.SdkAddr,
					sdk.AccAddress(bridgeAddress.Bytes()),
					assetsbridge.SendMsgURL,
					&banktypes.SendAuthorization{
						SpendLimit: sdk.NewCoins(sdk.NewCoin(evmtypes.DefaultEVMDenom, math.NewInt(1000))),
					},
					nil,
				)
				s.authzKeeper.SetDispatchSuccess(true)
				s.extBridgeKeeper.SetAssetsUnlockedSuccess(true)

				p2shAddress := []byte("3J98t1WpEZ73CNmQviecrnyiWrnqRhWNLy")
				recipient, _, _ := base58.CheckDecode(string(p2shAddress))
				recipient = append([]byte{0xa9, 0x14}, recipient...)
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
				// Setup ERC20 allowance first
				s.evmKeeper.SetBalance(s.account1.EvmAddr, big.NewInt(1000))
				s.evmKeeper.SetAllowance(s.account1.EvmAddr, bridgeAddress, big.NewInt(1000))

				s.authzKeeper.SetAuthorization(
					s.account1.SdkAddr,
					sdk.AccAddress(bridgeAddress.Bytes()),
					assetsbridge.SendMsgURL,
					&banktypes.SendAuthorization{
						SpendLimit: sdk.NewCoins(sdk.NewCoin(evmtypes.DefaultEVMDenom, math.NewInt(1000))),
					},
					nil,
				)
				s.authzKeeper.SetDispatchSuccess(true)
				s.extBridgeKeeper.SetAssetsUnlockedSuccess(true)

				bech32Address := []byte("notvalidaddress")

				return []interface{}{testBTCToken, big.NewInt(100), uint8(1), bech32Address}
			},
			as:          s.account1.EvmAddr,
			revert:      true,
			basicPass:   true,
			errContains: "invalid recipient address format for Bitcoin",
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
				// No authorization setup, but still need to pass ERC20 allowance check
				s.evmKeeper.SetBalance(s.account1.EvmAddr, big.NewInt(1000))
				s.evmKeeper.SetAllowance(s.account1.EvmAddr, bridgeAddress, big.NewInt(1000))
				return []interface{}{testBTCToken, big.NewInt(100), uint8(1), btcRecipient}
			},
			as:          s.account1.EvmAddr,
			basicPass:   true,
			revert:      true,
			errContains: "authorization type does not exist",
		},
		{
			name: "expired authorization",
			run: func() []interface{} {
				// Setup ERC20 allowance first
				s.evmKeeper.SetBalance(s.account1.EvmAddr, big.NewInt(1000))
				s.evmKeeper.SetAllowance(s.account1.EvmAddr, bridgeAddress, big.NewInt(1000))

				expiredTime := s.ctx.BlockTime().Add(-1 * time.Hour)
				s.authzKeeper.SetAuthorization(
					s.account1.SdkAddr,
					sdk.AccAddress(bridgeAddress.Bytes()),
					assetsbridge.SendMsgURL,
					&banktypes.SendAuthorization{
						SpendLimit: sdk.NewCoins(sdk.NewCoin(evmtypes.DefaultEVMDenom, math.NewInt(1000))),
					},
					&expiredTime,
				)
				return []interface{}{testBTCToken, big.NewInt(100), uint8(1), btcRecipient}
			},
			as:          s.account1.EvmAddr,
			basicPass:   true,
			revert:      true,
			errContains: "authorization expired",
		},
		// not sure we need this, not sure how that could happen,
		// but probably better to seeing that we use cast internally
		// to figure those types.
		{
			name: "wrong authorization type",
			run: func() []interface{} {
				// Setup ERC20 allowance first
				s.evmKeeper.SetBalance(s.account1.EvmAddr, big.NewInt(1000))
				s.evmKeeper.SetAllowance(s.account1.EvmAddr, bridgeAddress, big.NewInt(1000))

				// Set a different type of authorization (not SendAuthorization)
				s.authzKeeper.SetAuthorization(
					s.account1.SdkAddr,
					sdk.AccAddress(bridgeAddress.Bytes()),
					assetsbridge.SendMsgURL,
					&authz.GenericAuthorization{Msg: "/cosmos.bank.v1beta1.MsgSend"},
					nil,
				)
				return []interface{}{testBTCToken, big.NewInt(100), uint8(1), btcRecipient}
			},
			as:          s.account1.EvmAddr,
			basicPass:   true,
			revert:      true,
			errContains: "expected authorization to be a",
		},
		{
			name: "insufficient spend limit",
			run: func() []interface{} {
				// Setup ERC20 allowance first
				s.evmKeeper.SetBalance(s.account1.EvmAddr, big.NewInt(1000))
				s.evmKeeper.SetAllowance(s.account1.EvmAddr, bridgeAddress, big.NewInt(1000))

				s.authzKeeper.SetAuthorization(
					s.account1.SdkAddr,
					sdk.AccAddress(bridgeAddress.Bytes()),
					assetsbridge.SendMsgURL,
					&banktypes.SendAuthorization{
						SpendLimit: sdk.NewCoins(sdk.NewCoin(evmtypes.DefaultEVMDenom, math.NewInt(50))),
					},
					nil,
				)
				return []interface{}{testBTCToken, big.NewInt(100), uint8(1), btcRecipient}
			},
			as:          s.account1.EvmAddr,
			basicPass:   true,
			revert:      true,
			errContains: "requested amount",
		},
		{
			name: "nil spend limit",
			run: func() []interface{} {
				// Setup ERC20 allowance first
				s.evmKeeper.SetBalance(s.account1.EvmAddr, big.NewInt(1000))
				s.evmKeeper.SetAllowance(s.account1.EvmAddr, bridgeAddress, big.NewInt(1000))

				s.authzKeeper.SetAuthorization(
					s.account1.SdkAddr,
					sdk.AccAddress(bridgeAddress.Bytes()),
					assetsbridge.SendMsgURL,
					&banktypes.SendAuthorization{
						SpendLimit: nil,
					},
					nil,
				)
				return []interface{}{testBTCToken, big.NewInt(100), uint8(1), btcRecipient}
			},
			as:          s.account1.EvmAddr,
			basicPass:   true,
			revert:      true,
			errContains: "no allowance",
		},
		{
			name: "empty spend limit",
			run: func() []interface{} {
				// Setup ERC20 allowance first
				s.evmKeeper.SetBalance(s.account1.EvmAddr, big.NewInt(1000))
				s.evmKeeper.SetAllowance(s.account1.EvmAddr, bridgeAddress, big.NewInt(1000))

				s.authzKeeper.SetAuthorization(
					s.account1.SdkAddr,
					sdk.AccAddress(bridgeAddress.Bytes()),
					assetsbridge.SendMsgURL,
					&banktypes.SendAuthorization{
						SpendLimit: sdk.Coins{},
					},
					nil,
				)
				return []interface{}{testBTCToken, big.NewInt(100), uint8(1), btcRecipient}
			},
			as:          s.account1.EvmAddr,
			basicPass:   true,
			revert:      true,
			errContains: "no allowance",
		},
		{
			name: "wrong denomination",
			run: func() []interface{} {
				// Setup ERC20 allowance first
				s.evmKeeper.SetBalance(s.account1.EvmAddr, big.NewInt(1000))
				s.evmKeeper.SetAllowance(s.account1.EvmAddr, bridgeAddress, big.NewInt(1000))

				s.authzKeeper.SetAuthorization(
					s.account1.SdkAddr,
					sdk.AccAddress(bridgeAddress.Bytes()),
					assetsbridge.SendMsgURL,
					&banktypes.SendAuthorization{
						SpendLimit: sdk.NewCoins(sdk.NewCoin("wrongdenom", math.NewInt(1000))),
					},
					nil,
				)
				return []interface{}{testBTCToken, big.NewInt(100), uint8(1), btcRecipient}
			},
			as:          s.account1.EvmAddr,
			basicPass:   true,
			revert:      true,
			errContains: "no allowance for",
		},
		{
			name: "address not in allow list",
			run: func() []interface{} {
				// Setup ERC20 allowance first
				s.evmKeeper.SetBalance(s.account1.EvmAddr, big.NewInt(1000))
				s.evmKeeper.SetAllowance(s.account1.EvmAddr, bridgeAddress, big.NewInt(1000))

				s.authzKeeper.SetAuthorization(
					s.account1.SdkAddr,
					sdk.AccAddress(bridgeAddress.Bytes()),
					assetsbridge.SendMsgURL,
					&banktypes.SendAuthorization{
						SpendLimit: sdk.NewCoins(sdk.NewCoin(evmtypes.DefaultEVMDenom, math.NewInt(1000))),
						AllowList:  []string{"cosmos1differentaddress"},
					},
					nil,
				)
				return []interface{}{testBTCToken, big.NewInt(100), uint8(1), btcRecipient}
			},
			as:          s.account1.EvmAddr,
			basicPass:   true,
			revert:      true,
			errContains: "recipient address not in authorization allow list",
		},
		{
			name: "valid authorization",
			run: func() []interface{} {
				// Setup ERC20 allowance first
				s.evmKeeper.SetBalance(s.account1.EvmAddr, big.NewInt(1000))
				s.evmKeeper.SetAllowance(s.account1.EvmAddr, bridgeAddress, big.NewInt(1000))

				s.authzKeeper.SetAuthorization(
					s.account1.SdkAddr,
					sdk.AccAddress(bridgeAddress.Bytes()),
					assetsbridge.SendMsgURL,
					&banktypes.SendAuthorization{
						SpendLimit: sdk.NewCoins(sdk.NewCoin(evmtypes.DefaultEVMDenom, math.NewInt(1000))),
					},
					nil,
				)
				s.authzKeeper.SetDispatchSuccess(true)
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

// Test balance and allowance validation
func (s *BridgeOutTestSuite) TestBridgeOutBalanceAllowance() {
	testcases := []TestCase{
		{
			name: "exact balance and allowance",
			run: func() []interface{} {
				s.evmKeeper.SetBalance(s.account1.EvmAddr, big.NewInt(100))
				s.evmKeeper.SetAllowance(s.account1.EvmAddr, bridgeAddress, big.NewInt(100))
				s.evmKeeper.SetBurnFromSuccess(true)
				s.extBridgeKeeper.SetAssetsUnlockedSuccess(true)
				return []interface{}{testERC20Token, big.NewInt(100), uint8(0), ethRecipient}
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
				s.evmKeeper.SetBalance(s.account1.EvmAddr, big.NewInt(1000))
				s.evmKeeper.SetAllowance(s.account1.EvmAddr, bridgeAddress, big.NewInt(1000))
				s.evmKeeper.SetBurnFromSuccess(true)
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
			errContains: "invalid recipient address format for Ethereum chain",
		},
		{
			name: "bitcoin recipient - invalid address",
			run: func() []interface{} {
				return []interface{}{testERC20Token, big.NewInt(100), uint8(1), invalidBtcAddress}
			},
			as:          s.account1.EvmAddr,
			basicPass:   true,
			revert:      true,
			errContains: "invalid recipient address format for Bitcoin",
		},
		{
			name: "valid ethereum inputs",
			run: func() []interface{} {
				// Setup valid token and balances
				s.evmKeeper.SetBalance(s.account1.EvmAddr, big.NewInt(1000))
				s.evmKeeper.SetAllowance(s.account1.EvmAddr, bridgeAddress, big.NewInt(1000))
				s.evmKeeper.SetBurnFromSuccess(true)
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
				s.authzKeeper.SetDispatchSuccess(true)
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
			s.evmKeeper.Reset()
			s.authzKeeper.Reset()
			s.extBridgeKeeper.Reset()

			evm := &vm.EVM{
				StateDB: statedb.New(s.ctx, statedb.NewMockKeeper(), statedb.TxConfig{}),
			}

			assetsBridgePrecompile, err := assetsbridge.NewPrecompile(
				s.poaKeeper,
				s.extBridgeKeeper,
				s.bankKeeper,
				s.evmKeeper,
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
