package types

import (
	"math/big"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/stretchr/testify/require"
)

func TestTracerPrecompilesUseTimestampForks(t *testing.T) {
	osakaTime := sdkmath.NewInt(100)
	ethCfg := DefaultChainConfig()
	ethCfg.OsakaTime = &osakaTime
	cfg := ethCfg.EthereumConfig(big.NewInt(1))
	p256Precompile := common.HexToAddress("0x0100")

	precompilesBeforeOsaka := vm.ActivePrecompiles(cfg.Rules(big.NewInt(1), true, 99))
	precompilesAtOsaka := vm.ActivePrecompiles(cfg.Rules(big.NewInt(1), true, 100))

	require.NotContains(t, precompilesBeforeOsaka, p256Precompile)
	require.Contains(t, precompilesAtOsaka, p256Precompile)
}
