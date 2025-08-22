package sidecar

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"cosmossdk.io/log"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	ethconfig "github.com/keep-network/keep-common/pkg/chain/ethereum"
	ethconnect "github.com/mezo-org/mezod/ethereum"
	"github.com/mezo-org/mezod/ethereum/bindings/portal"
)

var defaultIsValidTickerDuration = time.Second

type attestationValidator struct {
	logger         log.Logger
	bridgeContract ethconnect.BridgeContract
	address        common.Address
}

func newAttestationValidation(
	logger log.Logger,
	bridgeContract ethconnect.BridgeContract,
	validatorAddress common.Address,
) *attestationValidator {
	return &attestationValidator{
		logger:         logger,
		bridgeContract: bridgeContract,
		address:        validatorAddress,
	}
}

func (av *attestationValidator) IsValid(ctx context.Context, bridgeAssetsUnlocked *portal.MezoBridgeAssetsUnlocked) (bool, error) {
	// ticker is used to retry the validation in case
	// of network transient failure.
	ticker := time.NewTicker(defaultIsValidTickerDuration)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ok, err := av.bridgeContract.ValidateAssetsUnlocked(*bridgeAssetsUnlocked)
			if err != nil {
				av.logger.Error("couldn't call validateAssetsUnlocked", "error", err)
				continue
			}
			return ok, nil
		case <-ctx.Done():
			av.logger.Info("stopping assets unlocked validation due to context cancellation")
			return false, ctx.Err()
		}
	}
}

func (av *attestationValidator) IsConfirmed(
	attestation *portal.MezoBridgeAssetsUnlocked,
) (bool, error) {
	ok, err := av.bridgeContract.ConfirmedUnlocks(attestation.Amount)
	if err != nil {
		return false, fmt.Errorf("couldn't get confirmedLocks: %w", err)
	}
	if ok {
		return true, nil
	}

	return av.checkOwnAttestation(attestation)
}

func (av *attestationValidator) WaitForAttestationConfirmation(
	blockHeightWaiter ethconfig.BlockHeightWaiter,
	startBlock, confirmations uint64,
	attestation *portal.MezoBridgeAssetsUnlocked,
) (bool, error) {
	return waitForBlockConfirmations(
		blockHeightWaiter,
		startBlock,
		confirmations,
		func() (bool, error) {
			return av.checkOwnAttestation(attestation)
		},
	)
}

func (av *attestationValidator) checkOwnAttestation(
	attestation *portal.MezoBridgeAssetsUnlocked,
) (bool, error) {
	encoded, err := abiEncodeAttestation(attestation)
	if err != nil {
		return false, fmt.Errorf("couldn't ABI encode attestation: %w", err)
	}

	hash := crypto.Keccak256Hash(encoded)

	bitmap, err := av.bridgeContract.Attestations(hash)
	if err != nil {
		return false, fmt.Errorf("couldn't get confirmedLock: %w", err)
	}

	validatorID, err := av.bridgeContract.ValidatorIDs(av.address)
	if err != nil {
		return false, fmt.Errorf("couldn't get validator ID: %w", err)
	}

	mask := new(big.Int).Lsh(big.NewInt(1), uint(validatorID))

	if new(big.Int).And(bitmap, mask).Int64() == 0 {
		return false, nil
	}

	return true, nil
}

func abiEncodeAttestation(attestation *portal.MezoBridgeAssetsUnlocked) ([]byte, error) {
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

	return abi.Arguments{
		{Type: uint256Type},
		{Type: bytesType},
		{Type: addressType},
		{Type: uint256Type},
		{Type: uint8Type},
	}.Pack(
		attestation.UnlockSequenceNumber,
		attestation.Recipient,
		attestation.Token,
		attestation.Amount,
		attestation.Chain,
	)
}

// waitForBlockConfirmations ensures that after receiving specific number of block
// confirmations the state of the chain is actually as expected. It waits for
// predefined number of blocks since the start block number provided. After the
// required block number is reached it performs a check of the chain state with
// a provided function returning a error.
func waitForBlockConfirmations(
	blockHeightWaiter ethconfig.BlockHeightWaiter,
	startBlockNumber uint64,
	blockConfirmations uint64,
	stateCheck func() (bool, error),
) (bool, error) {
	blockHeight := startBlockNumber + blockConfirmations

	err := blockHeightWaiter.WaitForBlockHeight(blockHeight)
	if err != nil {
		return false, fmt.Errorf("failed to wait for block height: [%v]", err)
	}

	return stateCheck()
}
