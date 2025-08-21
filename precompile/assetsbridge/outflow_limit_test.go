package assetsbridge_test

import (
	"math/big"
	"testing"

	"cosmossdk.io/math"
	"github.com/ethereum/go-ethereum/common"
	"github.com/mezo-org/mezod/precompile/assetsbridge"
	"github.com/stretchr/testify/suite"
)

var testTokenAddress = common.HexToAddress("0x1234567890123456789012345678901234567890")

type OutflowLimitTestSuite struct {
	PrecompileTestSuite
}

func TestOutflowLimitTestSuite(t *testing.T) {
	suite.Run(t, new(OutflowLimitTestSuite))
}

func (s *OutflowLimitTestSuite) TestSetOutflowLimitMethod() {
	testCases := []TestCase{
		{
			name: "success - owner sets valid limit",
			run: func() []interface{} {
				return []interface{}{
					testTokenAddress,
					big.NewInt(1000000),
				}
			},
			as:        s.account1.EvmAddr,
			basicPass: true,
			output:    []interface{}{true},
			postCheck: func() {
				limit := s.bridgeKeeper.GetOutflowLimit(s.ctx, testTokenAddress.Bytes())
				s.Require().Equal(math.NewInt(1000000), limit)
			},
		},
		{
			name: "success - owner sets zero limit",
			run: func() []interface{} {
				return []interface{}{
					testTokenAddress,
					big.NewInt(0),
				}
			},
			as:        s.account1.EvmAddr,
			basicPass: true,
			output:    []interface{}{true},
			postCheck: func() {
				limit := s.bridgeKeeper.GetOutflowLimit(s.ctx, testTokenAddress.Bytes())
				s.Require().Equal(math.NewInt(0), limit)
			},
		},
		{
			name: "failure - non-owner attempts to set limit",
			run: func() []interface{} {
				return []interface{}{
					testTokenAddress,
					big.NewInt(1000000),
				}
			},
			as:          s.account2.EvmAddr,
			basicPass:   true,
			revert:      true,
			errContains: "sender is not owner",
		},
		{
			name: "failure - invalid token address",
			run: func() []interface{} {
				return []interface{}{
					"invalid address",
					big.NewInt(1000000),
				}
			},
			as:        s.account1.EvmAddr,
			basicPass: false,
		},
		{
			name: "failure - invalid limit type",
			run: func() []interface{} {
				return []interface{}{
					testTokenAddress,
					"invalid limit",
				}
			},
			as:        s.account1.EvmAddr,
			basicPass: false,
		},
		{
			name: "failure - wrong number of inputs",
			run: func() []interface{} {
				return []interface{}{
					testTokenAddress,
				}
			},
			as:        s.account1.EvmAddr,
			basicPass: false,
		},
	}

	s.RunMethodTestCases(testCases, assetsbridge.SetOutflowLimitMethodName)
}

func (s *OutflowLimitTestSuite) TestGetOutflowLimitMethod() {
	testCases := []TestCase{
		{
			name: "success - returns set limit",
			run: func() []interface{} {
				// Set a limit first
				s.bridgeKeeper.SetOutflowLimit(s.ctx, testTokenAddress.Bytes(), math.NewInt(5000000))
				return []interface{}{
					testTokenAddress,
				}
			},
			as:        s.account1.EvmAddr,
			basicPass: true,
			output:    []interface{}{big.NewInt(5000000)},
		},
		{
			name: "success - returns zero for unset token",
			run: func() []interface{} {
				return []interface{}{
					common.HexToAddress("0x9999999999999999999999999999999999999999"),
				}
			},
			as:        s.account1.EvmAddr,
			basicPass: true,
			output:    []interface{}{big.NewInt(0)},
		},
		{
			name: "failure - invalid token address",
			run: func() []interface{} {
				return []interface{}{
					"invalid address",
				}
			},
			as:        s.account1.EvmAddr,
			basicPass: false,
		},
		{
			name: "failure - wrong number of inputs",
			run: func() []interface{} {
				return []interface{}{
					testTokenAddress,
					big.NewInt(123),
				}
			},
			as:        s.account1.EvmAddr,
			basicPass: false,
		},
	}

	s.RunMethodTestCases(testCases, assetsbridge.GetOutflowLimitMethodName)
}

func (s *OutflowLimitTestSuite) TestGetOutflowCapacityMethod() {
	testCases := []TestCase{
		{
			name: "success - when no outflow",
			run: func() []interface{} {
				// Set a limit for the token
				s.bridgeKeeper.SetOutflowLimit(s.ctx, testTokenAddress.Bytes(), math.NewInt(1000000))
				return []interface{}{
					testTokenAddress,
				}
			},
			as:        s.account1.EvmAddr,
			basicPass: true,
			output: []interface{}{
				big.NewInt(1000000), // capacity equals limit when no outflow
				big.NewInt(25000),   // reset height from fake keeper
			},
		},
		{
			name: "success - with outflow",
			run: func() []interface{} {
				// Set a limit for the token
				s.bridgeKeeper.SetOutflowLimit(s.ctx, testTokenAddress.Bytes(), math.NewInt(1000000))
				// Increase outflow by 150,000
				s.bridgeKeeper.increaseCurrentOutflow(testTokenAddress.Bytes(), math.NewInt(150000))
				return []interface{}{
					testTokenAddress,
				}
			},
			as:        s.account1.EvmAddr,
			basicPass: true,
			output: []interface{}{
				big.NewInt(850000), // capacity equals limit when no outflow
				big.NewInt(25000),  // reset height from fake keeper
			},
		},
		{
			name: "success - zero capacity for zero limit",
			run: func() []interface{} {
				// Don't set any limit (defaults to zero)
				return []interface{}{
					common.HexToAddress("0x8888888888888888888888888888888888888888"),
				}
			},
			as:        s.account1.EvmAddr,
			basicPass: true,
			output: []interface{}{
				big.NewInt(0), // zero capacity
				big.NewInt(25000),
			},
		},
		{
			name: "failure - invalid token address",
			run: func() []interface{} {
				return []interface{}{
					"invalid address",
				}
			},
			as:        s.account1.EvmAddr,
			basicPass: false,
		},
		{
			name: "failure - wrong number of inputs",
			run: func() []interface{} {
				return []interface{}{
					testTokenAddress,
					big.NewInt(123),
				}
			},
			as:        s.account1.EvmAddr,
			basicPass: false,
		},
	}

	s.RunMethodTestCases(testCases, assetsbridge.GetOutflowCapacityMethodName)
}
