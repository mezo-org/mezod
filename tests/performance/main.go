package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

var (
	mainAccount string = "" // used to bootstrap deposits and such
)

func main() {
	runCLI()
}

func getClient() *ethclient.Client {
	client, err := ethclient.Dial(rpcAddress)
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}

	return client
}

func getAccount(privKey string) (*ecdsa.PrivateKey, *ecdsa.PublicKey, common.Address) {
	privateKey, err := crypto.HexToECDSA(privKey)
	if err != nil {
		log.Fatalf("Failed to load private key from string: %v", err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("Error casting public key to ECDSA")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	return privateKey, publicKeyECDSA, fromAddress
}

func getAccountFromRaw(privKey []byte) (*ecdsa.PrivateKey, *ecdsa.PublicKey, common.Address) {
	privateKey, err := crypto.ToECDSA(privKey)
	if err != nil {
		log.Fatalf("Failed to load private key from raw: %v", err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("Error casting public key to ECDSA")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	return privateKey, publicKeyECDSA, fromAddress
}

// Helper function to wait for a transaction to be mined
func waitForTransaction(client *ethclient.Client, txHash common.Hash) (*types.Receipt, error) {
	ctx := context.Background()
	i := 0
	for {
		// initial wait of the time of a block,
		if i == 0 {
			time.Sleep(3500 * time.Millisecond)
		} else { // then poll often
			time.Sleep(500 * time.Millisecond)
		}
		receipt, err := client.TransactionReceipt(ctx, txHash)
		if err == nil {
			return receipt, nil
		}

		if err != ethereum.NotFound {
			return nil, err
		}

		fmt.Printf("transaction pending... waiting for confirmation %v\n", i)
		i++
	}
}
