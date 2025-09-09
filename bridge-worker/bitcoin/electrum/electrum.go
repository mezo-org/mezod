package electrum

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"golang.org/x/exp/slices"

	"github.com/checksum0/go-electrum/electrum"
	"github.com/ipfs/go-log"
	"github.com/keep-network/keep-common/pkg/wrappers"
	"github.com/mezo-org/mezod/bridge-worker/bitcoin"
)

var (
	supportedProtocolVersions = []string{"1.4"}
	logger                    = log.Logger("electrum")
)

// Connection is a handle for interactions with Electrum server.
type Connection struct {
	parentCtx   context.Context
	client      *electrum.Client
	clientMutex *sync.Mutex
	config      Config
}

// Connect initializes handle with provided Config.
func Connect(parentCtx context.Context, config Config) (bitcoin.Chain, error) {
	c := &Connection{
		parentCtx:   parentCtx,
		config:      config,
		clientMutex: &sync.Mutex{},
	}

	if err := c.electrumConnect(); err != nil {
		return nil, fmt.Errorf("failed to initialize electrum client: [%w]", err)
	}

	if err := c.verifyServer(); err != nil {
		return nil, fmt.Errorf("failed to verify electrum server: [%w]", err)
	}

	// Keep the connection alive and check the connection health.
	go c.keepAlive()

	return c, nil
}

// GetTransaction gets the transaction with the given transaction hash.
// If the transaction with the given hash was not found on the chain,
// this function returns an error.
func (c *Connection) GetTransaction(
	transactionHash bitcoin.Hash,
) (*bitcoin.Transaction, error) {
	txID := transactionHash.Hex(bitcoin.ReversedByteOrder)

	rawTransaction, err := requestWithRetry(
		c,
		func(ctx context.Context, client *electrum.Client) (string, error) {
			// We cannot use `GetTransaction` to get the the transaction details
			// as Esplora/Electrs doesn't support verbose transactions.
			// See: https://github.com/Blockstream/electrs/pull/36
			tx, err := client.GetRawTransaction(ctx, txID)
			if err != nil {
				if isTxNotFoundErr(err) {
					// The transaction was not found on the chain. There is
					// no point in retrying the request and losing time.
					return "", nil
				}

				return "", err
			}

			return tx, nil
		},
		"GetRawTransaction",
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to get raw transaction with ID [%s]: [%w]",
			txID,
			err,
		)
	}
	if len(rawTransaction) == 0 {
		return nil, fmt.Errorf(
			"failed to get raw transaction with ID [%s]: [%v]",
			txID,
			fmt.Errorf("not found"),
		)
	}

	result, err := convertRawTransaction(rawTransaction)
	if err != nil {
		return nil, fmt.Errorf("failed to convert transaction: [%w]", err)
	}

	return result, nil
}

// GetTxHashesForPublicKeyHash gets hashes of confirmed transactions that pays
// the given public key hash using either a P2PKH or P2WPKH script. The returned
// transactions hashes are ordered by block height in the ascending order, i.e.
// the latest transaction hash is at the end of the list. The returned list does
// not contain unconfirmed transactions hashes living in the mempool at the
// moment of request.
func (c *Connection) GetTxHashesForPublicKeyHash(
	publicKeyHash [20]byte,
) ([]bitcoin.Hash, error) {
	p2pkh, err := bitcoin.PayToPublicKeyHash(publicKeyHash)
	if err != nil {
		return nil, fmt.Errorf(
			"cannot build P2PKH for public key hash [0x%x]: [%v]",
			publicKeyHash,
			err,
		)
	}

	p2wpkh, err := bitcoin.PayToWitnessPublicKeyHash(publicKeyHash)
	if err != nil {
		return nil, fmt.Errorf(
			"cannot build P2WPKH for public key hash [0x%x]: [%v]",
			publicKeyHash,
			err,
		)
	}

	p2pkhItems, err := c.getConfirmedScriptHistory(p2pkh)
	if err != nil {
		return nil, fmt.Errorf(
			"cannot get P2PKH history for public key hash [0x%x]: [%v]",
			publicKeyHash,
			err,
		)
	}

	p2wpkhItems, err := c.getConfirmedScriptHistory(p2wpkh)
	if err != nil {
		return nil, fmt.Errorf(
			"cannot get P2WPKH history for public key hash [0x%x]: [%v]",
			publicKeyHash,
			err,
		)
	}

	items := append(p2pkhItems, p2wpkhItems...)

	sort.SliceStable(
		items,
		func(i, j int) bool {
			return items[i].blockHeight < items[j].blockHeight
		},
	)

	txHashes := make([]bitcoin.Hash, len(items))
	for i, item := range items {
		txHashes[i] = item.txHash
	}

	return txHashes, nil
}

// getConfirmedScriptHistory returns a history of confirmed transactions for
// the given script (P2PKH, P2WPKH, P2SH, P2WSH, etc.). The returned list
// is sorted by the block height in the ascending order, i.e. the latest
// transaction is at the end of the list. The resulting list does not contain
// unconfirmed transactions living in the mempool at the moment of request.
func (c *Connection) getConfirmedScriptHistory(
	script []byte,
) ([]*scriptHistoryItem, error) {
	scriptHash := sha256.Sum256(script)
	reversedScriptHash := bitcoin.Reverse(scriptHash[:])
	reversedScriptHashString := hex.EncodeToString(reversedScriptHash)

	items, err := requestWithRetry(
		c,
		func(
			ctx context.Context,
			client *electrum.Client,
		) ([]*electrum.GetMempoolResult, error) {
			return client.GetHistory(ctx, reversedScriptHashString)
		},
		"GetHistory",
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to get history for script [0x%x]: [%v]",
			script,
			err,
		)
	}

	// According to https://electrumx.readthedocs.io/en/latest/protocol-methods.html#blockchain-scripthash-get-history
	// unconfirmed items living in the mempool are appended at the end of the
	// returned list and their height value is either -1 or 0. That means
	// we need to take all items with height >0 to obtain a confirmed txs
	// history.
	confirmedItems := make([]*scriptHistoryItem, 0)
	for _, item := range items {
		if item.Height > 0 {
			txHash, err := bitcoin.NewHashFromString(
				item.Hash,
				bitcoin.ReversedByteOrder,
			)
			if err != nil {
				return nil, fmt.Errorf(
					"cannot parse hash [%s]: [%v]",
					item.Hash,
					err,
				)
			}

			confirmedItems = append(
				confirmedItems, &scriptHistoryItem{
					txHash:      txHash,
					blockHeight: item.Height,
				},
			)
		}
	}

	// The list returned from client.GetHistory is sorted by the block height
	// in the ascending order though we are sorting it again just in case
	// (e.g. API contract changes).
	sort.SliceStable(
		confirmedItems,
		func(i, j int) bool {
			return confirmedItems[i].blockHeight < confirmedItems[j].blockHeight
		},
	)

	return confirmedItems, nil
}

type scriptHistoryItem struct {
	txHash      bitcoin.Hash
	blockHeight int32
}

func isTxNotFoundErr(err error) bool {
	txNotFoundErrs := []string{
		"no such mempool or blockchain transaction",
		"missing transaction",
		"transaction not found",
	}

	errStr := strings.ToLower(err.Error())

	for _, txNotFoundErr := range txNotFoundErrs {
		if strings.Contains(errStr, txNotFoundErr) {
			return true
		}
	}

	return false
}

func (c *Connection) electrumConnect() error {
	var client *electrum.Client
	var err error

	logger.Debug("establishing connection to electrum server...")
	client, err = connectWithRetry(
		c,
		func(ctx context.Context) (*electrum.Client, error) {
			return electrum.NewClient(ctx, c.config.URL, nil)
		},
	)

	if err == nil {
		c.client = client
	}

	return err
}

func (c *Connection) keepAlive() {
	ticker := time.NewTicker(c.config.KeepAliveInterval)

	for {
		select {
		case <-ticker.C:
			_, err := requestWithRetry(
				c,
				func(ctx context.Context, client *electrum.Client) (interface{}, error) {
					return nil, client.Ping(ctx)
				},
				"Ping",
			)
			if err != nil {
				logger.Errorf(
					"failed to ping the electrum server; "+
						"please verify health of the electrum server: [%v]",
					err,
				)
			} else {
				// Adjust ticker starting at the time of the latest successful ping.
				ticker = time.NewTicker(c.config.KeepAliveInterval)
			}
		case <-c.parentCtx.Done():
			ticker.Stop()
			c.client.Shutdown()
			return
		}
	}
}

func (c *Connection) verifyServer() error {
	type Server struct {
		version  string
		protocol string
	}

	server, err := requestWithRetry(
		c,
		func(ctx context.Context, client *electrum.Client) (*Server, error) {
			serverVersion, protocolVersion, err := client.ServerVersion(ctx)
			if err != nil {
				return nil, err
			}
			return &Server{serverVersion, protocolVersion}, nil
		},
		"ServerVersion",
	)
	if err != nil {
		return fmt.Errorf("failed to get server version: [%w]", err)
	}

	logger.Infof(
		"connected to electrum server [version: [%s], protocol: [%s]]",
		server.version,
		server.protocol,
	)

	// Log a warning if connected to a server running an unsupported protocol version.
	if !slices.Contains(supportedProtocolVersions, server.protocol) {
		logger.Warnf(
			"electrum server [%s] runs an unsupported protocol version: [%s]; expected one of: [%s]",
			c.config.URL,
			server.protocol,
			strings.Join(supportedProtocolVersions, ","),
		)
	}

	return nil
}

func connectWithRetry(
	c *Connection,
	newClientFn func(ctx context.Context) (*electrum.Client, error),
) (*electrum.Client, error) {
	var result *electrum.Client
	err := wrappers.DoWithDefaultRetry(
		c.parentCtx,
		c.config.ConnectRetryTimeout,
		func(ctx context.Context) error {
			connectCtx, connectCancel := context.WithTimeout(
				ctx,
				c.config.ConnectTimeout,
			)
			defer connectCancel()

			client, err := newClientFn(connectCtx)
			if err == nil {
				result = client
			}

			return err
		},
	)

	return result, err
}

func requestWithRetry[K interface{}](
	c *Connection,
	requestFn func(ctx context.Context, client *electrum.Client) (K, error),
	requestName string,
) (K, error) {
	startTime := time.Now()
	logger.Debugf("starting [%s] request to Electrum server", requestName)

	var result K

	err := wrappers.DoWithDefaultRetry(
		c.parentCtx,
		c.config.RequestRetryTimeout,
		func(ctx context.Context) error {
			if err := c.reconnectIfShutdown(); err != nil {
				return err
			}

			requestCtx, requestCancel := context.WithTimeout(ctx, c.config.RequestTimeout)
			defer requestCancel()

			c.clientMutex.Lock()
			r, err := requestFn(requestCtx, c.client)
			c.clientMutex.Unlock()

			if err != nil {
				return fmt.Errorf("request failed: [%w]", err)
			}

			result = r
			return nil
		})

	solveRequestOutcome := func(err error) string {
		if err != nil {
			return fmt.Sprintf("error: [%v]", err)
		}
		return "success"
	}

	logger.Debugf("[%s] request to Electrum server completed with [%s] after [%s]",
		requestName,
		solveRequestOutcome(err),
		time.Since(startTime),
	)

	return result, err
}

func (c *Connection) reconnectIfShutdown() error {
	c.clientMutex.Lock()
	defer c.clientMutex.Unlock()

	isClientShutdown := c.client.IsShutdown()
	if isClientShutdown {
		logger.Warn("connection to electrum server is down; reconnecting...")
		err := c.electrumConnect()
		if err != nil {
			return fmt.Errorf("failed to reconnect to electrum server: [%w]", err)
		}
		logger.Info("reconnected to electrum server")
	}

	return nil
}
