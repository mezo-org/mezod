package testsuite

import (
	"math/big"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/core/vm"

	sdkmath "cosmossdk.io/math"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/mezo-org/mezod/precompile"
	"github.com/mezo-org/mezod/precompile/erc20"
	"github.com/mezo-org/mezod/x/evm/statedb"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
)

func (s *TestSuite) TestTransfer() {
	testcases := []struct {
		name          string
		run           func() []interface{}
		postCheck     func()
		basicPass     bool
		isCallerOwner bool
		errContains   string
	}{
		{
			name:        "empty args",
			run:         func() []interface{} { return nil },
			errContains: "argument count mismatch",
		},
		{
			name: "argument count mismatch",
			run: func() []interface{} {
				return []interface{}{
					1, 2, 3,
				}
			},
			errContains: "argument count mismatch",
		},
		{
			name: "invalid address",
			run: func() []interface{} {
				return []interface{}{
					"invalid address", big.NewInt(2),
				}
			},
			errContains: "cannot use string as type array as argument",
		},
		{
			name: "invalid amount",
			run: func() []interface{} {
				return []interface{}{
					s.account1.EvmAddr, "invalid amount",
				}
			},
			errContains: "cannot use string as type ptr as argument",
		},
		{
			name: "not enough balance",
			run: func() []interface{} {
				// Mint some coins to the module account and then send to the from address
				err := s.app.BankKeeper.MintCoins(s.ctx, evmtypes.ModuleName, sdk.Coins{sdk.NewCoin(s.denom, sdkmath.NewInt(1e18))})
				s.Require().NoError(err, "failed to mint coins")
				err = s.app.BankKeeper.SendCoinsFromModuleToAccount(s.ctx, evmtypes.ModuleName, s.account1.EvmAddr.Bytes(), sdk.Coins{sdk.NewCoin(s.denom, sdkmath.NewInt(42))})
				s.Require().NoError(err, "failed to send coins from module to account")

				return []interface{}{
					s.account2.EvmAddr, big.NewInt(43),
				}
			},
			basicPass:   true,
			errContains: "insufficient funds",
		},
		{
			name: "transfer zero amount",
			run: func() []interface{} {
				// Mint some coins to the module account and then send to the from address
				err := s.app.BankKeeper.MintCoins(s.ctx, evmtypes.ModuleName, sdk.Coins{sdk.NewCoin(s.denom, sdkmath.NewInt(1e18))})
				s.Require().NoError(err, "failed to mint coins")
				err = s.app.BankKeeper.SendCoinsFromModuleToAccount(s.ctx, evmtypes.ModuleName, s.account1.EvmAddr.Bytes(), sdk.Coins{sdk.NewCoin(s.denom, sdkmath.NewInt(42))})
				s.Require().NoError(err, "failed to send coins from module to account")

				return []interface{}{
					s.account2.EvmAddr, big.NewInt(0),
				}
			},
			basicPass: true,
			postCheck: func() {
				// Check the balance of the two accounts
				acc1 := s.app.BankKeeper.GetBalance(s.ctx, s.account1.SdkAddr, s.denom)
				acc2 := s.app.BankKeeper.GetBalance(s.ctx, s.account2.SdkAddr, s.denom)
				s.Require().Equal(sdkmath.NewInt(42), acc1.Amount)
				s.Require().Equal(sdkmath.NewInt(0), acc2.Amount)
			},
		},
		{
			name: "successful transfer",
			run: func() []interface{} {
				// Mint some coins to the module account and then send to the from address
				err := s.app.BankKeeper.MintCoins(s.ctx, evmtypes.ModuleName, sdk.Coins{sdk.NewCoin(s.denom, sdkmath.NewInt(1e18))})
				s.Require().NoError(err, "failed to mint coins")
				err = s.app.BankKeeper.SendCoinsFromModuleToAccount(s.ctx, evmtypes.ModuleName, s.account1.EvmAddr.Bytes(), sdk.Coins{sdk.NewCoin(s.denom, sdkmath.NewInt(42))})
				s.Require().NoError(err, "failed to send coins from module to account")

				return []interface{}{
					s.account2.EvmAddr, big.NewInt(40),
				}
			},
			basicPass: true,
			postCheck: func() {
				// Check the balance of the two accounts
				acc1 := s.app.BankKeeper.GetBalance(s.ctx, s.account1.SdkAddr, s.denom)
				acc2 := s.app.BankKeeper.GetBalance(s.ctx, s.account2.SdkAddr, s.denom)
				s.Require().Equal(sdkmath.NewInt(2), acc1.Amount)
				s.Require().Equal(sdkmath.NewInt(40), acc2.Amount)
			},
		},
	}

	for _, tc := range testcases {
		s.Run(tc.name, func() {
			s.SetupTest()
			evm := &vm.EVM{
				StateDB: statedb.New(s.ctx, s.app.EvmKeeper, statedb.TxConfig{}),
			}

			getEVMBalance := func(address common.Address) sdkmath.Int {
				return sdkmath.NewIntFromBigInt(evm.StateDB.GetBalance(address).ToBig())
			}

			erc20Precompile, err := s.precompileFactoryFn(s.app)
			s.Require().NoError(err)
			s.erc20Precompile = erc20Precompile

			var methodInputs []interface{}
			if tc.run != nil {
				methodInputs = tc.run()
			}

			method := s.erc20Precompile.Abi.Methods["transfer"]
			var methodInputArgs []byte
			methodInputArgs, err = method.Inputs.Pack(methodInputs...)

			if tc.basicPass {
				s.Require().NoError(err, "expected no error")
			} else {
				s.Require().Error(err, "expected error")
				s.Require().ErrorContains(err, tc.errContains, "expected different error message")
				return
			}

			vmContract := vm.NewContract(&precompile.Contract{}, nil, nil, 0)
			// These first 4 bytes correspond to the method ID (first 4 bytes of the
			// Keccak-256 hash of the function signature).
			// In this case a function signature is 'function transfer(address to, uint256 value)'
			vmContract.Input = append([]byte{0xa9, 0x05, 0x9c, 0xbb}, methodInputArgs...)
			vmContract.CallerAddress = s.account1.EvmAddr

			// Capture EVM balances of the from/to addresses before the transfer.
			// This must be done in this particular place to capture the balance
			// changes made in the test setup.
			initialEVMBalanceFrom := getEVMBalance(vmContract.CallerAddress)
			initialEVMBalanceTo := getEVMBalance(methodInputs[0].(common.Address))

			output, err := s.erc20Precompile.Run(evm, vmContract, false)
			if err != nil {
				s.Require().ErrorContains(err, tc.errContains, "expected different error message")
				return
			}
			s.Require().NoError(err, "expected no error")

			out, err := method.Outputs.Unpack(output)
			s.Require().NoError(err)
			s.Require().Equal(true, out[0], "expected different value")

			evmBalanceFrom := getEVMBalance(vmContract.CallerAddress)
			evmBalanceTo := getEVMBalance(methodInputs[0].(common.Address))

			if s.ensureEVMBalanceChange {
				// We ensure Cosmos balance changes made by the precompile are propagated
				// back to the EVM stateDB. This is the case for the EVM native gas token.
				// We expect that the `from` account balance decreased by the amount transferred
				// and the `to` account balance increased by the same amount.
				amountTransferred := sdkmath.NewIntFromBigInt(methodInputs[1].(*big.Int))

				s.Require().True(evmBalanceFrom.Equal(initialEVMBalanceFrom.Sub(amountTransferred)), "expected evm balance from decrease")
				s.Require().True(evmBalanceTo.Equal(initialEVMBalanceTo.Add(amountTransferred)), "expected evm balance to increase")
			} else {
				// If we do not expect balance changes to be propagated to the EVM,
				// we ensure that the EVM balances are unchanged.
				s.Require().True(evmBalanceFrom.Equal(initialEVMBalanceFrom), "expected evm balance from to be unchanged")
				s.Require().True(evmBalanceTo.Equal(initialEVMBalanceTo), "expected evm balance to be unchanged")
			}

			// we call  the statedb commit here to simulate end of transaction
			// processing and flush the cache context
			s.Require().NoError(evm.StateDB.(*statedb.StateDB).Commit())

			if tc.postCheck != nil {
				tc.postCheck()
			}
		})
	}
}

func (s *TestSuite) TestTransferFrom() {
	testcases := []struct {
		name          string
		run           func() []interface{}
		postCheck     func()
		basicPass     bool
		isCallerOwner bool
		errContains   string
	}{
		{
			name:        "empty args",
			run:         func() []interface{} { return nil },
			errContains: "argument count mismatch",
		},
		{
			name: "argument count mismatch",
			run: func() []interface{} {
				return []interface{}{
					1, 2,
				}
			},
			errContains: "argument count mismatch",
		},
		{
			name: "invalid address",
			run: func() []interface{} {
				return []interface{}{
					"invalid address", big.NewInt(2), big.NewInt(3),
				}
			},
			errContains: "cannot use string as type array as argument",
		},
		{
			name: "invalid amount",
			run: func() []interface{} {
				return []interface{}{
					s.account1.EvmAddr, "invalid amount", big.NewInt(3),
				}
			},
			errContains: "cannot use string as type array as argument",
		},
		{
			name: "invalid spender",
			run: func() []interface{} {
				return []interface{}{
					s.account1.EvmAddr, big.NewInt(2), "invalid spender",
				}
			},
			errContains: "cannot use ptr as type array as argument",
		},
		{
			name: "owner is spender - pass",
			run: func() []interface{} {
				// Mint some coins to the module account and then send to the from address
				err := s.app.BankKeeper.MintCoins(s.ctx, evmtypes.ModuleName, sdk.Coins{sdk.NewCoin(s.denom, sdkmath.NewInt(1e18))})
				s.Require().NoError(err, "failed to mint coins")
				err = s.app.BankKeeper.SendCoinsFromModuleToAccount(s.ctx, evmtypes.ModuleName, s.account1.EvmAddr.Bytes(), sdk.Coins{sdk.NewCoin(s.denom, sdkmath.NewInt(42))})
				s.Require().NoError(err, "failed to send coins from module to account")

				return []interface{}{
					s.account1.EvmAddr, s.account2.EvmAddr, big.NewInt(40),
				}
			},
			basicPass:     true,
			isCallerOwner: true,
			postCheck: func() {
				// Check the balance of the two accounts
				acc1 := s.app.BankKeeper.GetBalance(s.ctx, s.account1.SdkAddr, s.denom)
				acc2 := s.app.BankKeeper.GetBalance(s.ctx, s.account2.SdkAddr, s.denom)
				s.Require().Equal(sdkmath.NewInt(2), acc1.Amount)
				s.Require().Equal(sdkmath.NewInt(40), acc2.Amount)
			},
		},
		{
			name: "owner is not a spender - not enough allowance",
			run: func() []interface{} {
				// Mint some coins to the module account and then send to the from address
				err := s.app.BankKeeper.MintCoins(s.ctx, evmtypes.ModuleName, sdk.Coins{sdk.NewCoin(s.denom, sdkmath.NewInt(1e18))})
				s.Require().NoError(err, "failed to mint coins")
				err = s.app.BankKeeper.SendCoinsFromModuleToAccount(s.ctx, evmtypes.ModuleName, s.account1.EvmAddr.Bytes(), sdk.Coins{sdk.NewCoin(s.denom, sdkmath.NewInt(42))})
				s.Require().NoError(err, "failed to send coins from module to account")

				return []interface{}{
					s.account1.EvmAddr, s.account2.EvmAddr, big.NewInt(40),
				}
			},
			errContains:   "authorization type does not exist or is expired for address",
			basicPass:     true,
			isCallerOwner: false,
		},
		{
			name: "owner is not a spender - successful transferFrom",
			run: func() []interface{} {
				// Mint some coins to the module account and then send to the from address
				err := s.app.BankKeeper.MintCoins(s.ctx, evmtypes.ModuleName, sdk.Coins{sdk.NewCoin(s.denom, sdkmath.NewInt(1e18))})
				s.Require().NoError(err, "failed to mint coins")
				amount := sdk.Coins{sdk.NewCoin(s.denom, sdkmath.NewInt(42))}
				err = s.app.BankKeeper.SendCoinsFromModuleToAccount(s.ctx, evmtypes.ModuleName, s.account1.EvmAddr.Bytes(), amount)
				s.Require().NoError(err, "failed to send coins from module to account")

				// Approve the spender (acct2) to spend from the from (acct1) address
				sendAuthz := banktypes.NewSendAuthorization(amount, nil)
				expiration := s.ctx.BlockTime().Add(time.Hour * 24 * 365)
				err = s.app.AuthzKeeper.SaveGrant(s.ctx, s.account2.EvmAddr.Bytes(), s.account1.EvmAddr.Bytes(), sendAuthz, &expiration)
				s.Require().NoError(err, "expected no error saving the grant")

				return []interface{}{
					s.account1.EvmAddr, s.account2.EvmAddr, big.NewInt(41),
				}
			},
			basicPass:     true,
			isCallerOwner: false,
			postCheck: func() {
				// Check the balance of the two accounts
				acc1 := s.app.BankKeeper.GetBalance(s.ctx, s.account1.SdkAddr, s.denom)
				acc2 := s.app.BankKeeper.GetBalance(s.ctx, s.account2.SdkAddr, s.denom)
				s.Require().Equal(sdkmath.NewInt(1), acc1.Amount)
				s.Require().Equal(sdkmath.NewInt(41), acc2.Amount)
			},
		},
	}

	for _, tc := range testcases {
		s.Run(tc.name, func() {
			s.SetupTest()
			evm := &vm.EVM{
				StateDB: statedb.New(s.ctx, s.app.EvmKeeper, statedb.TxConfig{}),
			}

			getEVMBalance := func(address common.Address) sdkmath.Int {
				return sdkmath.NewIntFromBigInt(evm.StateDB.GetBalance(address).ToBig())
			}

			erc20Precompile, err := s.precompileFactoryFn(s.app)
			s.Require().NoError(err)
			s.erc20Precompile = erc20Precompile

			var methodInputs []interface{}
			if tc.run != nil {
				methodInputs = tc.run()
			}

			method := s.erc20Precompile.Abi.Methods["transferFrom"]
			var methodInputArgs []byte
			methodInputArgs, err = method.Inputs.Pack(methodInputs...)

			if tc.basicPass {
				s.Require().NoError(err, "expected no error")
			} else {
				s.Require().Error(err, "expected error")
				s.Require().ErrorContains(err, tc.errContains, "expected different error message")
				return
			}

			vmContract := vm.NewContract(&precompile.Contract{}, nil, nil, 0)
			// These first 4 bytes correspond to the method ID (first 4 bytes of the
			// Keccak-256 hash of the function signature).
			// In this case a function signature is 'function transferFrom(address from, address to, uint256 value)'
			vmContract.Input = append([]byte{0x23, 0xb8, 0x72, 0xdd}, methodInputArgs...)
			if tc.isCallerOwner {
				vmContract.CallerAddress = s.account1.EvmAddr
			} else {
				vmContract.CallerAddress = s.account2.EvmAddr
			}

			// Capture EVM balances of the from/to addresses before the transfer.
			// This must be done in this particular place to capture the balance
			// changes made in the test setup.
			initialEVMBalanceFrom := getEVMBalance(methodInputs[0].(common.Address))
			initialEVMBalanceTo := getEVMBalance(methodInputs[1].(common.Address))

			output, err := s.erc20Precompile.Run(evm, vmContract, false)
			if err != nil {

				s.Require().ErrorContains(err, tc.errContains, "expected different error message")
				return
			}
			s.Require().NoError(err, "expected no error")

			out, err := method.Outputs.Unpack(output)
			s.Require().NoError(err)
			s.Require().Equal(true, out[0], "expected different value")

			evmBalanceFrom := getEVMBalance(methodInputs[0].(common.Address))
			evmBalanceTo := getEVMBalance(methodInputs[1].(common.Address))

			if s.ensureEVMBalanceChange {
				// We ensure Cosmos balance changes made by the precompile are propagated
				// back to the EVM stateDB. This is the case for the EVM native gas token.
				// We expect that the `from` account balance decreased by the amount transferred
				// and the `to` account balance increased by the same amount.
				amountTransferred := sdkmath.NewIntFromBigInt(methodInputs[2].(*big.Int))

				s.Require().True(evmBalanceFrom.Equal(initialEVMBalanceFrom.Sub(amountTransferred)), "expected evm balance from decrease")
				s.Require().True(evmBalanceTo.Equal(initialEVMBalanceTo.Add(amountTransferred)), "expected evm balance to increase")
			} else {
				// If we do not expect balance changes to be propagated to the EVM,
				// we ensure that the EVM balances are unchanged.
				s.Require().True(evmBalanceFrom.Equal(initialEVMBalanceFrom), "expected evm balance from to be unchanged")
				s.Require().True(evmBalanceTo.Equal(initialEVMBalanceTo), "expected evm balance to be unchanged")
			}

			// we call  the statedb commit here to simulate end of transaction
			// processing and flush the cache context
			s.Require().NoError(evm.StateDB.(*statedb.StateDB).Commit())

			if tc.postCheck != nil {
				tc.postCheck()
			}
		})
	}
}

func (s *TestSuite) TestEmitTransferEvent() {
	testcases := []struct {
		name   string
		from   common.Address
		to     common.Address
		amount *big.Int
	}{
		{
			name:   "pass",
			from:   s.account1.EvmAddr,
			to:     s.account2.EvmAddr,
			amount: big.NewInt(100),
		},
	}

	for _, tc := range testcases {
		tc := tc
		s.Run(tc.name, func() {
			ae := erc20.NewTransferEvent(tc.from, tc.to, tc.amount)
			args := ae.Arguments()

			s.Require().Len(args, 3)

			// Check the first argument
			s.Require().True(args[0].Indexed)
			s.Require().Equal(tc.from, args[0].Value)

			// Check the second argument
			s.Require().True(args[1].Indexed)
			s.Require().Equal(tc.to, args[1].Value)

			// Check the third argument
			s.Require().False(args[2].Indexed)
			s.Require().Equal(tc.amount, args[2].Value)
		})
	}
}
