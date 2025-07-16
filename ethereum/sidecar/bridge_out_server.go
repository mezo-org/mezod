package sidecar

import (
	"context"
	"fmt"
	"net"

	"cosmossdk.io/log"
	pb "github.com/mezo-org/mezod/ethereum/sidecar/types"
	bridgekeeper "github.com/mezo-org/mezod/x/bridge/keeper"

	"google.golang.org/grpc"
)

// BridgeOutServer handles requests for `AssetsUnlocked` entries from Ethereum
// sidecar. It is intended to be used as part of the `mezod` validator node.
type BridgeOutServer struct {
	logger log.Logger

	grpcServer   *grpc.Server
	bridgeKeeper *bridgekeeper.Keeper
}

// RunBridgeOutServer initializes the bridge-out sidecar server and starts the
// gRPC server for handling `AssetsUnlocked` requests.
func RunBridgeOutServer(
	logger log.Logger,
	grpcAddress string,
	bridgeKeeper *bridgekeeper.Keeper,
) {
	logger.Info(
		"starting bridge-out sidecar server",
		"grpc_address", grpcAddress,
	)

	server := &BridgeOutServer{
		logger:       logger,
		grpcServer:   grpc.NewServer(),
		bridgeKeeper: bridgeKeeper,
	}

	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	go func() {
		defer cancelCtx()
		err := server.startGRPCServer(ctx, grpcAddress)
		if err != nil {
			server.logger.Error("gRPC server routine failed", "err", err)
		}

		server.logger.Info("gRPC server routine stopped")
	}()

	<-ctx.Done()

	server.logger.Error("bridge-out sidecar server stopped")
}

// startGRPCServer starts the gRPC server and registers the Ethereum sidecar
// bridge-out service.
func (bos *BridgeOutServer) startGRPCServer(
	ctx context.Context,
	address string,
) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to listen: [%w]", err)
	}

	pb.RegisterEthereumSidecarBridgeOutServiceServer(bos.grpcServer, bos)

	bos.logger.Info(
		"gRPC server started",
		"address", address,
	)

	defer bos.grpcServer.GracefulStop()

	errChan := make(chan error)

	go func() {
		err := bos.grpcServer.Serve(listener)
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

func (bos *BridgeOutServer) AssetsUnlockedEntries(
	_ context.Context,
	_ *pb.AssetsUnlockedEntriesRequest,
) (
	*pb.AssetsUnlockedEntriesResponse,
	error,
) {
	// TODO: Implement fetching of `AssetsUnlocked` entries from the bridge
	//       keeper.
	return nil, nil
}

func (bos *BridgeOutServer) AssetsUnlockedSequenceTip(
	_ context.Context,
	_ *pb.AssetsUnlockedSequenceTipRequest,
) (
	*pb.AssetsUnlockedSequenceTipResponse,
	error,
) {
	// TODO: Implement fetching of `AssetsUnlockedSequenceTip` from the bridge
	//       keeper.
	return nil, nil
}
