package sidecar

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"math/big"
	"testing"
	"time"

	"cosmossdk.io/log"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/mezo-org/mezod/ethereum/bindings/portal"
	"github.com/stretchr/testify/assert"
	gomock "go.uber.org/mock/gomock"
)

type testBatchAttestation struct {
	*batchAttestation

	t                  *testing.T
	ctrl               *gomock.Controller
	mockBridgeWorker   *MockBridgeWorker
	mockBridgeContract *MockBridgeContract
	privateKey         *ecdsa.PrivateKey
	address            common.Address
}

func newTestBatchAttestation(t *testing.T) *testBatchAttestation {
	t.Helper()

	ctrl := gomock.NewController(t)
	mockBridgeWorker := NewMockBridgeWorker(ctrl)
	mockBridgeContract := NewMockBridgeContract(ctrl)

	// Generate a test private key
	privateKey, err := crypto.GenerateKey()
	assert.NoError(t, err)

	address := crypto.PubkeyToAddress(privateKey.PublicKey)

	ba := newBatchAttestation(
		log.NewNopLogger(),
		privateKey,
		mockBridgeWorker,
		mockBridgeContract,
		big.NewInt(1),
	)

	return &testBatchAttestation{
		batchAttestation:   ba,
		t:                  t,
		ctrl:               ctrl,
		mockBridgeWorker:   mockBridgeWorker,
		mockBridgeContract: mockBridgeContract,
		privateKey:         privateKey,
		address:            address,
	}
}

func TestBatchAttestation_TryAttest(t *testing.T) {
	testCases := []struct {
		name            string
		attestation     *portal.MezoBridgeAssetsUnlocked
		pre             func(tba *testBatchAttestation)
		post            func(tba *testBatchAttestation)
		ctx             func() context.Context
		expectedSuccess bool
		expectedError   string
		timeout         time.Duration
	}{
		{
			name:        "SendSignature fails repeatedly until timeout",
			attestation: defaultUnlockAttestation(),
			pre: func(tba *testBatchAttestation) {
				expectedError := errors.New("network error")
				tba.mockBridgeWorker.EXPECT().
					SendSignature(defaultUnlockAttestation(), gomock.Any()).
					Return(expectedError).
					AnyTimes()
			},
			ctx:             context.Background,
			expectedSuccess: false,
			expectedError:   "failed to send payload: payload send terminated: context deadline exceeded",
			timeout:         100 * time.Millisecond, // Short timeout for testing
		},
		{
			name:        "SendSignature succeeds but confirmation never comes",
			attestation: defaultUnlockAttestation(),
			pre: func(tba *testBatchAttestation) {
				tba.mockBridgeWorker.EXPECT().
					SendSignature(defaultUnlockAttestation(), gomock.Any()).
					Return(nil).
					Times(1)
				tba.mockBridgeContract.EXPECT().
					ConfirmedUnlocks(defaultUnlockAttestation().UnlockSequenceNumber).
					Return(false, nil).
					AnyTimes()
			},
			ctx:             context.Background,
			expectedSuccess: false,
			expectedError:   "batch attestation terminated: context deadline exceeded",
			timeout:         100 * time.Millisecond,
		},
		{
			name:        "SendSignature succeeds and confirmation comes immediately",
			attestation: defaultUnlockAttestation(),
			pre: func(tba *testBatchAttestation) {
				tba.mockBridgeWorker.EXPECT().
					SendSignature(defaultUnlockAttestation(), gomock.Any()).
					Return(nil).
					Times(1)
				tba.mockBridgeContract.EXPECT().
					ConfirmedUnlocks(defaultUnlockAttestation().UnlockSequenceNumber).
					Return(true, nil).
					Times(1)
			},
			ctx:             context.Background,
			expectedSuccess: true,
			expectedError:   "",
			timeout:         5 * time.Second,
		},
		{
			name:        "SendSignature succeeds and confirmation comes after delay",
			attestation: defaultUnlockAttestation(),
			pre: func(tba *testBatchAttestation) {
				tba.mockBridgeWorker.EXPECT().
					SendSignature(defaultUnlockAttestation(), gomock.Any()).
					Return(nil).
					Times(1)
				// First few calls return false, then true
				gomock.InOrder(
					tba.mockBridgeContract.EXPECT().
						ConfirmedUnlocks(defaultUnlockAttestation().UnlockSequenceNumber).
						Return(false, nil).
						Times(2),
					tba.mockBridgeContract.EXPECT().
						ConfirmedUnlocks(defaultUnlockAttestation().UnlockSequenceNumber).
						Return(true, nil).
						Times(1),
				)
			},
			ctx:             context.Background,
			expectedSuccess: true,
			expectedError:   "",
			timeout:         5 * time.Second,
		},
		{
			name:        "SendSignature succeeds but ConfirmedUnlocks returns error",
			attestation: defaultUnlockAttestation(),
			pre: func(tba *testBatchAttestation) {
				tba.mockBridgeWorker.EXPECT().
					SendSignature(defaultUnlockAttestation(), gomock.Any()).
					Return(nil).
					Times(1)
				expectedError := errors.New("contract error")
				tba.mockBridgeContract.EXPECT().
					ConfirmedUnlocks(defaultUnlockAttestation().UnlockSequenceNumber).
					Return(false, expectedError).
					AnyTimes()
			},
			ctx:             context.Background,
			expectedSuccess: false,
			expectedError:   "batch attestation terminated: context deadline exceeded",
			timeout:         100 * time.Millisecond,
		},
		{
			name:        "Context canceled during execution",
			attestation: defaultUnlockAttestation(),
			pre:         func(_ *testBatchAttestation) {},
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				// cancel the context immediately so we fail at payload send
				cancel()
				return ctx
			},
			expectedSuccess: false,
			expectedError:   "failed to send payload: payload send terminated: context canceled",
			timeout:         5 * time.Second,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			tba := newTestBatchAttestation(t)

			// Override the default timeout
			batchAttestationTimeout = testCase.timeout

			// Also set a shorter check interval
			batchAttestationCheck = 10 * time.Millisecond

			// Set a shorter retry interval
			retrySendSignature = 10 * time.Millisecond

			// Prepare the test
			testCase.pre(tba)

			// Create context that might be canceled for one test case
			ctx := testCase.ctx()

			// Execute the transaction
			success, err := tba.TryAttest(ctx, testCase.attestation)

			// Verify results
			assert.Equal(t, testCase.expectedSuccess, success, "unexpected success value")

			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)
			} else {
				assert.NoError(t, err)
			}

			if testCase.post != nil {
				testCase.post(tba)
			}

			tba.ctrl.Finish()
		})
	}
}

func TestAttestationDigestHash(t *testing.T) {
	attestation := &portal.MezoBridgeAssetsUnlocked{
		UnlockSequenceNumber: big.NewInt(10),
		Recipient:            common.HexToAddress("0x87eaCD568b85dA3cfF39D3bb82F9329B23786b76").Bytes(),
		Token:                common.HexToAddress("0x517f2982701695D4E52f1ECFBEf3ba31Df470161"),
		Amount:               big.NewInt(77),
		Chain:                0,
	}

	digestHash, err := attestationDigestHash(attestation, big.NewInt(1))
	assert.NoError(t, err)

	// Expected digest hash computed using Solidity's `keccak256(abi.encode(1, AssetsUnlocked)).toEthSignedMessageHash()`
	expectedDigestHash := common.HexToHash("0xcb97fcb7f22cc5aadd1b5e6497d547723d4a652d31cf38e78d3644b44a71846d").Bytes()
	assert.Equal(t, expectedDigestHash, digestHash)
}
