package bridgeworker

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/mezo-org/mezod/bridge-worker/types"
	"github.com/mezo-org/mezod/ethereum/bindings/portal"
	bridgetypes "github.com/mezo-org/mezod/x/bridge/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestServer_submitAttestation(t *testing.T) {
	chainID := big.NewInt(1)

	// Generate a test private key for creating valid signatures
	privateKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	validatorAddress := crypto.PubkeyToAddress(privateKey.PublicKey)

	// Helper function to create a valid AssetsUnlocked entry
	createValidEntry := func() *bridgetypes.AssetsUnlockedEvent {
		return &bridgetypes.AssetsUnlockedEvent{
			UnlockSequence: sdkmath.NewInt(1),
			Recipient:      []byte("test-recipient"),
			Token:          "0x1234567890123456789012345678901234567890",
			Sender:         "0x9876543210987654321098765432109876543210",
			Amount:         sdkmath.NewInt(1000),
			Chain:          0,
			BlockTime:      1000,
		}
	}

	// Helper function to create a valid signature for an entry
	createValidSignature := func(entry *bridgetypes.AssetsUnlockedEvent, privKey *ecdsa.PrivateKey) string {
		attestation := &portal.MezoBridgeAssetsUnlocked{
			UnlockSequenceNumber: entry.UnlockSequence.BigInt(),
			Recipient:            entry.Recipient,
			Token:                common.HexToAddress(entry.Token),
			Amount:               entry.Amount.BigInt(),
			Chain:                uint8(entry.Chain),
		}

		hash, err := portal.AttestationDigestHash(attestation, chainID)
		require.NoError(t, err)

		signature, err := crypto.Sign(hash, privKey)
		require.NoError(t, err)

		return hexutil.Encode(signature)
	}

	testCases := []struct {
		name            string
		requestBody     interface{}
		expectedStatus  int
		expectedError   string
		setupRequest    func() *types.SubmitAttestationRequest
		setupBridgeMock func(*MockMezoBridge)
		setupStoreMock  func(*MockStore)
	}{
		{
			name:           "Valid signature submission",
			expectedStatus: http.StatusAccepted,
			setupRequest: func() *types.SubmitAttestationRequest {
				entry := createValidEntry()
				signature := createValidSignature(entry, privateKey)
				return &types.SubmitAttestationRequest{
					Entry:     entry,
					Signature: signature,
				}
			},
			setupBridgeMock: func(mockBridge *MockMezoBridge) {
				mockBridge.EXPECT().ValidatorIDs(validatorAddress).Return(uint8(1), nil)
				mockBridge.EXPECT().ConfirmedUnlocks(big.NewInt(1)).Return(false, nil)
				mockBridge.EXPECT().ValidateAssetsUnlocked(gomock.Any()).Return(true, nil)
			},
			setupStoreMock: func(mockStore *MockStore) {
				mockStore.EXPECT().SaveAttestation(gomock.Any()).Return(nil)
				mockStore.EXPECT().SaveSignature(gomock.Any(), gomock.Any()).Return(nil)
			},
		},
		{
			name:            "Invalid JSON body",
			requestBody:     "invalid-json",
			expectedStatus:  http.StatusBadRequest,
			expectedError:   "invalid json format",
			setupBridgeMock: func(mockBridge *MockMezoBridge) {},
			setupStoreMock:  func(mockStore *MockStore) {},
		},
		{
			name:            "Empty request body",
			requestBody:     "",
			expectedStatus:  http.StatusBadRequest,
			expectedError:   "invalid json format",
			setupBridgeMock: func(mockBridge *MockMezoBridge) {},
			setupStoreMock:  func(mockStore *MockStore) {},
		},
		{
			name:           "Missing Entry field",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "missing assets unlocked entry",
			setupRequest: func() *types.SubmitAttestationRequest {
				return &types.SubmitAttestationRequest{
					Entry:     nil,
					Signature: "0x" + strings.Repeat("00", 65),
				}
			},
			setupBridgeMock: func(mockBridge *MockMezoBridge) {},
			setupStoreMock:  func(mockStore *MockStore) {},
		},
		{
			name:           "Missing signature",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "empty hex string",
			setupRequest: func() *types.SubmitAttestationRequest {
				return &types.SubmitAttestationRequest{
					Entry:     createValidEntry(),
					Signature: "",
				}
			},
			setupBridgeMock: func(mockBridge *MockMezoBridge) {},
			setupStoreMock:  func(mockStore *MockStore) {},
		},
		{
			name:           "Missing 0x prefix in signature",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "hex string without 0x prefix",
			setupRequest: func() *types.SubmitAttestationRequest {
				return &types.SubmitAttestationRequest{
					Entry:     createValidEntry(),
					Signature: strings.Repeat("00", 65),
				}
			},
			setupBridgeMock: func(mockBridge *MockMezoBridge) {},
			setupStoreMock:  func(mockStore *MockStore) {},
		},
		{
			name:           "Invalid hex encoding in signature",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid hex string",
			setupRequest: func() *types.SubmitAttestationRequest {
				return &types.SubmitAttestationRequest{
					Entry:     createValidEntry(),
					Signature: "0x" + strings.Repeat("zz", 65),
				}
			},
			setupBridgeMock: func(mockBridge *MockMezoBridge) {},
			setupStoreMock:  func(mockStore *MockStore) {},
		},
		{
			name:           "Wrong signature length",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid signature length",
			setupRequest: func() *types.SubmitAttestationRequest {
				return &types.SubmitAttestationRequest{
					Entry:     createValidEntry(),
					Signature: "0x" + strings.Repeat("00", 32), // Too short
				}
			},
			setupBridgeMock: func(mockBridge *MockMezoBridge) {},
			setupStoreMock:  func(mockStore *MockStore) {},
		},
		{
			name:           "Missing sequence number",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid assets unlocked entry",
			setupRequest: func() *types.SubmitAttestationRequest {
				entry := createValidEntry()
				entry.UnlockSequence = sdkmath.Int{}
				signature := "0x" + strings.Repeat("00", 65)
				return &types.SubmitAttestationRequest{
					Entry:     entry,
					Signature: signature,
				}
			},
			setupBridgeMock: func(mockBridge *MockMezoBridge) {},
			setupStoreMock:  func(mockStore *MockStore) {},
		},
		{
			name:           "Invalid sequence number (zero)",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid assets unlocked entry",
			setupRequest: func() *types.SubmitAttestationRequest {
				entry := createValidEntry()
				entry.UnlockSequence = sdkmath.NewInt(0)
				signature := "0x" + strings.Repeat("00", 65)
				return &types.SubmitAttestationRequest{
					Entry:     entry,
					Signature: signature,
				}
			},
			setupBridgeMock: func(mockBridge *MockMezoBridge) {},
			setupStoreMock:  func(mockStore *MockStore) {},
		},
		{
			name:           "Invalid sequence number (negative)",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid assets unlocked entry",
			setupRequest: func() *types.SubmitAttestationRequest {
				entry := createValidEntry()
				entry.UnlockSequence = sdkmath.NewInt(-1)
				signature := "0x" + strings.Repeat("00", 65)
				return &types.SubmitAttestationRequest{
					Entry:     entry,
					Signature: signature,
				}
			},
			setupBridgeMock: func(mockBridge *MockMezoBridge) {},
			setupStoreMock:  func(mockStore *MockStore) {},
		},
		{
			name:           "Invalid recipient (empty)",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid assets unlocked entry",
			setupRequest: func() *types.SubmitAttestationRequest {
				entry := createValidEntry()
				entry.Recipient = []byte{}
				signature := "0x" + strings.Repeat("00", 65)
				return &types.SubmitAttestationRequest{
					Entry:     entry,
					Signature: signature,
				}
			},
			setupBridgeMock: func(mockBridge *MockMezoBridge) {},
			setupStoreMock:  func(mockStore *MockStore) {},
		},
		{
			name:           "Valid entry but invalid signature for recovery",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "recovery failed",
			setupRequest: func() *types.SubmitAttestationRequest {
				entry := createValidEntry()
				// Zero address is actually valid for the token field validation
				// This test will pass validation but fail at signature recovery
				signature := "0x" + strings.Repeat("00", 65)
				return &types.SubmitAttestationRequest{
					Entry:     entry,
					Signature: signature,
				}
			},
			setupBridgeMock: func(mockBridge *MockMezoBridge) {},
			setupStoreMock:  func(mockStore *MockStore) {},
		},
		{
			name:           "Missing amount",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid assets unlocked entry",
			setupRequest: func() *types.SubmitAttestationRequest {
				entry := createValidEntry()
				entry.Amount = sdkmath.Int{}
				signature := "0x" + strings.Repeat("00", 65)
				return &types.SubmitAttestationRequest{
					Entry:     entry,
					Signature: signature,
				}
			},
			setupBridgeMock: func(mockBridge *MockMezoBridge) {},
			setupStoreMock:  func(mockStore *MockStore) {},
		},
		{
			name:           "Invalid amount (zero)",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid assets unlocked entry",
			setupRequest: func() *types.SubmitAttestationRequest {
				entry := createValidEntry()
				entry.Amount = sdkmath.NewInt(0)
				signature := "0x" + strings.Repeat("00", 65)
				return &types.SubmitAttestationRequest{
					Entry:     entry,
					Signature: signature,
				}
			},
			setupBridgeMock: func(mockBridge *MockMezoBridge) {},
			setupStoreMock:  func(mockStore *MockStore) {},
		},
		{
			name:           "Invalid amount (negative)",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid assets unlocked entry",
			setupRequest: func() *types.SubmitAttestationRequest {
				entry := createValidEntry()
				entry.Amount = sdkmath.NewInt(-1)
				signature := "0x" + strings.Repeat("00", 65)
				return &types.SubmitAttestationRequest{
					Entry:     entry,
					Signature: signature,
				}
			},
			setupBridgeMock: func(mockBridge *MockMezoBridge) {},
			setupStoreMock:  func(mockStore *MockStore) {},
		},
		{
			name:           "Invalid chain (out of range)",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid assets unlocked entry",
			setupRequest: func() *types.SubmitAttestationRequest {
				entry := createValidEntry()
				entry.Chain = 99 // Invalid chain value
				signature := "0x" + strings.Repeat("00", 65)
				return &types.SubmitAttestationRequest{
					Entry:     entry,
					Signature: signature,
				}
			},
			setupBridgeMock: func(mockBridge *MockMezoBridge) {},
			setupStoreMock:  func(mockStore *MockStore) {},
		},
		// New test cases for the added validations
		{
			name:           "Unauthorized validator",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "not an authorized validator",
			setupRequest: func() *types.SubmitAttestationRequest {
				entry := createValidEntry()
				signature := createValidSignature(entry, privateKey)
				return &types.SubmitAttestationRequest{
					Entry:     entry,
					Signature: signature,
				}
			},
			setupBridgeMock: func(mockBridge *MockMezoBridge) {
				// Return 0 for validator ID to simulate unauthorized validator
				mockBridge.EXPECT().ValidatorIDs(validatorAddress).Return(uint8(0), nil)
			},
			setupStoreMock: func(mockStore *MockStore) {},
		},
		{
			name:           "Validator ID lookup error",
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "validator lookup error",
			setupRequest: func() *types.SubmitAttestationRequest {
				entry := createValidEntry()
				signature := createValidSignature(entry, privateKey)
				return &types.SubmitAttestationRequest{
					Entry:     entry,
					Signature: signature,
				}
			},
			setupBridgeMock: func(mockBridge *MockMezoBridge) {
				// Return error when looking up validator ID
				mockBridge.EXPECT().ValidatorIDs(validatorAddress).Return(uint8(0), errors.New("validator lookup error"))
			},
			setupStoreMock: func(mockStore *MockStore) {},
		},
		{
			name:           "Already confirmed unlock",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "already confirmed",
			setupRequest: func() *types.SubmitAttestationRequest {
				entry := createValidEntry()
				signature := createValidSignature(entry, privateKey)
				return &types.SubmitAttestationRequest{
					Entry:     entry,
					Signature: signature,
				}
			},
			setupBridgeMock: func(mockBridge *MockMezoBridge) {
				mockBridge.EXPECT().ValidatorIDs(validatorAddress).Return(uint8(1), nil)
				// Return true to simulate already confirmed unlock
				mockBridge.EXPECT().ConfirmedUnlocks(big.NewInt(1)).Return(true, nil)
			},
			setupStoreMock: func(mockStore *MockStore) {},
		},
		{
			name:           "Confirmed unlocks lookup error",
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "unlock check error",
			setupRequest: func() *types.SubmitAttestationRequest {
				entry := createValidEntry()
				signature := createValidSignature(entry, privateKey)
				return &types.SubmitAttestationRequest{
					Entry:     entry,
					Signature: signature,
				}
			},
			setupBridgeMock: func(mockBridge *MockMezoBridge) {
				mockBridge.EXPECT().ValidatorIDs(validatorAddress).Return(uint8(1), nil)
				// Return error when checking confirmed unlocks
				mockBridge.EXPECT().ConfirmedUnlocks(big.NewInt(1)).Return(false, errors.New("unlock check error"))
			},
			setupStoreMock: func(mockStore *MockStore) {},
		},
		{
			name:           "Invalid assets unlocked validation",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "not a valide asset unlocked event",
			setupRequest: func() *types.SubmitAttestationRequest {
				entry := createValidEntry()
				signature := createValidSignature(entry, privateKey)
				return &types.SubmitAttestationRequest{
					Entry:     entry,
					Signature: signature,
				}
			},
			setupBridgeMock: func(mockBridge *MockMezoBridge) {
				mockBridge.EXPECT().ValidatorIDs(validatorAddress).Return(uint8(1), nil)
				mockBridge.EXPECT().ConfirmedUnlocks(big.NewInt(1)).Return(false, nil)
				// Return false to simulate invalid assets unlocked
				mockBridge.EXPECT().ValidateAssetsUnlocked(gomock.Any()).Return(false, nil)
			},
			setupStoreMock: func(mockStore *MockStore) {},
		},
		{
			name:           "Assets unlocked validation error",
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "validation error",
			setupRequest: func() *types.SubmitAttestationRequest {
				entry := createValidEntry()
				signature := createValidSignature(entry, privateKey)
				return &types.SubmitAttestationRequest{
					Entry:     entry,
					Signature: signature,
				}
			},
			setupBridgeMock: func(mockBridge *MockMezoBridge) {
				mockBridge.EXPECT().ValidatorIDs(validatorAddress).Return(uint8(1), nil)
				mockBridge.EXPECT().ConfirmedUnlocks(big.NewInt(1)).Return(false, nil)
				// Return error during validation
				mockBridge.EXPECT().ValidateAssetsUnlocked(gomock.Any()).Return(false, errors.New("validation error"))
			},
			setupStoreMock: func(mockStore *MockStore) {},
		},
		// Store operation test cases
		{
			name:           "Store SaveAttestation failure",
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "save attestation error",
			setupRequest: func() *types.SubmitAttestationRequest {
				entry := createValidEntry()
				signature := createValidSignature(entry, privateKey)
				return &types.SubmitAttestationRequest{
					Entry:     entry,
					Signature: signature,
				}
			},
			setupBridgeMock: func(mockBridge *MockMezoBridge) {
				mockBridge.EXPECT().ValidatorIDs(validatorAddress).Return(uint8(1), nil)
				mockBridge.EXPECT().ConfirmedUnlocks(big.NewInt(1)).Return(false, nil)
				mockBridge.EXPECT().ValidateAssetsUnlocked(gomock.Any()).Return(true, nil)
			},
			setupStoreMock: func(mockStore *MockStore) {
				mockStore.EXPECT().SaveAttestation(gomock.Any()).Return(errors.New("save attestation error"))
			},
		},
		{
			name:           "Store SaveSignature failure",
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "save signature error",
			setupRequest: func() *types.SubmitAttestationRequest {
				entry := createValidEntry()
				signature := createValidSignature(entry, privateKey)
				return &types.SubmitAttestationRequest{
					Entry:     entry,
					Signature: signature,
				}
			},
			setupBridgeMock: func(mockBridge *MockMezoBridge) {
				mockBridge.EXPECT().ValidatorIDs(validatorAddress).Return(uint8(1), nil)
				mockBridge.EXPECT().ConfirmedUnlocks(big.NewInt(1)).Return(false, nil)
				mockBridge.EXPECT().ValidateAssetsUnlocked(gomock.Any()).Return(true, nil)
			},
			setupStoreMock: func(mockStore *MockStore) {
				mockStore.EXPECT().SaveAttestation(gomock.Any()).Return(nil)
				mockStore.EXPECT().SaveSignature(gomock.Any(), gomock.Any()).Return(errors.New("save signature error"))
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// Setup mocks
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockBridge := NewMockMezoBridge(ctrl)
			mockStore := NewMockStore(ctrl)
			testCase.setupBridgeMock(mockBridge)
			testCase.setupStoreMock(mockStore)

			// Create server with mocks
			server := NewServer(log.NewNopLogger(), 8080, chainID, mockBridge, mockStore)

			var requestBody []byte
			var err error

			if testCase.setupRequest != nil {
				req := testCase.setupRequest()
				requestBody, err = json.Marshal(req)
				require.NoError(t, err)
			} else if testCase.requestBody != nil {
				if str, ok := testCase.requestBody.(string); ok {
					requestBody = []byte(str)
				} else {
					requestBody, err = json.Marshal(testCase.requestBody)
					require.NoError(t, err)
				}
			}

			req := httptest.NewRequest(http.MethodPost, "/attestations", bytes.NewReader(requestBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()

			server.submitAttestation(w, req)

			assert.Equal(t, testCase.expectedStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var response types.SubmitAttestationResponse
			err = json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			// Check response content
			if testCase.expectedStatus == http.StatusAccepted {
				assert.True(t, response.Success)
				assert.Empty(t, response.Error)
			} else {
				assert.Equal(t, w.Code, testCase.expectedStatus)
				assert.False(t, response.Success)
				assert.Contains(t, response.Error, testCase.expectedError)
			}
		})
	}
}

func signAssetUnlock(
	t *testing.T,
	entry *bridgetypes.AssetsUnlockedEvent,
	signer *ecdsa.PrivateKey,
) string {
	t.Helper()

	chainID := big.NewInt(1)

	attestation := &portal.MezoBridgeAssetsUnlocked{
		UnlockSequenceNumber: entry.UnlockSequence.BigInt(),
		Recipient:            entry.Recipient,
		Token:                common.HexToAddress(entry.Token),
		Amount:               entry.Amount.BigInt(),
		Chain:                uint8(entry.Chain),
	}

	hash, err := portal.AttestationDigestHash(attestation, chainID)
	require.NoError(t, err)

	signature, err := crypto.Sign(hash, signer)
	require.NoError(t, err)

	return hexutil.Encode(signature)
}

func TestServer_recoverAddress(t *testing.T) {
	chainID := big.NewInt(1)

	// Setup mocks (not needed for this test, but required for constructor)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockBridge := NewMockMezoBridge(ctrl)
	mockStore := NewMockStore(ctrl)

	server := NewServer(log.NewNopLogger(), 8080, chainID, mockBridge, mockStore)

	privateKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	expectedAddress := crypto.PubkeyToAddress(privateKey.PublicKey)

	privateKey2, err := crypto.GenerateKey()
	require.NoError(t, err)
	expectedAddress2 := crypto.PubkeyToAddress(privateKey2.PublicKey)

	createValidEntry := func() *bridgetypes.AssetsUnlockedEvent {
		return &bridgetypes.AssetsUnlockedEvent{
			UnlockSequence: sdkmath.NewInt(1),
			Recipient:      []byte("test-recipient"),
			Token:          "0x1234567890123456789012345678901234567890",
			Sender:         "0x9876543210987654321098765432109876543210",
			Amount:         sdkmath.NewInt(1000),
			Chain:          0,
			BlockTime:      1000,
		}
	}

	testCases := []struct {
		name            string
		entry           *bridgetypes.AssetsUnlockedEvent
		signature       string
		expectedAddress *common.Address
		expectError     bool
		expectedError   string
		setupSignature  func() string
	}{
		{
			name:            "Valid signature recovery",
			entry:           createValidEntry(),
			expectedAddress: &expectedAddress,
			expectError:     false,
			setupSignature: func() string {
				return signAssetUnlock(t, createValidEntry(), privateKey)
			},
		},
		{
			name:            "Valid signature recovery with different key",
			entry:           createValidEntry(),
			expectedAddress: &expectedAddress2,
			expectError:     false,
			setupSignature: func() string {
				return signAssetUnlock(t, createValidEntry(), privateKey2)
			},
		},
		{
			name: "Altered entry results in different recovered address",
			setupSignature: func() string {
				return signAssetUnlock(t, createValidEntry(), privateKey)
			},
			entry: func() *bridgetypes.AssetsUnlockedEvent {
				// But try to recover with modified entry
				modifiedEntry := createValidEntry()
				modifiedEntry.Amount = sdkmath.NewInt(2000) // Modified amount
				return modifiedEntry
			}(),
			expectError: false,
			// We don't specify expectedAddress because we expect it to be
			// different from the original signer
		},
		// then test signature format validations
		{
			name:          "Invalid signature - missing 0x prefix",
			entry:         createValidEntry(),
			signature:     strings.Repeat("00", 65),
			expectError:   true,
			expectedError: "hex string without 0x prefix",
		},
		{
			name:          "Invalid signature - empty string",
			entry:         createValidEntry(),
			signature:     "",
			expectError:   true,
			expectedError: "empty hex string",
		},
		{
			name:          "Invalid signature - invalid hex",
			entry:         createValidEntry(),
			signature:     "0x" + strings.Repeat("zz", 65),
			expectError:   true,
			expectedError: "invalid hex string",
		},
		{
			name:          "Invalid signature - wrong length (too short)",
			entry:         createValidEntry(),
			signature:     "0x" + strings.Repeat("00", 32),
			expectError:   true,
			expectedError: "invalid signature length",
		},
		{
			name:          "Invalid signature - wrong length (too long)",
			entry:         createValidEntry(),
			signature:     "0x" + strings.Repeat("00", 100),
			expectError:   true,
			expectedError: "invalid signature length",
		},
		{
			name:          "Invalid signature - all zeros (recovery fails)",
			entry:         createValidEntry(),
			signature:     "0x" + strings.Repeat("00", 65),
			expectError:   true,
			expectedError: "recovery failed",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var signature string
			if testCase.setupSignature != nil {
				signature = testCase.setupSignature()
			} else {
				signature = testCase.signature
			}

			recoveredAddress, err := server.recoverAddress(testCase.entry, signature)

			if testCase.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)
				assert.Equal(t, common.Address{}, recoveredAddress)
			} else {

				// no error, so we either have a valid address recovered
				require.NoError(t, err)
				if testCase.expectedAddress != nil {
					assert.Equal(t, *testCase.expectedAddress, recoveredAddress)
				} else { // or a different address which is due to the parameter swap
					assert.NotEqual(t, common.Address{}, recoveredAddress)
					assert.NotEqual(t, expectedAddress, recoveredAddress)
				}
			}
		})
	}
}
