> **Disclaimer:** This is a temporary file used during implementation of
> MEZO-4227 (`eth_simulateV1`). It should be removed once the feature is
> complete.

# eth_simulateV1 in go-ethereum — decomposition

Source: go-ethereum v1.16.9 (commit `95665d57`). All file/line references point to that checkout.

## 1. What it is (one-sentence summary)

`eth_simulateV1` lets an RPC client run **a sequence of transactions packed into a sequence of simulated blocks** on top of a chosen base state, with per-block *state* and *block-header* overrides applied before execution, and get back fully-assembled blocks + per-call execution results — all without touching the real chain state. It's an EIP-7572/"multicall v3-style" JSON-RPC endpoint introduced by Nethermind/Besu/Geth and standardised as part of the `eth_simulate` family.

## 2. JSON-RPC surface (entry point)

**File:** `internal/ethapi/api.go:808`

```go
func (api *BlockChainAPI) SimulateV1(
    ctx context.Context,
    opts simOpts,
    blockNrOrHash *rpc.BlockNumberOrHash,
) ([]*simBlockResult, error)
```

And JS binding at `internal/web3ext/web3ext.go:594-598` exposes it as `web3.eth.simulateV1(opts, blockTag)` with `params: 2`.

The handler's body (api.go:809–837) does exactly three things:

1. **Input guards.**
   - `len(opts.BlockStateCalls) == 0` → `invalidParamsError("empty input")` (JSON-RPC code `-32602`).
   - `len(...) > maxSimulateBlocks` (=256, simulate.go:44) → `clientLimitExceededError("too many blocks")` (code `-38026`).
   - Default `blockNrOrHash` to `latest` if nil.
2. **Snapshot the base state + header** via `api.b.StateAndHeaderByNumberOrHash(ctx, *blockNrOrHash)`. This is an in-memory `*state.StateDB` — all overrides and tx executions mutate this copy; the live DB is never touched.
3. **Build a `*simulator`** holding everything execution needs (`state`, `base`, `chainConfig`, a per-request gas pool capped at `RPCGasCap`, and the three bool switches from `simOpts`), then call `sim.execute(ctx, opts.BlockStateCalls)`.

The `RPCGasCap` is the *total* gas budget for **all** txs in **all** simulated blocks combined (api.go:822–825, comment on 831). A value of 0 means "unbounded" (`math.MaxUint64`).

**Alternative considered:** Parity/OpenEthereum used `trace_callMany`, and Erigon's `eth_callMany` is similar but flatter (single list of calls). `eth_simulateV1`'s differentiator is the *block* abstraction — gas accounting, fees, BLOCKHASH, and receipts behave as if the txs were mined in real blocks.

## 3. Input / output data model

**File:** `internal/ethapi/simulate.go:50-109`

```go
type simOpts struct {
    BlockStateCalls        []simBlock
    TraceTransfers         bool
    Validation             bool
    ReturnFullTransactions bool
}

type simBlock struct {
    BlockOverrides *override.BlockOverrides   // header field overrides
    StateOverrides *override.StateOverride    // account/state overrides
    Calls          []TransactionArgs          // the txs
}

type simCallResult struct {
    ReturnValue hexutil.Bytes  `json:"returnData"`
    Logs        []*types.Log   `json:"logs"`
    GasUsed     hexutil.Uint64 `json:"gasUsed"`
    Status      hexutil.Uint64 `json:"status"`
    Error       *callError     `json:"error,omitempty"`
}

type simBlockResult struct {
    fullTx      bool
    chainConfig *params.ChainConfig
    Block       *types.Block
    Calls       []simCallResult
    senders     map[common.Hash]common.Address
}
```

Two small but important serialization quirks in `MarshalJSON` (simulate.go:66, 85):

- `simCallResult` forces `Logs` to `[]` (not `null`) when empty — a common JSON-RPC compatibility detail.
- `simBlockResult` re-uses `RPCMarshalBlock` (the standard block encoder), then injects a `calls` field, and **when `returnFullTransactions=true`** patches the `from` field of each tx using the `senders` map. That patching is needed because the tx objects produced during simulation are *unsigned* — the sender isn't recoverable from the signature, it's remembered from `TransactionArgs.from()` (simulate.go:286).

## 4. The simulator object

**File:** `simulate.go:152–163`

```go
type simulator struct {
    b              Backend
    state          *state.StateDB
    base           *types.Header
    chainConfig    *params.ChainConfig
    gp             *core.GasPool   // shared across ALL blocks + calls
    traceTransfers bool
    validate       bool
    fullTx         bool
}
```

Key design point: **one `*state.StateDB` and one `*core.GasPool` are shared across every block and every call**. This is what makes chained simulation meaningful — writes from block 1, call 1 are visible to block 3, call 7 — and what enforces a global gas budget.

A comment on `simulator` says "not safe for concurrent use" (simulate.go:153) — each request gets its own.

## 5. The driver: `simulator.execute`

**File:** `simulate.go:166–207`

```go
func (sim *simulator) execute(ctx context.Context, blocks []simBlock) ([]*simBlockResult, error) {
    if err := ctx.Err(); err != nil { return nil, err }

    timeout := sim.b.RPCEVMTimeout()
    if timeout > 0 {
        ctx, cancel = context.WithTimeout(ctx, timeout)
    } else {
        ctx, cancel = context.WithCancel(ctx)
    }
    defer cancel()

    blocks, err := sim.sanitizeChain(blocks)       // (a)
    headers, err := sim.makeHeaders(blocks)        // (b)

    results := make([]*simBlockResult, len(blocks))
    parent  := sim.base
    for bi, block := range blocks {
        result, callResults, senders, _ :=
            sim.processBlock(ctx, &block, headers[bi], parent, headers[:bi], timeout)
        headers[bi] = result.Header()                    // repair with post-exec header
        results[bi] = &simBlockResult{...}
        parent = result.Header()
    }
    return results, nil
}
```

The driver makes three passes:

- **(a) Sanitize** — fix up / validate the chain of blocks (gap fill, ordering).
- **(b) Make preliminary headers** — needed so step (c) can build a `ChainContext` that can return headers for *future* simulated blocks (the BLOCKHASH opcode reads parent-of-parent headers).
- **(c) Execute block-by-block**, passing the parent header forward and slicing `headers[:bi]` so each block can only see its past siblings (prevents a simulated block reading the hash of a not-yet-executed simulated block).

The `RPCEVMTimeout` applies to the **whole request** (not per-block). The context is plumbed into every inner `applyMessageWithEVM` call and a goroutine there calls `evm.Cancel()` on context-done.

## 6. `sanitizeChain` — ordering + gap filling

**File:** `simulate.go:400–459`

This is one of the most nuanced pieces. Rules:

1. **Default block number**: If the user doesn't set `BlockOverrides.Number`, it becomes `prev.Number + 1`.
2. **Default timestamp**: If the user doesn't set `BlockOverrides.Time`, it becomes `prev.Time + 12` (`timestampIncrement`, simulate.go:47). 12 s matches post-merge slot cadence.
3. **Strict ordering is enforced** — both number and timestamp must *increase* (not equal). Violations return:
   - `invalidBlockNumberError` (code `-38020`): `"block numbers must be in order: N <= M"`.
   - `invalidBlockTimestampError` (code `-38021`): `"block timestamps must be in order: T <= U"`.
4. **Gap filling** — if user jumps from block 11 to block 14, blocks 12 and 13 are synthesised as empty `simBlock`s. Each filler bumps the timestamp by 12 s. This is why `TestSimulateSanitizeBlockOrder` (simulate_test.go:49–57) shows that skipping from 10 → 13 with `Time: 80` produces intermediate blocks at (11, 62), (12, 74) and only *then* (13, 80).
5. **Absolute-range cap**: total span from base can never exceed `maxSimulateBlocks` (256); exceeding fires `clientLimitExceededError` (simulate.go:422-424).
6. **Withdrawals default** — empty withdrawals list is installed if not overridden (simulate.go:415-417).

This is pure-input pre-processing: it mutates block overrides only, doesn't touch state.

## 7. `makeHeaders` — preliminary header skeletons

**File:** `simulate.go:464-508`

Walks the (sanitized) block list and produces a `*types.Header` for each, with:
- `ParentHash` is **not yet set** (will be overwritten in `processBlock:212`).
- `UncleHash = EmptyUncleHash`, `ReceiptHash = EmptyReceiptsHash`, `TxHash = EmptyTxsHash` — these get re-computed at `FinalizeAndAssemble` time.
- `Coinbase` / `Difficulty` / `GasLimit` default-inherit from **the previous simulated header** (which chains back to `sim.base`).
- `WithdrawalsHash` is set to `EmptyWithdrawalsHash` when Shanghai is active.
- `ParentBeaconRoot` is initialised to zero for Cancun (or the override, if set). There's an interesting gate at simulate.go:485–488: `BeaconRoot` is only accepted *at this internal level* — but `BlockOverrides.Apply` itself rejects it for the RPC surface (override.go:141-143). So in practice `BeaconRoot` override isn't user-reachable, only `MakeHeader` uses it internally.
- `Difficulty` is forcibly zeroed when the chain is post-merge (simulate.go:492-494). A comment on this exact line explains *why*: without it, simulating on hoodi with `blockParameter: 0x0` produces headers with difficulty 1, which would make the hardfork rules treat the simulated chain as **pre-merge** — breaking all subsequent post-merge behavior.
- Finally calls `overrides.MakeHeader(...)` (override.go:178) which re-applies every user-set field onto the scaffold.

The pass is "preliminary" because `GasUsed`, `BlobGasUsed`, `RequestsHash`, and `BaseFee` are filled in later, post-execution.

## 8. `processBlock` — the heart of a simulated block

**File:** `simulate.go:209-356`

This is where a single block is actually executed. It's long, so I'll walk it in phases.

### 8a. Parent-dependent header fields (lines 212–231)

- `header.ParentHash = parent.Hash()` — now that the previous block is finalized, we can link.
- **EIP-1559 base fee**: if London is active and the user didn't set `BaseFee`:
  - `validate=true` → compute `eip1559.CalcBaseFee(cfg, parent)` (the real protocol rule).
  - `validate=false` → **set `BaseFee = 0`**. This is the critical "simulate" vs "validate" switch: the comment at simulate.go:220 explains that without it you'd hit `gasPrice < baseFee` for any call that specifies no fee.
- **EIP-4844 excess blob gas**: computed via `eip4844.CalcExcessBlobGas` when both parent and current are Cancun.

### 8b. Block context + precompile set (lines 232–239)

- `blockContext = core.NewEVMBlockContext(header, sim.newSimulatedChainContext(...), nil)`.
- The *chain context* is the interesting part: `newSimulatedChainContext` wraps a `simBackend` (simulate.go:510–563) that *hybridises* the real chain and the simulated headers. When the EVM executes `BLOCKHASH`:
  - Canonical blocks below base → delegate to real backend (`b.HeaderByNumber`).
  - The base block itself → return `sim.base`.
  - A previously-simulated sibling → scan the `headers` slice.
- `BlobBaseFee` from the block override is applied directly to `blockContext` (not just the header).
- `activePrecompiles(sim.base)` (simulate.go:388) derives the precompile set from **base block's rules** — important: even if block-number override says "Cancun", the active precompile list is whatever matches the base block's chain rules. This is then passed to `StateOverride.Apply` so it knows which slots are precompiles for `MovePrecompileTo`.
- `block.StateOverrides.Apply(sim.state, precompiles)` — applies the state diff **before any tx runs in this block** (see §9 below).

### 8c. EVM construction with hooked tracer (lines 241–266)

```go
tracer   := newTracer(sim.traceTransfers, blockContext.BlockNumber.Uint64(), ...)
vmConfig := &vm.Config{
    NoBaseFee: !sim.validate,
    Tracer:    tracer.Hooks(),
}
tracingStateDB := vm.StateDB(sim.state)
if hooks := tracer.Hooks(); hooks != nil {
    tracingStateDB = state.NewHookedState(sim.state, hooks)
}
evm := vm.NewEVM(blockContext, tracingStateDB, sim.chainConfig, *vmConfig)
if precompiles != nil { evm.SetPrecompiles(precompiles) }
```

Points:
- `NoBaseFee: !sim.validate` complements the base-fee-0 trick above. With `NoBaseFee` the EVM won't fail the `gasFeeCap >= baseFee` pre-check.
- The state is wrapped in `state.NewHookedState` so the tracer's `OnLog`, `OnBalanceChange`, etc. hooks fire — crucial for `traceTransfers`.
- `evm.SetPrecompiles(precompiles)` is what actually installs the moved/removed precompiles we rebuilt in §8b.

### 8d. EIP-4788 + EIP-2935 system contracts (lines 267–272)

```go
if sim.chainConfig.IsPrague(...) || sim.chainConfig.IsVerkle(...) {
    core.ProcessParentBlockHash(header.ParentHash, evm)   // EIP-2935
}
if header.ParentBeaconRoot != nil {
    core.ProcessBeaconBlockRoot(*header.ParentBeaconRoot, evm) // EIP-4788
}
```

These run the pseudo-"system transactions" that real block execution also runs at the top of the block.

### 8e. The per-call inner loop (lines 274–322)

For each call:
1. `ctx.Err()` check (bails on timeout/cancel).
2. `sim.sanitizeCall(&call, ...)` (see §10 below) — defaults nonce, gas, enforces block gas limit.
3. Build a `*types.Transaction` via `call.ToTransaction(DynamicFeeTxType)` **only to get a hash and for the assembled block** (the tx is unsigned). The hash becomes `txHash := tx.Hash()`.
4. Record the sender: `senders[txHash] = call.from()` — necessary because an unsigned tx has no recoverable sender.
5. `tracer.reset(txHash, i)` — clear call-frame log buffer for this tx, set tx-level metadata.
6. `sim.state.SetTxContext(txHash, i)` — for state logging / snapshots.
7. Convert to a message: `msg := call.ToMessage(header.BaseFee, !sim.validate)`. The second arg is `skipNonceCheck` — **nonce is only validated when `validate=true`**. EoA check is always skipped (comment simulate.go:289; this is done via `SkipTransactionChecks: true` always set in `ToMessage`, transaction_args.go:494).
8. `applyMessageWithEVM(ctx, evm, msg, timeout, sim.gp)` — the actual execution. It spawns a goroutine to `evm.Cancel()` on ctx-done, then calls `core.ApplyMessage(evm, msg, gp)`. On evm-cancellation it returns `"execution aborted (timeout = X)"`.
9. If the tx returns a non-revert error (`ErrNonceTooHigh/Low`, `ErrIntrinsicGas`, `ErrInsufficientFunds`, `ErrSenderNoEOA`, `ErrMaxInitCodeSizeExceeded`, …), `txValidationError(err)` wraps it into an `invalidTxError` with structured code (errors.go:120). **This aborts the whole simulation** — a single tx-level failure is fatal for the request (important contrast with a "soft" failure like a revert, which is kept per-call).
10. Update state root semantics:
    - Byzantium+ → `tracingStateDB.Finalise(true)` (the "empty root" behaviour).
    - Pre-Byzantium → compute `IntermediateRoot(IsEIP158(..))` and keep bytes as the receipt's post-state root.
11. `gasUsed += result.UsedGas` — block-level running gas total.
12. Build a receipt with `core.MakeReceipt(evm, result, state, blockNumber, common.Hash{}, time, tx, gasUsed, root)` — the empty block hash is a placeholder, repaired at the end.
13. Aggregate `blobGasUsed`.
14. Produce the `simCallResult`:
    - On revert: parse reason (`vm.ErrExecutionReverted`), attach `{code: -32000, data: "0x…"}` — that's `errCodeReverted`.
    - On other VM error: `{code: -32015}` — `errCodeVMError`.
    - On success: append tx logs to the block's `allLogs` (used only for EIP-7685 post-processing below).
15. Save into `callResults[i]`.

### 8f. Finalize block-level totals + EIP-7685 requests (lines 323–347)

```go
header.GasUsed = gasUsed
if Cancun { header.BlobGasUsed = &blobGasUsed }

if Prague {
    requests := [][]byte{}
    core.ParseDepositLogs(&requests, allLogs, sim.chainConfig)   // EIP-6110
    core.ProcessWithdrawalQueue(&requests, evm)                  // EIP-7002
    core.ProcessConsolidationQueue(&requests, evm)               // EIP-7251
    reqHash := types.CalcRequestsHash(requests)
    header.RequestsHash = &reqHash
}
```

These three calls invoke the Prague-era "post-block" system contracts. Note that `ProcessWithdrawalQueue` and `ProcessConsolidationQueue` take the live `evm` — meaning they execute state changes after all user txs, exactly as in real block execution.

### 8g. Engine finalization (lines 348–355)

```go
blockBody       := &types.Body{Transactions: txes, Withdrawals: *block.BlockOverrides.Withdrawals}
chainHeadReader := &simChainHeadReader{ctx, sim.b}
b, err          := sim.b.Engine().FinalizeAndAssemble(chainHeadReader, header, sim.state, blockBody, receipts)
repairLogs(callResults, b.Hash())
return b, callResults, senders, nil
```

The consensus engine is asked to build a real, fully-hashed `*types.Block` (computes `TxHash`, `ReceiptHash`, `WithdrawalsHash`, and the final block hash). This requires a `ChainHeaderReader`, which is what `simChainHeadReader` (simulate.go:112–150) provides — a thin adapter that serves the simulator's `Backend` as the header source (it's separate from `simBackend` because `FinalizeAndAssemble` expects a different interface).

### 8h. `repairLogs` — patching block hashes post-hoc

**File:** `simulate.go:361-367`

During execution, `tracer.captureLog` stamps `BlockHash` into each log — but the block hash isn't known yet. After `FinalizeAndAssemble` gives us `b.Hash()`, `repairLogs` walks every captured log and sets `log.BlockHash = hash`. This matches the behaviour clients expect from `eth_getLogs` (where `blockHash` is the canonical block hash).

## 9. State overrides — `override.StateOverride.Apply`

**File:** `internal/ethapi/override/override.go:57–120`

Structure:

```go
type OverrideAccount struct {
    Nonce            *hexutil.Uint64
    Code             *hexutil.Bytes
    Balance          *hexutil.Big
    State            map[common.Hash]common.Hash
    StateDiff        map[common.Hash]common.Hash
    MovePrecompileTo *common.Address
}
```

Ordered semantics (important for correctness):

1. **MovePrecompileTo** is checked first. It lets you *relocate* a precompile — say, move `SHA256` (0x02) to any address, then overwrite 0x02 with arbitrary bytecode. Rules (override.go:73–83):
   - You can only `MovePrecompileTo` from an address that **currently is** a precompile — otherwise `"account X is not a precompile"`.
   - The destination must not also be overridden in the same request (`"already overridden"`).
   - Destinations of prior moves are tracked in `dirtyAddrs` — a subsequent override that targets the same address errors out.
   - The move just does `precompiles[dst] = p` + `delete(precompiles, addr)`. The *state* at the destination is **not cleared** — see the code comment at override.go:71.
   - This is the feature tested by api_test.go:1839 and 3854; it's what makes things like "what if the SHA256 precompile were implemented in EVM bytecode?" or "what would happen if this address were a precompile?" answerable.
2. **Nonce** — `SetNonce(addr, nonce, tracing.NonceChangeUnspecified)`.
3. **Code** — `SetCode(addr, code, tracing.CodeChangeUnspecified)`.
4. **Balance** — `SetBalance(addr, uint256Bal, tracing.BalanceChangeUnspecified)`.
5. **State vs StateDiff** — *mutually exclusive* (override.go:101-103):
   - `state` — `SetStorage(addr, state)` **replaces the entire storage** of the account.
   - `stateDiff` — `SetState(addr, key, value)` for each — **merges** the specified slots.
6. Finally `statedb.Finalise(false)` — the comment at override.go:115-117 is interesting: the override is committed as though it happened in a phantom transaction immediately before the simulation's first real call.

One trade-off here: `MovePrecompileTo` is a go-ethereum-specific feature that makes the RPC more expressive than Parity-style `eth_call` overrides. It's more powerful but also makes caching / replaying at other clients harder.

## 10. Block overrides — `override.BlockOverrides`

**File:** `override.go:122–205`

Allowed fields:

| Field           | Effect                                                                                                                   |
| --------------- | ------------------------------------------------------------------------------------------------------------------------ |
| `Number`        | Overrides block number. Also used by simulate to fill gaps.                                                              |
| `Difficulty`    | No-op post-merge (see §7).                                                                                               |
| `Time`          | Block timestamp. Must strictly increase.                                                                                 |
| `GasLimit`      | Block gas limit.                                                                                                         |
| `FeeRecipient`  | `Coinbase` address.                                                                                                      |
| `PrevRandao`    | `MixDigest` / `RANDAO` (post-merge).                                                                                     |
| `BaseFeePerGas` | EIP-1559 base fee; crucial for the `validate=true` path.                                                                 |
| `BlobBaseFee`   | EIP-4844 blob base fee. Applied to `blockContext` but **not** into the header (`MakeHeader` ignores it — no such field). |
| `BeaconRoot`    | Explicitly **rejected** in `Apply` for this RPC (`errors.New("... not supported for this RPC method")`).                 |
| `Withdrawals`   | Same — rejected in `Apply`. Defaulted to empty elsewhere.                                                                |

Two entry points exist: `Apply(*vm.BlockContext)` (runtime EVM context) and `MakeHeader(*types.Header)` (the preliminary header scaffold in §7). They intentionally differ on `BlobBaseFee` and the rejected fields.

## 11. `sanitizeCall` — per-call defaulting + block-gas-limit enforcement

**File:** `simulate.go:369-386`

```go
if call.Nonce == nil {
    nonce := state.GetNonce(call.from())
    call.Nonce = (*hexutil.Uint64)(&nonce)        // read live nonce
}
if call.Gas == nil {
    remaining := blockContext.GasLimit - *gasUsed
    call.Gas   = (*hexutil.Uint64)(&remaining)    // "as much gas as remains"
}
if *gasUsed + *call.Gas > blockContext.GasLimit {
    return &blockGasLimitReachedError{...}        // code -38015
}
call.CallDefaults(sim.gp.Gas(), header.BaseFee, chainConfig.ChainID)
```

Two sensible defaults: a missing nonce is drawn from **current state** (which reflects all prior simulated txs), and a missing gas limit defaults to "whatever's left in this block". The block gas limit check is what returns `errCodeBlockGasLimitReached` (`-38015`).

`TransactionArgs.CallDefaults` (transaction_args.go:391-440) then does the ordinary "fill in zero for missing fee fields, validate chainID, cap gas at `globalGasCap`" logic. Important: with `baseFee` non-nil this forces the call into **1559-mode** (zeroing `maxFeePerGas` / `maxPriorityFeePerGas` when not set).

## 12. `TransactionArgs.ToMessage` / `ToTransaction`

**File:** `transaction_args.go:446-496` and `500+`

`ToMessage` produces a `*core.Message` (EVM-level input). Two crucial flags:

```go
SkipNonceChecks:       skipNonceCheck,   // = !sim.validate
SkipTransactionChecks: true,             // always true -> EoA check off
```

`ToTransaction(DynamicFeeTxType)` produces an **unsigned** `*types.Transaction` with the correct *shape* (1559 / blob / SetCode / access-list), only used to compute a stable hash and to assemble the block body. Because it's unsigned, `types.Sender(...)` can't recover the sender — which is the whole reason `simBlockResult.senders` exists.

## 13. Validation mode: `validate=true` vs `validate=false`

This is the single most important operational switch. Summary of what flips:

| Behaviour                                             | `validate=false` (default) | `validate=true`          |
| ----------------------------------------------------- | -------------------------- | ------------------------ |
| Base fee when not overridden                          | `0`                        | `CalcBaseFee(cfg, parent)` |
| `vm.Config.NoBaseFee`                                 | `true`                     | `false`                  |
| `msg.SkipNonceChecks` (via `ToMessage` second arg)    | `true`                     | `false`                  |
| EoA check                                             | off                        | off (`SkipTransactionChecks: true`) |
| Sender balance vs. fees                               | Effectively bypassed (fees=0) | Real balance check    |

In practice: `validate=false` is the "simulate this as if I'm a rich debug user" mode; `validate=true` is the "would this actually be accepted and executed on-chain right now?" mode.

## 14. Transfer tracing (`logtracer.go`)

**File:** `internal/ethapi/logtracer.go`

When `traceTransfers=true`, every native ETH transfer (txn value, call value, SELFDESTRUCT) is **synthesised as an ERC-20-shaped `Transfer` log**:

```go
transferTopic   = keccak256("Transfer(address,address,uint256)")       // ddf252ad...
transferAddress = 0xEeeeeEeeeEeEeeEeEeEeeEEEeeeeEeeeeeeeEEeE           // ERC-7528
```

Pseudo-log shape:
- `address`: `0xEeee…EEeE` (ERC-7528's "native token pseudo-address").
- `topics[0]`: the ERC-20 Transfer signature.
- `topics[1..2]`: `from`, `to` padded to 32 bytes.
- `data`: 32-byte value.

Tracer hooks (logtracer.go:73-79): `OnEnter`, `OnExit`, `OnLog`.
- `OnEnter`: pushes a new per-frame log slice. If the call *opcode isn't DELEGATECALL* and `value > 0`, emit a synthetic transfer log. DELEGATECALL is excluded because it doesn't move value — that's the only distinguishing rule.
- `OnExit`: pops the frame's logs. If the frame **reverted**, they're dropped; otherwise they're merged into the parent frame's buffer. This naturally mirrors the EVM's own log-on-revert behaviour.
- `OnLog` (real ERC-20 events from contracts): captured verbatim.
- `reset(txHash, idx)`: clears the stack between txs in the same block.

The side effect is that from the client's point of view, `eth_simulateV1` + `traceTransfers` gives you a unified "flow of value" feed — real ERC-20 Transfer events and native-ETH pseudo-events in one chronologically-ordered log list.

Trade-off: this is **not** a real execution trace (no stack/memory/storage ops exposed). If you want a proper trace, you still need `debug_traceCall` etc. `simulateV1` intentionally targets the "what happened" granularity that UIs need, not the "why".

## 15. Chain context shim: `simBackend`

**File:** `simulate.go:510-563`

```go
type simBackend struct {
    b       ChainContextBackend
    base    *types.Header
    headers []*types.Header
}
```

Implements the `ChainContextBackend` interface and is wrapped inside `NewChainContext` (api.go:653). It's what makes `BLOCKHASH` work correctly across simulated blocks — verified by `TestSimulateV1ChainLinkage` (api_test.go:2466), which asserts that in block 3 a contract calling `BLOCKHASH(1)` and `BLOCKHASH(2)` returns the hashes of the *simulated* blocks 1 and 2.

Resolution order in `HeaderByNumber`:
1. Requested == base block → return base.
2. Requested < base → delegate to the real backend (canonical chain).
3. Requested > base → linear scan `headers[]` of already-executed simulated blocks.

Point (3) works because `execute()` passes `headers[:bi]` — so block `bi` can only see *past* simulated blocks, and not future ones.

## 16. Error model

**File:** `internal/ethapi/errors.go:102-170`

Simulate uses a rich set of structured JSON-RPC codes (mostly in the -38xxx range reserved for simulate-style APIs):

| Code   | Name                              | Meaning                                              |
| ------ | --------------------------------- | ---------------------------------------------------- |
| -32602 | `errCodeInvalidParams`            | Empty or malformed input                             |
| -38026 | `errCodeClientLimitExceeded`      | > 256 blocks                                         |
| -38020 | `errCodeBlockNumberInvalid`       | Non-increasing block number                          |
| -38021 | `errCodeBlockTimestampInvalid`    | Non-increasing timestamp                             |
| -38015 | `errCodeBlockGasLimitReached`     | Cumulative gas > `header.GasLimit`                   |
| -38010 | `errCodeNonceTooLow`              | `validate=true` and nonce too low                    |
| -38011 | `errCodeNonceTooHigh`             | `validate=true` and nonce too high                   |
| -38013 | `errCodeIntrinsicGas`             | Below intrinsic gas                                  |
| -38014 | `errCodeInsufficientFunds`        | Sender can't pay `gasLimit*price + value`            |
| -38024 | `errCodeSenderIsNotEOA`           | (skipped in simulate; still defined)                 |
| -38025 | `errCodeMaxInitCodeSizeExceeded`  | Init code size exceeded                              |
| -32000 | `errCodeReverted`                 | **Per-call** revert (attached to `simCallResult.Error`, not the top-level err) |
| -32015 | `errCodeVMError`                  | Per-call non-revert VM error                         |

The key distinction: **tx-level validation failures kill the whole simulation** (via `txValidationError` at simulate.go:293-294). **Execution-time failures (revert, invalid opcode, OOG)** are reported per-call inside `simCallResult.Error` and the simulation continues (simulate.go:308-316).

## 17. Putting it together — lifecycle of one `eth_simulateV1` call

```
client sends eth_simulateV1(opts, blockTag)
 └─> BlockChainAPI.SimulateV1
      ├─ validate opts (empty? >256 blocks?)
      ├─ StateAndHeaderByNumberOrHash(blockTag)     [base snapshot, in-memory]
      ├─ build *simulator
      └─ simulator.execute
           ├─ setup timeout ctx
           ├─ sanitizeChain    (fill gaps, enforce ordering)
           ├─ makeHeaders      (preliminary scaffolds, post-merge diff=0, Cancun fields)
           └─ for each block:
                ├─ parent-linked header fields (ParentHash, BaseFee, ExcessBlobGas)
                ├─ build blockContext + simBackend-backed ChainContext
                ├─ resolve precompile set (incl. MovePrecompileTo mutations)
                ├─ StateOverride.Apply(state, precompiles)     <-- pre-tx state mutation
                ├─ construct EVM w/ hooked state + tracer
                ├─ EIP-2935 parent-hash sys-call  (Prague)
                ├─ EIP-4788 beacon-root sys-call  (Cancun)
                ├─ for each call:
                │    ├─ sanitizeCall (nonce/gas defaults; block-gas-limit check)
                │    ├─ ToTransaction to get hash; record sender
                │    ├─ ToMessage(skipNonce=!validate); applyMessageWithEVM
                │    ├─ Finalise state (post-Byzantium) or IntermediateRoot
                │    ├─ MakeReceipt; accumulate gas / blob gas
                │    └─ assemble simCallResult (OK/reverted/VMError)
                ├─ EIP-7685 post-block sys-calls (Prague): deposits, withdrawals Q, consolidation Q
                ├─ set RequestsHash, GasUsed, BlobGasUsed on header
                ├─ Engine.FinalizeAndAssemble(simChainHeadReader, header, state, body, receipts)
                └─ repairLogs(callResults, block.Hash())
           └─ return results
 <─ JSON-RPC marshalling (RPCMarshalBlock + "calls" field, optional from-patching)
```

## 18. Trade-offs and alternatives worth flagging

- **In-memory StateDB, no commit** — reversible and cheap; but the *entire simulation must fit in memory*, and there's no checkpointing within a request. If you want per-block rollback, you'd have to run multiple requests.
- **One shared gas pool across blocks** — matches intuition ("simulate as though these were mined") but means a run-away tx in block 1 can starve block 256. `RPCGasCap=0` disables the cap (`math.MaxUint64`) at the cost of unbounded compute.
- **`validate=false` by default** — convenient for UX (UIs don't need to fund fake accounts) but means simulation results can diverge from what a miner would accept. Clients must opt in to realism.
- **`MovePrecompileTo`** — powerful but geth-specific. Portability trade-off vs other clients.
- **Fatal tx-validation vs soft revert** — the choice to abort on any `core.ErrXxx` instead of reporting it per-call makes output cleaner but loses information: you can't find out *which* tx's nonce/funds were wrong without binary-searching.
- **Synthetic transfers via ERC-7528 address** — elegant, but consumers must special-case `address==0xEeee…EEeE` (and note it collides with any real contract living at that address — extremely unlikely but worth documenting).
- **Alternatives**: `eth_callMany` (Erigon; flatter, no block modelling), `debug_traceCall{Many}` (full traces, heavier), Tenderly-style external simulators (richer UX, but off-chain and central).

## 19. DoS protection / operator limits

Four stacked guards bound any single `eth_simulateV1` request. You cannot pass unbounded input.

### 19a. Hard cap on blocks per request — 256

`maxSimulateBlocks = 256` (simulate.go:44). Enforced twice:

- Up-front in `SimulateV1` on `len(opts.BlockStateCalls)` (api.go:811).
- Again in `sanitizeChain` against the *span* from base (simulate.go:422-424), so you also can't `number`-override your way past it.

Exceeding either returns `errCodeClientLimitExceeded` (`-38026`).

Gap-filler blocks **count against this cap** — if you pass two blocks with numbers `base+1` and `base+500`, sanitizeChain tries to synthesise 498 fillers and the span check fires.

### 19b. Global gas cap across the whole request — `RPCGasCap`

One `core.GasPool` is built from `api.b.RPCGasCap()` and shared across every block and every call (api.go:822-836, simulate.go:159). In vanilla geth this is `--rpc.gascap` (default **50,000,000**). `ApplyMessage` deducts from this pool on every tx; once drained, further calls fail with `core.ErrGasLimitReached`, which `txValidationError` turns into a **fatal request-level error** (simulate.go:293-294).

`RPCGasCap=0` disables the cap (`math.MaxUint64`). On a production node, this is the single most important DoS knob — leave it non-zero.

### 19c. Wall-clock timeout — `RPCEVMTimeout`

`sim.execute` wraps the whole request in `context.WithTimeout(ctx, sim.b.RPCEVMTimeout())` (simulate.go:172-178). Vanilla geth flag: `--rpc.evmtimeout` (default **5 s**). `applyMessageWithEVM` spawns a goroutine that calls `evm.Cancel()` on ctx-done (api.go:752-754), and the inner per-call loop checks `ctx.Err()` before each tx (simulate.go:275-277). On timeout: `"execution aborted (timeout = X)"`.

### 19d. Per-block gas limit — block-local, weaker

`sanitizeCall` rejects a call when `gasUsed + call.Gas > blockContext.GasLimit` (simulate.go:379-381, code `-38015`). But the user can override `GasLimit` via `BlockOverrides.GasLimit`, so this is **not** a hard DoS guard on its own — the real ceilings are 19b and 19c.

### 19e. Transactions per block/request

There's **no explicit tx-count limit**. You can submit as many `calls` per block as you want; you're only constrained by:

- `sum(gas used) ≤ block.GasLimit` (per-block; user-overridable)
- `sum(gas used across every block) ≤ RPCGasCap` (global)
- `total wall-clock ≤ RPCEVMTimeout` (global)

In practice 50 M gas + 5 s timeout bound a single request to a few thousand trivial (21 000-gas) transfers.

### 19f. Bounds summary under default geth config

| Axis             | Limit                                             |
| ---------------- | ------------------------------------------------- |
| Blocks           | 256                                               |
| Total gas        | `RPCGasCap` (50 M default, 0 = unbounded)         |
| Wall-clock       | `RPCEVMTimeout` (5 s default, 0 = unbounded)      |
| Txs per block    | none directly; bounded by block/global gas only   |
| Txs per request  | bounded by global gas + timeout                   |
| Per-call memory  | bounded only by the EVM's quadratic memory pricing — no simulate-specific cap |

### 19g. Non-guards worth noting

- **No request-payload size limit** in simulate itself. A 100 MB calldata blob will be read if the RPC server accepts it — that's a transport-layer concern (`--rpc.batch-request-limit`, HTTP body-size cap in the server front-end).
- **No per-IP rate limiting** in simulate. Operators typically front the endpoint with a reverse proxy for that.
- **Concurrent-request parallelism** — each request builds its own `*state.StateDB` snapshot and is single-threaded, but nothing prevents a client from issuing many requests concurrently. Connection-level limits (`--rpc.http.threads`, etc.) are the mitigation.

### 19h. Recommendation for public endpoints

If you expose `eth_simulateV1` publicly:
- Keep `--rpc.gascap` and `--rpc.evmtimeout` **non-zero** and size them for your hardware.
- Front with a reverse proxy that enforces per-IP rate limits and caps response size.
- Consider an allowlist for `traceTransfers=true` / `returnFullTransactions=true` if response-size amplification is a concern.

## Key files, one-liner recap

- `internal/ethapi/api.go:801-838` — RPC entry, input guards, simulator bootstrap.
- `internal/ethapi/simulate.go` — core types + driver + block processor + chain-shim.
- `internal/ethapi/override/override.go` — `StateOverride` / `BlockOverrides` + `MovePrecompileTo`.
- `internal/ethapi/logtracer.go` — log + synthetic-ETH-transfer tracer.
- `internal/ethapi/errors.go:102-170` — structured -38xxx JSON-RPC error codes.
- `internal/ethapi/transaction_args.go:391-496` — `CallDefaults` / `ToMessage` / `ToTransaction`.
- `internal/ethapi/simulate_test.go`, `api_test.go:1314+, 2466+, 2561+` — the behavioural contract.
