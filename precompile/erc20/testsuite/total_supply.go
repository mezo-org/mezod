package testsuite

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/mezo-org/mezod/app"
	"github.com/mezo-org/mezod/precompile"

	sdkmath "cosmossdk.io/math"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/mezo-org/mezod/x/evm/statedb"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
)

func (s *TestSuite) TestTotalSupply() {
	testcases := map[string]struct {
		run      func(sdk.Context, *app.Mezo)
		expTotal *big.Int
	}{
		// This is minted by the existing test_helpers.go file for the generated
		// account in the Setup() function. 10^14 is minted to the genesis account
		// resulting in 100000000000000 of the existing total coins.
		"existing total supply - no coins minted": {
			expTotal: big.NewInt(100000000000000),
		},
		"add to total supply - mint more coins": {
			expTotal: big.NewInt(100000000000042),
			run: func(ctx sdk.Context, app *app.Mezo) {
				// Mint more coins to the evm module
				err := app.BankKeeper.MintCoins(
					ctx, evmtypes.ModuleName,
					sdk.Coins{sdk.NewCoin(s.denom, sdkmath.NewInt(42))})
				s.Require().NoError(err)
			},
		},
	}

	for testName, tc := range testcases {
		s.Run(testName, func() {
			s.SetupTest()
			if tc.run != nil {
				tc.run(s.ctx, s.app)
			}

			evm := &vm.EVM{
				StateDB: statedb.New(s.ctx, statedb.NewMockKeeper(), statedb.TxConfig{}),
			}

			erc20Precompile, err := s.precompileFactoryFn(s.app)
			s.Require().NoError(err)
			s.erc20Precompile = erc20Precompile

			vmContract := vm.NewContract(&precompile.Contract{}, nil, nil, 0)
			// These first 4 bytes correspond to the method ID (first 4 bytes of the
			// Keccak-256 hash of the function signature).
			// In this case a function signature is 'function totalSupply()'
			vmContract.Input = []byte{0x18, 0x16, 0x0d, 0xdd}
			output, err := s.erc20Precompile.Run(evm, vmContract, true)
			s.Require().NoError(err)

			method := s.erc20Precompile.Abi.Methods["totalSupply"]

			out, err := method.Outputs.Unpack(output)
			s.Require().NoError(err)
			s.Require().Equal(tc.expTotal, out[0], "expected different value")
		})
	}
}
