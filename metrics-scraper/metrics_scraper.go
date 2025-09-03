package metricsscraper

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
)

type Config struct {
	NodesConfigPath string
	PrometheusPort  uint
	ChainID         string
	NodePollRate    time.Duration
	BridgePollRate  time.Duration
	MezoRpcUrl      string
	EthereumRpcUrl  string
}

func Start(config Config) error {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	nodesConfig, err := loadNodesConfig(config.NodesConfigPath)
	if err != nil {
		return fmt.Errorf("couldn't load nodes config: [%w]", err)
	}

	for _, n := range nodesConfig.Nodes {
		go func(nodeConfig NodeConfig) {
			runNodeMonitoring(ctx, nodeConfig, config.ChainID, config.NodePollRate)
		}(n)
	}

	go func() {
		// If this routine exits, cancel the context to
		// shutdown the entire process.
		defer cancelCtx()

		err := runBridgeMonitoring(
			ctx,
			config.ChainID,
			config.BridgePollRate,
			config.MezoRpcUrl,
			config.EthereumRpcUrl,
		)
		if err != nil {
			log.Printf("bridge monitoring routine exited with error: [%v]", err)
		}
	}()
	go startPrometheus(config.PrometheusPort)

	<-ctx.Done()

	return nil
}

func loadNodesConfig(nodesConfigPath string) (NodesConfig, error) {
	buffer, err := os.ReadFile(nodesConfigPath)
	if err != nil {
		return NodesConfig{}, fmt.Errorf("couldn't read nodes config file: [%w]", err)
	}

	nodesConfig := NodesConfig{}
	err = json.Unmarshal(buffer, &nodesConfig)
	if err != nil {
		return NodesConfig{}, fmt.Errorf("couldn't unmarshall nodes config file: [%w]", err)
	}

	return nodesConfig, nil
}
