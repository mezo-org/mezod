package sidecar

import (
	"context"
	"fmt"
	"math/big"
	"math/rand"
	"net"
	"sync"
	"time"

	"github.com/mezo-org/mezod/crypto/ethsecp256k1"
	"google.golang.org/grpc"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	pb "github.com/mezo-org/mezod/ethereum/sidecar/types"
	bridgetypes "github.com/mezo-org/mezod/x/bridge/types"
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
)

// Server observes events emitted by the Mezo `BitcoinBridge` contract on the
// Ethereum chain. It enables retrieval of information on the assets locked by
// the contract. It is intended to be run as a separate process.
type Server struct {
	sequenceTip sdkmath.Int
	eventsMutex sync.RWMutex
	events      []bridgetypes.AssetsLockedEvent
	grpcServer  *grpc.Server

	logger log.Logger
}

// RunServer initializes the server, starts the event observing routine and
// starts the gRPC server.
func RunServer(
	ctx context.Context,
	grpcAddress string,
	_ string,
	logger log.Logger,
) *Server {
	server := &Server{
		sequenceTip: sdkmath.ZeroInt(),
		events:      make([]bridgetypes.AssetsLockedEvent, 0),
		grpcServer:  grpc.NewServer(),
		logger:      logger,
	}

	// Start observing AssetsLocked events.
	go server.observeEvents(ctx)

	// Start the gRPC server.
	go server.startGRPCServer(
		ctx,
		grpcAddress,
	)

	return server
}

// observeEvents starts the AssetLocked observing routine. For now it is
// generating dummy events.
func (s *Server) observeEvents(ctx context.Context) {
	// TODO: Replace with the code that actually observes the Ethereum chain.
	//       Temporarily generate some dummy events.
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			//nolint:gosec
			eventsCount := rand.Intn(11) // [0, 10] events

			for i := 0; i < eventsCount; i++ {
				s.sequenceTip = s.sequenceTip.Add(sdkmath.OneInt())

				amount := new(big.Int).Mul(
					//nolint:gosec
					big.NewInt(rand.Int63n(10)+1),
					precision,
				)

				key, err := ethsecp256k1.GenerateKey()
				if err != nil {
					panic(err)
				}

				recipient := sdk.AccAddress(key.PubKey().Address().Bytes())

				event := bridgetypes.AssetsLockedEvent{
					Sequence:  s.sequenceTip,
					Recipient: recipient.String(),
					Amount:    sdkmath.NewIntFromBigInt(amount),
				}

				s.eventsMutex.Lock()
				s.events = append(s.events, event)

				s.logger.Info(
					"new AssetsLocked event",
					"sequence", event.Sequence.String(),
					"recipient", event.Recipient,
					"amount", event.Amount.String(),
				)

				// Prune old events. Once the cache reaches 10000 elements,
				// remove the oldest 100.
				if len(s.events) > 10000 {
					s.events = s.events[100:]
				}

				s.eventsMutex.Unlock()
			}
		case <-ctx.Done():
			return
		}
	}
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

	// The sequence start and end must be non-nil pointers in the request.
	// Notice that the sequence start and end may store nil values (which can be
	// tested using the `isNil` function) but the pointers themselves must be
	// non-nil.
	if req.SequenceStart == nil || req.SequenceEnd == nil {
		return nil, ErrSequencePointerNil
	}

	start, end := *req.SequenceStart, *req.SequenceEnd

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
