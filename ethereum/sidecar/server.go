package sidecar

import (
	"context"
	"fmt"
	"math/big"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	pb "github.com/mezo-org/mezod/ethereum/sidecar/types"
	bridgetypes "github.com/mezo-org/mezod/x/bridge/types"

	"github.com/ethereum/go-ethereum/common"
	ethconfig "github.com/keep-network/keep-common/pkg/chain/ethereum"
	ethconnect "github.com/mezo-org/mezod/ethereum"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/mezo-org/mezod/ethereum/bindings/portal/gen/abi"

	"github.com/ethereum/go-ethereum"

	"github.com/ethereum/go-ethereum/core/types"

	"github.com/ethereum/go-ethereum/ethclient"
)

var (
	// ErrSequenceStartNotLower is the error returned when the start of
	// the sequence is not lower than sequence end.
	ErrSequenceStartNotLower = fmt.Errorf(
		"sequence start is not lower than sequence end",
	)

	BitcoinBridgeName = "bitcoinbridge"
	// Number of blocks to look back for AssetsLocked events in the BitcoinBridge
	// contract. This is approximately 2 weeks window for 12 sec. block time.

	SearchedRange = new(big.Int).SetInt64(100000)
	// Size of Sequence: 32 bytes
	// Size of Recipient: 42 bytes
	// Size of Amount: 32 bytes
	// Struct overhead and padding: ~16bytes
	// Total size of one event: 122 bytes (0.12KB)
	// Assuiming we want to allocate ~64MB for the cache, we can store ~546k events.
	// For simplicity, let's make 500k events our limit. Even if deposits hit 1000
	// daily mark, that would give 50 days of cached events which should be more than
	// enough.

	CachedEvents = 500000
	// Number of blocks that have to pass before an AssetsLocked event can be confirmed
	// in a finalized block. Ideally a block is produced in every slot which takes
	// 12 sec. There are 32 slots in Epoch. A block is considered finalized
	// after 2 passed epochs which gives 64 slots. We can have some buffer here in
	// case a block is not produced in a given slot, hence the number is 70, not 64.
	// Since it takes 12 sec to create a block, it will take around 14 min to confirm
	// a new AssetsLocked event in this sidecar.
	ConfirmationBlocks = 70

	// Time in seconds needed to produce a new block on Ethereum PoS mainnet.
	BlockTime = 12 * time.Second
)

// Server observes events emitted by the Mezo `BitcoinBridge` contract on the
// Ethereum chain. It enables retrieval of information on the assets locked by
// the contract. It is intended to be run as a separate process.
type Server struct {
	sequenceTip sdkmath.Int
	eventsMutex sync.RWMutex
	// TODO: When we add the real implementation of the server, make sure the
	//       `AssetsLocked` events returned from the server will pass the
	//       validation in the Ethereum server client.
	events       []abi.BitcoinBridgeAssetsLocked
	bufferEvents []abi.BitcoinBridgeAssetsLocked
	grpcServer   *grpc.Server
	logger       log.Logger
}

type Block struct {
	Number string `json:"number"`
}

// RunServer initializes the server, starts the event observing routine and
// starts the gRPC server.
func RunServer(
	ctx context.Context,
	grpcAddress string,
	providerURL string,
	ethereumNetwork string,
	logger log.Logger,
) *Server {
	server := &Server{
		sequenceTip: sdkmath.ZeroInt(),
		events:      make([]abi.BitcoinBridgeAssetsLocked, 0),
		grpcServer:  grpc.NewServer(),
		logger:      logger,
	}

	bitcoinBridgeAddress, err := readBitcoinBridgeAddress()
	if err != nil {
		server.logger.Error("Failed to read the BitcoinBridge address: %v", err)
	}

	// Connect to the Ethereum network
	chain, client, err := ethconnect.Connect(ctx, ethconfig.Config{
		Network:           networkFromString(ethereumNetwork),
		URL:               providerURL,
		ContractAddresses: map[string]string{BitcoinBridgeName: bitcoinBridgeAddress.String()},
	})
	if err != nil {
		server.logger.Error("Failed to connect to the Ethereum network: %v", err)
	}

	go server.observeEvents(ctx, chain, client, bitcoinBridgeAddress)
	go server.startGRPCServer(ctx, grpcAddress)

	return server
}

// observeEvents starts the AssetLocked observing routine.
func (s *Server) observeEvents(ctx context.Context, chain *ethconnect.BaseChain, client *ethclient.Client, bitcoinBridgeAddress common.Address) {
	// Get the BitcoinBridge contract instance
	bitcoinBridge, err := initializeBitcoinBridgeContract(chain, bitcoinBridgeAddress)
	if err != nil {
		s.logger.Error("Failed to initialize BitcoinBridge contract: %v", err)
	}

	currentBlockNumber, err := client.BlockNumber(ctx)
	if err != nil {
		s.logger.Error("Failed to get the current block number: %v", err)
	}

	var finalized Block
	// false for only hashes of the txs, true for full tx objects
	err = client.Client().CallContext(ctx, &finalized, "eth_getBlockByNumber", "finalized", false)
	if err != nil {
		s.logger.Error("Failed to get the finalized block: %v", err)
	}
	finalizedBlock, ok := new(big.Int).SetString(finalized.Number[2:], 16) // hex to decimal
	if !ok {
		s.logger.Error("Failed to convert hexadecimal string")
	}

	s.fetchFinalizedEvents(bitcoinBridge, finalizedBlock)
	s.fetchNonFinalizedEvents(bitcoinBridge, finalizedBlock, currentBlockNumber)

	// Channel to receive new AssetsLocked events
	eventsChan := make(chan *abi.BitcoinBridgeAssetsLocked)
	sub, err := bitcoinBridge.WatchAssetsLocked(&bind.WatchOpts{Context: ctx}, eventsChan, nil, nil)
	if err != nil {
		s.logger.Error("Failed to start watching assets locked events: %v", err)
	}
	defer sub.Unsubscribe()

	tickerChan := s.startBlockTicker(currentBlockNumber)

	for {
		select {
		case err := <-sub.Err():
			s.logger.Error("Error while watching assets locked events: %v", err)
		case event := <-eventsChan:
			s.eventsMutex.Lock()
			s.bufferEvents = append(s.bufferEvents, *event)
			s.eventsMutex.Unlock()
			s.logger.Info(
				"New non-finalized AssetsLocked event",
				"sequence", event.SequenceNumber.String(),
				"recipient", event.Recipient.Hex(),
				"amount", event.TbtcAmount.String(),
			)
		case currentBlockNumber := <-tickerChan:
			s.processBufferedEvents(client, currentBlockNumber)
		}
	}
}

func (s *Server) fetchFinalizedEvents(bitcoinBridge *abi.BitcoinBridge, finalizedBlock *big.Int) {
	startBlock := new(big.Int).Sub(finalizedBlock, SearchedRange).Uint64()
	endBlock := finalizedBlock.Uint64()

	opts := &bind.FilterOpts{
		Start: startBlock,
		End:   &endBlock,
	}
	s.fetchEvents(bitcoinBridge, opts, true) // True for finalized events
}

func (s *Server) fetchNonFinalizedEvents(bitcoinBridge *abi.BitcoinBridge, finalizedBlock *big.Int, currentBlockNumber uint64) {
	opts := &bind.FilterOpts{
		Start: finalizedBlock.Uint64(),
		End:   &currentBlockNumber,
	}
	s.fetchEvents(bitcoinBridge, opts, false) // False for non-finalized events
}

func (s *Server) fetchEvents(bitcoinBridge *abi.BitcoinBridge, opts *bind.FilterOpts, finalized bool) {
	events, err := bitcoinBridge.FilterAssetsLocked(opts, nil, nil)
	if err != nil {
		s.logger.Error("Failed to filter AssetsLocked events: %v", err) // CHANGE: Use Errorf
		return
	}

	s.eventsMutex.Lock()
	defer s.eventsMutex.Unlock()

	for events.Next() {
		event := events.Event
		if finalized {
			s.events = append(s.events, *event)
			s.logger.Info(
				"Finalized AssetsLocked event",
				"sequence", event.SequenceNumber.String(),
				"recipient", event.Recipient.Hex(),
				"amount", event.TbtcAmount.String(),
			)
		} else {
			s.bufferEvents = append(s.bufferEvents, *event)
			s.logger.Info(
				"Non-finalized AssetsLocked event",
				"sequence", event.SequenceNumber.String(),
				"recipient", event.Recipient.Hex(),
				"amount", event.TbtcAmount.String(),
			)
		}
	}
}

// TODO: decide if we need to subscribe to the real block numbers. Historically,
// the block creation time was between 13-15 sec, since Post Merge (PoS) it
// takes 12 sec per block.
// Cons for subscribing to the block numbers:
// - subscription to the block numbers might create unnecessary cost. E.g. by RPC providers
// - things can go wrong due to e.g. connection issues
// Pros:
// - no need to track any upgrades to PoS around average time for blocks creation.
func (s *Server) startBlockTicker(currentBlockNumber uint64) chan uint64 {
	tickerChan := make(chan uint64)
	ticker := currentBlockNumber

	go func() {
		for {
			time.Sleep(BlockTime)
			ticker++
			tickerChan <- ticker
		}
	}()

	return tickerChan
}

func (s *Server) processBufferedEvents(client *ethclient.Client, currentBlockNumber uint64) {
	s.eventsMutex.Lock()
	defer s.eventsMutex.Unlock()

	var remainingEvents []abi.BitcoinBridgeAssetsLocked
	for _, event := range s.bufferEvents {
		if currentBlockNumber > event.Raw.BlockNumber+uint64(ConfirmationBlocks) {
			if isEventInBlock(client, int64(event.Raw.BlockNumber), event) {
				s.events = append(s.events, event)
			} else {
				remainingEvents = append(remainingEvents, event)
			}
		} else {
			remainingEvents = append(remainingEvents, event)
		}
	}
	s.bufferEvents = remainingEvents
	// TODO: remove logger
	fmt.Println("number of confirmed events", len(s.events))
	fmt.Println("number of unconfirmed events", len(s.bufferEvents))
	fmt.Println("------")

	if len(s.events) > CachedEvents {
		s.events = s.events[CachedEvents/10:]
	}
}

// Fetches events for a given block number from the blockchain.
func fetchEventsForBlock(client *ethclient.Client, blockNumber int64) ([]types.Log, error) {
	contractAddress, err := readBitcoinBridgeAddress()
	if err != nil {
		return nil, err
	}

	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(blockNumber),
		ToBlock:   big.NewInt(blockNumber),
		Addresses: []common.Address{contractAddress},
	}

	logs, err := client.FilterLogs(context.Background(), query)
	if err != nil {
		return nil, err
	}

	return logs, nil
}

func isEventInBlock(client *ethclient.Client, blockNumber int64, event abi.BitcoinBridgeAssetsLocked) bool {
	foundEvents, err := fetchEventsForBlock(client, blockNumber)
	if err != nil {
		return false
	}

	for _, e := range foundEvents {
		if e.TxHash == event.Raw.TxHash && e.Index == event.Raw.Index {
			// TODO: remove logger
			fmt.Println("Event is confirmed in block")
			return true
		}
	}
	// TODO: remove logger
	fmt.Println("Event was NOT found in block")
	return false
}

// Construct a new instance of the Ethereum Bitcoin Bridge contract.
func initializeBitcoinBridgeContract(
	baseChain *ethconnect.BaseChain,
	bitcoinBridgeAddress common.Address,
) (*abi.BitcoinBridge, error) {
	bitcoinBridge, err := abi.NewBitcoinBridge(bitcoinBridgeAddress, baseChain.Client)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to attach to Bitcoin Bridge contract: [%v]",
			err,
		)
	}

	return bitcoinBridge, nil
}

// startGRPCServer starts the gRPC server and registers the Ethereum sidecar
// service.
func (s *Server) startGRPCServer(
	ctx context.Context,
	address string,
) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		panic(fmt.Sprintf("failed to listen: [%v]", err))
	}

	pb.RegisterEthereumSidecarServer(s.grpcServer, s)

	s.logger.Info(
		"gRPC server started",
		"address", address,
	)

	go func() {
		if err := s.grpcServer.Serve(listener); err != nil {
			panic(fmt.Sprintf("gRPC server failure: [%v]", err))
		}
	}()

	<-ctx.Done()

	s.logger.Info("shutting down gRPC server...")
	s.grpcServer.GracefulStop()

	s.logger.Info("gRPC server stopped")
}

// AssetsLockedEvents returns a list of AssetsLocked events based on the
// passed request. It is executed by the gRPC server.
func (s *Server) AssetsLockedEvents(
	_ context.Context,
	req *pb.AssetsLockedEventsRequest,
) (
	*pb.AssetsLockedEventsResponse,
	error,
) {
	s.eventsMutex.RLock()
	defer s.eventsMutex.RUnlock()

	start, end := req.SequenceStart, req.SequenceEnd

	// The sequence start must be lower than the sequence end if both values are
	// non-nil.
	if !start.IsNil() && !end.IsNil() && start.GTE(end) {
		return nil, ErrSequenceStartNotLower
	}

	// Filter events that fit into the requested range.
	filteredEvents := []*bridgetypes.AssetsLockedEvent{}
	for _, event := range s.events {
		if (start.IsNil() || event.SequenceNumber.Cmp(start.BigInt()) >= 0) && (end.IsNil() || event.SequenceNumber.Cmp(end.BigInt()) < 0) {
			filteredEvents = append(filteredEvents, &bridgetypes.AssetsLockedEvent{
				Sequence:  sdkmath.NewIntFromBigInt(event.SequenceNumber),
				Recipient: event.Recipient.Hex(),
				Amount:    sdkmath.NewIntFromBigInt(event.TbtcAmount),
			})
		}
	}

	return &pb.AssetsLockedEventsResponse{
		Events: filteredEvents,
	}, nil
}
