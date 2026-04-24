package keeper

import (
	"fmt"
	"math/big"

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

	// simTimestampIncrement is the default gap between sequential
	// simulated blocks when the caller omits Time overrides. Matches
	// geth.
	simTimestampIncrement = 12
)

// sanitizeSimChain validates the block ordering rules and fills gaps
// with empty blocks. It mutates the input slice's BlockOverrides
// pointers in place (allocating when nil) and returns the concatenated
// slice including any gap-fill blocks. Failures are *types.SimError
// values carrying spec-reserved JSON-RPC codes (-38020 / -38021 /
// -38026), returned via the plain error channel.
//
// Design mirrors go-ethereum's internal/ethapi/simulate.go::sanitizeChain
// (v1.15.4 source). Divergences: we never propagate a nil slice input
// (caller has already enforced len > 0 at the RPC boundary), and the
// span check against maxSimulateBlocks is performed BEFORE gap-fill
// allocation so a pathological `[{Number: base+1}, {Number:
// base+10_000_000}]` input cannot drive the driver into a 10M-header
// allocation.
func sanitizeSimChain(base *ethtypes.Header, blocks []types.SimBlock) ([]types.SimBlock, error) {
	res := make([]types.SimBlock, 0, len(blocks))
	prevNumber := new(big.Int).Set(base.Number)
	prevTimestamp := base.Time

	for _, block := range blocks {
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
	// Start from the parent's inherited fields (coinbase, difficulty,
	// gas limit) so omitted overrides follow the predecessor block.
	h := &ethtypes.Header{
		ParentHash:  parent.Hash(),
		UncleHash:   ethtypes.EmptyUncleHash,
		TxHash:      ethtypes.EmptyTxsHash,
		ReceiptHash: ethtypes.EmptyReceiptsHash,
		Coinbase:    parent.Coinbase,
		Difficulty:  new(big.Int).Set(parent.Difficulty),
		GasLimit:    parent.GasLimit,
	}

	// Post-merge: Difficulty is zero and MixDigest (PREVRANDAO) must be
	// non-nil. Mezo does not support PREVRANDAO randomness — we set
	// MixDigest to zero but keep it populated so the merge rule-switch
	// inside go-ethereum works.
	if chainCfg.MergeNetsplitBlock != nil {
		h.Difficulty = new(big.Int)
	}

	// Number is required — the sanitize step has already defaulted it.
	if overrides != nil && overrides.Number != nil {
		h.Number = overrides.Number.ToInt()
	} else {
		h.Number = new(big.Int).Add(parent.Number, big.NewInt(1))
	}

	// Time is required — sanitize defaults it.
	if overrides != nil && overrides.Time != nil {
		h.Time = uint64(*overrides.Time)
	} else {
		h.Time = parent.Time + simTimestampIncrement
	}

	if overrides != nil {
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
	}

	// Base fee: caller override wins, otherwise validation=true derives
	// via CalcBaseFee against parent; validation=false reports a zero
	// baseFee so the caller's per-call gas-price checks don't fail.
	switch {
	case overrides != nil && overrides.BaseFeePerGas != nil:
		h.BaseFee = overrides.BaseFeePerGas.ToInt()
	case validation && rules.IsLondon:
		h.BaseFee = eip1559.CalcBaseFee(chainCfg, parent)
	default:
		h.BaseFee = new(big.Int)
	}

	return h
}

// assembleSimBlock turns a simulated header plus the calls that ran
// inside it into a spec-shaped block envelope suitable for JSON
// marshaling. Fields match RPCMarshalBlock's output. The assembled
// block carries tx hashes only; `returnFullTransactions` patching is
// not wired yet.
//
// `cumulativeGasUsed` is the sum of GasUsed across successful calls.
// It patches the header's GasUsed field prior to hashing so that
// downstream consumers see a consistent header/block relation.
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

	// Current scope: single block, single call. The driver is the
	// single source of truth for these bounds — callers hit the same
	// error surface regardless of entry point.
	if len(opts.BlockStateCalls) != 1 {
		return nil, fmt.Errorf("multi-block simulate not yet implemented")
	}
	if len(opts.BlockStateCalls[0].Calls) > 1 {
		return nil, fmt.Errorf("multi-call simulate not yet implemented")
	}

	sanitized, err := sanitizeSimChain(base, opts.BlockStateCalls)
	if err != nil {
		return nil, err
	}

	// Request-wide deadline is enforced by the gRPC backend ctx. The
	// gas pool, block/call caps, and evm.Cancel plumbing are not wired
	// yet — they land alongside the kill switch.
	rules := cfg.Rules(ctx.BlockHeight(), uint64(ctx.BlockTime().Unix())) //nolint:gosec

	results := make([]*types.SimBlockResult, 0, len(sanitized))
	parent := base

	for _, block := range sanitized {
		header := makeSimHeader(parent, block.BlockOverrides, rules, cfg.ChainConfig, opts.Validation)

		// Per-block StateDB. Multi-call support will switch to a
		// shared StateDB with FinaliseBetweenCalls between calls.
		txConfig := statedb.NewEmptyTxConfig(common.BytesToHash(ctx.HeaderHash()))
		sdb := statedb.New(ctx, k, txConfig)

		var moves map[common.Address]common.Address
		if len(block.StateOverrides) > 0 {
			m, applyErr := applyStateOverrides(sdb, block.StateOverrides, rules)
			if applyErr != nil {
				return nil, applyErr
			}
			moves = m
		}

		// Gate Random on the merge fork: go-ethereum flips its rule
		// set to Paris when BlockContext.Random != nil. Pre-merge we
		// must leave it nil to preserve the legacy opcode set,
		// matching the derivation used by NewEVMWithOverrides.
		var random *common.Hash
		if cfg.ChainConfig.MergeNetsplitBlock != nil {
			random = &header.MixDigest
		}
		blockCtx := vm.BlockContext{
			CanTransfer: core.CanTransfer,
			Transfer:    core.Transfer,
			GetHash:     k.GetHashFn(ctx),
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

		// validation=false preserves the spec-compliant relaxation
		// (no base-fee checks, caller may lack funds); validation=true
		// forces the realistic path.
		noBaseFee := !opts.Validation
		evmOverrides := &EVMOverrides{
			BlockContext: &blockCtx,
			Precompiles:  precompiles,
			NoBaseFee:    &noBaseFee,
		}

		calls := make([]types.SimCallResult, 0, len(block.Calls))
		txHashes := make([]common.Hash, 0, len(block.Calls))
		var cumGas uint64

		for _, args := range block.Calls {
			// Default nonce from the simulated StateDB so callers
			// omitting `nonce` see any preceding state override (and
			// eventually preceding calls in the same block).
			if args.Nonce == nil {
				h := hexutil.Uint64(sdb.GetNonce(args.GetFrom()))
				args.Nonce = &h
			}

			msg, msgErr := args.ToMessage(gasCap, header.BaseFee)
			if msgErr != nil {
				calls = append(calls, types.SimCallResult{
					Logs:  []*ethtypes.Log{},
					Error: types.NewSimInvalidParams(msgErr.Error()),
				})
				continue
			}

			res, _, runErr := k.applyMessageWithConfig(
				ctx,
				WrapMessage(msg),
				nil,   // tracer
				false, // commit = false (ephemeral simulate)
				cfg,
				txConfig,
				sdb,
				evmOverrides,
			)
			if runErr != nil {
				return nil, runErr
			}

			calls = append(calls, types.BuildSimCallResult(res))
			txHashes = append(txHashes, computeSimTxHash(msg))
			cumGas += res.GasUsed
		}

		results = append(results, &types.SimBlockResult{
			Block: assembleSimBlock(header, txHashes, cumGas),
			Calls: calls,
		})
		parent = header
	}

	return results, nil
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
