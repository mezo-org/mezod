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

// keccak256(encode(
//
//			keccak256(
//				"EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)"
//			),
//			keccak256(BTC),
//			keccak256(1),
//			31612,
//			0x7b7C000000000000000000000000000000000000
//		)
//	)
//
// Same in hex: f98315225d67e6d98b18f0a6b73bf711423ff8310bd350400db36e876c5cddf4
var DomainSeparator = []byte{
	249, 131, 21, 34, 93, 103, 230, 217, 139, 24, 240, 166, 183, 59, 247, 17,
	66, 63, 248, 49, 11, 211, 80, 64, 13, 179, 110, 135, 108, 92, 221, 244,
}

func TestPrecompileTestSuite(t *testing.T) {
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
