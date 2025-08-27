package sidecar

import (
	"errors"
	"math/big"
	"testing"

	"cosmossdk.io/log"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
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

func TestAbiEncodeAttestation(t *testing.T) {
	attestation := &portal.MezoBridgeAssetsUnlocked{
		UnlockSequenceNumber: big.NewInt(10),
		Recipient:            common.HexToAddress("0x87eaCD568b85dA3cfF39D3bb82F9329B23786b76").Bytes(),
		Token:                common.HexToAddress("0x517f2982701695D4E52f1ECFBEf3ba31Df470161"),
		Amount:               big.NewInt(77),
		Chain:                0,
	}

	encoded, err := abiEncodeAttestation(attestation)
	assert.NoError(t, err)

	hash := crypto.Keccak256Hash(encoded)
	// Expected hash computed using Solidity's `keccak256(abi.encode(AssetsUnlocked))`
	expectedHash := common.HexToHash("0x68e75c66f0779e7c240868e0c8149c51f14fdd74aad9a3eb2781500edfde3137")

	assert.Equal(t, expectedHash, hash)
}

func TestAbiEncodeAttestationWithChainID(t *testing.T) {
	attestation := &portal.MezoBridgeAssetsUnlocked{
		UnlockSequenceNumber: big.NewInt(10),
		Recipient:            common.HexToAddress("0x87eaCD568b85dA3cfF39D3bb82F9329B23786b76").Bytes(),
		Token:                common.HexToAddress("0x517f2982701695D4E52f1ECFBEf3ba31Df470161"),
		Amount:               big.NewInt(77),
		Chain:                0,
	}

	encoded, err := abiEncodeAttestationWithChainID(attestation, big.NewInt(1))
	assert.NoError(t, err)

	hash := crypto.Keccak256Hash(encoded)
	// Expected hash computed using Solidity's `keccak256(abi.encode(1, AssetsUnlocked))`
	expectedHash := common.HexToHash("0xa0c5a45f9393426db79c98ccd594e6f8ca6683268ee7e95101a4c85111f53318")

	assert.Equal(t, expectedHash, hash)
}
