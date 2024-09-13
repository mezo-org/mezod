package faucet

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"net/http"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"golang.org/x/crypto/sha3"
)

// TODO: move all of these values to a terrafrom conf
const transferAmount = 10000000000000000
const rpcURL = "http://127.0.0.1:8545"
const secret = ""
const token = "0x7b7c000000000000000000000000000000000000"

func init() {
	functions.HTTP("Distribute", distribute)
}

func distribute(w http.ResponseWriter, r *http.Request) {
	// Check a `to` url query is provided
	// note: URL.Query returns a map[string][]string
	// url: /index.html?to=1&to=2&to=3
	// map: { to: [ "1", "2", "3" ] }
	query := r.URL.Query()
	to, ok := query["to"]
	if !ok {
		http.Error(w, "missing `to` param", http.StatusBadRequest)
		return
	}

	// connect client to rpc endpoint
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		http.Error(w, "unable to dial `RPC_URL`", http.StatusInternalServerError)
		return
	}

	// import `from` using privkey
	privkey, err := crypto.HexToECDSA(secret)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	publicKey := privkey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		http.Error(w, "invalid pubkey", http.StatusInternalServerError)
		return
	}
	// from
	from := crypto.PubkeyToAddress(*publicKeyECDSA)
	fmt.Printf("from: %s\n", from.String())
	// token address
	tokenAddress := common.HexToAddress(token)
	fmt.Printf("token: %s\n", tokenAddress.String())

	// value (this is the transaction value, not the token transfer amount)
	value := big.NewInt(0) // in abtc (0 BTC)

	// amount (this is the amount of tokens to transfer in abtc)
	amount := big.NewInt(transferAmount)
	paddedAmount := common.LeftPadBytes(amount.Bytes(), 32)

	transferFnSignature := []byte("transfer(address,uint256)")
	hash := sha3.NewLegacyKeccak256()
	hash.Write(transferFnSignature)
	methodID := hash.Sum(nil)[:4]

	for _, v := range to {
		// Make sure we have a valid address
		if !common.IsHexAddress(v) {
			http.Error(w, "invalid address", http.StatusBadRequest)
			return
		}

		// nonce
		nonce, err := client.PendingNonceAt(context.Background(), from)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		gasPrice, err := client.SuggestGasPrice(context.Background())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		toAddress := common.HexToAddress(v)
		fmt.Printf("to: %s\n", toAddress.String())
		paddedAddress := common.LeftPadBytes(toAddress.Bytes(), 32)

		var data []byte
		data = append(data, methodID...)
		data = append(data, paddedAddress...)
		data = append(data, paddedAmount...)

		gasLimit, err := client.EstimateGas(context.Background(), ethereum.CallMsg{
			To:   &tokenAddress,
			Data: data,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		tx := types.NewTransaction(nonce, tokenAddress, value, gasLimit, gasPrice, data)
		chainID, err := client.NetworkID(context.Background())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		signedTx, err := types.SignTx(tx, types.LatestSignerForChainID(chainID), privkey)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = client.SendTransaction(context.Background(), signedTx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Printf("tx sent: %s", signedTx.Hash().Hex())
	}
}
