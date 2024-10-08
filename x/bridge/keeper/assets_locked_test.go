package keeper

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/mezo-org/mezod/cmd/config"
	"github.com/mezo-org/mezod/x/bridge/types"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
	"github.com/stretchr/testify/mock"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"
)

const (
	recipient1 = "mezo12wsc0qgyfwwfj3wrlpgm9q3lmndl2m4qmm34dp"
	recipient2 = "mezo1xqurxvvh8z2xpj6wltq0tajxm47xnv7q6rtvja"
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
		events       types.AssetsLockedEvents
		errContains  string
		postCheckFn  func(ctx sdk.Context, k Keeper)
	}{
		{
			name:         "empty events",
			bankKeeperFn: func(_ sdk.Context) *mockBankKeeper { return newMockBankKeeper() },
			events:       types.AssetsLockedEvents{},
			errContains:  "empty AssetsLocked sequence",
		},
		{
			name:         "nil events",
			bankKeeperFn: func(_ sdk.Context) *mockBankKeeper { return newMockBankKeeper() },
			events:       nil,
			errContains:  "empty AssetsLocked sequence",
		},
		{
			name:         "events invalid - event with invalid sequence",
			bankKeeperFn: func(_ sdk.Context) *mockBankKeeper { return newMockBankKeeper() },
			events: types.AssetsLockedEvents{
				mockEvent(0, recipient1, 1), // invalid sequence
				mockEvent(1, recipient2, 2), // proper event
			},
			errContains: "invalid AssetsLocked sequence",
		},
		{
			name:         "events invalid - event with invalid recipient",
			bankKeeperFn: func(_ sdk.Context) *mockBankKeeper { return newMockBankKeeper() },
			events: types.AssetsLockedEvents{
				mockEvent(1, "corrupted", 1), // invalid recipient
				mockEvent(2, recipient1, 2),  // proper event
			},
			errContains: "invalid AssetsLocked sequence",
		},
		{
			name:         "events invalid - event with invalid amount",
			bankKeeperFn: func(_ sdk.Context) *mockBankKeeper { return newMockBankKeeper() },
			events: types.AssetsLockedEvents{
				mockEvent(1, recipient1, 0), // invalid amount
				mockEvent(2, recipient2, 1), // proper event
			},
			errContains: "invalid AssetsLocked sequence",
		},
		{
			name:         "events invalid - not strictly increasing by one",
			bankKeeperFn: func(_ sdk.Context) *mockBankKeeper { return newMockBankKeeper() },
			events: types.AssetsLockedEvents{
				mockEvent(2, recipient1, 1),
				mockEvent(1, recipient2, 2),
			},
			errContains: "invalid AssetsLocked sequence",
		},
		{
			name:         "unexpected sequence start - gap",
			bankKeeperFn: func(_ sdk.Context) *mockBankKeeper { return newMockBankKeeper() },
			events: types.AssetsLockedEvents{
				// Sequence tip is 10, so the expected start is 11.
				mockEvent(12, recipient1, 1),
				mockEvent(13, recipient2, 2),
			},
			errContains: "unexpected AssetsLocked sequence start",
		},
		{
			name:         "unexpected sequence start - duplicate",
			bankKeeperFn: func(_ sdk.Context) *mockBankKeeper { return newMockBankKeeper() },
			events: types.AssetsLockedEvents{
				// Sequence tip is 10, so the expected start is 11.
				mockEvent(10, recipient1, 1),
				mockEvent(11, recipient2, 2),
			},
			errContains: "unexpected AssetsLocked sequence start",
		},
		{
			name: "bank keeper minting error",
			bankKeeperFn: func(ctx sdk.Context) *mockBankKeeper {
				bankKeeper := newMockBankKeeper()

				bankKeeper.On(
					"MintCoins",
					ctx,
					types.ModuleName,
					sdk.NewCoins(
						sdk.NewCoin(evmtypes.DefaultEVMDenom, math.NewInt(3)),
					),
				).Return(fmt.Errorf("mint error"))

				return bankKeeper
			},
			events: types.AssetsLockedEvents{
				// Sequence tip is 10, so the expected start is 11.
				mockEvent(11, recipient1, 1),
				mockEvent(12, recipient2, 2),
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
						sdk.NewCoin(evmtypes.DefaultEVMDenom, math.NewInt(3)),
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
			events: types.AssetsLockedEvents{
				// Sequence tip is 10, so the expected start is 11.
				mockEvent(11, recipient1, 1),
				mockEvent(12, recipient2, 2),
			},
			errContains: "failed to send coins",
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
						sdk.NewCoin(evmtypes.DefaultEVMDenom, math.NewInt(3)),
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
			events: types.AssetsLockedEvents{
				// Sequence tip is 10, so the expected start is 11.
				mockEvent(11, recipient1, 1),
				mockEvent(12, recipient2, 2),
			},
			errContains: "",
			postCheckFn: func(ctx sdk.Context, k Keeper) {
				k.bankKeeper.(*mockBankKeeper).AssertExpectations(t)

				require.EqualValues(
					t,
					math.NewInt(12),
					k.GetAssetsLockedSequenceTip(ctx),
				)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx, k := mockContext()

			k.bankKeeper = test.bankKeeperFn(ctx)

			// Set the sequence tip to 10.
			k.setAssetsLockedSequenceTip(ctx, math.NewInt(10))

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
) types.AssetsLockedEvent {
	return types.AssetsLockedEvent{
		Sequence:  math.NewInt(sequence),
		Recipient: recipient,
		Amount:    math.NewInt(amount),
	}
}
