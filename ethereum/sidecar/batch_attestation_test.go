package sidecar

import (
	"context"
	"crypto/ecdsa"
	"errors"
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
		address,
		mockBridgeWorker,
		mockBridgeContract,
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
					SendSignature(tba.address, gomock.Any()).
					Return(expectedError).
					AnyTimes()
			},
			ctx:             context.Background,
			expectedSuccess: false,
			expectedError:   "",
			timeout:         100 * time.Millisecond, // Short timeout for testing
		},
		{
			name:        "SendSignature succeeds but confirmation never comes",
			attestation: defaultUnlockAttestation(),
			pre: func(tba *testBatchAttestation) {
				tba.mockBridgeWorker.EXPECT().
					SendSignature(tba.address, gomock.Any()).
					Return(nil).
					Times(1)
				tba.mockBridgeContract.EXPECT().
					ConfirmedUnlocks(defaultUnlockAttestation().UnlockSequenceNumber).
					Return(false, nil).
					AnyTimes()
			},
			ctx:             context.Background,
			expectedSuccess: false,
			expectedError:   "",
			timeout:         100 * time.Millisecond,
		},
		{
			name:        "SendSignature succeeds and confirmation comes immediately",
			attestation: defaultUnlockAttestation(),
			pre: func(tba *testBatchAttestation) {
				tba.mockBridgeWorker.EXPECT().
					SendSignature(tba.address, gomock.Any()).
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
					SendSignature(tba.address, gomock.Any()).
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
					SendSignature(tba.address, gomock.Any()).
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
			expectedError:   "",
			timeout:         100 * time.Millisecond,
		},
		{
			name:        "Context canceled during execution",
			attestation: defaultUnlockAttestation(),
			pre:         func(_ *testBatchAttestation) {},
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			},
			expectedSuccess: false,
			expectedError:   "context canceled",
			timeout:         5 * time.Second,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			tba := newTestBatchAttestation(t)

			// Override the default timeout
			defaultBatchAttestationTimeout = testCase.timeout

			// Also set a shorter check interval
			defaultBatchAttestationCheck = 10 * time.Millisecond

			// Set a shorter retry interval
			defaultRetrySendSignature = 10 * time.Millisecond

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
