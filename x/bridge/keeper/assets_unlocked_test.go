package keeper

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"testing"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/mezo-org/mezod/cmd/config"
	"github.com/mezo-org/mezod/x/bridge/types"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
	"github.com/stretchr/testify/require"
)

func TestGetAssetsUnlockedSequenceTip(t *testing.T) {
	ctx, k := mockContext()

	newTip := math.NewInt(100)
	k.setAssetsUnlockedSequenceTip(ctx, newTip)

	require.EqualValues(t, newTip, k.GetAssetsUnlockedSequenceTip(ctx))
}

func TestSaveAssetsUnlocked(t *testing.T) {
	// Set bech32 prefixes to make the recipient address validation in
	// AssetsLocked events possible (see AssetsLockedEvent.IsValid).
	cfg := sdk.GetConfig()
	config.SetBech32Prefixes(cfg)

	blockTime := uint32(10000)

	toBytes := func(address string) sdk.AccAddress {
		account, err := sdk.AccAddressFromBech32(address)
		if err != nil {
			t.Fatal(err)
		}
		return account
	}

	//nolint:gosec
	invalidToken := "0x57cc23c7f5ec21f0225187281dd61ef7dfb5c476"

	tests := []struct {
		name         string
		bankKeeperFn func(ctx sdk.Context) *mockBankKeeper
		evmKeeperFn  func(ctx sdk.Context) *mockEvmKeeper
		event        types.AssetsUnlockedEvent
		errContains  string
		run          func(ctx sdk.Context, k Keeper) (*types.AssetsUnlockedEvent, error)
		postCheckFn  func(ctx sdk.Context, k Keeper)
	}{
		{
			name:         "success erc20 mapped token",
			bankKeeperFn: func(_ sdk.Context) *mockBankKeeper { return newMockBankKeeper() },
			evmKeeperFn:  func(_ sdk.Context) *mockEvmKeeper { return newMockEvmKeeper() },
			event: types.AssetsUnlockedEvent{
				UnlockSequence: math.NewInt(11),
				Recipient:      toBytes(recipient1),
				Token:          testSourceERC20Token1,
				Sender:         sender,
				Amount:         math.NewInt(1),
				Chain:          0,
				BlockTime:      blockTime,
			},
			run: func(ctx sdk.Context, k Keeper) (*types.AssetsUnlockedEvent, error) {
				token, _ := hex.DecodeString(testMezoERC20Token1[2:])
				recipient := toBytes(recipient1)
				sender := common.HexToAddress(sender).Bytes()
				return k.SaveAssetsUnlocked(ctx, recipient, token, sender, math.NewInt(1), 0)
			},
		},
		{
			name:         "success btc token",
			bankKeeperFn: func(_ sdk.Context) *mockBankKeeper { return newMockBankKeeper() },
			evmKeeperFn:  func(_ sdk.Context) *mockEvmKeeper { return newMockEvmKeeper() },
			event: types.AssetsUnlockedEvent{
				UnlockSequence: math.NewInt(11),
				Recipient:      toBytes(recipient1),
				Token:          testSourceBTCToken,
				Sender:         sender,
				Amount:         math.NewInt(1),
				Chain:          0,
				BlockTime:      blockTime,
			},
			run: func(ctx sdk.Context, k Keeper) (*types.AssetsUnlockedEvent, error) {
				btcToken := evmtypes.HexAddressToBytes(
					evmtypes.BTCTokenPrecompileAddress,
				)
				recipient := toBytes(recipient1)
				sender := common.HexToAddress(sender).Bytes()
				return k.SaveAssetsUnlocked(ctx, recipient, btcToken, sender, math.NewInt(1), 0)
			},
		},
		{
			name:         "invalid token",
			bankKeeperFn: func(_ sdk.Context) *mockBankKeeper { return newMockBankKeeper() },
			evmKeeperFn:  func(_ sdk.Context) *mockEvmKeeper { return newMockEvmKeeper() },
			errContains:  fmt.Sprintf("unknown token %s", invalidToken[2:]),
			run: func(ctx sdk.Context, k Keeper) (*types.AssetsUnlockedEvent, error) {
				// not a mapped address
				token, _ := hex.DecodeString(invalidToken[2:])
				recipient := toBytes(recipient1)
				sender := common.HexToAddress(sender).Bytes()
				return k.SaveAssetsUnlocked(ctx, recipient, token, sender, math.NewInt(1), 0)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx, k := mockContext()
			ctx = ctx.WithBlockTime(time.Unix(int64(blockTime), 0).UTC())

			k.bankKeeper = test.bankKeeperFn(ctx)
			k.evmKeeper = test.evmKeeperFn(ctx)

			// Set the sequence tip to 10.
			k.setAssetsUnlockedSequenceTip(ctx, math.NewInt(10))

			// Set test mappings.
			k.setERC20TokensMappings(
				ctx, []*types.ERC20TokenMapping{
					{
						SourceToken: testSourceERC20Token1,
						MezoToken:   testMezoERC20Token1,
					},
					{
						SourceToken: testSourceERC20Token2,
						MezoToken:   testMezoERC20Token2,
					},
				},
			)

			// Set the outflow limit to the maximum value of uint256
			maxUint256 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))
			k.SetOutflowLimit(ctx, evmtypes.HexAddressToBytes(testMezoERC20Token1), math.NewIntFromBigInt(maxUint256))
			k.SetOutflowLimit(ctx, evmtypes.HexAddressToBytes(invalidToken), math.NewIntFromBigInt(maxUint256))
			k.SetOutflowLimit(ctx, evmtypes.HexAddressToBytes(evmtypes.BTCTokenPrecompileAddress), math.NewIntFromBigInt(maxUint256))

			evt, err := test.run(ctx, k)

			if len(test.errContains) == 0 {
				require.NoError(t, err, "expected no error")
				require.EqualValues(t, test.event, *evt, "expected same events")
			} else {
				// ErrorContains checks if the error is non-nil so no need
				// for an explicit check here.
				require.ErrorContains(
					t,
					err,
					test.errContains,
					"expected different error message",
				)
			}

			if test.postCheckFn != nil {
				test.postCheckFn(ctx, k)
			}
		})
	}
}

func TestBurnBTC(t *testing.T) {
	// Set bech32 prefixes to make the recipient address validation in
	// AssetsLocked events possible (see AssetsLockedEvent.IsValid).
	cfg := sdk.GetConfig()
	config.SetBech32Prefixes(cfg)

	toBytes := func(address string) sdk.AccAddress {
		account, err := sdk.AccAddressFromBech32(address)
		if err != nil {
			t.Fatal(err)
		}
		return account
	}

	tests := []struct {
		name         string
		bankKeeperFn func(ctx sdk.Context) *mockBankKeeper
		evmKeeperFn  func(ctx sdk.Context) *mockEvmKeeper
		errContains  string
		run          func(ctx sdk.Context, k Keeper) error
		postCheckFn  func(ctx sdk.Context, k Keeper, t *testing.T)
	}{
		{
			name: "burn btc success",
			bankKeeperFn: func(ctx sdk.Context) *mockBankKeeper {
				bankKeeper := newMockBankKeeper()
				recipient := toBytes(recipient1)
				coins := sdk.NewCoins(sdk.NewCoin(evmtypes.DefaultEVMDenom, math.NewInt(1)))
				bankKeeper.On("SendCoinsFromAccountToModule", ctx, recipient, types.ModuleName, coins).Return(nil)
				bankKeeper.On("BurnCoins", ctx, types.ModuleName, coins).Return(nil)

				return bankKeeper
			},
			evmKeeperFn: func(_ sdk.Context) *mockEvmKeeper { return newMockEvmKeeper() },
			run: func(ctx sdk.Context, k Keeper) error {
				recipient := toBytes(recipient1)

				return k.BurnBTC(ctx, recipient, math.NewInt(1))
			},
			postCheckFn: func(ctx sdk.Context, k Keeper, t *testing.T) {
				require.Equal(t, math.NewInt(1), k.GetBTCBurnt(ctx), "invalid burn amount")
			},
		},
		{
			name: "burn btc bank keeper failed to send coins",
			bankKeeperFn: func(ctx sdk.Context) *mockBankKeeper {
				bankKeeper := newMockBankKeeper()
				recipient := toBytes(recipient1)
				coins := sdk.NewCoins(sdk.NewCoin(evmtypes.DefaultEVMDenom, math.NewInt(1)))
				bankKeeper.On("SendCoinsFromAccountToModule", ctx, recipient, types.ModuleName, coins).Return(errors.New("failed to send coins"))

				return bankKeeper
			},
			evmKeeperFn: func(_ sdk.Context) *mockEvmKeeper { return newMockEvmKeeper() },
			run: func(ctx sdk.Context, k Keeper) error {
				recipient := toBytes(recipient1)
				return k.BurnBTC(ctx, recipient, math.NewInt(1))
			},
			postCheckFn: func(ctx sdk.Context, k Keeper, t *testing.T) {
				require.Equal(t, math.NewInt(0), k.GetBTCBurnt(ctx), "invalid burn amount")
			},
			errContains: "failed to send coins",
		},
		{
			name: "burn btc bank keeper failed to burn",
			bankKeeperFn: func(ctx sdk.Context) *mockBankKeeper {
				bankKeeper := newMockBankKeeper()

				recipient := toBytes(recipient1)
				coins := sdk.NewCoins(sdk.NewCoin(evmtypes.DefaultEVMDenom, math.NewInt(1)))
				bankKeeper.On("SendCoinsFromAccountToModule", ctx, recipient, types.ModuleName, coins).Return(nil)
				bankKeeper.On("BurnCoins", ctx, types.ModuleName, coins).Return(errors.New("failed to burn coins"))
				return bankKeeper
			},
			evmKeeperFn: func(_ sdk.Context) *mockEvmKeeper { return newMockEvmKeeper() },
			run: func(ctx sdk.Context, k Keeper) error {
				recipient := toBytes(recipient1)
				return k.BurnBTC(ctx, recipient, math.NewInt(1))
			},
			postCheckFn: func(ctx sdk.Context, k Keeper, t *testing.T) {
				require.Equal(t, math.NewInt(0), k.GetBTCBurnt(ctx), "invalid burn amount")
			},
			errContains: "failed to burn coins",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx, k := mockContext()

			k.bankKeeper = test.bankKeeperFn(ctx)
			k.evmKeeper = test.evmKeeperFn(ctx)

			// Set the sequence tip to 10.
			k.setAssetsUnlockedSequenceTip(ctx, math.NewInt(10))

			// Set test mappings.
			k.setERC20TokensMappings(
				ctx, []*types.ERC20TokenMapping{
					{
						SourceToken: testSourceERC20Token1,
						MezoToken:   testMezoERC20Token1,
					},
					{
						SourceToken: testSourceERC20Token2,
						MezoToken:   testMezoERC20Token2,
					},
				},
			)

			err := test.run(ctx, k)

			if len(test.errContains) == 0 {
				require.NoError(t, err, "expected no error")
			} else {
				// ErrorContains checks if the error is non-nil so no need
				// for an explicit check here.
				require.ErrorContains(
					t,
					err,
					test.errContains,
					"expected different error message",
				)
			}

			if test.postCheckFn != nil {
				test.postCheckFn(ctx, k, t)
			}
		})
	}
}

func TestBurnERC20(t *testing.T) {
	// Set bech32 prefixes to make the recipient address validation in
	// AssetsLocked events possible (see AssetsLockedEvent.IsValid).
	cfg := sdk.GetConfig()
	config.SetBech32Prefixes(cfg)

	toBytes := func(address string) sdk.AccAddress {
		account, err := sdk.AccAddressFromBech32(address)
		if err != nil {
			t.Fatal(err)
		}
		return account
	}

	tests := []struct {
		name         string
		bankKeeperFn func(ctx sdk.Context) *mockBankKeeper
		evmKeeperFn  func(ctx sdk.Context) *mockEvmKeeper
		errContains  string
		run          func(ctx sdk.Context, k Keeper) error
		postCheckFn  func(ctx sdk.Context, k Keeper, t *testing.T)
	}{
		{
			name: "burn erc20 success",
			bankKeeperFn: func(_ sdk.Context) *mockBankKeeper {
				return newMockBankKeeper()
			},
			evmKeeperFn: func(ctx sdk.Context) *mockEvmKeeper {
				evmKeeper := newMockEvmKeeper()

				fromAddr := toBytes(recipient1)
				token, _ := hex.DecodeString(testSourceERC20Token1[2:])
				amount := big.NewInt(1)
				bridgeAddrBytes := evmtypes.HexAddressToBytes(
					evmtypes.AssetsBridgePrecompileAddress,
				)

				call, err := evmtypes.NewERC20BurnFromCall(
					bridgeAddrBytes,
					token,
					fromAddr,
					amount,
				)
				if err != nil {
					panic(fmt.Sprintf("couldn't create burnFrom call: %v", err))
				}

				evmKeeper.On("ExecuteContractCall", ctx, call).Return(nil, nil)

				return evmKeeper
			},
			run: func(ctx sdk.Context, k Keeper) error {
				recipient := toBytes(recipient1)

				token, _ := hex.DecodeString(testSourceERC20Token1[2:])
				return k.BurnERC20(ctx, token, recipient, big.NewInt(1))
			},
		},
		{
			name: "burn erc20 failure",
			bankKeeperFn: func(_ sdk.Context) *mockBankKeeper {
				return newMockBankKeeper()
			},
			evmKeeperFn: func(ctx sdk.Context) *mockEvmKeeper {
				evmKeeper := newMockEvmKeeper()

				fromAddr := toBytes(recipient1)
				token, _ := hex.DecodeString(testSourceERC20Token1[2:])
				amount := big.NewInt(1)
				bridgeAddrBytes := evmtypes.HexAddressToBytes(
					evmtypes.AssetsBridgePrecompileAddress,
				)

				call, err := evmtypes.NewERC20BurnFromCall(
					bridgeAddrBytes,
					token,
					fromAddr,
					amount,
				)
				if err != nil {
					panic(fmt.Sprintf("couldn't create burnFrom call: %v", err))
				}

				evmKeeper.On("ExecuteContractCall", ctx, call).Return(nil, errors.New("execution reverted"))

				return evmKeeper
			},
			run: func(ctx sdk.Context, k Keeper) error {
				recipient := toBytes(recipient1)

				token, _ := hex.DecodeString(testSourceERC20Token1[2:])
				return k.BurnERC20(ctx, token, recipient, big.NewInt(1))
			},
			errContains: "execution reverted",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx, k := mockContext()

			k.bankKeeper = test.bankKeeperFn(ctx)
			k.evmKeeper = test.evmKeeperFn(ctx)

			// Set the sequence tip to 10.
			k.setAssetsUnlockedSequenceTip(ctx, math.NewInt(10))

			// Set test mappings.
			k.setERC20TokensMappings(
				ctx, []*types.ERC20TokenMapping{
					{
						SourceToken: testSourceERC20Token1,
						MezoToken:   testMezoERC20Token1,
					},
					{
						SourceToken: testSourceERC20Token2,
						MezoToken:   testMezoERC20Token2,
					},
				},
			)

			err := test.run(ctx, k)

			if len(test.errContains) == 0 {
				require.NoError(t, err, "expected no error")
			} else {
				// ErrorContains checks if the error is non-nil so no need
				// for an explicit check here.
				require.ErrorContains(
					t,
					err,
					test.errContains,
					"expected different error message",
				)
			}

			if test.postCheckFn != nil {
				test.postCheckFn(ctx, k, t)
			}
		})
	}
}

func TestSaveAssetsUnlockedWithOutflowLimits(t *testing.T) {
	cfg := sdk.GetConfig()
	config.SetBech32Prefixes(cfg)

	btcToken := evmtypes.HexAddressToBytes(evmtypes.BTCTokenPrecompileAddress)
	erc20Token := common.HexToAddress("0x1234567890123456789012345678901234567890").Bytes()

	t.Run("SaveAssetsUnlocked respects outflow limits for BTC", func(t *testing.T) {
		ctx, keeper := mockContext()
		// Set outflow limit for BTC token
		keeper.SetOutflowLimit(ctx, btcToken, math.NewInt(1000))

		// First let's verify the outflow limit was set correctly
		limit := keeper.GetOutflowLimit(ctx, btcToken)
		require.Equal(t, math.NewInt(1000), limit, "limit should be set correctly")

		// Verify capacity
		capacity, _ := keeper.GetOutflowCapacity(ctx, btcToken)
		require.Equal(t, math.NewInt(1000), capacity, "capacity should equal limit initially")

		// Test direct outflow limit check
		err := keeper.checkOutflowLimit(ctx, btcToken, math.NewInt(500))
		require.NoError(t, err, "outflow limit check should pass for amount within limit")

		// First transaction within limit should succeed
		event, err := keeper.SaveAssetsUnlocked(
			ctx,
			[]byte("recipient"),      // recipient
			btcToken,                 // token
			[]byte("sender_address"), // sender
			math.NewInt(500),         // amount
			1,                        // chain
		)
		require.NoError(t, err, "transaction within limit should succeed")
		require.NotNil(t, event)

		// Verify outflow was tracked
		currentOutflow := keeper.getCurrentOutflow(ctx, btcToken)
		require.Equal(t, math.NewInt(500), currentOutflow, "outflow should be tracked")

		// Second transaction within remaining capacity should succeed
		_, err = keeper.SaveAssetsUnlocked(
			ctx,
			[]byte("recipient2"),     // recipient
			btcToken,                 // token
			[]byte("sender_address"), // sender
			math.NewInt(400),         // amount
			1,
		)
		require.NoError(t, err, "transaction within remaining capacity should succeed")

		// Verify total outflow
		currentOutflow = keeper.getCurrentOutflow(ctx, btcToken)
		require.Equal(t, math.NewInt(900), currentOutflow, "total outflow should be tracked")

		// Third transaction exceeding limit should fail
		_, err = keeper.SaveAssetsUnlocked(
			ctx,
			[]byte("recipient3"),     // recipient
			btcToken,                 // token
			[]byte("sender_address"), // sender
			math.NewInt(200),         // amount
			1,
		)
		require.Error(t, err, "transaction exceeding limit should fail")
		require.ErrorContains(t, err, "outflow limit check error", "should be outflow limit error")
		require.ErrorIs(t, err, types.ErrOutflowLimitExceeded)

		// Verify outflow was not increased for failed transaction
		currentOutflow = keeper.getCurrentOutflow(ctx, btcToken)
		require.Equal(t, math.NewInt(900), currentOutflow, "outflow should not increase for failed transaction")
	})

	t.Run("SaveAssetsUnlocked respects outflow limits for ERC20", func(t *testing.T) {
		ctx, keeper := mockContext()

		// Create ERC20 token mapping first
		sourceToken := common.HexToAddress("0xA0b86a33E6441b5B6F7BB33b8F2D8F9FD6D5F0C2").Bytes()
		mapping := types.NewERC20TokenMapping(sourceToken, erc20Token)
		keeper.setERC20TokenMapping(ctx, mapping)

		// Set outflow limit for ERC20 token
		keeper.SetOutflowLimit(ctx, erc20Token, math.NewInt(2000))

		// Transaction within limit should succeed
		event, err := keeper.SaveAssetsUnlocked(
			ctx,
			[]byte("recipient"),      // recipient
			erc20Token,               // token
			[]byte("sender_address"), // sender
			math.NewInt(1500),        // amount
			2,
		)
		require.NoError(t, err, "transaction within limit should succeed")
		require.NotNil(t, event)

		// Verify outflow was tracked
		currentOutflow := keeper.getCurrentOutflow(ctx, erc20Token)
		require.Equal(t, math.NewInt(1500), currentOutflow, "outflow should be tracked")

		// Transaction exceeding remaining capacity should fail
		_, err = keeper.SaveAssetsUnlocked(
			ctx,
			[]byte("recipient2"),     // recipient
			erc20Token,               // token
			[]byte("sender_address"), // sender
			math.NewInt(600),         // amount
			2,
		)
		require.Error(t, err, "transaction exceeding limit should fail")
		require.ErrorContains(t, err, "outflow limit check error")
	})

	t.Run("SaveAssetsUnlocked with zero outflow limit", func(t *testing.T) {
		ctx, keeper := mockContext()
		zeroLimitToken := common.HexToAddress("0x9999999999999999999999999999999999999999").Bytes()

		// Create ERC20 token mapping
		sourceToken := common.HexToAddress("0xB1b86a33E6441b5B6F7BB33b8F2D8F9FD6D5F0C2").Bytes()
		mapping := types.NewERC20TokenMapping(sourceToken, zeroLimitToken)
		keeper.setERC20TokenMapping(ctx, mapping)

		// Zero limit means no outflow allowed
		keeper.SetOutflowLimit(ctx, zeroLimitToken, math.ZeroInt())

		// Any transaction should fail
		_, err := keeper.SaveAssetsUnlocked(
			ctx,
			[]byte("recipient"),      // recipient
			zeroLimitToken,           // token
			[]byte("sender_address"), // sender
			math.NewInt(1),           // amount
			2,
		)
		require.Error(t, err, "transaction with zero limit should fail")
		require.ErrorContains(t, err, "outflow limit check error")
	})

	t.Run("SaveAssetsUnlocked with no outflow limit set", func(t *testing.T) {
		ctx, keeper := mockContext()
		noLimitToken := common.HexToAddress("0x8888888888888888888888888888888888888888").Bytes()

		// No limit set means zero limit by default
		_, err := keeper.SaveAssetsUnlocked(
			ctx,
			[]byte("recipient"),      // recipient
			noLimitToken,             // token
			[]byte("sender_address"), // sender
			math.NewInt(1),           // amount
			2,
		)
		require.Error(t, err, "transaction with no limit set should fail")
		require.ErrorContains(t, err, "outflow limit check error")
	})

	t.Run("SaveAssetsUnlocked exactly at limit", func(t *testing.T) {
		ctx, keeper := mockContext()
		exactLimitToken := common.HexToAddress("0x7777777777777777777777777777777777777777").Bytes()

		// Create ERC20 token mapping
		sourceToken := common.HexToAddress("0xC2b86a33E6441b5B6F7BB33b8F2D8F9FD6D5F0C2").Bytes()
		mapping := types.NewERC20TokenMapping(sourceToken, exactLimitToken)
		keeper.setERC20TokenMapping(ctx, mapping)

		// Set limit
		keeper.SetOutflowLimit(ctx, exactLimitToken, math.NewInt(100))

		// Transaction exactly at limit should succeed
		event, err := keeper.SaveAssetsUnlocked(
			ctx,
			[]byte("recipient"),      // recipient
			exactLimitToken,          // token
			[]byte("sender_address"), // sender
			math.NewInt(100),         // amount
			2,
		)
		require.NoError(t, err, "transaction exactly at limit should succeed")
		require.NotNil(t, event)

		// Verify outflow is at limit
		currentOutflow := keeper.getCurrentOutflow(ctx, exactLimitToken)
		require.Equal(t, math.NewInt(100), currentOutflow)

		// Any additional transaction should fail
		_, err = keeper.SaveAssetsUnlocked(
			ctx,
			[]byte("recipient2"),     // recipient
			exactLimitToken,          // token
			[]byte("sender_address"), // sender
			math.NewInt(1),           // amount
			2,
		)
		require.Error(t, err, "any additional transaction should fail")
	})

	t.Run("SaveAssetsUnlocked with large limit", func(t *testing.T) {
		ctx, keeper := mockContext()
		largeLimitToken := common.HexToAddress("0x6666666666666666666666666666666666666666").Bytes()

		// Create ERC20 token mapping
		sourceToken := common.HexToAddress("0xD3b86a33E6441b5B6F7BB33b8F2D8F9FD6D5F0C2").Bytes()
		mapping := types.NewERC20TokenMapping(sourceToken, largeLimitToken)
		keeper.setERC20TokenMapping(ctx, mapping)

		// Set very large limit
		largeLimit := math.NewIntFromBigInt(big.NewInt(0).Exp(big.NewInt(10), big.NewInt(19), nil)) // 10^19
		keeper.SetOutflowLimit(ctx, largeLimitToken, largeLimit)

		// Multiple transactions should succeed
		for i := 0; i < 5; i++ {
			_, err := keeper.SaveAssetsUnlocked(
				ctx,
				[]byte(fmt.Sprintf("recipient%d", i)), // recipient
				largeLimitToken,                       // token
				[]byte("sender_address"),              // sender
				math.NewInt(1000000),                  // amount
				2,
			)
			require.NoError(t, err, "transaction with large limit should succeed")
		}

		// Verify cumulative outflow
		currentOutflow := keeper.getCurrentOutflow(ctx, largeLimitToken)
		require.Equal(t, math.NewInt(5000000), currentOutflow)
	})
}

func TestGetMinBridgeOutAmount(t *testing.T) {
	ctx, k := mockContext()

	token1, err := hex.DecodeString(testMezoERC20Token1[2:])
	if err != nil {
		t.Fatal(err)
	}

	minAmount := math.NewInt(20000)
	err = k.SetMinBridgeOutAmount(ctx, token1, minAmount)
	if err != nil {
		t.Fatal(err)
	}

	actualMinAmount1 := k.GetMinBridgeOutAmount(ctx, token1)
	require.EqualValues(t, minAmount, actualMinAmount1)

	// Check for unknown token
	token2, err := hex.DecodeString(testMezoERC20Token2[2:])
	if err != nil {
		t.Fatal(err)
	}

	actualMinAmount2 := k.GetMinBridgeOutAmount(ctx, token2)
	require.EqualValues(t, math.ZeroInt(), actualMinAmount2)
}

func TestGetMinBridgeOutAmountForBitcoinChain(t *testing.T) {
	ctx, k := mockContext()

	// Test when minimum amount is not set
	actualMinAmount := k.GetMinBridgeOutAmountForBitcoinChain(ctx)
	require.EqualValues(t, math.ZeroInt(), actualMinAmount)

	// Test setting and getting minimum amount
	minAmount := math.NewInt(50000)
	k.SetMinBridgeOutAmountForBitcoinChain(ctx, minAmount)

	actualMinAmount = k.GetMinBridgeOutAmountForBitcoinChain(ctx)
	require.EqualValues(t, minAmount, actualMinAmount)

	// Test overwriting with zero amount
	k.SetMinBridgeOutAmountForBitcoinChain(ctx, math.ZeroInt())

	actualMinAmount = k.GetMinBridgeOutAmountForBitcoinChain(ctx)
	require.EqualValues(t, math.ZeroInt(), actualMinAmount)

	// Test overwriting with another value
	newMinAmount := math.NewInt(75000)
	k.SetMinBridgeOutAmountForBitcoinChain(ctx, newMinAmount)

	actualMinAmount = k.GetMinBridgeOutAmountForBitcoinChain(ctx)
	require.EqualValues(t, newMinAmount, actualMinAmount)
}

func TestSetMinBridgeOutAmountForBitcoinChain(t *testing.T) {
	ctx, k := mockContext()

	testCases := []struct {
		name      string
		amount    math.Int
		expectErr bool
	}{
		{
			name:      "set zero amount",
			amount:    math.ZeroInt(),
			expectErr: false,
		},
		{
			name:      "set positive amount",
			amount:    math.NewInt(100000),
			expectErr: false,
		},
		{
			name:      "set large amount",
			amount:    math.NewIntFromBigInt(big.NewInt(0).Exp(big.NewInt(10), big.NewInt(18), nil)),
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			k.SetMinBridgeOutAmountForBitcoinChain(ctx, tc.amount)
			actualAmount := k.GetMinBridgeOutAmountForBitcoinChain(ctx)
			require.EqualValues(t, tc.amount, actualAmount)
		})
	}
}
