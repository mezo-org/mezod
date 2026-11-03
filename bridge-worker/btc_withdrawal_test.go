package bridgeworker

import (
	"context"
	"encoding/hex"
	"errors"
	"math/big"
	"testing"

	sdkmath "cosmossdk.io/math"

	"github.com/mezo-org/mezod/ethereum/bindings/tbtc"
	gomock "go.uber.org/mock/gomock"

	"cosmossdk.io/log"

	"github.com/mezo-org/mezod/bridge-worker/bitcoin"
	"github.com/mezo-org/mezod/ethereum/bindings/portal"
	bridgetypes "github.com/mezo-org/mezod/x/bridge/types"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestUpdateLiveWallets(t *testing.T) {
	t.Helper()
	ctrl := gomock.NewController(t)

	mockTbtcBridgeContract := NewMockTbtcBridgeContract(ctrl)
	mockEthereumChain := NewMockEthereumChain(ctrl)

	env := &environment{
		logger:             log.NewNopLogger(),
		chain:              mockEthereumChain,
		tbtcBridgeContract: mockTbtcBridgeContract,
	}
	bwj := btcWithdrawalJob{
		env:                           env,
		liveWallets:                   [][20]byte{{0x11}, {0x22}}, // Two wallets already stored.
		liveWalletsLastProcessedBlock: 100,
	}

	// Already stored wallets checks.
	mockTbtcBridgeContract.EXPECT().Wallets(
		[20]byte{0x11},
	).Return(
		tbtc.Wallet{State: 1}, nil, // Live
	).Times(1)

	mockTbtcBridgeContract.EXPECT().Wallets(
		[20]byte{0x22},
	).Return(
		tbtc.Wallet{State: 4}, nil, // Closed
	).Times(1)

	// Searching for new wallets.
	mockEthereumChain.EXPECT().FinalizedBlock(gomock.Any()).Return(
		big.NewInt(105), nil,
	).Times(1)

	mockTbtcBridgeContract.EXPECT().PastNewWalletRegisteredEvents(
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
	).Return(
		[]*tbtc.BridgeNewWalletRegistered{
			// Three new wallets.
			{
				WalletPubKeyHash: [20]byte{0x33},
			},
			{
				WalletPubKeyHash: [20]byte{0x44},
			},
			{
				WalletPubKeyHash: [20]byte{0x55},
			},
		}, nil,
	).Times(1)

	// New wallets checks.
	mockTbtcBridgeContract.EXPECT().Wallets(
		[20]byte{0x33},
	).Return(
		tbtc.Wallet{State: 1}, nil, // Live
	).Times(1)

	mockTbtcBridgeContract.EXPECT().Wallets(
		[20]byte{0x44},
	).Return(
		tbtc.Wallet{State: 4}, nil, // Closed
	).Times(1)

	mockTbtcBridgeContract.EXPECT().Wallets(
		[20]byte{0x55},
	).Return(
		tbtc.Wallet{State: 1}, nil, // Live
	).Times(1)

	err := bwj.updateLiveWallets(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	expectedLiveWalletsLastProcessedBlock := uint64(105)
	require.Equal(
		t,
		expectedLiveWalletsLastProcessedBlock,
		bwj.liveWalletsLastProcessedBlock,
	)

	expectedLiveWallets := [][20]byte{{0x11}, {0x33}, {0x55}}
	require.Equal(
		t,
		expectedLiveWallets,
		bwj.liveWallets,
	)
}

func TestEnqueueBTCWithdrawal(t *testing.T) {
	env := &environment{
		logger: log.NewNopLogger(),
	}
	bwj := btcWithdrawalJob{
		env: env,
	}

	event1 := portal.MezoBridgeAssetsUnlockConfirmed{
		UnlockSequenceNumber: big.NewInt(1),
		Recipient:            common.HexToHash("0x0A219c03938FBC93aA23cAd65f7c480f52665C2a"),
		Token:                common.HexToAddress("0x3A128b915bee3645396d43Fe7A13A59a66C427d6"),
		Amount:               big.NewInt(1000000),
	}

	event2 := portal.MezoBridgeAssetsUnlockConfirmed{
		UnlockSequenceNumber: big.NewInt(2),
		Recipient:            common.HexToHash("0x0A219c03938FBC93aA23cAd65f7c480f52665C2a"),
		Token:                common.HexToAddress("0x3A128b915bee3645396d43Fe7A13A59a66C427d6"),
		Amount:               big.NewInt(2000000),
	}

	event3 := portal.MezoBridgeAssetsUnlockConfirmed{
		UnlockSequenceNumber: big.NewInt(3),
		Recipient:            common.HexToHash("0x0A219c03938FBC93aA23cAd65f7c480f52665C2a"),
		Token:                common.HexToAddress("0x3A128b915bee3645396d43Fe7A13A59a66C427d6"),
		Amount:               big.NewInt(3000000),
	}

	bwj.enqueueBTCWithdrawal(
		&event1,
	)
	bwj.enqueueBTCWithdrawal(
		&event2,
	)
	bwj.enqueueBTCWithdrawal(
		&event3,
	)

	expectedWithdrawalQueue := []portal.MezoBridgeAssetsUnlockConfirmed{
		event1, event2, event3,
	}

	require.Equal(t, expectedWithdrawalQueue, bwj.btcWithdrawalQueue)
}

func TestDequeueBTCWithdrawal(t *testing.T) {
	event1 := portal.MezoBridgeAssetsUnlockConfirmed{
		UnlockSequenceNumber: big.NewInt(1),
		Recipient:            common.HexToHash("0x0A219c03938FBC93aA23cAd65f7c480f52665C2a"),
		Token:                common.HexToAddress("0x3A128b915bee3645396d43Fe7A13A59a66C427d6"),
		Amount:               big.NewInt(1000000),
	}

	event2 := portal.MezoBridgeAssetsUnlockConfirmed{
		UnlockSequenceNumber: big.NewInt(2),
		Recipient:            common.HexToHash("0x0A219c03938FBC93aA23cAd65f7c480f52665C2a"),
		Token:                common.HexToAddress("0x3A128b915bee3645396d43Fe7A13A59a66C427d6"),
		Amount:               big.NewInt(2000000),
	}

	env := &environment{
		logger: log.NewNopLogger(),
	}
	bwj := btcWithdrawalJob{
		env: env,
		btcWithdrawalQueue: []portal.MezoBridgeAssetsUnlockConfirmed{
			event1, event2,
		},
	}

	event := bwj.dequeueBTCWithdrawal()
	require.Equal(t, event1, *event)
	require.Equal(
		t,
		[]portal.MezoBridgeAssetsUnlockConfirmed{event2},
		bwj.btcWithdrawalQueue,
	)

	event = bwj.dequeueBTCWithdrawal()
	require.Equal(t, event2, *event)
	require.Equal(
		t,
		[]portal.MezoBridgeAssetsUnlockConfirmed{},
		bwj.btcWithdrawalQueue,
	)

	event = bwj.dequeueBTCWithdrawal()
	require.Nil(t, event)
	require.Equal(
		t,
		[]portal.MezoBridgeAssetsUnlockConfirmed{},
		bwj.btcWithdrawalQueue,
	)
}

func TestBtcToErc20Amount(t *testing.T) {
	fromDec := func(s string) *big.Int {
		n, ok := new(big.Int).SetString(s, 10)
		if !ok {
			t.Fatalf("invalid decimal: %s", s)
		}
		return n
	}

	// Positive
	erc20Amount := btcToErc20Amount(fromDec("123456789"))
	expectedErc20Amount := fromDec("1234567890000000000")
	require.Equal(t, expectedErc20Amount, erc20Amount)

	// Zero
	erc20Amount = btcToErc20Amount(fromDec("0"))
	expectedErc20Amount = fromDec("0")
	require.Equal(t, expectedErc20Amount, erc20Amount)
}

func TestComputeMainUtxoHash(t *testing.T) {
	transactionHash, err := bitcoin.NewHashFromString(
		"089bd0671a4481c3584919b4b9b6751cb3f8586dab41cb157adec43fd10ccc00",
		bitcoin.InternalByteOrder,
	)
	if err != nil {
		t.Fatal(err)
	}

	mainUtxo := &bitcoin.UnspentTransactionOutput{
		Outpoint: &bitcoin.TransactionOutpoint{
			TransactionHash: transactionHash,
			OutputIndex:     5,
		},
		Value: 143565433,
	}

	mainUtxoHash := computeMainUtxoHash(mainUtxo)

	expectedMainUtxoHash, err := hex.DecodeString(
		"1216f8e993c4c57d3c4c971c0d2651140fc4ab09d41960d9ccd7b41fdcd270d6",
	)
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, expectedMainUtxoHash, mainUtxoHash[:])
}

func TestComputeRedemptionKey(t *testing.T) {
	fromHex := func(hexString string) []byte {
		b, err := hex.DecodeString(hexString)
		if err != nil {
			t.Fatal(err)
		}
		return b
	}

	walletPublicKeyHashBytes := fromHex("8db50eb52063ea9d98b3eac91489a90f738986f6")
	var walletPublicKeyHash [20]byte
	copy(walletPublicKeyHash[:], walletPublicKeyHashBytes)

	redeemerOutputScript := fromHex("76a9144130879211c54df460e484ddf9aac009cb38ee7488ac")

	redemptionKey, err := computeRedemptionKey(walletPublicKeyHash, redeemerOutputScript)
	if err != nil {
		t.Fatal(err)
	}

	expectedRedemptionKey := "cb493004c645792101cfa4cc5da4c16aa3148065034371a6f1478b7df4b92d39"

	require.Equal(t, expectedRedemptionKey, redemptionKey.Text(16))
}

func TestComputeAttestationKey(t *testing.T) {
	unlockSeq := big.NewInt(1)
	token := common.HexToAddress("0x5FbDB2315678afecb367f032d93F642f64180aa3")
	amount := big.NewInt(1000)
	chain := uint8(1)
	recipient, err := hex.DecodeString(
		"1976a914f4eedc8f40d4b8e30771f792b065ebec0abaddef88ac",
	)
	if err != nil {
		t.Fatal(err)
	}

	expectedHash, err := hex.DecodeString(
		"aa4b0f8491b6dacd340e19be39225c5ac39da0e242b23c84b4ff15e7cf8c948d",
	)
	if err != nil {
		t.Fatal(err)
	}

	hash, err := computeAttestationKey(
		unlockSeq,
		recipient,
		token,
		amount,
		chain,
	)
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, expectedHash, hash[:])
}

type stubAssetsUnlockedEndpoint struct {
	events []bridgetypes.AssetsUnlockedEvent
}

func (s *stubAssetsUnlockedEndpoint) GetAssetsUnlockedEvents(
	_ context.Context,
	_ sdkmath.Int,
	_ sdkmath.Int,
) ([]bridgetypes.AssetsUnlockedEvent, error) {
	return s.events, nil
}

func TestPrepareBTCWithdrawal(t *testing.T) {
	failingWallet := [20]byte{0x11}
	fundedWallet := [20]byte{0x22}

	recipientScript, err := bitcoin.PayToWitnessPublicKeyHash([20]byte{0x33})
	require.NoError(t, err)
	recipient, err := recipientScript.ToVarLenData()
	require.NoError(t, err)

	event := &portal.MezoBridgeAssetsUnlockConfirmed{
		UnlockSequenceNumber: big.NewInt(1),
		Amount:               btcToErc20Amount(big.NewInt(50000000)), // 0.5 BTC
		Chain:                bitcoinTargetChain,
	}

	// The funded wallet's main UTXO holds 1 BTC.
	fundedWalletScript, err := bitcoin.PayToWitnessPublicKeyHash(fundedWallet)
	require.NoError(t, err)
	fundedWalletTx := &bitcoin.Transaction{
		Version: 1,
		Outputs: []*bitcoin.TransactionOutput{
			{Value: 100000000, PublicKeyScript: fundedWalletScript},
		},
	}
	fundedWalletMainUtxo := &bitcoin.UnspentTransactionOutput{
		Outpoint: &bitcoin.TransactionOutpoint{
			TransactionHash: fundedWalletTx.Hash(),
			OutputIndex:     0,
		},
		Value: 100000000,
	}

	errHistory := errors.New("history too large")

	tests := map[string]struct {
		expectFailingWallet func(*MockBitcoinChain)
		expectFundedWallet  bool
	}{
		"first wallet's history cannot be fetched": {
			expectFailingWallet: func(m *MockBitcoinChain) {
				m.EXPECT().GetTxHashesForPublicKeyHash(failingWallet).
					Return(nil, errHistory).Times(1)
			},
			expectFundedWallet: true,
		},
		"first wallet's main UTXO is not in its history": {
			expectFailingWallet: func(m *MockBitcoinChain) {
				m.EXPECT().GetTxHashesForPublicKeyHash(failingWallet).
					Return([]bitcoin.Hash{}, nil).Times(1)
			},
			expectFundedWallet: true,
		},
		"no wallet's main UTXO can be determined": {
			expectFailingWallet: func(m *MockBitcoinChain) {
				m.EXPECT().GetTxHashesForPublicKeyHash(failingWallet).
					Return(nil, errHistory).Times(1)
				m.EXPECT().GetTxHashesForPublicKeyHash(fundedWallet).
					Return(nil, errHistory).Times(1)
			},
		},
	}

	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			mockTbtcBridgeContract := NewMockTbtcBridgeContract(ctrl)
			mockBitcoinChain := NewMockBitcoinChain(ctrl)

			mockTbtcBridgeContract.EXPECT().Wallets(failingWallet).Return(
				tbtc.Wallet{State: walletStateLive, MainUtxoHash: [32]byte{0x01}},
				nil,
			).AnyTimes()
			mockTbtcBridgeContract.EXPECT().Wallets(fundedWallet).Return(
				tbtc.Wallet{
					State:        walletStateLive,
					MainUtxoHash: computeMainUtxoHash(fundedWalletMainUtxo),
				},
				nil,
			).AnyTimes()

			test.expectFailingWallet(mockBitcoinChain)

			if test.expectFundedWallet {
				mockBitcoinChain.EXPECT().GetTxHashesForPublicKeyHash(fundedWallet).
					Return([]bitcoin.Hash{fundedWalletTx.Hash()}, nil).Times(1)
				mockBitcoinChain.EXPECT().GetTransaction(fundedWalletTx.Hash()).
					Return(fundedWalletTx, nil).Times(1)
				mockTbtcBridgeContract.EXPECT().PendingRedemptions(gomock.Any()).
					Return(tbtc.RedemptionRequest{}, nil).Times(1)
			}

			bwj := btcWithdrawalJob{
				env: &environment{
					logger:             log.NewNopLogger(),
					tbtcBridgeContract: mockTbtcBridgeContract,
					btcChain:           mockBitcoinChain,
				},
				assetsUnlockedEndpoint: &stubAssetsUnlockedEndpoint{
					events: []bridgetypes.AssetsUnlockedEvent{
						{Recipient: recipient},
					},
				},
				liveWallets: [][20]byte{failingWallet, fundedWallet},
			}

			assetsUnlocked, walletPublicKeyHash, mainUtxo, err := bwj.prepareBTCWithdrawal(
				context.Background(),
				event,
			)

			if test.expectFundedWallet {
				require.NoError(t, err)
				require.Equal(t, recipient, assetsUnlocked.Recipient)
				require.Equal(t, fundedWallet, walletPublicKeyHash)
				require.Equal(
					t,
					portal.BitcoinTxUTXO{
						TxHash:        fundedWalletTx.Hash(),
						TxOutputIndex: 0,
						TxOutputValue: 100000000,
					},
					*mainUtxo,
				)
				return
			}

			require.ErrorContains(t, err, "cannot find wallet to cover BTC withdrawal")
			require.ErrorContains(t, err, hex.EncodeToString(failingWallet[:]))
			require.ErrorContains(t, err, hex.EncodeToString(fundedWallet[:]))
			require.ErrorContains(t, err, errHistory.Error())
		})
	}
}
