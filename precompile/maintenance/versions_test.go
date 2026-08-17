package maintenance_test

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/mezo-org/mezod/precompile"
	"github.com/mezo-org/mezod/precompile/maintenance"
	"github.com/mezo-org/mezod/x/evm/statedb"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
)

// callMethod runs the given method of the given precompile version and returns
// the execution error, if any.
func (s *PrecompileTestSuite) callMethod(
	contract *precompile.Contract,
	methodName string,
	as common.Address,
	inputs ...interface{},
) error {
	evm := &vm.EVM{
		StateDB: statedb.New(s.ctx, statedb.NewMockKeeper(), statedb.TxConfig{}),
	}

	method, ok := contract.Abi.Methods[methodName]
	s.Require().True(ok, "method %s not found in the ABI", methodName)

	methodInputArgs, err := method.Inputs.Pack(inputs...)
	s.Require().NoError(err)

	vmContract := vm.NewPrecompile(as, common.Address{}, nil, 0)
	vmContract.Input = append(vmContract.Input, method.ID...)
	vmContract.Input = append(vmContract.Input, methodInputArgs...)

	_, err = contract.Run(evm, vmContract, false)

	return err
}

func (s *PrecompileTestSuite) TestEmergencyControlsVersions() {
	versionMap, err := maintenance.NewPrecompileVersionMap(
		s.poaKeeper,
		s.evmKeeper,
		s.feeMarketKeeper,
		s.bridgeKeeper,
		s.upgradeKeeper,
	)
	s.Require().NoError(err)

	// The emergency controls are folded into the unreleased version 6.
	s.Require().Equal(6, evmtypes.MaintenancePrecompileLatestVersion)
	s.Require().Equal(
		evmtypes.MaintenancePrecompileLatestVersion,
		versionMap.GetLatestVersion(),
	)

	contractV5, ok := versionMap.GetByVersion(5)
	s.Require().True(ok)

	contractV6, ok := versionMap.GetByVersion(6)
	s.Require().True(ok)

	calls := []struct {
		methodName string
		inputs     []interface{}
	}{
		{"setEmergencyTeam", []interface{}{s.account3.EvmAddr}},
		{"getEmergencyTeam", nil},
		{"setBridgeLockdown", []interface{}{false, false}},
		{"getBridgeLockdown", nil},
		{"setTxLockdown", []interface{}{false}},
		{"getTxLockdown", nil},
		{
			"setTxLockdownAllowlist",
			[]interface{}{[]common.Address{}, []common.Address{}},
		},
		{"getTxLockdownAllowlist", nil},
		{"setChainLockdown", []interface{}{"v13.0.0"}},
	}

	for _, call := range calls {
		s.Run(call.methodName+" is not registered in v5", func() {
			err := s.callMethod(
				contractV5,
				call.methodName,
				s.account1.EvmAddr,
				call.inputs...,
			)
			s.Require().ErrorContains(err, "method not found in precompile")
		})

		s.Run(call.methodName+" is registered in v6", func() {
			err := s.callMethod(
				contractV6,
				call.methodName,
				s.account1.EvmAddr,
				call.inputs...,
			)
			s.Require().NoError(err)
		})
	}

	// The SELFDESTRUCT toggle folded into version 6 as well.
	s.Run("setSelfDestructDisabled is not registered in v5", func() {
		err := s.callMethod(
			contractV5,
			"setSelfDestructDisabled",
			s.account1.EvmAddr,
			true,
		)
		s.Require().ErrorContains(err, "method not found in precompile")
	})
}
