package bridgeworker

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"sync"

	"cosmossdk.io/log"

	"github.com/ethereum/go-ethereum/common"
	ethconfig "github.com/keep-network/keep-common/pkg/chain/ethereum"
	ethconnect "github.com/mezo-org/mezod/ethereum"
	"github.com/mezo-org/mezod/ethereum/bindings/portal"
	"github.com/mezo-org/mezod/ethereum/bindings/tbtc"
)

// mezoBridgeName is the name of the MezoBridge contract.
const mezoBridgeName = "MezoBridge"

// BridgeWorker is a component responsible for tasks related to bridge-out
// process.
type BridgeWorker struct {
	logger log.Logger

	mezoBridgeContract *portal.MezoBridge
	tbtcBridgeContract *tbtc.Bridge

	chain *ethconnect.BaseChain

	batchSize         uint64
	requestsPerMinute uint64

	btcWithdrawalLastProcessedBlock uint64

	btcWithdrawalMutex sync.Mutex
	btcWithdrawalQueue []portal.MezoBridgeAssetsUnlockConfirmed

	withdrawalFinalityChecksMutex sync.Mutex
	withdrawalFinalityChecks      map[string]*withdrawalFinalityCheck
}

func RunBridgeWorker(
	logger log.Logger,
	providerURL string,
	ethereumNetwork string,
	privateKey *ecdsa.PrivateKey,
) {
	network := ethconnect.NetworkFromString(ethereumNetwork)
	mezoBridgeAddress := portal.MezoBridgeAddress(network)
	tbtcBridgeAddress := tbtc.BridgeAddress(network)

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
		"ethereum_network", network,
	)

	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	// Connect to the Ethereum network
	chain, err := ethconnect.Connect(
		ctx,
		ethconfig.Config{
			Network:           network,
			URL:               providerURL,
			ContractAddresses: map[string]string{mezoBridgeName: mezoBridgeAddress},
		},
		privateKey,
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

	bw := &BridgeWorker{
		logger:                   logger,
		mezoBridgeContract:       mezoBridgeContract,
		tbtcBridgeContract:       tbtcBridgeContract,
		chain:                    chain,
		batchSize:                defaultBatchSize,
		requestsPerMinute:        defaultRequestsPerMinute,
		btcWithdrawalQueue:       []portal.MezoBridgeAssetsUnlockConfirmed{},
		withdrawalFinalityChecks: map[string]*withdrawalFinalityCheck{},
	}

	go func() {
		defer cancelCtx()
		err := bw.observeBitcoinWithdrawals(ctx)
		if err != nil {
			bw.logger.Error(
				"Bitcoin withdrawal observation routine failed",
				"err", err,
			)
		}

		bw.logger.Warn("Bitcoin withdrawal observation routine stopped")
	}()

	go func() {
		defer cancelCtx()
		bw.processBtcWithdrawalQueue(ctx)
		bw.logger.Warn("Bitcoin withdrawal processing loop stopped")
	}()

	go func() {
		defer cancelCtx()
		bw.processWithdrawalFinalityChecks(ctx)
		bw.logger.Warn("Bitcoin withdrawal finality checks loop stopped")
	}()

	<-ctx.Done()

	bw.logger.Info("bridge worker stopped")
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
