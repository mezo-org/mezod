package maintenance_test

import (
	"github.com/ethereum/go-ethereum/common"
)

// zeroAddress is the EVM zero address. The transaction lockdown allowlists
// reject it.
var zeroAddress = common.Address{}

// setTxLockdownParams writes the transaction lockdown state into the EVM
// parameters. The test cases use it to seed the state a method operates on.
func (s *PrecompileTestSuite) setTxLockdownParams(
	enabled bool,
	senders []string,
	targets []string,
) {
	params := s.evmKeeper.GetParams(s.ctx)
	params.TxLockdownEnabled = enabled
	params.TxLockdownSenders = senders
	params.TxLockdownTargets = targets
	s.Require().NoError(s.evmKeeper.SetParams(s.ctx, params))
}

func (s *PrecompileTestSuite) TestSetTxLockdown() {
	testcases := []TestCase{
		{
			name: "argument count mismatch",
			run: func() []interface{} {
				return []interface{}{
					true,
					true,
				}
			},
			errContains: "argument count mismatch",
		},
		{
			name: "invalid enabled argument type",
			run: func() []interface{} {
				return []interface{}{
					"true",
				}
			},
			errContains: "cannot use string as type bool as argument",
		},
		{
			name: "sender is neither the owner nor the emergency team",
			run: func() []interface{} {
				return []interface{}{
					true,
				}
			},
			as:          s.account2.EvmAddr,
			basicPass:   true,
			revert:      true,
			errContains: "emergency team address is empty",
			postCheck: func() {
				s.Require().False(s.evmKeeper.GetParams(s.ctx).TxLockdownEnabled)
			},
		},
		{
			name: "valid call - the owner enables the lockdown",
			run: func() []interface{} {
				return []interface{}{
					true,
				}
			},
			as:        s.account1.EvmAddr,
			basicPass: true,
			output:    []interface{}{true},
			postCheck: func() {
				s.Require().True(s.evmKeeper.GetParams(s.ctx).TxLockdownEnabled)

				topics, arguments := s.requireEmittedEvent("TxLockdownSet")
				s.Require().Empty(topics)
				s.Require().Equal([]interface{}{true}, arguments)
			},
		},
		{
			name: "sender is not the emergency team",
			run: func() []interface{} {
				err := s.poaKeeper.SetEmergencyTeam(
					s.ctx,
					s.account1.SdkAddr,
					s.account3.SdkAddr,
				)
				s.Require().NoError(err)

				return []interface{}{
					false,
				}
			},
			as:          s.account2.EvmAddr,
			basicPass:   true,
			revert:      true,
			errContains: "sender is not the emergency team",
			postCheck: func() {
				// The lockdown keeps the value set by the previous test case.
				s.Require().True(s.evmKeeper.GetParams(s.ctx).TxLockdownEnabled)
			},
		},
		{
			name: "valid call - the emergency team disables the lockdown",
			run: func() []interface{} {
				s.setTxLockdownParams(
					true,
					[]string{s.account2.EvmAddr.Hex()},
					[]string{s.account3.EvmAddr.Hex()},
				)

				return []interface{}{
					false,
				}
			},
			as:        s.account3.EvmAddr,
			basicPass: true,
			output:    []interface{}{true},
			postCheck: func() {
				params := s.evmKeeper.GetParams(s.ctx)
				s.Require().False(params.TxLockdownEnabled)
				// Disabling the lockdown clears both extra allowlists.
				s.Require().Empty(params.TxLockdownSenders)
				s.Require().Empty(params.TxLockdownTargets)

				_, arguments := s.requireEmittedEvent("TxLockdownSet")
				s.Require().Equal([]interface{}{false}, arguments)
			},
		},
		{
			name: "valid call - the owner enables the lockdown again",
			run: func() []interface{} {
				s.setTxLockdownParams(
					false,
					[]string{s.account2.EvmAddr.Hex()},
					nil,
				)

				return []interface{}{
					true,
				}
			},
			as:        s.account1.EvmAddr,
			basicPass: true,
			output:    []interface{}{true},
			postCheck: func() {
				params := s.evmKeeper.GetParams(s.ctx)
				s.Require().True(params.TxLockdownEnabled)
				// Enabling the lockdown keeps the extra allowlists intact.
				s.Require().Equal(
					[]string{s.account2.EvmAddr.Hex()},
					params.TxLockdownSenders,
				)
			},
		},
	}

	s.RunMethodTestCases(testcases, "setTxLockdown")
}

func (s *PrecompileTestSuite) TestGetTxLockdown() {
	testcases := []TestCase{
		{
			name: "argument count mismatch",
			run: func() []interface{} {
				return []interface{}{
					true, // Additional argument
				}
			},
			errContains: "argument count mismatch",
		},
		{
			name:      "valid call - no lockdown",
			run:       func() []interface{} { return nil },
			as:        s.account2.EvmAddr,
			basicPass: true,
			output:    []interface{}{false},
		},
		{
			name: "valid call - lockdown enabled",
			run: func() []interface{} {
				s.setTxLockdownParams(true, nil, nil)
				return nil
			},
			as:        s.account2.EvmAddr,
			basicPass: true,
			output:    []interface{}{true},
			postCheck: func() {
				// A read method must not emit events.
				s.requireNoEmittedEvent()
			},
		},
	}

	s.RunMethodTestCases(testcases, "getTxLockdown")
}

func (s *PrecompileTestSuite) TestSetTxLockdownAllowlist() {
	testcases := []TestCase{
		{
			name: "argument count mismatch",
			run: func() []interface{} {
				return []interface{}{
					[]common.Address{s.account2.EvmAddr},
				}
			},
			errContains: "argument count mismatch",
		},
		{
			name: "invalid senders argument type",
			run: func() []interface{} {
				return []interface{}{
					s.account2.EvmAddr.Hex(),
					[]common.Address{s.account3.EvmAddr},
				}
			},
			errContains: "cannot use string as type",
		},
		{
			name: "invalid targets argument type",
			run: func() []interface{} {
				return []interface{}{
					[]common.Address{s.account2.EvmAddr},
					s.account3.EvmAddr.Hex(),
				}
			},
			errContains: "cannot use string as type",
		},
		{
			name: "sender is neither the owner nor the emergency team",
			run: func() []interface{} {
				return []interface{}{
					[]common.Address{s.account2.EvmAddr},
					[]common.Address{s.account3.EvmAddr},
				}
			},
			as:          s.account2.EvmAddr,
			basicPass:   true,
			revert:      true,
			errContains: "emergency team address is empty",
			postCheck: func() {
				params := s.evmKeeper.GetParams(s.ctx)
				s.Require().Empty(params.TxLockdownSenders)
				s.Require().Empty(params.TxLockdownTargets)
			},
		},
		{
			name: "valid call - the owner sets both lists",
			run: func() []interface{} {
				return []interface{}{
					[]common.Address{s.account2.EvmAddr},
					[]common.Address{s.account3.EvmAddr},
				}
			},
			as:        s.account1.EvmAddr,
			basicPass: true,
			output:    []interface{}{true},
			postCheck: func() {
				params := s.evmKeeper.GetParams(s.ctx)
				s.Require().Equal(
					[]string{s.account2.EvmAddr.Hex()},
					params.TxLockdownSenders,
				)
				s.Require().Equal(
					[]string{s.account3.EvmAddr.Hex()},
					params.TxLockdownTargets,
				)

				topics, arguments := s.requireEmittedEvent("TxLockdownAllowlistSet")
				s.Require().Empty(topics)
				s.Require().Equal(
					[]interface{}{
						[]common.Address{s.account2.EvmAddr},
						[]common.Address{s.account3.EvmAddr},
					},
					arguments,
				)
			},
		},
		{
			name: "valid call - the emergency team replaces both lists",
			run: func() []interface{} {
				err := s.poaKeeper.SetEmergencyTeam(
					s.ctx,
					s.account1.SdkAddr,
					s.account3.SdkAddr,
				)
				s.Require().NoError(err)

				return []interface{}{
					[]common.Address{s.account1.EvmAddr},
					[]common.Address{s.account2.EvmAddr},
				}
			},
			as:        s.account3.EvmAddr,
			basicPass: true,
			output:    []interface{}{true},
			postCheck: func() {
				params := s.evmKeeper.GetParams(s.ctx)
				// The call replaces the lists, it does not append to them.
				s.Require().Equal(
					[]string{s.account1.EvmAddr.Hex()},
					params.TxLockdownSenders,
				)
				s.Require().Equal(
					[]string{s.account2.EvmAddr.Hex()},
					params.TxLockdownTargets,
				)
			},
		},
		{
			name: "the zero address in the senders reverts",
			run: func() []interface{} {
				return []interface{}{
					[]common.Address{s.account2.EvmAddr, zeroAddress},
					[]common.Address{s.account2.EvmAddr},
				}
			},
			as:          s.account1.EvmAddr,
			basicPass:   true,
			revert:      true,
			errContains: "zero EVM address",
			postCheck: func() {
				params := s.evmKeeper.GetParams(s.ctx)
				// The lists keep the values set by the previous test case.
				s.Require().Equal(
					[]string{s.account1.EvmAddr.Hex()},
					params.TxLockdownSenders,
				)
			},
		},
		{
			name: "the zero address in the targets reverts",
			run: func() []interface{} {
				return []interface{}{
					[]common.Address{s.account2.EvmAddr},
					[]common.Address{zeroAddress},
				}
			},
			as:          s.account1.EvmAddr,
			basicPass:   true,
			revert:      true,
			errContains: "zero EVM address",
		},
		{
			name: "a duplicate sender reverts",
			run: func() []interface{} {
				return []interface{}{
					[]common.Address{s.account2.EvmAddr, s.account2.EvmAddr},
					[]common.Address{s.account3.EvmAddr},
				}
			},
			as:          s.account1.EvmAddr,
			basicPass:   true,
			revert:      true,
			errContains: "duplicate",
		},
		{
			name: "valid call - empty lists clear the allowlist",
			run: func() []interface{} {
				return []interface{}{
					[]common.Address{},
					[]common.Address{},
				}
			},
			as:        s.account1.EvmAddr,
			basicPass: true,
			output:    []interface{}{true},
			postCheck: func() {
				params := s.evmKeeper.GetParams(s.ctx)
				s.Require().Empty(params.TxLockdownSenders)
				s.Require().Empty(params.TxLockdownTargets)

				_, arguments := s.requireEmittedEvent("TxLockdownAllowlistSet")
				s.Require().Equal(
					[]interface{}{
						[]common.Address{},
						[]common.Address{},
					},
					arguments,
				)
			},
		},
		{
			name: "valid call - the lockdown flag does not gate the allowlist",
			run: func() []interface{} {
				s.setTxLockdownParams(false, nil, nil)

				return []interface{}{
					[]common.Address{s.account2.EvmAddr},
					[]common.Address{},
				}
			},
			as:        s.account1.EvmAddr,
			basicPass: true,
			output:    []interface{}{true},
			postCheck: func() {
				params := s.evmKeeper.GetParams(s.ctx)
				s.Require().False(params.TxLockdownEnabled)
				s.Require().Equal(
					[]string{s.account2.EvmAddr.Hex()},
					params.TxLockdownSenders,
				)
				s.Require().Empty(params.TxLockdownTargets)
			},
		},
	}

	s.RunMethodTestCases(testcases, "setTxLockdownAllowlist")
}

func (s *PrecompileTestSuite) TestGetTxLockdownAllowlist() {
	testcases := []TestCase{
		{
			name: "argument count mismatch",
			run: func() []interface{} {
				return []interface{}{
					[]common.Address{}, // Additional argument
				}
			},
			errContains: "argument count mismatch",
		},
		{
			name:      "valid call - empty allowlist",
			run:       func() []interface{} { return nil },
			as:        s.account2.EvmAddr,
			basicPass: true,
			output: []interface{}{
				[]common.Address{},
				[]common.Address{},
			},
		},
		{
			name: "valid call - senders only",
			run: func() []interface{} {
				s.setTxLockdownParams(
					true,
					[]string{s.account2.EvmAddr.Hex()},
					nil,
				)
				return nil
			},
			as:        s.account2.EvmAddr,
			basicPass: true,
			output: []interface{}{
				[]common.Address{s.account2.EvmAddr},
				[]common.Address{},
			},
		},
		{
			name: "valid call - both lists",
			run: func() []interface{} {
				s.setTxLockdownParams(
					true,
					[]string{s.account1.EvmAddr.Hex(), s.account2.EvmAddr.Hex()},
					[]string{s.account3.EvmAddr.Hex()},
				)
				return nil
			},
			as:        s.account2.EvmAddr,
			basicPass: true,
			output: []interface{}{
				[]common.Address{s.account1.EvmAddr, s.account2.EvmAddr},
				[]common.Address{s.account3.EvmAddr},
			},
			postCheck: func() {
				// A read method must not emit events.
				s.requireNoEmittedEvent()
			},
		},
	}

	s.RunMethodTestCases(testcases, "getTxLockdownAllowlist")
}
