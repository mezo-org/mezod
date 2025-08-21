package assetsbridge_test

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/mezo-org/mezod/precompile"
)

var testPauserAddr = common.HexToAddress("0xAbCdEfAbCdEfAbCdEfAbCdEfAbCdEfAbCdEfAbCdEf")

func (s *PrecompileTestSuite) TestSetPauser() {
	testcases := []TestCase{
		{
			name: "caller is not owner",
			run: func() []interface{} {
				return []interface{}{testPauserAddr}
			},
			as:          s.account2.EvmAddr, // account2 is not the owner
			basicPass:   true,
			revert:      true,
			errContains: "sender is not owner",
		},
		{
			name: "happy path - set valid pauser",
			run: func() []interface{} {
				return []interface{}{testPauserAddr}
			},
			as:        s.account1.EvmAddr, // account1 is the owner
			basicPass: true,
			output:    []interface{}{true},
			postCheck: func() {
				pauser := s.bridgeKeeper.GetPauser(s.ctx)
				s.Require().True(precompile.TypesConverter.Address.FromSDK(pauser) == testPauserAddr)
			},
		},
		{
			name: "happy path - set zero address (remove pauser)",
			run: func() []interface{} {
				return []interface{}{common.Address{}}
			},
			as:        s.account1.EvmAddr, // account1 is the owner
			basicPass: true,
			output:    []interface{}{true},
			postCheck: func() {
				pauser := s.bridgeKeeper.GetPauser(s.ctx)
				s.Require().True(precompile.TypesConverter.Address.FromSDK(pauser) == common.Address{})
			},
		},
	}

	s.RunMethodTestCases(testcases, "setPauser")
}

func (s *PrecompileTestSuite) TestGetPauser() {
	testcases := []TestCase{
		{
			name: "no pauser is set - returns zero address",
			run: func() []interface{} {
				// Ensure no pauser is set
				s.bridgeKeeper.SetPauser(s.ctx, nil)
				return []interface{}{}
			},
			as:        s.account1.EvmAddr, // anyone can call this method
			basicPass: true,
			output:    []interface{}{common.Address{}}, // zero address
		},
		{
			name: "pauser is set - returns correct address",
			run: func() []interface{} {
				// Set a pauser
				s.bridgeKeeper.SetPauser(s.ctx, testPauserAddr.Bytes())
				return []interface{}{}
			},
			as:        s.account2.EvmAddr, // anyone can call this method
			basicPass: true,
			output:    []interface{}{testPauserAddr},
		},
	}

	s.RunMethodTestCases(testcases, "getPauser")
}

func (s *PrecompileTestSuite) TestPauseBridgeOut() {
	testcases := []TestCase{
		{
			name: "no pauser is set",
			run: func() []interface{} {
				return []interface{}{}
			},
			as:          testPauserAddr, // doesn't matter who calls it
			basicPass:   true,
			revert:      true,
			errContains: "no pauser is set",
		},
		{
			name: "caller is not the pauser",
			run: func() []interface{} {
				// First set a pauser
				s.bridgeKeeper.SetPauser(s.ctx, testPauserAddr.Bytes())
				return []interface{}{}
			},
			as:          s.account1.EvmAddr, // not the pauser
			basicPass:   true,
			revert:      true,
			errContains: "caller is not the pauser",
		},
		{
			name: "happy path - pauser calls pause",
			run: func() []interface{} {
				// Set the pauser
				s.bridgeKeeper.SetPauser(s.ctx, testPauserAddr.Bytes())
				return []interface{}{}
			},
			as:        testPauserAddr, // the pauser calls it
			basicPass: true,
			output:    []interface{}{true},
			postCheck: func() {
				s.Require().True(s.bridgeKeeper.isPaused())
			},
		},
	}

	s.RunMethodTestCases(testcases, "pauseBridgeOut")
}
