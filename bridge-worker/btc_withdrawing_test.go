package bridgeworker

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestComputeAssetsUnlockedHash(t *testing.T) {
	unlockSeq := big.NewInt(1)
	token := common.HexToAddress("0x5FbDB2315678afecb367f032d93F642f64180aa3")
	amount := big.NewInt(1000)
	chain := uint8(1)
	recipient, err := hex.DecodeString(
		"1976a914f4eedc8f40d4b8e30771f792b065ebec0abaddef88ac",
	)
	if err != nil {
		t.Fatal(err)
	}

	expectedHash, err := hex.DecodeString(
		"aa4b0f8491b6dacd340e19be39225c5ac39da0e242b23c84b4ff15e7cf8c948d",
	)
	if err != nil {
		t.Fatal(err)
	}

	hash, err := computeAssetsUnlockedHash(
		unlockSeq,
		recipient,
		token,
		amount,
		chain,
	)
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, expectedHash, hash[:])
}
