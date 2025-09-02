package metricsscraper

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func Start(configPath string) {
	buf, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatalf("couldn't read config file: [%v]", err)
	}

	config := Config{}
	err = json.Unmarshal(buf, &config)
	if err != nil {
		log.Fatalf("couldn't unmarshall config file: [%v]", err)
	}

	log.Printf("monitoring Mezo chain [%v]", config.ChainID)

	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	for _, n := range config.Nodes {
		wg.Add(1)

		// start the job
		go func(nodeConfig NodeConfig) {
			defer wg.Done()
			runNodeMonitoring(ctx, config.NodePollRate.Duration, config.ChainID, nodeConfig)
		}(n)
	}

	go runBridgeMonitoring(ctx, config.BridgePollRate.Duration, config.ChainID)

	go startPrometheus()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	log.Println("monitoring started; listening for SIGINT or SIGTERM to terminate")
	<-sigChan

	cancel()
	wg.Wait()
}
