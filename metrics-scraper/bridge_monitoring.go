package metricsscraper

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/big"
	"slices"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	keepethereum "github.com/keep-network/keep-common/pkg/chain/ethereum"
	mezodethereum "github.com/mezo-org/mezod/ethereum"
	"github.com/mezo-org/mezod/ethereum/bindings/portal"
	"github.com/mezo-org/mezod/ethereum/bindings/tbtc"
	mezotypes "github.com/mezo-org/mezod/types"
	"github.com/mezo-org/mezod/utils"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
)

const (
	mezoAssetsUnlockedLookBackBlocks = uint64(172800) // ~7 days (assuming 1 block per 3.5s)
	// We make the look-back period longer for Ethereum than for Mezo to make sure
	// we don't miss any AssetsUnlockedConfirmed events that indiciate the completion
	// of the attestation process on Ethereum. Missing those events could result
	// in false positives in the pending_assets_unlocked metric.
	ethereumAssetsUnlockedConfirmedLookBackBlocks = uint64(100800) // ~14 days (assuming 1 block per 12s)
)

var pendingAssetsUnlockedCache = map[string]bool{} // unlock seqno -> bool

type ethereumChain struct {
	client          *mezodethereum.BaseChain
	mezoBridge      *portal.MezoBridge
	tbtcBankAddress common.Address

	assetsUnlockedConfirmedCheckpointHeight uint64
}

type mezoChain struct {
	client       *ethclient.Client
	assetsBridge *AssetsBridge

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

	tbtcBridgeAddress := tbtc.BridgeAddress(ethereumNetwork)
	if len(tbtcBridgeAddress) == 0 {
		// If this happened, bindings are broken and the only option is to panic.
		panic("cannot get address of the tBTC Bridge contract on Ethereum")
	}

	tbtcBridge, err := tbtc.NewTbtcBridge(
		common.HexToAddress(tbtcBridgeAddress),
		client.ChainID(),
		client.Key(),
		client.Client(),
		client.NonceManager(),
		client.MiningWaiter(),
		client.BlockCounter(),
		client.TransactionMutex(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to attach to tBTC Bridge contract: [%w]", err)
	}

	tbtcContractReferences, err := tbtcBridge.ContractReferences()
	if err != nil {
		return nil, fmt.Errorf("failed to get contract references from tBTC Bridge contract: [%w]", err)
	}

	return &ethereumChain{
		client:          client,
		mezoBridge:      mezoBridge,
		tbtcBankAddress: tbtcContractReferences.Bank,
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
		client:       client,
		assetsBridge: assetsBridge,
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

	err = ethereumBridgeValsBalance(
		ctx,
		mezoChainID,
		ethereum,
	)
	if err != nil {
		errs = append(errs, err)
	}

	err = tbtcRedeemerBankBalance(
		ctx,
		mezoChainID,
		ethereum,
	)
	if err != nil {
		errs = append(errs, err)
	}

	err = outflowLimitSaturation(
		mezoChainID,
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
			log.Printf("error while determining pending AssetsLocked: [%v]", err)
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

	pending, _ := new(big.Int).Sub(mezoBridgeSequence, assetsBridgeSequence).Float64()

	log.Printf(
		"MezoBridge lock sequence tip: [%d]; "+
			"AssetsBridge lock sequence tip: [%d]; "+
			"pending AssetsLocked count: [%d]",
		mezoBridgeSequence,
		assetsBridgeSequence,
		int(pending),
	)

	pendingAssetsLockedGauge.WithLabelValues(mezoChainID).Set(pending)

	return
}

func pendingAssetsUnlocked(
	ctx context.Context,
	mezoChainID string,
	ethereum *ethereumChain,
	mezo *mezoChain,
) (err error) {
	defer func() {
		if err != nil {
			log.Printf("error while determining pending AssetsUnlocked: [%v]", err)
			// set gauge to -1 to indicate error
			pendingAssetsUnlockedGauge.WithLabelValues(mezoChainID).Set(-1)
		}
	}()

	requestedUnlockSeqnos, err := mezo.requestedUnlockSeqnos(ctx)
	if err != nil {
		err = fmt.Errorf(
			"failed to get requested unlock sequence numbers from Mezo: [%w]",
			err,
		)
		//nolint:nakedret
		return err
	}

	// Populate the cache with all requested unlock seqnos.
	for _, seqno := range requestedUnlockSeqnos {
		pendingAssetsUnlockedCache[seqno.String()] = true
	}

	processedUnlockSeqnos, err := ethereum.processedUnlockSeqnos(ctx)
	if err != nil {
		err = fmt.Errorf(
			"failed to get processed unlock sequence numbers from Ethereum: [%w]",
			err,
		)
		//nolint:nakedret
		return
	}

	// Remove processed unlock seqnos from the cache.
	for _, seqno := range processedUnlockSeqnos {
		delete(pendingAssetsUnlockedCache, seqno.String())
	}

	// Prepare a list of pending seqnos for logging purposes.
	pendingSeqnos := make([]*big.Int, 0, len(pendingAssetsUnlockedCache))
	for seqno := range pendingAssetsUnlockedCache {
		seqnoBigInt, _ := new(big.Int).SetString(seqno, 10)
		pendingSeqnos = append(pendingSeqnos, seqnoBigInt)
	}
	slices.SortFunc(pendingSeqnos, func(a, b *big.Int) int {
		return a.Cmp(b)
	})

	// Calculate the number of pending seqnos for metrics.
	pending := len(pendingAssetsUnlockedCache)

	log.Printf(
		"fetched [%d] requested and [%d] processed AssetsUnlocked entries; "+
			"pending AssetsUnlocked count: [%d] (seqnos: %s)",
		len(requestedUnlockSeqnos),
		len(processedUnlockSeqnos),
		pending,
		pendingSeqnos,
	)

	pendingAssetsUnlockedGauge.WithLabelValues(mezoChainID).Set(float64(pending))

	//nolint:nakedret
	return
}

func (mc *mezoChain) requestedUnlockSeqnos(ctx context.Context) ([]*big.Int, error) {
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

	unlockSequenceNumbers := make([]*big.Int, len(events))
	for i, event := range events {
		unlockSequenceNumbers[i] = event.UnlockSequenceNumber
	}

	slices.SortFunc(unlockSequenceNumbers, func(a, b *big.Int) int {
		return a.Cmp(b)
	})

	return unlockSequenceNumbers, nil
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

func (ec *ethereumChain) processedUnlockSeqnos(ctx context.Context) ([]*big.Int, error) {
	currentHeight, err := ec.client.LatestBlock(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get current Ethereum height: [%w]", err)
	}

	startHeight := ec.assetsUnlockedConfirmedCheckpointHeight
	endHeight := currentHeight.Uint64()

	if startHeight == 0 {
		if endHeight > ethereumAssetsUnlockedConfirmedLookBackBlocks {
			startHeight = endHeight - ethereumAssetsUnlockedConfirmedLookBackBlocks
		}
	}

	log.Printf(
		"getting AssetsUnlockedConfirmed events from [%d] to [%d] on Ethereum chain",
		startHeight,
		endHeight,
	)

	events, err := withBatchFetch(ec.assetsUnlockConfirmedEvents, startHeight, endHeight)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to get AssetsUnlockedConfirmed events on Ethereum chain: [%w]",
			err,
		)
	}

	ec.assetsUnlockedConfirmedCheckpointHeight = endHeight + 1

	unlockSequenceNumbers := make([]*big.Int, len(events))
	for i, event := range events {
		unlockSequenceNumbers[i] = event.UnlockSequenceNumber
	}

	slices.SortFunc(unlockSequenceNumbers, func(a, b *big.Int) int {
		return a.Cmp(b)
	})

	return unlockSequenceNumbers, nil
}

func (ec *ethereumChain) assetsUnlockConfirmedEvents(
	startHeight, endHeight uint64,
) ([]*portal.MezoBridgeAssetsUnlockConfirmed, error) {
	return ec.mezoBridge.PastAssetsUnlockConfirmedEvents(startHeight, &endHeight, nil, nil, nil)
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

func ethereumBridgeValsBalance(
	ctx context.Context,
	mezoChainID string,
	ethereum *ethereumChain,
) error {
	validatorsCount, err := ethereum.mezoBridge.BridgeValidatorsCount()
	if err != nil {
		return fmt.Errorf("failed to get bridge validators count from Ethereum: [%w]", err)
	}

	ethereumBridgeValBalanceGauge.Reset()

	for i := uint64(0); i < validatorsCount.Uint64(); i++ {
		//nolint:gosec
		validator, err := ethereum.mezoBridge.BridgeValidators(big.NewInt(int64(i)))
		if err != nil {
			return fmt.Errorf(
				"failed to get bridge validator [%d] address from Ethereum: [%w]",
				i,
				err,
			)
		}

		balance, err := ethereum.client.Client().BalanceAt(ctx, validator, nil)
		if err != nil {
			return fmt.Errorf(
				"failed to get bridge validator [%s] balance from Ethereum: [%w]",
				validator.Hex(),
				err,
			)
		}

		ethValue := new(big.Float).Quo(new(big.Float).SetInt(balance), big.NewFloat(1e18))
		ethValueFloat, _ := ethValue.Float64()

		ethereumBridgeValBalanceGauge.WithLabelValues(validator.Hex(), mezoChainID).Set(ethValueFloat)
	}

	return nil
}

func tbtcRedeemerBankBalance(
	ctx context.Context,
	mezoChainID string,
	ethereum *ethereumChain,
) (err error) {
	defer func() {
		if err != nil {
			log.Printf("error while determining tBTC Redeemer bank balance: [%v]", err)
			// set gauge to -1 to indicate error
			tbtcRedeemerBankBalanceGauge.WithLabelValues(mezoChainID).Set(-1)
		}
	}()

	tbtcRedeemer, err := ethereum.mezoBridge.TbtcRedeemer()
	if err != nil {
		err = fmt.Errorf("failed to get tBTC redeemer address from Ethereum: [%w]", err)
		return
	}

	balance, err := ethereum.tbtcBankBalanceOf(ctx, tbtcRedeemer)
	if err != nil {
		err = fmt.Errorf("failed to get tBTC Bank balance from Ethereum: [%w]", err)
		return
	}

	btcValue := new(big.Float).Quo(new(big.Float).SetInt(balance), big.NewFloat(1e8))
	btcValueFloat, _ := btcValue.Float64()

	tbtcRedeemerBankBalanceGauge.WithLabelValues(mezoChainID).Set(btcValueFloat)

	return
}

func (ec *ethereumChain) tbtcBankBalanceOf(ctx context.Context, account common.Address) (*big.Int, error) {
	// Create the balanceOf function call data
	// Function signature: balanceOf(address) -> uint256
	balanceOfSig := crypto.Keccak256([]byte("balanceOf(address)"))[:4]

	// Encode the address parameter (pad to 32 bytes)
	addressParam := common.LeftPadBytes(account.Bytes(), 32)

	// Combine function signature and parameters
	//nolint:gocritic
	callData := append(balanceOfSig, addressParam...)

	// Create the call message
	msg := ethereum.CallMsg{
		To:   &ec.tbtcBankAddress,
		Data: callData,
	}

	// Make the contract call
	result, err := ec.client.Client().CallContract(ctx, msg, nil)
	if err != nil {
		return nil, fmt.Errorf("tBTC Bank balanceOf contract call failed: %w", err)
	}

	// Convert the result to *big.Int
	balance := new(big.Int).SetBytes(result)

	return balance, nil
}

func outflowLimitSaturation(
	mezoChainID string,
	mezo *mezoChain,
) error {
	outflowLimitSaturationGauge.Reset()

	mappings, err := mezo.assetsBridge.GetERC20TokensMappings(nil)
	if err != nil {
		return fmt.Errorf("failed to get ERC20 tokens mappings from Mezo: [%w]", err)
	}

	mezoTokens := []common.Address{common.HexToAddress(evmtypes.BTCTokenPrecompileAddress)}
	for _, mapping := range mappings {
		mezoTokens = append(mezoTokens, mapping.MezoToken)
	}

	for _, mezoToken := range mezoTokens {
		capacityData, err := mezo.assetsBridge.GetOutflowCapacity(nil, mezoToken)
		if err != nil {
			return fmt.Errorf("failed to get outflow capacity for token [%s]: [%w]", mezoToken, err)
		}

		limit, err := mezo.assetsBridge.GetOutflowLimit(nil, mezoToken)
		if err != nil {
			return fmt.Errorf("failed to get outflow limit from for token [%s]: [%w]", mezoToken, err)
		}

		capacityFloat := new(big.Float).SetInt(capacityData.Capacity)
		limitFloat := new(big.Float).SetInt(limit)
		limitUsed := new(big.Float).Sub(limitFloat, capacityFloat)

		saturation := new(big.Float).Quo(new(big.Float).Mul(limitUsed, big.NewFloat(100)), limitFloat)
		saturationFloat, _ := saturation.Float64()

		outflowLimitSaturationGauge.WithLabelValues(mezoToken.Hex(), mezoChainID).Set(saturationFloat)
	}

	return nil
}
