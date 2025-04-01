package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/mezo-org/mezod/crypto/hd"
	mezotypes "github.com/mezo-org/mezod/types"
)

type LocalKey struct {
	Secret string `json:"secret"`
}

func importMnemonic(mnemonic string) string {
	privKey, _ := hd.EthSecp256k1.Derive()(mnemonic, keyring.DefaultBIP39Passphrase, mezotypes.BIP44HDPath)

	_, _, address := getAccountFromRaw(privKey)
	fmt.Printf("loaded menomic key: %v\n", address)

	return hex.EncodeToString(privKey)
}

func importKey(path string) string {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("couldn't load local key file: %v", err)
	}

	var lkey LocalKey
	err = json.Unmarshal(buf, &lkey)
	if err != nil {
		log.Fatalf("invalid local key file: %v", err)
	}

	privKey, _ := hd.EthSecp256k1.Derive()(lkey.Secret, keyring.DefaultBIP39Passphrase, mezotypes.BIP44HDPath)

	_, _, address := getAccountFromRaw(privKey)
	fmt.Printf("loaded local key: %v\n", address)

	return hex.EncodeToString(privKey)
}
