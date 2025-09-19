package bridgeworker

import (
	"context"
	"crypto/ecdsa"
	"fmt"

	"cosmossdk.io/log"

	"github.com/mezo-org/mezod/bridge-worker/bitcoin"
	"github.com/mezo-org/mezod/bridge-worker/bitcoin/electrum"

	"github.com/ethereum/go-ethereum/common"
	ethconfig "github.com/keep-network/keep-common/pkg/chain/ethereum"
	ethconnect "github.com/mezo-org/mezod/ethereum"
	"github.com/mezo-org/mezod/ethereum/bindings/portal"
	"github.com/mezo-org/mezod/ethereum/bindings/tbtc"
)

// mezoBridgeName is the name of the MezoBridge contract.
const mezoBridgeName = "MezoBridge"

// bridgeWorkerJob represent a bridge worker job.
type bridgeWorkerJob interface {
	run(ctx context.Context)
}

// environment groups common elements for all bridge worker jobs.
type environment struct {
	logger log.Logger

	mezoBridgeContract *portal.MezoBridge
	tbtcBridgeContract *tbtc.Bridge

	chain *ethconnect.BaseChain

	batchSize         uint64
	requestsPerMinute uint64

	btcChain bitcoin.Chain
}

func RunBridgeWorker(
	logger log.Logger,
	cfg Config,
	ethPrivateKey *ecdsa.PrivateKey,
) error {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	ethereumNetwork := ethconnect.NetworkFromString(cfg.Ethereum.Network)
	mezoBridgeAddress := portal.MezoBridgeAddress(ethereumNetwork)
	tbtcBridgeAddress := tbtc.BridgeAddress(ethereumNetwork)

	if mezoBridgeAddress == "" {
		panic(
			"cannot get address of the MezoBridge contract on Ethereum; " +
				"make sure you run 'make bindings' before building the binary",
		)
	}

	if tbtcBridgeAddress == "" {
		panic(
			"cannot get address of the Tbtc Bridge contract on Ethereum",
		)
	}

	logger.Info(
		"resolved contract addresses and Ethereum network",
		"mezo_bridge_address", mezoBridgeAddress,
		"tbtc_bridge_address", tbtcBridgeAddress,
		"ethereum_network", ethereumNetwork,
	)

	// Connect to the Ethereum network
	chain, err := ethconnect.Connect(
		ctx,
		ethconfig.Config{
			Network:           ethereumNetwork,
			URL:               cfg.Ethereum.ProviderURL,
			ContractAddresses: map[string]string{mezoBridgeName: mezoBridgeAddress},
		},
		ethPrivateKey,
	)
	if err != nil {
		panic(fmt.Sprintf("failed to connect to the Ethereum network: %v", err))
	}

	// Initialize the MezoBridge contract instance.
	mezoBridgeContract, err := initializeMezoBridgeContract(common.HexToAddress(mezoBridgeAddress), chain)
	if err != nil {
		panic(fmt.Sprintf("failed to initialize MezoBridge contract: %v", err))
	}

	// Initialize the tBTC Bridge contract instance.
	tbtcBridgeContract, err := initializeTbtcBridgeContract(common.HexToAddress(tbtcBridgeAddress), chain)
	if err != nil {
		panic(fmt.Sprintf("failed to initialize tBTC Bridge contract: %v", err))
	}

	logger.Info(
		"connecting to electrum node",
		"electrum_url", cfg.Bitcoin.Electrum.URL,
		"network", cfg.Bitcoin.Network.String(),
	)

	btcChain, err := electrum.Connect(
		ctx,
		cfg.Bitcoin.Electrum,
		logger.With(log.ModuleKey, "electrum"),
	)
	if err != nil {
		panic(fmt.Sprintf("could not connect to Electrum chain: %v", err))
	}

	env := &environment{
		logger:             logger,
		mezoBridgeContract: mezoBridgeContract,
		tbtcBridgeContract: tbtcBridgeContract,
		chain:              chain,
		batchSize:          cfg.Ethereum.BatchSize,
		requestsPerMinute:  cfg.Ethereum.RequestsPerMinute,
		btcChain:           btcChain,
	}

	jobs := []bridgeWorkerJob{
		newBTCWithdrawalJob(
			env,
			cfg.Mezo.AssetsUnlockEndpoint,
			cfg.Job.BTCWithdrawal.QueueCheckFrequency,
		),
	}

	for _, job := range jobs {
		go func(j bridgeWorkerJob) {
			defer cancelCtx()
			j.run(ctx)
		}(job)
	}

	<-ctx.Done()

	logger.Info("bridge worker stopped")

	return nil
}

// Construct a new instance of the Ethereum MezoBridge contract.
func initializeMezoBridgeContract(
	address common.Address,
	chain *ethconnect.BaseChain,
) (*portal.MezoBridge, error) {
	bridgeContract, err := portal.NewMezoBridge(
		address,
		chain.ChainID(),
		chain.Key(),
		chain.Client(),
		chain.NonceManager(),
		chain.MiningWaiter(),
		chain.BlockCounter(),
		chain.TransactionMutex(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to attach to MezoBridge contract. %v", err)
	}

	return bridgeContract, nil
}

// Construct a new instance of the Ethereum Tbtc Bridge contract.
func initializeTbtcBridgeContract(
	address common.Address,
	chain *ethconnect.BaseChain,
) (*tbtc.Bridge, error) {
	bridgeContract, err := tbtc.NewTbtcBridge(
		address,
		chain.ChainID(),
		chain.Key(),
		chain.Client(),
		chain.NonceManager(),
		chain.MiningWaiter(),
		chain.BlockCounter(),
		chain.TransactionMutex(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to attach to Tbtc Bridge contract. %v", err)
	}

	return bridgeContract, nil
}
