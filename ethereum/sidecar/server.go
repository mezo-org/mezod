package sidecar

import (
	"context"
	"fmt"
	"math/big"
	"net"
	"sort"
	"sync"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	ethconfig "github.com/keep-network/keep-common/pkg/chain/ethereum"
	ethconnect "github.com/mezo-org/mezod/ethereum"
	"github.com/mezo-org/mezod/ethereum/bindings/portal/gen"
	"github.com/mezo-org/mezod/ethereum/bindings/portal/gen/abi"
	pb "github.com/mezo-org/mezod/ethereum/sidecar/types"
	bridgetypes "github.com/mezo-org/mezod/x/bridge/types"
	"google.golang.org/grpc"
)

var (
	// errSequenceStartNotLower is the error returned when the start of
	// the sequence is not lower than sequence end.
	errSequenceStartNotLower = fmt.Errorf(
		"sequence start is not lower than sequence end",
	)

	// bitcoinBridgeName is the name of the BitcoinBridge contract.
	bitcoinBridgeName = "BitcoinBridge"

	// searchedRange is a number of blocks to look back for AssetsLocked events in
	// the BitcoinBridge contract. This is approximately 2 weeks window for 12 sec.
	// block time.
	searchedRange = new(big.Int).SetInt64(100000)

	// cachedEventsLimit is a number of events to keep in the cache.
	// Size of Sequence: 32 bytes
	// Size of Recipient: 42 bytes
	// Size of Amount: 32 bytes
	// Struct overhead and padding: ~16bytes
	// Total size of one event: 122 bytes (0.12KB)
	// Assuming we want to allocate ~64MB for the cache, we can store ~546k events.
	// For simplicity, let's make 500k events our limit. Even if deposits hit 1000
	// daily mark, that would give 50 days of cached events which should be more than
	// enough.
	cachedEventsLimit = 500000
)

// Server observes events emitted by the Mezo `BitcoinBridge` contract on the
// Ethereum chain. It enables retrieval of information on the assets locked by
// the contract. It is intended to be run as a separate process.
type Server struct {
	logger log.Logger

	grpcServer *grpc.Server

	eventsMutex sync.RWMutex
	events      []bridgetypes.AssetsLockedEvent

	lastFinalizedBlockMutex sync.RWMutex
	lastFinalizedBlock      *big.Int

	bitcoinBridge *abi.BitcoinBridge

	chain *ethconnect.BaseChain
}

// RunServer initializes the server, starts the event observing routine and
// starts the gRPC server.
func RunServer(
	ctx context.Context,
	cancel context.CancelFunc,
	grpcAddress string,
	providerURL string,
	ethereumNetwork string,
	logger log.Logger,
) *Server {
	server := &Server{
		events:             make([]bridgetypes.AssetsLockedEvent, 0),
		grpcServer:         grpc.NewServer(),
		logger:             logger,
		lastFinalizedBlock: new(big.Int),
	}

	var err error

	if gen.BitcoinBridgeAddress == "" {
    panic("BitcoinBridgeAddress is empty")
	}

	// Connect to the Ethereum network
	server.chain, err = ethconnect.Connect(ctx, ethconfig.Config{
		Network:           ethconnect.NetworkFromString(ethereumNetwork),
		URL:               providerURL,
		ContractAddresses: map[string]string{bitcoinBridgeName: gen.BitcoinBridgeAddress},
	})
	if err != nil {
		panic(fmt.Sprintf("failed to connect to the Ethereum network: %v", err))
	}

	// Initialize the BitcoinBridge contract instance
	server.bitcoinBridge, err = server.initializeBitcoinBridgeContract(common.HexToAddress(gen.BitcoinBridgeAddress))
	if err != nil {
		panic(fmt.Sprintf("failed to initialize BitcoinBridge contract: %v", err))
	}

	errChan := make(chan error, 2)

	go func() {
		if err := server.observeEvents(ctx); err != nil {
			errChan <- err
		}
	}()
	go func() {
		if err := server.startGRPCServer(ctx, grpcAddress); err != nil {
			errChan <- err
		}
	}()

	// Wait for an error from either goroutine
	err = <-errChan
	server.logger.Error("Error encountered: %v. Shutting down application.", err)

	// Signal goroutines to stop
	cancel()

	// Exit the application
	server.logger.Error("Sidecar stopped.")

	return server
}

// observeEvents monitors and processes events from the BitcoinBridge smart contract.
//
//   - Initializes a BitcoinBridge contract instance using the provided blockchain connection and contract address.
//   - Retrieves the most recent finalized block number from the blockchain.
//   - Calculates a start block that is two weeks prior to the current finalized block. It then fetches
//     `AssetsLocked` events for this range.
//   - Sets up a ticker channel to continuously monitor new finalized blocks.
//   - On each new block notification from the ticker channel, calls `processEvents` to handle new events
//     since the last finalized block.
func (s *Server) observeEvents(ctx context.Context) error {
	finalizedBlock, err := s.chain.FinalizedBlock(ctx)
	if err != nil {
		s.logger.Error("failed to get the finalized block")
		return err
	}

	// Fetch AssetsLocked events two weeks back from the finalized block
	startBlock := new(big.Int).Sub(finalizedBlock, searchedRange).Uint64()
	err = s.fetchFinalizedEvents(startBlock, finalizedBlock.Uint64())
	if err != nil {
		s.logger.Error("failed to fetch historical events")
		return err
	}

	s.lastFinalizedBlockMutex.Lock()
	s.lastFinalizedBlock = finalizedBlock
	s.lastFinalizedBlockMutex.Unlock()

	// Start a ticker to periodically check the current block number
	tickerChan := s.chain.BlockCounter().WatchBlocks(ctx)

	for {
		select {
		case <-ctx.Done(): // Handle context cancellation
			s.logger.Info("stopping event observation due to context cancellation")
			return nil
		case <-tickerChan:
			// On each tick check if the current finalized block is greater than the last
			// finalized block.
			// TODO: add a simple counter to retry in case a connection issue occurs
			err := s.processEvents(ctx)
			if err != nil {
				s.logger.Error("failed to monitor newly emitted events: %v", err)
			}
			// return from here in case of a an issue
		}
	}
}

// processEvents processes events from Ethereum by fetching the AssetsLocked
// finalized events within a specified block range and managing memory usage
// for cached events.
func (s *Server) processEvents(ctx context.Context) error {
	currentFinalizedBlock, err := s.chain.FinalizedBlock(ctx)
	if err != nil {
		return err
	}

	s.lastFinalizedBlockMutex.RLock()
	shouldFetchEvents := currentFinalizedBlock.Cmp(s.lastFinalizedBlock)
	s.lastFinalizedBlockMutex.RUnlock()

	if shouldFetchEvents > 0 {
		// Specified range in FilterOps is inclusive.
		// 1 is added to the lastFinalizedBlock to make the range exclusive at
		// the beginning of the range.
		// E.g.
		// lastFinalizedBlock = 100
		// currentFinalizedBlock = 132
		// fetching events in the following range [101, 132]
		exclusiveLastFinalizedBlock := s.lastFinalizedBlock.Uint64() + 1
		err := s.fetchFinalizedEvents(exclusiveLastFinalizedBlock, currentFinalizedBlock.Uint64())
		if err != nil {
			return err
		}
		s.lastFinalizedBlockMutex.Lock()
		s.lastFinalizedBlock = currentFinalizedBlock
		s.lastFinalizedBlockMutex.Unlock()
	}

	// Free up memory up to the length that exceeds the cache size.
	if len(s.events) > cachedEventsLimit {
		s.eventsMutex.Lock()
		trim := len(s.events) - cachedEventsLimit
		s.events = s.events[trim:]
		s.eventsMutex.Unlock()
	}

	return nil
}

// fetchFinalizedEvents retrieves and processes finalized `AssetsLocked` events
// from the Ethereum network, within a specified block range. It uses the
// provided BitcoinBridge contract to filter these events. Each event is
// transformed into an `AssetsLockedEvent` type compatible with the bridgetypes
// package and added to the server's event list with mutex protection.
func (s *Server) fetchFinalizedEvents(startBlock uint64, endBlock uint64) error {
	opts := &bind.FilterOpts{
		Start: startBlock,
		End:   &endBlock,
	}

	events, err := s.bitcoinBridge.FilterAssetsLocked(opts, nil, nil)
	if err != nil {
		s.logger.Error("failed to filter AssetsLocked events")
		return err
	}

	var bufferedEvents []bridgetypes.AssetsLockedEvent

	for events.Next() {
		event := bridgetypes.AssetsLockedEvent{
			Sequence:  sdkmath.NewIntFromBigInt(events.Event.SequenceNumber),
			Recipient: sdk.AccAddress(events.Event.Recipient.Bytes()).String(),
			Amount:    sdkmath.NewIntFromBigInt(events.Event.TbtcAmount),
		}
		bufferedEvents = append(bufferedEvents, event)
		s.logger.Info(
			"finalized AssetsLocked event",
			"sequence", event.Sequence.String(),
			"recipient", events.Event.Recipient,
			"amount", event.Amount.String(),
		)
	}

	// Make sure the events are sorted in ascending order by the sequence number
	sort.Slice(bufferedEvents, func(i, j int) bool {
		return bufferedEvents[i].Sequence.LT(bufferedEvents[j].Sequence)
	})

	s.eventsMutex.RLock()
	isEventsListEmpty := len(s.events) == 0
	s.eventsMutex.RUnlock()
	// Make sure there are no gaps between events and bufferedEvents lists
	if !isEventsListEmpty && len(bufferedEvents) > 0 {
		lastEvent := s.events[len(s.events)-1]
		firstEvent := bufferedEvents[0]
		if !lastEvent.Sequence.Add(sdkmath.NewInt(1)).Equal(firstEvent.Sequence) {
			s.logger.Error("sequence gap between events")
			return err
		}
	}

	if !bridgetypes.AssetsLockedEvents(bufferedEvents).IsValid() {
		s.logger.Error("invalid AssetsLocked events")
		return err
	}

	s.eventsMutex.Lock()
	s.events = append(s.events, bufferedEvents...)
	s.eventsMutex.Unlock()

	return nil
}

// startGRPCServer starts the gRPC server and registers the Ethereum sidecar
// service.
func (s *Server) startGRPCServer(
	ctx context.Context,
	address string,
) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		s.logger.Error("failed to listen")
		return err
	}

	pb.RegisterEthereumSidecarServer(s.grpcServer, s)

	s.logger.Info(
		"gRPC server started",
		"address", address,
	)

	if err := s.grpcServer.Serve(listener); err != nil {
		s.logger.Error("gRPC server failure")
		return err
	}

	<-ctx.Done()

	s.logger.Info("shutting down gRPC server...")
	s.grpcServer.GracefulStop()

	s.logger.Info("gRPC server stopped")
	return nil
}

// Construct a new instance of the Ethereum Bitcoin Bridge contract.
func (s *Server) initializeBitcoinBridgeContract(
	bitcoinBridgeAddress common.Address,
) (*abi.BitcoinBridge, error) {
	bitcoinBridge, err := abi.NewBitcoinBridge(bitcoinBridgeAddress, s.chain.Client())
	if err != nil {
		s.logger.Error("failed to attach to Bitcoin Bridge contract")
		return nil, err
	}

	return bitcoinBridge, nil
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
		return nil, errSequenceStartNotLower
	}

	// Filter events that fit into the requested range.
	filteredEvents := []*bridgetypes.AssetsLockedEvent{}
	for _, event := range s.events {
		if (start.IsNil() || event.Sequence.GTE(start)) && (end.IsNil() || event.Sequence.LT(end)) {
			filteredEvents = append(filteredEvents, &bridgetypes.AssetsLockedEvent{
				Sequence:  event.Sequence,
				Recipient: event.Recipient,
				Amount:    event.Amount,
			})
		}
	}

	return &pb.AssetsLockedEventsResponse{
		Events: filteredEvents,
	}, nil
}
