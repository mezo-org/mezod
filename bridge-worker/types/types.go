package types

import (
	"encoding/hex"
	"errors"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/mezo-org/mezod/ethereum/bindings/portal"
)

const (
	chainEthereum = iota
	chainBitcoin
)

var (
	ErrInvalidSignatureFormat     = errors.New("invalid signature format")
	ErrMissingAssetsUnlockedEntry = errors.New("missing assets unlocked entry")
	ErrMissingSequenceNumber      = errors.New("missing sequence number")
	ErrInvalidSequenceNumber      = errors.New("invalid sequence number")
	ErrInvalidRecipient           = errors.New("invalid recipient")
	ErrInvalidToken               = errors.New("invalid token")
	ErrMissingAmount              = errors.New("missing amount")
	ErrInvalidAmount              = errors.New("invalid amount")
	ErrInvalidChain               = errors.New("invalid chain")

	zero = big.NewInt(0)
)

type AssetsUnlocked struct {
	UnlockSequenceNumber *big.Int       `json:"unlock_sequence_number"`
	Recipient            []byte         `json:"recipient"`
	Token                common.Address `json:"token"`
	Amount               *big.Int       `json:"amount"`
	Chain                uint8          `json:"chain"`
}

func AssetsUnlockedFromPortal(assetsUnlocked *portal.MezoBridgeAssetsUnlocked) *AssetsUnlocked {
	return &AssetsUnlocked{
		UnlockSequenceNumber: assetsUnlocked.UnlockSequenceNumber,
		Recipient:            assetsUnlocked.Recipient,
		Token:                assetsUnlocked.Token,
		Amount:               assetsUnlocked.Amount,
		Chain:                assetsUnlocked.Chain,
	}
}

func (a *AssetsUnlocked) Validate() error {
	if a.UnlockSequenceNumber == nil {
		return ErrMissingSequenceNumber
	}
	if a.UnlockSequenceNumber.Cmp(zero) <= 0 {
		return ErrInvalidSequenceNumber
	}

	if len(a.Recipient) == 0 {
		return ErrInvalidRecipient
	}

	if len(a.Token) == 0 {
		return ErrInvalidToken
	}

	if a.Amount == nil {
		return ErrMissingAmount
	}
	if a.Amount.Cmp(zero) <= 0 {
		return ErrInvalidAmount
	}

	if a.Chain < chainEthereum || a.Chain > chainBitcoin {
		return ErrInvalidChain
	}

	return nil
}

type SubmitAttestationRequest struct {
	Entry     *AssetsUnlocked
	Signature string
}

func (s *SubmitAttestationRequest) Validate() error {
	if s.Entry == nil {
		return ErrMissingAssetsUnlockedEntry
	}

	if err := validateSignature(s.Signature); err != nil {
		return err
	}

	return s.Entry.Validate()
}

type SubmitAttestationResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

func validateSignature(signature string) error {
	if !strings.HasPrefix(signature, "0x") {
		return ErrInvalidSignatureFormat
	}

	signature = signature[2:]

	bytes, err := hex.DecodeString(signature)
	if err != nil {
		return ErrInvalidSignatureFormat
	}

	if len(bytes) != 65 {
		return ErrInvalidSignatureFormat
	}

	return nil
}
