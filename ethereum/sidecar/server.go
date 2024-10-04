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

	go server.observeEvents(chain, bitcoinBridgeAddress)
	go server.startGRPCServer(ctx, grpcAddress)

	return server
}

// observeEvents starts the AssetLocked observing routine.
func (s *Server) observeEvents(chain *ethconnect.BaseChain, bitcoinBridgeAddress common.Address) {
	// Get the BitcoinBridge contract instance
	bitcoinBridge, err := initializeBitcoinBridgeContract(chain, bitcoinBridgeAddress)
	if err != nil {
		s.logger.Error("Failed to initialize BitcoinBridge contract: %v", err)
	}

	// TODO: On Sidecar start, fetch events from 2 weeks old finilized blocks and
	//       store them in cache.
	fromBlock := big.NewInt(6762000)
	opts := &bind.FilterOpts{
		Start: fromBlock.Uint64(),
		End:   nil, // Set to nil for now to get events from the latest block
	}

	// TODO: Subscribe to the AssetsLocked event instead of simple polling.
	events, err := bitcoinBridge.FilterAssetsLocked(opts, nil, nil)
	if err != nil {
		s.logger.Error(
			"failed to filter to Assets Locked events: [%v]",
			err,
		)
	}

	// TODO: Use only those events which blocks were finalized on Ethereum.
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

	// Prune old events. Once the cache reaches 10000 elements,
	// remove the oldest 100.
	// TODO: Decide on the actual cleaning based on memory allocation for each event.
	if len(s.events) > 10000 {
		s.events = s.events[100:]
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
