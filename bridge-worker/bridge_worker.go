package bridgeworker

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"sync"
	"time"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"

	"github.com/mezo-org/mezod/bridge-worker/bitcoin"
	"github.com/mezo-org/mezod/bridge-worker/bitcoin/electrum"

	"github.com/ethereum/go-ethereum/common"
	ethconfig "github.com/keep-network/keep-common/pkg/chain/ethereum"
	ethconnect "github.com/mezo-org/mezod/ethereum"
	"github.com/mezo-org/mezod/ethereum/bindings/portal"
	"github.com/mezo-org/mezod/ethereum/bindings/tbtc"
	bridgetypes "github.com/mezo-org/mezod/x/bridge/types"
)

// mezoBridgeName is the name of the MezoBridge contract.
const mezoBridgeName = "MezoBridge"

// AssetsUnlockedEndpoint is a client enabling communication with the `Mezo`
// chain.
type AssetsUnlockedEndpoint interface {
	// GetAssetsUnlockedEvents gets the AssetsUnlocked events from the Mezo
	// chain. The requested range of events is inclusive on the lower side and
	// exclusive on the upper side.
	GetAssetsUnlockedEvents(
		ctx context.Context,
		sequenceStart sdkmath.Int,
		sequenceEnd sdkmath.Int,
	) ([]bridgetypes.AssetsUnlockedEvent, error)
}

// BridgeWorker is a component responsible for tasks related to bridge-out
// process.
type BridgeWorker struct {
	logger log.Logger

	mezoBridgeContract *portal.MezoBridge
	tbtcBridgeContract *tbtc.Bridge

	chain *ethconnect.BaseChain

	batchSize         uint64
	requestsPerMinute uint64

	btcChain               bitcoin.Chain
	assetsUnlockedEndpoint AssetsUnlockedEndpoint

	// Channel used to indicate whether the initial fetching of live wallets
	// is done. Once this channel is closed we can proceed with the Bitcoin
	// withdrawing routines.
	liveWalletsReady chan struct{}

	liveWalletsMutex sync.Mutex
	liveWallets      [][20]byte

	liveWalletsLastProcessedBlock uint64 // single-routine use; no mutex locking needed.

	btcWithdrawalLastProcessedBlock uint64

	btcWithdrawalMutex sync.Mutex
	btcWithdrawalQueue []portal.MezoBridgeAssetsUnlockConfirmed

	withdrawalFinalityChecksMutex sync.Mutex
	withdrawalFinalityChecks      map[string]*withdrawalFinalityCheck

	btcWithdrawalQueueCheckFrequency time.Duration
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

	logger.Info(
		"connecting to Mezo assets unlock endpoint",
		"endpoint", cfg.Mezo.AssetsUnlockEndpoint,
	)

	// The messages handled by the bridge-worker contain custom types.
	// Add codecs so that the messages can be marshaled/unmarshalled.
	assetsUnlockedGrpcEndpoint, err := NewAssetsUnlockedGrpcEndpoint(
		cfg.Mezo.AssetsUnlockEndpoint,
		codectypes.NewInterfaceRegistry(),
	)
	if err != nil {
		panic(fmt.Sprintf("failed to create assets unlocked endpoint: %v", err))
	}

	go func() {
		// Test the connection to the assets unlocked endpoint by verifying we
		// can successfully execute `GetAssetsUnlockedEvents`.
		ctxWithTimeout, cancel := context.WithTimeout(
			context.Background(),
			requestTimeout,
		)
		defer cancel()

		_, err := assetsUnlockedGrpcEndpoint.GetAssetsUnlockedEvents(
			ctxWithTimeout,
			sdkmath.NewInt(1),
			sdkmath.NewInt(2),
		)
		if err != nil {
			logger.Error(
				"assets unlocked endpoint connection test failed; possible "+
					"problem with configuration or connectivity",
				"err", err,
			)
		} else {
			logger.Info(
				"assets unlocked endpoint connection test completed successfully",
			)
		}
	}()

	bw := &BridgeWorker{
		logger:                           logger,
		mezoBridgeContract:               mezoBridgeContract,
		tbtcBridgeContract:               tbtcBridgeContract,
		chain:                            chain,
		batchSize:                        cfg.Ethereum.BatchSize,
		requestsPerMinute:                cfg.Ethereum.RequestsPerMinute,
		btcChain:                         btcChain,
		assetsUnlockedEndpoint:           assetsUnlockedGrpcEndpoint,
		liveWalletsReady:                 make(chan struct{}),
		btcWithdrawalQueue:               []portal.MezoBridgeAssetsUnlockConfirmed{},
		withdrawalFinalityChecks:         map[string]*withdrawalFinalityCheck{},
		btcWithdrawalQueueCheckFrequency: cfg.Job.BTCWithdrawal.QueueCheckFrequency,
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

	// Wait until the initial fetching of live wallets is done.
	bw.logger.Info("waiting for initial fetching of live wallet")
	select {
	case <-bw.liveWalletsReady:
		bw.logger.Info("initial fetching of live wallets completed")
	case <-ctx.Done():
		bw.logger.Warn(
			"context canceled while waiting; exiting without launching " +
				"Bitcoin withdrawal routines",
		)
		return ctx.Err()
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
