package types

import (
	"math/big"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/stretchr/testify/require"
)

func TestAccessListTracerExclusions(t *testing.T) {
	osakaTime := sdkmath.NewInt(100)
	ethCfg := DefaultChainConfig()
	ethCfg.OsakaTime = &osakaTime
	cfg := ethCfg.EthereumConfig(big.NewInt(1))
	from := common.HexToAddress("0x01")
	to := common.HexToAddress("0x02")
	p256Precompile := common.HexToAddress("0x0100")
	customPrecompile := common.HexToAddress(BTCTokenPrecompileAddress)
	msg := core.Message{From: from, To: &to}

	exclusionsBeforeOsaka := accessListTracerExclusions(
		msg,
		cfg,
		1,
		99,
		[]common.Address{customPrecompile},
	)
	exclusionsAtOsaka := accessListTracerExclusions(
		msg,
		cfg,
		1,
		100,
		[]common.Address{customPrecompile},
	)

	require.Contains(t, exclusionsAtOsaka, from)
	require.Contains(t, exclusionsAtOsaka, to)
	require.Contains(t, exclusionsAtOsaka, customPrecompile)
	require.NotContains(t, exclusionsBeforeOsaka, p256Precompile)
	require.Contains(t, exclusionsAtOsaka, p256Precompile)
}
