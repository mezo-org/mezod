package bridgeworker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"

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
	mux.HandleFunc("POST /submit-signature", s.submitSignature)

	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
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

func (s *Server) submitSignature(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	req, err := readSubmitSignatureRequest(r)
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

func readSubmitSignatureRequest(r *http.Request) (*types.SubmitSignatureRequest, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, errors.New("failed to read request body")
	}
	defer r.Body.Close()

	req := types.SubmitSignatureRequest{}
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, errors.New("invalid json format")
	}

	return &req, nil
}

func writeError(w http.ResponseWriter, err error, status int) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(types.SubmitSignatureResponse{
		Error:   err.Error(),
		Success: false,
	})
}

func writeSuccess(w http.ResponseWriter, status int) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(types.SubmitSignatureResponse{
		Success: true,
	})
}

// abiEncodeAttestationWithChainID is used to encode the attestation with the chain ID
// which is used to produce a signature for the batch attestation process.
func abiEncodeAttestationWithChainID(attestation *types.AssetsUnlocked, chainID *big.Int) ([]byte, error) {
	uint256Type, err := abi.NewType("uint256", "uint256", nil)
	if err != nil {
		return nil, err
	}
	bytesType, err := abi.NewType("bytes", "bytes", nil)
	if err != nil {
		return nil, err
	}
	addressType, err := abi.NewType("address", "address", nil)
	if err != nil {
		return nil, err
	}
	uint8Type, err := abi.NewType("uint8", "uint8", nil)
	if err != nil {
		return nil, err
	}

	var argumentsTypes abi.Arguments
	var arguments []any
	if chainID != nil {
		argumentsTypes = append(argumentsTypes, abi.Argument{Type: uint256Type})
		arguments = append(arguments, chainID)
	}

	argumentsTypes = append(
		argumentsTypes,
		abi.Arguments{
			{Type: uint256Type},
			{Type: bytesType},
			{Type: addressType},
			{Type: uint256Type},
			{Type: uint8Type},
		}...,
	)

	arguments = append(
		arguments,
		[]any{
			attestation.UnlockSequenceNumber,
			attestation.Recipient,
			attestation.Token,
			attestation.Amount,
			attestation.Chain,
		}...,
	)

	return argumentsTypes.Pack(arguments...)
}
