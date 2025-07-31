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
				Sender:         toBytes(recipient1),
				Amount:         math.NewInt(1),
				Chain:          0,
				BlockTime:      blockTime,
			},
			run: func(ctx sdk.Context, k Keeper) (*types.AssetsUnlockedEvent, error) {
				token, _ := hex.DecodeString(testMezoERC20Token1[2:])
				recipient := toBytes(recipient1)
				return k.SaveAssetsUnlocked(ctx, recipient, token, recipient, math.NewInt(1), 0)
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
				Sender:         toBytes(recipient1),
				Amount:         math.NewInt(1),
				Chain:          0,
				BlockTime:      blockTime,
			},
			run: func(ctx sdk.Context, k Keeper) (*types.AssetsUnlockedEvent, error) {
				btcToken := evmtypes.HexAddressToBytes(
					evmtypes.BTCTokenPrecompileAddress,
				)
				recipient := toBytes(recipient1)
				return k.SaveAssetsUnlocked(ctx, recipient, btcToken, recipient, math.NewInt(1), 0)
			},
		},
		{
			name:         "invalid token",
			bankKeeperFn: func(_ sdk.Context) *mockBankKeeper { return newMockBankKeeper() },
			evmKeeperFn:  func(_ sdk.Context) *mockEvmKeeper { return newMockEvmKeeper() },
			errContains:  "unknown token 57cc23c7f5ec21f0225187281dd61ef7dfb5c476",
			run: func(ctx sdk.Context, k Keeper) (*types.AssetsUnlockedEvent, error) {
				// not a mapped address
				token, _ := hex.DecodeString("57CC23C7f5Ec21f0225187281dD61Ef7dFb5C476")
				recipient := toBytes(recipient1)
				return k.SaveAssetsUnlocked(ctx, recipient, token, recipient, math.NewInt(1), 0)
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
