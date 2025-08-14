package sidecar

import (
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"testing"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/keep-network/keep-common/pkg/chain/ethereum/ethutil"
	bridgetypes "github.com/mezo-org/mezod/x/bridge/types"
	"github.com/stretchr/testify/assert"
	gomock "go.uber.org/mock/gomock"
)

var (
	privateKey    = loadPrivateKey("ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80")
	bridgeAddress = common.HexToAddress("0xB81057aB093B161b53049bDC9edb6c6cE8472784")
)

type mockChain struct {
	client *MockEthereumClient
}

func (c *mockChain) ChainID() *big.Int {
	return big.NewInt(1)
}

func (c *mockChain) Client() ethutil.EthereumClient {
	return c.client
}

type mockBridgeTransactor struct {
	tx  *types.Transaction
	err error
}

func (bt *mockBridgeTransactor) AttestBridgeOut(
	_ *bind.TransactOpts,
	_ *bridgetypes.AssetsUnlockedEvent,
) (*types.Transaction, error) {
	return bt.tx, bt.err
}

type testTransactionExecutor struct {
	*IndividualAttestationTransactionExecutor

	t                    *testing.T
	ctrl                 *gomock.Controller
	mockChain            *mockChain
	mockBridgeTransactor *mockBridgeTransactor
}

func newTransactionExecutor(t *testing.T) *testTransactionExecutor {
	t.Helper()

	logger := log.NewTestLogger(t)
	ctrl := gomock.NewController(t)
	mockChain := &mockChain{
		client: NewMockEthereumClient(ctrl),
	}
	mockBridgeTransactor := &mockBridgeTransactor{}
	iate, err := NewIndividualAttestationTransactionExecutor(
		logger, privateKey, mockChain, bridgeAddress, mockBridgeTransactor,
	)
	if err != nil {
		t.Fatalf("couldn't create IndividualAttestationTransactionExecutor: %v", err)
	}

	return &testTransactionExecutor{
		IndividualAttestationTransactionExecutor: iate,
		t:                                        t,
		ctrl:                                     ctrl,
		mockChain:                                mockChain,
		mockBridgeTransactor:                     mockBridgeTransactor,
	}
}

func TestTransactionExecutor(t *testing.T) {
	testCases := []struct {
		name        string
		attestation *bridgetypes.AssetsUnlockedEvent
		pre         func(tte *testTransactionExecutor)
		post        func(tte *testTransactionExecutor)
		expectErr   string
	}{
		{
			name:        "SuggestGasPrice failed",
			attestation: defaultAttestation(),
			pre: func(tte *testTransactionExecutor) {
				expectedError := errors.New("gas price estimation failed")
				tte.mockChain.client.EXPECT().
					SuggestGasPrice(gomock.Any()).
					Return(nil, expectedError).
					Times(1)
			},
			expectErr: "gas price estimation failed",
		},
		{
			name:        "PendingNonceAt failed",
			attestation: defaultAttestation(),
			pre: func(tte *testTransactionExecutor) {
				tte.mockChain.client.EXPECT().
					SuggestGasPrice(gomock.Any()).
					Return(big.NewInt(10), nil).
					Times(1)
				expectedError := errors.New("pending nonce failed")
				tte.mockChain.client.EXPECT().
					PendingNonceAt(gomock.Any(), gomock.Any()).
					Return(uint64(0), expectedError).
					Times(1)
			},
			expectErr: "pending nonce failed",
		},
		{
			name:        "EstimateGas failed",
			attestation: defaultAttestation(),
			pre: func(tte *testTransactionExecutor) {
				tte.mockChain.client.EXPECT().
					SuggestGasPrice(gomock.Any()).
					Return(big.NewInt(10), nil).
					Times(1)
				tte.mockChain.client.EXPECT().
					PendingNonceAt(gomock.Any(), gomock.Any()).
					Return(uint64(42), nil).
					Times(1)
				expectedError := errors.New("estimate gas failed")
				tte.mockChain.client.EXPECT().
					EstimateGas(gomock.Any(), gomock.Any()).
					Return(uint64(0), expectedError).
					Times(1)
			},
			expectErr: "estimate gas failed",
		},
		{
			name:        "AttestBridgeOut failed",
			attestation: defaultAttestation(),
			pre: func(tte *testTransactionExecutor) {
				tte.mockChain.client.EXPECT().
					SuggestGasPrice(gomock.Any()).
					Return(big.NewInt(10), nil).
					Times(1)
				tte.mockChain.client.EXPECT().
					PendingNonceAt(gomock.Any(), gomock.Any()).
					Return(uint64(42), nil).
					Times(1)
				tte.mockChain.client.EXPECT().
					EstimateGas(gomock.Any(), gomock.Any()).
					Return(uint64(5000), nil).
					Times(1)
				tte.mockBridgeTransactor.err = errors.New("attest bridge out failed")
				tte.mockBridgeTransactor.tx = nil
			},
			expectErr: "attest bridge out failed",
		},
		{
			name:        "AttestBridgeOut succeed",
			attestation: defaultAttestation(),
			pre: func(tte *testTransactionExecutor) {
				tte.mockChain.client.EXPECT().
					SuggestGasPrice(gomock.Any()).
					Return(big.NewInt(10), nil).
					Times(1)
				tte.mockChain.client.EXPECT().
					PendingNonceAt(gomock.Any(), gomock.Any()).
					Return(uint64(42), nil).
					Times(1)
				tte.mockChain.client.EXPECT().
					EstimateGas(gomock.Any(), gomock.Any()).
					Return(uint64(5000), nil).
					Times(1)
				tte.mockBridgeTransactor.err = nil
				tte.mockBridgeTransactor.tx = types.NewTx(defaultTxData())

				tte.mockChain.client.EXPECT().
					TransactionReceipt(gomock.Any(), gomock.Any()).
					Return(&types.Receipt{Status: 1}, nil).
					Times(1)
			},
			post: func(tte *testTransactionExecutor) {
				// extra asserts
				assert.Equal(t, tte.auth.GasPrice.Cmp(adjustGasPrice(big.NewInt(10))), 0, "invalid adjusted gas price")
				assert.Equal(t, tte.auth.GasLimit, adjustGasLimit(5000), "invalid adjusted gas limit")
			},
			expectErr: "",
		},
	}

	// set the default transaction wait to 0
	defaultTransactionReceiptTicker = 1

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			tte := newTransactionExecutor(t)

			// prepare the test
			testCase.pre(tte)

			// execute the transaction
			err := tte.Send(testCase.attestation)

			// verify
			if len(testCase.expectErr) > 0 {
				assert.ErrorContains(t, err, testCase.expectErr, "not the expected error")
			} else {
				assert.NoError(t, err, "expected no error")
			}

			if testCase.post != nil {
				testCase.post(tte)
			}

			tte.ctrl.Finish()
		})
	}
}

func defaultAttestation() *bridgetypes.AssetsUnlockedEvent {
	return &bridgetypes.AssetsUnlockedEvent{
		UnlockSequence: math.NewInt(1),
		Recipient:      []byte{},
		Token:          "0x...",
		Sender:         "0x...",
		Amount:         math.NewInt(1),
		Chain:          0,
		BlockTime:      1755331749,
	}
}

func defaultTxData() types.TxData {
	return &types.LegacyTx{
		Nonce:    0,
		GasPrice: big.NewInt(1000000000),
		Gas:      21000,
		To:       &common.Address{},
		Value:    big.NewInt(1000000000000000000),
		Data:     []byte{},
		V:        big.NewInt(27),
		R:        big.NewInt(1),
		S:        big.NewInt(1),
	}
}

func loadPrivateKey(hexKey string) *ecdsa.PrivateKey {
	if len(hexKey) >= 2 && hexKey[0:2] == "0x" {
		hexKey = hexKey[2:]
	}

	privateKeyBytes, err := hex.DecodeString(hexKey)
	if err != nil {
		panic(fmt.Sprintf("failed to decode hex string: %v", err))
	}

	privateKey, err := crypto.ToECDSA(privateKeyBytes)
	if err != nil {
		panic(fmt.Sprintf("failed to load private key: %v", err))
	}

	return privateKey
}
