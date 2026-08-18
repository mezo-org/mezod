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
					"v13.0.0",
					true, // Additional argument
				}
			},
			errContains: "argument count mismatch",
		},
		{
			name: "invalid plan name argument type",
			run: func() []interface{} {
				return []interface{}{
					true,
				}
			},
			errContains: "cannot use bool as type string as argument",
		},
		{
			name: "sender is neither the owner nor the emergency team",
			run: func() []interface{} {
				s.upgradeKeeper.reset()
				return []interface{}{"v13.0.0"}
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

				return []interface{}{"v13.0.0"}
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
			name: "the plan name is empty",
			run: func() []interface{} {
				s.upgradeKeeper.reset()
				return []interface{}{""}
			},
			as:          s.account1.EvmAddr,
			basicPass:   true,
			revert:      true,
			errContains: "plan name cannot be empty",
			postCheck: func() {
				s.Require().Nil(s.upgradeKeeper.scheduledPlan)
				s.requireNoEmittedEvent()
			},
		},
		{
			name: "the plan name has a registered handler",
			run: func() []interface{} {
				s.upgradeKeeper.reset()
				s.upgradeKeeper.handlers = map[string]bool{"v12.0.0": true}

				return []interface{}{"v12.0.0"}
			},
			as:          s.account1.EvmAddr,
			basicPass:   true,
			revert:      true,
			errContains: `the binary has an upgrade handler for plan name "v12.0.0"`,
			postCheck: func() {
				s.Require().Nil(s.upgradeKeeper.scheduledPlan)
				s.requireNoEmittedEvent()
			},
		},
		{
			name: "the plan name was already completed",
			run: func() []interface{} {
				s.upgradeKeeper.reset()
				s.upgradeKeeper.doneHeights = map[string]int64{"v12.0.0": 500}

				return []interface{}{"v12.0.0"}
			},
			as:          s.account1.EvmAddr,
			basicPass:   true,
			revert:      true,
			errContains: `an upgrade with plan name "v12.0.0" was already completed`,
			postCheck: func() {
				s.Require().Nil(s.upgradeKeeper.scheduledPlan)
				s.requireNoEmittedEvent()
			},
		},
		{
			name: "reading the completion height fails",
			run: func() []interface{} {
				s.upgradeKeeper.reset()
				s.upgradeKeeper.doneHeightErr = fmt.Errorf("store is broken")

				return []interface{}{"v13.0.0"}
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
				s.upgradeKeeper.scheduleUpgradeErr = fmt.Errorf(
					"upgrade cannot be scheduled in the past",
				)

				return []interface{}{"v13.0.0"}
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
		{
			name: "valid call - the owner halts the chain",
			run: func() []interface{} {
				s.upgradeKeeper.reset()
				return []interface{}{"v13.0.0"}
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
			name: "valid call - the emergency team halts the chain",
			run: func() []interface{} {
				s.upgradeKeeper.reset()

				err := s.poaKeeper.SetEmergencyTeam(
					s.ctx,
					s.account1.SdkAddr,
					s.account3.SdkAddr,
				)
				s.Require().NoError(err)

				return []interface{}{"emergency-halt"}
			},
			as:        s.account3.EvmAddr,
			basicPass: true,
			output:    []interface{}{true},
			postCheck: func() {
				plan := s.upgradeKeeper.scheduledPlan
				s.Require().NotNil(plan)
				s.Require().Equal("emergency-halt", plan.Name)
				s.Require().Equal(s.ctx.BlockHeight()+1, plan.Height)

				_, arguments := s.requireEmittedEvent("ChainLockdownSet")
				s.Require().Equal([]interface{}{"emergency-halt"}, arguments)
			},
		},
	}

	s.RunMethodTestCases(testcases, "setChainLockdown")
}
