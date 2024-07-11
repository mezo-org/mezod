package ethereum

import (
	"context"
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ipfs/go-log"

	"github.com/keep-network/keep-common/pkg/chain/ethereum"
	"github.com/keep-network/keep-common/pkg/chain/ethereum/ethutil"
	"github.com/keep-network/keep-common/pkg/rate"
)

var logger = log.Logger("mezo-ethereum")

// baseChain represents a base, non-application-specific chain handle. It
// provides the implementation of generic features like balance monitor,
// block counter and similar.
type baseChain struct {
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
}

// Connect creates Portal Ethereum chain handle.
// TODO: Test the connectivity with the `Portal` smart contract. Due to the
//
//	outdated go-ethereum package it could not be tested so far.
func ConnectPortal(
	ctx context.Context,
	config ethereum.Config,
) (
	*PortalChain,
	error,
) {
	client, err := ethclient.Dial(config.URL)
	if err != nil {
		return nil, fmt.Errorf(
			"error Connecting to Ethereum Server: %s [%v]",
			config.URL,
			err,
		)
	}

	baseChain, err := newBaseChain(ctx, config, client)
	if err != nil {
		return nil, fmt.Errorf(
			"could not create base chain handle: [%v]",
			err,
		)
	}

	portalChain, err := newPortalChain(config, baseChain)
	if err != nil {
		return nil, fmt.Errorf(
			"could not create portal chain handle: [%v]",
			err,
		)
	}

	return portalChain, nil
}

// newChain construct a new instance of the Ethereum chain handle.
func newBaseChain(
	ctx context.Context,
	config ethereum.Config,
	client *ethclient.Client,
) (*baseChain, error) {
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

	// TODO: Validators interact with Ethereum in read-only mode. Therefore,
	//       they should not be required to have an Ethereum key.
	key, err := decryptKey(config)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to decrypt Ethereum key: [%v]",
			err,
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

	return &baseChain{
		key:              key,
		client:           clientWithAddons,
		chainID:          chainID,
		blockCounter:     blockCounter,
		nonceManager:     nonceManager,
		miningWaiter:     miningWaiter,
		transactionMutex: transactionMutex,
	}, nil
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

// decryptKey decrypts the chain key pointed by the config.
func decryptKey(config ethereum.Config) (*keystore.Key, error) {
	return ethutil.DecryptKeyFile(
		config.Account.KeyFile,
		config.Account.KeyFilePassword,
	)
}
