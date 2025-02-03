package types_test

import (
	"errors"
	"math/big"
	"testing"

	"github.com/cosmos/cosmos-sdk/client"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"

	"github.com/ethereum/go-ethereum/common"

	proto "github.com/cosmos/gogoproto/proto"

	"github.com/mezo-org/mezod/app"
	"github.com/mezo-org/mezod/encoding"
	utiltx "github.com/mezo-org/mezod/testutil/tx"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"

	"github.com/stretchr/testify/require"
)

func TestEvmDataEncoding(t *testing.T) {
	ret := []byte{0x5, 0x8}

	data := &evmtypes.MsgEthereumTxResponse{
		Hash: common.BytesToHash([]byte("hash")).String(),
		Logs: []*evmtypes.Log{{
			Data:        []byte{1, 2, 3, 4},
			BlockNumber: 17,
		}},
		Ret: ret,
	}

	anyData := codectypes.UnsafePackAny(data)
	txData := &sdk.TxMsgData{
		MsgResponses: []*codectypes.Any{anyData},
	}

	txDataBz, err := proto.Marshal(txData)
	require.NoError(t, err)

	res, err := evmtypes.DecodeTxResponse(txDataBz)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, data.Logs, res.Logs)
	require.Equal(t, ret, res.Ret)
}

func TestUnwrapEthererumMsg(t *testing.T) {
	_, err := evmtypes.UnwrapEthereumMsg(nil, common.Hash{})
	require.NotNil(t, err)

	encodingConfig := encoding.MakeConfig(app.ModuleBasics)
	clientCtx := client.Context{}.WithTxConfig(encodingConfig.TxConfig)
	builder, _ := clientCtx.TxConfig.NewTxBuilder().(authtx.ExtensionOptionsTxBuilder)

	tx := builder.GetTx().(sdk.Tx)
	_, err = evmtypes.UnwrapEthereumMsg(&tx, common.Hash{})
	require.NotNil(t, err)

	evmTxParams := &evmtypes.EvmTxArgs{
		ChainID:  big.NewInt(1),
		Nonce:    0,
		To:       &common.Address{},
		Amount:   big.NewInt(0),
		GasLimit: 0,
		GasPrice: big.NewInt(0),
		Input:    []byte{},
	}

	msg := evmtypes.NewTx(evmTxParams)
	err = builder.SetMsgs(msg)
	require.Nil(t, err)

	tx = builder.GetTx().(sdk.Tx)
	unwrappedMsg, err := evmtypes.UnwrapEthereumMsg(&tx, msg.AsTransaction().Hash())
	require.Nil(t, err)
	require.Equal(t, unwrappedMsg, msg)
}

func TestBinSearch(t *testing.T) {
	successExecutable := func(gas uint64) (bool, *evmtypes.MsgEthereumTxResponse, error) {
		target := uint64(21000)
		return gas < target, nil, nil
	}
	failedExecutable := func(_ uint64) (bool, *evmtypes.MsgEthereumTxResponse, error) {
		return true, nil, errors.New("contract failed")
	}

	gas, err := evmtypes.BinSearch(20000, 21001, successExecutable)
	require.NoError(t, err)
	require.Equal(t, gas, uint64(21000))

	gas, err = evmtypes.BinSearch(20000, 21001, failedExecutable)
	require.Error(t, err)
	require.Equal(t, gas, uint64(0))
}

func TestTransactionLogsEncodeDecode(t *testing.T) {
	addr := utiltx.GenerateAddress().String()

	txLogs := evmtypes.TransactionLogs{
		Hash: common.BytesToHash([]byte("tx_hash")).String(),
		Logs: []*evmtypes.Log{
			{
				Address:     addr,
				Topics:      []string{common.BytesToHash([]byte("topic")).String()},
				Data:        []byte("data"),
				BlockNumber: 1,
				TxHash:      common.BytesToHash([]byte("tx_hash")).String(),
				TxIndex:     1,
				BlockHash:   common.BytesToHash([]byte("block_hash")).String(),
				Index:       1,
				Removed:     false,
			},
		},
	}

	txLogsEncoded, encodeErr := evmtypes.EncodeTransactionLogs(&txLogs)
	require.Nil(t, encodeErr)

	txLogsEncodedDecoded, decodeErr := evmtypes.DecodeTransactionLogs(txLogsEncoded)
	require.Nil(t, decodeErr)
	require.Equal(t, txLogs, txLogsEncodedDecoded)
}

func TestIsHexAddress(t *testing.T) {
	tests := map[string]struct {
		address string
		valid   bool
	}{
		"checksum": {
			address: "0x7471bc0f225D4e00586057412cc491b6c755DeD1",
			valid:   true,
		},
		"lowercase": {
			address: "0x7471bc0f225d4e00586057412cc491b6c755ded1",
			valid:   true,
		},
		"uppercase": {
			address: "0x7471BC0F225D4E00586057412CC491B6C755DED1",
			valid:   true,
		},
		"unprefixed": {
			address: "7471bc0f225D4e00586057412cc491b6c755DeD1",
			valid:   true,
		},
		"too short - uneven": {
			address: "0x7471bc0f225D4e00586057412cc491b6c755DeD",
			valid:   false,
		},
		"too short - even": {
			address: "0x7471bc0f225D4e00586057412cc491b6c755De",
			valid:   false,
		},
		"too long - uneven": {
			address: "0x7471bc0f225D4e00586057412cc491b6c755DeD1D",
			valid:   false,
		},
		"too long - even": {
			address: "0x7471bc0f225D4e00586057412cc491b6c755DeD1D1",
			valid:   false,
		},
		"zero": {
			address: "0000000000000000000000000000000000000000",
			valid:   true,
		},
		"empty": {
			address: "",
			valid:   false,
		},
		"non-hex": {
			address: "0xzzz",
			valid:   false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, test.valid, evmtypes.IsHexAddress(test.address))
		})
	}
}

func TestIsZeroHexAddress(t *testing.T) {
	tests := map[string]struct {
		address string
		valid   bool
	}{
		"non-zero": {
			address: "0x7471bc0f225D4e00586057412cc491b6c755DeD1",
			valid:   false,
		},
		"zero": {
			address: "0x0000000000000000000000000000000000000000",
			valid:   true,
		},
		"zero unprefixed": {
			address: "0000000000000000000000000000000000000000",
			valid:   true,
		},
		"empty": {
			address: "",
			valid:   true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, test.valid, evmtypes.IsZeroHexAddress(test.address))
		})
	}
}

func TestHexAddressToBytes(t *testing.T) {
	expectedBytes := []byte{0x74, 0x71, 0xbc, 0x0f, 0x22, 0x5d, 0x4e, 0x00, 0x58, 0x60, 0x57, 0x41, 0x2c, 0xc4, 0x91, 0xb6, 0xc7, 0x55, 0xde, 0xd1}
	zeroBytes := make([]byte, 20)

	tests := map[string]struct {
		address string
		bytes   []byte
	}{
		"checksum": {
			address: "0x7471bc0f225D4e00586057412cc491b6c755DeD1",
			bytes:   expectedBytes,
		},
		"lowercase": {
			address: "0x7471bc0f225d4e00586057412cc491b6c755ded1",
			bytes:   expectedBytes,
		},
		"uppercase": {
			address: "0x7471BC0F225D4E00586057412CC491B6C755DED1",
			bytes:   expectedBytes,
		},
		"unprefixed": {
			address: "7471bc0f225D4e00586057412cc491b6c755DeD1",
			bytes:   expectedBytes,
		},
		"too short - uneven": {
			address: "0x7471bc0f225D4e00586057412cc491b6c755DeD",
			bytes:   []byte{0x7, 0x47, 0x1b, 0xc0, 0xf2, 0x25, 0xd4, 0xe0, 0x5, 0x86, 0x5, 0x74, 0x12, 0xcc, 0x49, 0x1b, 0x6c, 0x75, 0x5d, 0xed},
		},
		"too short - even": {
			address: "0x7471bc0f225D4e00586057412cc491b6c755De",
			bytes:   []byte{0x0, 0x74, 0x71, 0xbc, 0xf, 0x22, 0x5d, 0x4e, 0x0, 0x58, 0x60, 0x57, 0x41, 0x2c, 0xc4, 0x91, 0xb6, 0xc7, 0x55, 0xde},
		},
		"too long - uneven": {
			address: "0x7471bc0f225D4e00586057412cc491b6c755DeD1D",
			bytes:   []byte{0x47, 0x1b, 0xc0, 0xf2, 0x25, 0xd4, 0xe0, 0x5, 0x86, 0x5, 0x74, 0x12, 0xcc, 0x49, 0x1b, 0x6c, 0x75, 0x5d, 0xed, 0x1d},
		},
		"too long - even": {
			address: "0x7471bc0f225D4e00586057412cc491b6c755DeD1D1",
			bytes:   []byte{0x71, 0xbc, 0xf, 0x22, 0x5d, 0x4e, 0x0, 0x58, 0x60, 0x57, 0x41, 0x2c, 0xc4, 0x91, 0xb6, 0xc7, 0x55, 0xde, 0xd1, 0xd1},
		},
		"zero": {
			address: "0000000000000000000000000000000000000000",
			bytes:   zeroBytes,
		},
		"empty": {
			address: "",
			bytes:   zeroBytes,
		},
		"non-hex": {
			address: "0xzzz",
			bytes:   zeroBytes,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, test.bytes, evmtypes.HexAddressToBytes(test.address))
		})
	}
}

func TestBytesToHexAddress(t *testing.T) {
	tests := map[string]struct {
		bytes   []byte
		address string
	}{
		"proper": {
			bytes:   []byte{0x74, 0x71, 0xbc, 0x0f, 0x22, 0x5d, 0x4e, 0x00, 0x58, 0x60, 0x57, 0x41, 0x2c, 0xc4, 0x91, 0xb6, 0xc7, 0x55, 0xde, 0xd1},
			address: "0x7471bc0f225D4e00586057412cc491b6c755DeD1",
		},
		"too short": {
			bytes:   []byte{0x74, 0x71, 0xbc, 0x0f, 0x22, 0x5d, 0x4e, 0x00, 0x58, 0x60, 0x57, 0x41, 0x2c, 0xc4, 0x91, 0xb6, 0xc7, 0x55, 0xde},
			address: "0x007471BC0f225d4E00586057412CC491b6c755DE",
		},
		"too long": {
			bytes:   []byte{0x74, 0x71, 0xbc, 0x0f, 0x22, 0x5d, 0x4e, 0x00, 0x58, 0x60, 0x57, 0x41, 0x2c, 0xc4, 0x91, 0xb6, 0xc7, 0x55, 0xde, 0xd1, 0xd1},
			address: "0x71bc0f225d4e00586057412Cc491B6c755DED1D1",
		},
		"zero": {
			bytes:   make([]byte, 20),
			address: "0x0000000000000000000000000000000000000000",
		},
		"empty": {
			bytes:   []byte{},
			address: "0x0000000000000000000000000000000000000000",
		},
		"nil": {
			bytes:   nil,
			address: "0x0000000000000000000000000000000000000000",
		},
	}

	for name, test := range tests {
		t.Run(
			name, func(t *testing.T) {
				require.Equal(
					t,
					test.address,
					evmtypes.BytesToHexAddress(test.bytes),
				)
			},
		)
	}
}
