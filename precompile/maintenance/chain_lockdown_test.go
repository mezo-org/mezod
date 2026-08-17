package maintenance_test

import (
	"fmt"
)

func (s *PrecompileTestSuite) TestSetChainLockdown() {
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
			name: "sender is neither the owner nor the emergency team",
			run: func() []interface{} {
				s.upgradeKeeper.reset()
				return nil
			},
			as:          s.account2.EvmAddr,
			basicPass:   true,
			revert:      true,
			errContains: "emergency team address is empty",
			postCheck: func() {
				s.Require().Nil(s.upgradeKeeper.scheduledPlan)
				s.requireNoEmittedEvent()
			},
		},
		{
			name: "sender is not the emergency team",
			run: func() []interface{} {
				s.upgradeKeeper.reset()

				err := s.poaKeeper.SetEmergencyTeam(
					s.ctx,
					s.account1.SdkAddr,
					s.account3.SdkAddr,
				)
				s.Require().NoError(err)

				return nil
			},
			as:          s.account2.EvmAddr,
			basicPass:   true,
			revert:      true,
			errContains: "sender is not the emergency team",
			postCheck: func() {
				s.Require().Nil(s.upgradeKeeper.scheduledPlan)
				s.requireNoEmittedEvent()
			},
		},
		{
			name: "valid call - the owner halts the chain",
			run: func() []interface{} {
				s.upgradeKeeper.reset()
				s.upgradeKeeper.lastCompletedName = "v12.0.0"
				s.upgradeKeeper.lastCompletedHeight = 500

				return nil
			},
			as:        s.account1.EvmAddr,
			basicPass: true,
			output:    []interface{}{true},
			postCheck: func() {
				plan := s.upgradeKeeper.scheduledPlan
				s.Require().NotNil(plan)
				s.Require().Equal("v13.0.0", plan.Name)
				s.Require().Equal(s.ctx.BlockHeight()+1, plan.Height)
				s.Require().Empty(plan.Info)

				topics, arguments := s.requireEmittedEvent("ChainLockdownSet")
				s.Require().Empty(topics)
				s.Require().Equal([]interface{}{"v13.0.0"}, arguments)
			},
		},
		{
			name: "valid call - the emergency team halts a chain without a completed upgrade",
			run: func() []interface{} {
				s.upgradeKeeper.reset()

				err := s.poaKeeper.SetEmergencyTeam(
					s.ctx,
					s.account1.SdkAddr,
					s.account3.SdkAddr,
				)
				s.Require().NoError(err)

				return nil
			},
			as:        s.account3.EvmAddr,
			basicPass: true,
			output:    []interface{}{true},
			postCheck: func() {
				plan := s.upgradeKeeper.scheduledPlan
				s.Require().NotNil(plan)
				s.Require().Equal("v1.0.0", plan.Name)
				s.Require().Equal(s.ctx.BlockHeight()+1, plan.Height)

				_, arguments := s.requireEmittedEvent("ChainLockdownSet")
				s.Require().Equal([]interface{}{"v1.0.0"}, arguments)
			},
		},
		{
			name: "valid call - the binary has a handler for the first candidate",
			run: func() []interface{} {
				s.upgradeKeeper.reset()
				s.upgradeKeeper.lastCompletedName = "v0.4.0"
				s.upgradeKeeper.handlers = map[string]bool{"v1.0.0": true}

				return nil
			},
			as:        s.account1.EvmAddr,
			basicPass: true,
			output:    []interface{}{true},
			postCheck: func() {
				plan := s.upgradeKeeper.scheduledPlan
				s.Require().NotNil(plan)
				s.Require().Equal("v2.0.0", plan.Name)
				s.Require().Equal(s.ctx.BlockHeight()+1, plan.Height)

				_, arguments := s.requireEmittedEvent("ChainLockdownSet")
				s.Require().Equal([]interface{}{"v2.0.0"}, arguments)
			},
		},
		{
			name: "the last completed upgrade name does not parse",
			run: func() []interface{} {
				s.upgradeKeeper.reset()
				s.upgradeKeeper.lastCompletedName = "garbage"

				return nil
			},
			as:          s.account1.EvmAddr,
			basicPass:   true,
			revert:      true,
			errContains: `cannot derive the halt name from the last completed upgrade "garbage"`,
			postCheck: func() {
				s.Require().Nil(s.upgradeKeeper.scheduledPlan)
				s.requireNoEmittedEvent()
			},
		},
		{
			name: "reading the last completed upgrade fails",
			run: func() []interface{} {
				s.upgradeKeeper.reset()
				s.upgradeKeeper.lastCompletedErr = fmt.Errorf("store is broken")

				return nil
			},
			as:          s.account1.EvmAddr,
			basicPass:   true,
			revert:      true,
			errContains: "store is broken",
			postCheck: func() {
				s.Require().Nil(s.upgradeKeeper.scheduledPlan)
				s.requireNoEmittedEvent()
			},
		},
		{
			name: "scheduling the upgrade plan fails",
			run: func() []interface{} {
				s.upgradeKeeper.reset()
				s.upgradeKeeper.lastCompletedName = "v12.0.0"
				s.upgradeKeeper.scheduleUpgradeErr = fmt.Errorf(
					"upgrade cannot be scheduled in the past",
				)

				return nil
			},
			as:          s.account1.EvmAddr,
			basicPass:   true,
			revert:      true,
			errContains: "upgrade cannot be scheduled in the past",
			postCheck: func() {
				s.Require().Nil(s.upgradeKeeper.scheduledPlan)
				s.requireNoEmittedEvent()
			},
		},
	}

	s.RunMethodTestCases(testcases, "setChainLockdown")
}
