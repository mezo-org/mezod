package metricsscraper

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/big"
	"slices"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
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

const (
	mezoAssetsUnlockedLookBackBlocks = uint64(172800) // ~7 days (assuming 1 block per 3.5s)
)

type ethereumChain struct {
	client     *mezodethereum.BaseChain
	mezoBridge *portal.MezoBridge
}

type mezoChain struct {
	client       *ethclient.Client
	assetsBridge *AssetsBridge

	assetsUnlockSequenceTip        *big.Int
	assetsUnlockedCheckpointHeight uint64
}

func runBridgeMonitoring(
	ctx context.Context,
	mezoChainID string,
	pollRate time.Duration,
	mezoRPCURL string,
	ethereumRPCURL string,
) error {
	log.Printf("starting bridge monitoring")

	ethereum, err := connectEthereum(
		ctx,
		mapMezoChainIDToEthereumNetwork(mezoChainID),
		ethereumRPCURL,
	)
	if err != nil {
		return fmt.Errorf(
			"failed to connect to MezoBridge contract on Ethereum: [%w]",
			err,
		)
	}

	mezo, err := connectMezo(
		ctx,
		mezoChainID,
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
				mezoChainID,
				ethereum,
				mezo,
			); err != nil {
				log.Printf("error while polling bridge data: %v", err)
			} else {
				log.Printf("bridge data polled successfully")
			}
		}
	}
}

func connectEthereum(
	ctx context.Context,
	ethereumNetwork keepethereum.Network,
	ethereumRPCURL string,
) (*ethereumChain, error) {
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

	client, err := mezodethereum.Connect(
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
		client.ChainID(),
		client.Key(),
		client.Client(),
		client.NonceManager(),
		client.MiningWaiter(),
		client.BlockCounter(),
		client.TransactionMutex(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to attach to MezoBridge contract: [%w]", err)
	}

	return &ethereumChain{
		client:     client,
		mezoBridge: mezoBridge,
	}, nil
}

func connectMezo(
	ctx context.Context,
	mezoChainID string,
	mezoRPCURL string,
) (*mezoChain, error) {
	client, err := ethclient.DialContext(ctx, mezoRPCURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Mezo network: [%w]", err)
	}

	eip155MezoChainID, err := mezotypes.ParseChainID(mezoChainID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Mezo chain ID to EIP155 format: [%w]", err)
	}

	clientMezoChainID, err := client.ChainID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get chain ID from Mezo RPC: [%w]", err)
	}

	if eip155MezoChainID.Cmp(clientMezoChainID) != 0 {
		return nil, fmt.Errorf("mezo RPC uses different chain ID than expected")
	}

	assetsBridge, err := NewAssetsBridge(
		common.HexToAddress(evmtypes.AssetsBridgePrecompileAddress),
		client,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to attach to AssetsBridge contract: [%w]", err)
	}

	return &mezoChain{
		client:                  client,
		assetsBridge:            assetsBridge,
		assetsUnlockSequenceTip: big.NewInt(0),
	}, nil
}

func mapMezoChainIDToEthereumNetwork(mezoChainID string) keepethereum.Network {
	switch {
	case utils.IsMainnet(mezoChainID):
		return keepethereum.Mainnet
	case utils.IsTestnet(mezoChainID):
		return keepethereum.Sepolia
	default:
		panic(fmt.Sprintf("unknown Mezo chain id: %s", mezoChainID))
	}
}

func pollBridgeData(
	ctx context.Context,
	mezoChainID string,
	ethereum *ethereumChain,
	mezo *mezoChain,
) error {
	errs := []error{}

	err := pendingAssetsLocked(
		mezoChainID,
		ethereum,
		mezo,
	)
	if err != nil {
		errs = append(errs, err)
	}

	err = pendingAssetsUnlocked(
		ctx,
		mezoChainID,
		ethereum,
		mezo,
	)
	if err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

func pendingAssetsLocked(
	mezoChainID string,
	ethereum *ethereumChain,
	mezo *mezoChain,
) (err error) {
	defer func() {
		if err != nil {
			// set gauge to -1 to indicate error
			pendingAssetsLockedGauge.WithLabelValues(mezoChainID).Set(-1)
		}
	}()

	mezoBridgeSequence, err := ethereum.mezoBridge.Sequence()
	if err != nil {
		err = fmt.Errorf("failed to get MezoBridge lock sequence tip: [%w]", err)
		return
	}

	assetsBridgeSequence, err := mezo.assetsBridge.GetCurrentSequenceTip(nil)
	if err != nil {
		err = fmt.Errorf("failed to get AssetsBridge lock sequence tip: [%w]", err)
		return
	}

	log.Printf(
		"MezoBridge lock sequence tip: [%d]; AssetsBridge lock sequence tip: [%d]",
		mezoBridgeSequence,
		assetsBridgeSequence,
	)

	pending, _ := new(big.Int).Sub(mezoBridgeSequence, assetsBridgeSequence).Float64()

	pendingAssetsLockedGauge.WithLabelValues(mezoChainID).Set(pending)

	return
}

func pendingAssetsUnlocked(
	ctx context.Context,
	mezoChainID string,
	_ *ethereumChain,
	mezo *mezoChain,
) (err error) {
	defer func() {
		if err != nil {
			// set gauge to -1 to indicate error
			pendingAssetsUnlockedGauge.WithLabelValues(mezoChainID).Set(-1)
		}
	}()

	assetsBridgeSequence, err := mezo.assetsUnlockedSequenceTip(ctx)
	if err != nil {
		return fmt.Errorf("failed to get AssetsBridge unlock sequence tip: [%w]", err)
	}

	// TODO: Determine the MezoBridge seqno tip.
	mezoBridgeSequence := big.NewInt(0)

	log.Printf(
		"AssetsBridge unlock sequence tip: [%d]; MezoBridge unlock sequence tip: [%d]",
		assetsBridgeSequence,
		mezoBridgeSequence,
	)

	pending, _ := new(big.Int).Sub(assetsBridgeSequence, mezoBridgeSequence).Float64()

	pendingAssetsUnlockedGauge.WithLabelValues(mezoChainID).Set(pending)

	return nil
}

func (mc *mezoChain) assetsUnlockedSequenceTip(ctx context.Context) (*big.Int, error) {
	currentHeader, err := mc.client.HeaderByNumber(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get current Mezo height: [%w]", err)
	}

	startHeight := mc.assetsUnlockedCheckpointHeight
	endHeight := currentHeader.Number.Uint64()

	if startHeight == 0 {
		if endHeight > mezoAssetsUnlockedLookBackBlocks {
			startHeight = endHeight - mezoAssetsUnlockedLookBackBlocks
		}
	}

	log.Printf(
		"getting AssetsUnlocked events from [%d] to [%d] on Mezo chain",
		startHeight,
		endHeight,
	)

	events, err := withBatchFetch(mc.assetsUnlockedEvents, startHeight, endHeight)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to get AssetsUnlocked events on Mezo chain: [%w]",
			err,
		)
	}

	mc.assetsUnlockedCheckpointHeight = endHeight + 1

	slices.SortFunc(events, func(a, b *AssetsBridgeAssetsUnlocked) int {
		return a.UnlockSequenceNumber.Cmp(b.UnlockSequenceNumber)
	})

	newTip := big.NewInt(0)
	if len(events) > 0 {
		newTip = events[len(events)-1].UnlockSequenceNumber
	}

	if newTip.Cmp(mc.assetsUnlockSequenceTip) > 0 {
		mc.assetsUnlockSequenceTip = newTip
	}

	return mc.assetsUnlockSequenceTip, nil
}

func (mc *mezoChain) assetsUnlockedEvents(
	startHeight, endHeight uint64,
) ([]*AssetsBridgeAssetsUnlocked, error) {
	iterator, err := mc.assetsBridge.FilterAssetsUnlocked(
		&bind.FilterOpts{
			Start: startHeight,
			End:   &endHeight,
		},
		nil,
		nil,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to get AssetsUnlocked events on Mezo chain: [%w]",
			err,
		)
	}
	defer iterator.Close()

	events := make([]*AssetsBridgeAssetsUnlocked, 0)
	for iterator.Next() {
		events = append(events, iterator.Event)
	}

	return events, nil
}

func withBatchFetch[T any](
	fetchFunc func(batchStartHeight, batchEndHeight uint64) ([]T, error),
	startHeight, endHeight uint64,
) ([]T, error) {
	batchSize := uint64(1000)
	batchStartHeight := startHeight
	result := make([]T, 0)

	for batchStartHeight <= endHeight {
		batchEndHeight := min(batchStartHeight+batchSize, endHeight)

		batchEvents, batchErr := fetchFunc(batchStartHeight, batchEndHeight)
		if batchErr != nil {
			return nil, fmt.Errorf(
				"batched event fetch failed: [%w]",
				batchErr,
			)
		}

		result = append(result, batchEvents...)

		batchStartHeight = batchEndHeight + 1
	}

	return result, nil
}
