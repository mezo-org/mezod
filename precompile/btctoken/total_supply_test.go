package btctoken_test

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v12/app"
	"github.com/evmos/evmos/v12/precompile"
	"github.com/evmos/evmos/v12/precompile/btctoken"
	"github.com/evmos/evmos/v12/utils"

	sdkmath "cosmossdk.io/math"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/evmos/evmos/v12/x/evm/statedb"
	evmtypes "github.com/evmos/evmos/v12/x/evm/types"
)

func (s *PrecompileTestSuite) TestTotalSupply() {
	testcases := map[string]struct {
		run      func(sdk.Context, *app.Evmos)
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
			run: func(ctx sdk.Context, app *app.Evmos) {
				// Mint more coins to the evm module
				err := app.BankKeeper.MintCoins(
					ctx, evmtypes.ModuleName,
					sdk.Coins{sdk.NewCoin(utils.BaseDenom, sdkmath.NewInt(42))})
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
				StateDB: statedb.New(s.ctx, nil, statedb.TxConfig{}),
			}

			bankKeeper := s.app.BankKeeper
			authzKeeper := s.app.AuthzKeeper
			emvKeeper := *s.app.EvmKeeper
			btcTokenPrecompile, err := btctoken.NewPrecompile(bankKeeper, authzKeeper, emvKeeper)
			s.Require().NoError(err)
			s.btcTokenPrecompile = btcTokenPrecompile

			vmContract := vm.NewContract(&precompile.Contract{}, nil, nil, 0)
			// These first 4 bytes correspond to the method ID (first 4 bytes of the
			// Keccak-256 hash of the function signature).
			// In this case a function signature is 'function totalSupply()'
			vmContract.Input = []byte{0x18, 0x16, 0x0d, 0xdd}
			output, err := s.btcTokenPrecompile.Run(evm, vmContract, true)
			s.Require().NoError(err)

			method := s.btcTokenPrecompile.Abi.Methods["totalSupply"]

			out, err := method.Outputs.Unpack(output)
			s.Require().NoError(err)
			s.Require().Equal(tc.expTotal, out[0], "expected different value")
		})
	}
}
