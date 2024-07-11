package btctoken_test

import (
	"github.com/evmos/evmos/v12/precompile"
	"github.com/evmos/evmos/v12/precompile/btctoken"

	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/evmos/evmos/v12/x/evm/statedb"
)

const (
	Name     = "BTC"
	Symbol   = "BTC"
	Decimals = uint8(18)
)

func (s *PrecompileTestSuite) setup() {
	bankKeeper := s.app.BankKeeper
	btcTokenPrecompile, err := btctoken.NewPrecompile(bankKeeper)
	s.Require().NoError(err)
	s.btcTokenPrecompile = btcTokenPrecompile
}

func (s *PrecompileTestSuite) runTest(input []byte, expected interface{}, methodName string) {
	evm := &vm.EVM{
		StateDB: statedb.New(s.ctx, nil, statedb.TxConfig{}),
	}

	vmContract := vm.NewContract(&precompile.Contract{}, nil, nil, 0)
	vmContract.Input = input
	output, err := s.btcTokenPrecompile.Run(evm, vmContract, true)
	s.Require().NoError(err)

	method := s.btcTokenPrecompile.Abi.Methods[methodName]

	out, err := method.Outputs.Unpack(output)
	s.Require().NoError(err)
	s.Require().Equal(expected, out[0], "expected different result")
}

func (s *PrecompileTestSuite) TestName() {
	s.setup()
	s.runTest([]byte{0x06, 0xfd, 0xde, 0x03}, Name, "name")
}

func (s *PrecompileTestSuite) TestSymbol() {
	s.setup()
	s.runTest([]byte{0x95, 0xd8, 0x9b, 0x41}, Symbol, "symbol")
}

func (s *PrecompileTestSuite) TestDecimals() {
	s.setup()
	s.runTest([]byte{0x31, 0x3c, 0xe5, 0x67}, Decimals, "decimals")
}
