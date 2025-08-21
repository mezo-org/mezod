package sidecar

import (
	"errors"
	"math/big"
	"testing"
	"time"

	"cosmossdk.io/log"
	"github.com/ethereum/go-ethereum/common"
	"github.com/mezo-org/mezod/ethereum/bindings/portal"
	"github.com/stretchr/testify/assert"
	gomock "go.uber.org/mock/gomock"
)

type testSubmissionQueue struct {
	*submissionQueue

	t                  *testing.T
	ctrl               *gomock.Controller
	mockBridgeContract *MockBridgeContract
}

func newTestSubmissionQueue(t *testing.T) *testSubmissionQueue {
	t.Helper()

	ctrl := gomock.NewController(t)
	mockBridgeContract := NewMockBridgeContract(ctrl)
	address := common.HexToAddress("0x123")
	sq := newSubmissionQueue(
		log.NewNopLogger(),
		mockBridgeContract,
		address,
	)

	return &testSubmissionQueue{
		submissionQueue:    sq,
		t:                  t,
		ctrl:               ctrl,
		mockBridgeContract: mockBridgeContract,
	}
}

func TestSubmissionQueue_GetSubmissionDelay(t *testing.T) {
	testCases := []struct {
		name          string
		attestation   *portal.MezoBridgeAssetsUnlocked
		pre           func(tsq *testSubmissionQueue)
		post          func(tsq *testSubmissionQueue)
		expectedDelay time.Duration
		expectErr     string
	}{
		{
			name:        "BridgeValidatorsCount fails",
			attestation: defaultUnlockAttestation(),
			pre: func(tsq *testSubmissionQueue) {
				expectedError := errors.New("network error")
				tsq.mockBridgeContract.EXPECT().
					BridgeValidatorsCount().
					Return(nil, expectedError).
					Times(1)
			},
			expectedDelay: time.Duration(0),
		},
		{
			name:        "No validators found",
			attestation: defaultUnlockAttestation(),
			pre: func(tsq *testSubmissionQueue) {
				tsq.mockBridgeContract.EXPECT().
					BridgeValidatorsCount().
					Return(big.NewInt(0), nil).
					Times(1)
			},
			expectedDelay: time.Duration(0),
		},
		{
			name:        "ValidatorIDs fails",
			attestation: defaultUnlockAttestation(),
			pre: func(tsq *testSubmissionQueue) {
				tsq.mockBridgeContract.EXPECT().
					BridgeValidatorsCount().
					Return(big.NewInt(3), nil).
					Times(1)
				expectedError := errors.New("validator ID error")
				tsq.mockBridgeContract.EXPECT().
					ValidatorIDs(gomock.Any()).
					Return(uint8(0), expectedError).
					Times(1)
			},
			expectedDelay: time.Duration(0),
		},
		{
			name:        "Single validator - should have no delay",
			attestation: defaultUnlockAttestation(),
			pre: func(tsq *testSubmissionQueue) {
				tsq.mockBridgeContract.EXPECT().
					BridgeValidatorsCount().
					Return(big.NewInt(1), nil).
					Times(1)
				tsq.mockBridgeContract.EXPECT().
					ValidatorIDs(gomock.Any()).
					Return(uint8(0), nil).
					Times(1)
			},
			expectedDelay: 0,
		},
		{
			name:        "Multiple validators - first in queue",
			attestation: defaultUnlockAttestation(),
			pre: func(tsq *testSubmissionQueue) {
				tsq.mockBridgeContract.EXPECT().
					BridgeValidatorsCount().
					Return(big.NewInt(3), nil).
					Times(1)
				tsq.mockBridgeContract.EXPECT().
					ValidatorIDs(gomock.Any()).
					Return(uint8(0), nil).
					Times(1)
			},
			expectedDelay: 0,
		},
		{
			name:        "Multiple validators - last in queue",
			attestation: defaultUnlockAttestation(),
			pre: func(tsq *testSubmissionQueue) {
				tsq.mockBridgeContract.EXPECT().
					BridgeValidatorsCount().
					Return(big.NewInt(3), nil).
					Times(1)
				tsq.mockBridgeContract.EXPECT().
					ValidatorIDs(gomock.Any()).
					Return(uint8(2), nil).
					Times(1)
			},
			expectedDelay: 1 * time.Minute,
		},
		{
			name:        "Multiple validators - validator not in queue returns 0",
			attestation: defaultUnlockAttestation(),
			pre: func(tsq *testSubmissionQueue) {
				tsq.mockBridgeContract.EXPECT().
					BridgeValidatorsCount().
					Return(big.NewInt(3), nil).
					Times(1)
				tsq.mockBridgeContract.EXPECT().
					ValidatorIDs(gomock.Any()).
					Return(uint8(5), nil).
					Times(1)
			},
			expectedDelay: 0,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			tsq := newTestSubmissionQueue(t)

			// prepare the test
			testCase.pre(tsq)

			// execute the transaction
			delay := tsq.GetSubmissionDelay(testCase.attestation)

			assert.Equal(t, testCase.expectedDelay, delay, "unexpected delay")

			if testCase.post != nil {
				testCase.post(tsq)
			}

			tsq.ctrl.Finish()
		})
	}
}

func TestSubmissionQueue_shuffleValidatorIDs(t *testing.T) {
	tsq := newTestSubmissionQueue(t)

	t.Run("Deterministic shuffle with same seed", func(t *testing.T) {
		validatorIDs1 := []uint8{0, 1, 2, 3, 4}
		validatorIDs2 := []uint8{0, 1, 2, 3, 4}
		seed := int64(12345)

		tsq.shuffleValidatorIDs(validatorIDs1, seed)
		tsq.shuffleValidatorIDs(validatorIDs2, seed)

		assert.Equal(t, validatorIDs1, validatorIDs2, "same seed should produce same shuffle")
	})

	t.Run("Different seeds produce different results", func(t *testing.T) {
		original := []uint8{0, 1, 2, 3, 4}
		validatorIDs1 := make([]uint8, len(original))
		validatorIDs2 := make([]uint8, len(original))
		copy(validatorIDs1, original)
		copy(validatorIDs2, original)

		res1 := tsq.shuffleValidatorIDs(validatorIDs1, 12345)
		res2 := tsq.shuffleValidatorIDs(validatorIDs2, 54321)

		// Very unlikely to be the same after shuffle with different seeds
		assert.NotEqual(t, res1, res2, "different seeds should produce different shuffles")
	})
}

func defaultUnlockAttestation() *portal.MezoBridgeAssetsUnlocked {
	return &portal.MezoBridgeAssetsUnlocked{
		UnlockSequenceNumber: big.NewInt(1),
		Recipient:            []byte{},
		Token:                common.Address{},
		Amount:               big.NewInt(100),
		Chain:                0,
	}
}
