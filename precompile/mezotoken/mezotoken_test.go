package mezotoken_test

import (
	"testing"

	"github.com/mezo-org/mezod/app"
	"github.com/mezo-org/mezod/precompile"
	erc20testsuite "github.com/mezo-org/mezod/precompile/erc20/testsuite"
	"github.com/mezo-org/mezod/precompile/mezotoken"
	"github.com/stretchr/testify/suite"
)

const (
	Denom    = "amezo"
	Name     = "MEZO"
	Symbol   = "MEZO"
	Decimals = uint8(18)
)

// This domain separator was computed using the following Solidity code:
//
//	keccak256(
//		abi.encode(
//			keccak256("EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)"),
//			keccak256("MEZO"),
//			keccak256("1"),
//			31612,
//			0x7B7c000000000000000000000000000000000001
//		)
//	);
//
// Same in hex: 43d4db585ef58130660466546b06e8802f4479372409cf0636169cfd80f6a186
var DomainSeparator = []byte{
	0x43, 0xd4, 0xdb, 0x58, 0x5e, 0xf5, 0x81, 0x30, 0x66, 0x04, 0x66, 0x54, 0x6b, 0x06, 0xe8, 0x80,
	0x2f, 0x44, 0x79, 0x37, 0x24, 0x09, 0xcf, 0x06, 0x36, 0x16, 0x9c, 0xfd, 0x80, 0xf6, 0xa1, 0x86,
}

func TestMEZOPrecompile(t *testing.T) {
	precompileFactoryFn := func(app *app.Mezo) (*precompile.Contract, error) {
		return mezotoken.NewPrecompile(app.BankKeeper, app.AuthzKeeper, *app.EvmKeeper, "mezo_31612-1")
	}

	suiteInstance := erc20testsuite.New(
		Denom,
		Name,
		Symbol,
		Decimals,
		DomainSeparator,
		precompileFactoryFn,
		false,
	)

	suite.Run(t, suiteInstance)
}
