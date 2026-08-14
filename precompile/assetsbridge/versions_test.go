package assetsbridge_test

import (
	"fmt"

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
	method, ok := contract.Abi.Methods[methodName]
	s.Require().True(ok, "method %s not found in the ABI", methodName)

	methodInputArgs, err := method.Inputs.Pack(inputs...)
	s.Require().NoError(err)

	return s.callRawInput(contract, as, append(method.ID, methodInputArgs...))
}

// callRawInput runs the given precompile version with the given raw call input
// and returns the execution error, if any.
func (s *PrecompileTestSuite) callRawInput(
	contract *precompile.Contract,
	as common.Address,
	input []byte,
) error {
	evm := &vm.EVM{
		StateDB: statedb.New(s.ctx, statedb.NewMockKeeper(), statedb.TxConfig{}),
	}

	vmContract := vm.NewPrecompile(as, common.Address{}, nil, 0)
	vmContract.Input = input

	_, err := contract.Run(evm, vmContract, false)

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

	// The retired pause methods are removed from the shared ABI, so calls to
	// their selectors revert on every version.
	retiredPauseSelectors := map[string][]byte{
		"setPauser(address)":  {0x2d, 0x88, 0xaf, 0x4a},
		"getPauser()":         {0x70, 0x08, 0xb5, 0x48},
		"pauseBridgeOut()":    {0x2f, 0x1d, 0x44, 0x8f},
		"pauseTriparty(bool)": {0x9a, 0xfc, 0x02, 0xff},
		"isTripartyPaused()":  {0x5d, 0x1b, 0x3f, 0xc4},
	}

	for version := 0; version <= versionMap.GetLatestVersion(); version++ {
		contract, ok := versionMap.GetByVersion(version)
		s.Require().True(ok)

		for signature, selector := range retiredPauseSelectors {
			s.Run(fmt.Sprintf("%s reverts in v%d", signature, version), func() {
				err := s.callRawInput(
					contract,
					s.account1.EvmAddr,
					selector,
				)
				s.Require().ErrorContains(err, "method not found in ABI")
			})
		}
	}

	// The methods kept by version 6 stay registered.
	contractV6, ok := versionMap.GetByVersion(6)
	s.Require().True(ok)

	s.Run("getTripartyLimits is registered in v6", func() {
		err := s.callMethod(
			contractV6,
			"getTripartyLimits",
			s.account1.EvmAddr,
		)
		s.Require().NoError(err)
	})
}
