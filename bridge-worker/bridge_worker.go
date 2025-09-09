package bridgeworker

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"sync"

	"cosmossdk.io/log"

	bwconfig "github.com/mezo-org/mezod/bridge-worker/config"

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

// BridgeWorker is a component responsible for tasks related to bridge-out
// process.
type BridgeWorker struct {
	logger log.Logger

	btcChain bitcoin.Chain // TODO: Initialize

	mezoBridgeContract *portal.MezoBridge
	tbtcBridgeContract *tbtc.Bridge

	chain *ethconnect.BaseChain

	batchSize         uint64
	requestsPerMinute uint64

	liveWalletsMutex sync.Mutex
	liveWallets      [][20]byte

	liveWalletsLastProcessedBlock uint64 // single-routine use; no mutex locking needed.

	btcWithdrawalLastProcessedBlock uint64

	btcWithdrawalMutex sync.Mutex
	btcWithdrawalQueue []portal.MezoBridgeAssetsUnlockConfirmed

	withdrawalFinalityChecksMutex sync.Mutex
	withdrawalFinalityChecks      map[string]*withdrawalFinalityCheck
}

func RunBridgeWorker(
	logger log.Logger,
	cfg bwconfig.Config,
	ethPrivateKey *ecdsa.PrivateKey,
) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	btcChain, err := electrum.Connect(ctx, cfg.Bitcoin.Electrum)
	if err != nil {
		panic(fmt.Sprintf("could not connect to Electrum chain: %v", err))
	}

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

	bw := &BridgeWorker{
		logger:                   logger,
		btcChain:                 btcChain,
		mezoBridgeContract:       mezoBridgeContract,
		tbtcBridgeContract:       tbtcBridgeContract,
		chain:                    chain,
		batchSize:                cfg.Ethereum.BatchSize,
		requestsPerMinute:        cfg.Ethereum.RequestsPerMinute,
		btcWithdrawalQueue:       []portal.MezoBridgeAssetsUnlockConfirmed{},
		withdrawalFinalityChecks: map[string]*withdrawalFinalityCheck{},
	}

	go func() {
		defer cancelCtx()
		err := bw.observeLiveWallets(ctx)
		if err != nil {
			bw.logger.Error(
				"live wallets observation routine failed",
				"err", err,
			)
		}

		bw.logger.Warn("live wallets observation routine stopped")
	}()

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
