package assetsbridge_test

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/mezo-org/mezod/precompile"
	"github.com/mezo-org/mezod/precompile/assetsbridge"
	"github.com/mezo-org/mezod/x/evm/statedb"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
	"github.com/stretchr/testify/suite"
)

const (
	ethereumChain = uint8(0)
	bitcoinChain  = uint8(1)
	unknownChain  = uint8(5)
)

type BridgeOutChainsTestSuite struct {
	PrecompileTestSuite

	authzKeeper *FakeAuthzKeeper
	stateDB     *statedb.StateDB
	contract    *precompile.Contract

	// emergencyTeamMember is a plain account. The precompile authorizes the
	// owner only, so the emergency team has no more rights than a third party.
	emergencyTeamMember Key
}

func TestBridgeOutChainsTestSuite(t *testing.T) {
	suite.Run(t, new(BridgeOutChainsTestSuite))
}

func (s *BridgeOutChainsTestSuite) SetupTest() {
	s.PrecompileTestSuite.SetupTest()

	s.authzKeeper = NewFakeAuthzKeeper()
	s.emergencyTeamMember = NewKey()
}

func (s *BridgeOutChainsTestSuite) TestRemoveBridgeOutChain() {
	testcases := []TestCase{
		{
			name: "caller is not owner",
			run: func() []interface{} {
				s.setBridgeOutChains(ethereumChain, bitcoinChain)
				return []interface{}{bitcoinChain}
			},
			as:          s.account2.EvmAddr,
			basicPass:   true,
			revert:      true,
			errContains: "sender is not owner",
			postCheck: func() {
				s.Require().Equal(
					[]uint8{ethereumChain, bitcoinChain},
					s.bridgeKeeper.GetBridgeOutChains(s.ctx),
				)
			},
		},
		{
			name: "caller is emergency team member",
			run: func() []interface{} {
				s.setBridgeOutChains(ethereumChain, bitcoinChain)
				return []interface{}{bitcoinChain}
			},
			as:          s.emergencyTeamMember.EvmAddr,
			basicPass:   true,
			revert:      true,
			errContains: "sender is not owner",
			postCheck: func() {
				s.Require().Equal(
					[]uint8{ethereumChain, bitcoinChain},
					s.bridgeKeeper.GetBridgeOutChains(s.ctx),
				)
			},
		},
		{
			name: "chain outside the enum",
			run: func() []interface{} {
				s.setBridgeOutChains(ethereumChain, bitcoinChain)
				return []interface{}{unknownChain}
			},
			as:          s.account1.EvmAddr,
			basicPass:   true,
			revert:      true,
			errContains: "unsupported chain",
		},
		{
			name: "chain is not in the set",
			run: func() []interface{} {
				s.setBridgeOutChains(ethereumChain)
				return []interface{}{bitcoinChain}
			},
			as:          s.account1.EvmAddr,
			basicPass:   true,
			revert:      true,
			errContains: "chain is not enabled for bridge-outs",
		},
		{
			name: "invalid chain type - string instead of uint8",
			run: func() []interface{} {
				return []interface{}{"bitcoin"}
			},
			as:          s.account1.EvmAddr,
			basicPass:   false,
			errContains: "cannot use string as type",
		},
		{
			name: "happy path - remove Bitcoin",
			run: func() []interface{} {
				s.setBridgeOutChains(ethereumChain, bitcoinChain)
				return []interface{}{bitcoinChain}
			},
			as:        s.account1.EvmAddr,
			basicPass: true,
			output:    []interface{}{true},
			postCheck: func() {
				s.Require().Equal(
					[]uint8{ethereumChain},
					s.bridgeKeeper.GetBridgeOutChains(s.ctx),
				)

				arguments := s.requireEmittedEvent("BridgeOutChainRemoved")
				s.Require().Equal([]interface{}{bitcoinChain}, arguments)
			},
		},
		{
			name: "happy path - remove Ethereum",
			run: func() []interface{} {
				s.setBridgeOutChains(ethereumChain, bitcoinChain)
				return []interface{}{ethereumChain}
			},
			as:        s.account1.EvmAddr,
			basicPass: true,
			output:    []interface{}{true},
			postCheck: func() {
				s.Require().Equal(
					[]uint8{bitcoinChain},
					s.bridgeKeeper.GetBridgeOutChains(s.ctx),
				)

				arguments := s.requireEmittedEvent("BridgeOutChainRemoved")
				s.Require().Equal([]interface{}{ethereumChain}, arguments)
			},
		},
	}

	s.runMethodTestCases(testcases, "removeBridgeOutChain")
}

func (s *BridgeOutChainsTestSuite) TestAddBridgeOutChain() {
	testcases := []TestCase{
		{
			name: "caller is not owner",
			run: func() []interface{} {
				s.setBridgeOutChains(ethereumChain)
				return []interface{}{bitcoinChain}
			},
			as:          s.account2.EvmAddr,
			basicPass:   true,
			revert:      true,
			errContains: "sender is not owner",
			postCheck: func() {
				s.Require().Equal(
					[]uint8{ethereumChain},
					s.bridgeKeeper.GetBridgeOutChains(s.ctx),
				)
			},
		},
		{
			name: "caller is emergency team member",
			run: func() []interface{} {
				s.setBridgeOutChains(ethereumChain)
				return []interface{}{bitcoinChain}
			},
			as:          s.emergencyTeamMember.EvmAddr,
			basicPass:   true,
			revert:      true,
			errContains: "sender is not owner",
			postCheck: func() {
				s.Require().Equal(
					[]uint8{ethereumChain},
					s.bridgeKeeper.GetBridgeOutChains(s.ctx),
				)
			},
		},
		{
			name: "chain outside the enum",
			run: func() []interface{} {
				s.setBridgeOutChains(ethereumChain)
				return []interface{}{unknownChain}
			},
			as:          s.account1.EvmAddr,
			basicPass:   true,
			revert:      true,
			errContains: "unsupported chain",
		},
		{
			name: "chain is already in the set",
			run: func() []interface{} {
				s.setBridgeOutChains(ethereumChain, bitcoinChain)
				return []interface{}{bitcoinChain}
			},
			as:          s.account1.EvmAddr,
			basicPass:   true,
			revert:      true,
			errContains: "chain is already enabled for bridge-outs",
		},
		{
			name: "invalid chain type - string instead of uint8",
			run: func() []interface{} {
				return []interface{}{"bitcoin"}
			},
			as:          s.account1.EvmAddr,
			basicPass:   false,
			errContains: "cannot use string as type",
		},
		{
			name: "happy path - add Bitcoin back",
			run: func() []interface{} {
				s.setBridgeOutChains(ethereumChain)
				return []interface{}{bitcoinChain}
			},
			as:        s.account1.EvmAddr,
			basicPass: true,
			output:    []interface{}{true},
			postCheck: func() {
				s.Require().Equal(
					[]uint8{ethereumChain, bitcoinChain},
					s.bridgeKeeper.GetBridgeOutChains(s.ctx),
				)

				arguments := s.requireEmittedEvent("BridgeOutChainAdded")
				s.Require().Equal([]interface{}{bitcoinChain}, arguments)
			},
		},
		{
			name: "happy path - add a chain to an empty set",
			run: func() []interface{} {
				s.setBridgeOutChains()
				return []interface{}{ethereumChain}
			},
			as:        s.account1.EvmAddr,
			basicPass: true,
			output:    []interface{}{true},
			postCheck: func() {
				s.Require().Equal(
					[]uint8{ethereumChain},
					s.bridgeKeeper.GetBridgeOutChains(s.ctx),
				)

				arguments := s.requireEmittedEvent("BridgeOutChainAdded")
				s.Require().Equal([]interface{}{ethereumChain}, arguments)
			},
		},
	}

	s.runMethodTestCases(testcases, "addBridgeOutChain")
}

func (s *BridgeOutChainsTestSuite) TestGetBridgeOutChains() {
	testcases := []TestCase{
		{
			name: "both chains enabled",
			run: func() []interface{} {
				s.setBridgeOutChains(ethereumChain, bitcoinChain)
				return []interface{}{}
			},
			as:        s.account2.EvmAddr,
			basicPass: true,
			output:    []interface{}{[]uint8{ethereumChain, bitcoinChain}},
		},
		{
			name: "Bitcoin removed",
			run: func() []interface{} {
				s.setBridgeOutChains(ethereumChain)
				return []interface{}{}
			},
			as:        s.account2.EvmAddr,
			basicPass: true,
			output:    []interface{}{[]uint8{ethereumChain}},
		},
		{
			name: "empty set",
			run: func() []interface{} {
				s.setBridgeOutChains()
				return []interface{}{}
			},
			as:        s.account2.EvmAddr,
			basicPass: true,
			output:    []interface{}{[]uint8{}},
		},
		{
			name: "chains are returned in ascending order",
			run: func() []interface{} {
				s.setBridgeOutChains(bitcoinChain, ethereumChain)
				return []interface{}{}
			},
			as:        s.account2.EvmAddr,
			basicPass: true,
			output:    []interface{}{[]uint8{ethereumChain, bitcoinChain}},
			postCheck: func() {
				s.requireNoEmittedEvent()
			},
		},
	}

	s.runMethodTestCases(testcases, "getBridgeOutChains")
}

func (s *BridgeOutChainsTestSuite) TestBridgeOutChainsRoundTrip() {
	s.setBridgeOutChains(ethereumChain, bitcoinChain)

	contract := s.latestContract()

	s.Require().NoError(s.callMethod(
		contract,
		"removeBridgeOutChain",
		s.account1.EvmAddr,
		bitcoinChain,
	))
	s.Require().Equal(
		[]uint8{ethereumChain},
		s.bridgeKeeper.GetBridgeOutChains(s.ctx),
	)

	s.Require().NoError(s.callMethod(
		contract,
		"addBridgeOutChain",
		s.account1.EvmAddr,
		bitcoinChain,
	))
	s.Require().Equal(
		[]uint8{ethereumChain, bitcoinChain},
		s.bridgeKeeper.GetBridgeOutChains(s.ctx),
	)
}

// setBridgeOutChains replaces the whole set with the given chains. Every test
// case sets the state it needs, because the suite keeps one keeper for all
// subtests.
func (s *BridgeOutChainsTestSuite) setBridgeOutChains(chains ...uint8) {
	for _, chain := range s.bridgeKeeper.GetBridgeOutChains(s.ctx) {
		s.bridgeKeeper.DisableBridgeOutChain(s.ctx, chain)
	}

	for _, chain := range chains {
		s.bridgeKeeper.EnableBridgeOutChain(s.ctx, chain)
	}
}

// latestContract returns the latest version of the assets bridge precompile.
func (s *BridgeOutChainsTestSuite) latestContract() *precompile.Contract {
	versionMap, err := assetsbridge.NewPrecompileVersionMap(
		s.poaKeeper,
		s.bridgeKeeper,
		s.authzKeeper,
	)
	s.Require().NoError(err)

	contract, ok := versionMap.GetByVersion(
		evmtypes.AssetsBridgePrecompileLatestVersion,
	)
	s.Require().True(ok)

	return contract
}

// requireEmittedEvent asserts that the last precompile call emitted exactly one
// event with the given name. It returns the unpacked non-indexed arguments.
func (s *BridgeOutChainsTestSuite) requireEmittedEvent(
	eventName string,
) []interface{} {
	logs := s.stateDB.Logs()
	s.Require().Len(logs, 1, "expected a single emitted event")

	log := logs[0]
	s.Require().Equal(common.HexToAddress(assetsbridge.EvmAddress), log.Address)

	event, ok := s.contract.Abi.Events[eventName]
	s.Require().True(ok, "event %s not found in the ABI", eventName)
	s.Require().Equal(event.ID, log.Topics[0])
	s.Require().Len(log.Topics, 1, "expected no indexed arguments")

	arguments, err := event.Inputs.NonIndexed().Unpack(log.Data)
	s.Require().NoError(err)

	return arguments
}

// requireNoEmittedEvent asserts that the last precompile call emitted no events.
func (s *BridgeOutChainsTestSuite) requireNoEmittedEvent() {
	s.Require().Empty(s.stateDB.Logs())
}

// runMethodTestCases runs the given cases against the latest version of the
// assets bridge precompile and keeps the state database, so that a post check
// can assert the emitted events.
func (s *BridgeOutChainsTestSuite) runMethodTestCases(
	testcases []TestCase,
	methodName string,
) {
	for _, tc := range testcases {
		s.Run(tc.name, func() {
			s.stateDB = statedb.New(s.ctx, statedb.NewMockKeeper(), statedb.TxConfig{})
			evm := &vm.EVM{
				StateDB: s.stateDB,
			}

			s.contract = s.latestContract()

			var methodInputs []interface{}
			if tc.run != nil {
				methodInputs = tc.run()
			}

			method, ok := s.contract.Abi.Methods[methodName]
			s.Require().True(ok, "method %s not found in the ABI", methodName)

			methodInputArgs, err := method.Inputs.Pack(methodInputs...)

			if tc.basicPass {
				s.Require().NoError(err, "expected no error")
			} else {
				s.Require().Error(err, "expected error")
				s.Require().ErrorContains(err, tc.errContains, "expected different error message")
				return
			}

			vmContract := vm.NewPrecompile(tc.as, common.Address{}, nil, 0)
			vmContract.Input = append(vmContract.Input, method.ID...)
			vmContract.Input = append(vmContract.Input, methodInputArgs...)

			output, err := s.contract.Run(evm, vmContract, false)
			if tc.revert {
				s.Require().Error(err, "expected error")
				s.Require().ErrorContains(err, tc.errContains, "expected different error message")
				s.requireNoEmittedEvent()

				if tc.postCheck != nil {
					tc.postCheck()
				}

				return
			}
			s.Require().NoError(err, "expected no error")

			out, err := method.Outputs.Unpack(output)
			s.Require().NoError(err)
			s.Require().Len(out, len(tc.output))
			for i, expected := range tc.output {
				if expected, ok := expected.(*big.Int); ok {
					actual, _ := out[i].(*big.Int)
					s.Require().True(expected.Cmp(actual) == 0, "expected different value")
					continue
				}
				s.Require().Equal(expected, out[i], "expected different value")
			}

			if tc.postCheck != nil {
				tc.postCheck()
			}
		})
	}
}
