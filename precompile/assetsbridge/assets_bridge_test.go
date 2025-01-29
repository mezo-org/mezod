package assetsbridge

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestPackEventsToInput(t *testing.T) {
	events := []AssetsLockedEvent{
		{
			SequenceNumber: big.NewInt(1),
			Recipient:      common.HexToAddress("0xa9bDddA0816EE183C51B28e0C33664B2DB069481"),
			Amount:         big.NewInt(1000000),
			Token:          common.HexToAddress("0x7d738d48b5c30f224aB86DaedE96CD95AB4854d9"),
		},
		{
			SequenceNumber: big.NewInt(2),
			Recipient:      common.HexToAddress("0xFc62EBEccDfccc1610844B4e0E3b85FE33829e84"),
			Amount:         big.NewInt(1000),
			Token:          common.HexToAddress("0x3127b53ab26ba3f70A25F5bC82BDAf5f4fd4f197"),
		},
		{
			SequenceNumber: big.NewInt(3),
			Recipient:      common.HexToAddress("0xA0209205d902c3a70EDba71C0e30420bd84AE1b8"),
			Amount:         big.NewInt(20000),
			Token:          common.HexToAddress("0x388Dc627C8876cC6a83209788e9CDE033060F03A"),
		},
	}

	expectedInput, err := hex.DecodeString("5b5eef90000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000030000000000000000000000000000000000000000000000000000000000000001000000000000000000000000a9bddda0816ee183c51b28e0c33664b2db06948100000000000000000000000000000000000000000000000000000000000f42400000000000000000000000000000000000000000000000000000000000000002000000000000000000000000fc62ebeccdfccc1610844b4e0e3b85fe33829e8400000000000000000000000000000000000000000000000000000000000003e80000000000000000000000000000000000000000000000000000000000000003000000000000000000000000a0209205d902c3a70edba71c0e30420bd84ae1b80000000000000000000000000000000000000000000000000000000000004e20")
	require.NoError(t, err)

	input, err := PackEventsToInput(events)
	require.NoError(t, err)
	require.Equal(t, expectedInput, input)
}
