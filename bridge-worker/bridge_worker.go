package bridgeworker

import (
	"context"
	"crypto/ecdsa"
	"fmt"

	"cosmossdk.io/log"

	"github.com/ethereum/go-ethereum/common"
	ethconfig "github.com/keep-network/keep-common/pkg/chain/ethereum"
	ethconnect "github.com/mezo-org/mezod/ethereum"
	"github.com/mezo-org/mezod/ethereum/bindings/portal"
)

// mezoBridgeName is the name of the MezoBridge contract.
const mezoBridgeName = "MezoBridge"

type BridgeWorker struct {
	logger log.Logger

	bridgeContract *portal.MezoBridge
	chain          *ethconnect.BaseChain

	batchSize         uint64
	requestsPerMinute uint64
}

func RunBridgeWorker(
	logger log.Logger,
	providerURL string,
	ethereumNetwork string,
	privateKey *ecdsa.PrivateKey,
) {
	network := ethconnect.NetworkFromString(ethereumNetwork)
	mezoBridgeAddress := portal.MezoBridgeAddress(network)

	if mezoBridgeAddress == "" {
		panic(
			"cannot get address of the MezoBridge contract on Ethereum; " +
				"make sure you run 'make bindings' before building the binary",
		)
	}

	logger.Info(
		"resolved MezoBridge contract and Ethereum network",
		"mezo_bridge_address", mezoBridgeAddress,
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
	bridgeContractBinding, err := initializeBridgeContract(common.HexToAddress(mezoBridgeAddress), chain)
	if err != nil {
		panic(fmt.Sprintf("failed to initialize MezoBridge contract: %v", err))
	}

	bw := &BridgeWorker{
		logger:            logger,
		bridgeContract:    bridgeContractBinding,
		chain:             chain,
		batchSize:         defaultBatchSize,
		requestsPerMinute: defaultRequestsPerMinute,
	}

	go func() {
		defer cancelCtx()
		err := bw.handleBitcoinWithdrawing(ctx)
		if err != nil {
			bw.logger.Info(
				"Bitcoin withdrawing routine failed",
				"err", err,
			)
		}

		bw.logger.Info("Bitcoin withdrawing routine stopped")
	}()

	<-ctx.Done()

	bw.logger.Info("bridge worker stopped")
}

// Construct a new instance of the Ethereum MezoBridge contract.
func initializeBridgeContract(
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
