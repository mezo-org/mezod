package ethereum

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/ethclient"
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
	client           ethutil.EthereumClient
	chainID          *big.Int
	blockCounter     *ethereum.BlockCounter
	finalizedBlockFn func(ctx context.Context) (*big.Int, error)
}

type block struct {
	Number string `json:"number"`
}

// Connect creates Ethereum chain handle.
func Connect(
	ctx context.Context,
	config ethereum.Config,
) (
	*BaseChain,
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

	return baseChain, nil
}

// newChain construct a new instance of the Ethereum chain handle.
func newBaseChain(
	ctx context.Context,
	config ethereum.Config,
	client *ethclient.Client,
) (*BaseChain, error) {
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
		client:           clientWithAddons,
		chainID:          chainID,
		blockCounter:     blockCounter,
		finalizedBlockFn: finalizedBlockFn,
	}, nil
}

func (bc *BaseChain) BlockCounter() *ethereum.BlockCounter {
	return bc.blockCounter
}

func (bc *BaseChain) FinalizedBlock(ctx context.Context) (*big.Int, error) {
	return bc.finalizedBlockFn(ctx)
}

func (bc *BaseChain) Client() ethutil.EthereumClient {
	return bc.client
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
