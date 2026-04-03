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

func TestNewTripartyCallbackCall(t *testing.T) {
	from := common.HexToAddress("0xbb9A13411Ee01F163B5FeB8cd8Ec4cE5bcc9500c")
	to := common.HexToAddress("0x9609B36D7feF4D8A641170aC403F16609ffc0EDC")
	recipient := common.HexToAddress("0x337B4f715fdb2Bcf7872c657505d3A347F8c3419")
	requestID := big.NewInt(1)
	amount := big.NewInt(1000)
	callbackData := []byte{0xab, 0xcd}

	call, err := NewTripartyCallbackCall(
		from.Bytes(),
		to.Bytes(),
		requestID,
		recipient.Bytes(),
		amount,
		callbackData,
	)
	require.NoError(t, err)

	// Expected data is a concatenation of the following:
	// 1. Function selector for onTripartyBridgeCompleted(uint256,address,uint256,bytes).
	// 2. requestId as uint256.
	// 3. Recipient address left-padded to 32 bytes.
	// 4. Amount as uint256.
	// 5. Offset to callbackData (0x80).
	// 6. callbackData length.
	// 7. callbackData right-padded to 32 bytes.
	expectedData := "2c144c16" +
		"0000000000000000000000000000000000000000000000000000000000000001" +
		"000000000000000000000000337b4f715fdb2bcf7872c657505d3a347f8c3419" +
		"00000000000000000000000000000000000000000000000000000000000003e8" +
		"0000000000000000000000000000000000000000000000000000000000000080" +
		"0000000000000000000000000000000000000000000000000000000000000002" +
		"abcd000000000000000000000000000000000000000000000000000000000000"
	expectedDataBytes, err := hex.DecodeString(expectedData)
	require.NoError(t, err)

	require.Equal(t, from, call.From())
	require.Equal(t, &to, call.To())
	require.Equal(t, TripartyCallbackGasLimit, call.GasLimit())
	require.Equal(t, expectedDataBytes, call.Data())
}
