package ethereum

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"net/url"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/google/uuid"
	"github.com/ipfs/go-log"

	"github.com/keep-network/keep-common/pkg/chain/ethereum"
	"github.com/keep-network/keep-common/pkg/chain/ethereum/ethutil"
	"github.com/keep-network/keep-common/pkg/rate"
)

var logger = log.Logger("mezo-ethereum")

// BaseChain represents a base, non-application-specific chain handle. It
// provides the implementation of generic features like balance monitor,
// block counter and similar.
type BaseChain struct {
	key     *keystore.Key
	client  ethutil.EthereumClient
	chainID *big.Int

	blockCounter *ethereum.BlockCounter
	nonceManager *ethereum.NonceManager
	miningWaiter *ethutil.MiningWaiter

	// transactionMutex allows interested parties to forcibly serialize
	// transaction submission.
	//
	// When transactions are submitted, they require a valid nonce. The nonce is
	// equal to the count of transactions the account has submitted so far, and
	// for a transaction to be accepted it should be monotonically greater than
	// any previous submitted transaction. To do this, transaction submission
	// asks the Ethereum client it is connected to for the next pending nonce,
	// and uses that value for the transaction. Unfortunately, if multiple
	// transactions are submitted in short order, they may all get the same
	// nonce. Serializing submission ensures that each nonce is requested after
	// a previous transaction has been submitted.
	transactionMutex *sync.Mutex

	finalizedBlockFn func(ctx context.Context) (*big.Int, error)
}

type block struct {
	Number string `json:"number"`
}

// Connect creates Ethereum chain handle.
func Connect(
	ctx context.Context,
	config ethereum.Config,
	privateKey *ecdsa.PrivateKey,
) (
	*BaseChain,
	error,
) {
	parsedURL, err := url.Parse(config.URL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL provided for ETH client")
	}

	// Enforce the connection via WebSockets as other protocols may not support
	// subscriptions.
	if parsedURL.Scheme != "wss" && parsedURL.Scheme != "ws" {
		return nil, fmt.Errorf(
			"ETH client requires a WebSocket URL starting with wss:// " +
				"(recommended) or ws://",
		)
	}

	client, err := ethclient.Dial(config.URL)
	if err != nil {
		return nil, fmt.Errorf("error connecting to ETH Server: [%v]", err)
	}

	// Double-check if subscriptions are supported.
	if !client.Client().SupportsSubscriptions() {
		client.Close()
		return nil, fmt.Errorf("ETH client does not support subscriptions")
	}

	baseChain, err := newBaseChain(ctx, config, client, privateKey)
	if err != nil {
		return nil, fmt.Errorf(
			"could not create base chain handle: [%v]",
			err,
		)
	}

	return baseChain, nil
}

// newChain construct a new instance of the Ethereum chain handle.
func newBaseChain(
	ctx context.Context,
	config ethereum.Config,
	client *ethclient.Client,
	privateKey *ecdsa.PrivateKey,
) (*BaseChain, error) {
	key := newKeyFromECDSA(privateKey)

	chainID, err := client.ChainID(ctx)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to resolve Ethereum chain id: [%v]",
			err,
		)
	}

	if config.Network != ethereum.Developer &&
		big.NewInt(config.Network.ChainID()).Cmp(chainID) != 0 {
		return nil, fmt.Errorf(
			"chain id returned from ethereum api [%s] "+
				"doesn't match the expected chain id [%d] for [%s] network; "+
				"please verify the configured ethereum.url",
			chainID.String(),
			config.Network.ChainID(),
			config.Network,
		)
	}

	clientWithAddons := wrapClientAddons(config, client)

	blockCounter, err := ethutil.NewBlockCounter(clientWithAddons)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to create Ethereum blockcounter: [%v]",
			err,
		)
	}

	nonceManager := ethutil.NewNonceManager(
		clientWithAddons,
		key.Address,
	)

	miningWaiter := ethutil.NewMiningWaiter(clientWithAddons, config)

	transactionMutex := &sync.Mutex{}

	finalizedBlockFn := func(ctx context.Context) (*big.Int, error) {
		var finalized block
		err = client.Client().CallContext(ctx, &finalized, "eth_getBlockByNumber", "finalized", false)
		if err != nil {
			return nil, fmt.Errorf(
				"failed to get the finalized block: %v",
				err,
			)
		}

		hexNumber := strings.TrimPrefix(finalized.Number, "0x")
		finalizedBlock, ok := new(big.Int).SetString(hexNumber, 16) // hex to decimal
		if !ok {
			return nil, fmt.Errorf(
				"failed to convert finalized block number to integer: %v",
				err,
			)
		}

		return finalizedBlock, nil
	}

	return &BaseChain{
		key:              key,
		client:           clientWithAddons,
		chainID:          chainID,
		blockCounter:     blockCounter,
		nonceManager:     nonceManager,
		miningWaiter:     miningWaiter,
		transactionMutex: transactionMutex,
		finalizedBlockFn: finalizedBlockFn,
	}, nil
}

func (bc *BaseChain) Key() *keystore.Key {
	return bc.key
}

func (bc *BaseChain) Client() ethutil.EthereumClient {
	return bc.client
}

func (bc *BaseChain) ChainID() *big.Int {
	return bc.chainID
}

func (bc *BaseChain) BlockCounter() *ethereum.BlockCounter {
	return bc.blockCounter
}

func (bc *BaseChain) NonceManager() *ethereum.NonceManager {
	return bc.nonceManager
}

func (bc *BaseChain) MiningWaiter() *ethutil.MiningWaiter {
	return bc.miningWaiter
}

func (bc *BaseChain) TransactionMutex() *sync.Mutex {
	return bc.transactionMutex
}

func (bc *BaseChain) FinalizedBlock(ctx context.Context) (*big.Int, error) {
	return bc.finalizedBlockFn(ctx)
}

func (bc *BaseChain) CurrentBlock() (uint64, error) {
	return bc.blockCounter.CurrentBlock()
}

func (bc *BaseChain) LatestBlock(ctx context.Context) (*big.Int, error) {
	r, err := bc.Client().HeaderByNumber(ctx, nil)
	if err != nil {
		return nil, err
	}

	return r.Number, nil
}

func (bc *BaseChain) WatchBlocks(ctx context.Context) <-chan uint64 {
	return bc.blockCounter.WatchBlocks(ctx)
}

// wrapClientAddons wraps the client instance with add-ons like logging, rate
// limiting and so on.
func wrapClientAddons(
	config ethereum.Config,
	client ethutil.EthereumClient,
) ethutil.EthereumClient {
	loggingClient := ethutil.WrapCallLogging(logger, client)

	if config.RequestsPerSecondLimit > 0 || config.ConcurrencyLimit > 0 {
		logger.Infof(
			"enabled ethereum client request rate limiter; "+
				"rps limit [%v]; "+
				"concurrency limit [%v]",
			config.RequestsPerSecondLimit,
			config.ConcurrencyLimit,
		)

		return ethutil.WrapRateLimiting(
			loggingClient,
			&rate.LimiterConfig{
				RequestsPerSecondLimit: config.RequestsPerSecondLimit,
				ConcurrencyLimit:       config.ConcurrencyLimit,
			},
		)
	}

	return loggingClient
}

// NetworkFromString converts a string to an ethereum.Network.
func NetworkFromString(networkStr string) ethereum.Network {
	switch networkStr {
	case ethereum.Mainnet.String():
		return ethereum.Mainnet
	case ethereum.Sepolia.String():
		return ethereum.Sepolia
	case ethereum.Developer.String():
		return ethereum.Developer
	default:
		return ethereum.Unknown
	}
}

func newKeyFromECDSA(privateKey *ecdsa.PrivateKey) *keystore.Key {
	id, err := uuid.NewRandom()
	if err != nil {
		panic(fmt.Sprintf("could not create random uuid: %v", err))
	}

	if privateKey == nil {
		panic("private key is required")
	}

	return &keystore.Key{
		Id:         id,
		Address:    crypto.PubkeyToAddress(privateKey.PublicKey),
		PrivateKey: privateKey,
	}
}
