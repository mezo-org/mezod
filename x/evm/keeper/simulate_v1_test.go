package keeper

import (
	"errors"
	"math"
	"math/big"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/require"

	"github.com/mezo-org/mezod/x/evm/statedb"
	"github.com/mezo-org/mezod/x/evm/types"
)

// TestSimulateV1_CoveredInGrpcQueryTest exists as a pointer for future
// readers: the public SimulateV1 gRPC handler is exercised end-to-end in
// grpc_query_test.go against a fully-wired KeeperTestSuite app. The
// private simulateV1 driver is not covered with dedicated tests here —
// repeating the same scenarios against a hand-built keeper would reimplement
// the suite setup for no additional signal. Helper coverage (sanitize-
// chain, resolve-call, make-header) lives below and runs without a
// keeper instance.
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

// Gap-fill shape mirrors the go-ethereum TestSimulateSanitizeBlockOrder
// fixture (simulate_test.go:46-51 upstream) retimed to mezo's ~3s block
// cadence: base at 10/50, caller skips to block 13 with Time=80.
// Expected fill: 11 @ 53, 12 @ 56, 13 @ 80.
func TestSanitizeSimChain_GapFill(t *testing.T) {
	base := &ethtypes.Header{Number: big.NewInt(10), Time: 50}
	blocks := []types.SimBlock{
		{BlockOverrides: &types.SimBlockOverrides{Number: hbig(13), Time: hu64(80)}},
	}
	out, err := sanitizeSimChain(base, blocks)
	require.NoError(t, err)
	require.Len(t, out, 3)

	require.Equal(t, int64(11), out[0].BlockOverrides.Number.ToInt().Int64())
	require.Equal(t, uint64(53), uint64(*out[0].BlockOverrides.Time))
	require.Equal(t, int64(12), out[1].BlockOverrides.Number.ToInt().Int64())
	require.Equal(t, uint64(56), uint64(*out[1].BlockOverrides.Time))
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
	require.Equal(t, uint64(1003), uint64(*out[0].BlockOverrides.Time))
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
		{BlockOverrides: &types.SimBlockOverrides{Number: hbig(10 + int64(types.MaxSimulateBlocks))}},
	}
	out, err := sanitizeSimChain(base, blocks)
	require.NoError(t, err)
	require.Len(t, out, types.MaxSimulateBlocks)

	// 257 over is rejected.
	blocks = []types.SimBlock{
		{BlockOverrides: &types.SimBlockOverrides{Number: hbig(10 + int64(types.MaxSimulateBlocks) + 1)}},
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
	rules := cfg.Rules(big.NewInt(101), true, 1003)
	h := makeSimHeader(parent, nil, rules, cfg, false)
	require.Equal(t, int64(101), h.Number.Int64())
	require.Equal(t, uint64(1003), h.Time)
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

// --- resolveSimCallNonce / resolveSimCallGas ----------------------------

// fakeNonceSource implements the nonceSource interface for
// resolveSimCallNonce unit tests so we avoid standing up a full
// StateDB / keeper.
type fakeNonceSource struct{ n uint64 }

func (f fakeNonceSource) GetNonce(common.Address) uint64 { return f.n }

func TestResolveSimCallNonce_DefaultsFromStateDB(t *testing.T) {
	from := common.HexToAddress("0xaaaa000000000000000000000000000000000001")
	src := fakeNonceSource{n: 7}
	args := &types.TransactionArgs{From: &from}

	nonce := resolveSimCallNonce(src, args)
	require.NotNil(t, nonce)
	require.Equal(t, uint64(7), uint64(*nonce))
}

func TestResolveSimCallNonce_PreservesExplicit(t *testing.T) {
	from := common.HexToAddress("0xaaaa000000000000000000000000000000000002")
	src := fakeNonceSource{n: 7}
	explicit := hexutil.Uint64(42)
	args := &types.TransactionArgs{From: &from, Nonce: &explicit}

	nonce := resolveSimCallNonce(src, args)
	require.Same(t, &explicit, nonce)
	require.Equal(t, uint64(42), uint64(*nonce))
}

func TestResolveSimCallGas_DefaultsToRemaining(t *testing.T) {
	from := common.HexToAddress("0xaaaa000000000000000000000000000000000003")
	args := &types.TransactionArgs{From: &from}
	header := &ethtypes.Header{GasLimit: 1_000_000}

	gas, simErr := resolveSimCallGas(args, header, 400_000, newSimGasBudget(0))
	require.Nil(t, simErr)
	require.NotNil(t, gas)
	require.Equal(t, uint64(600_000), uint64(*gas))
}

func TestResolveSimCallGas_BlockGasLimitReached(t *testing.T) {
	from := common.HexToAddress("0xaaaa000000000000000000000000000000000004")
	gas := hexutil.Uint64(700_000)
	args := &types.TransactionArgs{From: &from, Gas: &gas}
	header := &ethtypes.Header{GasLimit: 1_000_000}

	resolved, simErr := resolveSimCallGas(args, header, 400_000, newSimGasBudget(0))
	require.Nil(t, resolved)
	require.NotNil(t, simErr)
	require.Equal(t, types.SimErrCodeBlockGasLimitReached, simErr.ErrorCode())
	require.Contains(t, simErr.Message, "700000")
	require.Contains(t, simErr.Message, "600000")
}

// header.GasLimit=0 with implicit args.Gas → -38015.
func TestResolveSimCallGas_ZeroRemaining_DefaultEmitsSimError(t *testing.T) {
	from := common.HexToAddress("0xaaaa000000000000000000000000000000000005")
	args := &types.TransactionArgs{From: &from}
	header := &ethtypes.Header{GasLimit: 0}

	resolved, simErr := resolveSimCallGas(args, header, 0, newSimGasBudget(0))
	require.Nil(t, resolved)
	require.NotNil(t, simErr)
	require.Equal(t, types.SimErrCodeBlockGasLimitReached, simErr.ErrorCode())
}

// cumGasUsed == header.GasLimit with implicit args.Gas → -38015.
func TestResolveSimCallGas_BudgetFullyConsumed_DefaultEmitsSimError(t *testing.T) {
	from := common.HexToAddress("0xaaaa00000000000000000000000000000000000a")
	args := &types.TransactionArgs{From: &from}
	header := &ethtypes.Header{GasLimit: 1_000_000}

	resolved, simErr := resolveSimCallGas(args, header, 1_000_000, newSimGasBudget(0))
	require.Nil(t, resolved)
	require.NotNil(t, simErr)
	require.Equal(t, types.SimErrCodeBlockGasLimitReached, simErr.ErrorCode())
}

// header.GasLimit=0 with explicit args.Gas → -38015.
func TestResolveSimCallGas_ZeroGasLimit_ExplicitGasRejected(t *testing.T) {
	from := common.HexToAddress("0xaaaa000000000000000000000000000000000006")
	gas := hexutil.Uint64(21_000)
	args := &types.TransactionArgs{From: &from, Gas: &gas}
	header := &ethtypes.Header{GasLimit: 0}

	resolved, simErr := resolveSimCallGas(args, header, 0, newSimGasBudget(0))
	require.Nil(t, resolved)
	require.NotNil(t, simErr)
	require.Equal(t, types.SimErrCodeBlockGasLimitReached, simErr.ErrorCode())
}

// --- newSimGetHashFn -----------------------------------------------------

// canonicalHash is a fixed magic byte pattern a synthetic canonical
// resolver returns so tests can assert the closure delegated to it.
func canonicalHashForHeight(h uint64) common.Hash {
	var out common.Hash
	out[0] = 0xCA
	out[1] = byte(h)
	return out
}

// fakeCanonical returns a deterministic canonicalHashForHeight value
// for every height — an in-test stand-in for k.GetHashFn(ctx). Used
// across the newSimGetHashFn cases so the assertions can distinguish
// canonical hits from simulated-sibling hits and zero-hash misses.
func fakeCanonical(h uint64) common.Hash { return canonicalHashForHeight(h) }

func TestNewSimGetHashFn_HitBase(t *testing.T) {
	base := &ethtypes.Header{Number: big.NewInt(100)}
	fn := newSimGetHashFn(fakeCanonical, base, nil)
	require.Equal(t, canonicalHashForHeight(100), fn(100))
}

func TestNewSimGetHashFn_BelowBase_Canonical(t *testing.T) {
	base := &ethtypes.Header{Number: big.NewInt(100)}
	fn := newSimGetHashFn(fakeCanonical, base, nil)
	require.Equal(t, canonicalHashForHeight(42), fn(42))
	require.Equal(t, canonicalHashForHeight(99), fn(99))
}

func TestNewSimGetHashFn_AboveBase_ScansSim(t *testing.T) {
	base := &ethtypes.Header{Number: big.NewInt(100)}
	sim := []*ethtypes.Header{
		{Number: big.NewInt(101), Time: 10},
		{Number: big.NewInt(102), Time: 20},
		{Number: big.NewInt(103), Time: 30},
	}
	fn := newSimGetHashFn(fakeCanonical, base, sim)

	require.Equal(t, sim[0].Hash(), fn(101))
	require.Equal(t, sim[1].Hash(), fn(102))
	require.Equal(t, sim[2].Hash(), fn(103))
}

func TestNewSimGetHashFn_NotFound_Zero(t *testing.T) {
	base := &ethtypes.Header{Number: big.NewInt(100)}
	sim := []*ethtypes.Header{
		{Number: big.NewInt(101), Time: 10},
	}
	fn := newSimGetHashFn(fakeCanonical, base, sim)
	require.Equal(t, common.Hash{}, fn(102)) // beyond sim[] but future
	require.Equal(t, common.Hash{}, fn(200)) // far future
}

// TestNewSimGetHashFn_CanonicalUnforgeable verifies that the closure
// never serves an attacker-controlled hash for a canonical-range height
// even if, hypothetically, a sim[] slot carries a Number <= base.Number
// (bypassing sanitizeSimChain's monotonic check). Guards against future
// refactors breaking the invariant the multi-block security argument
// rests on.
func TestNewSimGetHashFn_CanonicalUnforgeable(t *testing.T) {
	base := &ethtypes.Header{Number: big.NewInt(100)}
	// Craft a sim[] header with Number == base.Number - 1 that sanitize
	// would never produce.
	evil := &ethtypes.Header{
		Number:     big.NewInt(99),
		Time:       1,
		Difficulty: big.NewInt(7),
		Extra:      []byte("malicious"),
	}
	fn := newSimGetHashFn(fakeCanonical, base, []*ethtypes.Header{evil})

	// Canonical range: must come from the canonical source, NOT from evil.
	got := fn(99)
	require.Equal(t, canonicalHashForHeight(99), got)
	require.NotEqual(t, evil.Hash(), got, "attacker-controlled sim[] header must not surface as canonical hash")
}

// Two cfg.Rules invocations at the same (height, time) must compare equal,
// even though ChainConfig.Rules allocates a fresh *big.Int ChainID on every
// call — the whole point of the helper is to look past that pointer.
func TestSameForks_DistinctChainIDPointers(t *testing.T) {
	cfg := postMergeConfig()
	a := cfg.Rules(big.NewInt(100), true, 1000)
	b := cfg.Rules(big.NewInt(100), true, 1000)
	require.NotSame(t, a.ChainID, b.ChainID, "fixture relies on Rules() returning fresh ChainID pointers")
	require.True(t, sameForks(a, b))
}

func TestSameForks_SameForkDifferentHeights(t *testing.T) {
	cfg := postMergeConfig()
	require.True(t, sameForks(
		cfg.Rules(big.NewInt(100), true, 1000),
		cfg.Rules(big.NewInt(999_999), true, 9_999_999),
	))
}

// The fork-span rejection emits a *SimError on the spec-reserved
// -38026 (ClientLimitExceeded) channel. Pins the wire shape so a future
// refactor doesn't accidentally regress to a bare fmt.Errorf and route
// to gRPC Internal.
func TestNewSimForkSpanUnsupported(t *testing.T) {
	e := types.NewSimForkSpanUnsupported()
	require.Equal(t, types.SimErrCodeClientLimitExceeded, e.ErrorCode())
	require.Contains(t, e.Error(), "fork boundary")
}

// Crossing a time-gated fork must flip sameForks to false — this is the
// boundary the simulate driver's sentinel relies on.
func TestSameForks_AcrossForkBoundary(t *testing.T) {
	cfg := postMergeConfig()
	shanghai := uint64(1_000_000)
	cancun := uint64(2_000_000)
	cfg.ShanghaiTime = &shanghai
	cfg.CancunTime = &cancun

	pre := cfg.Rules(big.NewInt(100), true, cancun-1)
	post := cfg.Rules(big.NewInt(100), true, cancun)
	require.True(t, post.IsCancun && !pre.IsCancun, "fixture must straddle Cancun activation")
	require.False(t, sameForks(pre, post))
}

// --- simGasBudget ---------------------------------------------------------

func TestSimGasBudget_ZeroIsUnlimited(t *testing.T) {
	b := newSimGasBudget(0)
	require.Equal(t, uint64(math.MaxUint64), b.remaining)
	require.Equal(t, uint64(1_000_000), b.clamp(1_000_000))
}

func TestSimGasBudget_ClampToRemaining(t *testing.T) {
	b := newSimGasBudget(500_000)
	require.Equal(t, uint64(100_000), b.clamp(100_000))
	require.Equal(t, uint64(500_000), b.clamp(700_000))
}

func TestSimGasBudget_ConsumeDeducts(t *testing.T) {
	b := newSimGasBudget(1_000_000)
	require.Nil(t, b.consume(300_000))
	require.Equal(t, uint64(700_000), b.remaining)
	require.Nil(t, b.consume(700_000))
	require.Equal(t, uint64(0), b.remaining)
}

func TestSimGasBudget_ConsumeOverflowsReturnsError(t *testing.T) {
	b := newSimGasBudget(100_000)
	err := b.consume(200_000)
	require.Error(t, err)
	require.Contains(t, err.Error(), "RPC gas cap exhausted")
	require.Contains(t, err.Error(), "200000")
	require.Contains(t, err.Error(), "100000")
	require.Equal(t, uint64(100_000), b.remaining)
}

// Budget remaining < args.Gas → effective gas clamps to budget remaining.
func TestResolveSimCallGas_ClampsByBudget(t *testing.T) {
	from := common.HexToAddress("0xaaaa00000000000000000000000000000000000b")
	gas := hexutil.Uint64(800_000)
	args := &types.TransactionArgs{From: &from, Gas: &gas}
	header := &ethtypes.Header{GasLimit: 1_000_000}
	budget := newSimGasBudget(300_000)

	resolved, simErr := resolveSimCallGas(args, header, 0, budget)
	require.Nil(t, simErr)
	require.NotNil(t, resolved)
	require.Equal(t, uint64(300_000), uint64(*resolved))
}

// Unlimited budget (gas-cap=0) does not shrink the resolved gas.
func TestResolveSimCallGas_UnlimitedBudgetIsNoop(t *testing.T) {
	from := common.HexToAddress("0xaaaa00000000000000000000000000000000000c")
	gas := hexutil.Uint64(500_000)
	args := &types.TransactionArgs{From: &from, Gas: &gas}
	header := &ethtypes.Header{GasLimit: 1_000_000}

	resolved, simErr := resolveSimCallGas(args, header, 0, newSimGasBudget(0))
	require.Nil(t, simErr)
	require.NotNil(t, resolved)
	require.Equal(t, uint64(500_000), uint64(*resolved))
}

// --- validateSimCall (validation=true gate) ------------------------------
//
// The gate runs five checks in geth's preCheck order: nonce -> init-code
// (CREATE only, Shanghai-gated) -> intrinsic gas -> fee-cap (when
// header.BaseFee is non-nil) -> balance. Boundary tests below pin each
// gate's pass/fail edge with a single-field perturbation against a
// shared baseline.
//
// Construction: the Keeper itself is empty (validateSimCall only calls
// k.GetEthIntrinsicGas, which reads ctx height/time + chain config and
// no keeper-stored state). The StateDB is backed by a MockKeeper so the
// nonce/balance reads come from the in-memory mock rather than a real
// IAVL store. The chain config has every fork active at height 0 so
// rules.IsShanghai is true; the contract-creation gate fires only when
// To==nil regardless of the height/time arithmetic.

// validateSimCallFixture builds the shared inputs for the gate tests.
// Each test perturbs one field at a time off this baseline so failures
// trace back to a single gate.
type validateSimCallFixture struct {
	k      *Keeper
	sdb    *statedb.StateDB
	ctx    sdk.Context
	header *ethtypes.Header
	rules  params.Rules
	cfg    *params.ChainConfig
	from   common.Address
	to     common.Address
}

// senderBaseline pre-funds the sender with a generous balance and seeds
// nonce 0 so each test can perturb either field cleanly.
const validateSimCallSenderBalance = "1000000000000000000" // 1e18 wei

func newValidateSimCallFixture(t *testing.T) *validateSimCallFixture {
	t.Helper()
	cfg := postMergeConfig()
	// Shanghai active at genesis so the init-code gate is reachable.
	zeroTime := uint64(0)
	cfg.ShanghaiTime = &zeroTime

	mk := statedb.NewMockKeeper()
	from := common.HexToAddress("0xaaaa000000000000000000000000000000000010")
	to := common.HexToAddress("0xbbbb000000000000000000000000000000000020")
	sdb := statedb.New(sdk.Context{}, mk, testTxConfig)
	bal, _ := uint256.FromDecimal(validateSimCallSenderBalance)
	sdb.AddBalance(from, bal, 0)
	require.NoError(t, sdb.Commit())

	// Re-instantiate so the committed balance is the persisted state
	// rather than the journal's pending change. validateSimCall calls
	// GetBalance / GetNonce which read through the StateDB's keeper.
	sdb = statedb.New(sdk.Context{}, mk, testTxConfig)

	rules := cfg.Rules(big.NewInt(1), true, 1)
	require.True(t, rules.IsShanghai, "fixture must activate Shanghai for the init-code gate")
	header := &ethtypes.Header{
		Number:   big.NewInt(1),
		Time:     1,
		GasLimit: 30_000_000,
		BaseFee:  big.NewInt(1_000_000_000),
	}
	return &validateSimCallFixture{
		k:      &Keeper{},
		sdb:    sdb,
		ctx:    sdk.Context{},
		header: header,
		rules:  rules,
		cfg:    cfg,
		from:   from,
		to:     to,
	}
}

// validateSimCallBaselineMsg returns a message that passes every gate:
// nonce 0 (matches the seeded state nonce), 21k gas (intrinsic for a
// pure transfer), gasFeeCap >= header.BaseFee, value < balance.
func (f *validateSimCallFixture) baselineMsg() *core.Message {
	return &core.Message{
		From:      f.from,
		To:        &f.to,
		Nonce:     0,
		GasLimit:  21_000,
		GasFeeCap: new(big.Int).Set(f.header.BaseFee),
		GasTipCap: big.NewInt(0),
		Value:     big.NewInt(0),
	}
}

// Happy path baseline. Subsequent tests perturb one field at a time, so
// any failure here would point at fixture drift rather than gate logic.
func TestValidateSimCall_AllGatesPass(t *testing.T) {
	f := newValidateSimCallFixture(t)
	require.Nil(t, f.k.validateSimCall(f.ctx, f.sdb, f.baselineMsg(), f.header, f.rules, f.cfg))
}

// --- per-gate boundaries -------------------------------------------------

func TestValidateSimCall_Nonce_EqualPasses(t *testing.T) {
	f := newValidateSimCallFixture(t)
	msg := f.baselineMsg()
	msg.Nonce = 0 // state nonce is 0
	require.Nil(t, f.k.validateSimCall(f.ctx, f.sdb, msg, f.header, f.rules, f.cfg))
}

func TestValidateSimCall_Nonce_OneLowFails_38010(t *testing.T) {
	f := newValidateSimCallFixture(t)
	// Bump state nonce so the gate has a non-zero target to compare against.
	f.sdb.SetNonce(f.from, 5)
	msg := f.baselineMsg()
	msg.Nonce = 4
	simErr := f.k.validateSimCall(f.ctx, f.sdb, msg, f.header, f.rules, f.cfg)
	require.NotNil(t, simErr)
	require.Equal(t, types.SimErrCodeNonceTooLow, simErr.ErrorCode())
}

func TestValidateSimCall_Nonce_OneHighFails_38011(t *testing.T) {
	f := newValidateSimCallFixture(t)
	msg := f.baselineMsg()
	msg.Nonce = 1 // state nonce is 0
	simErr := f.k.validateSimCall(f.ctx, f.sdb, msg, f.header, f.rules, f.cfg)
	require.NotNil(t, simErr)
	require.Equal(t, types.SimErrCodeNonceTooHigh, simErr.ErrorCode())
}

func TestValidateSimCall_IntrinsicGas_BoundaryEqual(t *testing.T) {
	f := newValidateSimCallFixture(t)
	msg := f.baselineMsg()
	msg.GasLimit = 21_000 // exactly intrinsic for a pure transfer
	require.Nil(t, f.k.validateSimCall(f.ctx, f.sdb, msg, f.header, f.rules, f.cfg))
}

func TestValidateSimCall_IntrinsicGas_OneBelowFails_38013(t *testing.T) {
	f := newValidateSimCallFixture(t)
	msg := f.baselineMsg()
	msg.GasLimit = 20_999
	simErr := f.k.validateSimCall(f.ctx, f.sdb, msg, f.header, f.rules, f.cfg)
	require.NotNil(t, simErr)
	require.Equal(t, types.SimErrCodeIntrinsicGas, simErr.ErrorCode())
}

func TestValidateSimCall_FeeCap_BoundaryEqual(t *testing.T) {
	f := newValidateSimCallFixture(t)
	msg := f.baselineMsg()
	msg.GasFeeCap = new(big.Int).Set(f.header.BaseFee) // exactly equal -> passes
	require.Nil(t, f.k.validateSimCall(f.ctx, f.sdb, msg, f.header, f.rules, f.cfg))
}

func TestValidateSimCall_FeeCap_OneBelowFails_32005(t *testing.T) {
	f := newValidateSimCallFixture(t)
	msg := f.baselineMsg()
	msg.GasFeeCap = new(big.Int).Sub(f.header.BaseFee, big.NewInt(1))
	simErr := f.k.validateSimCall(f.ctx, f.sdb, msg, f.header, f.rules, f.cfg)
	require.NotNil(t, simErr)
	require.Equal(t, types.SimErrCodeFeeCapTooLow, simErr.ErrorCode())
}

func TestValidateSimCall_Balance_BoundaryEqual(t *testing.T) {
	f := newValidateSimCallFixture(t)
	msg := f.baselineMsg()
	// balance == cost: gasLimit*gasFeeCap + value
	bal, _ := uint256.FromDecimal(validateSimCallSenderBalance)
	cost := new(big.Int).Mul(new(big.Int).SetUint64(msg.GasLimit), msg.GasFeeCap)
	cost.Add(cost, msg.Value)
	require.Equal(t, bal.ToBig().Cmp(cost), 1, "fixture must leave headroom; lower gasLimit*gasFeeCap if it doesn't")

	// Now adjust value so balance == cost exactly.
	msg.Value = new(big.Int).Sub(bal.ToBig(), new(big.Int).Mul(new(big.Int).SetUint64(msg.GasLimit), msg.GasFeeCap))
	require.Nil(t, f.k.validateSimCall(f.ctx, f.sdb, msg, f.header, f.rules, f.cfg))
}

func TestValidateSimCall_Balance_OneBelowFails_38014(t *testing.T) {
	f := newValidateSimCallFixture(t)
	msg := f.baselineMsg()
	bal, _ := uint256.FromDecimal(validateSimCallSenderBalance)
	// value = balance + 1 -> cost > balance.
	msg.Value = new(big.Int).Add(bal.ToBig(), big.NewInt(1))
	simErr := f.k.validateSimCall(f.ctx, f.sdb, msg, f.header, f.rules, f.cfg)
	require.NotNil(t, simErr)
	require.Equal(t, types.SimErrCodeInsufficientFunds, simErr.ErrorCode())
}

func TestValidateSimCall_InitCode_BoundaryEqual(t *testing.T) {
	f := newValidateSimCallFixture(t)
	msg := f.baselineMsg()
	msg.To = nil // CREATE
	msg.Data = make([]byte, params.MaxInitCodeSize)
	// Intrinsic for a CREATE this size is well above 21k; bump GasLimit.
	msg.GasLimit = 30_000_000
	require.Nil(t, f.k.validateSimCall(f.ctx, f.sdb, msg, f.header, f.rules, f.cfg))
}

func TestValidateSimCall_InitCode_OneOverFails_38025(t *testing.T) {
	f := newValidateSimCallFixture(t)
	msg := f.baselineMsg()
	msg.To = nil
	msg.Data = make([]byte, params.MaxInitCodeSize+1)
	msg.GasLimit = 30_000_000
	simErr := f.k.validateSimCall(f.ctx, f.sdb, msg, f.header, f.rules, f.cfg)
	require.NotNil(t, simErr)
	require.Equal(t, types.SimErrCodeMaxInitCodeSizeExceeded, simErr.ErrorCode())
}

// --- fork gating / null-guard --------------------------------------------

// Pre-Shanghai chain rules: the init-code gate must not fire even when
// data length exceeds MaxInitCodeSize. The gate is forked at Shanghai
// activation per EIP-3860.
func TestValidateSimCall_InitCode_PreShanghai_NoCheck(t *testing.T) {
	f := newValidateSimCallFixture(t)
	preShanghaiRules := f.rules
	preShanghaiRules.IsShanghai = false

	msg := f.baselineMsg()
	msg.To = nil
	msg.Data = make([]byte, params.MaxInitCodeSize+1)
	msg.GasLimit = 30_000_000
	require.Nil(t, f.k.validateSimCall(f.ctx, f.sdb, msg, f.header, preShanghaiRules, f.cfg))
}

// CALL with oversized calldata is NOT subject to the init-code gate;
// only CREATE (To == nil) is.
func TestValidateSimCall_InitCode_Call_NotCreate(t *testing.T) {
	f := newValidateSimCallFixture(t)
	msg := f.baselineMsg()
	msg.Data = make([]byte, params.MaxInitCodeSize+1)
	msg.GasLimit = 30_000_000
	// To is non-nil from baseline -> CALL.
	require.Nil(t, f.k.validateSimCall(f.ctx, f.sdb, msg, f.header, f.rules, f.cfg))
}

// Pre-London / no override: header.BaseFee == nil. Fee-cap check is
// skipped per the explicit nil guard in validateSimCall.
func TestValidateSimCall_FeeCap_NoBaseFeeHeader(t *testing.T) {
	f := newValidateSimCallFixture(t)
	header := *f.header
	header.BaseFee = nil
	msg := f.baselineMsg()
	msg.GasFeeCap = big.NewInt(0)
	require.Nil(t, f.k.validateSimCall(f.ctx, f.sdb, msg, &header, f.rules, f.cfg))
}

// --- ordering ------------------------------------------------------------
//
// Geth's preCheck order is nonce -> init-code -> intrinsic -> fee-cap ->
// balance. The driver mirrors that order so a tx that would fail
// multiple gates surfaces the leftmost code (matching geth's wire
// behavior). Each test below sets two gates to fail simultaneously and
// asserts the leftmost wins.

// Nonce mismatch + insufficient funds -> nonce wins.
func TestValidateSimCall_Order_NonceBeforeInsufficientFunds(t *testing.T) {
	f := newValidateSimCallFixture(t)
	msg := f.baselineMsg()
	msg.Nonce = 1 // nonce too high (state == 0)
	bal, _ := uint256.FromDecimal(validateSimCallSenderBalance)
	msg.Value = new(big.Int).Add(bal.ToBig(), big.NewInt(1)) // also fails balance

	simErr := f.k.validateSimCall(f.ctx, f.sdb, msg, f.header, f.rules, f.cfg)
	require.NotNil(t, simErr)
	require.Equal(t, types.SimErrCodeNonceTooHigh, simErr.ErrorCode())
}

// Init-code overflow + below-intrinsic gas -> init-code wins.
func TestValidateSimCall_Order_InitCodeBeforeIntrinsic(t *testing.T) {
	f := newValidateSimCallFixture(t)
	msg := f.baselineMsg()
	msg.To = nil
	msg.Data = make([]byte, params.MaxInitCodeSize+1)
	msg.GasLimit = 0 // also fails intrinsic
	simErr := f.k.validateSimCall(f.ctx, f.sdb, msg, f.header, f.rules, f.cfg)
	require.NotNil(t, simErr)
	require.Equal(t, types.SimErrCodeMaxInitCodeSizeExceeded, simErr.ErrorCode())
}

// Below-base fee-cap + below-intrinsic gas -> fee-cap wins.
func TestValidateSimCall_Order_FeeCapBeforeIntrinsic(t *testing.T) {
	f := newValidateSimCallFixture(t)
	msg := f.baselineMsg()
	msg.GasLimit = 20_999         // fails intrinsic
	msg.GasFeeCap = big.NewInt(0) // also fails fee-cap
	simErr := f.k.validateSimCall(f.ctx, f.sdb, msg, f.header, f.rules, f.cfg)
	require.NotNil(t, simErr)
	require.Equal(t, types.SimErrCodeFeeCapTooLow, simErr.ErrorCode())
}

// Below-base fee-cap + insufficient balance -> fee-cap wins.
func TestValidateSimCall_Order_FeeCapBeforeBalance(t *testing.T) {
	f := newValidateSimCallFixture(t)
	msg := f.baselineMsg()
	msg.GasFeeCap = big.NewInt(0) // fails fee-cap
	bal, _ := uint256.FromDecimal(validateSimCallSenderBalance)
	msg.Value = new(big.Int).Add(bal.ToBig(), big.NewInt(1)) // also fails balance
	simErr := f.k.validateSimCall(f.ctx, f.sdb, msg, f.header, f.rules, f.cfg)
	require.NotNil(t, simErr)
	require.Equal(t, types.SimErrCodeFeeCapTooLow, simErr.ErrorCode())
}

// --- determinism / no mutation -------------------------------------------

func TestValidateSimCall_Determinism(t *testing.T) {
	f := newValidateSimCallFixture(t)
	msg := f.baselineMsg()
	msg.Nonce = 1 // forces NonceTooHigh
	a := f.k.validateSimCall(f.ctx, f.sdb, msg, f.header, f.rules, f.cfg)
	b := f.k.validateSimCall(f.ctx, f.sdb, msg, f.header, f.rules, f.cfg)
	require.NotNil(t, a)
	require.NotNil(t, b)
	require.Equal(t, a.Code, b.Code)
	require.Equal(t, a.Message, b.Message)
}

// validateSimCall must not mutate the message it inspects; the driver
// reuses the same Message for the EVM call after the gate clears.
func TestValidateSimCall_DoesNotMutateMessage(t *testing.T) {
	f := newValidateSimCallFixture(t)
	msg := f.baselineMsg()
	snapshot := *msg

	require.Nil(t, f.k.validateSimCall(f.ctx, f.sdb, msg, f.header, f.rules, f.cfg))
	require.Equal(t, snapshot.From, msg.From)
	require.Equal(t, snapshot.Nonce, msg.Nonce)
	require.Equal(t, snapshot.GasLimit, msg.GasLimit)
	require.Equal(t, snapshot.GasFeeCap.String(), msg.GasFeeCap.String())
	require.Equal(t, snapshot.Value.String(), msg.Value.String())
}
