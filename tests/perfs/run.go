package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"math/rand/v2"
	"slices"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/mezo-org/mezod/tests/perfs/token"
	"golang.org/x/exp/maps"
)

var (
	ongoingTotalTxs atomic.Int64
)

func runNative(
	cnt int,
	runTime time.Duration,
	destination common.Address,
) {
	client := getClient()
	accounts := loadAccounts()

	if len(accounts.Accounts) < cnt {
		log.Fatalf("only %v accounts, requested %v clients", len(accounts.Accounts), cnt)
	}

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		log.Fatalf("Failed to get chain ID: %v", err)
	}

	accountsAddress := maps.Keys(accounts.Accounts)
	slices.Sort(accountsAddress)

	go aggregateBlockDataParallel()
	time.Sleep(10 * time.Second) // wait to get a few blocks first to see baseline

	var stop atomic.Bool
	stop.Store(false)

	startedAt := time.Now()
	for i := 0; i < cnt; i++ {
		address := accountsAddress[i]
		time.Sleep(50 * time.Millisecond)
		go func() {
			runNativeOne(
				client,
				accounts.Accounts[address],
				destination,
				chainID,
				&stop,
			)
		}()
	}

	// just run it for a while
	time.Sleep(runTime)
	timeTaken := time.Since(startedAt)
	fmt.Printf("total transaction sent: %v in: %v\n", ongoingTotalTxs.Load(), timeTaken)
	stop.Store(true)
	log.Printf("stopping wallets")
	time.Sleep(1 * time.Minute)
}

func runNativeOne(
	client *ethclient.Client,
	privKeyRaw string,
	toAddress common.Address,
	chainID *big.Int,
	stop *atomic.Bool,
) {
	// Load your private key
	privateKey, err := crypto.HexToECDSA(privKeyRaw)
	if err != nil {
		log.Fatalf("Failed to load private key: %v", err)
	}

	// Get the public key and address from the private key
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("Error casting public key to ECDSA")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	// Get the nonce for the sender's address
	// just once before, we can increase it next..
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatalf("Failed to get nonce: %v", err)
	}

	// standard gas limit for ETH transfers
	gasLimit := uint64(21000)

	value := big.NewInt(1)

	for !stop.Load() {
		time.Sleep(rateLimit)
		// ask every time, price might change over time
		gasPrice, err := client.SuggestGasPrice(context.Background())
		if err != nil {
			log.Printf("Failed to suggest gas price: %v", err)
			continue
		}

		// Create the transaction
		var txData = &types.DynamicFeeTx{
			ChainID:   chainID, // Mainnet
			Nonce:     nonce,
			GasTipCap: gasPrice,
			GasFeeCap: gasPrice.Mul(gasPrice, big.NewInt(2)),
			Gas:       gasLimit,
			To:        &toAddress,
			Value:     value,
			Data:      nil,
		}

		// Create transaction and sign it
		tx := types.NewTx(txData)

		signedTx, err := types.SignTx(tx, types.NewLondonSigner(chainID), privateKey)
		if err != nil {
			log.Fatalf("Failed to sign transaction: %v", err)
		}

		// Send the transaction
		err = client.SendTransaction(context.Background(), signedTx)
		if err != nil {
			log.Printf("Failed to send transaction: %v", err)
			// most errors are about nonce
			nonce, err = client.PendingNonceAt(context.Background(), fromAddress)
			if err != nil {
				log.Printf("Failed to get nonce: %v", err)
			}
			continue
		}

		// Print the transaction hash
		fmt.Printf("transaction sent: %s\n", signedTx.Hash().Hex())

		// Wait for transaction to be mined
		if waitForReceipt {
			_, err = waitForTransaction(client, signedTx.Hash())
			if err != nil {
				log.Fatalf("failed waiting for transaction: %v", err)
			}
		}

		// time.Sleep(3000 * time.Millisecond)
		nonce += 1
		ongoingTotalTxs.Add(1)
	}
}

func runERC20(
	cnt int,
	runTime time.Duration,
	destination common.Address,
	tokenAddress common.Address,
) {
	client := getClient()
	accounts := loadAccounts()

	if len(accounts.Accounts) < cnt {
		log.Fatalf("only %v accounts, requested %v clients", len(accounts.Accounts), cnt)
	}

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		log.Fatalf("Failed to get chain ID: %v", err)
	}

	accountsAddress := maps.Keys(accounts.Accounts)
	slices.Sort(accountsAddress)

	var stop atomic.Bool
	stop.Store(false)
	go aggregateBlockDataParallel()
	time.Sleep(10 * time.Second) // wait to get a few blocks first to see baseline

	startedAt := time.Now()
	for i := 0; i < cnt; i++ {
		address := accountsAddress[i]
		time.Sleep(30 * time.Millisecond)
		go func() {
			runERC20One(
				client,
				accounts.Accounts[address],
				destination,
				chainID,
				tokenAddress,
				i,
				&stop,
			)
		}()
	}

	// just run it for a while
	time.Sleep(runTime)
	timeTaken := time.Since(startedAt)
	fmt.Printf("total transaction sent: %v in: %v\n", ongoingTotalTxs.Load(), timeTaken)
	stop.Store(true)
	log.Printf("stopping wallets")
	time.Sleep(1 * time.Minute) // cooldown

}

func randRange(min, max int) int {
	return rand.IntN(max-min) + min
}

func runERC20One(
	client *ethclient.Client,
	privKeyRaw string,
	toAddress common.Address,
	chainID *big.Int,
	tokenAddress common.Address,
	index int,
	stop *atomic.Bool,
) {
	// Load your private key
	privateKey, err := crypto.HexToECDSA(privKeyRaw)
	if err != nil {
		log.Fatalf("Failed to load private key: %v", err)
	}

	// Get the public key and address from the private key
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("Error casting public key to ECDSA")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	// Get the nonce for the sender's address
	// just once before, we can increase it next..
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatalf("Failed to get nonce: %v", err)
	}

	// gasLimit := uint64(65000) // more or less an ERC20 gas requirement
	// gasLimit := uint64(22000)
	// gasLimit := uint64(50000)
	gasLimit := uint64(25000)

	// Initialize the ERC20 contract
	token, err := token.NewTokenTransactor(tokenAddress, client)
	if err != nil {
		log.Fatalf("Failed to instantiate ERC20 contract: %v", err)
	}

	value := big.NewInt(1)

	for !stop.Load() {
		if rateLimit == 0 {
			time.Sleep(
				time.Duration(randRange(0, 1000)) * time.Millisecond,
			)
		} else {
			time.Sleep(rateLimit)
		}

		// Create an auth transactor
		auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
		if err != nil {
			log.Fatalf("Failed to create transactor: %v", err)
		}

		// ask every time, price might change over time
		gasPrice, err := client.SuggestGasPrice(context.Background())
		if err != nil {
			log.Printf("Failed to suggest gas price: %v", err)
			continue
		}

		auth.Nonce = big.NewInt(int64(nonce))
		auth.GasPrice = gasPrice
		auth.GasLimit = gasLimit

		// Execute the transfer
		tx, err := token.Transfer(auth, toAddress, value)
		if err != nil {
			log.Printf("Failed to execute transfer: index(%v), %v", index, err)
			// most errors are about nonce
			nonce, err = client.PendingNonceAt(context.Background(), fromAddress)
			if err != nil {
				log.Fatalf("Failed to get nonce: %v", err)
			}

			continue
		}

		// Print the transaction hash
		fmt.Printf("transaction sent: %s\n", tx.Hash().Hex())

		// Wait for transaction to be mined
		if waitForReceipt {
			receipt, err := waitForTransaction(client, tx.Hash())
			if err != nil {
				log.Fatalf("failed waiting for transaction: %v", err)
			}
			_ = receipt
		}

		// time.Sleep(3000 * time.Millisecond)
		nonce += 1
		ongoingTotalTxs.Add(1)
	}
}
