package types

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestNewERC20MintCall(t *testing.T) {
	from := common.HexToAddress("0xbb9A13411Ee01F163B5FeB8cd8Ec4cE5bcc9500c")
	to := common.HexToAddress("0x9609B36D7feF4D8A641170aC403F16609ffc0EDC")
	recipient := common.HexToAddress("0x337B4f715fdb2Bcf7872c657505d3A347F8c3419")
	amount := big.NewInt(1000)

	call, err := NewERC20MintCall(from.Bytes(), to.Bytes(), recipient.Bytes(), amount)
	require.NoError(t, err)

	// Expected data is a concatenation of the following:
	// 1. Function selector for mint(address,uint256).
	// 2. Recipient address left-padded to 32 bytes.
	// 3. Amount as hex, left-padded to 32 bytes.
	expectedData := "40c10f19" +
		"000000000000000000000000337b4f715fdb2bcf7872c657505d3a347f8c3419" +
		"00000000000000000000000000000000000000000000000000000000000003e8"
	expectedDataBytes, err := hex.DecodeString(expectedData)
	require.NoError(t, err)

	require.Equal(t, from, call.From())
	require.Equal(t, &to, call.To())
	require.Equal(t, expectedDataBytes, call.Data())
}
