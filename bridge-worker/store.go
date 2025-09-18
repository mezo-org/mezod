package bridgeworker

import (
	"fmt"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	bridgetypes "github.com/mezo-org/mezod/x/bridge/types"
	"github.com/supabase-community/supabase-go"
)

const (
	AssetsUnlockedEventsTable  = "assets_unlocked_events"
	AttestationSignaturesTable = "attestations_signatures"
)

type SupabaseStore struct {
	logger log.Logger
	client *supabase.Client
}

func NewSupabaseStore(
	logger log.Logger,
	databaseURL string,
	anonKey string,
) (*SupabaseStore, error) {
	client, err := supabase.NewClient(databaseURL, anonKey, &supabase.ClientOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create supabase client: %w", err)
	}

	return &SupabaseStore{
		logger: logger,
		client: client,
	}, nil
}

// SaveAttestation will save an attestation in the database
// because there's no good atomic way of checking the error that
// occurred, we take an optimistic approach, meaning:
// we first try to insert
// if we have an error we then try to get the entry
// if we are successful at loading the entry, then we return no error
// if we are not then we return the initial error
func (s *SupabaseStore) SaveAttestation(entry *bridgetypes.AssetsUnlockedEvent) error {
	_, _, err := s.client.From(AssetsUnlockedEventsTable).Insert(entry, false, "", "", "").Execute()
	if err != nil {
		// we've had an error, let's try to load it as well
		// if it succeed that means it was inserted before
		attestation, loadErr := s.LoadAttestation(entry.UnlockSequence)
		if loadErr != nil {
			return fmt.Errorf("couldn't save attestation: %w, %w", err, loadErr)
		}
		// attestation is nil, that means it didn't exists, and we could not
		// create a new one, that's some internal error as well
		if attestation == nil {
			return fmt.Errorf("couldn't create attestation: %w", err)
		}

		// finally attestation is not nil, that means we couldn't create
		// it because it already existed, nothing to do.
	}

	return nil
}

func (s *SupabaseStore) LoadAttestation(unlockSequence math.Int) (*bridgetypes.AssetsUnlockedEvent, error) {
	attestations := []bridgetypes.AssetsUnlockedEvent{}
	// first is the count, there' can be only 1 anyway
	// because the sequence number is unique in DB and with
	// filter on it, if it's 0 then we return nil, meaning it doesn't
	// exists
	_, err := s.client.From(AssetsUnlockedEventsTable).
		Select("*", "", false).
		Eq("unlock_sequence", unlockSequence.String()).
		ExecuteTo(&attestations)
	if err != nil {
		return nil, fmt.Errorf("couldn't load attestation: %w", err)
	}

	if len(attestations) == 0 {
		return nil, nil
	}

	return &attestations[0], nil
}

type signature struct {
	UnlockSequence math.Int `json:"unlock_sequence"`
	Signature      string   `json:"signature"`
}

func (s *SupabaseStore) SaveSignature(unlockSequence math.Int, sig string) error {
	signature := signature{
		UnlockSequence: unlockSequence,
		Signature:      sig,
	}
	_, _, err := s.client.From(AttestationSignaturesTable).Insert(signature, false, "", "", "").Execute()
	if err != nil {
		return fmt.Errorf("couldn't save signature: %w", err)
	}

	return nil
}

func (s *SupabaseStore) LoadSignature(unlockSequence math.Int) ([]string, error) {
	rawSignatures := []signature{}
	_, err := s.client.From(AttestationSignaturesTable).
		Select("*", "", false).
		Eq("unlock_sequence", unlockSequence.String()).
		ExecuteTo(&rawSignatures)
	if err != nil {
		return nil, fmt.Errorf("couldn't load signatures: %w", err)
	}

	signatures := make([]string, 0, len(rawSignatures))
	for _, s := range rawSignatures {
		signatures = append(signatures, s.Signature)
	}

	return signatures, nil
}
