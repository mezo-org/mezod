package electrum

import (
	"context"

	"github.com/mezo-org/mezod/bridge-worker/bitcoin"
)

// Connect initializes handle with provided Config.
func Connect(parentCtx context.Context, config Config) (bitcoin.Chain, error) {
	// if config.ConnectTimeout == 0 {
	// 	config.ConnectTimeout = DefaultConnectTimeout
	// }
	// if config.ConnectRetryTimeout == 0 {
	// 	config.ConnectRetryTimeout = DefaultConnectRetryTimeout
	// }
	// if config.RequestTimeout == 0 {
	// 	config.RequestTimeout = DefaultRequestTimeout
	// }
	// if config.RequestRetryTimeout == 0 {
	// 	config.RequestRetryTimeout = DefaultRequestRetryTimeout
	// }
	// if config.KeepAliveInterval == 0 {
	// 	config.KeepAliveInterval = DefaultKeepAliveInterval
	// }

	// c := &Connection{
	// 	parentCtx:   parentCtx,
	// 	config:      config,
	// 	clientMutex: &sync.Mutex{},
	// }

	// if err := c.electrumConnect(); err != nil {
	// 	return nil, fmt.Errorf("failed to initialize electrum client: [%w]", err)
	// }

	// if err := c.verifyServer(); err != nil {
	// 	return nil, fmt.Errorf("failed to verify electrum server: [%w]", err)
	// }

	// // Keep the connection alive and check the connection health.
	// go c.keepAlive()

	// return c, nil

	// TODO: Implement
	return nil, nil
}
