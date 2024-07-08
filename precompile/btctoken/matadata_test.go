package btctoken_test

import (
	"github.com/evmos/evmos/v12/precompile"
	"github.com/evmos/evmos/v12/precompile/btctoken"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/evmos/evmos/v12/x/evm/statedb"
)

var validMetadata = banktypes.Metadata{
	Base:   "abtc",
	Name:   "abtc",
	Symbol: "BTC",
	DenomUnits: []*banktypes.DenomUnit{
		{
			Denom:    "abtc",
			Exponent: uint32(18),
		},
	},
}

func (s *PrecompileTestSuite) TestName() {
	bankKeeper := s.app.BankKeeper
	btcTokenPrecompile, err := btctoken.NewPrecompile(bankKeeper)
	s.Require().NoError(err)
	s.btcTokenPrecompile = btcTokenPrecompile

	s.app.BankKeeper.SetDenomMetaData(s.ctx, validMetadata)

	evm := &vm.EVM{
		StateDB: statedb.New(s.ctx, nil, statedb.TxConfig{}),
	}

	vmContract := vm.NewContract(&precompile.Contract{}, nil, nil, 0)
	// These first 4 bytes correspond to the method ID (first 4 bytes of the
	// Keccak-256 hash of the function signature).
	// In this case a function signature is 'function name()'
	vmContract.Input = []byte{0x06, 0xfd, 0xde, 0x03}
	output, err := s.btcTokenPrecompile.Run(evm, vmContract, true)
	s.Require().NoError(err)

	method := s.btcTokenPrecompile.Abi.Methods["name"]

	out, err := method.Outputs.Unpack(output)
	s.Require().NoError(err)
	s.Require().Equal(validMetadata.Name, out[0], "expected different name")
}

func (s *PrecompileTestSuite) TestSymbol() {
	bankKeeper := s.app.BankKeeper
	btcTokenPrecompile, err := btctoken.NewPrecompile(bankKeeper)
	s.Require().NoError(err)
	s.btcTokenPrecompile = btcTokenPrecompile

	s.app.BankKeeper.SetDenomMetaData(s.ctx, validMetadata)

	evm := &vm.EVM{
		StateDB: statedb.New(s.ctx, nil, statedb.TxConfig{}),
	}

	vmContract := vm.NewContract(&precompile.Contract{}, nil, nil, 0)
	// These first 4 bytes correspond to the method ID (first 4 bytes of the
	// Keccak-256 hash of the function signature).
	// In this case a function signature is 'function symbol()'
	vmContract.Input = []byte{0x95, 0xd8, 0x9b, 0x41}
	output, err := s.btcTokenPrecompile.Run(evm, vmContract, true)
	s.Require().NoError(err)

	method := s.btcTokenPrecompile.Abi.Methods["symbol"]

	out, err := method.Outputs.Unpack(output)
	s.Require().NoError(err)
	s.Require().Equal(validMetadata.Symbol, out[0], "expected different name")
}

func (s *PrecompileTestSuite) TestDecimals() {
	bankKeeper := s.app.BankKeeper
	btcTokenPrecompile, err := btctoken.NewPrecompile(bankKeeper)
	s.Require().NoError(err)
	s.btcTokenPrecompile = btcTokenPrecompile

	s.app.BankKeeper.SetDenomMetaData(s.ctx, validMetadata)

	evm := &vm.EVM{
		StateDB: statedb.New(s.ctx, nil, statedb.TxConfig{}),
	}

	vmContract := vm.NewContract(&precompile.Contract{}, nil, nil, 0)
	// These first 4 bytes correspond to the method ID (first 4 bytes of the
	// Keccak-256 hash of the function signature).
	// In this case a function signature is 'function decimals()'
	vmContract.Input = []byte{0x31, 0x3c, 0xe5, 0x67}
	output, err := s.btcTokenPrecompile.Run(evm, vmContract, true)
	s.Require().NoError(err)

	method := s.btcTokenPrecompile.Abi.Methods["decimals"]

	out, err := method.Outputs.Unpack(output)
	s.Require().NoError(err)
	s.Require().Equal(validMetadata.DenomUnits[0].Exponent, out[0], "expected different name")
}
