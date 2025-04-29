package testsuite

import (
	"github.com/mezo-org/mezod/precompile"

	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/mezo-org/mezod/x/evm/statedb"
)

func (s *TestSuite) setupMetadataTest() {
	erc20Precompile, err := s.precompileFactoryFn(s.app)
	s.Require().NoError(err)
	s.erc20Precompile = erc20Precompile
}

func (s *TestSuite) runMetadataTest(input []byte, expected interface{}, methodName string) {
	evm := &vm.EVM{
		StateDB: statedb.New(s.ctx, statedb.NewMockKeeper(), statedb.TxConfig{}),
	}

	vmContract := vm.NewContract(&precompile.Contract{}, nil, nil, 0)
	vmContract.Input = input
	output, err := s.erc20Precompile.Run(evm, vmContract, true)
	s.Require().NoError(err)

	method := s.erc20Precompile.Abi.Methods[methodName]

	out, err := method.Outputs.Unpack(output)
	s.Require().NoError(err)
	s.Require().Equal(expected, out[0], "expected different result")
}

func (s *TestSuite) TestName() {
	s.setupMetadataTest()
	s.runMetadataTest([]byte{0x06, 0xfd, 0xde, 0x03}, s.name, "name")
}

func (s *TestSuite) TestSymbol() {
	s.setupMetadataTest()
	s.runMetadataTest([]byte{0x95, 0xd8, 0x9b, 0x41}, s.symbol, "symbol")
}

func (s *TestSuite) TestDecimals() {
	s.setupMetadataTest()
	s.runMetadataTest([]byte{0x31, 0x3c, 0xe5, 0x67}, s.decimals, "decimals")
}
