package metricsscraper

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	keepethereum "github.com/keep-network/keep-common/pkg/chain/ethereum"
	mezodethereum "github.com/mezo-org/mezod/ethereum"
	"github.com/mezo-org/mezod/ethereum/bindings/portal"
	mezotypes "github.com/mezo-org/mezod/types"
	"github.com/mezo-org/mezod/utils"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
)

func runBridgeMonitoring(
	ctx context.Context,
	chainID string,
	pollRate time.Duration,
	mezoRPCURL string,
	ethereumRPCURL string,
) error {
	log.Printf("starting bridge monitoring")

	mezoBridge, err := connectMezoBridgeEthereumContract(
		ctx,
		mapMezoChainIDToEthereumNetwork(chainID),
		ethereumRPCURL,
	)
	if err != nil {
		return fmt.Errorf(
			"failed to connect to MezoBridge contract on Ethereum: [%w]",
			err,
		)
	}

	assetsBridge, err := connectAssetsBridgeMezoContract(
		ctx,
		chainID,
		mezoRPCURL,
	)
	if err != nil {
		return fmt.Errorf(
			"failed to connect to AssetsBridge contract on Mezo: [%w]",
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
			if err := pollBridgeData(
				ctx,
				chainID,
				mezoBridge,
				assetsBridge,
			); err != nil {
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
	ethereumRPCURL string,
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
			URL:     ethereumRPCURL,
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

func connectAssetsBridgeMezoContract(
	ctx context.Context,
	chainID string,
	mezoRPCURL string,
) (*AssetsBridge, error) {
	client, err := ethclient.DialContext(ctx, mezoRPCURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Mezo network: [%w]", err)
	}

	eip155ChainID, err := mezotypes.ParseChainID(chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse chain ID to EIP155 format: [%w]", err)
	}

	clientChainID, err := client.ChainID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get chain ID from Mezo RPC: [%w]", err)
	}

	if eip155ChainID.Cmp(clientChainID) != 0 {
		return nil, fmt.Errorf("mezo RPC uses different chain ID than expected")
	}

	assetsBridge, err := NewAssetsBridge(
		common.HexToAddress(evmtypes.AssetsBridgePrecompileAddress),
		client,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to attach to AssetsBridge contract: [%w]", err)
	}

	return assetsBridge, nil
}

func mapMezoChainIDToEthereumNetwork(chainID string) keepethereum.Network {
	switch {
	case utils.IsMainnet(chainID):
		return keepethereum.Mainnet
	case utils.IsTestnet(chainID):
		return keepethereum.Sepolia
	default:
		panic(fmt.Sprintf("unknown Mezo chain id: %s", chainID))
	}
}

func pollBridgeData(
	ctx context.Context,
	chainID string,
	mezoBridge *portal.MezoBridge,
	assetsBridge *AssetsBridge,
) error {
	errs := []error{}

	err := pendingAssetsLocked(
		ctx,
		chainID,
		mezoBridge,
		assetsBridge,
	)
	if err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

func pendingAssetsLocked(
	ctx context.Context,
	chainID string,
	mezoBridge *portal.MezoBridge,
	assetsBridge *AssetsBridge,
) (err error) {
	defer func() {
		if err != nil {
			// set gauge to -1 to indicate error
			pendingAssetsLockedGauge.WithLabelValues(chainID).Set(-1)
		}
	}()

	mezoBridgeSequence, err := mezoBridge.Sequence()
	if err != nil {
		err = fmt.Errorf("failed to get MezoBridge sequence: [%w]", err)
		return
	}

	assetsBridgeSequence, err := assetsBridge.GetCurrentSequenceTip(nil)
	if err != nil {
		err = fmt.Errorf("failed to get AssetsBridge sequence: [%w]", err)
		return
	}

	pending, _ := new(big.Int).Sub(mezoBridgeSequence, assetsBridgeSequence).Float64()

	pendingAssetsLockedGauge.WithLabelValues(chainID).Set(pending)

	return
}
