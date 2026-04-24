package keeper

import (
	"fmt"
	"math/big"
	"slices"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/consensus/misc/eip1559"
	"github.com/ethereum/go-ethereum/core"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"

	"github.com/mezo-org/mezod/x/evm/statedb"
	"github.com/mezo-org/mezod/x/evm/types"
)

const (
	// maxSimulateBlocks caps the number of simulated blocks in a single
	// request. Matches geth's hard-coded bound.
	maxSimulateBlocks = 256

	// simTimestampIncrement is the default gap, in seconds, between
	// sequential simulated blocks when the caller omits Time overrides.
	// Matches mezo's ~3s average CometBFT block time so callers who let
	// the sim fabricate timestamps land in a realistic ballpark.
	simTimestampIncrement = 3
)

// sanitizeSimChain validates the block ordering rules and fills gaps
// with empty blocks. It works on a shallow clone of the input slice so
// top-level SimBlock field writes in the loop (notably the
// BlockOverrides pointer reassignment for nil entries) never reach the
// caller's slice. Returns the concatenated slice including any
// gap-fill blocks. Failures are *types.SimError values carrying
// spec-reserved JSON-RPC codes (-38020 / -38021 / -38026), returned
// via the plain error channel.
//
// Design mirrors go-ethereum's internal/ethapi/simulate.go::sanitizeChain
// (v1.15.4 source). Divergences: we never propagate a nil slice input
// (caller has already enforced len > 0 at the RPC boundary), and the
// span check against maxSimulateBlocks is performed BEFORE gap-fill
// allocation so a pathological `[{Number: base+1}, {Number:
// base+10_000_000}]` input cannot drive the driver into a 10M-header
// allocation.
func sanitizeSimChain(base *ethtypes.Header, blocks []types.SimBlock) ([]types.SimBlock, error) {
	// Work on a clone so the loop's per-block edits stay confined to
	// our copy and never touch the caller's slice.
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
		// allocation: any input Number more than maxSimulateBlocks
		// past base fails immediately without materializing the
		// in-between headers.
		if span := new(big.Int).Sub(num, base.Number); span.Cmp(big.NewInt(maxSimulateBlocks)) > 0 {
			return nil, types.NewSimClientLimitExceeded(span, maxSimulateBlocks)
		}

		diff := new(big.Int).Sub(num, prevNumber)
		if diff.Sign() <= 0 {
			return nil, types.NewSimInvalidBlockNumber(num, prevNumber)
		}

		if diff.Cmp(big.NewInt(1)) > 0 {
			gap := new(big.Int).Sub(diff, big.NewInt(1))
			for i := uint64(0); i < gap.Uint64(); i++ {
				n := new(big.Int).Add(prevNumber, big.NewInt(int64(i+1))) //nolint:gosec
				t := prevTimestamp + simTimestampIncrement
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
			t = prevTimestamp + simTimestampIncrement
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
		Time:        parent.Time + simTimestampIncrement,
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

// assembleSimBlock turns a simulated header plus the calls that ran
// inside it into a spec-shaped block envelope suitable for JSON
// marshaling. Fields match RPCMarshalBlock's output. The assembled
// block carries tx hashes only; `returnFullTransactions` patching is
// not wired yet.
//
// `cumulativeGasUsed` is the sum of GasUsed across every call in the
// block (including reverts, whose gas is still consumed). It patches
// the header's GasUsed field prior to hashing so that downstream
// consumers see a consistent header/block relation.
func assembleSimBlock(
	header *ethtypes.Header,
	txHashes []common.Hash,
	cumulativeGasUsed uint64,
) map[string]interface{} {
	// Patch the header's GasUsed before hashing; keep the rest of the
	// scaffolding as-is.
	header.GasUsed = cumulativeGasUsed
	blockHash := header.Hash()

	txList := make([]interface{}, 0, len(txHashes))
	for _, h := range txHashes {
		txList = append(txList, h)
	}

	return map[string]interface{}{
		"number":           (*hexutil.Big)(new(big.Int).Set(header.Number)),
		"hash":             blockHash,
		"parentHash":       header.ParentHash,
		"nonce":            ethtypes.BlockNonce{},
		"mixHash":          header.MixDigest,
		"sha3Uncles":       header.UncleHash,
		"logsBloom":        header.Bloom,
		"stateRoot":        header.Root,
		"miner":            header.Coinbase,
		"difficulty":       (*hexutil.Big)(header.Difficulty),
		"extraData":        hexutil.Bytes(header.Extra),
		"size":             hexutil.Uint64(0),
		"gasLimit":         hexutil.Uint64(header.GasLimit),
		"gasUsed":          hexutil.Uint64(header.GasUsed),
		"timestamp":        hexutil.Uint64(header.Time),
		"transactionsRoot": header.TxHash,
		"receiptsRoot":     header.ReceiptHash,
		"baseFeePerGas":    (*hexutil.Big)(header.BaseFee),
		"uncles":           []common.Hash{},
		"transactions":     txList,
	}
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
// ephemeral. See .claude/MEZO-4227-eth-simulate-v1/precompiles-caveat.md
// ("Option A") for the full rationale and the failure mode a fresh
// StateDB per block would produce.
//
// Error policy: spec-coded failures (sanitize-chain violation,
// state-override validation, per-call preflight) are returned as
// *types.SimError through the plain error channel; callers branch with
// errors.As. Everything else is a genuine internal — the gRPC handler
// maps it to codes.Internal.
func (k *Keeper) simulateV1(
	ctx sdk.Context,
	cfg *statedb.EVMConfig,
	base *ethtypes.Header,
	opts *types.SimOpts,
	gasCap uint64,
) ([]*types.SimBlockResult, error) {
	if len(opts.BlockStateCalls) == 0 {
		return []*types.SimBlockResult{}, nil
	}

	sanitized, err := sanitizeSimChain(base, opts.BlockStateCalls)
	if err != nil {
		return nil, err
	}

	// One Rules value covers every simulated block, derived from ctx
	// (i.e. base) rather than from each block's own number/time. This
	// is deliberate: applyMessageWithConfig reads ctx.BlockHeight() /
	// ctx.BlockTime() internally for fork-gated behavior (signer rules,
	// access-list prep at state_transition.go:568, intrinsic gas at
	// state_transition.go:533). Since those internal reads are anchored
	// at ctx, every piece of fork-gated machinery below the driver sees
	// the *base* ruleset regardless of which simulated block's header
	// we hand the EVM. Using a single base-derived `rules` for header
	// construction keeps the driver aligned with that internal truth.
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
	// pre-fork rules from ctx).
	rules := cfg.Rules(ctx.BlockHeight(), uint64(ctx.BlockTime().Unix())) //nolint:gosec
	lastBlock := sanitized[len(sanitized)-1]
	lastRules := cfg.Rules(
		lastBlock.BlockOverrides.Number.ToInt().Int64(),
		uint64(*lastBlock.BlockOverrides.Time),
	)
	if !sameForks(rules, lastRules) {
		return nil, fmt.Errorf(
			"simulate: span crosses a fork boundary; not yet supported",
		)
	}

	// Preliminary headers for every simulated block. The parent chain is
	// fixed up front so newSimGetHashFn can resolve simulated-sibling
	// hashes during an earlier block's execution. Each header's GasUsed
	// (and therefore final Hash) is patched inside processSimBlock after
	// the block's calls run — later blocks only see already-finalized
	// past siblings via headers[:bi] from the call site, so the patch
	// order never surfaces a stale hash.
	headers := make([]*ethtypes.Header, len(sanitized))
	parent := base
	for bi, block := range sanitized {
		headers[bi] = makeSimHeader(parent, block.BlockOverrides, rules, cfg.ChainConfig, opts.Validation)
		parent = headers[bi]
	}

	// ONE StateDB for the entire request — see the package comment above
	// and the precompile caveat doc. Per-block StateDB would silently
	// drop custom-precompile Cosmos writes between blocks.
	//
	// The initial TxConfig carries only the base-block hash as a
	// placeholder; processSimBlock immediately overwrites it per-call
	// via SetTxConfig so each call's logs carry their own TxHash /
	// TxIndex, and back-stamps log.BlockHash with the simulated block
	// hash after the block finalizes.
	sdb := statedb.New(ctx, k, statedb.NewEmptyTxConfig(common.BytesToHash(ctx.HeaderHash())))

	results := make([]*types.SimBlockResult, 0, len(sanitized))
	for bi, block := range sanitized {
		res, blockErr := k.processSimBlock(
			ctx, cfg, sdb, base, headers, bi, block, rules, opts, gasCap,
		)
		if blockErr != nil {
			return nil, blockErr
		}
		results = append(results, res)
	}

	return results, nil
}

// processSimBlock executes one simulated block against the shared
// StateDB. It applies the block's StateOverrides (incl.
// MovePrecompileTo), builds the per-block BlockContext with the
// simulate-aware GetHashFn, runs the block's calls sequentially with
// cumulative gas accounting and per-call ephemeral resets, and
// assembles the response envelope.
//
// The header at headers[bi] is mutated in place: assembleSimBlock
// patches GasUsed before computing the block hash. Later blocks see
// this block through headers[:laterIdx] via newSimGetHashFn.
func (k *Keeper) processSimBlock(
	ctx sdk.Context,
	cfg *statedb.EVMConfig,
	sdb *statedb.StateDB,
	base *ethtypes.Header,
	headers []*ethtypes.Header,
	bi int,
	block types.SimBlock,
	rules params.Rules,
	opts *types.SimOpts,
	gasCap uint64,
) (*types.SimBlockResult, error) {
	header := headers[bi]

	var moves map[common.Address]common.Address
	if len(block.StateOverrides) > 0 {
		m, applyErr := applyStateOverrides(sdb, block.StateOverrides, rules)
		if applyErr != nil {
			return nil, applyErr
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
		// finalized past siblings (headers[:bi]). Canonical-range hashes
		// are therefore unforgeable by any BlockOverrides field.
		GetHash:     newSimGetHashFn(k.GetHashFn(ctx), base, headers[:bi]),
		Coinbase:    header.Coinbase,
		GasLimit:    header.GasLimit,
		BlockNumber: new(big.Int).Set(header.Number),
		Time:        header.Time,
		Difficulty:  new(big.Int).Set(header.Difficulty),
		BaseFee:     header.BaseFee,
		BlobBaseFee: big.NewInt(0),
		Random:      random,
	}

	precompiles := k.precompilesWithMoves(ctx, cfg, moves)

	// validation=false preserves the spec-compliant relaxation (no
	// base-fee checks, caller may lack funds); validation=true forces
	// the realistic path.
	noBaseFee := !opts.Validation
	evmOverrides := &EVMOverrides{
		BlockContext: &blockCtx,
		Precompiles:  precompiles,
		NoBaseFee:    &noBaseFee,
	}

	calls := make([]types.SimCallResult, 0, len(block.Calls))
	txHashes := make([]common.Hash, 0, len(block.Calls))
	var cumGas uint64

	for i := range block.Calls {
		// Clear per-call ephemerals (logs, refund, transient storage,
		// precompile call counter) while preserving account/storage
		// mutations. Runs unconditionally at the top of every call —
		// idempotent on a fresh StateDB, covers both call boundaries
		// within a block and block boundaries (call 0 of block N+1).
		sdb.FinaliseBetweenCalls()

		args := block.Calls[i]

		if simErr := sanitizeSimCall(sdb, &args, header, cumGas); simErr != nil {
			calls = append(calls, types.SimCallResult{
				Logs:  []*ethtypes.Log{},
				Error: simErr,
			})
			continue
		}

		msg, msgErr := args.ToMessage(gasCap, header.BaseFee)
		if msgErr != nil {
			calls = append(calls, types.SimCallResult{
				Logs:  []*ethtypes.Log{},
				Error: types.NewSimInvalidParams(msgErr.Error()),
			})
			continue
		}

		// Per-call TxConfig so AddLog stamps distinct TxHash / TxIndex
		// on every emitted log. BlockHash is left zero here because
		// the block hash depends on cumulative GasUsed, which isn't
		// known until every call in this block has executed; the
		// post-pass below patches log.BlockHash once that hash is
		// computed.
		callTxHash := computeSimTxHash(msg)
		callCfg := statedb.NewTxConfig(common.Hash{}, callTxHash, uint(i), 0) //nolint:gosec
		sdb.SetTxConfig(callCfg)

		res, _, runErr := k.applyMessageWithConfig(
			ctx,
			WrapMessage(msg),
			nil,   // tracer
			false, // commit = false (ephemeral simulate)
			cfg,
			callCfg,
			sdb,
			evmOverrides,
		)
		if runErr != nil {
			return nil, runErr
		}

		calls = append(calls, types.BuildSimCallResult(res))
		txHashes = append(txHashes, callTxHash)
		cumGas += res.GasUsed
	}

	// Finalize the header (GasUsed must be set before Hash()) and
	// back-stamp every log's BlockHash with the resulting hash so
	// consumers can join per-call logs against the per-block envelope.
	// assembleSimBlock re-patches GasUsed idempotently below.
	header.GasUsed = cumGas
	finalBlockHash := header.Hash()
	for i := range calls {
		for _, log := range calls[i].Logs {
			log.BlockHash = finalBlockHash
		}
	}

	return &types.SimBlockResult{
		Block: assembleSimBlock(header, txHashes, cumGas),
		Calls: calls,
	}, nil
}

// nonceSource narrows *statedb.StateDB down to the single method
// sanitizeSimCall consumes. Keeping the function signature interface-
// typed lets helper unit tests pass a fake without instantiating a
// full keeper.
type nonceSource interface {
	GetNonce(common.Address) uint64
}

// sanitizeSimCall defaults the nonce from the shared StateDB (so
// multi-call chains implicitly advance for CREATE callers — note that
// applyMessageWithConfig only bumps the StateDB nonce on contract
// creation, so back-to-back non-CREATE calls from the same sender read
// the same default nonce; harmless for simulate, which has no
// uniqueness check) and the gas limit from the remaining per-block gas
// budget. A requested gas that would push the cumulative block gas
// past header.GasLimit is rejected with -38015 — emitted as a per-call
// error so preceding valid calls still surface.
//
// header.GasLimit is treated as authoritative: a caller who sets
// blockOverrides.gasLimit to 0 (or otherwise produces a zero-limit
// block) gets a block where every call either defaults to Gas=0 and
// fails on intrinsic gas inside applyMessageWithConfig, or — if Gas is
// explicitly >0 — trips the -38015 preflight. This matches geth's
// behavior (empty core.GasPool) and keeps the per-block work bound
// real even under hostile BlockOverrides.
func sanitizeSimCall(
	sdb nonceSource,
	args *types.TransactionArgs,
	header *ethtypes.Header,
	cumGasUsed uint64,
) *types.SimError {
	if args.Nonce == nil {
		n := hexutil.Uint64(sdb.GetNonce(args.GetFrom()))
		args.Nonce = &n
	}

	var remaining uint64
	if header.GasLimit > cumGasUsed {
		remaining = header.GasLimit - cumGasUsed
	}

	if args.Gas == nil {
		g := hexutil.Uint64(remaining)
		args.Gas = &g
		return nil
	}

	requested := uint64(*args.Gas)
	if requested > remaining {
		return types.NewSimBlockGasLimitReached(requested, remaining)
	}
	return nil
}

// newSimGetHashFn builds the simulate-aware BLOCKHASH resolver. The
// closure captures base + the already-finalized past-sibling headers
// (headers[:bi] from the caller). Resolution order:
//
//   - height <= base.Number → delegate to canonical. Callers supply
//     k.GetHashFn(ctx), whose ctx is anchored at base by the gRPC layer
//     (rpctypes.ContextWithHeight): Case 1 returns ctx.HeaderHash
//     (canonical CometBFT hash); Case 2 consults
//     stakingKeeper.GetHistoricalInfo.
//   - height  > base.Number → look up the simulated sibling by height;
//     return header.Hash() on hit.
//   - not found → zero hash.
//
// Canonical-range hashes (height <= base.Number) are thus unforgeable
// by any BlockOverrides field: sim[] only holds simulated blocks whose
// Number > base.Number (enforced by sanitizeSimChain). This diverges
// from go-ethereum's simulate.go (which uses base.Hash() for the
// height==base case) because mezod surfaces CometBFT header hashes as
// canonical block hashes, not ethtypes.Header.Hash() values.
//
// The function takes a canonical vm.GetHashFunc (rather than a Keeper +
// ctx) so unit tests can exercise all four resolution cases against a
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
// params.Rules carries a *big.Int ChainID that prevents direct struct
// equality; we ignore ChainID (unchanged across blocks of one chain)
// and compare every boolean activation field exhaustively so the
// "bitmap" framing remains true even when go-ethereum adds new
// rule flags.
func sameForks(a, b params.Rules) bool {
	return a.IsHomestead == b.IsHomestead &&
		a.IsEIP150 == b.IsEIP150 &&
		a.IsEIP155 == b.IsEIP155 &&
		a.IsEIP158 == b.IsEIP158 &&
		a.IsEIP2929 == b.IsEIP2929 &&
		a.IsEIP4762 == b.IsEIP4762 &&
		a.IsByzantium == b.IsByzantium &&
		a.IsConstantinople == b.IsConstantinople &&
		a.IsPetersburg == b.IsPetersburg &&
		a.IsIstanbul == b.IsIstanbul &&
		a.IsBerlin == b.IsBerlin &&
		a.IsLondon == b.IsLondon &&
		a.IsMerge == b.IsMerge &&
		a.IsShanghai == b.IsShanghai &&
		a.IsCancun == b.IsCancun &&
		a.IsPrague == b.IsPrague &&
		a.IsVerkle == b.IsVerkle
}

// computeSimTxHash derives a deterministic "tx hash" for a simulated
// call. Simulate txs are unsigned and have no canonical Ethereum hash;
// we synthesize one from the message fields so the assembled block
// carries stable, distinct identifiers per call.
func computeSimTxHash(msg core.Message) common.Hash {
	tx := ethtypes.NewTx(&ethtypes.LegacyTx{
		Nonce:    msg.Nonce,
		GasPrice: msg.GasPrice,
		Gas:      msg.GasLimit,
		To:       msg.To,
		Value:    msg.Value,
		Data:     msg.Data,
	})
	return tx.Hash()
}
