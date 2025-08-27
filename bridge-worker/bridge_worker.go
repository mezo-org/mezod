package bridgeworker

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/common"

	ethconfig "github.com/keep-network/keep-common/pkg/chain/ethereum"
	ethconnect "github.com/mezo-org/mezod/ethereum"
	"github.com/mezo-org/mezod/ethereum/bindings/portal"
)

// mezoBridgeName is the name of the MezoBridge contract.
var mezoBridgeName = "MezoBridge"

type BridgeWorker struct {
	bridgeContract *portal.MezoBridge
}

func RunBridgeWorker(
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

	log.Printf(
		"resolved MezoBridge: %s and ethereum_network: %s",
		mezoBridgeAddress,
		network,
	)

	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	var err error
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
		bridgeContract: bridgeContractBinding,
	}

	go func() {
		defer cancelCtx()
		err := bw.handleBitcoinWithdrawing(ctx)
		if err != nil {
			log.Printf("Bitcoin withdrawing routine failed: %v", err)
		}

		log.Printf("Bitcoin withdrawing routine stopped")
	}()

	<-ctx.Done()

	log.Print("bridge worker stopped")
}

func (bw *BridgeWorker) handleBitcoinWithdrawing(ctx context.Context) error {
	// TODO: Implement; log for now.
	log.Print("Inside Bitcoin withdrawal logic")
	<-ctx.Done()
	return ctx.Err()
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
