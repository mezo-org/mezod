package ethereum

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/mezo-org/mezod/ethereum/bindings/portal"
)

type AssetsLockedIterator interface {
	Next() bool
	Error() error
	Close() error
	Event() *portal.MezoBridgeAssetsLocked
}

type BridgeContract interface {
	FilterAssetsLocked(
		opts *bind.FilterOpts,
		sequenceNumber []*big.Int,
		recipient []common.Address,
		token []common.Address,
	) (AssetsLockedIterator, error)

	PastAssetsUnlockConfirmedEvents(
		startBlock uint64,
		endBlock *uint64,
		unlockSequenceNumberFilter []*big.Int,
		recipientFilter [][]byte,
		tokenFilter []common.Address,
	) ([]*MezoBridgeAssetsUnlockConfirmed, error)

	ConfirmedUnlocks(arg0 *big.Int) (bool, error)
}

// TODO: Remove once bindings for `MezoBridge` re-generated and contain features
//
//	related to attestation.
type MezoBridgeAssetsUnlockConfirmed struct {
	UnlockSequenceNumber *big.Int
	Recipient            common.Hash
	Token                common.Address
	Amount               *big.Int
	Chain                uint8
	Raw                  Log
}

// TODO: Remove once bindings for `MezoBridge` re-generated and contain features
//
//	related to attestation.
type Log struct {
	// Consensus fields:
	// address of the contract that generated the event
	Address common.Address `json:"address" gencodec:"required"`
	// list of topics provided by the contract.
	Topics []common.Hash `json:"topics" gencodec:"required"`
	// supplied by the contract, usually ABI-encoded
	Data []byte `json:"data" gencodec:"required"`

	// Derived fields. These fields are filled in by the node
	// but not secured by consensus.
	// block in which the transaction was included
	BlockNumber uint64 `json:"blockNumber" rlp:"-"`
	// hash of the transaction
	TxHash common.Hash `json:"transactionHash" gencodec:"required" rlp:"-"`
	// index of the transaction in the block
	TxIndex uint `json:"transactionIndex" rlp:"-"`
	// hash of the block in which the transaction was included
	BlockHash common.Hash `json:"blockHash" rlp:"-"`
	// index of the log in the block
	Index uint `json:"logIndex" rlp:"-"`

	// The Removed field is true if this log was reverted due to a chain reorganisation.
	// You must pay attention to this field if you receive logs through a filter query.
	Removed bool `json:"removed" rlp:"-"`
}
