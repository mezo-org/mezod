package bridgeworker

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"cosmossdk.io/log"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/mezo-org/mezod/bridge-worker/ethereum"

	ethconfig "github.com/keep-network/keep-common/pkg/chain/ethereum"
	"github.com/mezo-org/mezod/bridge-worker/bitcoin"
	"github.com/mezo-org/mezod/bridge-worker/bitcoin/electrum"
	ethconnect "github.com/mezo-org/mezod/ethereum"
	"github.com/mezo-org/mezod/ethereum/bindings/portal"
	"github.com/mezo-org/mezod/ethereum/bindings/tbtc"
)

// mezoBridgeName is the name of the MezoBridge contract.
const mezoBridgeName = "MezoBridge"

// EthereumChain provides basic information about the Ethereum chain.
type EthereumChain interface {
	FinalizedBlock(ctx context.Context) (*big.Int, error)
	CurrentBlock() (uint64, error)
	WatchBlocks(ctx context.Context) <-chan uint64
}

// MezoBridgeContract represents a handle to the MezoBridge smart contract.
type MezoBridgeContract interface {
	PastAssetsUnlockConfirmedEvents(
		startBlock uint64,
		endBlock *uint64,
		unlockSequenceNumber []*big.Int,
		recipient [][]byte,
		token []common.Address,
	) ([]*portal.MezoBridgeAssetsUnlockConfirmed, error)
	TbtcToken() (common.Address, error)
	PendingBTCWithdrawals([32]byte) (bool, error)
	WithdrawBTC(
		entry portal.MezoBridgeAssetsUnlocked,
		walletPubKeyHash [20]byte,
		mainUtxo portal.BitcoinTxUTXO,
	) (*types.Transaction, error)
}

// TbtcBridgeContract represents a handle to the tBTC Bridge smart contract.
type TbtcBridgeContract interface {
	PastNewWalletRegisteredEvents(
		startBlock uint64,
		endBlock *uint64,
		ecdsaWalletID [][32]byte,
		walletPubKeyHash [][20]byte,
	) ([]*tbtc.BridgeNewWalletRegistered, error)
	Wallets(walletPublicKeyHash [20]byte) (tbtc.Wallet, error)
	PendingRedemptions(redemptionKey *big.Int) (tbtc.RedemptionRequest, error)
	RedemptionDustThreshold() (uint64, error)
}

// bridgeWorkerJob represent a bridge worker job.
type bridgeWorkerJob interface {
	run(ctx context.Context)
}

// environment groups common elements for all bridge worker jobs.
type environment struct {
	logger log.Logger

	mezoBridgeContract MezoBridgeContract
	tbtcBridgeContract TbtcBridgeContract

	chain EthereumChain

	batchSize         uint64
	requestsPerMinute uint64

	btcChain bitcoin.Chain

	server *Server
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
	mezoBridgeContractBindings, err := initializeMezoBridgeContract(common.HexToAddress(mezoBridgeAddress), chain)
	if err != nil {
		panic(fmt.Sprintf("failed to initialize MezoBridge contract bindings: %v", err))
	}

	mezoBridgeContract := ethereum.NewMezoBridgeContract(mezoBridgeContractBindings)

	// Initialize the tBTC Bridge contract instance.
	tbtcBridgeContractBindings, err := initializeTbtcBridgeContract(common.HexToAddress(tbtcBridgeAddress), chain)
	if err != nil {
		panic(fmt.Sprintf("failed to initialize tBTC Bridge contract bindings: %v", err))
	}

	tbtcBridgeContract := ethereum.NewTbtcBridgeContract(tbtcBridgeContractBindings)

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

	// The messages handled by the bridge-worker contain custom types.
	// Add codecs so that the messages can be marshaled/unmarshalled.
	assetsUnlockEndpoint, err := NewAssetsUnlockedGrpcEndpoint(
		cfg.Mezo.AssetsUnlockEndpoint,
		codectypes.NewInterfaceRegistry(),
	)
	if err != nil {
		panic(fmt.Sprintf("failed to create assets unlocked endpoint: %v", err))
	}

	store, err := NewSupabaseStore(logger, cfg.Supabase.URL, cfg.Supabase.Key)
	if err != nil {
		panic(fmt.Sprintf("couldn't initialize supabase store: %v", err))
	}

	env := &environment{
		logger:             logger,
		mezoBridgeContract: mezoBridgeContract,
		tbtcBridgeContract: tbtcBridgeContract,
		chain:              chain,
		batchSize:          cfg.Ethereum.BatchSize,
		requestsPerMinute:  cfg.Ethereum.RequestsPerMinute,
		btcChain:           btcChain,
		server:             NewServer(logger, cfg.Server.Port, chain.ChainID(), mezoBridgeContract, store, assetsUnlockEndpoint),
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

	go func() {
		defer cancelCtx()
		env.server.Start()
		env.logger.Warn("Http server stopped")
	}()

	<-ctx.Done()

	err = env.server.Stop(ctx)
	if err != nil {
		env.logger.Error("couldn't shutdown the http server properly", "error", err)
	}

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
