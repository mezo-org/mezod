package maintenance_test

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/mezo-org/mezod/precompile"
)

func (s *PrecompileTestSuite) TestSetEmergencyTeam() {
	testcases := []TestCase{
		{
			name: "argument count mismatch",
			run: func() []interface{} {
				return []interface{}{}
			},
			errContains: "argument count mismatch",
		},
		{
			name: "invalid team argument type",
			run: func() []interface{} {
				return []interface{}{
					"0x0000000000000000000000000000000000000001",
				}
			},
			errContains: "cannot use string as type array as argument",
		},
		{
			name: "sender is not owner",
			run: func() []interface{} {
				return []interface{}{
					s.account3.EvmAddr,
				}
			},
			as:          s.account2.EvmAddr,
			basicPass:   true,
			revert:      true,
			errContains: "sender is not owner",
			postCheck: func() {
				s.Require().True(s.poaKeeper.GetEmergencyTeam(s.ctx).Empty())
			},
		},
		{
			name: "valid call - grant the role",
			run: func() []interface{} {
				return []interface{}{
					s.account3.EvmAddr,
				}
			},
			as:        s.account1.EvmAddr,
			basicPass: true,
			output:    []interface{}{true},
			postCheck: func() {
				emergencyTeam := s.poaKeeper.GetEmergencyTeam(s.ctx)
				s.Require().Equal(
					s.account3.EvmAddr,
					precompile.TypesConverter.Address.FromSDK(emergencyTeam),
				)

				topics, arguments := s.requireEmittedEvent("EmergencyTeamSet")
				s.Require().Empty(arguments)
				s.Require().Len(topics, 2)
				s.Require().Equal(common.Hash{}, topics[0])
				s.Require().Equal(
					common.BytesToHash(s.account3.EvmAddr.Bytes()),
					topics[1],
				)
			},
		},
		{
			name: "valid call - rotate the role",
			run: func() []interface{} {
				return []interface{}{
					s.account2.EvmAddr,
				}
			},
			as:        s.account1.EvmAddr,
			basicPass: true,
			output:    []interface{}{true},
			postCheck: func() {
				emergencyTeam := s.poaKeeper.GetEmergencyTeam(s.ctx)
				s.Require().Equal(
					s.account2.EvmAddr,
					precompile.TypesConverter.Address.FromSDK(emergencyTeam),
				)

				topics, _ := s.requireEmittedEvent("EmergencyTeamSet")
				s.Require().Equal(
					common.BytesToHash(s.account3.EvmAddr.Bytes()),
					topics[0],
				)
				s.Require().Equal(
					common.BytesToHash(s.account2.EvmAddr.Bytes()),
					topics[1],
				)
			},
		},
		{
			name: "valid call - revoke the role",
			run: func() []interface{} {
				return []interface{}{
					common.Address{},
				}
			},
			as:        s.account1.EvmAddr,
			basicPass: true,
			output:    []interface{}{true},
			postCheck: func() {
				s.Require().True(s.poaKeeper.GetEmergencyTeam(s.ctx).Empty())

				topics, _ := s.requireEmittedEvent("EmergencyTeamSet")
				s.Require().Equal(
					common.BytesToHash(s.account2.EvmAddr.Bytes()),
					topics[0],
				)
				s.Require().Equal(common.Hash{}, topics[1])
			},
		},
	}

	s.RunMethodTestCases(testcases, "setEmergencyTeam")
}

func (s *PrecompileTestSuite) TestGetEmergencyTeam() {
	testcases := []TestCase{
		{
			name: "argument count mismatch",
			run: func() []interface{} {
				return []interface{}{
					1, // Additional argument
				}
			},
			errContains: "argument count mismatch",
		},
		{
			name:      "valid call - role not granted",
			run:       func() []interface{} { return nil },
			as:        s.account2.EvmAddr,
			basicPass: true,
			output:    []interface{}{common.Address{}},
		},
		{
			name: "valid call - role granted",
			run: func() []interface{} {
				err := s.poaKeeper.SetEmergencyTeam(
					s.ctx,
					s.account1.SdkAddr,
					s.account3.SdkAddr,
				)
				s.Require().NoError(err)
				return nil
			},
			as:        s.account2.EvmAddr,
			basicPass: true,
			output:    []interface{}{s.account3.EvmAddr},
			postCheck: func() {
				// A read method must not emit events.
				s.requireNoEmittedEvent()
			},
		},
	}

	s.RunMethodTestCases(testcases, "getEmergencyTeam")
}
