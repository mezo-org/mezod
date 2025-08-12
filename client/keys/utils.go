// Copyright 2022 Evmos Foundation
// This file is part of the Evmos Network packages.
//
// Evmos is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Evmos packages are distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Evmos packages. If not, see https://github.com/evmos/evmos/blob/main/LICENSE
package keys

import (
	"bufio"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"io"

	"sigs.k8s.io/yaml"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/crypto"
	cryptokeyring "github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/mezo-org/mezod/crypto/ethsecp256k1"
	"github.com/spf13/cobra"
)

// available output formats.
const (
	OutputFormatText = "text"
	OutputFormatJSON = "json"
)

type bechKeyOutFn func(k *cryptokeyring.Record) (keys.KeyOutput, error)

func printKeyringRecord(w io.Writer, k *cryptokeyring.Record, bechKeyOut bechKeyOutFn, output string) error {
	ko, err := bechKeyOut(k)
	if err != nil {
		return err
	}

	switch output {
	case OutputFormatText:
		if err := printTextRecords(w, []keys.KeyOutput{ko}); err != nil {
			return err
		}

	case OutputFormatJSON:
		out, err := json.Marshal(ko)
		if err != nil {
			return err
		}

		if _, err := fmt.Fprintln(w, string(out)); err != nil {
			return err
		}
	}

	return nil
}

func printTextRecords(w io.Writer, kos []keys.KeyOutput) error {
	out, err := yaml.Marshal(&kos)
	if err != nil {
		return err
	}

	if _, err := fmt.Fprintln(w, string(out)); err != nil {
		return err
	}

	return nil
}

func ExtractPrivateKey(cmd *cobra.Command, keyName string) (*ecdsa.PrivateKey, error) {
	clientCtx, err := client.GetClientTxContext(cmd)
	if err != nil {
		return nil, fmt.Errorf("unable to get client transaction context: %v", err)
	}

	decryptPassword := ""
	inBuf := bufio.NewReader(cmd.InOrStdin())
	if clientCtx.Keyring.Backend() == cryptokeyring.BackendFile {
		decryptPassword, err = input.GetPassword("Exporting private key. \nEnter key password:", inBuf)
		if err != nil {
			return nil, fmt.Errorf("failed to get password: %v", err)
		}
	}

	armor, err := clientCtx.Keyring.ExportPrivKeyArmor(keyName, decryptPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to export private key: %v", err)
	}

	privKey, algo, err := crypto.UnarmorDecryptPrivKey(armor, decryptPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt private key: %v", err)
	}
	if algo != ethsecp256k1.KeyType {
		return nil, fmt.Errorf("invalid key algorithm, got %s, expected %s", algo, ethsecp256k1.KeyType)
	}

	ethPrivKey, ok := privKey.(*ethsecp256k1.PrivKey)
	if !ok {
		return nil, fmt.Errorf("invalid private key type %T, expected %T", privKey, &ethsecp256k1.PrivKey{})
	}

	return ethPrivKey.ToECDSA()
}
