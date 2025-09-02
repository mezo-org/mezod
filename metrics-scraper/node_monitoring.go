package metricsscraper

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/mezo-org/mezod/rpc/namespaces/ethereum/net"
)

const (
	netSidecarsEndpoint         = "net_sidecars"
	web3ClientVersionEndpoint   = "web3_clientVersion"
	ethGetBlockByNumberEndpoint = "eth_getBlockByNumber"
)

func runNodeMonitoring(
	ctx context.Context,
	pollRate time.Duration,
	chainID string,
	config NodeConfig,
) {
	log.Printf("starting node [%v] monitoring", config.Moniker)

	c, ok := tryConnectNode(ctx, pollRate, config)
	if !ok {
		// only case this happen is that the context been canceled
		return
	}

	ticker := time.NewTicker(pollRate)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Printf("terminated node [%v] monitoring", config.Moniker)
			return
		case <-ticker.C:
			if err := pollNodeData(ctx, c, config.Moniker, chainID); err != nil {
				log.Printf("error while polling node [%v] data: [%v]", config.Moniker, err)
			} else {
				log.Printf("node [%v] data polled successfully", config.Moniker)
			}
		}
	}
}

// tryConnect, try to connect to the node forever until it works
// it'll stop when it acquire a connection or it's asked to stop
func tryConnectNode(
	ctx context.Context,
	pollRate time.Duration,
	config NodeConfig,
) (*rpc.Client, bool) {
	c, err := rpc.DialContext(ctx, config.RPCURL)
	if err == nil {
		// no errors, let's start polling
		return c, true
	}

	ticker := time.NewTicker(pollRate)
	for {
		select {
		case <-ctx.Done():
			log.Printf("connection job terminated for node [%v]", config.Moniker)
			return nil, false
		case <-ticker.C:
			c, err = rpc.DialContext(ctx, config.RPCURL)
			if err != nil {
				log.Printf("couldn't connect to node [%v] at [%v]: [%v]", config.Moniker, config.RPCURL, err)
				continue
			}

			return c, true
		}
	}
}

func pollNodeData(ctx context.Context, client *rpc.Client, moniker, chainID string) error {
	errs := []error{}
	if err := sidecarsVersion(ctx, client, moniker, chainID); err != nil {
		errs = append(errs, err)
	}

	if err := nodeVersion(ctx, client, moniker, chainID); err != nil {
		errs = append(errs, err)
	}

	if err := latestBlockAndTimestamp(ctx, client, moniker, chainID); err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		mezodUpGauge.WithLabelValues(moniker, chainID).Set(0)
	} else {
		mezodUpGauge.WithLabelValues(moniker, chainID).Set(1)
	}

	return errors.Join(errs...)
}

func latestBlockAndTimestamp(ctx context.Context, client *rpc.Client, moniker, chainID string) (err error) {
	defer func() {
		if err != nil {
			// set latest block to 0 for error
			mezoLatestBlockGauge.WithLabelValues(moniker, chainID).Set(0)
			// set latest timestamp to 0 for error
			mezoLatestTimestampGauge.WithLabelValues(moniker, chainID).Set(0)
		}
	}()

	result := struct {
		Number    string `json:"number"`
		Timestamp string `json:"timestamp"`
	}{}

	err = client.CallContext(ctx, &result, ethGetBlockByNumberEndpoint, "latest", false)
	if err != nil {
		err = fmt.Errorf("couldn't call [%v] for node [%v]: [%w]", ethGetBlockByNumberEndpoint, moniker, err)
		return
	}

	latestBlock, err := strconv.ParseUint(strings.TrimPrefix(result.Number, "0x"), 16, 64)
	if err != nil {
		err = fmt.Errorf("invalid latestBlock: [%w]", err)
		return
	}

	ts, err := strconv.ParseUint(strings.TrimPrefix(result.Timestamp, "0x"), 16, 64)
	if err != nil {
		err = fmt.Errorf("invalid timestamp: [%w]", err)
		return
	}

	mezoLatestTimestampGauge.WithLabelValues(moniker, chainID).Set(float64(ts))
	mezoLatestBlockGauge.WithLabelValues(moniker, chainID).Set(float64(latestBlock))

	return
}

func nodeVersion(ctx context.Context, client *rpc.Client, moniker, chainID string) (err error) {
	defer func() {
		if err != nil {
			// set it to a valid semver showing nicely there's an error
			mezodVersionGauge.WithLabelValues(moniker, chainID, "0.0.0-unknown").Set(1)
		}
	}()

	// always remove the previous metrics to ensure that there's no duplicates when
	// the labels version changes
	mezodVersionGauge.DeletePartialMatch(map[string]string{
		"moniker":  moniker,
		"chain_id": chainID,
	})

	var result string
	err = client.CallContext(ctx, &result, web3ClientVersionEndpoint)
	if err != nil {
		err = fmt.Errorf("couldn't call [%v] for node [%v]: [%w]", web3ClientVersionEndpoint, moniker, err)
		return
	}

	// here we expect the following pattern:
	// Mezod/<VERSION>/amd64/go1.22.8
	segments := strings.Split(result, "/")
	if len(segments) != 4 {
		err = fmt.Errorf("invalid version string, expected 4 segments, got [%v]: [%v]", len(segments), result)
		return
	}

	mezodVersionGauge.WithLabelValues(moniker, chainID, segments[1]).Set(1)

	return
}

func sidecarsVersion(ctx context.Context, client *rpc.Client, moniker, chainID string) (err error) {
	defer func() {
		if err != nil {
			ethereumSidecarGauge.WithLabelValues(moniker, chainID, "0.0.0-unknown").Set(0)
			connectSidecarGauge.WithLabelValues(moniker, chainID, "0.0.0-unknown").Set(0)
		}
	}()

	// always remove the previous metrics to ensure that there's no duplicates when
	// the labels version changes
	ethereumSidecarGauge.DeletePartialMatch(map[string]string{
		"moniker":  moniker,
		"chain_id": chainID,
	})
	connectSidecarGauge.DeletePartialMatch(map[string]string{
		"moniker":  moniker,
		"chain_id": chainID,
	})

	result := map[string]net.SidecarInfos{}
	err = client.CallContext(ctx, &result, netSidecarsEndpoint)
	if err != nil {
		err = fmt.Errorf("couldn't call [%v] for node [%v]: [%w]", netSidecarsEndpoint, moniker, err)
		return
	}

	if ethereumSidecar, ok := result["ethereum"]; ok {
		var isConnected float64
		if ethereumSidecar.Connected {
			isConnected = 1
		}
		ethereumSidecarGauge.WithLabelValues(moniker, chainID, ethereumSidecar.Version).Set(isConnected)
	}

	if connectSidecar, ok := result["connect"]; ok {
		var isConnected float64
		if connectSidecar.Connected {
			isConnected = 1
		}
		connectSidecarGauge.WithLabelValues(moniker, chainID, connectSidecar.Version).Set(isConnected)
	}

	return
}
