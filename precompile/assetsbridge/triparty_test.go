package assetsbridge_test

import (
	"math/big"

	"cosmossdk.io/math"
	"github.com/ethereum/go-ethereum/common"
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
			name: "triparty is paused",
			run: func() []interface{} {
				s.bridgeKeeper.SetTripartyPaused(s.ctx, true)
				s.bridgeKeeper.AllowTripartyController(s.ctx, testTripartyController.Bytes(), true)
				return []interface{}{s.account2.EvmAddr, big.NewInt(1000), []byte{}}
			},
			as:          testTripartyController,
			basicPass:   true,
			revert:      true,
			errContains: "triparty bridging is paused",
		},
		{
			name: "caller is not a controller",
			run: func() []interface{} {
				s.bridgeKeeper.SetTripartyPaused(s.ctx, false)
				return []interface{}{s.account2.EvmAddr, big.NewInt(1000), []byte{}}
			},
			as:          s.account2.EvmAddr,
			basicPass:   true,
			revert:      true,
			errContains: "caller is not an allowed triparty controller",
		},
		{
			name: "zero recipient",
			run: func() []interface{} {
				s.bridgeKeeper.SetTripartyPaused(s.ctx, false)
				s.bridgeKeeper.AllowTripartyController(s.ctx, testTripartyController.Bytes(), true)
				return []interface{}{common.Address{}, big.NewInt(1000), []byte{}}
			},
			as:          testTripartyController,
			basicPass:   true,
			revert:      true,
			errContains: "recipient address must not be the zero address",
		},
		{
			name: "zero amount",
			run: func() []interface{} {
				s.bridgeKeeper.SetTripartyPaused(s.ctx, false)
				s.bridgeKeeper.AllowTripartyController(s.ctx, testTripartyController.Bytes(), true)
				return []interface{}{s.account2.EvmAddr, big.NewInt(0), []byte{}}
			},
			as:          testTripartyController,
			basicPass:   true,
			revert:      true,
			errContains: "amount must be positive",
		},
		{
			name: "per-request limit exceeded",
			run: func() []interface{} {
				s.bridgeKeeper.SetTripartyPaused(s.ctx, false)
				s.bridgeKeeper.AllowTripartyController(s.ctx, testTripartyController.Bytes(), true)
				s.bridgeKeeper.SetTripartyPerRequestLimit(s.ctx, math.NewInt(500))
				return []interface{}{s.account2.EvmAddr, big.NewInt(1000), []byte{}}
			},
			as:          testTripartyController,
			basicPass:   true,
			revert:      true,
			errContains: "triparty per-request limit exceeded",
		},
		{
			name: "happy path - returns requestId 1",
			run: func() []interface{} {
				s.bridgeKeeper.SetTripartyPaused(s.ctx, false)
				s.bridgeKeeper.AllowTripartyController(s.ctx, testTripartyController.Bytes(), true)
				s.bridgeKeeper.SetTripartyPerRequestLimit(s.ctx, math.ZeroInt())
				return []interface{}{s.account2.EvmAddr, big.NewInt(1000), []byte("callback")}
			},
			as:        testTripartyController,
			basicPass: true,
			output:    []interface{}{big.NewInt(1)},
		},
		{
			name: "happy path - sequential requestIds",
			run: func() []interface{} {
				s.bridgeKeeper.SetTripartyPaused(s.ctx, false)
				s.bridgeKeeper.AllowTripartyController(s.ctx, testTripartyController.Bytes(), true)
				s.bridgeKeeper.SetTripartyPerRequestLimit(s.ctx, math.ZeroInt())
				return []interface{}{s.account2.EvmAddr, big.NewInt(2000), []byte{}}
			},
			as:        testTripartyController,
			basicPass: true,
			output:    []interface{}{big.NewInt(2)},
		},
		{
			name: "happy path - amount equals per-request limit",
			run: func() []interface{} {
				s.bridgeKeeper.SetTripartyPaused(s.ctx, false)
				s.bridgeKeeper.AllowTripartyController(s.ctx, testTripartyController.Bytes(), true)
				s.bridgeKeeper.SetTripartyPerRequestLimit(s.ctx, math.NewInt(1000))
				return []interface{}{s.account2.EvmAddr, big.NewInt(1000), []byte{}}
			},
			as:        testTripartyController,
			basicPass: true,
			output:    []interface{}{big.NewInt(3)},
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
					uint64(5),
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
				s.Require().True(
					s.bridgeKeeper.GetTripartyPerRequestLimit(s.ctx).
						Equal(
							s.bridgeKeeper.GetTripartyPerRequestLimit(s.ctx),
						),
				)
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
