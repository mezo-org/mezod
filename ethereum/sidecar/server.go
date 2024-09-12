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

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bridgetypes "github.com/mezo-org/mezod/x/bridge/types"
	pb "github.com/mezo-org/mezod/x/ethereum_sidecar/types"
)

// Server observes events emitted by the Mezo `BitcoinBridge` contract on the
// Ethereum chain. It enables retrieval of information on the assets locked by
// the contract. It is intended to be run as a separate process.
type Server struct {
	sequenceTip sdkmath.Int
	eventsMutex sync.Mutex
	events      []bridgetypes.AssetsLockedEvent
	grpcServer  *grpc.Server
}

// RunServer initializes the server, starts the event observing routine and
// starts the gRPC server.
func RunServer(
	ctx context.Context,
	grpcAddress string,
	_ string,
) *Server {
	server := &Server{
		sequenceTip: sdkmath.ZeroInt(),
		events:      make([]bridgetypes.AssetsLockedEvent, 0),
		grpcServer:  grpc.NewServer(),
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

				fmt.Printf(
					"New AssetsLocked event (sequence: %s, recipient: %s, amount: %s)\n",
					event.Sequence.String(),
					event.Recipient,
					event.Amount.String(),
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
	fmt.Printf("gRPC server started on [%s]\n", address)

	go func() {
		if err := s.grpcServer.Serve(listener); err != nil {
			panic(fmt.Sprintf("gRPC server failure: [%v]", err))
		}
	}()

	<-ctx.Done()

	fmt.Println("Shutting down gRPC server...")
	s.grpcServer.GracefulStop()

	fmt.Println("gRPC server stopped")
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
	s.eventsMutex.Lock()
	defer s.eventsMutex.Unlock()

	var sequenceStart, sequenceEnd sdkmath.Int

	// Use the requested sequence start if it is not nil. Otherwise, set the
	// sequence start to zero.
	if req.SequenceStart.IsNil() {
		sequenceStart = sdkmath.ZeroInt()
	} else {
		sequenceStart = *req.SequenceStart
	}

	// Use the requested sequence end if it is not nil. Otherwise, set the
	// sequence end to the total amount of events so far.
	if req.SequenceEnd.IsNil() {
		sequenceEnd = sdkmath.NewInt(int64(len(s.events)))
	} else {
		sequenceEnd = *req.SequenceEnd
	}

	// Filter events that fit into the requested range.
	filteredEvents := []*bridgetypes.AssetsLockedEvent{}
	for _, event := range s.events {
		if event.Sequence.GTE(sequenceStart) && event.Sequence.LT(sequenceEnd) {
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
