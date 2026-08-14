package assetsbridge_test

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/mezo-org/mezod/precompile"
	"github.com/mezo-org/mezod/precompile/assetsbridge"
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

func (s *PrecompileTestSuite) TestRetiredPauseMethodsVersions() {
	versionMap, err := assetsbridge.NewPrecompileVersionMap(
		s.poaKeeper,
		s.bridgeKeeper,
		&FakeAuthzKeeper{},
	)
	s.Require().NoError(err)

	// Version 6 retires the pause methods in favor of the maintenance
	// precompile.
	s.Require().Equal(6, evmtypes.AssetsBridgePrecompileLatestVersion)
	s.Require().Equal(
		evmtypes.AssetsBridgePrecompileLatestVersion,
		versionMap.GetLatestVersion(),
	)

	contractV5, ok := versionMap.GetByVersion(5)
	s.Require().True(ok)

	contractV6, ok := versionMap.GetByVersion(6)
	s.Require().True(ok)

	for methodName, inputs := range retiredPauseMethods {
		s.Run(methodName+" is registered in v5", func() {
			err := s.callMethod(
				contractV5,
				methodName,
				s.account1.EvmAddr,
				inputs...,
			)
			if err != nil {
				s.Require().NotContains(err.Error(), "method not found in precompile")
			}
		})

		s.Run(methodName+" is not registered in v6", func() {
			err := s.callMethod(
				contractV6,
				methodName,
				s.account1.EvmAddr,
				inputs...,
			)
			s.Require().ErrorContains(err, "method not found in precompile")
		})
	}

	// The methods kept by version 6 stay registered.
	s.Run("getTripartyLimits is registered in v6", func() {
		err := s.callMethod(
			contractV6,
			"getTripartyLimits",
			s.account1.EvmAddr,
		)
		s.Require().NoError(err)
	})
}
