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
	"cosmossdk.io/math"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/mezo-org/mezod/bridge-worker/types"
	"github.com/mezo-org/mezod/ethereum/bindings/portal"
	bridgetypes "github.com/mezo-org/mezod/x/bridge/types"
)

type MezoBridge interface {
	BridgeValidatorIDs(common.Address) (uint8, error)
	ValidateAssetsUnlocked(portal.MezoBridgeAssetsUnlocked) (bool, error)
	ConfirmedUnlocks(*big.Int) (bool, error)
}

type Store interface {
	SaveAttestation(entry *bridgetypes.AssetsUnlockedEvent) error
	SaveSignature(unlockSequence math.Int, sig string) error
}

type Server struct {
	logger     log.Logger
	server     *http.Server
	chainID    *big.Int
	mezoBridge MezoBridge
	store      Store
}

func NewServer(
	logger log.Logger,
	port uint16,
	chainID *big.Int,
	mezoBridge MezoBridge,
	store Store,
) *Server {
	s := &Server{
		logger:     logger,
		chainID:    chainID,
		mezoBridge: mezoBridge,
		store:      store,
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
		s.writeSubmitAttestationError(w, err, http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		s.writeSubmitAttestationError(w, err, http.StatusBadRequest)
		return
	}

	// TODO: ensure this is a valid attestation

	// first recover the address out of the signature
	address, err := s.recoverAddress(req.Entry, req.Signature)
	if err != nil {
		s.writeSubmitAttestationError(w, err, http.StatusBadRequest)
		return
	}

	// now validate the address is a registered validator address
	index, err := s.mezoBridge.BridgeValidatorIDs(address)
	if err != nil {
		s.logger.Error("couldn't get bridge validator ID", "error", err)
		s.writeSubmitAttestationError(w, err, http.StatusInternalServerError)
		return
	}

	// if default value, then address is not a validator
	if index == 0 {
		s.writeSubmitAttestationError(w, errors.New("not an authorized validator"), http.StatusUnauthorized)
		return
	}

	// then the attestation has not been confirmed yet
	isConfirmed, err := s.mezoBridge.ConfirmedUnlocks(req.Entry.UnlockSequence.BigInt())
	if err != nil {
		s.logger.Error("couldn't confirm unlock", "error", err)
		s.writeSubmitAttestationError(w, err, http.StatusInternalServerError)
		return
	}

	// if already confirmed, nothing to do
	if isConfirmed {
		s.writeSubmitAttestationError(w, errors.New("already confirmed"), http.StatusBadRequest)
		return
	}

	// finally, is it a valid attestation
	attestation := toPortalAssetsUnlock(req.Entry)
	ok, err := s.mezoBridge.ValidateAssetsUnlocked(*attestation)
	if err != nil {
		s.logger.Error("couldn't validate assets unlock event", "error", err)
		s.writeSubmitAttestationError(w, err, http.StatusInternalServerError)
		return
	}

	// if already confirmed, nothing to do
	if !ok {
		s.writeSubmitAttestationError(w, errors.New("not a valide asset unlocked event"), http.StatusBadRequest)
		return
	}

	// now we save it into the store
	if err := s.store.SaveAttestation(req.Entry); err != nil {
		s.logger.Error("couldn't save assets unlock event", "error", err)
		s.writeSubmitAttestationError(w, err, http.StatusInternalServerError)
		return
	}

	if err := s.store.SaveSignature(req.Entry.UnlockSequence, req.Signature); err != nil {
		s.logger.Error("couldn't save signature", "error", err)
		s.writeSubmitAttestationError(w, err, http.StatusInternalServerError)
		return
	}

	s.writeSubmitAttestationSuccess(w, http.StatusAccepted)
}

func (s *Server) writeSubmitAttestationError(w http.ResponseWriter, err error, status int) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(types.SubmitAttestationResponse{
		Error:   err.Error(),
		Success: false,
	})
}

func (s *Server) writeSubmitAttestationSuccess(w http.ResponseWriter, status int) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(types.SubmitAttestationResponse{
		Success: true,
	})
}

func (s *Server) recoverAddress(entry *bridgetypes.AssetsUnlockedEvent, signature string) (common.Address, error) {
	attestation := toPortalAssetsUnlock(entry)

	hash, err := portal.AttestationDigestHash(attestation, s.chainID)
	if err != nil {
		return common.Address{}, err
	}

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

func toPortalAssetsUnlock(entry *bridgetypes.AssetsUnlockedEvent) *portal.MezoBridgeAssetsUnlocked {
	return &portal.MezoBridgeAssetsUnlocked{
		UnlockSequenceNumber: entry.UnlockSequence.BigInt(),
		Recipient:            entry.Recipient,
		Token:                common.HexToAddress(entry.Token),
		Amount:               entry.Amount.BigInt(),
		Chain:                uint8(entry.Chain),
	}
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
