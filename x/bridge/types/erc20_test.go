package types

import (
	"testing"

	evmtypes "github.com/mezo-org/mezod/x/evm/types"
	"github.com/stretchr/testify/require"
)

func TestERC20TokenMapping(t *testing.T) {
	testSourceERC20Token := evmtypes.HexAddressToBytes("0xac7f043Cf1BF10143926CC0035dBc46999512732")
	testMezoERC20Token := evmtypes.HexAddressToBytes("0x54F551EF4a7754d810461F554Edf9495063bc6e5")

	mapping := NewERC20TokenMapping(testSourceERC20Token, testMezoERC20Token)

	require.Equal(t, testSourceERC20Token, mapping.SourceTokenBytes())
	require.Equal(t, testMezoERC20Token, mapping.MezoTokenBytes())
}
