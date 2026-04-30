package keeper

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/big"
	"reflect"
	"slices"
	"sync/atomic"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/consensus/misc/eip1559"
	"github.com/ethereum/go-ethereum/core"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/eth/tracers"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/trie"

	"github.com/mezo-org/mezod/x/evm/statedb"
	"github.com/mezo-org/mezod/x/evm/types"
)

// simGasBudget tracks the request-wide gas pool for one eth_simulateV1
// invocation.
type simGasBudget struct{ remaining uint64 }

// newSimGasBudget builds a gas budget for one eth_simulateV1 request.
//
// gasCap == 0 means unlimited. The RPC backend injects the operator's
// RPCGasCap (positive default), so the keeper trusts the value as-is.
// A direct gRPC peer passing 0 disables the request-wide gas budget —
// by design, since operator config is not visible at this layer.
func newSimGasBudget(gasCap uint64) *simGasBudget {
	if gasCap == 0 {
		return &simGasBudget{remaining: math.MaxUint64}
	}
	return &simGasBudget{remaining: gasCap}
}

// clamp returns min(gas, b.remaining).
func (b *simGasBudget) clamp(gas uint64) uint64 {
	if gas > b.remaining {
		return b.remaining
	}
	return gas
}

// consume deducts amount from the remaining budget. The clamp invariant
// keeps amount <= remaining for valid inputs; the error path is the
// safety net.
func (b *simGasBudget) consume(amount uint64) error {
	if amount > b.remaining {
		return fmt.Errorf("RPC gas cap exhausted: need %d, remaining %d", amount, b.remaining)
	}
	b.remaining -= amount
	return nil
}

// sanitizeSimChain validates the block ordering rules and fills gaps
// with empty blocks. Returns the concatenated slice including any
// gap-fill blocks. Failures are *types.SimError values carrying
// spec-reserved JSON-RPC codes (-38020 / -38021 / -38026), returned
// via the plain error channel.
//
// Design mirrors go-ethereum's internal/ethapi/simulate.go::sanitizeChain
// (v1.15.4 source). Divergences: we never propagate a nil slice input
// (caller has already enforced len > 0 at the RPC boundary), and the
// span check against types.MaxSimulateBlocks is performed BEFORE gap-fill
// allocation so a pathological `[{Number: base+1}, {Number:
// base+10_000_000}]` input cannot drive the driver into a 10M-header
// allocation.
//
// Mutation contract: the input slice is cloned, but caller-supplied
// non-nil *SimBlockOverrides are aliased — when their Number or Time
// is nil, the resolver writes the defaulted value back through that
// shared pointer, visible to the caller.
func sanitizeSimChain(base *ethtypes.Header, blocks []types.SimBlock) ([]types.SimBlock, error) {
	cloned := slices.Clone(blocks)

	res := make([]types.SimBlock, 0, len(cloned))
	prevNumber := new(big.Int).Set(base.Number)
	prevTimestamp := base.Time

	for _, block := range cloned {
		if block.BlockOverrides == nil {
			block.BlockOverrides = new(types.SimBlockOverrides)
		}

		if block.BlockOverrides.Number == nil {
			n := new(big.Int).Add(prevNumber, big.NewInt(1))
			block.BlockOverrides.Number = (*hexutil.Big)(n)
		}
		num := block.BlockOverrides.Number.ToInt()

		// Span check against base.Number runs before gap-fill
		// allocation: any input Number more than types.MaxSimulateBlocks
		// past base fails immediately without materializing the
		// in-between headers.
		if span := new(big.Int).Sub(num, base.Number); span.Cmp(big.NewInt(types.MaxSimulateBlocks)) > 0 {
			return nil, types.NewSimClientLimitExceeded(span, types.MaxSimulateBlocks)
		}

		diff := new(big.Int).Sub(num, prevNumber)
		if diff.Sign() <= 0 {
			return nil, types.NewSimInvalidBlockNumber(num, prevNumber)
		}

		if diff.Cmp(big.NewInt(1)) > 0 {
			gap := new(big.Int).Sub(diff, big.NewInt(1))
			for i := uint64(0); i < gap.Uint64(); i++ {
				n := new(big.Int).Add(prevNumber, big.NewInt(int64(i+1))) //nolint:gosec
				t := prevTimestamp + types.SimTimestampIncrement
				res = append(res, types.SimBlock{BlockOverrides: &types.SimBlockOverrides{
					Number: (*hexutil.Big)(n),
					Time:   (*hexutil.Uint64)(&t),
				}})
				prevTimestamp = t
			}
		}
		prevNumber = num

		var t uint64
		if block.BlockOverrides.Time == nil {
			t = prevTimestamp + types.SimTimestampIncrement
			block.BlockOverrides.Time = (*hexutil.Uint64)(&t)
		} else {
			t = uint64(*block.BlockOverrides.Time)
			if t <= prevTimestamp {
				return nil, types.NewSimInvalidBlockTimestamp(t, prevTimestamp)
			}
		}
		prevTimestamp = t

		res = append(res, block)
	}
	return res, nil
}

// makeSimHeader constructs the preliminary header for a simulated
// block. Fields that depend on execution outcomes (GasUsed, final
// state/tx/recpt roots, block hash) are left unset; the driver patches
// them after the block's calls execute.
//
// Inputs:
//   - parent is the immediate predecessor header (base block for the
//     first simulated block, previous simulated header otherwise).
//   - overrides carries caller-supplied field overrides. May be nil.
//   - rules carries the active fork ruleset; used to derive the
//     header's fork-gated fields.
//   - chainCfg is used to compute the default base fee when the caller
//     leaves BaseFeePerGas unset and validation is enabled.
//   - validation chooses the default BaseFee: when true, derive from
//     eip1559.CalcBaseFee on the parent; when false, set to 0 (matches
//     the spec-conformant "skip base-fee checks" behavior).
//
// Mezo-specific: the header leaves WithdrawalsHash, ParentBeaconRoot,
// RequestsHash, BlobGasUsed, and ExcessBlobGas nil — the chain model
// does not support the underlying EIPs (4895, 4788, 7685, 4844).
func makeSimHeader(
	parent *ethtypes.Header,
	overrides *types.SimBlockOverrides,
	rules params.Rules,
	chainCfg *params.ChainConfig,
	validation bool,
) *ethtypes.Header {
	// Defaults derived from the parent and chain config. Override
	// application below layers caller-supplied values on top of these.
	h := &ethtypes.Header{
		ParentHash:  parent.Hash(),
		UncleHash:   ethtypes.EmptyUncleHash,
		TxHash:      ethtypes.EmptyTxsHash,
		ReceiptHash: ethtypes.EmptyReceiptsHash,
		Coinbase:    parent.Coinbase,
		GasLimit:    parent.GasLimit,
		Difficulty:  new(big.Int).Set(parent.Difficulty),
		Number:      new(big.Int).Add(parent.Number, big.NewInt(1)),
		Time:        parent.Time + types.SimTimestampIncrement,
	}

	// Post-merge: Difficulty is zero. MixDigest carries PREVRANDAO; mezo
	// does not support PREVRANDAO randomness, so we leave MixDigest at
	// its zero-value default — common.Hash is a value type ([32]byte),
	// so the field is already populated and does not need an explicit
	// assignment to satisfy go-ethereum's merge rule-switch.
	if chainCfg.MergeNetsplitBlock != nil {
		h.Difficulty = new(big.Int)
	}

	// Default base fee: validation=true derives via CalcBaseFee against
	// parent; validation=false reports zero so per-call gas-price checks
	// don't fail (spec-conformant "skip base-fee checks" relaxation).
	switch {
	case validation && rules.IsLondon:
		h.BaseFee = eip1559.CalcBaseFee(chainCfg, parent)
	default:
		h.BaseFee = new(big.Int)
	}

	// Caller-supplied overrides win over the defaults above.
	if overrides != nil {
		if overrides.Number != nil {
			h.Number = overrides.Number.ToInt()
		}
		if overrides.Time != nil {
			h.Time = uint64(*overrides.Time)
		}
		if overrides.Difficulty != nil {
			h.Difficulty = overrides.Difficulty.ToInt()
		}
		if overrides.GasLimit != nil {
			h.GasLimit = uint64(*overrides.GasLimit)
		}
		if overrides.FeeRecipient != nil {
			h.Coinbase = *overrides.FeeRecipient
		}
		if overrides.PrevRandao != nil {
			h.MixDigest = *overrides.PrevRandao
		}
		if overrides.BaseFeePerGas != nil {
			h.BaseFee = overrides.BaseFeePerGas.ToInt()
		}
	}

	return h
}

// assembleSimBlock builds the synthetic *ethtypes.Block for a simulated
// block. NewBlock derives transactionsRoot, receiptsRoot, and bloom
// from the supplied txs and receipts; CreateBloom in turn ORs each
// receipt's pre-computed Bloom. Caller must seal header.GasUsed before
// invoking so block.Hash() (referenced by NewSimBlockResult's marshal
// path) is stable.
//
// stateRoot stays at the header's zero Root: mezod's StateDB wraps a
// Cosmos cached multistore and has no MPT to call IntermediateRoot on,
// so any non-zero value would be misleading. Documented as a known
// Mezo divergence from the geth simulate envelope.
func assembleSimBlock(
	header *ethtypes.Header,
	txs []*ethtypes.Transaction,
	receipts []*ethtypes.Receipt,
) *ethtypes.Block {
	return ethtypes.NewBlock(header, &ethtypes.Body{Transactions: txs}, receipts, trie.NewStackTrie(nil))
}

// simulateV1 is the keeper-side entry point for eth_simulateV1. It
// sanitizes the chain, builds preliminary headers, applies state
// overrides, executes per-call, and assembles response envelopes.
//
// State-propagation invariant: ONE *statedb.StateDB is shared across
// every call and every block in the request. Both the EVM journal and
// mezo's StateDB-scoped cached-ctx (where custom precompile Cosmos-side
// writes live) ride on that single StateDB, so call/block continuity
// covers both layers uniformly. commit=false keeps the whole thing
// ephemeral. A fresh StateDB per block would silently drop
// custom-precompile Cosmos writes between blocks.
//
// Error policy: spec-coded failures (sanitize-chain violation,
// state-override validation, per-call preflight) are returned as
// *types.SimError through the plain error channel; callers branch with
// errors.As. Everything else is a genuine internal — the gRPC handler
// maps it to codes.Internal.
func (k *Keeper) simulateV1(
	ctx context.Context,
	sdkCtx sdk.Context,
	cfg *statedb.EVMConfig,
	base *ethtypes.Header,
	baseHash common.Hash,
	opts *types.SimOpts,
	gasCap uint64,
	timeout time.Duration,
) ([]*types.SimBlockResult, error) {
	if len(opts.BlockStateCalls) == 0 {
		return []*types.SimBlockResult{}, nil
	}

	sanitized, err := sanitizeSimChain(base, opts.BlockStateCalls)
	if err != nil {
		return nil, err
	}

	// Request-wide gas budget. Seeded from json-rpc.gas-cap; a zero cap
	// is interpreted as unlimited.
	budget := newSimGasBudget(gasCap)

	// One Rules value covers every simulated block, derived from sdkCtx
	// (i.e. base) rather than from each block's own number/time. This
	// is deliberate: applyMessageWithConfig reads sdkCtx.BlockHeight() /
	// sdkCtx.BlockTime() internally for fork-gated behavior (signer
	// rules, access-list prep, intrinsic gas). Since those internal
	// reads are anchored at sdkCtx, every piece of fork-gated machinery
	// below the driver sees the *base* ruleset regardless of which
	// simulated block's header we hand the EVM. Using a single
	// base-derived `rules` for header construction keeps the driver
	// aligned with that internal truth.
	//
	// The sentinel enforces this anchoring: if the span's last block
	// would cross onto a different fork than base, rules() anywhere in
	// the simulation would disagree with applyMessageWithConfig's
	// internal rules() — silently producing wrong-fork output. Fork
	// activation is monotonic in (height, time), so comparing base vs.
	// last-sim is sufficient: intermediate blocks are guaranteed to
	// match both endpoints when the endpoints match. Conservative where
	// it matters (a span fully inside a post-fork region whose base is
	// pre-fork is rejected rather than silently executed with
	// pre-fork rules from sdkCtx).
	rules := cfg.Rules(sdkCtx.BlockHeight(), uint64(sdkCtx.BlockTime().Unix())) //nolint:gosec
	lastBlock := sanitized[len(sanitized)-1]
	lastRules := cfg.Rules(
		lastBlock.BlockOverrides.Number.ToInt().Int64(),
		uint64(*lastBlock.BlockOverrides.Time),
	)
	if !sameForks(rules, lastRules) {
		return nil, types.NewSimForkSpanUnsupported()
	}

	// ONE StateDB for the entire request — see the package comment
	// above. Per-block StateDB would silently drop custom-precompile
	// Cosmos writes between blocks.
	//
	// The initial TxConfig carries a zero BlockHash because the
	// simulated block hash for any block is not known until that
	// block's calls have all executed and cumulative gas is sealed
	// into the header. processSimBlock updates per-call TxHash /
	// TxIndex via SetTxContext (geth-aligned: BlockHash untouched)
	// and back-stamps log.BlockHash with the simulated block hash
	// after the block finalizes.
	sdb := statedb.New(sdkCtx, k, statedb.NewEmptyTxConfig(common.Hash{}))

	// Build, execute, and link headers in a single pass. processSimBlock
	// finalizes its own header (GasUsed sealed, hash stable) before
	// returning, so the next iteration's parent.Hash() already reflects
	// the final hash — the response envelope's parent chain is coherent
	// by construction.
	headers := make([]*ethtypes.Header, 0, len(sanitized))
	results := make([]*types.SimBlockResult, 0, len(sanitized))
	for _, block := range sanitized {
		header, res, blockErr := k.processSimBlock(
			ctx, sdkCtx, cfg, sdb, base, headers, baseHash, block, rules, opts, gasCap, budget, timeout,
		)
		if blockErr != nil {
			return nil, blockErr
		}

		headers = append(headers, header)
		results = append(results, res)
	}

	return results, nil
}

// processSimBlock executes one simulated block against the shared
// StateDB. It builds the block's header from the last `pastSiblings`
// entry (or `base` when empty), applies the block's StateOverrides
// (incl. MovePrecompileTo), builds the per-block BlockContext with the
// simulate-aware GetHashFn, runs the block's calls sequentially with
// cumulative gas accounting and per-call ephemeral resets, and
// assembles the response envelope.
//
// `pastSiblings` is read-only; `newSimGetHashFn` consults it for
// `BLOCKHASH` resolution in the simulated-sibling range. For the first
// block (empty `pastSiblings`), a non-zero `baseHash` is used as
// `header.ParentHash` — the canonical CometBFT block hash supplied by
// the rpc/backend layer, since `base.Hash()` is the
// `baseHeaderFromContext` synthetic hash, not the canonical one.
//
// Returns the finalized header (GasUsed sealed, Hash() stable) alongside
// the assembled block result.
func (k *Keeper) processSimBlock(
	ctx context.Context,
	sdkCtx sdk.Context,
	cfg *statedb.EVMConfig,
	sdb *statedb.StateDB,
	base *ethtypes.Header,
	pastSiblings []*ethtypes.Header,
	baseHash common.Hash,
	block types.SimBlock,
	rules params.Rules,
	opts *types.SimOpts,
	gasCap uint64,
	budget *simGasBudget,
	timeout time.Duration,
) (*ethtypes.Header, *types.SimBlockResult, error) {
	parent := base
	if n := len(pastSiblings); n > 0 {
		parent = pastSiblings[n-1]
	}
	header := makeSimHeader(parent, block.BlockOverrides, rules, cfg.ChainConfig, opts.Validation)

	if len(pastSiblings) == 0 && baseHash != (common.Hash{}) {
		header.ParentHash = baseHash
	}

	// -38012: caller-supplied BlockOverrides.BaseFeePerGas must not fall
	// below the chain-computed eip1559 floor. Per-block (not per-call),
	// hoisted out of validateSimCall to keep that helper focused on
	// message-level checks. Skipped when validation=false, when no
	// override is supplied, or pre-London (no floor exists).
	if opts.Validation &&
		rules.IsLondon &&
		block.BlockOverrides != nil &&
		block.BlockOverrides.BaseFeePerGas != nil {
		floor := eip1559.CalcBaseFee(cfg.ChainConfig, parent)
		if override := block.BlockOverrides.BaseFeePerGas.ToInt(); override.Cmp(floor) < 0 {
			return nil, nil, types.NewSimBaseFeeTooLow(override, floor)
		}
	}

	var moves map[common.Address]common.Address
	if len(block.StateOverrides) > 0 {
		m, applyErr := applyStateOverrides(sdb, block.StateOverrides, rules)
		if applyErr != nil {
			return nil, nil, applyErr
		}
		moves = m
	}

	// Gate Random on the merge fork: go-ethereum flips its rule set to
	// Paris when BlockContext.Random != nil. Pre-merge we must leave it
	// nil to preserve the legacy opcode set, matching the derivation
	// used by NewEVMWithOverrides.
	var random *common.Hash
	if cfg.ChainConfig.MergeNetsplitBlock != nil {
		random = &header.MixDigest
	}

	blockCtx := vm.BlockContext{
		CanTransfer: core.CanTransfer,
		Transfer:    core.Transfer,
		// Simulate-aware GetHashFn: canonical-range delegated to
		// k.GetHashFn, simulated-sibling range scanned from already-
		// finalized past siblings. Canonical-range hashes are therefore
		// unforgeable by any BlockOverrides field.
		GetHash:     newSimGetHashFn(k.GetHashFn(sdkCtx), base, pastSiblings),
		Coinbase:    header.Coinbase,
		GasLimit:    header.GasLimit,
		BlockNumber: new(big.Int).Set(header.Number),
		Time:        header.Time,
		Difficulty:  new(big.Int).Set(header.Difficulty),
		BaseFee:     header.BaseFee,
		BlobBaseFee: big.NewInt(0),
		Random:      random,
	}

	precompiles := k.precompilesWithMoves(sdkCtx, cfg, moves)

	// One watcher goroutine per block. OnEVMConstructed publishes the
	// per-call *vm.EVM into liveEVM synchronously, before
	// applyMessageWithConfig starts the EVM, so a non-nil load here
	// always points at the call we want to cancel. The ctx.Err() guard
	// distinguishes upstream cancellation (evm.Cancel()) from the
	// deferred cancelBlock that disarms the watcher on normal exit.
	watchCtx, cancelBlock := context.WithCancel(ctx)
	defer cancelBlock()

	var liveEVM atomic.Pointer[vm.EVM]
	go func() {
		<-watchCtx.Done()
		if ctx.Err() != nil {
			if e := liveEVM.Load(); e != nil {
				e.Cancel()
			}
		}
	}()

	// validation=false skips base-fee and balance checks; validation=true
	// runs the realistic path.
	noBaseFee := !opts.Validation
	evmOverrides := &EVMOverrides{
		BlockContext: &blockCtx,
		Precompiles:  precompiles,
		NoBaseFee:    &noBaseFee,
		OnEVMConstructed: func(evm *vm.EVM) {
			liveEVM.Store(evm)
		},
	}

	var (
		tt            *simTracer
		perCallTracer *tracers.Tracer
	)
	if opts.TraceTransfers {
		tt = newSimTracer(opts.TraceTransfers, header.Number.Uint64(), common.Hash{})
		hooks := tt.Hooks()
		perCallTracer = &tracers.Tracer{Hooks: hooks}
		sdb.SetTracingHooks(hooks)
		defer sdb.SetTracingHooks(nil)
	}

	calls := make([]types.SimCallResult, 0, len(block.Calls))
	// Synthetic txs and receipts feed NewBlock so the assembled envelope
	// has correct transactionsRoot, receiptsRoot, and bloom. They are
	// kept for the lifetime of one block — bounded by MaxSimulateCalls
	// and the per-block GasLimit.
	txs := make([]*ethtypes.Transaction, 0, len(block.Calls))
	receipts := make([]*ethtypes.Receipt, 0, len(block.Calls))
	senders := make([]common.Address, 0, len(block.Calls))
	var cumGas uint64

	for i := range block.Calls {
		if err := ctx.Err(); err != nil {
			return nil, nil, types.NewSimTimeout(timeout)
		}

		// Clear per-call ephemerals (logs, refund, transient storage,
		// precompile call counter) while preserving account/storage
		// mutations. Runs unconditionally at the top of every call —
		// idempotent on a fresh StateDB, covers both call boundaries
		// within a block and block boundaries (call 0 of block N+1).
		sdb.FinaliseBetweenCalls()

		args := block.Calls[i]
		args.Nonce = resolveSimCallNonce(sdb, &args)

		var simErr *types.SimError
		args.Gas, simErr = resolveSimCallGas(&args, header, cumGas, budget)
		if simErr != nil {
			// -38015 is a request-level code per the geth execution spec
			// (ethereum/execution-apis `execute.yaml`).
			return nil, nil, simErr
		}

		msg, msgErr := args.ToMessage(gasCap, header.BaseFee)
		if msgErr != nil {
			// -32602 is a request-level code per the geth execution spec
			// (ethereum/execution-apis `execute.yaml`).
			return nil, nil, types.NewSimInvalidParams(msgErr.Error())
		}

		// State overrides may make `from` a contract; the EoA check
		// must stay off for the simulator regardless of validation mode.
		msg.SkipAccountChecks = true

		if opts.Validation {
			if simErr := k.validateSimCall(sdkCtx, sdb, &msg, header, rules, cfg.ChainConfig); simErr != nil {
				return nil, nil, simErr
			}
		}

		// Per-call TxConfig so AddLog stamps distinct TxHash / TxIndex
		// on every emitted log. BlockHash is left zero in callCfg
		// because the block hash depends on cumulative GasUsed, which
		// isn't known until every call in this block has executed; the
		// post-pass below patches log.BlockHash once that hash is
		// computed. SetTxContext mirrors geth's
		// `(*state.StateDB).SetTxContext(thash, ti)` — it only updates
		// the StateDB's per-call TxHash / TxIndex while leaving the
		// pre-set BlockHash and LogIndex alone.
		// DynamicFeeTxType matches go-ethereum v1.16's eth_simulateV1
		// driver (`internal/ethapi/simulate.go`), so when Mezo upgrades
		// off v1.14 the wire-shape of the response stays stable.
		simTx := buildSimTx(&args, cfg.ChainConfig.ChainID, ethtypes.DynamicFeeTxType)
		callTxHash := simTx.Hash()
		callIdx := len(txs)
		callCfg := statedb.NewTxConfig(common.Hash{}, callTxHash, uint(callIdx), 0) //nolint:gosec
		sdb.SetTxContext(callCfg.TxHash, callIdx)

		if perCallTracer != nil {
			tt.reset(callTxHash, callIdx)
		}

		res, _, runErr := k.applyMessageWithConfig(
			sdkCtx,
			WrapMessage(msg),
			perCallTracer,
			false, // commit = false (ephemeral simulate)
			cfg,
			callCfg,
			sdb,
			evmOverrides,
		)

		// Canceled mid-call: ignore any vm-error artifact and return -32016.
		if err := ctx.Err(); err != nil {
			return nil, nil, types.NewSimTimeout(timeout)
		}

		if runErr != nil {
			// Per the geth execution spec (ethereum/execution-apis
			// `execute.yaml`), CallResultFailure permits only codes 3
			// and -32015, so ErrIntrinsicGas must surface at the request
			// level. Capture args.Gas here while it is still in scope so
			// the SimError carries the actual provided value rather than
			// a (0, 0) reconstruction at the gRPC handler.
			if errors.Is(runErr, core.ErrIntrinsicGas) {
				var provided uint64
				if args.Gas != nil {
					provided = uint64(*args.Gas)
				}
				return nil, nil, types.NewSimIntrinsicGas(provided, 0)
			}
			return nil, nil, runErr
		}

		// Mirror the CREATE branch's post-call SetNonce in
		// applyMessageWithConfig (state_transition.go:580). For
		// MsgEthereumTx, CALL nonce progression is owned by the
		// ante handler — but eth_simulateV1 bypasses ante, so the
		// driver has to advance the StateDB nonce here. Without
		// this, multi-CALL chains from the same sender all read
		// the same defaulted nonce, producing colliding synthetic
		// tx hashes and wrong CREATE addresses for any follow-up
		// CREATE in the same block. Anchored on msg.Nonce+1 (not
		// GetNonce+1) so the post-state mirrors CREATE when the
		// caller supplies an explicit args.Nonce that diverges
		// from the StateDB value. Runs regardless of res.VmError —
		// a reverted CALL still consumes the nonce on the real
		// chain.
		if msg.To != nil {
			sdb.SetNonce(msg.From, msg.Nonce+1)
		}

		// Request-wide gas pool. The clamp in resolveSimCallGas keeps
		// res.GasUsed <= budget.remaining under valid inputs; the error
		// path is the safety net.
		if budgetErr := budget.consume(res.GasUsed); budgetErr != nil {
			return nil, nil, budgetErr
		}

		callResult := types.BuildSimCallResult(res)
		if perCallTracer != nil {
			tracerLogs := tt.Logs()
			if tracerLogs == nil {
				tracerLogs = []*ethtypes.Log{}
			}
			callResult.Logs = tracerLogs
		}
		calls = append(calls, callResult)
		cumGas += res.GasUsed

		// Build the synthetic receipt for this call. Fields mirror
		// core/state_processor.go's receipt assembly; PostState stays
		// nil because mezod's StateDB has no MPT root, BlockHash is
		// back-stamped after header.Hash() stabilizes below.
		status := ethtypes.ReceiptStatusSuccessful
		if res.Failed() {
			status = ethtypes.ReceiptStatusFailed
		}
		receipt := &ethtypes.Receipt{
			Type:              simTx.Type(),
			Status:            status,
			CumulativeGasUsed: cumGas,
			TxHash:            callTxHash,
			GasUsed:           res.GasUsed,
			Logs:              callResult.Logs,
			BlockNumber:       new(big.Int).Set(header.Number),
			TransactionIndex:  uint(callIdx), //nolint:gosec
		}
		if msg.To == nil {
			receipt.ContractAddress = crypto.CreateAddress(msg.From, msg.Nonce)
		}
		receipt.Bloom = ethtypes.CreateBloom(ethtypes.Receipts{receipt})

		txs = append(txs, simTx)
		receipts = append(receipts, receipt)
		senders = append(senders, msg.From)
	}

	// Finalize the header (GasUsed must be set before NewBlock derives
	// the trie roots and bloom) and assemble the *ethtypes.Block. NewBlock
	// stamps header.TxHash, header.ReceiptHash, and header.Bloom from the
	// supplied txs/receipts; block.Hash() reads from the resulting header.
	header.GasUsed = cumGas
	ethBlock := assembleSimBlock(header, txs, receipts)
	finalBlockHash := ethBlock.Hash()

	// Back-stamp the values that aren't knowable until the block is
	// sealed: log.BlockHash (depends on cumulative GasUsed via
	// header.Hash()), log.Index (must be per-block monotonic; AddLog
	// stamps from txConfig.LogIndex+len(s.logs), both of which reset
	// between calls), and receipt.BlockHash (mirrors the per-log stamp
	// for envelope consistency). log.BlockNumber is also normalized
	// here: custom precompiles stamp it from sdkCtx.BlockHeight() in
	// EmitEvent, and the driver anchors sdkCtx at base for fork-gated
	// reads — so without this normalization, precompile-emitted logs
	// report the parent block's number while regular EVM logs report
	// the simulated header's number.
	var logIdx uint
	for i := range calls {
		for _, log := range calls[i].Logs {
			log.BlockNumber = header.Number.Uint64()
			log.BlockHash = finalBlockHash
			log.Index = logIdx
			logIdx++
		}
	}
	for _, r := range receipts {
		r.BlockHash = finalBlockHash
	}

	return ethBlock.Header(), types.NewSimBlockResult(
		ethBlock, senders, opts.ReturnFullTransactions, cfg.ChainConfig, calls,
	), nil
}

// nonceSource narrows *statedb.StateDB down to the single method
// resolveSimCallNonce consumes. Keeping the function signature
// interface-typed lets helper unit tests pass a fake without
// instantiating a full keeper.
type nonceSource interface {
	GetNonce(common.Address) uint64
}

// resolveSimCallNonce returns the nonce to apply to a sim call:
// args.Nonce when explicitly set, otherwise a freshly allocated
// pointer holding the StateDB nonce for args.From. Defaulting from the
// shared StateDB lets multi-call chains implicitly advance for CREATE
// callers — note that applyMessageWithConfig only bumps the StateDB
// nonce on contract creation, so back-to-back non-CREATE calls from
// the same sender read the same default nonce; harmless for simulate,
// which has no uniqueness check.
func resolveSimCallNonce(
	sdb nonceSource,
	args *types.TransactionArgs,
) *hexutil.Uint64 {
	if args.Nonce != nil {
		return args.Nonce
	}
	n := hexutil.Uint64(sdb.GetNonce(args.GetFrom()))
	return &n
}

// resolveSimCallGas returns the gas limit to apply to a sim call as
// min(args.Gas | header-remaining, budget-remaining). When args.Gas is
// nil the resolver defaults to header-remaining. -38015 is returned
// when args.Gas exceeds the per-block remaining or when the per-block
// remaining is zero; the caller surfaces it as a request-level fatal.
//
// header.GasLimit is authoritative: a caller who sets
// blockOverrides.gasLimit to 0 (or otherwise produces a zero-limit
// block) gets every call rejected at the preflight, keeping the
// per-block work bound real even under hostile BlockOverrides.
func resolveSimCallGas(
	args *types.TransactionArgs,
	header *ethtypes.Header,
	cumGasUsed uint64,
	budget *simGasBudget,
) (*hexutil.Uint64, *types.SimError) {
	var remaining uint64
	if header.GasLimit > cumGasUsed {
		remaining = header.GasLimit - cumGasUsed
	}

	if remaining == 0 {
		var requested uint64
		if args.Gas != nil {
			requested = uint64(*args.Gas)
		}
		return nil, types.NewSimBlockGasLimitReached(requested, 0)
	}

	var resolved uint64
	if args.Gas == nil {
		resolved = remaining
	} else {
		requested := uint64(*args.Gas)
		if requested > remaining {
			return nil, types.NewSimBlockGasLimitReached(requested, remaining)
		}
		resolved = requested
	}
	resolved = budget.clamp(resolved)
	g := hexutil.Uint64(resolved)
	return &g, nil
}

// newSimGetHashFn builds the simulate-aware BLOCKHASH resolver. The
// closure captures base + the already-finalized past-sibling headers
// (headers[:bi] from the caller). Resolution order:
//
//   - height <= base.Number → delegate to the canonical resolver,
//     which returns the canonical CometBFT header hash. Callers supply
//     k.GetHashFn(ctx), whose ctx is anchored at base by the gRPC
//     layer via rpctypes.ContextWithHeight.
//   - height  > base.Number → look up the simulated sibling by height;
//     return header.Hash() on hit.
//   - not found → zero hash.
//
// Canonical-range hashes are thus unforgeable by any BlockOverrides
// field: sim[] only holds simulated blocks whose Number > base.Number
// (enforced by sanitizeSimChain). This diverges from go-ethereum's
// simulate.go (which uses base.Hash() for the height==base case)
// because mezod surfaces CometBFT header hashes as canonical block
// hashes, not ethtypes.Header.Hash() values.
//
// The function takes a canonical vm.GetHashFunc (rather than a Keeper +
// ctx) so unit tests can exercise every resolution path against a
// synthetic canonical resolver without a full keeper instance.
func newSimGetHashFn(
	canonical vm.GetHashFunc,
	base *ethtypes.Header,
	sim []*ethtypes.Header,
) vm.GetHashFunc {
	baseN := base.Number.Uint64()
	// Index simulated siblings by height so BLOCKHASH stays O(1) per
	// opcode invocation (a single contract call can issue many).
	simByHeight := make(map[uint64]*ethtypes.Header, len(sim))
	for _, h := range sim {
		simByHeight[h.Number.Uint64()] = h
	}
	return func(height uint64) common.Hash {
		if height <= baseN {
			return canonical(height)
		}
		if h, ok := simByHeight[height]; ok {
			return h.Hash()
		}
		return common.Hash{}
	}
}

// sameForks reports whether two Rules values share the same fork
// activation bitmap. The driver's fork-boundary sentinel uses it to
// ensure all simulated blocks fall within a single fork.
//
// ChainConfig.Rules allocates a fresh *big.Int ChainID on every call,
// so direct struct equality always fails on the pointer. reflect.DeepEqual
// dereferences the ChainID and compares every remaining field by value,
// which keeps the "bitmap" framing correct as go-ethereum adds new rule
// flags. Within one chain ChainID is invariant, so including it in the
// comparison is a no-op.
func sameForks(a, b params.Rules) bool {
	return reflect.DeepEqual(a, b)
}

// buildSimTx materializes the synthetic *ethtypes.Transaction for a
// simulated call. Returning the transaction (rather than just the hash)
// lets the driver pass the same value into NewBlock to derive
// transactionsRoot and into the Senders map without rebuilding it.
//
// The envelope shape mirrors go-ethereum v1.16's
// `internal/ethapi.TransactionArgs.ToTransaction(defaultType)` so that
// Mezo's wire shape stays consistent across the pinned v1.14 baseline
// and the v1.16 upgrade target. The selection rule, in order:
//
//  1. `MaxFeePerGas` set, or `defaultType == DynamicFeeTxType` →
//     DynamicFeeTx (accessList is nested if provided).
//  2. `AccessList` set, or `defaultType == AccessListTxType` →
//     AccessListTx.
//  3. Otherwise LegacyTx.
//  4. As a final override, if the caller provided `gasPrice`, force
//     LegacyTx — geth uses this to fall back to a legacy envelope
//     whenever the legacy fee field is present, even when defaultType
//     would otherwise pick a typed one.
//
// Simulate txs are unsigned: V/R/S = 0, so tx.Hash() is stable and
// distinct per request — sufficient for the assembled block's
// per-call identifiers, transactionsRoot, and the typed JSON-RPC
// representation returned to the caller. Mezo rejects blob (type 3)
// and set-code (type 4) txs upstream, so this switch covers types 0/1/2
// only.
//
// chainID flows in from the chain config so type-1/2 hashes match the
// network the simulator runs against. args.Nonce and args.Gas are
// expected to be resolved by the caller before this function runs;
// other args are tolerant of nil pointers (treated as zero).
func buildSimTx(args *types.TransactionArgs, chainID *big.Int, defaultType uint8) *ethtypes.Transaction {
	nonce := uint64(0)
	if args.Nonce != nil {
		nonce = uint64(*args.Nonce)
	}
	gas := uint64(0)
	if args.Gas != nil {
		gas = uint64(*args.Gas)
	}
	value := bigOrZero(args.Value)
	data := args.GetData()

	usedType := uint8(ethtypes.LegacyTxType)
	switch {
	case args.MaxFeePerGas != nil || defaultType == ethtypes.DynamicFeeTxType:
		usedType = ethtypes.DynamicFeeTxType
	case args.AccessList != nil || defaultType == ethtypes.AccessListTxType:
		usedType = ethtypes.AccessListTxType
	}
	if args.GasPrice != nil {
		usedType = ethtypes.LegacyTxType
	}

	switch usedType {
	case ethtypes.DynamicFeeTxType:
		al := ethtypes.AccessList{}
		if args.AccessList != nil {
			al = *args.AccessList
		}
		return ethtypes.NewTx(&ethtypes.DynamicFeeTx{
			ChainID:    chainID,
			Nonce:      nonce,
			GasTipCap:  bigOrZero(args.MaxPriorityFeePerGas),
			GasFeeCap:  bigOrZero(args.MaxFeePerGas),
			Gas:        gas,
			To:         args.To,
			Value:      value,
			Data:       data,
			AccessList: al,
		})
	case ethtypes.AccessListTxType:
		return ethtypes.NewTx(&ethtypes.AccessListTx{
			ChainID:    chainID,
			Nonce:      nonce,
			GasPrice:   bigOrZero(args.GasPrice),
			Gas:        gas,
			To:         args.To,
			Value:      value,
			Data:       data,
			AccessList: *args.AccessList,
		})
	default:
		return ethtypes.NewTx(&ethtypes.LegacyTx{
			Nonce:    nonce,
			GasPrice: bigOrZero(args.GasPrice),
			Gas:      gas,
			To:       args.To,
			Value:    value,
			Data:     data,
		})
	}
}

// bigOrZero unwraps a hexutil.Big into a *big.Int, returning a fresh
// zero when the pointer is nil. Keeps buildSimTx's typed-tx
// constructors free of nil-fee fields, which would otherwise panic
// inside ethtypes' RLP encoding.
func bigOrZero(v *hexutil.Big) *big.Int {
	if v == nil {
		return new(big.Int)
	}
	return v.ToInt()
}

// validateSimCall runs the per-call validation=true gates in the order
// geth's state_transition.go enforces: nonce, fee-cap-vs-baseFee, then
// balance (preCheck), and finally init-code-size and intrinsic gas
// (execute). The block-baseFee floor (-38012) lives in processSimBlock
// since it depends on the parent and the override pointer, not on the
// message.
func (k *Keeper) validateSimCall(
	ctx sdk.Context,
	sdb *statedb.StateDB,
	msg *core.Message,
	header *ethtypes.Header,
	rules params.Rules,
	chainCfg *params.ChainConfig,
) *types.SimError {
	stateNonce := sdb.GetNonce(msg.From)
	switch {
	case msg.Nonce < stateNonce:
		return types.NewSimNonceTooLow(msg.From, msg.Nonce, stateNonce)
	case msg.Nonce > stateNonce:
		return types.NewSimNonceTooHigh(msg.From, msg.Nonce, stateNonce)
	}

	if header.BaseFee != nil && msg.GasFeeCap.Cmp(header.BaseFee) < 0 {
		return types.NewSimFeeCapTooLow(msg.GasFeeCap, header.BaseFee)
	}

	cost := new(big.Int).SetUint64(msg.GasLimit)
	cost.Mul(cost, msg.GasFeeCap)
	cost.Add(cost, msg.Value)
	balance := sdb.GetBalance(msg.From).ToBig()
	if balance.Cmp(cost) < 0 {
		return types.NewSimInsufficientFunds(msg.From, balance, cost)
	}

	contractCreation := msg.To == nil
	if contractCreation && rules.IsShanghai && len(msg.Data) > params.MaxInitCodeSize {
		return types.NewSimInitcodeTooLarge(len(msg.Data), params.MaxInitCodeSize)
	}

	intrinsic, err := k.GetEthIntrinsicGas(ctx, *msg, chainCfg, contractCreation)
	if err != nil {
		return types.NewSimIntrinsicGas(msg.GasLimit, math.MaxUint64)
	}
	if msg.GasLimit < intrinsic {
		return types.NewSimIntrinsicGas(msg.GasLimit, intrinsic)
	}

	return nil
}
