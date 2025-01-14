// Copyright 2022 Evmos Foundation
// This file is part of the Evmos Network packages.
//
// Evmos is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Evmos packages are distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Evmos packages. If not, see https://github.com/evmos/evmos/blob/main/LICENSE
package net

import (
	"context"
	"fmt"
	"time"

	"cosmossdk.io/log"
	rpcclient "github.com/cometbft/cometbft/rpc/client"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/mezo-org/mezod/ethereum/sidecar"
	"github.com/mezo-org/mezod/server/config"
	"github.com/mezo-org/mezod/types"
	oracleclient "github.com/skip-mev/connect/v2/service/clients/oracle"
	servicemetrics "github.com/skip-mev/connect/v2/service/metrics"
	oracletypes "github.com/skip-mev/connect/v2/service/servers/oracle/types"
)

// PublicAPI is the eth_ prefixed set of APIs in the Web3 JSON-RPC spec.
type PublicAPI struct {
	logger                log.Logger
	networkVersion        uint64
	tmClient              rpcclient.Client
	oracleClient          oracleclient.OracleClient
	ethereumSidecarClient *sidecar.Client
}

// NewPublicAPI creates an instance of the public Net Web3 API.
func NewPublicAPI(
	ctx *server.Context,
	clientCtx client.Context,
) *PublicAPI {
	appConf, err := config.GetConfig(ctx.Viper)
	if err != nil {
		panic(err)
	}

	oracleClient, err := oracleclient.NewClientFromConfig(
		appConf.Oracle, ctx.Logger, servicemetrics.NewNopMetrics(),
	)
	if err != nil {
		panic(err)
	}

	err = oracleClient.Start(context.Background())
	if err != nil {
		panic(err)
	}

	ethereumSidecarClient, err := sidecar.NewClient(
		ctx.Logger,
		appConf.EthereumSidecar.ServerAddress,
		appConf.EthereumSidecar.RequestTimeout,
		clientCtx.InterfaceRegistry,
	)
	if err != nil {
		panic(err)
	}

	// parse the chainID from a integer string
	chainIDEpoch, err := types.ParseChainID(clientCtx.ChainID)
	if err != nil {
		panic(err)
	}

	return &PublicAPI{
		logger:                ctx.Logger,
		networkVersion:        chainIDEpoch.Uint64(),
		tmClient:              clientCtx.Client.(rpcclient.Client),
		oracleClient:          oracleClient,
		ethereumSidecarClient: ethereumSidecarClient,
	}
}

// Version returns the current ethereum protocol version.
func (s *PublicAPI) Version() string {
	return fmt.Sprintf("%d", s.networkVersion)
}

// Listening returns if client is actively listening for network connections.
func (s *PublicAPI) Listening() bool {
	ctx := context.Background()
	netInfo, err := s.tmClient.NetInfo(ctx)
	if err != nil {
		return false
	}
	return netInfo.Listening
}

// PeerCount returns the number of peers currently connected to the client.
func (s *PublicAPI) PeerCount() int {
	ctx := context.Background()
	netInfo, err := s.tmClient.NetInfo(ctx)
	if err != nil {
		return 0
	}
	return len(netInfo.Peers)
}

type SidecarInfos struct {
	Version   string
	Connected bool
}

// Sidecars returns informations about the ethereum
func (s *PublicAPI) Sidecars() map[string]SidecarInfos {
	// FIXME(jeremy): use better timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var (
		connectVersion  = "unknown"
		connectStatus   = false
		ethereumVersion = "unknown"
		ethereumStatus  = false
	)

	resp, err := s.oracleClient.Version(ctx, &oracletypes.QueryVersionRequest{})
	if err != nil {
		s.logger.Error("couldn't reach oracle", "error", err)
	} else {
		connectVersion = resp.Version
		connectStatus = true
	}

	ethv, err := s.ethereumSidecarClient.Version(ctx)
	if err != nil {
		s.logger.Error("couldn't reach ethereum sidecar", "error", err)
	} else {
		ethereumVersion = ethv
		ethereumStatus = true
	}

	return map[string]SidecarInfos{
		"ethereum": {
			Version:   ethereumVersion,
			Connected: ethereumStatus,
		},
		"connect": {
			Version:   connectVersion,
			Connected: connectStatus,
		},
	}
}
