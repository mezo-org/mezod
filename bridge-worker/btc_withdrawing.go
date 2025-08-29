package bridgeworker

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/mezo-org/mezod/ethereum/bindings/portal"
)

const (
	// bitcoinTargetChain is a numerical value representing Bitcoin target chain.
	bitcoinTargetChain = uint8(1)

	// TODO: Should the following be options?
	defaultBatchSize                      = uint64(1000)
	defaultRequestsPerMinute              = uint64(600) // 10 requests per second
	assetsUnlockedConfirmedLookBackPeriod = 216000      // ~30 days
)

func (bw *BridgeWorker) handleBitcoinWithdrawing(ctx context.Context) error {
	bw.logger.Info("Inside Bitcoin withdrawal logic")

	currentBlock, err := bw.chain.BlockCounter().CurrentBlock()
	if err != nil {
		return fmt.Errorf("failed to get current block: [%w]", err)
	}

	startBlock := uint64(0)
	if currentBlock > assetsUnlockedConfirmedLookBackPeriod {
		startBlock = currentBlock - assetsUnlockedConfirmedLookBackPeriod
	}

	recentEvents, err := bw.fetchAssetsUnlockConfirmedEvents(
		startBlock,
		currentBlock,
	)
	if err != nil {
		return fmt.Errorf(
			"failed to fetch AssetsUnlockConfirmed events: [%w]",
			err,
		)
	}

	for _, event := range recentEvents {
		isPendingWithdrawal, err := bw.isPendingBTCWithdrawal(event)
		if err != nil {
			return fmt.Errorf(
				"failed to check if event represents pending BTC withdrawal: [%w]",
				err,
			)
		}
		fmt.Println("isPendingWithdrawal", isPendingWithdrawal)
	}

	// TODO: Continue with implementation.

	<-ctx.Done()
	return ctx.Err()
}

func (bw *BridgeWorker) fetchAssetsUnlockConfirmedEvents(
	startBlock uint64,
	endBlock uint64,
) ([]*portal.MezoBridgeAssetsUnlockConfirmed, error) {
	bw.logger.Info(
		"fetching AssetsUnlockConfirmed events from range",
		"start_block", startBlock,
		"end_block", endBlock,
	)

	result := make([]*portal.MezoBridgeAssetsUnlockConfirmed, 0)

	ticker := time.NewTicker(time.Minute / time.Duration(bw.requestsPerMinute)) //nolint:gosec
	defer ticker.Stop()

	events, err := bw.bridgeContract.PastAssetsUnlockConfirmedEvents(
		startBlock,
		&endBlock,
		nil,
		nil,
		nil,
	)
	if err != nil {
		bw.logger.Warn(
			"failed to fetch AssetsUnlockConfirmed events from the entire "+
				"range; falling back to batched events fetch",
			"start_block", startBlock,
			"end_block", endBlock,
			"err", err,
		)

		batchStartBlock := startBlock

		for batchStartBlock <= endBlock {
			batchEndBlock := batchStartBlock + bw.batchSize
			if batchEndBlock > endBlock {
				batchEndBlock = endBlock
			}

			bw.logger.Info(
				"fetching a batch of AssetsUnlockConfirmed events from range",
				"batch_start_block", batchStartBlock,
				"batch_end_block", batchEndBlock,
			)

			<-ticker.C

			batchEvents, batchErr := bw.bridgeContract.PastAssetsUnlockConfirmedEvents(
				batchStartBlock,
				&batchEndBlock,
				nil,
				nil,
				nil,
			)
			if batchErr != nil {
				return nil, fmt.Errorf(
					"batched AssetsUnlockConfirmed fetch failed: [%w]; giving up",
					batchErr,
				)
			}

			result = append(result, batchEvents...)

			batchStartBlock = batchEndBlock + 1
		}
	} else {
		result = append(result, events...)
	}

	return result, nil
}

func (bw *BridgeWorker) isPendingBTCWithdrawal(
	event *portal.MezoBridgeAssetsUnlockConfirmed,
) (bool, error) {
	if event.Chain == bitcoinTargetChain {
		return false, nil
	}

	hash, err := computeAssetsUnlockedHash(
		event.UnlockSequenceNumber,
		event.Recipient[:],
		event.Token,
		event.Amount,
		event.Chain,
	)
	if err != nil {
		return false, fmt.Errorf(
			"failed to calculate AssetsUnlocked hash: [%w]",
			err,
		)
	}

	isPendingBTCWithdrawal, err := bw.bridgeContract.PendingBTCWithdrawals(hash)
	if err != nil {
		return false, fmt.Errorf(
			"failed to get pending BTC withdrawals info: [%w]",
			err,
		)
	}

	return isPendingBTCWithdrawal, nil
}

func computeAssetsUnlockedHash(
	unlockSeq *big.Int,
	recipient []byte,
	token common.Address,
	amount *big.Int,
	chain uint8,
) (common.Hash, error) {
	type assetsUnlockedTuple struct {
		UnlockSequenceNumber *big.Int       `abi:"unlockSequenceNumber"`
		Recipient            []byte         `abi:"recipient"`
		Token                common.Address `abi:"token"`
		Amount               *big.Int       `abi:"amount"`
		Chain                uint8          `abi:"chain"`
	}

	tupleType, err := abi.NewType("tuple", "tuple", []abi.ArgumentMarshaling{
		{Name: "unlockSequenceNumber", Type: "uint256"}, // unlockSequenceNumber
		{Name: "recipient", Type: "bytes"},              // recipient
		{Name: "token", Type: "address"},                // token
		{Name: "amount", Type: "uint256"},               // amount
		{Name: "chain", Type: "uint8"},                  // chain
	})
	if err != nil {
		return common.Hash{}, err
	}

	entry := assetsUnlockedTuple{
		UnlockSequenceNumber: unlockSeq,
		Recipient:            recipient,
		Token:                token,
		Amount:               amount,
		Chain:                chain,
	}

	bytes, err := (abi.Arguments{{Type: tupleType}}).Pack(entry)
	if err != nil {
		return common.Hash{}, err
	}

	return crypto.Keccak256Hash(bytes), nil
}
