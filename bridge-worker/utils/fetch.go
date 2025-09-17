package utils

import (
	"fmt"
	"time"
)

// WithBatchFetch fetches events from the given range using the provided
// fetch function. It first tries to fetch the events from the entire range and
// if it fails it switches to batch fetching.
func WithBatchFetch[T any](
	fetchFunc func(batchStartHeight, batchEndHeight uint64) ([]T, error),
	startHeight uint64,
	endHeight uint64,
	requestsPerMinute uint64,
	batchSize uint64,
) ([]T, error) {
	if requestsPerMinute == 0 {
		requestsPerMinute = 600
	}
	if batchSize == 0 {
		batchSize = 1000
	}

	result := make([]T, 0)

	ticker := time.NewTicker(time.Minute / time.Duration(requestsPerMinute)) //nolint:gosec
	defer ticker.Stop()

	events, err := fetchFunc(startHeight, endHeight)
	if err != nil {
		batchStartHeight := startHeight

		for batchStartHeight <= endHeight {
			batchEndHeight := batchStartHeight + batchSize
			if batchEndHeight > endHeight {
				batchEndHeight = endHeight
			}

			<-ticker.C

			batchEvents, batchErr := fetchFunc(batchStartHeight, batchEndHeight)
			if batchErr != nil {
				return nil, fmt.Errorf(
					"batched fetch failed: [%w]; giving up",
					batchErr,
				)
			}

			result = append(result, batchEvents...)

			batchStartHeight = batchEndHeight + 1
		}
	} else {
		result = append(result, events...)
	}

	return result, nil
}
