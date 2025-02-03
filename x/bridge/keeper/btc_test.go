package keeper

import (
	"testing"

	evmtypes "github.com/mezo-org/mezod/x/evm/types"
	"github.com/stretchr/testify/require"
)

func TestGetSourceBTCToken(t *testing.T) {
	ctx, k := mockContext()
	expectedToken := testSourceBTCToken

	k.setSourceBTCToken(ctx, evmtypes.HexAddressToBytes(expectedToken))
	actualToken := k.GetSourceBTCToken(ctx)
	require.Equal(t, expectedToken, evmtypes.BytesToHexAddress(actualToken))
}
