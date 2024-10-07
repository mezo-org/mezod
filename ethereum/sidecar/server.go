package sidecar

import (
	"context"
	"fmt"
	"math/big"
	"net"
	"sync"

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
)

var (
	// ErrSequencePointerNil is the error returned when the start or end of the
	// sequence is nil in the received request.
	ErrSequencePointerNil = fmt.Errorf("sequence start or end is a nil pointer")

	// ErrSequenceStartNotLower is the error returned when the start of
	// the sequence is not lower than sequence end.
	ErrSequenceStartNotLower = fmt.Errorf(
		"sequence start is not lower than sequence end",
	)

	BitcoinBridgeName = "bitcoinbridge"

	// Number of blocks to look back for AssetsLocked events in the BitcoinBridge
	// contract. This is approximately 2 weeks window for 12s block time.
	EventsLookupInBlocks = new(big.Int).SetInt64(100000)

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
	events     []bridgetypes.AssetsLockedEvent
	grpcServer *grpc.Server
	logger     log.Logger
}

type Block struct {
	Number string `json:"number"`
}

// RunServer initializes the server, starts the event observing routine and
// starts the gRPC server.
func RunServer(
	ctx context.Context,
	grpcAddress string,
	providerUrl string,
	ethereumNetwork string,
	logger log.Logger,
) *Server {
	server := &Server{
		sequenceTip: sdkmath.ZeroInt(),
		events:      make([]bridgetypes.AssetsLockedEvent, 0),
		grpcServer:  grpc.NewServer(),
		logger:      logger,
	}

	bitcoinBridgeAddress, err := readBitcoinBridgeAddress()
	if err != nil {
		server.logger.Error("Failed to read the BitcoinBridge address: %v", err)
	}

	// Connect to the Ethereum network
	chain, err := ethconnect.Connect(ctx, ethconfig.Config{
		Network:           networkFromString(ethereumNetwork),
		URL:               providerUrl,
		ContractAddresses: map[string]string{BitcoinBridgeName: bitcoinBridgeAddress.String()},
	})
	if err != nil {
		server.logger.Error("Failed to connect to the Ethereum network: %v", err)
	}

	go server.observeEvents(ctx, chain, bitcoinBridgeAddress)
	go server.startGRPCServer(ctx, grpcAddress)

	return server
}

// observeEvents starts the AssetLocked observing routine.
func (s *Server) observeEvents(ctx context.Context, chain *ethconnect.BaseChain, bitcoinBridgeAddress common.Address) {
	// Get the BitcoinBridge contract instance
	bitcoinBridge, err := initializeBitcoinBridgeContract(chain, bitcoinBridgeAddress)
	if err != nil {
		s.logger.Error("Failed to initialize BitcoinBridge contract: %v", err)
	}

	var finalizedBlock Block
	err = chain.RPCClient.CallContext(ctx, &finalizedBlock, "eth_getBlockByNumber", "finalized", false)
	if err != nil {
		s.logger.Error("Failed to get the finalized block: %v", err)
	}

	finalizedBlockInt, ok := new(big.Int).SetString(finalizedBlock.Number[2:], 16)
	if !ok {
		s.logger.Error("Failed to convert hexadecimal string")
	}

	fmt.Printf("Hexadecimal: %s\nDecimal: %d\n", finalizedBlock.Number, finalizedBlockInt)

	// Latest finalized block - ~2 weeks of blocks
	fromBlock := new(big.Int).Sub(finalizedBlockInt, EventsLookupInBlocks).Uint64()
	endBlock := finalizedBlockInt.Uint64()

	if (s.events == nil) || (len(s.events) == 0) {

		opts := &bind.FilterOpts{
			Start: fromBlock,
			End:   &endBlock,
		}
		events, err := bitcoinBridge.FilterAssetsLocked(opts, nil, nil)
		if err != nil {
			s.logger.Error(
				"failed to filter to Assets Locked events: [%v]",
				err,
			)
		}

		for events.Next() {
			event := events.Event
			s.eventsMutex.Lock()
			s.events = append(s.events, bridgetypes.AssetsLockedEvent{
				Sequence:  sdkmath.NewIntFromBigInt(event.SequenceNumber),
				Recipient: event.Recipient.Hex(),
				Amount:    sdkmath.NewIntFromBigInt(event.TbtcAmount),
			})
			s.eventsMutex.Unlock()
			s.logger.Info(
				"new AssetsLocked event",
				"sequence", event.SequenceNumber.String(),
				"recipient", event.Recipient.Hex(),
				"amount", event.TbtcAmount.String(),
			)
		}
	}

	// TODO: Subscribe to the AssetsLocked events and add the new ones to cache.

	// Prune old events. Once the cache reaches `CachedEvents` elements limit,
	// free 10% of the oldest ones.
	if len(s.events) > CachedEvents {
		s.events = s.events[CachedEvents/10:]
	}
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
