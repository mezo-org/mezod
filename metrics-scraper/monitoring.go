package metricsscraper

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/mezo-org/mezod/rpc/namespaces/ethereum/net"
)

const (
	netSidecarsEndpoint         = "net_sidecars"
	web3ClientVersionEndpoint   = "web3_clientVersion"
	ethGetBlockByNumberEndpoint = "eth_getBlockByNumber"
)

func Start(configPath string) {
	buf, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatalf("errror: reading config file: %s", err)
	}

	config := Config{}
	err = json.Unmarshal(buf, &config)
	if err != nil {
		log.Fatalf("error: unmarshalling config file: %v", err)
	}

	if len(config.Nodes) == 0 {
		log.Fatalf("error: empty config")
	}

	log.Printf("monitoring network %v", config.ChainID)
	log.Printf("poll rate %v", config.PollRate.Duration)

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	for _, c := range config.Nodes {
		wg.Add(1)

		log.Printf("starting monitoring for %v", c.Moniker)

		// start the job
		go run(ctx, &wg, config.PollRate.Duration, config.ChainID, c)
	}

	log.Printf("monitoring %v nodes", len(config.Nodes))

	log.Printf("starting prometheus")
	go startPrometheus()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	log.Println("monitoring started press ctrl+c to terminate")
	<-sigChan

	// cancel context
	cancel()
	// wait for the waitgroup to terminate
	wg.Wait()
}

// tryConnect, try to connect to the node forever until it works
// it'll stop when it acquire a connection or it's asked to stop
func tryConnect(
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
			log.Printf("job terminated for %v", config.Moniker)
			return nil, false
		case <-ticker.C:
			c, err = rpc.DialContext(ctx, config.RPCURL)
			if err != nil {
				log.Printf("error: couldn't connect to %v at %v: %v", config.Moniker, config.RPCURL, err)
				continue
			}

			return c, true
		}
	}
}

func run(
	ctx context.Context,
	wg *sync.WaitGroup,
	pollRate time.Duration,
	networkID string,
	config NodeConfig,
) {
	defer wg.Done()

	c, ok := tryConnect(ctx, pollRate, config)
	if !ok {
		// only case this happen is that the context been canceled
		return
	}

	ticker := time.NewTicker(pollRate)
	for {
		select {
		case <-ctx.Done():
			log.Printf("job terminated for %v", config.Moniker)
			return
		case <-ticker.C:
			if err := pollData(ctx, c, config.Moniker, networkID); err != nil {
				log.Printf("error polling data for %v: %v", config.Moniker, err)
			} else {
				log.Printf("data polled successfully for: %v", config.Moniker)
			}
		}
	}
}

func pollData(ctx context.Context, client *rpc.Client, moniker, networkID string) error {
	errs := []error{}
	if err := sidecarsVersion(ctx, client, moniker, networkID); err != nil {
		errs = append(errs, err)
	}

	if err := nodeVersion(ctx, client, moniker, networkID); err != nil {
		errs = append(errs, err)
	}

	if err := latestBlockAndTimestamp(ctx, client, moniker, networkID); err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		mezodUpGauge.WithLabelValues(moniker, networkID).Set(0)
	} else {
		mezodUpGauge.WithLabelValues(moniker, networkID).Set(1)
	}

	return errors.Join(errs...)
}

func latestBlockAndTimestamp(ctx context.Context, client *rpc.Client, moniker, networkID string) (err error) {
	defer func() {
		if err != nil {
			// set latest block to 0 for error
			mezoLatestBlockGauge.WithLabelValues(moniker, networkID).Set(0)
			// set latest timestamp to 0 for error
			mezoLatestTimestampGauge.WithLabelValues(moniker, networkID).Set(0)
		}
	}()

	result := struct {
		Number    string `json:"number"`
		Timestamp string `json:"timestamp"`
	}{}

	err = client.CallContext(ctx, &result, ethGetBlockByNumberEndpoint, "latest", false)
	if err != nil {
		return fmt.Errorf("couldn't call %v for %v: %v", ethGetBlockByNumberEndpoint, moniker, err)
	}

	errs := []error{}
	latestBlock, err := strconv.ParseUint(strings.TrimPrefix(result.Number, "0x"), 16, 64)
	if err != nil {
		errs = append(errs, fmt.Errorf("invalid latestBlock: %v", err))
	} else {
		mezoLatestBlockGauge.WithLabelValues(moniker, networkID).Set(float64(latestBlock))
	}

	ts, err := strconv.ParseUint(strings.TrimPrefix(result.Timestamp, "0x"), 16, 64)
	if err != nil {
		errs = append(errs, fmt.Errorf("invalid timestamp: %v", err))
	} else {
		mezoLatestTimestampGauge.WithLabelValues(moniker, networkID).Set(float64(ts))
	}

	return errors.Join(errs...)
}

func nodeVersion(ctx context.Context, client *rpc.Client, moniker, networkID string) (err error) {
	defer func() {
		if err != nil {
			// set it to a valid semver showing nicely there's an error
			mezodVersionGauge.WithLabelValues(moniker, networkID, "0.0.0-unknown").Set(1)
		}
	}()

	// always remove the previous metrics to ensure that there's no duplicates when
	// the labels version changes
	mezodVersionGauge.DeletePartialMatch(map[string]string{
		"moniker":  moniker,
		"chain_id": networkID,
	})

	var result string
	err = client.CallContext(ctx, &result, web3ClientVersionEndpoint)
	if err != nil {
		return fmt.Errorf("couldn't call %v for %v: %v", web3ClientVersionEndpoint, moniker, err)
	}

	// here we expect the following pattern:
	// Mezod/<VERSION>/amd64/go1.22.8
	segments := strings.Split(result, "/")
	if len(segments) != 4 {
		return fmt.Errorf("invalid version string, expected 4 segments, got %v: %v", len(segments), result)
	}

	mezodVersionGauge.WithLabelValues(moniker, networkID, segments[1]).Set(1)

	return nil
}

func sidecarsVersion(ctx context.Context, client *rpc.Client, moniker, networkID string) (err error) {
	defer func() {
		if err != nil {
			ethereumSidecarGauge.WithLabelValues(moniker, networkID, "0.0.0-unknown").Set(0)
			connectSidecarGauge.WithLabelValues(moniker, networkID, "0.0.0-unknown").Set(0)
		}
	}()

	// always remove the previous metrics to ensure that there's no duplicates when
	// the labels version changes
	ethereumSidecarGauge.DeletePartialMatch(map[string]string{
		"moniker":  moniker,
		"chain_id": networkID,
	})
	connectSidecarGauge.DeletePartialMatch(map[string]string{
		"moniker":  moniker,
		"chain_id": networkID,
	})

	result := map[string]net.SidecarInfos{}
	err = client.CallContext(ctx, &result, netSidecarsEndpoint)
	if err != nil {
		// just format nicely the error, and continue execution
		// so the prometheus variable are still set to defaults
		return fmt.Errorf("couldn't call %v for %v: %v", netSidecarsEndpoint, moniker, err)
	}

	if ethereumSidecar, ok := result["ethereum"]; ok {
		var isConnected float64
		if ethereumSidecar.Connected {
			isConnected = 1
		}
		ethereumSidecarGauge.WithLabelValues(moniker, networkID, ethereumSidecar.Version).Set(isConnected)
	}

	if connectSidecar, ok := result["connect"]; ok {
		var isConnected float64
		if connectSidecar.Connected {
			isConnected = 1
		}
		connectSidecarGauge.WithLabelValues(moniker, networkID, connectSidecar.Version).Set(isConnected)
	}

	return err
}
