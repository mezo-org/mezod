package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"slices"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	token "github.com/mezo-org/mezod/tests/performance/bindings"
	"golang.org/x/exp/maps"
)

const (
	// initialTransfer = "10000000000"
	initialTransfer     = "100000000000000000"
	tokenInitialBalance = "1"
)

var (
	btcTokenAddress = common.HexToAddress("0x7b7C000000000000000000000000000000000000")
)

func newAccount() (string, common.Address) {
	// Generate a new private key
	privateKey, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
	if err != nil {
		log.Fatalf("Failed to generate private key: %v", err)
	}

	// Get the private key bytes
	privateKeyBytes := crypto.FromECDSA(privateKey)

	// Convert private key to hex string
	privateKeyHex := hex.EncodeToString(privateKeyBytes)
	fmt.Printf("private key: 0x%s\n", privateKeyHex)

	// Derive the public key and address (for verification)
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("Failed to convert public key to ECDSA")
	}

	// Get the public address
	address := crypto.PubkeyToAddress(*publicKeyECDSA)
	fmt.Printf("ethereum address: 0x%x\n", address)

	return privateKeyHex, address
}

func transfer(
	client *ethclient.Client,
	chainID *big.Int,
	fromPrivateKey *ecdsa.PrivateKey,
	fromAddress, toAddress common.Address,
	amount *big.Int,
	tokenAddress common.Address,
) bool {
	// Create an auth transactor
	auth, err := bind.NewKeyedTransactorWithChainID(fromPrivateKey, chainID)
	if err != nil {
		log.Fatalf("Failed to create transactor: %v", err)
	}

	// Get nonce
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatalf("Failed to get nonce: %v", err)
	}
	auth.Nonce = big.NewInt(int64(nonce))

	// Get gas price
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatalf("Failed to get gas price: %v", err)
	}
	auth.GasPrice = gasPrice

	// Set gas limit (or let the function estimate it)
	auth.GasLimit = uint64(300000)

	token, err := token.NewTokenTransactor(tokenAddress, client)

	// Initialize the ERC20 contract
	if err != nil {
		log.Fatalf("Failed to instantiate ERC20 contract: %v", err)
	}

	// Execute the transfer
	tx, err := token.Transfer(auth, toAddress, amount)
	if err != nil {
		log.Fatalf("Failed to execute transfer: %v", err)
	}

	fmt.Printf("transaction sent: %s\n", tx.Hash().Hex())

	// Wait for transaction to be mined
	if waitForReceipt {

		receipt, err := waitForTransaction(client, tx.Hash())
		if err != nil {
			log.Fatalf("failed waiting for transaction: %v", err)
		}

		if receipt.Status == 1 {
			fmt.Println("transaction was successful")
			return true
		}

		fmt.Println("transaction failed!")

		return false
	}

	return true
}

func topupERC20Precompile(cnt int) {
	topupERC20(cnt, btcTokenAddress)
}

func topupERC20(cnt int, tokenAddress common.Address) {
	client := getClient()
	fromPrivateKey, _, fromAddress := getAccount(mainAccount)

	chainID, err := client.ChainID(context.Background())
	if err != nil {
		log.Fatalf("Failed to get chain ID: %v", err)
	}

	amount := new(big.Int)
	amount.SetString(initialTransfer, 10)

	accounts := loadAccounts()
	if len(accounts.Accounts) < cnt {
		log.Fatalf("only %v accounts, requested %v clients", len(accounts.Accounts), cnt)
	}

	accountsAddress := maps.Keys(accounts.Accounts)
	slices.Sort(accountsAddress)

	for i := 0; i < int(cnt); i++ {
		toAddress := common.HexToAddress(accountsAddress[i])
		fmt.Printf("toping up account: %v\n", i)
		time.Sleep(100 * time.Millisecond)

		ok := transfer(client, chainID, fromPrivateKey, fromAddress, toAddress, amount, tokenAddress)
		if !ok {
			fmt.Printf("couldn't tranfer funds")
		}
	}

}

func generateAndTransferERC20Precompile(cnt uint) {
	client := getClient()
	fromPrivateKey, _, fromAddress := getAccount(mainAccount)

	chainID, err := client.ChainID(context.Background())
	if err != nil {
		log.Fatalf("Failed to get chain ID: %v", err)
	}

	amount := new(big.Int)
	amount.SetString(initialTransfer, 10)

	accounts := loadAccounts()

	for i := len(accounts.Accounts); i < int(cnt); i++ {
		fmt.Printf("generating account: %v\n", i)
		time.Sleep(100 * time.Millisecond)

		newPriv, newAddress := newAccount()

		ok := transfer(client, chainID, fromPrivateKey, fromAddress, newAddress, amount, btcTokenAddress)
		if !ok {
			fmt.Printf("couldn't tranfer funds")
		}

		// All went well, we can save the updated accounts
		accounts.Accounts[newAddress.String()] = newPriv
		saveAccounts(accounts)
	}
}

func deployToken() {
	client := getClient()
	fromPrivateKey, _, fromAddress := getAccount(mainAccount)

	chainID, err := client.ChainID(context.Background())
	if err != nil {
		log.Fatalf("Failed to get chain ID: %v", err)
	}

	// Create an auth transactor
	auth, err := bind.NewKeyedTransactorWithChainID(fromPrivateKey, chainID)
	if err != nil {
		log.Fatalf("Failed to create transactor: %v", err)
	}

	// Get nonce
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatalf("Failed to get nonce: %v", err)
	}
	auth.Nonce = big.NewInt(int64(nonce))

	// Get gas price
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatalf("Failed to get gas price: %v", err)
	}
	auth.GasPrice = gasPrice

	// Set gas limit (or let the function estimate it)
	// auth.GasLimit = uint64(300000)

	supply := big.NewInt(0).Mul(
		big.NewInt(1000000000000000000), // 1 TOKEN
		big.NewInt(10000),               // x10000
	)

	address, tx, _, err := token.DeployToken(auth, client, "test_token", "TEST", supply, fromAddress)
	if err != nil {
		log.Fatalf("unable to deploy token: %v", err)
	}

	fmt.Printf("token deployed at tx hash: %v with address: %v\n", tx.Hash(), address)

	if waitForReceipt {

		receipt, err := waitForTransaction(client, tx.Hash())
		if err != nil {
			log.Fatalf("failed waiting for transaction: %v", err)
		}

		if receipt.Status == 1 {
			fmt.Println("transaction was successful")
			return
		}

		fmt.Printf("transaction failed: %v\n", receipt.Status)
	}
}
