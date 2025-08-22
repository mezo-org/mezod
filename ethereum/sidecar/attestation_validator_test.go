package sidecar

import (
	"context"
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

var bridgeAddress = common.HexToAddress("0xB81057aB093B161b53049bDC9edb6c6cE8472784")

type testAttestationValidator struct {
	*attestationValidator

	t                  *testing.T
	ctrl               *gomock.Controller
	mockBridgeContract *MockBridgeContract
}

func newTestAttestationValidator(t *testing.T) *testAttestationValidator {
	t.Helper()

	ctrl := gomock.NewController(t)
	mockBridgeContract := NewMockBridgeContract(ctrl)
	av := newAttestationValidation(
		log.NewNopLogger(), mockBridgeContract, bridgeAddress,
	)

	return &testAttestationValidator{
		attestationValidator: av,
		t:                    t,
		ctrl:                 ctrl,
		mockBridgeContract:   mockBridgeContract,
	}
}

func TestAttestationValidationIsValid(t *testing.T) {
	testCases := []struct {
		name        string
		attestation *portal.MezoBridgeAssetsUnlocked
		pre         func(tte *testAttestationValidator)
		getCtx      func() context.Context
		expect      bool
		expectErr   string
	}{
		{
			name:        "IsValid - succeess",
			attestation: defaultAttestation(),
			pre: func(tav *testAttestationValidator) {
				defaultIsValidTickerDuration = 1
				tav.mockBridgeContract.EXPECT().
					ValidateAssetsUnlocked(gomock.Any()).
					Return(true, nil).
					Times(1)
			},
			getCtx: context.Background,
			expect: true,
		},
		{
			name:        "IsValid - failure",
			attestation: defaultAttestation(),
			pre: func(tav *testAttestationValidator) {
				defaultIsValidTickerDuration = 1
				tav.mockBridgeContract.EXPECT().
					ValidateAssetsUnlocked(gomock.Any()).
					Return(false, nil).
					Times(1)
			},
			getCtx: context.Background,
			expect: false,
		},
		{
			name:        "ctx canceled",
			attestation: defaultAttestation(),
			pre: func(_ *testAttestationValidator) {
				// just to make sure the context cancel is trigger first
				defaultIsValidTickerDuration = 10 * time.Second
			},
			getCtx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			},
			expect:    false,
			expectErr: "context canceled",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			tte := newTestAttestationValidator(t)

			// prepare the test
			testCase.pre(tte)

			ctx := testCase.getCtx()

			// execute the transaction
			ok, err := tte.IsValid(ctx, testCase.attestation)
			assert.Equal(t, ok, testCase.expect)
			if len(testCase.expectErr) > 0 {
				assert.ErrorContains(t, err, testCase.expectErr)
			}

			tte.ctrl.Finish()
		})
	}
}

func TestAttestationValidation(t *testing.T) {
	testCases := []struct {
		name        string
		attestation *portal.MezoBridgeAssetsUnlocked
		pre         func(tte *testAttestationValidator)
		post        func(tte *testAttestationValidator)
		expectErr   string
	}{
		{
			name:        "ConfirmedUnlocks failed with error",
			attestation: defaultAttestation(),
			pre: func(tav *testAttestationValidator) {
				expectedError := errors.New("network error")
				tav.mockBridgeContract.EXPECT().
					ConfirmedUnlocks(gomock.Any()).
					Return(false, expectedError).
					Times(1)
			},
			expectErr: "network error",
		},
		{
			name:        "ConfirmedUnlocks is confirmed",
			attestation: defaultAttestation(),
			pre: func(tav *testAttestationValidator) {
				tav.mockBridgeContract.EXPECT().
					ConfirmedUnlocks(gomock.Any()).
					Return(true, nil).
					Times(1)
			},
		},
		{
			name:        "Attestation failed with error",
			attestation: defaultAttestation(),
			pre: func(tav *testAttestationValidator) {
				tav.mockBridgeContract.EXPECT().
					ConfirmedUnlocks(gomock.Any()).
					Return(false, nil).
					Times(1)
				expectedError := errors.New("network error")
				tav.mockBridgeContract.EXPECT().
					Attestations(gomock.Any()).
					Return(big.NewInt(0), expectedError).
					Times(1)
			},
			expectErr: "network error",
		},
		{
			name:        "ValidatorIDs failed with error",
			attestation: defaultAttestation(),
			pre: func(tav *testAttestationValidator) {
				tav.mockBridgeContract.EXPECT().
					ConfirmedUnlocks(gomock.Any()).
					Return(false, nil).
					Times(1)
				tav.mockBridgeContract.EXPECT().
					Attestations(gomock.Any()).
					Return(big.NewInt(0), nil).
					Times(1)
				expectedError := errors.New("network error")
				tav.mockBridgeContract.EXPECT().
					ValidatorIDs(gomock.Any()).
					Return(uint8(10), expectedError).
					Times(1)
			},
			expectErr: "network error",
		},
		{
			name:        "Validation succeeded, nothing to do",
			attestation: defaultAttestation(),
			pre: func(tav *testAttestationValidator) {
				tav.mockBridgeContract.EXPECT().
					ConfirmedUnlocks(gomock.Any()).
					Return(false, nil).
					Times(1)
				tav.mockBridgeContract.EXPECT().
					Attestations(gomock.Any()).
					Return(new(big.Int).SetBit(big.NewInt(0), int(10), 1), nil).
					Times(1)
				tav.mockBridgeContract.EXPECT().
					ValidatorIDs(gomock.Any()).
					Return(uint8(10), nil).
					Times(1)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			tte := newTestAttestationValidator(t)

			// prepare the test
			testCase.pre(tte)

			// execute the transaction
			ok, err := tte.IsConfirmed(testCase.attestation)

			// verify
			if len(testCase.expectErr) > 0 {
				assert.ErrorContains(t, err, testCase.expectErr, "not the expected error")
				assert.False(t, ok)
			} else {
				assert.NoError(t, err, "expected no error")
				assert.True(t, ok)
			}

			if testCase.post != nil {
				testCase.post(tte)
			}

			tte.ctrl.Finish()
		})
	}
}

func defaultAttestation() *portal.MezoBridgeAssetsUnlocked {
	return &portal.MezoBridgeAssetsUnlocked{
		UnlockSequenceNumber: big.NewInt(1),
		Recipient:            []byte{},
		Token:                common.Address{},
		Amount:               big.NewInt(1),
		Chain:                0,
	}
}
