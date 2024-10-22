package sidecar

import (
	"context"
	"fmt"
	"math/big"
	"net"
	"sync"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	ethconfig "github.com/keep-network/keep-common/pkg/chain/ethereum"
	ethconnect "github.com/mezo-org/mezod/ethereum"
	"github.com/mezo-org/mezod/ethereum/bindings/portal/gen/abi"
	pb "github.com/mezo-org/mezod/ethereum/sidecar/types"
	bridgetypes "github.com/mezo-org/mezod/x/bridge/types"
	"google.golang.org/grpc"
	"github.com/mezo-org/mezod/ethereum/bindings/portal/gen"
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

	// cachedEvents is a number of events to keep in the cache.
	// Size of Sequence: 32 bytes
	// Size of Recipient: 42 bytes
	// Size of Amount: 32 bytes
	// Struct overhead and padding: ~16bytes
	// Total size of one event: 122 bytes (0.12KB)
	// Assuming we want to allocate ~64MB for the cache, we can store ~546k events.
	// For simplicity, let's make 500k events our limit. Even if deposits hit 1000
	// daily mark, that would give 50 days of cached events which should be more than
	// enough.
	cachedEvents = 500000
)

// Server observes events emitted by the Mezo `BitcoinBridge` contract on the
// Ethereum chain. It enables retrieval of information on the assets locked by
// the contract. It is intended to be run as a separate process.
type Server struct {
	eventsMutex             sync.RWMutex
	events                  []bridgetypes.AssetsLockedEvent
	grpcServer              *grpc.Server
	logger                  log.Logger
	lastFinalizedBlock      *big.Int
	lastFinalizedBlockMutex sync.RWMutex
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
		events:             make([]bridgetypes.AssetsLockedEvent, 0),
		grpcServer:         grpc.NewServer(),
		logger:             logger,
		lastFinalizedBlock: new(big.Int),
	}

	// Connect to the Ethereum network
	chain, err := ethconnect.Connect(ctx, ethconfig.Config{
		Network:           networkFromString(ethereumNetwork),
		URL:               providerURL,
		ContractAddresses: map[string]string{bitcoinBridgeName: gen.BitcoinBridgeAddress},
	})
	if err != nil {
		server.logger.Error("failed to connect to the Ethereum network: %v", err)
		return nil
	}

	go server.observeEvents(ctx, chain, common.HexToAddress(gen.BitcoinBridgeAddress))
	go server.startGRPCServer(ctx, grpcAddress)

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
func (s *Server) observeEvents(ctx context.Context, chain *ethconnect.BaseChain, bitcoinBridgeAddress common.Address) {
	// Initialize the BitcoinBridge contract instance
	bitcoinBridge, err := initializeBitcoinBridgeContract(chain, bitcoinBridgeAddress)
	if err != nil {
		s.logger.Error("failed to initialize BitcoinBridge contract: %v", err)
		return
	}

	finalizedBlock, err := chain.FinalizedBlock(ctx)
	if err != nil {
		s.logger.Error("failed to get the finalized block: %v", err)
		return
	}

	// Fetch AssetsLocked events two weeks back from the finalized block
	startBlock := new(big.Int).Sub(finalizedBlock, searchedRange).Uint64()
	s.fetchFinalizedEvents(bitcoinBridge, startBlock, finalizedBlock.Uint64())

	s.lastFinalizedBlockMutex.Lock()
	s.lastFinalizedBlock = finalizedBlock
	s.lastFinalizedBlockMutex.Unlock()

	// Start a ticker to periodically check the current block number
	tickerChan := chain.BlockCounter().WatchBlocks(ctx)

	for {
		select {
		case <-ctx.Done(): // Handle context cancellation
			s.logger.Info("stopping event observation due to context cancellation")
			return
		case <-tickerChan:
			// On each tick check if the current finalized block is greater than the last
			// finalized block.
			err := s.processEvents(ctx, chain, bitcoinBridge)
			if err != nil {
				s.logger.Error("failed to process events: %v", err)
				return
			}
		}
	}
}

// processEvents processes events from Ethereum by fetching the AssetsLocked
// finalized events within a specified block range and managing memory usage
// for cached events.
func (s *Server) processEvents(ctx context.Context, chain *ethconnect.BaseChain, bitcoinBridge *abi.BitcoinBridge) error {
	currentFinalizedBlock, err := chain.FinalizedBlock(ctx)
	if err != nil {
		return err
	}

	if currentFinalizedBlock.Cmp(s.lastFinalizedBlock) > 0 {
		// Specified range in FilterOps is inclusive.
		// 1 is added to the lastFinalizedBlock to make the range exclusive at
		// the beginning of the range.
		// E.g.
		// lastFinalizedBlock = 100
		// currentFinalizedBlock = 132
		// fetching events in the following range [101, 132]
		exclusiveLastFinalizedBlock := s.lastFinalizedBlock.Uint64() + 1
		s.fetchFinalizedEvents(bitcoinBridge, exclusiveLastFinalizedBlock, currentFinalizedBlock.Uint64())
		s.lastFinalizedBlockMutex.Lock()
		s.lastFinalizedBlock = currentFinalizedBlock
		s.lastFinalizedBlockMutex.Unlock()
	}

	// Free up memory up to the length that exceeds the cache size.
	if len(s.events) > cachedEvents {
		s.eventsMutex.Lock()
		trim := len(s.events) - cachedEvents
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
func (s *Server) fetchFinalizedEvents(bitcoinBridge *abi.BitcoinBridge, startBlock uint64, endBlock uint64) {
	opts := &bind.FilterOpts{
		Start: startBlock,
		End:   &endBlock,
	}

	events, err := bitcoinBridge.FilterAssetsLocked(opts, nil, nil)
	if err != nil {
		s.logger.Error("failed to filter AssetsLocked events: %v", err)
		return
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

	if !bridgetypes.AssetsLockedEvents(bufferedEvents).IsValid() {
		s.logger.Error("invalid AssetsLocked events")
		return
	}

	s.eventsMutex.Lock()
	s.events = append(s.events, bufferedEvents...)
	s.eventsMutex.Unlock()
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
		return nil, errSequenceStartNotLower
	}

	// Filter events that fit into the requested range.
	filteredEvents := []*bridgetypes.AssetsLockedEvent{}
	for _, event := range s.events {
		if (start.IsNil() || event.Sequence.BigInt().Cmp(start.BigInt()) >= 0) && (end.IsNil() || event.Sequence.BigInt().Cmp(end.BigInt()) < 0) {
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

// Construct a new instance of the Ethereum Bitcoin Bridge contract.
func initializeBitcoinBridgeContract(
	baseChain *ethconnect.BaseChain,
	bitcoinBridgeAddress common.Address,
) (*abi.BitcoinBridge, error) {
	bitcoinBridge, err := abi.NewBitcoinBridge(bitcoinBridgeAddress, baseChain.Client())
	if err != nil {
		return nil, fmt.Errorf(
			"failed to attach to Bitcoin Bridge contract: [%v]",
			err,
		)
	}

	return bitcoinBridge, nil
}
