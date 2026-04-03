package assetsbridge_test

import (
	"math/big"

	"cosmossdk.io/math"
	"github.com/ethereum/go-ethereum/common"
	bridgekeeper "github.com/mezo-org/mezod/x/bridge/keeper"
)

var testTripartyController = common.HexToAddress("0x1234567890AbCdEf1234567890AbCdEf12345678")

func (s *PrecompileTestSuite) TestAllowTripartyController() {
	testcases := []TestCase{
		{
			name: "caller is not owner",
			run: func() []interface{} {
				return []interface{}{testTripartyController, true}
			},
			as:          s.account2.EvmAddr,
			basicPass:   true,
			revert:      true,
			errContains: "sender is not owner",
		},
		{
			name: "zero address controller",
			run: func() []interface{} {
				return []interface{}{common.Address{}, true}
			},
			as:          s.account1.EvmAddr,
			basicPass:   true,
			revert:      true,
			errContains: "controller address must not be the zero address",
		},
		{
			name: "happy path - allow controller",
			run: func() []interface{} {
				return []interface{}{testTripartyController, true}
			},
			as:        s.account1.EvmAddr,
			basicPass: true,
			output:    []interface{}{true},
			postCheck: func() {
				s.Require().True(
					s.bridgeKeeper.IsAllowedTripartyController(s.ctx, testTripartyController.Bytes()),
				)
			},
		},
		{
			name: "happy path - disallow controller",
			run: func() []interface{} {
				s.bridgeKeeper.AllowTripartyController(s.ctx, testTripartyController.Bytes(), true)
				return []interface{}{testTripartyController, false}
			},
			as:        s.account1.EvmAddr,
			basicPass: true,
			output:    []interface{}{true},
			postCheck: func() {
				s.Require().False(
					s.bridgeKeeper.IsAllowedTripartyController(s.ctx, testTripartyController.Bytes()),
				)
			},
		},
	}

	s.RunMethodTestCases(testcases, "allowTripartyController")
}

func (s *PrecompileTestSuite) TestIsAllowedTripartyController() {
	testcases := []TestCase{
		{
			name: "not allowed - returns false",
			run: func() []interface{} {
				return []interface{}{testTripartyController}
			},
			as:        s.account1.EvmAddr,
			basicPass: true,
			output:    []interface{}{false},
		},
		{
			name: "allowed - returns true",
			run: func() []interface{} {
				s.bridgeKeeper.AllowTripartyController(s.ctx, testTripartyController.Bytes(), true)
				return []interface{}{testTripartyController}
			},
			as:        s.account1.EvmAddr,
			basicPass: true,
			output:    []interface{}{true},
		},
	}

	s.RunMethodTestCases(testcases, "isAllowedTripartyController")
}

func (s *PrecompileTestSuite) TestPauseTriparty() {
	testcases := []TestCase{
		{
			name: "no pauser is set",
			run: func() []interface{} {
				return []interface{}{true}
			},
			as:          testPauserAddr,
			basicPass:   true,
			revert:      true,
			errContains: "no pauser is set",
		},
		{
			name: "caller is not the pauser",
			run: func() []interface{} {
				s.bridgeKeeper.SetPauser(s.ctx, testPauserAddr.Bytes())
				return []interface{}{true}
			},
			as:          s.account1.EvmAddr,
			basicPass:   true,
			revert:      true,
			errContains: "caller is not the pauser",
		},
		{
			name: "happy path - pause",
			run: func() []interface{} {
				s.bridgeKeeper.SetPauser(s.ctx, testPauserAddr.Bytes())
				return []interface{}{true}
			},
			as:        testPauserAddr,
			basicPass: true,
			output:    []interface{}{true},
			postCheck: func() {
				s.Require().True(s.bridgeKeeper.tripartyPaused)
			},
		},
		{
			name: "happy path - unpause",
			run: func() []interface{} {
				s.bridgeKeeper.SetPauser(s.ctx, testPauserAddr.Bytes())
				s.bridgeKeeper.SetTripartyPaused(s.ctx, true)
				return []interface{}{false}
			},
			as:        testPauserAddr,
			basicPass: true,
			output:    []interface{}{true},
			postCheck: func() {
				s.Require().False(s.bridgeKeeper.tripartyPaused)
			},
		},
	}

	s.RunMethodTestCases(testcases, "pauseTriparty")
}

func (s *PrecompileTestSuite) TestBridgeTriparty() {
	testcases := []TestCase{
		{
			name: "happy path - with callbackData",
			run: func() []interface{} {
				s.bridgeKeeper.AllowTripartyController(s.ctx, testTripartyController.Bytes(), true)
				return []interface{}{s.account2.EvmAddr, bridgekeeper.MinTripartyAmount.BigInt(), []byte("callback")}
			},
			as:        testTripartyController,
			basicPass: true,
			output:    []interface{}{big.NewInt(1)},
			postCheck: func() {
				params := s.bridgeKeeper.lastTripartyBridgeRequestParams
				s.Require().NotNil(params)
				s.Require().Equal(s.account2.EvmAddr.Hex(), params.recipient)
				s.Require().True(bridgekeeper.MinTripartyAmount.Equal(params.amount))
				s.Require().Equal([]byte("callback"), params.callbackData)
				s.Require().Equal(testTripartyController.Hex(), params.controller)
			},
		},
		{
			name: "happy path - with empty callbackData",
			run: func() []interface{} {
				s.bridgeKeeper.AllowTripartyController(s.ctx, testTripartyController.Bytes(), true)
				doubleMin := new(big.Int).Mul(bridgekeeper.MinTripartyAmount.BigInt(), big.NewInt(2))
				return []interface{}{s.account2.EvmAddr, doubleMin, []byte{}}
			},
			as:        testTripartyController,
			basicPass: true,
			output:    []interface{}{big.NewInt(2)},
			postCheck: func() {
				params := s.bridgeKeeper.lastTripartyBridgeRequestParams
				s.Require().NotNil(params)
				s.Require().Equal(s.account2.EvmAddr.Hex(), params.recipient)
				s.Require().True(bridgekeeper.MinTripartyAmount.MulRaw(2).Equal(params.amount))
				s.Require().Empty(params.callbackData)
				s.Require().Equal(testTripartyController.Hex(), params.controller)
			},
		},
	}

	s.RunMethodTestCases(testcases, "bridgeTriparty")
}

func (s *PrecompileTestSuite) TestBridgeTripartyInvalidInputs() {
	testcases := []TestCase{
		{
			name: "invalid recipient type",
			run: func() []interface{} {
				return []interface{}{"not-an-address", big.NewInt(1), []byte{}}
			},
			as:        testTripartyController,
			basicPass: false,
		},
		{
			name: "invalid amount type",
			run: func() []interface{} {
				return []interface{}{s.account2.EvmAddr, "not-a-number", []byte{}}
			},
			as:        testTripartyController,
			basicPass: false,
		},
		{
			name: "invalid callbackData type",
			run: func() []interface{} {
				return []interface{}{s.account2.EvmAddr, big.NewInt(1), 123}
			},
			as:        testTripartyController,
			basicPass: false,
		},
		{
			name: "wrong number of inputs - too few",
			run: func() []interface{} {
				return []interface{}{s.account2.EvmAddr}
			},
			as:        testTripartyController,
			basicPass: false,
		},
		{
			name: "wrong number of inputs - too many",
			run: func() []interface{} {
				return []interface{}{s.account2.EvmAddr, big.NewInt(1), []byte{}, "extra"}
			},
			as:        testTripartyController,
			basicPass: false,
		},
	}

	s.RunMethodTestCases(testcases, "bridgeTriparty")
}

func (s *PrecompileTestSuite) TestSetTripartyBlockDelay() {
	testcases := []TestCase{
		{
			name: "caller is not owner",
			run: func() []interface{} {
				return []interface{}{big.NewInt(5)}
			},
			as:          s.account2.EvmAddr,
			basicPass:   true,
			revert:      true,
			errContains: "sender is not owner",
		},
		{
			name: "delay is zero",
			run: func() []interface{} {
				return []interface{}{big.NewInt(0)}
			},
			as:          s.account1.EvmAddr,
			basicPass:   true,
			revert:      true,
			errContains: "delay must be at least 1",
		},
		{
			name: "invalid delay type",
			run: func() []interface{} {
				return []interface{}{"invalid delay"}
			},
			as:        s.account1.EvmAddr,
			basicPass: false,
		},
		{
			name: "wrong number of inputs",
			run: func() []interface{} {
				return []interface{}{}
			},
			as:        s.account1.EvmAddr,
			basicPass: false,
		},
		{
			name: "happy path - set delay to 5",
			run: func() []interface{} {
				return []interface{}{big.NewInt(5)}
			},
			as:        s.account1.EvmAddr,
			basicPass: true,
			output:    []interface{}{true},
			postCheck: func() {
				s.Require().Equal(
					int64(5),
					s.bridgeKeeper.GetTripartyBlockDelay(s.ctx),
				)
			},
		},
	}

	s.RunMethodTestCases(testcases, "setTripartyBlockDelay")
}

func (s *PrecompileTestSuite) TestGetTripartyBlockDelay() {
	testcases := []TestCase{
		{
			name: "default value - returns 1",
			run: func() []interface{} {
				return []interface{}{}
			},
			as:        s.account1.EvmAddr,
			basicPass: true,
			output:    []interface{}{big.NewInt(1)},
		},
		{
			name: "returns set value",
			run: func() []interface{} {
				//nolint:errcheck
				s.bridgeKeeper.SetTripartyBlockDelay(s.ctx, 5)
				return []interface{}{}
			},
			as:        s.account1.EvmAddr,
			basicPass: true,
			output:    []interface{}{big.NewInt(5)},
		},
		{
			name: "wrong number of inputs",
			run: func() []interface{} {
				return []interface{}{big.NewInt(1)}
			},
			as:        s.account1.EvmAddr,
			basicPass: false,
		},
	}

	s.RunMethodTestCases(testcases, "getTripartyBlockDelay")
}

func (s *PrecompileTestSuite) TestSetTripartyLimits() {
	testcases := []TestCase{
		{
			name: "caller is not owner",
			run: func() []interface{} {
				return []interface{}{big.NewInt(100), big.NewInt(1000)}
			},
			as:          s.account2.EvmAddr,
			basicPass:   true,
			revert:      true,
			errContains: "sender is not owner",
		},
		{
			name: "invalid per-request limit type",
			run: func() []interface{} {
				return []interface{}{"invalid limit", big.NewInt(1000)}
			},
			as:        s.account1.EvmAddr,
			basicPass: false,
		},
		{
			name: "invalid window limit type",
			run: func() []interface{} {
				return []interface{}{big.NewInt(100), "invalid limit"}
			},
			as:        s.account1.EvmAddr,
			basicPass: false,
		},
		{
			name: "wrong number of inputs",
			run: func() []interface{} {
				return []interface{}{big.NewInt(100)}
			},
			as:        s.account1.EvmAddr,
			basicPass: false,
		},
		{
			name: "happy path - set limits",
			run: func() []interface{} {
				return []interface{}{big.NewInt(100), big.NewInt(1000)}
			},
			as:        s.account1.EvmAddr,
			basicPass: true,
			output:    []interface{}{true},
			postCheck: func() {
				s.Require().Equal(
					int64(100),
					s.bridgeKeeper.GetTripartyPerRequestLimit(s.ctx).Int64(),
				)
				s.Require().Equal(
					int64(1000),
					s.bridgeKeeper.GetTripartyWindowLimit(s.ctx).Int64(),
				)
			},
		},
		{
			name: "happy path - set zero limits",
			run: func() []interface{} {
				return []interface{}{big.NewInt(0), big.NewInt(0)}
			},
			as:        s.account1.EvmAddr,
			basicPass: true,
			output:    []interface{}{true},
			postCheck: func() {
				s.Require().True(
					s.bridgeKeeper.GetTripartyPerRequestLimit(s.ctx).IsZero(),
				)
				s.Require().True(
					s.bridgeKeeper.GetTripartyWindowLimit(s.ctx).IsZero(),
				)
			},
		},
	}

	s.RunMethodTestCases(testcases, "setTripartyLimits")
}

func (s *PrecompileTestSuite) TestGetTripartyLimits() {
	testcases := []TestCase{
		{
			name: "default values - both zero",
			run: func() []interface{} {
				return []interface{}{}
			},
			as:        s.account1.EvmAddr,
			basicPass: true,
			output:    []interface{}{big.NewInt(0), big.NewInt(0)},
		},
		{
			name: "returns set values",
			run: func() []interface{} {
				s.bridgeKeeper.SetTripartyPerRequestLimit(s.ctx, math.NewInt(500))
				s.bridgeKeeper.SetTripartyWindowLimit(s.ctx, math.NewInt(5000))
				return []interface{}{}
			},
			as:        s.account1.EvmAddr,
			basicPass: true,
			output:    []interface{}{big.NewInt(500), big.NewInt(5000)},
		},
		{
			name: "wrong number of inputs",
			run: func() []interface{} {
				return []interface{}{big.NewInt(1)}
			},
			as:        s.account1.EvmAddr,
			basicPass: false,
		},
	}

	s.RunMethodTestCases(testcases, "getTripartyLimits")
}
