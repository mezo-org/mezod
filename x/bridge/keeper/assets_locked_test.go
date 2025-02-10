package keeper

import (
	"fmt"
	"math/big"
	"testing"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/mezo-org/mezod/cmd/config"
	"github.com/mezo-org/mezod/x/bridge/types"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
	"github.com/stretchr/testify/mock"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"
)

const (
	//nolint:gosec
	recipient1 = "mezo12wsc0qgyfwwfj3wrlpgm9q3lmndl2m4qmm34dp"
	//nolint:gosec
	recipient2 = "mezo1xqurxvvh8z2xpj6wltq0tajxm47xnv7q6rtvja"
	//nolint:gosec
	testSourceERC20Token1 = "0xac7f043Cf1BF10143926CC0035dBc46999512732"
	//nolint:gosec
	testMezoERC20Token1 = "0x546758f4C2EfA4f37d66fF53644170F1d27AA1A0"
	//nolint:gosec
	testSourceERC20Token2 = "0x8c1442D1a806dd27F3562b54d964778BA4a18708"
	//nolint:gosec
	testMezoERC20Token2 = "0x891347CfC5ED716b4C178dCD28e66fC9f3802B09"
)

func TestGetAssetsLockedSequenceTip(t *testing.T) {
	ctx, k := mockContext()

	newTip := math.NewInt(100)
	k.setAssetsLockedSequenceTip(ctx, newTip)

	require.EqualValues(t, newTip, k.GetAssetsLockedSequenceTip(ctx))
}

func TestAcceptAssetsLocked(t *testing.T) {
	// Set bech32 prefixes to make the recipient address validation in
	// AssetsLocked events possible (see AssetsLockedEvent.IsValid).
	cfg := sdk.GetConfig()
	config.SetBech32Prefixes(cfg)

	toSdkAccount := func(address string) sdk.AccAddress {
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
		events       types.AssetsLockedEvents
		errContains  string
		postCheckFn  func(ctx sdk.Context, k Keeper)
	}{
		{
			name:         "empty events",
			bankKeeperFn: func(_ sdk.Context) *mockBankKeeper { return newMockBankKeeper() },
			evmKeeperFn:  func(_ sdk.Context) *mockEvmKeeper { return newMockEvmKeeper() },
			events:       types.AssetsLockedEvents{},
			errContains:  "empty AssetsLocked sequence",
		},
		{
			name:         "nil events",
			bankKeeperFn: func(_ sdk.Context) *mockBankKeeper { return newMockBankKeeper() },
			evmKeeperFn:  func(_ sdk.Context) *mockEvmKeeper { return newMockEvmKeeper() },
			events:       nil,
			errContains:  "empty AssetsLocked sequence",
		},
		{
			name:         "events invalid - event with invalid sequence",
			bankKeeperFn: func(_ sdk.Context) *mockBankKeeper { return newMockBankKeeper() },
			evmKeeperFn:  func(_ sdk.Context) *mockEvmKeeper { return newMockEvmKeeper() },
			events: types.AssetsLockedEvents{
				mockEvent(0, recipient1, 1, testSourceBTCToken), // invalid sequence
				mockEvent(1, recipient2, 2, testSourceBTCToken), // proper event
			},
			errContains: "invalid AssetsLocked sequence",
		},
		{
			name:         "events invalid - event with invalid recipient",
			bankKeeperFn: func(_ sdk.Context) *mockBankKeeper { return newMockBankKeeper() },
			evmKeeperFn:  func(_ sdk.Context) *mockEvmKeeper { return newMockEvmKeeper() },
			events: types.AssetsLockedEvents{
				mockEvent(1, "corrupted", 1, testSourceBTCToken), // invalid recipient
				mockEvent(2, recipient1, 2, testSourceBTCToken),  // proper event
			},
			errContains: "invalid AssetsLocked sequence",
		},
		{
			name:         "events invalid - event with invalid amount",
			bankKeeperFn: func(_ sdk.Context) *mockBankKeeper { return newMockBankKeeper() },
			evmKeeperFn:  func(_ sdk.Context) *mockEvmKeeper { return newMockEvmKeeper() },
			events: types.AssetsLockedEvents{
				mockEvent(1, recipient1, 0, testSourceBTCToken), // invalid amount
				mockEvent(2, recipient2, 1, testSourceBTCToken), // proper event
			},
			errContains: "invalid AssetsLocked sequence",
		},
		{
			name:         "events invalid - not strictly increasing by one",
			bankKeeperFn: func(_ sdk.Context) *mockBankKeeper { return newMockBankKeeper() },
			evmKeeperFn:  func(_ sdk.Context) *mockEvmKeeper { return newMockEvmKeeper() },
			events: types.AssetsLockedEvents{
				mockEvent(2, recipient1, 1, testSourceBTCToken),
				mockEvent(1, recipient2, 2, testSourceBTCToken),
			},
			errContains: "invalid AssetsLocked sequence",
		},
		{
			name:         "unexpected sequence start - gap",
			bankKeeperFn: func(_ sdk.Context) *mockBankKeeper { return newMockBankKeeper() },
			evmKeeperFn:  func(_ sdk.Context) *mockEvmKeeper { return newMockEvmKeeper() },
			events: types.AssetsLockedEvents{
				// Sequence tip is 10, so the expected start is 11.
				mockEvent(12, recipient1, 1, testSourceBTCToken),
				mockEvent(13, recipient2, 2, testSourceBTCToken),
			},
			errContains: "unexpected AssetsLocked sequence start",
		},
		{
			name:         "unexpected sequence start - duplicate",
			bankKeeperFn: func(_ sdk.Context) *mockBankKeeper { return newMockBankKeeper() },
			evmKeeperFn:  func(_ sdk.Context) *mockEvmKeeper { return newMockEvmKeeper() },
			events: types.AssetsLockedEvents{
				// Sequence tip is 10, so the expected start is 11.
				mockEvent(10, recipient1, 1, testSourceBTCToken),
				mockEvent(11, recipient2, 2, testSourceBTCToken),
			},
			errContains: "unexpected AssetsLocked sequence start",
		},
		{
			name: "bank keeper minting error",
			bankKeeperFn: func(ctx sdk.Context) *mockBankKeeper {
				bankKeeper := newMockBankKeeper()

				// Fail on the first mint.
				bankKeeper.On(
					"MintCoins",
					ctx,
					types.ModuleName,
					sdk.NewCoins(
						sdk.NewCoin(evmtypes.DefaultEVMDenom, math.NewInt(1)),
					),
				).Return(fmt.Errorf("mint error"))

				return bankKeeper
			},
			evmKeeperFn: func(_ sdk.Context) *mockEvmKeeper { return newMockEvmKeeper() },
			events: types.AssetsLockedEvents{
				// Sequence tip is 10, so the expected start is 11.
				mockEvent(11, recipient1, 1, testSourceBTCToken),
				mockEvent(12, recipient2, 2, testSourceBTCToken),
			},
			errContains: "failed to mint coins",
		},
		{
			name: "bank keeper transfer error",
			bankKeeperFn: func(ctx sdk.Context) *mockBankKeeper {
				bankKeeper := newMockBankKeeper()

				bankKeeper.On(
					"MintCoins",
					ctx,
					types.ModuleName,
					sdk.NewCoins(
						sdk.NewCoin(evmtypes.DefaultEVMDenom, math.NewInt(1)),
					),
				).Return(nil)

				bankKeeper.On(
					"MintCoins",
					ctx,
					types.ModuleName,
					sdk.NewCoins(
						sdk.NewCoin(evmtypes.DefaultEVMDenom, math.NewInt(2)),
					),
				).Return(nil)

				bankKeeper.On(
					"SendCoinsFromModuleToAccount",
					ctx,
					types.ModuleName,
					mock.Anything,
					mock.Anything,
				).Return(fmt.Errorf("transfer error"))

				return bankKeeper
			},
			evmKeeperFn: func(_ sdk.Context) *mockEvmKeeper { return newMockEvmKeeper() },
			events: types.AssetsLockedEvents{
				// Sequence tip is 10, so the expected start is 11.
				mockEvent(11, recipient1, 1, testSourceBTCToken),
				mockEvent(12, recipient2, 2, testSourceBTCToken),
			},
			errContains: "failed to send coins",
		},
		{
			name:         "ERC20 mapping not found",
			bankKeeperFn: func(_ sdk.Context) *mockBankKeeper { return newMockBankKeeper() },
			evmKeeperFn: func(ctx sdk.Context) *mockEvmKeeper {
				evmKeeper := newMockEvmKeeper()

				evmKeeper.On(
					"ExecuteContractCall",
					ctx,
					mock.Anything,
				).Return(&evmtypes.MsgEthereumTxResponse{}, nil)

				return evmKeeper
			},
			events: types.AssetsLockedEvents{
				// Sequence tip is 10, so the expected start is 11.
				// Use an arbitrary token whose mapping is not set.
				mockEvent(11, recipient1, 1, "0x96c1493e5a9efa5C7572D98b5e9544B17bD9AE72"),
			},
			errContains: "", // This case shouldn't lead to an error.
			postCheckFn: func(ctx sdk.Context, k Keeper) {
				// The EVM state change should not have been executed.
				k.evmKeeper.(*mockEvmKeeper).AssertNotCalled(
					t,
					"ExecuteContractCall",
					mock.Anything,
					mock.Anything,
				)

				// The sequence tip should have been updated as the event
				// was skipped.
				require.EqualValues(
					t,
					math.NewInt(11),
					k.GetAssetsLockedSequenceTip(ctx),
				)
			},
		},
		{
			name:         "ERC20 mint error",
			bankKeeperFn: func(_ sdk.Context) *mockBankKeeper { return newMockBankKeeper() },
			evmKeeperFn: func(ctx sdk.Context) *mockEvmKeeper {
				evmKeeper := newMockEvmKeeper()

				recipient1SdkAddr, err := sdk.AccAddressFromBech32(recipient1)
				require.NoError(t, err)

				evmCall, err := evmtypes.NewERC20MintCall(
					authtypes.NewModuleAddress(types.ModuleName).Bytes(),
					evmtypes.HexAddressToBytes(testMezoERC20Token1),
					recipient1SdkAddr,
					big.NewInt(1),
				)
				require.NoError(t, err)

				evmKeeper.On(
					"ExecuteContractCall",
					ctx,
					evmCall,
				).Return(nil, fmt.Errorf("ERC20 mint error"))

				return evmKeeper
			},
			events: types.AssetsLockedEvents{
				// Sequence tip is 10, so the expected start is 11.
				mockEvent(11, recipient1, 1, testSourceERC20Token1),
			},
			errContains: "", // This case shouldn't lead to an error.
			postCheckFn: func(ctx sdk.Context, k Keeper) {
				k.evmKeeper.(*mockEvmKeeper).AssertExpectations(t)

				// The sequence tip should have been updated as the event
				// was skipped.
				require.EqualValues(
					t,
					math.NewInt(11),
					k.GetAssetsLockedSequenceTip(ctx),
				)
			},
		},
		{
			name: "events accepted successfully",
			bankKeeperFn: func(ctx sdk.Context) *mockBankKeeper {
				bankKeeper := newMockBankKeeper()

				bankKeeper.On(
					"MintCoins",
					ctx,
					types.ModuleName,
					sdk.NewCoins(
						sdk.NewCoin(evmtypes.DefaultEVMDenom, math.NewInt(1)),
					),
				).Return(nil)

				bankKeeper.On(
					"MintCoins",
					ctx,
					types.ModuleName,
					sdk.NewCoins(
						sdk.NewCoin(evmtypes.DefaultEVMDenom, math.NewInt(2)),
					),
				).Return(nil)

				bankKeeper.On(
					"SendCoinsFromModuleToAccount",
					ctx,
					types.ModuleName,
					toSdkAccount(recipient1),
					sdk.NewCoins(
						sdk.NewCoin(evmtypes.DefaultEVMDenom, math.NewInt(1)),
					),
				).Return(nil)

				bankKeeper.On(
					"SendCoinsFromModuleToAccount",
					ctx,
					types.ModuleName,
					toSdkAccount(recipient2),
					sdk.NewCoins(
						sdk.NewCoin(evmtypes.DefaultEVMDenom, math.NewInt(2)),
					),
				).Return(nil)

				return bankKeeper
			},
			evmKeeperFn: func(ctx sdk.Context) *mockEvmKeeper {
				evmKeeper := newMockEvmKeeper()

				recipient1SdkAddr, err := sdk.AccAddressFromBech32(recipient1)
				require.NoError(t, err)

				evmCall1, err := evmtypes.NewERC20MintCall(
					authtypes.NewModuleAddress(types.ModuleName).Bytes(),
					evmtypes.HexAddressToBytes(testMezoERC20Token1),
					recipient1SdkAddr,
					big.NewInt(3),
				)
				require.NoError(t, err)

				evmKeeper.On(
					"ExecuteContractCall",
					ctx,
					evmCall1,
				).Return(&evmtypes.MsgEthereumTxResponse{}, nil)

				recipient2SdkAddr, err := sdk.AccAddressFromBech32(recipient2)
				require.NoError(t, err)

				evmCall2, err := evmtypes.NewERC20MintCall(
					authtypes.NewModuleAddress(types.ModuleName).Bytes(),
					evmtypes.HexAddressToBytes(testMezoERC20Token2),
					recipient2SdkAddr,
					big.NewInt(4),
				)
				require.NoError(t, err)

				evmKeeper.On(
					"ExecuteContractCall",
					ctx,
					evmCall2,
				).Return(&evmtypes.MsgEthereumTxResponse{}, nil)

				return evmKeeper
			},
			events: types.AssetsLockedEvents{
				// Sequence tip is 10, so the expected start is 11.
				mockEvent(11, recipient1, 1, testSourceBTCToken),
				mockEvent(12, recipient2, 2, testSourceBTCToken),
				mockEvent(13, recipient1, 3, testSourceERC20Token1),
				mockEvent(14, recipient2, 4, testSourceERC20Token2),
			},
			errContains: "",
			postCheckFn: func(ctx sdk.Context, k Keeper) {
				k.bankKeeper.(*mockBankKeeper).AssertExpectations(t)
				k.evmKeeper.(*mockEvmKeeper).AssertExpectations(t)
				require.EqualValues(
					t,
					math.NewInt(14),
					k.GetAssetsLockedSequenceTip(ctx),
				)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx, k := mockContext()

			k.bankKeeper = test.bankKeeperFn(ctx)
			k.evmKeeper = test.evmKeeperFn(ctx)

			// Set the sequence tip to 10.
			k.setAssetsLockedSequenceTip(ctx, math.NewInt(10))

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

			err := k.AcceptAssetsLocked(ctx, test.events)

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
				test.postCheckFn(ctx, k)
			}
		})
	}
}

func mockEvent(
	sequence int64,
	recipient string,
	amount int64,
	token string,
) types.AssetsLockedEvent {
	return types.AssetsLockedEvent{
		Sequence:  math.NewInt(sequence),
		Recipient: recipient,
		Amount:    math.NewInt(amount),
		Token:     token,
	}
}
