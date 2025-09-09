package assetsbridge_test

import (
	"math/big"

	"cosmossdk.io/math"

	"github.com/ethereum/go-ethereum/common"
)

func (s *BridgeOutTestSuite) TestSetMinBridgeOutAmount() {
	token := common.HexToAddress("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48")

	testcases := []TestCase{
		{
			name: "caller is not owner",
			run: func() []interface{} {
				return []interface{}{token, big.NewInt(123)}
			},
			as:          s.account2.EvmAddr, // not the owner
			basicPass:   true,
			revert:      true,
			errContains: "sender is not owner",
		},
		{
			name: "zero address token",
			run: func() []interface{} {
				return []interface{}{common.Address{}, big.NewInt(100)}
			},
			as:          s.account1.EvmAddr,
			basicPass:   true,
			revert:      true,
			errContains: "mezo token cannot be the zero address",
		},
		{
			name: "non-positive minimum amount",
			run: func() []interface{} {
				return []interface{}{token, big.NewInt(0)}
			},
			as:          s.account1.EvmAddr,
			basicPass:   true,
			revert:      true,
			errContains: "minimum amount must be positive",
		},
		{
			name: "happy path - set new minimum",
			run: func() []interface{} {
				return []interface{}{token, big.NewInt(250)}
			},
			as:        s.account1.EvmAddr,
			basicPass: true,
			output:    []interface{}{true},
			postCheck: func() {
				minAmount := s.extBridgeKeeper.GetMinBridgeOutAmount(s.ctx, token.Bytes())
				s.Require().EqualValues(math.NewInt(250), minAmount)
			},
		},
		{
			name: "overwrite existing minimum",
			run: func() []interface{} {
				// first set to 250 then update to 999
				_ = s.extBridgeKeeper.SetMinBridgeOutAmount(s.ctx, token.Bytes(), math.NewInt(250))
				return []interface{}{token, big.NewInt(999)}
			},
			as:        s.account1.EvmAddr,
			basicPass: true,
			output:    []interface{}{true},
			postCheck: func() {
				minAmount := s.extBridgeKeeper.GetMinBridgeOutAmount(s.ctx, token.Bytes())
				s.Require().EqualValues(math.NewInt(999), minAmount)
			},
		},
	}

	s.RunMethodTestCasesWithKeepers(testcases, "setMinBridgeOutAmount")
}

func (s *BridgeOutTestSuite) TestGetMinBridgeOutAmount() {
	token := common.HexToAddress("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48")

	testcases := []TestCase{
		{
			name:      "minimum amount not set",
			run:       func() []interface{} { return []interface{}{token} },
			basicPass: true,
			output:    []interface{}{big.NewInt(0)},
		},
		{
			name: "minimum amount set",
			run: func() []interface{} {
				_ = s.extBridgeKeeper.SetMinBridgeOutAmount(s.ctx, token.Bytes(), math.NewInt(123456))
				return []interface{}{token}
			},
			basicPass: true,
			output:    []interface{}{big.NewInt(123456)},
		},
	}

	s.RunMethodTestCasesWithKeepers(testcases, "getMinBridgeOutAmount")
}

func (s *BridgeOutTestSuite) TestSetMinBridgeOutAmountForBitcoinChain() {
	testcases := []TestCase{
		{
			name: "caller is not owner",
			run: func() []interface{} {
				return []interface{}{big.NewInt(123)}
			},
			as:          s.account2.EvmAddr, // not the owner
			basicPass:   true,
			revert:      true,
			errContains: "sender is not owner",
		},
		{
			name: "zero minimum amount",
			run: func() []interface{} {
				return []interface{}{big.NewInt(0)}
			},
			as:        s.account1.EvmAddr,
			basicPass: true,
			output:    []interface{}{true},
			postCheck: func() {
				minAmount := s.extBridgeKeeper.GetMinBridgeOutAmountForBitcoinChain(s.ctx)
				s.Require().EqualValues(math.ZeroInt(), minAmount)
			},
		},
		{
			name: "happy path - set new minimum",
			run: func() []interface{} {
				return []interface{}{big.NewInt(250000)}
			},
			as:        s.account1.EvmAddr,
			basicPass: true,
			output:    []interface{}{true},
			postCheck: func() {
				minAmount := s.extBridgeKeeper.GetMinBridgeOutAmountForBitcoinChain(s.ctx)
				s.Require().EqualValues(math.NewInt(250000), minAmount)
			},
		},
		{
			name: "overwrite existing minimum",
			run: func() []interface{} {
				// first set to 250000 then update to 500000
				s.extBridgeKeeper.SetMinBridgeOutAmountForBitcoinChain(s.ctx, math.NewInt(250000))
				return []interface{}{big.NewInt(500000)}
			},
			as:        s.account1.EvmAddr,
			basicPass: true,
			output:    []interface{}{true},
			postCheck: func() {
				minAmount := s.extBridgeKeeper.GetMinBridgeOutAmountForBitcoinChain(s.ctx)
				s.Require().EqualValues(math.NewInt(500000), minAmount)
			},
		},
	}

	s.RunMethodTestCasesWithKeepers(testcases, "setMinBridgeOutAmountForBitcoinChain")
}

func (s *BridgeOutTestSuite) TestGetMinBridgeOutAmountForBitcoinChain() {
	testcases := []TestCase{
		{
			name:      "minimum amount not set",
			run:       func() []interface{} { return []interface{}{} },
			basicPass: true,
			output:    []interface{}{big.NewInt(0)},
		},
		{
			name: "minimum amount set",
			run: func() []interface{} {
				s.extBridgeKeeper.SetMinBridgeOutAmountForBitcoinChain(s.ctx, math.NewInt(75000))
				return []interface{}{}
			},
			basicPass: true,
			output:    []interface{}{big.NewInt(75000)},
		},
	}

	s.RunMethodTestCasesWithKeepers(testcases, "getMinBridgeOutAmountForBitcoinChain")
}
