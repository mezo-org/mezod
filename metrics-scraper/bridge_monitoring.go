package metricsscraper

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	keepethereum "github.com/keep-network/keep-common/pkg/chain/ethereum"
	mezodethereum "github.com/mezo-org/mezod/ethereum"
	"github.com/mezo-org/mezod/ethereum/bindings/portal"
	"github.com/mezo-org/mezod/utils"
)

func runBridgeMonitoring(
	ctx context.Context,
	chainID string,
	pollRate time.Duration,
	mezoRpcUrl string,
	ethereumRpcUrl string,
) error {
	log.Printf("starting bridge monitoring")

	_, err := connectMezoBridgeEthereumContract(
		ctx,
		mapMezoChainIdToEthereumNetwork(chainID),
		ethereumRpcUrl,
	)
	if err != nil {
		return fmt.Errorf(
			"failed to connect to MezoBridge contract on Ethereum: [%w]",
			err,
		)
	}

	ticker := time.NewTicker(pollRate)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Printf("terminated bridge monitoring")
			return ctx.Err()
		case <-ticker.C:
			if err := pollBridgeData(ctx, chainID); err != nil {
				log.Printf("error while polling bridge data: %v", err)
			} else {
				log.Printf("bridge data polled successfully")
			}
		}
	}
}

func connectMezoBridgeEthereumContract(
	ctx context.Context,
	ethereumNetwork keepethereum.Network,
	ethereumRpcUrl string,
) (*portal.MezoBridge, error) {
	mezoBridgeAddress := portal.MezoBridgeAddress(ethereumNetwork)
	if len(mezoBridgeAddress) == 0 {
		// If this happened, bindings are broken and the only option is to panic.
		panic("cannot get address of the MezoBridge contract on Ethereum")
	}

	log.Printf(
		"resolved Ethereum network [%s] and MezoBridge contract address [%s]",
		ethereumNetwork,
		mezoBridgeAddress,
	)

	// Generate a dummy private key to use it with Connect. If we pass nil,
	// Connect will panic.
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate dummy private key: [%w]", err)
	}

	ethereumChain, err := mezodethereum.Connect(
		ctx,
		keepethereum.Config{
			Network: ethereumNetwork,
			URL:     ethereumRpcUrl,
		},
		privateKey,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the Ethereum network: [%w]", err)
	}

	mezoBridge, err := portal.NewMezoBridge(
		common.HexToAddress(mezoBridgeAddress),
		ethereumChain.ChainID(),
		ethereumChain.Key(),
		ethereumChain.Client(),
		ethereumChain.NonceManager(),
		ethereumChain.MiningWaiter(),
		ethereumChain.BlockCounter(),
		ethereumChain.TransactionMutex(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to attach to MezoBridge contract: [%w]", err)
	}

	return mezoBridge, nil
}

func mapMezoChainIdToEthereumNetwork(chainID string) keepethereum.Network {
	if utils.IsMainnet(chainID) {
		return keepethereum.Mainnet
	} else if utils.IsTestnet(chainID) {
		return keepethereum.Sepolia
	} else {
		panic(fmt.Sprintf("unknown Mezo chain id: %s", chainID))
	}
}

func pollBridgeData(_ context.Context, _ string) error {
	// TODO: Implement bridge data polling and expose metrics.
	return nil
}
