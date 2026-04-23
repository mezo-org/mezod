package keeper

import (
	"errors"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/stretchr/testify/require"

	"github.com/mezo-org/mezod/x/evm/types"
)

// TestSimulateV1_CoveredInGrpcQueryTest exists as a pointer for future
// readers: the public SimulateV1 gRPC handler is exercised end-to-end in
// grpc_query_test.go against a fully-wired KeeperTestSuite app. The
// private simulateV1 driver is not covered with dedicated tests here —
// repeating the same scenarios against a hand-built keeper would reimplement
// the suite setup for no additional signal. Helper coverage (sanitize +
// make-header) lives below and runs without a keeper instance.
func TestSimulateV1_CoveredInGrpcQueryTest(_ *testing.T) {}

func hbig(n int64) *hexutil.Big { return (*hexutil.Big)(big.NewInt(n)) }

func hu64(v uint64) *hexutil.Uint64 { h := hexutil.Uint64(v); return &h }

// postMergeConfig returns the minimal chain config the simulate header
// builder needs to exercise fork-gated branches. London is active so
// CalcBaseFee works; merge is active so Difficulty is zeroed.
func postMergeConfig() *params.ChainConfig {
	mergeHeight := big.NewInt(0)
	return &params.ChainConfig{
		ChainID:             big.NewInt(31611),
		HomesteadBlock:      big.NewInt(0),
		EIP150Block:         big.NewInt(0),
		EIP155Block:         big.NewInt(0),
		EIP158Block:         big.NewInt(0),
		ByzantiumBlock:      big.NewInt(0),
		ConstantinopleBlock: big.NewInt(0),
		PetersburgBlock:     big.NewInt(0),
		IstanbulBlock:       big.NewInt(0),
		MuirGlacierBlock:    big.NewInt(0),
		BerlinBlock:         big.NewInt(0),
		LondonBlock:         big.NewInt(0),
		MergeNetsplitBlock:  mergeHeight,
	}
}

// The go-ethereum TestSimulateSanitizeBlockOrder fixture: base at
// 10/50, caller skips to block 13 with Time=80. Expected fill:
// 11 @ 62, 12 @ 74, 13 @ 80. Verified against simulate_test.go:46-51
// upstream.
func TestSanitizeSimChain_GapFill(t *testing.T) {
	base := &ethtypes.Header{Number: big.NewInt(10), Time: 50}
	blocks := []types.SimBlock{
		{BlockOverrides: &types.SimBlockOverrides{Number: hbig(13), Time: hu64(80)}},
	}
	out, err := sanitizeSimChain(base, blocks)
	require.NoError(t, err)
	require.Len(t, out, 3)

	require.Equal(t, int64(11), out[0].BlockOverrides.Number.ToInt().Int64())
	require.Equal(t, uint64(62), uint64(*out[0].BlockOverrides.Time))
	require.Equal(t, int64(12), out[1].BlockOverrides.Number.ToInt().Int64())
	require.Equal(t, uint64(74), uint64(*out[1].BlockOverrides.Time))
	require.Equal(t, int64(13), out[2].BlockOverrides.Number.ToInt().Int64())
	require.Equal(t, uint64(80), uint64(*out[2].BlockOverrides.Time))
}

func TestSanitizeSimChain_DefaultsFromParent(t *testing.T) {
	base := &ethtypes.Header{Number: big.NewInt(100), Time: 1000}
	blocks := []types.SimBlock{{}}
	out, err := sanitizeSimChain(base, blocks)
	require.NoError(t, err)
	require.Len(t, out, 1)
	require.Equal(t, int64(101), out[0].BlockOverrides.Number.ToInt().Int64())
	require.Equal(t, uint64(1012), uint64(*out[0].BlockOverrides.Time))
}

func TestSanitizeSimChain_NonMonotonicNumber(t *testing.T) {
	base := &ethtypes.Header{Number: big.NewInt(10), Time: 60}
	blocks := []types.SimBlock{
		{BlockOverrides: &types.SimBlockOverrides{Number: hbig(10)}},
	}
	_, err := sanitizeSimChain(base, blocks)
	requireSimError(t, err, types.SimErrCodeBlockNumberInvalid)
}

func TestSanitizeSimChain_NonMonotonicTimestamp(t *testing.T) {
	base := &ethtypes.Header{Number: big.NewInt(10), Time: 60}
	blocks := []types.SimBlock{
		{BlockOverrides: &types.SimBlockOverrides{Number: hbig(11), Time: hu64(60)}},
	}
	_, err := sanitizeSimChain(base, blocks)
	requireSimError(t, err, types.SimErrCodeBlockTimestampInvalid)
}

func TestSanitizeSimChain_SpanBoundGuard(t *testing.T) {
	base := &ethtypes.Header{Number: big.NewInt(10), Time: 60}
	// Jumping 10M blocks forward must fail BEFORE the gap-fill
	// allocates millions of headers.
	blocks := []types.SimBlock{
		{BlockOverrides: &types.SimBlockOverrides{Number: hbig(10_000_010)}},
	}
	_, err := sanitizeSimChain(base, blocks)
	requireSimError(t, err, types.SimErrCodeClientLimitExceeded)
}

func TestSanitizeSimChain_MaxBlocksBoundary(t *testing.T) {
	base := &ethtypes.Header{Number: big.NewInt(10), Time: 60}
	// 256 blocks forward is on the allowed edge.
	blocks := []types.SimBlock{
		{BlockOverrides: &types.SimBlockOverrides{Number: hbig(10 + int64(maxSimulateBlocks))}},
	}
	out, err := sanitizeSimChain(base, blocks)
	require.NoError(t, err)
	require.Len(t, out, maxSimulateBlocks)

	// 257 over is rejected.
	blocks = []types.SimBlock{
		{BlockOverrides: &types.SimBlockOverrides{Number: hbig(10 + int64(maxSimulateBlocks) + 1)}},
	}
	_, err = sanitizeSimChain(base, blocks)
	requireSimError(t, err, types.SimErrCodeClientLimitExceeded)
}

// Sanitize failures must surface as *types.SimError with the
// spec-reserved code — callers branch on the code via errors.As.
func TestSanitizeSimChain_ErrorsAsSimError(t *testing.T) {
	base := &ethtypes.Header{Number: big.NewInt(10), Time: 60}
	_, err := sanitizeSimChain(base, []types.SimBlock{
		{BlockOverrides: &types.SimBlockOverrides{Number: hbig(5)}},
	})
	requireSimError(t, err, types.SimErrCodeBlockNumberInvalid)
}

// requireSimError asserts that err unwraps to a *types.SimError
// carrying the expected spec-reserved JSON-RPC code.
func requireSimError(t *testing.T, err error, wantCode int) {
	t.Helper()
	require.Error(t, err)
	var simErr *types.SimError
	require.True(t, errors.As(err, &simErr), "expected *types.SimError, got %T: %v", err, err)
	require.Equal(t, wantCode, simErr.ErrorCode())
}

func TestMakeSimHeader_NilOverridesDefaultsFromParent(t *testing.T) {
	cfg := postMergeConfig()
	parent := &ethtypes.Header{
		Number:     big.NewInt(100),
		Time:       1000,
		Difficulty: big.NewInt(0),
		GasLimit:   20_000_000,
		Coinbase:   common.HexToAddress("0xabc"),
		BaseFee:    big.NewInt(1_000_000_000),
	}
	rules := cfg.Rules(big.NewInt(101), true, 1012)
	h := makeSimHeader(parent, nil, rules, cfg, false)
	require.Equal(t, int64(101), h.Number.Int64())
	require.Equal(t, uint64(1012), h.Time)
	require.Equal(t, parent.Coinbase, h.Coinbase)
	require.Equal(t, parent.GasLimit, h.GasLimit)
	// Parent-hash wired explicitly.
	require.Equal(t, parent.Hash(), h.ParentHash)
	// Scaffolding roots set.
	require.Equal(t, ethtypes.EmptyUncleHash, h.UncleHash)
	require.Equal(t, ethtypes.EmptyReceiptsHash, h.ReceiptHash)
	require.Equal(t, ethtypes.EmptyTxsHash, h.TxHash)
}

func TestMakeSimHeader_PostMergeDifficultyZero(t *testing.T) {
	cfg := postMergeConfig()
	parent := &ethtypes.Header{Number: big.NewInt(1), Time: 10, Difficulty: big.NewInt(42)}
	rules := cfg.Rules(big.NewInt(2), true, 22)
	h := makeSimHeader(parent, nil, rules, cfg, false)
	require.Equal(t, int64(0), h.Difficulty.Int64())
}

func TestMakeSimHeader_BaseFeeOverrideWins(t *testing.T) {
	cfg := postMergeConfig()
	parent := &ethtypes.Header{Number: big.NewInt(1), Time: 10, Difficulty: big.NewInt(0), GasLimit: 30_000_000, BaseFee: big.NewInt(1_000_000_000)}
	rules := cfg.Rules(big.NewInt(2), true, 22)

	explicit := hbig(7)
	h := makeSimHeader(parent, &types.SimBlockOverrides{BaseFeePerGas: explicit}, rules, cfg, true)
	require.Equal(t, int64(7), h.BaseFee.Int64())
}

func TestMakeSimHeader_ValidationDerivesBaseFee(t *testing.T) {
	cfg := postMergeConfig()
	parent := &ethtypes.Header{
		Number:     big.NewInt(1),
		Time:       10,
		Difficulty: big.NewInt(0),
		GasLimit:   30_000_000,
		GasUsed:    15_000_000,
		BaseFee:    big.NewInt(1_000_000_000),
	}
	rules := cfg.Rules(big.NewInt(2), true, 22)

	// validation=true without an override → CalcBaseFee on parent.
	h := makeSimHeader(parent, nil, rules, cfg, true)
	require.NotNil(t, h.BaseFee)
	require.Equal(t, int64(1_000_000_000), h.BaseFee.Int64(), "gasUsed == target keeps base fee flat")

	// validation=false without an override → zero base fee.
	h2 := makeSimHeader(parent, nil, rules, cfg, false)
	require.Equal(t, int64(0), h2.BaseFee.Int64())
}

func TestMakeSimHeader_OverrideFields(t *testing.T) {
	cfg := postMergeConfig()
	parent := &ethtypes.Header{
		Number:     big.NewInt(1),
		Time:       10,
		Difficulty: big.NewInt(0),
		GasLimit:   30_000_000,
	}
	rules := cfg.Rules(big.NewInt(2), true, 22)

	feeRecipient := common.HexToAddress("0xdeadbeef")
	prevRandao := common.HexToHash("0xcafe")
	o := &types.SimBlockOverrides{
		Number:       hbig(999),
		Time:         hu64(42_000),
		GasLimit:     hu64(1_234_567),
		FeeRecipient: &feeRecipient,
		PrevRandao:   &prevRandao,
	}
	h := makeSimHeader(parent, o, rules, cfg, false)
	require.Equal(t, int64(999), h.Number.Int64())
	require.Equal(t, uint64(42_000), h.Time)
	require.Equal(t, uint64(1_234_567), h.GasLimit)
	require.Equal(t, feeRecipient, h.Coinbase)
	require.Equal(t, prevRandao, h.MixDigest)
}
