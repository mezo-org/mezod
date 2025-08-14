package sidecar

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"cosmossdk.io/log"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/keep-network/keep-common/pkg/chain/ethereum/ethutil"
	"github.com/mezo-org/mezod/ethereum/bindings/portal"
	bridgetypes "github.com/mezo-org/mezod/x/bridge/types"
)

// NOTE: these two function are set here at the top
// level because I see them like consts more or less
// happy to put it somewhere else if people do not
// like it.
var (
	// a 10% buffer on top the gasLimit
	adjustGasLimit = func(gasLimit uint64) uint64 {
		return gasLimit + gasLimit/10
	}

	// a 20% buffer on top of the gasPrice
	adjustGasPrice = func(gasPrice *big.Int) *big.Int {
		gasPriceCpy := new(big.Int).Set(gasPrice)
		return gasPrice.Add(gasPrice, gasPriceCpy.Div(gasPrice, big.NewInt(5)))
	}

	defaultTransactionReceiptTicker = 1 * time.Second
)

type MezoBridgeTransactor interface {
	AttestBridgeOut(auth *bind.TransactOpts, entry *bridgetypes.AssetsUnlockedEvent) (*types.Transaction, error)
}

type Chain interface {
	ChainID() *big.Int
	Client() ethutil.EthereumClient
}

type IndividualAttestationTransactionExecutor struct {
	logger           log.Logger
	privateKey       *ecdsa.PrivateKey
	address          common.Address
	bridgeAddress    common.Address
	chain            Chain
	bridgeTransactor MezoBridgeTransactor
	auth             *bind.TransactOpts
}

func NewIndividualAttestationTransactionExecutor(
	logger log.Logger,
	privateKey *ecdsa.PrivateKey,
	chain Chain,
	bridgeAddress common.Address,
	bridgeTransactor MezoBridgeTransactor,
) (*IndividualAttestationTransactionExecutor, error) {
	// get the public key and address from the private key
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("error casting public key to ECDSA")
	}
	address := crypto.PubkeyToAddress(*publicKeyECDSA)

	// prepare our auth which will be used on every request
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chain.ChainID())
	if err != nil {
		return nil, fmt.Errorf("couldn't create ethereum keyed transactor: %w", err)
	}

	return &IndividualAttestationTransactionExecutor{
		logger:           logger,
		privateKey:       privateKey,
		address:          address,
		bridgeAddress:    bridgeAddress,
		chain:            chain,
		bridgeTransactor: bridgeTransactor,
		auth:             auth,
	}, nil
}

func (te *IndividualAttestationTransactionExecutor) Send(
	attestation *bridgetypes.AssetsUnlockedEvent,
) error {
	if err := te.updateAuth(); err != nil {
		return err
	}

	if err := te.estimateGasLimit(attestation); err != nil {
		return err
	}

	// execute the transaction
	tx, err := te.bridgeTransactor.AttestBridgeOut(te.auth, attestation)
	if err != nil {
		return fmt.Errorf("couldn't send ethereum transaction: %w", err)
	}

	receipt, err := te.waitForTransactionReceipt(tx.Hash())
	if err != nil {
		return err
	}

	if receipt.Status != 1 {
		return fmt.Errorf("attestBridgeOut call failed: %v", tx.Hash().Hex())
	}

	te.logger.Info("attestation: %v execute successfully at %v", attestation.String(), tx.Hash().Hex())

	return nil
}

func (te *IndividualAttestationTransactionExecutor) updateAuth() error {
	gasPrice, err := te.suggestedGasPrice()
	if err != nil {
		return fmt.Errorf("failed to suggest gas price: %v", err)
	}

	gasPrice = adjustGasPrice(gasPrice)

	nonce, err := te.pendingNonce()
	if err != nil {
		return fmt.Errorf("failed to get nonce: %v", err)
	}

	// when everything have succeeded we update the auth
	te.auth.Nonce = nonce
	te.auth.GasPrice = gasPrice

	return nil
}

func (te *IndividualAttestationTransactionExecutor) pendingNonce() (*big.Int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	nonce, err := te.chain.Client().PendingNonceAt(ctx, te.address)
	if err != nil {
		return nil, err
	}

	// #nosec G115
	return big.NewInt(int64(nonce)), nil
}

func (te *IndividualAttestationTransactionExecutor) suggestedGasPrice() (*big.Int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return te.chain.Client().SuggestGasPrice(ctx)
}

func (te *IndividualAttestationTransactionExecutor) estimateGasLimit(
	attestation *bridgetypes.AssetsUnlockedEvent,
) error {
	client := te.chain.Client()

	contractABI, err := abi.JSON(strings.NewReader(portal.MezoBridgeABI))
	if err != nil {
		return fmt.Errorf("couldn't create MezoBridge ABI: %w", err)
	}

	// TODO: remove when the following is implemented
	_ = contractABI
	data := []byte{}
	_ = attestation

	// TODO: uncomment when the ABI is implemented
	// data, err := contractABI.Pack("attestBridgeOut", attestation)
	// if err != nil {
	// 	return fmt.Errorf("couldn't pack attestBridgeOut arguments: %w", err)
	// }

	msg := ethereum.CallMsg{
		From: te.address,
		To:   &te.bridgeAddress,
		Data: data,
	}

	gasLimit, err := client.EstimateGas(context.Background(), msg)
	if err != nil {
		return fmt.Errorf("couldn't estimate gasLimit for attestBridgeOut: %w", err)
	}

	// finally set the adjusted gasLimit
	te.auth.GasLimit = adjustGasLimit(gasLimit)

	return nil
}

func (te *IndividualAttestationTransactionExecutor) getReceipt(txHash common.Hash) *types.Receipt {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	receipt, err := te.chain.Client().TransactionReceipt(ctx, txHash)
	if err != nil {
		// If error is "not found", transaction is still pending
		// no need to log an error though
		if err != ethereum.NotFound {
			te.logger.Error("couldn't get transaction receipt: %v", err)
		}

		return nil
	}

	return receipt
}

func (te *IndividualAttestationTransactionExecutor) waitForTransactionReceipt(txHash common.Hash) (*types.Receipt, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	ticker := time.NewTicker(defaultTransactionReceiptTicker)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if tx := te.getReceipt(txHash); tx != nil {
				return tx, nil
			}
		case <-ctx.Done():

			return nil, fmt.Errorf("couldn't find the receipt for transaction: %v", txHash.Hex())
		}
	}
}
