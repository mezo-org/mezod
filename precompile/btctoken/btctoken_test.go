package btctoken_test

import (
	"testing"

	"github.com/mezo-org/mezod/app"
	"github.com/mezo-org/mezod/precompile"
	"github.com/mezo-org/mezod/precompile/btctoken"
	erc20testsuite "github.com/mezo-org/mezod/precompile/erc20/testsuite"
	"github.com/stretchr/testify/suite"
)

const (
	Denom    = "abtc"
	Name     = "BTC"
	Symbol   = "BTC"
	Decimals = uint8(18)
)

// This domain separator was computed using the following Solidity code:
//
//	keccak256(
//		abi.encode(
//			keccak256("EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)"),
//			keccak256("BTC"),
//			keccak256("1"),
//			31612,
//			0x7b7C000000000000000000000000000000000000
//		)
//	);
//
// Same in hex: f98315225d67e6d98b18f0a6b73bf711423ff8310bd350400db36e876c5cddf4
var DomainSeparator = []byte{
	0xf9, 0x83, 0x15, 0x22, 0x5d, 0x67, 0xe6, 0xd9, 0x8b, 0x18, 0xf0, 0xa6, 0xb7, 0x3b, 0xf7, 0x11,
	0x42, 0x3f, 0xf8, 0x31, 0x0b, 0xd3, 0x50, 0x40, 0x0d, 0xb3, 0x6e, 0x87, 0x6c, 0x5c, 0xdd, 0xf4,
}

func TestBTCPrecompile(t *testing.T) {
	precompileFactoryFn := func(app *app.Mezo) (*precompile.Contract, error) {
		return btctoken.NewPrecompile(app.BankKeeper, app.AuthzKeeper, *app.EvmKeeper, "mezo_31612-1")
	}

	suiteInstance := erc20testsuite.New(
		Denom,
		Name,
		Symbol,
		Decimals,
		DomainSeparator,
		precompileFactoryFn,
		true,
	)

	suite.Run(t, suiteInstance)
}
