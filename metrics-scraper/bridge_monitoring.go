package metricsscraper

import (
	"context"
	"log"
	"time"
)

func runBridgeMonitoring(ctx context.Context, chainID string, pollRate time.Duration) {
	log.Printf("starting bridge monitoring")

	ticker := time.NewTicker(pollRate)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Printf("terminated bridge monitoring")
			return
		case <-ticker.C:
			if err := pollBridgeData(ctx, chainID); err != nil {
				log.Printf("error while polling bridge data: %v", err)
			} else {
				log.Printf("bridge data polled successfully")
			}
		}
	}
}

func pollBridgeData(_ context.Context, _ string) error {
	// TODO: Implement bridge data polling and expose metrics.
	return nil
}
