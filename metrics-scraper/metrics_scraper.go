package metricsscraper

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type Config struct {
	NodesConfigPath string
	PrometheusPort  uint
	ChainID         string
	NodePollRate    time.Duration
	BridgePollRate  time.Duration
}

func Start(config Config) error {
	ctx := context.Background()

	nodesConfig, err := loadNodesConfig(config.NodesConfigPath)
	if err != nil {
		return fmt.Errorf("couldn't load nodes config: [%w]", err)
	}

	for _, n := range nodesConfig.Nodes {
		go func(nodeConfig NodeConfig) {
			runNodeMonitoring(ctx, nodeConfig, config.ChainID, config.NodePollRate)
		}(n)
	}

	go runBridgeMonitoring(ctx, config.ChainID, config.BridgePollRate)
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
