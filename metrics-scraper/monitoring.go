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
	wg *sync.WaitGroup,
	config NodeConfig,
) (c *rpc.Client, err error) {
	c, err = rpc.DialContext(ctx, config.RPCURL)
	if err == nil {
		// no errors, let's start polling
		return
	}

	ticker := time.NewTicker(5 * time.Second) // arbitrarily retry every 5 seconds
	for {
		select {
		case <-ctx.Done():
			wg.Done()
			log.Printf("job terminated for %v", config.Moniker)
			return
		case <-ticker.C:
			c, err = rpc.DialContext(ctx, config.RPCURL)
			if err != nil {
				log.Printf("error: couldn't connect to %v at %v: %v", config.Moniker, config.RPCURL, err)
				continue
			}

			return
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

	c, err := tryConnect(ctx, wg, config)
	if err != nil {
		// only case this happen is that the context been cancelled
		return
	}

	ticker := time.NewTicker(pollRate)
	for {
		select {
		case <-ctx.Done():
			wg.Done()
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

	return errors.Join(errs...)
}

func latestBlockAndTimestamp(ctx context.Context, client *rpc.Client, moniker, networkID string) error {
	result := struct {
		Number    string `json:"number"`
		Timestamp string `json:"timestamp"`
	}{}

	err := client.CallContext(ctx, &result, ethGetBlockByNumberEndpoint, "latest", false)
	if err != nil {
		mezodUpGauge.WithLabelValues(moniker, networkID).Set(0)
		// set latest block to 0 for error
		mezoLatestBlockGauge.WithLabelValues(moniker, networkID).Set(0)
		// set latest timestamp to 0 for error
		mezoLatestTimestampGauge.WithLabelValues(moniker, networkID).Set(0)
		return fmt.Errorf("couldn't call %v for %v: %v", ethGetBlockByNumberEndpoint, moniker, err)
	}

	mezodUpGauge.WithLabelValues(moniker, networkID).Set(1)

	errs := []error{}
	latestBlock, err := strconv.ParseUint(strings.TrimPrefix(result.Number, "0x"), 16, 64)
	if err != nil {
		// set latest block to 0 for error
		mezoLatestBlockGauge.WithLabelValues(moniker, networkID).Set(0)
		errs = append(errs, fmt.Errorf("invalid latestBlock: %v", err))
	} else {
		mezoLatestBlockGauge.WithLabelValues(moniker, networkID).Set(float64(latestBlock))
	}

	ts, err := strconv.ParseUint(strings.TrimPrefix(result.Timestamp, "0x"), 16, 64)
	if err != nil {
		// set latest timestamp to 0 for error
		mezoLatestTimestampGauge.WithLabelValues(moniker, networkID).Set(0)
		errs = append(errs, fmt.Errorf("invalid timestamp: %v", err))
	} else {
		mezoLatestTimestampGauge.WithLabelValues(moniker, networkID).Set(float64(ts))
	}
	return errors.Join(errs...)
}

func nodeVersion(ctx context.Context, client *rpc.Client, moniker, networkID string) error {
	var result string
	err := client.CallContext(ctx, &result, web3ClientVersionEndpoint)
	if err != nil {
		// set it to a valid semver showing nicely there's an error
		mezodVersionGauge.WithLabelValues(moniker, networkID, "0.0.0-unknown").Set(1)
		return fmt.Errorf("couldn't call %v for %v: %v", web3ClientVersionEndpoint, moniker, err)
	}

	// here we expect the following pattern:
	// Mezod/<VERSION>/amd64/go1.22.8
	segments := strings.Split(result, "/")
	if len(segments) != 4 {
		// set it to a valid semver showing nicely there's an error
		mezodVersionGauge.WithLabelValues(moniker, networkID, "0.0.0-unknown").Set(1)
		return fmt.Errorf("invalid version string, expected 4 segments, got %v: %v", len(segments), result)
	}

	mezodVersionGauge.WithLabelValues(moniker, networkID, segments[1]).Set(1)

	return nil
}

func sidecarsVersion(ctx context.Context, client *rpc.Client, moniker, networkID string) error {
	result := map[string]net.SidecarInfos{}
	err := client.CallContext(ctx, &result, netSidecarsEndpoint)
	if err != nil {
		// just format nicely the error, and continue execution
		// so the prometheus variable are still set to defaults
		err = fmt.Errorf("couldn't call %v for %v: %v", netSidecarsEndpoint, moniker, err)
	}

	var (
		ethereumVersion, ethereumIsRunning = "0.0.0-unknown", 0.
		connectVersion, connectIsRunning   = "0.0.0-unknown", 0.
	)

	if ethereumSidecar, ok := result["ethereum"]; ok {
		if ethereumSidecar.Connected {
			ethereumIsRunning = 1
		}
		ethereumVersion = ethereumSidecar.Version
	}
	ethereumSidecarGauge.WithLabelValues(moniker, networkID, ethereumVersion).Set(ethereumIsRunning)

	if connectSidecar, ok := result["connect"]; ok {
		if connectSidecar.Connected {
			connectIsRunning = 1
		}
		connectVersion = connectSidecar.Version
	}
	connectSidecarGauge.WithLabelValues(moniker, networkID, connectVersion).Set(connectIsRunning)

	return err
}
