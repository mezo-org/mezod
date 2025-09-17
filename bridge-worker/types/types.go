package types

import (
	"errors"

	"github.com/ethereum/go-ethereum/common/hexutil"
	bridgetypes "github.com/mezo-org/mezod/x/bridge/types"
)

var (
	ErrInvalidSignatureFormat     = errors.New("invalid signature format")
	ErrMissingAssetsUnlockedEntry = errors.New("missing assets unlocked entry")
	ErrInvalidAssetsUnlockedEntry = errors.New("invalid assets unlocked entry")
)

type SubmitAttestationRequest struct {
	Entry     *bridgetypes.AssetsUnlockedEvent
	Signature string
}

type SubmitAttestationResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

func (s *SubmitAttestationRequest) Validate() error {
	if s.Entry == nil {
		return ErrMissingAssetsUnlockedEntry
	}

	if _, err := hexutil.Decode(s.Signature); err != nil {
		return err
	}

	if !s.Entry.IsValid() {
		return ErrInvalidAssetsUnlockedEntry
	}

	return nil
}
