package types

import (
	"context"
	"fmt"
	"sync"
	"time"

	"cosmossdk.io/log"
	rpcclient "github.com/cometbft/cometbft/rpc/jsonrpc/client"
)

// CometWSClient is a wrapper around the cometbft rpc client that provides a
// mechanism to restore subscriptions after a reconnect.
type CometWSClient struct {
	*rpcclient.WSClient

	logger             log.Logger
	subscriptionsMutex sync.Mutex
	subscriptions      map[string]bool
}

func ConnectCometWS(logger log.Logger, address, endpoint, moniker string) (*CometWSClient, error) {
	clientLogger := logger.With(
		"client-moniker", moniker,
		"server-address", address+endpoint,
	)

	reconnectedCh := make(chan struct{}, 1)

	wsClient, err := rpcclient.NewWS(address, endpoint,
		rpcclient.MaxReconnectAttempts(256),
		rpcclient.ReadWait(120*time.Second),
		rpcclient.WriteWait(120*time.Second),
		rpcclient.PingPeriod(50*time.Second),
		rpcclient.OnReconnect(func() {
			reconnectedCh <- struct{}{}
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("internal cometbft ws client could not be created: %w", err)
	} else if err := wsClient.OnStart(); err != nil {
		return nil, fmt.Errorf("internal cometbft ws client could not be started: %w", err)
	}

	wrapper := &CometWSClient{
		logger:        clientLogger,
		WSClient:      wsClient,
		subscriptions: make(map[string]bool),
	}

	go func() {
		for {
			select {
			case <-reconnectedCh:
				clientLogger.Info("internal cometbft ws client reconnected")
				wrapper.restoreSubscriptions()
			case <-wsClient.Quit():
				clientLogger.Warn("internal cometbft ws client quit")
				return
			}
		}
	}()

	return wrapper, nil
}

func (cwc *CometWSClient) restoreSubscriptions() {
	cwc.subscriptionsMutex.Lock()
	defer cwc.subscriptionsMutex.Unlock()

	ctx, cancelFn := context.WithCancel(context.Background())
	defer cancelFn()

	for query := range cwc.subscriptions {
		// This is a fresh connection so there shouldn't be any subscriptions but unsubscribe just in case.
		err := cwc.WSClient.Unsubscribe(ctx, query)
		if err != nil {
			cwc.logger.Error(
				"internal cometbft ws client failed to unsubscribe from query while restoring subscriptions",
				"query", query,
				"error", err,
			)
		}

		err = cwc.WSClient.Subscribe(ctx, query)
		if err != nil {
			cwc.logger.Error(
				"internal cometbft ws client failed to subscribe to query while restoring subscriptions",
				"query", query,
				"error", err,
			)
		}
	}

	cwc.logger.Info(
		"internal cometbft ws client restored subscriptions after reconnect",
		"subscriptions", cwc.subscriptions,
	)
}

func (cwc *CometWSClient) Subscribe(ctx context.Context, query string) error {
	cwc.subscriptionsMutex.Lock()
	defer cwc.subscriptionsMutex.Unlock()

	cwc.logger.Debug("internal cometbft ws client subscribing to query", "query", query)

	err := cwc.WSClient.Subscribe(ctx, query)
	if err != nil {
		return err
	}

	cwc.subscriptions[query] = true

	return nil
}

func (cwc *CometWSClient) Unsubscribe(ctx context.Context, query string) error {
	cwc.subscriptionsMutex.Lock()
	defer cwc.subscriptionsMutex.Unlock()

	cwc.logger.Debug("internal cometbft ws client unsubscribing from query", "query", query)

	err := cwc.WSClient.Unsubscribe(ctx, query)
	if err != nil {
		return err
	}

	delete(cwc.subscriptions, query)

	return nil
}

func (cwc *CometWSClient) UnsubscribeAll(ctx context.Context) error {
	cwc.subscriptionsMutex.Lock()
	defer cwc.subscriptionsMutex.Unlock()

	cwc.logger.Debug("internal cometbft ws client unsubscribing from all queries")

	err := cwc.WSClient.UnsubscribeAll(ctx)
	if err != nil {
		return err
	}

	for query := range cwc.subscriptions {
		delete(cwc.subscriptions, query)
	}

	return nil
}
