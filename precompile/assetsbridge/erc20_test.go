package assetsbridge_test

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	bridgetypes "github.com/mezo-org/mezod/x/bridge/types"
)

var (
	testSourceToken = common.HexToAddress("0x0d2Ff728BA2CA86a4E9883af35f2aeE6a6a85e11")
	testMezoToken   = common.HexToAddress("0xeDc4c5Af4424d0138375C31Df216Ec7A92AF2cA5")
)

type mappingDescriptor = struct {
	SourceToken common.Address "json:\"sourceToken\""
	MezoToken   common.Address "json:\"mezoToken\""
}

func (s *PrecompileTestSuite) TestCreateERC20TokenMapping() {
	testcases := []TestCase{
		{
			name: "caller is not owner",
			run: func() []interface{} {
				return []interface{}{
					testSourceToken,
					testMezoToken,
				}
			},
			as:          s.account2.EvmAddr, // account2 is not the owner
			basicPass:   true,
			revert:      true,
			errContains: "sender is not owner",
		},
		{
			name: "happy path",
			run: func() []interface{} {
				return []interface{}{
					testSourceToken,
					testMezoToken,
				}
			},
			as:        s.account1.EvmAddr, // account1 is the owner
			basicPass: true,
			output:    []interface{}{true},
			postCheck: func() {
				mappings := s.bridgeKeeper.GetERC20TokensMappings(s.ctx)
				s.Require().Len(mappings, 1)
				s.Require().Equal(testSourceToken.Hex(), mappings[0].SourceToken)
				s.Require().Equal(testMezoToken.Hex(), mappings[0].MezoToken)
			},
		},
	}

	s.RunMethodTestCases(testcases, "createERC20TokenMapping")
}

func (s *PrecompileTestSuite) TestDeleteERC20TokenMapping() {
	testcases := []TestCase{
		{
			name: "caller is not owner",
			run: func() []interface{} {
				return []interface{}{
					testSourceToken,
				}
			},
			as:          s.account2.EvmAddr, // account2 is not the owner
			basicPass:   true,
			revert:      true,
			errContains: "sender is not owner",
		},
		{
			name: "happy path",
			run: func() []interface{} {
				// Create a mapping first.
				err := s.bridgeKeeper.CreateERC20TokenMapping(
					s.ctx,
					testSourceToken.Bytes(),
					testMezoToken.Bytes(),
				)
				s.Require().NoError(err)

				return []interface{}{
					testSourceToken,
				}
			},
			as:        s.account1.EvmAddr, // account1 is the owner
			basicPass: true,
			output:    []interface{}{true},
			postCheck: func() {
				mappings := s.bridgeKeeper.GetERC20TokensMappings(s.ctx)
				s.Require().Len(mappings, 0)
			},
		},
	}

	s.RunMethodTestCases(testcases, "deleteERC20TokenMapping")
}

func (s *PrecompileTestSuite) TestGetERC20TokenMapping() {
	testcases := []TestCase{
		{
			name: "happy path",
			run: func() []interface{} {
				// Create a mapping first.
				err := s.bridgeKeeper.CreateERC20TokenMapping(
					s.ctx,
					testSourceToken.Bytes(),
					testMezoToken.Bytes(),
				)
				s.Require().NoError(err)

				return []interface{}{
					testSourceToken,
				}
			},
			basicPass: true,
			output: []interface{}{
				mappingDescriptor{
					testSourceToken,
					testMezoToken,
				},
			},
		},
	}

	s.RunMethodTestCases(testcases, "getERC20TokenMapping")
}

func (s *PrecompileTestSuite) TestGetERC20TokensMappings() {
	otherSourceToken := common.HexToAddress("0xE1210B4E5E6D97e2B3eE694543649D7f344F6329")
	otherMezoToken := common.HexToAddress("0xBA2806dc37bFe04D7C8B4EdbD9B8814ab028a7BD")

	testcases := []TestCase{
		{
			name: "happy path",
			run: func() []interface{} {
				// Create some mappings first.
				err := s.bridgeKeeper.CreateERC20TokenMapping(
					s.ctx,
					testSourceToken.Bytes(),
					testMezoToken.Bytes(),
				)
				s.Require().NoError(err)

				err = s.bridgeKeeper.CreateERC20TokenMapping(
					s.ctx,
					otherSourceToken.Bytes(),
					otherMezoToken.Bytes(),
				)
				s.Require().NoError(err)

				return nil
			},
			basicPass: true,
			output: []interface{}{
				[]mappingDescriptor{
					{
						testSourceToken,
						testMezoToken,
					},
					{
						otherSourceToken,
						otherMezoToken,
					},
				},
			},
		},
	}

	s.RunMethodTestCases(testcases, "getERC20TokensMappings")
}

func (s *PrecompileTestSuite) TestGetMaxERC20TokensMappings() {
	testcases := []TestCase{
		{
			name: "happy path",
			run: func() []interface{} {
				return nil
			},
			basicPass: true,
			output:    []interface{}{big.NewInt(int64(bridgetypes.DefaultParams().MaxErc20TokensMappings))},
		},
	}

	s.RunMethodTestCases(testcases, "getMaxERC20TokensMappings")
}
