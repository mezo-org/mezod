package faucet

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"math/big"
	"net/http"
	"os"
	"strings"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	btctoken "github.com/mezo-org/mezod/infrastructure/terraform/mezo-staging/gcf/faucet/bindings"
)

// transfer amount - 0.01 BTC
var transferAmount = big.NewInt(10000000000000000)

// BTC token precompile address
const token = "0x7b7c000000000000000000000000000000000000"

func init() {
	functions.HTTP("Distribute", distribute)
}

func distribute(w http.ResponseWriter, r *http.Request) {
	// set CORS headers for the preflight request
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Max-Age", "3600")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// get env args
	rpcURL := os.Getenv("RPCURL")
	secret := os.Getenv("SECRET")

	// make sure we have a valid address
	address := strings.TrimPrefix(r.URL.Path, "/")
	if !common.IsHexAddress(address) {
		http.Error(w, "invalid address", http.StatusBadRequest)
		return
	}
	to := common.HexToAddress(address)

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

	// token address
	tokenAddress := common.HexToAddress(token)

	// contract
	btc, err := btctoken.NewBtctoken(tokenAddress, client)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// chainID
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// nonce
	nonce, err := client.PendingNonceAt(context.Background(), from)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// transfer btc
	transfer, err := btc.BtctokenTransactor.Transfer(&bind.TransactOpts{
		From:  from,
		Nonce: big.NewInt(int64(nonce)),
		Signer: func(a common.Address, tx *types.Transaction) (*types.Transaction, error) {
			return types.SignTx(tx, types.LatestSignerForChainID(chainID), privkey)
		},
		Value: big.NewInt(0),
	}, to, transferAmount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// set CORS header
	w.Header().Set("Access-Control-Allow-Origin", "*")
	// set Content-Type header
	w.Header().Set("Content-Type", "application/json")
	// write response
	err = json.NewEncoder(w).Encode(transfer)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
