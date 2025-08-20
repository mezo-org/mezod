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

type SubmissionQueue struct {
	logger         log.Logger
	bridgeContract ethconnect.BridgeContract
	address        common.Address
}

func NewSubmissionQueue(
	logger log.Logger,
	bridgeContract ethconnect.BridgeContract,
	address common.Address,
) *SubmissionQueue {
	return &SubmissionQueue{
		logger:         logger,
		bridgeContract: bridgeContract,
		address:        address,
	}
}

func (s *SubmissionQueue) GetSubmissionDelay(attestation *portal.MezoBridgeAssetsUnlocked) (time.Duration, error) {
	queue, err := s.calculateSubmissionQueue(attestation.UnlockSequenceNumber)
	if err != nil {
		return time.Duration(0), fmt.Errorf("failed to calculate submission queue: %w", err)
	}

	myValidatorID, err := s.bridgeContract.ValidatorIDs(s.address)
	if err != nil {
		return time.Duration(0), fmt.Errorf("failed to get validator ID: %w", err)
	}

	return s.calculateSubmissionDelay(queue, myValidatorID), nil
}

func (s *SubmissionQueue) calculateSubmissionQueue(sequenceNumber *big.Int) ([]uint8, error) {
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
		validatorIDs[i] = uint8(i) // #nosec G115 - this is fine, max is 255
	}

	seed := sequenceNumber.Int64()
	slices.Sort(validatorIDs)
	s.shuffleValidatorIDs(validatorIDs, seed)

	return validatorIDs, nil
}

func (s *SubmissionQueue) shuffleValidatorIDs(validatorIDs []uint8, seed int64) {
	source := rand.NewSource(seed)
	// #nosec G404 - this is alright for such kind of shuffling
	rng := rand.New(source)

	rng.Shuffle(len(validatorIDs), func(i, j int) {
		validatorIDs[i], validatorIDs[j] = validatorIDs[j], validatorIDs[i]
	})
}

func (s *SubmissionQueue) calculateSubmissionDelay(queue []uint8, myValidatorID uint8) time.Duration {
	for i, validatorID := range queue {
		if validatorID == myValidatorID {
			return time.Duration(i) * defaultSubmissionDelay
		}
	}

	return 0
}
