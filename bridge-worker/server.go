package bridgeworker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"time"

	"cosmossdk.io/log"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/mezo-org/mezod/bridge-worker/types"
)

type Server struct {
	logger  log.Logger
	server  *http.Server
	chainID *big.Int
}

func NewServer(
	logger log.Logger,
	port uint16,
	chainID *big.Int,
) *Server {
	s := &Server{
		logger:  logger,
		chainID: chainID,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /attestations", s.submitAttestation)

	s.server = &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           mux,
		ReadHeaderTimeout: time.Second,
	}

	return s
}

func (s *Server) Start() {
	s.logger.Info("http server starting", "address", s.server.Addr)
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			// should we panic here? likely yes
			s.logger.Error("server failed to start", "error", err)
		}
	}()
}

func (s *Server) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func (s *Server) submitAttestation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	req, err := readSubmitAttestationRequest(r)
	if err != nil {
		writeError(w, err, http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		writeError(w, err, http.StatusBadRequest)
		return
	}

	address, err := s.recoverAddress(req.Entry, req.Signature)
	if err != nil {
		writeError(w, err, http.StatusBadRequest)
		return
	}

	// now do call some other service which will
	// process the asset unlock if the address is a
	// valid signer
	_ = address

	writeSuccess(w, http.StatusAccepted)
}

func (s *Server) recoverAddress(entry *types.AssetsUnlocked, signature string) (common.Address, error) {
	abiEncoded, err := abiEncodeAttestationWithChainID(entry, s.chainID)
	if err != nil {
		return common.Address{}, err
	}

	// Apply the same TextHash transformation
	hash := accounts.TextHash(abiEncoded)

	// Decode the signature from hex
	signatureBytes, err := hexutil.Decode(signature)
	if err != nil {
		return common.Address{}, err
	}

	// Recover the public key from the signature
	publicKeyBytes, err := crypto.SigToPub(hash, signatureBytes)
	if err != nil {
		return common.Address{}, err
	}

	// Convert public key to address
	return crypto.PubkeyToAddress(*publicKeyBytes), nil
}

func readSubmitAttestationRequest(r *http.Request) (*types.SubmitAttestationRequest, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, errors.New("failed to read request body")
	}
	defer r.Body.Close()

	req := types.SubmitAttestationRequest{}
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, errors.New("invalid json format")
	}

	return &req, nil
}

func writeError(w http.ResponseWriter, err error, status int) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(types.SubmitAttestationResponse{
		Error:   err.Error(),
		Success: false,
	})
}

func writeSuccess(w http.ResponseWriter, status int) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(types.SubmitAttestationResponse{
		Success: true,
	})
}

// abiEncodeAttestationWithChainID is used to encode the attestation with the chain ID
// which is used to produce a signature for the batch attestation process.
func abiEncodeAttestationWithChainID(attestation *types.AssetsUnlocked, chainID *big.Int) ([]byte, error) {
	var argumentsTypes abi.Arguments
	var arguments []any

	if chainID != nil {
		uint256Type, err := abi.NewType("uint256", "uint256", nil)
		if err != nil {
			return nil, err
		}
		argumentsTypes = append(argumentsTypes, abi.Argument{Type: uint256Type})
		arguments = append(arguments, chainID)
	}

	// Create tuple type for AssetsUnlocked struct
	tupleType, err := abi.NewType("tuple", "tuple", []abi.ArgumentMarshaling{
		{Name: "unlockSequenceNumber", Type: "uint256"},
		{Name: "recipient", Type: "bytes"},
		{Name: "token", Type: "address"},
		{Name: "amount", Type: "uint256"},
		{Name: "chain", Type: "uint8"},
	})
	if err != nil {
		return nil, err
	}

	// Add the tuple as a single argument instead of individual fields
	argumentsTypes = append(argumentsTypes, abi.Argument{Type: tupleType})

	// Create the struct as a single tuple argument
	assetsUnlockedTuple := struct {
		UnlockSequenceNumber *big.Int
		Recipient            []byte
		Token                common.Address
		Amount               *big.Int
		Chain                uint8
	}{
		UnlockSequenceNumber: attestation.UnlockSequenceNumber,
		Recipient:            attestation.Recipient,
		Token:                attestation.Token,
		Amount:               attestation.Amount,
		Chain:                attestation.Chain,
	}

	arguments = append(arguments, assetsUnlockedTuple)

	return argumentsTypes.Pack(arguments...)
}
