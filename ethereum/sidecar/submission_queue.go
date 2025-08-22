package sidecar

import (
	"fmt"
	"math/big"
	"math/rand"
	"slices"
	"time"

	"cosmossdk.io/log"
	"github.com/ethereum/go-ethereum/common"
	ethconnect "github.com/mezo-org/mezod/ethereum"
	"github.com/mezo-org/mezod/ethereum/bindings/portal"
)

const (
	defaultSubmissionDelay = 30 * time.Second
)

type submissionQueue struct {
	logger         log.Logger
	bridgeContract ethconnect.BridgeContract
	address        common.Address
}

func newSubmissionQueue(
	logger log.Logger,
	bridgeContract ethconnect.BridgeContract,
	address common.Address,
) *submissionQueue {
	return &submissionQueue{
		logger:         logger,
		bridgeContract: bridgeContract,
		address:        address,
	}
}

func (s *submissionQueue) GetSubmissionDelay(attestation *portal.MezoBridgeAssetsUnlocked) time.Duration {
	queue, err := s.calculateSubmissionQueue(attestation.UnlockSequenceNumber)
	if err != nil {
		return time.Duration(0)
	}

	myValidatorID, err := s.bridgeContract.ValidatorIDs(s.address)
	if err != nil {
		return time.Duration(0)
	}

	return s.calculateSubmissionDelay(queue, myValidatorID)
}

func (s *submissionQueue) calculateSubmissionQueue(sequenceNumber *big.Int) ([]uint8, error) {
	validatorCount, err := s.bridgeContract.BridgeValidatorsCount()
	if err != nil {
		return nil, fmt.Errorf("failed to get validator count: %w", err)
	}

	count := validatorCount.Uint64()
	if count == 0 {
		return nil, fmt.Errorf("no validators found")
	}

	validatorIDs := make([]uint8, count)
	for i := uint64(0); i < count; i++ {
		validatorIDs[i] = uint8(i + 1) // #nosec G115 - this is fine, max is 255
	}

	seed := sequenceNumber.Int64()
	slices.Sort(validatorIDs)

	return s.shuffleValidatorIDs(validatorIDs, seed), nil
}

func (s *submissionQueue) shuffleValidatorIDs(validatorIDs []uint8, seed int64) []uint8 {
	cpy := make([]uint8, len(validatorIDs))
	copy(cpy, validatorIDs)

	source := rand.NewSource(seed)
	// #nosec G404 - this is alright for such kind of shuffling
	rng := rand.New(source)

	rng.Shuffle(len(cpy), func(i, j int) {
		cpy[i], cpy[j] = cpy[j], cpy[i]
	})

	return cpy
}

func (s *submissionQueue) calculateSubmissionDelay(queue []uint8, myValidatorID uint8) time.Duration {
	for i, validatorID := range queue {
		if validatorID == myValidatorID {
			return time.Duration(i) * defaultSubmissionDelay
		}
	}

	return 0
}
