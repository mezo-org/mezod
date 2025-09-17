package bridgeworker

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/mezo-org/mezod/bridge-worker/types"
	bridgetypes "github.com/mezo-org/mezod/x/bridge/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServer_submitSignature(t *testing.T) {
	chainID := big.NewInt(1)
	server := NewServer(log.NewNopLogger(), 8080, chainID)

	// Generate a test private key for creating valid signatures
	privateKey, err := crypto.GenerateKey()
	require.NoError(t, err)

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
		abiEncoded, err := abiEncodeAttestationWithChainID(entry, chainID)
		require.NoError(t, err)

		hash := accounts.TextHash(abiEncoded)
		signature, err := crypto.Sign(hash, privKey)
		require.NoError(t, err)

		return hexutil.Encode(signature)
	}

	testCases := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		expectedError  string
		setupRequest   func() *types.SubmitAttestationRequest
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
		},
		{
			name:           "Invalid JSON body",
			requestBody:    "invalid-json",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid json format",
		},
		{
			name:           "Empty request body",
			requestBody:    "",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid json format",
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
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
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

			req := httptest.NewRequest(http.MethodPost, "/submit-signature", bytes.NewReader(requestBody))
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

func TestServer_recoverAddress(t *testing.T) {
	chainID := big.NewInt(1)
	server := NewServer(log.NewNopLogger(), 8080, chainID)

	privateKey, err := crypto.GenerateKey()
	require.NoError(t, err)

	expectedAddress := crypto.PubkeyToAddress(privateKey.PublicKey)

	entry := &bridgetypes.AssetsUnlockedEvent{
		UnlockSequence: sdkmath.NewInt(1),
		Recipient:      []byte("test-recipient"), // doesn't matter here
		Token:          "0x1234567890123456789012345678901234567890",
		Sender:         "0x9876543210987654321098765432109876543210",
		Amount:         sdkmath.NewInt(1000),
		Chain:          0,
		BlockTime:      1000,
	}

	abiEncoded, err := abiEncodeAttestationWithChainID(entry, chainID)
	require.NoError(t, err)

	hash := accounts.TextHash(abiEncoded)
	signature, err := crypto.Sign(hash, privateKey)
	require.NoError(t, err)

	signatureHex := hexutil.Encode(signature)

	recoveredAddress, err := server.recoverAddress(entry, signatureHex)
	require.NoError(t, err)
	assert.Equal(t, expectedAddress, recoveredAddress)

	_, err = server.recoverAddress(entry, "invalid-signature")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "hex string without 0x prefix")

	invalidSignature := "0x" + strings.Repeat("00", 65)
	_, err = server.recoverAddress(entry, invalidSignature)
	assert.Error(t, err)
}
