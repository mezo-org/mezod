package sidecar

import (
	"context"
	"fmt"
	"math/big"
	"net"
	"sort"
	"sync"
	"time"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	ethconfig "github.com/keep-network/keep-common/pkg/chain/ethereum"
	"github.com/keep-network/keep-common/pkg/chain/ethereum/ethutil"
	ethconnect "github.com/mezo-org/mezod/ethereum"
	"github.com/mezo-org/mezod/ethereum/bindings/portal"
	pb "github.com/mezo-org/mezod/ethereum/sidecar/types"
	"github.com/mezo-org/mezod/version"
	bridgetypes "github.com/mezo-org/mezod/x/bridge/types"
	"google.golang.org/grpc"
)

var (
	// mezoBridgeName is the name of the MezoBridge contract.
	mezoBridgeName = "MezoBridge"

	// searchedRange is a number of blocks to look back for AssetsLocked events in
	// the MezoBridge contract. This is approximately 30 days window for 12 sec.
	// block time.
	searchedRange = new(big.Int).SetInt64(220000)

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

	// errSequenceGap is the error reported when there is a gap between
	// sequences of events.
	errSequenceGap = fmt.Errorf("sequence gap between events")

	// errInvalidEvents is the error reported when events are invalid.
	errInvalidEvents = fmt.Errorf("invalid AssetsLocked events")
)

// Server observes events emitted by the Mezo `MezoBridge` contract on the
// Ethereum chain. It enables retrieval of information on the assets locked by
// the contract. It is intended to be run as a separate process.
type Server struct {
	logger log.Logger

	// bridging-in
	grpcServer *grpc.Server

	eventsMutex sync.RWMutex
	events      []bridgetypes.AssetsLockedEvent

	lastFinalizedBlockMutex sync.RWMutex
	lastFinalizedBlock      *big.Int

	bridgeContract ethconnect.BridgeContract

	chain *ethconnect.BaseChain

	batchSize         uint64
	requestsPerMinute uint64

	// bridging-out
	bridgeOutClient *BridgeOutClient

	attestationMutex sync.RWMutex
	attestationQueue []bridgetypes.AssetsUnlockedEvent
}

// RunServer initializes the server, starts the event observing routine and
// starts the gRPC server.
func RunServer(
	logger log.Logger,
	grpcAddress string,
	providerURL string,
	ethereumNetwork string,
	batchSize uint64,
	requestsPerMinute uint64,
	bridgeOutServerAddress string,
	bridgeOutRequestTimeout time.Duration,
	registry codectypes.InterfaceRegistry,
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
		"Sidecar server resolved MezoBridge contract and Ethereum network",
		"mezo_bridge_address", mezoBridgeAddress,
		"ethereum_network", network,
	)

	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	var err error
	// Connect to the Ethereum network
	chain, err := ethconnect.Connect(ctx, ethconfig.Config{
		Network:           network,
		URL:               providerURL,
		ContractAddresses: map[string]string{mezoBridgeName: mezoBridgeAddress},
	})
	if err != nil {
		panic(fmt.Sprintf("failed to connect to the Ethereum network: %v", err))
	}

	// Initialize the MezoBridge contract instance.
	bridgeContract, err := initializeBridgeContract(common.HexToAddress(mezoBridgeAddress), chain.Client())
	if err != nil {
		panic(fmt.Sprintf("failed to initialize MezoBridge contract: %v", err))
	}

	bridgeOutClient, err := NewBridgeOutClient(
		logger,
		bridgeOutServerAddress,
		bridgeOutRequestTimeout,
		registry,
	)
	if err != nil {
		panic(fmt.Sprintf("failed to create bridge-out client: %v", err))
	}

	server := &Server{
		logger:             logger,
		grpcServer:         grpc.NewServer(),
		events:             make([]bridgetypes.AssetsLockedEvent, 0),
		lastFinalizedBlock: new(big.Int),
		bridgeContract:     NewBridgeContract(bridgeContract),
		chain:              chain,
		batchSize:          batchSize,
		requestsPerMinute:  requestsPerMinute,
		bridgeOutClient:    bridgeOutClient,
		attestationQueue:   make([]bridgetypes.AssetsUnlockedEvent, 0),
	}

	go func() {
		defer cancelCtx()
		err := server.observeAssetsLockedEvents(ctx)
		if err != nil {
			server.logger.Error(
				"AssetsLocked events observation routine failed",
				"err", err,
			)
		}

		server.logger.Info("AssetsLocked events observation routine stopped")
	}()

	go func() {
		defer cancelCtx()
		err := server.startGRPCServer(ctx, grpcAddress)
		if err != nil {
			server.logger.Error("gRPC server routine failed", "err", err)
		}

		server.logger.Info("gRPC server routine stopped")
	}()

	go func() {
		defer cancelCtx()
		err := server.observeAssetsUnlockedEvents(ctx)
		if err != nil {
			server.logger.Error(
				"AssetsUnlocked events observation routine failed",
				"err", err,
			)
		}

		server.logger.Info("AssetsUnlocked events observation routine stopped")
	}()

	go func() {
		defer cancelCtx()
		err := server.attestAssetsUnlockedEvents(ctx)
		if err != nil {
			server.logger.Error(
				"AssetsUnlocked events attestation routine failed",
				"err", err,
			)
		}

		server.logger.Info("AssetsUnlocked events attestation routine stopped")
	}()

	<-ctx.Done()

	server.logger.Error("sidecar stopped")
}

// observeAssetsLockedEvents monitors and processes AssetsLocked events from the MezoBridge smart contract.
//
//   - Initializes a MezoBridge contract instance using the provided blockchain connection and contract address.
//   - Retrieves the most recent finalized block number from the blockchain.
//   - Calculates a start block that is two weeks prior to the current finalized block. It then fetches
//     `AssetsLocked` events for this range.
//   - Sets up a ticker channel to continuously monitor new finalized blocks.
//   - On each new block notification from the ticker channel, calls `processEvents` to handle new events
//     since the last finalized block.
func (s *Server) observeAssetsLockedEvents(ctx context.Context) error {
	finalizedBlock, err := s.chain.FinalizedBlock(ctx)
	if err != nil {
		return fmt.Errorf("failed to get the finalized block: [%w]", err)
	}

	var startBlock uint64
	if searchedRange.Cmp(finalizedBlock) <= 0 {
		// Fetch AssetsLocked events two weeks back from the finalized block
		startBlock = new(big.Int).Sub(finalizedBlock, searchedRange).Uint64()
	} else {
		startBlock = 0
	}
	err = s.fetchFinalizedEvents(startBlock, finalizedBlock.Uint64())
	if err != nil {
		return fmt.Errorf("failed to fetch historical events: [%w]", err)
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
			// TODO: Add a basic counter to manage validation issues that may occur
			//			 when processing events. This counter should allow a few retry
			//			 attempts to handle temporary connection issues, but should halt
			//			 the sidecar if an error arises specifically with event processing.”
			err := s.processEvents(ctx)
			if err != nil {
				s.logger.Error("failed to monitor newly emitted events", "err", err)
			}
		}
	}
}

// processEvents processes events from Ethereum by fetching the AssetsLocked
// finalized events within a specified block range and managing memory usage
// for cached events.
func (s *Server) processEvents(ctx context.Context) error {
	currentFinalizedBlock, err := s.chain.FinalizedBlock(ctx)
	if err != nil {
		return fmt.Errorf("cannot get finalized block: [%w]", err)
	}

	s.lastFinalizedBlockMutex.RLock()
	shouldFetchEvents := currentFinalizedBlock.Cmp(s.lastFinalizedBlock) > 0
	s.lastFinalizedBlockMutex.RUnlock()

	if shouldFetchEvents {
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
			return fmt.Errorf("cannot fetch finalized events: [%w]", err)
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
// provided MezoBridge contract to filter these events. Each event is
// transformed into an `AssetsLockedEvent` type compatible with the bridgetypes
// package and added to the server's event list with mutex protection.
func (s *Server) fetchFinalizedEvents(startBlock uint64, endBlock uint64) error {
	abiEvents, err := s.fetchABIEvents(startBlock, endBlock)
	if err != nil {
		return fmt.Errorf("failed to fetch ABI events: [%w]", err)
	}

	//nolint:all
	var bufferedEvents []bridgetypes.AssetsLockedEvent
	for _, abiEvent := range abiEvents {
		event := bridgetypes.AssetsLockedEvent{
			Sequence:  sdkmath.NewIntFromBigInt(abiEvent.SequenceNumber),
			Recipient: sdk.AccAddress(abiEvent.Recipient.Bytes()).String(),
			Token:     abiEvent.Token.Hex(),
			Amount:    sdkmath.NewIntFromBigInt(abiEvent.Amount),
		}
		bufferedEvents = append(bufferedEvents, event)
		s.logger.Info(
			"finalized AssetsLocked event",
			"sequence", event.Sequence.String(),
			"recipient", abiEvent.Recipient,
			"token", abiEvent.Token,
			"amount", event.Amount.String(),
		)
	}

	if len(bufferedEvents) == 0 {
		s.logger.Info("no new events to process")
		return nil
	}

	// Make sure the events are sorted in ascending order by the sequence number
	sort.Slice(bufferedEvents, func(i, j int) bool {
		return bufferedEvents[i].Sequence.LT(bufferedEvents[j].Sequence)
	})

	s.eventsMutex.Lock()
	defer s.eventsMutex.Unlock()

	// Make sure there are no gaps between events and bufferedEvents lists
	if len(s.events) > 0 && len(bufferedEvents) > 0 {
		lastEvent := s.events[len(s.events)-1]
		firstEvent := bufferedEvents[0]
		if !lastEvent.Sequence.Add(sdkmath.NewInt(1)).Equal(firstEvent.Sequence) {
			return errSequenceGap
		}
	}

	if !bridgetypes.AssetsLockedEvents(bufferedEvents).IsValid() {
		return errInvalidEvents
	}

	s.events = append(s.events, bufferedEvents...)

	return nil
}

// fetchABIEvents retrieves raw `AssetsLocked` ABI events from the MezoBridge
// contract within a specified block range. The function fetches events in batches if
// the entire range is too large to fetch at once.
func (s *Server) fetchABIEvents(
	startBlock uint64,
	endBlock uint64,
) ([]*portal.MezoBridgeAssetsLocked, error) {
	s.logger.Info(
		"fetching AssetsLocked events from range",
		"startBlock", startBlock,
		"endBlock", endBlock,
	)

	abiEvents := make([]*portal.MezoBridgeAssetsLocked, 0)

	ticker := time.NewTicker(time.Minute / time.Duration(s.requestsPerMinute)) //nolint:gosec
	defer ticker.Stop()

	iterator, err := s.bridgeContract.FilterAssetsLocked(
		&bind.FilterOpts{
			Start: startBlock,
			End:   &endBlock,
		}, nil, nil, nil,
	)
	if err != nil {
		s.logger.Warn(
			"failed to fetch AssetsLocked events from the entire range; "+
				"falling back to batched events fetch",
			"startBlock", startBlock,
			"endBlock", endBlock,
			"err", err,
		)

		batchStartBlock := startBlock

		for batchStartBlock <= endBlock {
			batchEndBlock := batchStartBlock + s.batchSize
			if batchEndBlock > endBlock {
				batchEndBlock = endBlock
			}

			s.logger.Info(
				"fetching a batch of AssetsLocked events from range",
				"batchStartBlock", batchStartBlock,
				"batchEndBlock", batchEndBlock,
			)

			<-ticker.C

			batchIterator, batchErr := s.bridgeContract.FilterAssetsLocked(
				&bind.FilterOpts{
					Start: batchStartBlock,
					End:   &batchEndBlock,
				}, nil, nil, nil,
			)
			if batchErr != nil {
				return nil, fmt.Errorf(
					"batched AssetsLocked fetch failed: [%w]; giving up",
					batchErr,
				)
			}

			for batchIterator.Next() {
				abiEvents = append(abiEvents, batchIterator.Event())
			}

			batchStartBlock = batchEndBlock + 1
		}
	} else {
		for iterator.Next() {
			abiEvents = append(abiEvents, iterator.Event())
		}
	}

	return abiEvents, nil
}

// startGRPCServer starts the gRPC server and registers the Ethereum sidecar
// service.
func (s *Server) startGRPCServer(
	ctx context.Context,
	address string,
) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to listen: [%w]", err)
	}

	pb.RegisterEthereumSidecarServer(s.grpcServer, s)

	s.logger.Info(
		"gRPC server started",
		"address", address,
	)

	defer s.grpcServer.GracefulStop()

	errChan := make(chan error)

	go func() {
		err := s.grpcServer.Serve(listener)
		if err != nil {
			errChan <- err
		}
	}()

	select {
	case err := <-errChan:
		return fmt.Errorf("serve failed: [%w]", err)
	case <-ctx.Done():
		return nil
	}
}

func (s *Server) observeAssetsUnlockedEvents(ctx context.Context) error {
	// TODO: Implement
	<-ctx.Done()
	return nil
}

func (s *Server) attestAssetsUnlockedEvents(ctx context.Context) error {
	// TODO: Implement
	<-ctx.Done()
	return nil
}

// Version return the current version of the ethereum sidecar.
func (s *Server) Version(
	_ context.Context,
	_ *pb.VersionRequest,
) (*pb.VersionResponse, error) {
	return &pb.VersionResponse{
		Version: version.AppVersion,
	}, nil
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
		return nil, fmt.Errorf("sequence start is not lower than sequence end")
	}

	// The sequence start and end must be positive.
	if (!start.IsNil() && !start.IsPositive()) || (!end.IsNil() && !end.IsPositive()) {
		return nil, fmt.Errorf("invalid non positive sequence range")
	}

	// Filter events that fit into the requested range.
	filteredEvents := []*bridgetypes.AssetsLockedEvent{}
	for _, event := range s.events {
		if (start.IsNil() || event.Sequence.GTE(start)) && (end.IsNil() || event.Sequence.LT(end)) {
			filteredEvents = append(filteredEvents, &bridgetypes.AssetsLockedEvent{
				Sequence:  event.Sequence,
				Recipient: event.Recipient,
				Token:     event.Token,
				Amount:    event.Amount,
			})
		}
	}

	return &pb.AssetsLockedEventsResponse{
		Events: filteredEvents,
	}, nil
}

// Construct a new instance of the Ethereum MezoBridge contract.
func initializeBridgeContract(
	address common.Address,
	client ethutil.EthereumClient,
) (*portal.MezoBridge, error) {
	bridgeContract, err := portal.NewMezoBridge(address, client)
	if err != nil {
		return nil, fmt.Errorf("failed to attach to MezoBridge contract. %v", err)
	}

	return bridgeContract, nil
}
