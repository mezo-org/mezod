> **Disclaimer:** Temporary implementation file for MEZO-4227 (`eth_simulateV1`).
> Remove once the feature is complete.

# Implementation plan: `eth_simulateV1` in mezod

**Status.** Phases 1-8 shipped: [#658](https://github.com/mezo-org/mezod/pull/658) (scaffold), [#660](https://github.com/mezo-org/mezod/pull/660) (`MovePrecompileTo`), [#662](https://github.com/mezo-org/mezod/pull/662) (keeper seams + proto + single-call execution), [#664](https://github.com/mezo-org/mezod/pull/664) (multi-call + multi-block + simulated `GetHashFn`), [#665](https://github.com/mezo-org/mezod/pull/665) (DoS guards + kill switch + request-level fatal lifts). Next up: **Phase 9** (`TraceTransfers` synthetic ERC-20 logs).

## Context

Add `eth_simulateV1` to mezod's EVM JSON-RPC surface. Reference impl: go-ethereum v1.16.9 `internal/ethapi/simulate.go` (full walkthrough in `research.md`). Authoritative spec: `ethereum/execution-apis`, with 92 conformance fixtures in `tests/eth_simulateV1/`.

Mezod is a mission-critical Cosmos-SDK EVM chain (Evmos-derived, CometBFT consensus). The broader EVM tooling ecosystem (ethers v6, viem, MetaMask/Rabby, debug UIs) increasingly assumes `eth_simulateV1` is available for multi-tx multi-block simulation with overrides. No Cosmos/Evmos-family chain has shipped it yet; mezod becomes the reference.

**Security posture.** Every phase must (a) build green on its own, (b) ship targeted tests, (c) never touch consensus-critical paths. Phases 7, 8, 10, and 15 are security-critical kernels and gate a `/security-review` invocation before merge.

**Two-part delivery.** Sequenced around the separate geth v1.14.8 → v1.16.9 + Prague/Osaka [upgrade project](https://linear.app/thesis-co/project/chain-geth-v116-upgrade-and-osaka-fork-compatibility-b08591b25fb5).

- **Part 1 — Phases 1-12 (executes now against v1.14.8).** No dependency on the upgrade project; ships independently.
- **Part 2 — Phases 13-16 (post-upgrade).** Mechanical port to v1.16.9 interfaces, then add Prague/Osaka features that fit mezo's chain model.

The port cost measured against `git diff v1.14.8..v1.16.9` on the interfaces simulate touches is ~15 mechanical lines (net -10 LOC after replacing the custom `FinaliseBetweenCalls` helper with geth's new `StateDB.Finalise(true)`).

## Decisions (locked in)

| Decision | Choice |
|---|---|
| Feature parity | **Full parity** with execution-apis spec (`TraceTransfers`, `Validation`, `ReturnFullTransactions`, `MovePrecompileTo`) |
| `MovePrecompileTo` for custom mezo precompiles (`0x7b7c…`) | **Blocked** — stdlib precompiles (0x01-0x0A) only; custom rejects with structured error |
| DoS config | **Kill switch only**: `SimulateDisabled bool` on `JSONRPCConfig`; reuse existing `RPCGasCap` + `RPCEVMTimeout`; hard-code 256-block cap |
| Validation error semantics | **Spec-conformant** — tx-level validation failures (`-38010..-38025`) are fatal top-level errors; revert / VM errors stay per-call (`3` / `-32015`) |
| Error handling | **Single typed `*SimError{Code, Message, Data}` end-to-end.** Catalog in `x/evm/types/simulate_v1_errors.go`. Rides gRPC on a dedicated `SimError error = 2` field of `SimulateV1Response`. No enum, no kind-string, no translator. Genuine internals collapse to `status.Error(codes.Internal, …)` |
| Gas numerics | **mezod-native** — reported `GasUsed` honors `MinGasMultiplier` (matches on-chain receipts); raw EVM gas only for internal pool accounting |
| EIPs (Part 1, v1.14.8) | Skip EIP-4844 / 4788 / 2935 / 7685 (not in chain config); reject explicit overrides for those fields |
| EIPs (Part 2, post-upgrade) | Add EIP-2935 pre-block hook, EIP-7702 SetCode, EIP-7825 per-tx gas cap, `MaxUsedGas` field. Continue rejecting EIP-4844 / 4788 / 7685 / 6110 / 7002 / 7251 (no beacon chain, no DA layer, no EL↔CL requests) |

## Already shipped (Phases 1-8)

The architectural seam is **bare types in `x/evm/types/`, flow logic in `x/evm/keeper/`**. The driver lives in the `keeper` package because it needs unexported access to `applyStateOverrides` and `applyMessageWithConfig`. **There is no `simulate/` sub-package** — driver and helpers live in a single file, `x/evm/keeper/simulate_v1.go`. All request/response JSON shapes live under `x/evm/types/` — `rpc/types/simulate_v1.go` is gone; there is no duplicate RPC-side shape.

### File map and symbols

| File | Symbols / role |
|---|---|
| `x/evm/types/simulate_v1.go` | All JSON shapes and helpers: `SimOpts`, `SimBlock`, `SimBlockOverrides`, `SimCallResult` (+ `MarshalJSON` forcing `Logs: []` over `null`), `SimBlockResult` (+ `MarshalJSON` / `UnmarshalJSON` that flatten block fields alongside `calls`), `UnmarshalSimOpts` (strict-validates input — rejects `BlockOverrides.BeaconRoot/Withdrawals/BlobBaseFee`), `BuildSimCallResult`. Also hosts the `MaxSimulateBlocks = 256` / `MaxSimulateCalls = 1000` envelope caps and the `CountSimCalls` helper — the cumulative call count is the post-sanitize total (gap-fills included), not the raw input, so the keeper recounts after `sanitizeSimChain` returns. Single shape used by both keeper and backend; no RPC-side duplicate |
| `x/evm/types/state_overrides.go` | `StateOverride`, `OverrideAccount` (incl. `MovePrecompileTo *common.Address`) — unified override types used by both `eth_call` and `eth_simulateV1`; no RPC-side duplicate |
| `x/evm/types/simulate_v1_errors.go` | `SimError{Code, Message, Data}` implementing geth's `Error()/ErrorCode()/ErrorData()`; `SimErrCode*` constants for every spec-reserved code; `NewSim*` constructors: `NewSimInvalidParams`, `NewSimInvalidBlockNumber`, `NewSimInvalidBlockTimestamp`, `NewSimClientLimitExceeded`, `NewSimBlockCountExceeded`, `NewSimCallLimitExceeded`, `NewSimBlockGasLimitReached` (-38015), `NewSimIntrinsicGas` (-38013), `NewSimForkSpanUnsupported`, `NewSimTimeout` (-32016), `NewSimMethodNotFound` (-32601), `NewSimMovePrecompileSelfRef`, `NewSimMovePrecompileDupDest`, `NewSimStateAndStateDiff`, `NewSimAccountTainted`, `NewSimDestAlreadyOverridden`, `NewSimMoveMezoCustom`, `NewSimNotAPrecompile`, `NewSimReverted`, `NewSimVMError` |
| `x/evm/keeper/simulate_v1.go` | All driver + helpers (private): `simulateV1` (top-level entry; one shared `*statedb.StateDB` for the whole request; sanitizes, post-sanitize call cap, fork-span sentinel, request-wide `simGasBudget`), `processSimBlock` (per-block execution; StateOverrides + BlockContext + per-call loop + envelope assembly; per-block cancel-watcher goroutine that calls `evm.Cancel()` on ctx-done via `OnEVMConstructed` published `*vm.EVM`), `sanitizeSimChain` (chain ordering + gap fill + `-38020` / `-38021` / `-38026`), `resolveSimCallNonce` / `resolveSimCallGas` (per-call defaults: nonce from shared StateDB, gas from `header.GasLimit - cumGasUsed` clamped by request-wide budget; `-38015` preflight), `simGasBudget` (`newSimGasBudget(gasCap)` — `gasCap == 0` means unlimited; `clamp` / `consume`), `makeSimHeader`, `assembleSimBlock`, `computeSimTxHash`, `newSimGetHashFn` (simulate-aware `BLOCKHASH`: canonical via `k.GetHashFn(ctx)`, simulated siblings via in-memory height map, zero-hash otherwise — canonical range unforgeable by any `BlockOverrides`), `sameForks` (fork-boundary sentinel; compares base vs. last-sim `params.Rules` exhaustively and rejects spans that would cross a fork) |
| `x/evm/keeper/state_override.go` | `applyStateOverrides(db, overrides, rules) (map[addr]addr, error)` — returns the validated `MovePrecompileTo` move set; mutates `db` in place; uses `vm.DefaultPrecompiles(rules)` to identify stdlib precompiles and `mezoCustomPrecompileAddrs` for the deny-list. Returns `*types.SimError` via `NewSim*` constructors on every spec-coded failure |
| `x/evm/keeper/state_transition.go` | Seams: `EVMOverrides{BlockContext *vm.BlockContext, Precompiles map[addr]vm.PrecompiledContract, NoBaseFee *bool}`, `NewEVMWithOverrides(ctx, msg, cfg, tracer, stateDB, *EVMOverrides) *vm.EVM`, `precompilesWithMoves`, `activePrecompiles`. `applyMessageWithConfig` takes `*EVMOverrides`. `NewEVM` and the public `ApplyMessageWithConfig` delegate with `nil` — consensus path is byte-identical to main. The simulate driver calls `applyMessageWithConfig` with overrides carrying `BlockContext` (simulate-aware `GetHashFn`), `Precompiles` (with any `MovePrecompileTo` moves), and `NoBaseFee = &!opts.Validation` |
| `x/evm/statedb/statedb.go` | `FinaliseBetweenCalls()` — clears per-call ephemeral state (logs, refund, transient storage) and resets the precompile-call counter while preserving state objects, access list, and journal across sequential calls in a shared StateDB. `SetTxContext(thash common.Hash, ti int)` — geth-aligned setter that replaces the StateDB's per-tx hash and tx index in place, so each simulated call stamps distinct `TxHash` / `TxIndex` on its emitted logs |
| `proto/ethermint/evm/v1/query.proto` (+ `x/evm/types/query.pb.go`) | `rpc SimulateV1(SimulateV1Request) returns (SimulateV1Response)`. `SimulateV1Request{opts bytes, block_number_or_hash bytes, gas_cap uint64, proposer_address ConsAddress, chain_id int64, timeout_ms int64}`. `SimulateV1Response{result bytes, error SimError}`. `SimError{code int32, message string, data string}` |
| `x/evm/keeper/grpc_query.go` | `Keeper.SimulateV1` handler — calls `validateSimulateV1Anchor` (defense-in-depth for direct-gRPC callers that bypass `rpc/backend`), parses chain ID / gas cap, unmarshals `req.Opts`, enforces `MaxSimulateBlocks` / `MaxSimulateCalls` envelope caps before the driver clones (defense-in-depth mirror of the post-sanitize keeper-side check), wires `req.TimeoutMs` into a `context.WithTimeout` (timeout `<= 0` means no keeper-side deadline — by design, since operator config is not visible at this layer), calls `k.simulateV1(...)`, marshals block results to JSON. Helper `simulateV1ErrResponse(err)` does `errors.As(err, &simErr)` to route typed failures to `response.Error`, translates raw `core.ErrIntrinsicGas` to a structured `-38013`, and collapses everything else to `status.Error(codes.Internal, …)` |
| `rpc/backend/simulate_v1.go` | Real adapter: marshals `SimOpts` to JSON, resolves caller-supplied `BlockNumberOrHash` to a concrete numeric height via `BlockNumberFromTendermint` + `TendermintBlockByNumber`, emits that concrete height in the request (sentinel `BlockNumber`s do not round-trip through JSON, so resolving here keeps the keeper anchor validator consistent), plumbs `b.RPCGasCap()` into `req.GasCap` (consumed as a request-wide gas budget by the keeper, not just a per-call cap), plumbs `b.RPCEVMTimeout()` into `req.TimeoutMs` and a wrapping `context.WithTimeout` anchored via `rpctypes.ContextWithHeight(resolvedHeight)`, invokes gRPC. On gRPC `DeadlineExceeded` (which can fire before the keeper writes a structured timeout response), translates to `NewSimTimeout`. On `response.Error`, returns the `*evmtypes.SimError` directly so geth's RPC server emits `{code, message, data}`. Unmarshals `response.Result` to `[]*evmtypes.SimBlockResult` on success |
| `rpc/namespaces/ethereum/eth/simulate_v1.go` | `PublicAPI.SimulateV1` — checks `b.SimulateDisabled()` kill switch (returns `-32601 NewSimMethodNotFound("eth_simulateV1")` so an operator-hidden endpoint is indistinguishable from a node that does not implement it), otherwise passes through to the backend |
| `rpc/namespaces/ethereum/eth/api.go` | `SimulateV1` on the `EthereumAPI` interface |
| `rpc/backend/backend.go` | `SimulateV1` and `SimulateDisabled() bool` on the `EVMBackend` interface |
| `server/config/config.go` | `SimulateDisabled bool` on `JSONRPCConfig` (default `false`); `simulate-disabled` in the TOML template at `server/config/toml.go` documents the kill switch's scope (JSON-RPC only — direct gRPC peers on port 9090 must be restricted at the network layer) |

### Test layout (established convention)

- **Public-handler tests** live in `x/evm/keeper/grpc_query_test.go` and exercise the full stack against a fully-wired `KeeperTestSuite`. Shipped cases: the Phase 1-5 set (`TestSimulateV1_EmptyOpts`, `TestSimulateV1_SingleCallHappyPath`, `TestSimulateV1_StateOverrideSentinelBubblesUp`, `TestSimulateV1_MovePrecompileToSha256`, `TestSimulateV1_NilRequest`, `TestSimulateV1_UnsupportedOverrideRejected`), the Phase 6-7 set (`TestSimulateV1_MultiCall_*`, `TestSimulateV1_MultiBlock_*` — state chaining, revert isolation, block-gas preflight, log tx-index, nonce auto-increment, CALL→CREATE nonce, multi-block state, chain linkage, precompile state, parent-hash coherence, cumulative log index), and the Phase 8 set (`TestSimulateV1_DoS_BlockCap`, `TestSimulateV1_DoS_CallCap`, `TestSimulateV1_DoS_GasPool_ClampsCallGas`, `TestSimulateV1_DoS_GasPool_ZeroIsUnlimited`, `TestSimulateV1_Timeout_LoopCheck`, `TestSimulateV1_Timeout_MidCall`, `TestSimulateV1_MultiCall_ImplicitGasAfterExhaustedBudget`, `TestSimulateV1_ExplicitGasBelowIntrinsic`). Build opts as raw JSON so tests never touch private driver types.
- **Helper unit tests** in `x/evm/keeper/simulate_v1_test.go` (`package keeper`, white-box) cover every stateless helper: `sanitizeSimChain` (gap fill, monotonic number/timestamp, span bound, `*SimError` surface), `makeSimHeader` (defaults from parent, post-merge difficulty, base-fee override, validation-derived base fee, field overrides), `resolveSimCallNonce` (default vs. explicit), `resolveSimCallGas` (default Gas, `-38015` preflight, zero-gas-limit behavior, budget-clamp, unlimited-budget noop), `simGasBudget` (zero is unlimited, clamp, consume, overflow guard), `newSimGetHashFn` (hit-base, below-base canonical, above-base sibling, not-found, canonical-unforgeability), `NewSimForkSpanUnsupported` shape.
- **Override unit tests** in `x/evm/keeper/state_override_test.go` cover `MovePrecompileTo` validation and the mezo-custom deny-list.
- **State-transition unit tests** in `x/evm/keeper/state_transition_test.go` cover `NewEVMWithOverrides` byte-equivalence with `NewEVM` and override behavior.
- **StateDB tests** in `x/evm/statedb/statedb_test.go` cover `FinaliseBetweenCalls` and `SetTxContext`.
- **Backend tests** in `rpc/backend/simulate_v1_test.go` use a mocked query client to assert proto request shape (including the resolved numeric height in `BlockNumberOrHash` and the `BaseBlockHash` field), timeout context, gRPC `DeadlineExceeded` → structured `-32016` translation, response unmarshaling, and `*SimError` bubble-up.
- **Kill-switch tests** in `rpc/namespaces/ethereum/eth/simulate_v1_test.go` cover `TestSimulateV1_KillSwitch` (disabled → structured `-32601 method not found`) and `TestSimulateV1_KillSwitchOff` (passthrough to backend).
- **Types unit tests** in `x/evm/types/simulate_v1_test.go` cover JSON round-trip for every shape — `SimOpts`, `SimBlockResult`, `SimCallResult` — and the explicit rejections baked into `UnmarshalSimOpts`.
- **System tests** under `tests/system/test/SimulateV1_*.test.ts` are TypeScript Hardhat suites run via `./tests/system/system-tests.sh`. Current files: `SimulateV1_SingleCall`, `SimulateV1_MultiCall`, `SimulateV1_MultiBlock`, `SimulateV1_MovePrecompile_ethCall`, `SimulateV1_RejectedOverrides`, `SimulateV1_Limits` (Phase 8 — request-level fatals: block cap, call cap, gas-pool exhaustion, intrinsic-gas, timeout). Phase 12 collapses these into conformance + divergence suites.

### What works end-to-end today

- Multi-call, multi-block `eth_simulateV1` round-trips end-to-end over JSON-RPC, with a single `*statedb.StateDB` threaded through every call of every block. Ephemeral writes, `commit=false`.
- State mutations propagate across calls within a block and across blocks within a request — for both the EVM journal (accounts, storage) and mezo's StateDB-scoped cached-ctx layer (custom precompile Cosmos-side writes). Covered by the `TestSimulateV1_MultiBlock_PrecompileStateChains` keeper test and the `SimulateV1_MultiCall` / `SimulateV1_MultiBlock` system tests' `btctoken` cases.
- State overrides honored per-block (balance, nonce, code, state, stateDiff).
- `MovePrecompileTo` works for stdlib precompiles (0x01-0x0A); blocks all 8 mezo custom precompiles at `0x7b7c…` with structured `-32602`.
- Per-call results: `returnData`, `logs` (with distinct `TxHash` / `TxIndex` / `BlockHash` per call via `SetTxContext` + post-block back-stamp), `gasUsed`, `status`.
- Reverts → per-call `error.code = 3`; VM errors → per-call `error.code = -32015`. Block-gas-limit overflow (`-38015`) and intrinsic-gas (`-38013`) are request-level fatals that abort the whole simulate, not per-call entries — `execute.yaml`'s `CallResultFailure` only permits `3` and `-32015` per-call, so any other reserved code surfaces top-level.
- `sanitizeSimChain` enforces strictly-increasing block numbers (`-38020`), strictly-increasing timestamps (`-38021`), and the hard 256-block span bound (`-38026`) — the latter enforced *before* gap-fill allocation to prevent pathological inputs from driving oversized header allocations.
- `sameForks` sentinel rejects simulated spans that would cross a fork boundary: `applyMessageWithConfig` reads `ctx.BlockHeight` / `ctx.BlockTime` internally for fork-gated behavior, so a span straddling forks would silently execute with the base ruleset. Conservative rejection rather than silent wrong-fork output.
- `BLOCKHASH` inside a simulated block resolves correctly over both tiers: canonical range (`height <= base.Number`) delegates to `k.GetHashFn(ctx)` which returns `ctx.HeaderHash` for the base height and consults `stakingKeeper.GetHistoricalInfo` below that; simulated-sibling range (`height > base.Number`) looks up an O(1) height-indexed map of already-finalized past siblings. Canonical-range hashes are unforgeable by any `BlockOverrides` field because `sanitizeSimChain` refuses simulated blocks whose `Number <= base.Number`.
- `rpc/backend/simulate_v1.go` resolves the caller's `BlockNumberOrHash` to a concrete numeric height before marshaling; the keeper's `validateSimulateV1Anchor` rejects direct-gRPC callers whose numeric `BlockNumberOrHash` disagrees with the anchored context.
- `baseHeaderFromContext` derives `GasLimit` from `mezotypes.BlockGasLimit(ctx)` with a fallback to `req.GasCap` — a gRPC query context anchored at a past height may carry no consensus params, which would otherwise collapse every default-Gas call to `0`.
- Strict input validation rejects `BeaconRoot`, `Withdrawals`, `BlobBaseFee` overrides as user-observable errors.
- **DoS hardening (Phase 8).** Layered defense bounds every concrete resource: hard 256-block span (`MaxSimulateBlocks`), hard 1000-call cumulative cap (`MaxSimulateCalls`, post-sanitize), request-wide gas budget seeded from `b.RPCGasCap()` and threaded as a `simGasBudget` through `processSimBlock` / `resolveSimCallGas` (clamps default-gas calls and rejects explicit-gas requests beyond remaining via `-38015`), and an `RPCEVMTimeout`-derived deadline plumbed through `req.TimeoutMs`. Each bound aborts the request as a top-level fatal; the per-block `resolveSimCallGas` preflight (gates one call against its block) and the request-wide budget (gates the whole request against node config) are independent. Direct gRPC peers passing zero values for `GasCap` / `TimeoutMs` get unbounded behavior — by design, since operator config is not visible at the keeper layer.
- **Kill switch (Phase 8).** `JSONRPCConfig.SimulateDisabled` (TOML key `simulate-disabled`) hides the JSON-RPC endpoint with `-32601 NewSimMethodNotFound` so an operator-disabled node is indistinguishable from one that does not implement the method. The flag does not affect the SDK gRPC port (default 9090); operators who need to suppress simulate entirely must additionally restrict that port at the network layer (documented inline in `server/config/toml.go`).
- **Mid-call cancellation (Phase 8).** `processSimBlock` spawns one watcher goroutine per simulated block; `OnEVMConstructed` (fires once per `applyMessageWithConfig`, i.e. per call) publishes the live `*vm.EVM` via an atomic pointer. On upstream `ctx.Done()` the watcher calls `evm.Cancel()` on whichever call is currently executing. Per-call `ctx.Err()` checks bracket every iteration so the loop exits with a structured `-32016 NewSimTimeout` rather than papering over a partial result.
- All errors flow as `*types.SimError` from constructor → keeper → gRPC `SimulateV1Response.Error` → backend → geth's RPC server emits `{code, message, data}`.

## Conventions for remaining phases

- **Errors.** Build at the call site via a `NewSim*` constructor; return through the single `error` channel; the gRPC handler splits with `errors.As(err, &simErr)`. Don't declare new error codes anywhere except `x/evm/types/simulate_v1_errors.go`. Don't translate codes between layers.
- **Tests.** End-to-end coverage goes in `x/evm/keeper/grpc_query_test.go` against the public handler. Stateless helper coverage goes in `x/evm/keeper/simulate_v1_test.go`. System tests in `tests/system/test/SimulateV1_*.test.ts`.
- **File layout.** Driver and helpers continue to live in `x/evm/keeper/simulate_v1.go` as private symbols. New flow logic lands as private functions in the same file unless it's a standalone concern (e.g., the transfer tracer in Phase 9 lives in its own package).
- **Security gates.** Phases 7, 8, 10, 15 require `/security-review` before merge. Phase 12 invokes a final review for the release cut.
- **Consensus path — never edit:** `app/ante/evm/*.go`, `x/evm/keeper/msg_server.go`, `x/evm/keeper/state_transition.go:185` (`ApplyTransaction`), `x/evm/keeper/state_transition.go:319` (`ApplyMessage`). The public wrappers `ApplyMessageWithConfig` and `SimulateMessage` may be refactored internally as long as the consensus call path produces byte-identical state transitions and all pre-existing keeper + ante tests pass unchanged. `EVMOverrides` is always passed as a pointer; the consensus path passes `nil`, which is the compiler-verified "no overrides" case.

## External references

- `~/projects/ethereum/go-ethereum/internal/ethapi/simulate.go` (reference impl)
- `~/projects/ethereum/go-ethereum/internal/ethapi/errors.go` (`-38010..-38026` codes)
- `~/projects/ethereum/go-ethereum/internal/ethapi/override/override.go` (override semantics)
- `~/projects/ethereum/go-ethereum/internal/ethapi/logtracer.go` (transfer tracer reference)
- `ethereum/execution-apis` — `src/eth/execute.yaml` (schema), `tests/eth_simulateV1/*.io` (92 conformance fixtures)

---

# Part 1 — remaining phases (against v1.14.8)

## Phase 9 — `TraceTransfers`: synthetic ERC-20 logs (ERC-7528)

**Goal.** When `opts.TraceTransfers=true`, emit synthetic `Transfer(address,address,uint256)` logs at pseudo-address `0xEeeeeEeeeEeEeeEeEeEeeEEEeeeeEeeeeeeeEEeE` (ERC-7528) for every native BTC value transfer.

**Design.** New tracer in its own package to keep `x/evm/types/tracer.go` focused.

```go
// x/evm/tracer/transfertracer/tracer.go
type Tracer struct {...}
func New() *Tracer
func (t *Tracer) Hooks() *tracing.Hooks  // OnEnter, OnExit, OnLog
func (t *Tracer) Reset(txHash common.Hash, txIdx int)
func (t *Tracer) Logs() []*ethtypes.Log
```

Per-frame log stack: `OnEnter` pushes a new frame and emits a synthetic log if `value > 0 && op != DELEGATECALL`; `OnExit` pops — on revert, drops the frame's logs; otherwise merges into parent.

In the driver: when `TraceTransfers=true`, wrap StateDB via `state.NewHookedState(stateDB, tracer.Hooks())`; pass tracer to `applyMessageWithConfig`.

**Mezo-specific.** Custom precompiles at `0x7b7c…` emit their own `Transfer` events via `AddLog`. Skip synthetic emission when `to` is a mezo custom precompile address (hard-coded list from `types.DefaultPrecompilesVersions`) to avoid double-counting.

**Files.**
- NEW `x/evm/tracer/transfertracer/tracer.go` — tracer implementation.
- NEW `x/evm/tracer/transfertracer/tracer_test.go` — unit tests.
- EDIT `x/evm/keeper/simulate_v1.go` — wire tracer when `opts.TraceTransfers=true`.

**Risks.**
- **Log amplification.** Deep call stack × N transactions = O(depth × N) synthetic logs. Bounded by Phase 8.
- **Mezo double-counting.** Exclusion list for custom precompiles. Test every custom precompile.

**Verification.**
- `transfertracer/tracer_test.go`: plain value transfer → 1 synthetic log with correct topics/data; nested 3-level call with middle revert → middle-level logs absent; DELEGATECALL with value → no synthetic log; SELFDESTRUCT with balance → synthetic log emitted; value to mezo BTC precompile → NO synthetic log.
- System: `tests/system/test/SimulateV1_TraceTransfers.test.ts` — contract sending value to EOA; parse log at ERC-7528 address; assert topic = `keccak256("Transfer(address,address,uint256)")`.

**DoD.**
- All ERC-7528 spec cases pass.
- Mezo custom-precompile exclusion verified.
- No regression with `TraceTransfers=false`.

---

## Phase 10 — `Validation=true` mode ⚠️ SECURITY-CRITICAL

**Goal.** Implement `validation=true` semantics per the execution-apis spec: tx-level validation failures are **fatal top-level errors** that abort the whole simulate request.

**Already in place.** `makeSimHeader` derives the header's `BaseFee` via `eip1559.CalcBaseFee(chainCfg, parent)` when `validation && rules.IsLondon`, otherwise zero. `processSimBlock` sets `EVMOverrides.NoBaseFee = &!opts.Validation`, so `validation=false` already relaxes base-fee checks and `validation=true` already forces them.

**Design (remaining work).** In the driver:
- `validation=true` → before each call: nonce check (`-38010`/`-38011`), balance check for `gasLimit*gasPrice + value` (`-38014`), intrinsic-gas check (`-38013`), init-code-size check (`-38025`). Any failure aborts the request and returns the top-level structured error.
- `validation=true` + derived base fee → if `msg.GasFeeCap < baseFee` → top-level `-32005` (`SimErrCodeFeeCapTooLow`).
- `validation=true` + `BlockOverrides.BaseFeePerGas` lower than the chain would accept → top-level `-38012` (`SimErrCodeBaseFeeTooLow`). Distinct from `-32005`: `-32005` is about the *transaction's* fee cap; `-38012` is about the *block's* overridden baseFee.
- `validation=true` → `msg.SkipNonceChecks = false`.
- `validation=false` (default) → `msg.SkipNonceChecks = true`. (Base-fee / `NoBaseFee` handling is already branched by `opts.Validation` in `makeSimHeader` and `processSimBlock`.)
- Revert / VM errors stay per-call regardless of validation mode (revert → code `3`; VM → `-32015`).

`SkipAccountChecks = true` always (EoA check off — custom overrides may well be a contract at the from address).

**Files.**
- EDIT `x/evm/keeper/simulate_v1.go` — pre-call validation gates inside `processSimBlock`; `skipNonceCheck` flag into the message builder.
- EDIT `x/evm/types/simulate_v1_errors.go` — add `NewSimNonceTooLow`, `NewSimNonceTooHigh`, `NewSimInsufficientFunds`, `NewSimIntrinsicGas`, `NewSimInitcodeTooLarge`, `NewSimFeeCapTooLow`, `NewSimBaseFeeTooLow` constructors as needed (constants `SimErrCodeNonceTooLow`, `SimErrCodeNonceTooHigh`, `SimErrCodeBaseFeeTooLow`, `SimErrCodeIntrinsicGas`, `SimErrCodeInsufficientFunds`, `SimErrCodeMaxInitCodeSizeExceeded` already declared).

**Risks.**
- **Divergence from fee-market `NoBaseFee` param.** `validation=true` MUST override regardless of node config. Test explicitly.
- **Fatal abort is user-observable.** Ensure deterministic — same inputs → same fatal error.
- **DoS through early-rejected txs.** Bounded by Phase 8.

**Verification.**
- `grpc_query_test.go`: `TestSimulateV1_Validation_NonceLow` — `validation=true` + nonce-too-low → top-level `-38010`.
- `grpc_query_test.go`: `TestSimulateV1_Validation_NonceHigh` — `validation=true` + nonce-too-high → top-level `-38011`.
- `grpc_query_test.go`: `TestSimulateV1_Validation_InsufficientFunds` — `validation=true` + insufficient funds → top-level `-38014`.
- `grpc_query_test.go`: `TestSimulateV1_Validation_FeeCapBelowBaseFee` — `validation=true` + `gasFeeCap < baseFee` → top-level `-32005`.
- `grpc_query_test.go`: `TestSimulateV1_Validation_BaseFeeOverrideTooLow` — `validation=true` + low `BaseFeePerGas` override → top-level `-38012`.
- `grpc_query_test.go`: `TestSimulateV1_Validation_NodeNoBaseFeeIgnored` — `validation=true` + node fee-market `NoBaseFee=true` → still enforces base fee.
- `grpc_query_test.go`: `TestSimulateV1_Validation_RevertStaysPerCall` — `validation=true` + reverting call → per-call `error.code = 3`, not fatal.
- `grpc_query_test.go`: `TestSimulateV1_NoValidation_NonceLowSucceeds` — `validation=false` + nonce-too-low → success per spec.
- **Port conformance fixtures** from `ethereum/execution-apis/tests/eth_simulateV1/` — at minimum the `-38014` and `-38011` fatal-abort cases plus matching `validation=false` success cases.
- System: `tests/system/test/SimulateV1_Validation.test.ts` — Hardhat, underfunded tx under both modes.
- **`/security-review` on the branch before merge.**

**DoD.**
- All spec-conformance fixture behaviors match.
- No regression in `validation=false` default path.
- Security review clean.

---

## Phase 11 — `ReturnFullTransactions` + sender patching + full block envelope

**Goal.** Response shape parity with spec on **two axes**:
1. `returnFullTransactions=true` emits fully-populated tx objects with `from` patched from an internal `senders` map.
2. Block-envelope fields that today are zero / placeholder / empty-tree-root scaffolding are populated correctly: `logsBloom`, `transactionsRoot`, `receiptsRoot`, `size`. `stateRoot` is the one exception — it stays zero by design (see Phase 12 known-divergence note).

The current driver hand-builds a `map[string]interface{}` in `assembleSimBlock` and never derives the post-execution roots, bloom, or block size. This phase replaces that scaffolding with a real `*types.Block` envelope so the remaining fields fall out naturally.

**Design.**

*Sender patching.* Simulated txs are unsigned (no sender recoverable from signature). The driver tracks `senders map[common.Hash]common.Address` keyed by tx hash. On response marshaling:
- `returnFullTransactions=false` (default) → tx hashes only (current behavior — `assembleSimBlock` builds the `transactions` list from `txHashes`).
- `returnFullTransactions=true` → full tx objects with `from` patched in `MarshalJSON`.

*Block-envelope completeness.* Replace the `map[string]interface{}` shortcut with a real `*types.Block`:
- Materialize a per-call `*ethtypes.Transaction` (the legacy tx already constructed by `computeSimTxHash` — reuse rather than re-hash) and accumulate them into `txs []*ethtypes.Transaction`.
- Materialize a per-call `*ethtypes.Receipt` carrying `Status`, `CumulativeGasUsed` (running total of `res.GasUsed`), `Logs` (the per-call slice already attached to `SimCallResult`), `Bloom = ethtypes.CreateBloom(types.Receipts{r})`, `TxHash`, `BlockHash` (post back-stamp), `TransactionIndex`, and `BlockNumber`. Accumulate into `receipts []*ethtypes.Receipt`.
- After the per-call loop and back-stamp, derive header fields:
  - `header.TxHash = ethtypes.DeriveSha(ethtypes.Transactions(txs), trie.NewStackTrie(nil))`
  - `header.ReceiptHash = ethtypes.DeriveSha(ethtypes.Receipts(receipts), trie.NewStackTrie(nil))`
  - `header.Bloom = ethtypes.CreateBloom(receipts)` (merged bloom across all receipts in the block)
- Build `block := ethtypes.NewBlock(header, &ethtypes.Body{Transactions: txs}, receipts, trie.NewStackTrie(nil))` so `block.Size()` returns a real RLP-encoded body length.
- Custom `MarshalJSON` for the block envelope: invokes `RPCMarshalBlock` (existing in `rpc/backend/blocks.go`), injects `calls` field, patches `from` (mirrors go-ethereum `simulate.go:85-100`).

**Out of scope for this phase — `stateRoot`.** Mezod's `statedb.StateDB` wraps a Cosmos cached multistore; there is no MPT to call `IntermediateRoot()` on. `header.Root` stays the zero hash, surfacing in JSON as `0x000…000`. Document this in Phase 12's "Known divergences" section. Echoing `base.Root` would be misleading (ignores everything the simulation did) and is explicitly rejected.

**Files.**
- EDIT `x/evm/keeper/simulate_v1.go`:
  - In `processSimBlock`: build per-call `*ethtypes.Transaction` and `*ethtypes.Receipt` alongside the existing `SimCallResult`. Reuse the `LegacyTx` already constructed by `computeSimTxHash` to avoid hashing twice.
  - After the back-stamp loop, derive `header.TxHash`, `header.ReceiptHash`, `header.Bloom` from the synthetic txs/receipts.
  - Replace `assembleSimBlock`'s hand-built map with a real `*ethtypes.Block` constructed via `ethtypes.NewBlock(...)`. Plug it into the existing `SimBlockResult.Block` shape (or replace the `map[string]interface{}` field with `*ethtypes.Block`, with `MarshalJSON` doing the flattening + `calls` injection via `RPCMarshalBlock`).
  - Populate `senders` map keyed by synthetic tx hash; thread it through to the marshaller.
- EDIT `x/evm/types/simulate_v1.go`:
  - Update `SimBlockResult` to carry the real `*ethtypes.Block` (or keep `map[string]interface{}` populated from `RPCMarshalBlock(block, …)`).
  - Extend `SimBlockResult.MarshalJSON` with `from` patching and `calls` injection; mirror go-ethereum `simulate.go:simBlockResult.MarshalJSON`.
  - Thread `ReturnFullTransactions bool` + `Senders map[common.Hash]common.Address` through the response shape.
  - Update `UnmarshalJSON` to round-trip the new shape (the existing flatten/unflatten logic still applies).
- EDIT `rpc/backend/simulate_v1.go`: no new logic expected; add a regression test asserting the populated envelope fields round-trip through the gRPC `response.Result` envelope unchanged.

**Risks.**
- **Synthetic-receipt fidelity.** `Receipt.PostState` stays nil (we don't have a state root); `Receipt.ContractAddress` is set only on CREATE; `Receipt.GasUsed` is the per-call `res.GasUsed`; `Receipt.CumulativeGasUsed` is the running block total. Verify against `core.MakeReceipt` in go-ethereum for any field we might miss.
- **`DeriveSha` stack trie**: passing `trie.NewStackTrie(nil)` matches geth's `simulate.go`. Double-check that the v1.14.8 `DeriveSha` signature accepts the StackTrie; if not, fall back to `ethtypes.DeriveSha(list, trie.NewEmpty(...))`.
- **`Logs: []` vs `Logs: null`** — already handled by `SimCallResult.MarshalJSON`; verify it still applies to the per-receipt logs list.
- **Tx hash stability**: unsigned tx `Hash()` depends on all fields — don't mutate tx between hashing and block assembly.
- **Memory cost.** Receipts and synthetic txs are kept in memory for the whole request. Bounded by Phase 8 caps; document in code comments next to the per-call loop.

**Verification.**
- `x/evm/keeper/grpc_query_test.go`: `TestSimulateV1_BlockEnvelope_BloomMatchesLogs` — call emits a `Transfer` log → `logsBloom` contains the expected indexed-topic bits; computed via `ethtypes.LogsBloom(logs)` cross-check.
- `x/evm/keeper/grpc_query_test.go`: `TestSimulateV1_BlockEnvelope_TxRootNonEmpty` — single CALL → `transactionsRoot != EmptyTxsHash`; matches `ethtypes.DeriveSha(Transactions{tx}, …)` over the same synthetic tx that produces `transactions[0]`.
- `x/evm/keeper/grpc_query_test.go`: `TestSimulateV1_BlockEnvelope_ReceiptsRootNonEmpty` — single CALL → `receiptsRoot != EmptyReceiptsHash`.
- `x/evm/keeper/grpc_query_test.go`: `TestSimulateV1_BlockEnvelope_SizeNonZero` — single CALL → `size > 0` and equals `block.Size()`.
- `x/evm/keeper/grpc_query_test.go`: `TestSimulateV1_BlockEnvelope_StateRootZero` — `stateRoot == 0x0…0`; pinned as the deliberate Mezo divergence (would fail loudly if a future change accidentally populated it from base or anywhere else).
- `x/evm/keeper/grpc_query_test.go`: `TestSimulateV1_BlockEnvelope_EmptyBlockUsesEmptyRoots` — block with zero calls (gap-fill block) → roots are the empty constants and bloom is zero. Sanity counterpart to the above tests.
- `x/evm/types/simulate_v1_test.go`: `TestSimBlockResult_FullTx_FromPatched` — `returnFullTransactions=true` → tx objects with correct `from`.
- `x/evm/types/simulate_v1_test.go`: `TestSimBlockResult_HashOnly` — `returnFullTransactions=false` → tx hashes only (existing default behavior).
- `x/evm/types/simulate_v1_test.go`: existing `TestSimCallResult_MarshalsEmptyLogsAsArray` already covers the `Logs: []` invariant — no change needed, but verify it still passes under the new patch.
- System: `tests/system/test/SimulateV1_FullTx.test.ts` — assert full tx shape round-trips.
- System: `tests/system/test/SimulateV1_BlockEnvelope.test.ts` — execute the exact ERC-20 transfer scenario from <https://github.com/mezo-org/mezod/pull/664#issuecomment-4333600375> against a localnode and assert: `logsBloom` non-zero and contains the Transfer topic bits, `transactionsRoot != EmptyTxsHash`, `receiptsRoot != EmptyReceiptsHash`, `size > 0`. The agent implementing this phase MUST resolve every concern raised in that comment except `stateRoot` (which moves to Phase 12 documented divergences).

**DoD.**
- Response JSON shape matches go-ethereum byte-for-byte on identical inputs **except** for the documented Mezo divergences (`stateRoot` zero, fields tied to EIPs mezod doesn't support).
- All five envelope fields flagged in <https://github.com/mezo-org/mezod/pull/664#issuecomment-4333600375> are addressed: `logsBloom`, `transactionsRoot`, `receiptsRoot`, and `size` are correctly derived; `stateRoot` is documented as a known divergence in Phase 12.
- All Phase 1-10 tests still green.

---

## Phase 12 — Spec conformance, fuzzing, operator docs

**Goal.** Catch behavior drift vs the execution-apis spec. Harden for attack. Ship operator docs.

**Tasks.**
- NEW `x/evm/keeper/simulate_v1_fuzz_test.go` — Go fuzz target `FuzzSimulateV1Opts` mutating JSON inputs; invariant: never panic, always returns either valid response or structured error.
- NEW `tests/system/test/SimulateV1_Conformance.test.ts` — port high-signal scenarios from `ethereum/execution-apis/tests/eth_simulateV1/`: multi-block chaining, state/block overrides, `MovePrecompileTo` (stdlib only), `validation=true` fatal aborts (-38014, -38011), `traceTransfers`, block-gas-limit overflow (-38015), span > 256 (-38026).
- **System-test consolidation pass.** Phases 1-11 each land a focused `tests/system/test/SimulateV1_*.test.ts` for easy attribution. Current files on disk: `SimulateV1_SingleCall`, `SimulateV1_MultiCall`, `SimulateV1_MultiBlock`, `SimulateV1_MovePrecompile_ethCall`, `SimulateV1_RejectedOverrides` (plus the Phase 8-11 additions: `SimulateV1_Limits`, `SimulateV1_TraceTransfers`, `SimulateV1_Validation`, `SimulateV1_FullTx`). With Phase 12's conformance suite in place, collapse:
  - DELETE each `SimulateV1_*.test.ts` whose cases the conformance suite already covers: `SingleCall`, `MultiCall`, `MultiBlock`, `MovePrecompile_ethCall`, `Validation`, `TraceTransfers`, `Limits`, `FullTx`. Do this only after confirming the conformance suite asserts the same response shapes.
  - KEEP/CREATE `SimulateV1_MezoDivergences.test.ts` for behavior the execution-apis fixtures cannot cover: custom-precompile immovability, custom-precompile `cachedCtx` continuity across calls and blocks (currently in `SimulateV1_MultiCall` / `SimulateV1_MultiBlock` — must be lifted before those are dropped), `MinGasMultiplier` gas reporting, kill-switch returning `-32601`, rejected overrides for unsupported EIPs (`BeaconRoot`, `Withdrawals`, blob fields — currently in `SimulateV1_RejectedOverrides`, fold here).
  - Target end state: **2 files** — `SimulateV1_Conformance.test.ts` + `SimulateV1_MezoDivergences.test.ts`.
- EDIT `CHANGELOG.md`, `docs/` — document:
  - New `eth_simulateV1` method.
  - `SimulateDisabled` config flag.
  - Mezo-specific divergences: custom precompiles immovable; gas reported with `MinGasMultiplier`; no EIP-4844/4788/2935/7685 support.
  - Operator guidance: front public endpoints with a reverse proxy for rate limiting; bound `RPCGasCap` + `RPCEVMTimeout` for hardware.
- **Final `/security-review` invocation** against the merged feature branch before release cut.

**Verification.**
- `go test -fuzz=FuzzSimulateV1Opts -fuzztime=10m` — no panics.
- Full system-test suite green.
- Manual smoke test against localnet with viem's `simulateCalls` equivalent.

**DoD.**
- CI green with new tests.
- Zero fuzz panics in 10-minute run.
- Docs merged.
- System tests collapsed to the two files above; no stub/obsolete files remain.
- Final security review clean.

---

# Part 2 — post-upgrade phases

**⚠ Blocked on** the [geth v1.16 upgrade project](https://linear.app/thesis-co/project/chain-geth-v116-upgrade-and-osaka-fork-compatibility-b08591b25fb5) merging to `main`. Target: 2026-05-15. Do not start Phase 13 before the upgrade merges.

**Scope discipline.** Prague/Osaka activates many EIPs simultaneously. We pick up only those that fit mezo's chain model:

| EIP / feature | Applies to mezo? | Simulate action |
|---|---|---|
| EIP-2935 parent-hash state contract | **Yes** — pure EVM | Phase 14: `ProcessParentBlockHash` pre-block system call |
| EIP-7702 SetCode txs | **Yes** — pure EVM tx type | Phase 15: accept type-4 txs + auth-list validation |
| EIP-7825 per-tx gas cap (16,777,216) | **Yes** — general tx bound | Phase 16: new DoS guard + new error code |
| `MaxUsedGas` response field (geth v1.16.9 PR #32789) | **Yes** — spec-conformant addition | Phase 16: add to `SimCallResult` |
| EIP-2537 BLS12-381 precompiles (0x0b-0x11) | **Yes** — stdlib precompiles | Absorbed automatically: `MovePrecompileTo` allow-list driven by `vm.DefaultPrecompiles(rules)` |
| EIP-7951 secp256r1 precompile | **Yes** — stdlib precompile | Absorbed automatically (same mechanism) |
| EIP-7623 calldata cost change | **Yes** — intrinsic gas | Absorbed by `k.GetEthIntrinsicGas` keeper wrapper |
| EIP-7883 / EIP-7939 / EIP-7918 | **Yes** — transparent | No simulate work |
| EIP-4844 blob transactions | **No** — chain policy rejects blob txs | Continue rejecting blob-related overrides |
| EIP-4788 parent beacon root | **No** — CometBFT, no beacon chain | Continue rejecting `BlockOverrides.BeaconRoot`; update reason text |
| EIP-7685 requests + EIP-6110/7002/7251 | **No** — no EL↔CL messaging; validator ops via `x/poa`+`x/bridge` | Continue rejecting `BlockOverrides.Withdrawals`; skip post-block `ProcessWithdrawalQueue`/`ProcessConsolidationQueue`/`ParseDepositLogs`; `RequestsHash` stays nil |

## Phase 13 — Port simulate to v1.16.9 interfaces (mechanical)

**Goal.** Update call sites where v1.16.9's signatures differ from v1.14.8's. Pure mechanical edits. ~15-20 lines modified, ~10 lines deleted net.

**What changes** (measured from `git diff v1.14.8..v1.16.9` on `core/vm/evm.go`, `core/vm/contracts.go`, `core/vm/interface.go`, `core/state/statedb.go`, `core/state_transition.go`):

| Interface | v1.14.8 → v1.16.9 change | Simulate-code fix |
|---|---|---|
| `vm.NewEVM` | drops `TxContext` param | Phase 3's `NewEVMWithOverrides`; call `evm.SetTxContext(core.NewEVMTxContext(msg))` separately where TxContext was passed |
| `vm.StateDB.SetNonce` | gains `tracing.NonceChangeReason` param | `applyStateOverrides` (Phase 2) + Phase 3 helpers: pass `tracing.NonceChangeUnspecified` |
| `vm.StateDB.SetCode` | gains `tracing.CodeChangeReason` param, returns prev code | Same; ignore return |
| `vm.StateDB.Finalise(bool)` | **NEW on interface** | **Simplification**: remove Phase 3's custom `FinaliseBetweenCalls()`; call `stateDB.Finalise(true)` (matches geth's own `simulate.go:299-303`). Saves ~20 lines |
| `evm.Call`/`Create` first param | `ContractRef` → `common.Address` | Simulate invokes `core.ApplyMessage`, not these directly; no fix |
| `core.IntrinsicGas` | gains `authList []types.SetCodeAuthorization` param | Absorbed by `k.GetEthIntrinsicGas` keeper wrapper (updated by upgrade project) |
| `vm.PrecompiledContract` | gains `Name() string` method | Mezo custom precompiles get `Name()` via upgrade project |
| `ExecutionResult.RefundedGas` | renamed to `MaxUsedGas` | Handled in Phase 16 |

**Files.**
- EDIT `x/evm/keeper/state_override.go` — add `tracing.*ChangeReason` params to affected setters.
- EDIT `x/evm/keeper/state_transition.go` — update `NewEVMWithOverrides` to the new `NewEVM` signature; insert `evm.SetTxContext(...)` calls. In `NewEVM`, swap the inline clone-and-layer precompile build for `evm.WithCustomPrecompiles(k.customPrecompiles, ...)`. In `SimulateMessage`, replace the duplicated precompile-registry rebuild with `precompiles := evm.Precompiles()` (live map), apply moves, `evm.SetPrecompiles(precompiles)` — the explicit `evm.WithPrecompiles(...)` re-attach goes away.
- EDIT `x/evm/statedb/statedb.go` — **remove** `FinaliseBetweenCalls` **only after the verification gate below**.
- EDIT `x/evm/keeper/simulate_v1.go` — replace `stateDB.FinaliseBetweenCalls()` call sites with `stateDB.Finalise(true)`.

**⚠ VERIFY BEFORE DELETING `FinaliseBetweenCalls`.** Phase 3's helper does two things: (a) standard finalise (clear logs/refund/transientStorage, preserve stateObjects), and (b) reset mezod's custom `ongoingPrecompilesCallsCounter`. Geth's new `StateDB.Finalise(true)` covers (a). Whether it also performs (b) depends on how mezod's StateDB override of `Finalise` is written on the upgrade branch. Before removing the helper:
1. Read mezod's `Finalise(true)` impl on the post-upgrade branch.
2. If it resets `ongoingPrecompilesCallsCounter`, remove the helper as planned.
3. If it does NOT, either fold the counter reset into mezod's `Finalise` override, or keep a thin wrapper that resets the counter and then calls `Finalise(true)`.

Skipping this check will silently break any simulate request that exceeds `maxPrecompilesCallsPerExecution` across call boundaries.

**Risks.** None new in the type-safe sense — purely mechanical — but see the counter-reset gate above.

**Verification.**
- All Phase 1-12 tests pass unchanged.
- The multi-call / multi-block tests that touch custom precompiles (`TestSimulateV1_MultiBlock_PrecompileStateChains`; system-side `btctoken` cases in `SimulateV1_MultiCall` / `SimulateV1_MultiBlock`) still pass — canary for the counter-reset gap.
- `go build ./...` clean; `make test-unit` green.

**DoD.**
- Simulate compiles clean against v1.16.9.
- All Phase 1-12 behavior tests green.
- No functional delta.

---

## Phase 14 — EIP-2935 parent-hash state contract

**Goal.** Post-Prague, `BLOCKHASH` can be served from the system contract at `0x…fffffffffffffffffffffffffffffffffffffffe` for up to the last 8192 blocks. Simulate must invoke `core.ProcessParentBlockHash` at the top of each simulated block (matches go-ethereum `simulate.go:267-272`) so BLOCKHASH works across the full 1..8192 range.

**Design.** In `processSimBlock`, after EVM construction and before executing user calls:

```go
if cfg.ChainConfig.IsPrague(header.Number, header.Time) {
    core.ProcessParentBlockHash(header.ParentHash, evm)
}
```

The existing `newSimGetHashFn` closure stays — it covers the `[base, base+N]` simulated-sibling range that the parent-hash contract cannot serve. Post-Prague split:
- `height > base` (simulated siblings) — `newSimGetHashFn` from in-memory headers.
- `height == base` — `newSimGetHashFn`.
- `height ∈ [base-256, base-1]` (recent canonical) — EVM `BLOCKHASH` opcode via `GetHashFn` delegating to `k.GetHashFn(ctx)`.
- `height ∈ [base-8192, base-257]` (older canonical) — parent-hash contract state.
- `height < base-8192` — zero hash.

**Files.**
- EDIT `x/evm/keeper/simulate_v1.go` — add Prague-gated `ProcessParentBlockHash` call in `processSimBlock`.

**Risks.**
- **Fork-gate correctness.** Use `cfg.ChainConfig.IsPrague(...)`; firing pre-Prague produces nonsensical state writes.
- **No divergence** with real block processing — the upgrade project adds the same call to `ApplyTransaction`; we mirror.

**Verification.**
- `simulate_v1_test.go`: `TestProcessSimBlock_Prague_BlockHashRange` — `BLOCKHASH(base - N)` for N = 100, 500, 5000, 9000 → first three return real hashes, last returns zero.
- System: `tests/system/test/SimulateV1_EIP2935.test.ts` — multi-block simulate; inside block 3 read `BLOCKHASH(base - 1000)`; cross-check against `eth_getBlockByNumber(base - 1000).hash`.

**DoD.**
- BLOCKHASH 257..8192 range works in simulated blocks (lifting the standard 256-block cap for Prague-activated simulations).

---

## Phase 15 — EIP-7702 SetCode transactions ⚠️ SECURITY-CRITICAL

**Goal.** Accept type-4 (SetCode) transactions in `calls[]`. Handle delegation-prefix (`0xef0100…`) state overrides correctly. Validate authorization lists when `validation=true`.

**Depends on** the upgrade project's "EIP-7702 SetCode transaction support" scope item — that lands Type-4 tx handling, authorization validation in ante handlers, and delegation-prefix handling in `statedb.StateDB`. Simulate extends the new machinery; we don't build it from scratch.

**Design.**
- **Input.** `TransactionArgs.AuthorizationList` is populated by the upgrade project. Simulate's JSON unmarshal passes it through unchanged; `call.ToMessage` at the keeper level absorbs it.
- **Validation mode.** When `validation=true`, validate each auth per EIP-7702: `chainID ∈ {0, chain.ID}`, nonce matches current state, signer not a contract (unless already delegated), signature recoverable. Any invalid auth → top-level fatal with new structured code (await upstream assignment; add to `simulate_v1_errors.go`).
- **State overrides + delegation.** `OverrideAccount.Code` set to `0xef0100` + 20-byte address is a delegation. `applyStateOverrides` passes through unchanged — mezod's upgraded StateDB handles the prefix semantics.
- **Cross-call nonce consistency.** Auth nonces reference current state; between calls in a simulated block, nonce advances. Validation must consult the shared StateDB, not a snapshot.

**Files.**
- EDIT `x/evm/keeper/simulate_v1.go` — recognize `authList` in the call loop; invoke per-call auth validation when `validation=true`.
- EDIT `x/evm/types/simulate_v1.go` — allow `authorizationList` in `SimBlock` calls JSON unmarshal.
- EDIT `x/evm/types/transaction_args.go` (or equivalent owned by the upgrade project) — surface `AuthorizationList` in the serializable call-args shape if not already from the upgrade.
- EDIT `x/evm/types/simulate_v1_errors.go` — add EIP-7702 auth-invalid error codes + `NewSim*` constructors.

**Risks.**
- **Delegation amplification in state overrides.** A caller could chain delegations across N EOAs to inflate storage reads per call. Bounded by Phase 8's caps + the new Phase 16 per-tx 16M cap.
- **Signature verification cost.** ~40-50μs per auth (ecdsa); 100 auths = 5ms. Negligible vs wall-clock timeout.
- **Auth signature replay across simulated blocks.** Each auth has a nonce; replay bounded by nonce increments; test explicitly that auth N in block 1 cannot be replayed in block 2.
- **`/security-review` before merge** — new tx type + auth-list validation is a rich attack surface.

**Verification.**
- `grpc_query_test.go`: `TestSimulateV1_EIP7702_ValidAuth` — single-auth type-4 tx → delegation installed; call to authorizer's address reaches delegate.
- `grpc_query_test.go`: `TestSimulateV1_EIP7702_InvalidSig_Fatal` — invalid auth signature + `validation=true` → top-level fatal.
- `grpc_query_test.go`: `TestSimulateV1_EIP7702_InvalidNonce_Fatal` — invalid auth nonce + `validation=true` → top-level fatal.
- `grpc_query_test.go`: `TestSimulateV1_EIP7702_Revocation` — auth to `0x0000…` → subsequent call reverts to EOA.
- `grpc_query_test.go`: `TestSimulateV1_EIP7702_NoValidationProceeds` — `validation=false` + invalid auth → call proceeds.
- `grpc_query_test.go`: `TestSimulateV1_EIP7702_AuthReplayBlocked` — same auth in two blocks; second fails.
- System: `tests/system/test/SimulateV1_EIP7702.test.ts` — Hardhat end-to-end delegation.
- Port upstream spec conformance fixtures for 7702 once `execution-apis/tests/eth_simulateV1/` publishes them.

**DoD.**
- Type-4 tx round-trips end-to-end.
- Auth-list validation matches spec conformance.
- Security review clean.

---

## Phase 16 — EIP-7825 per-tx gas cap + `MaxUsedGas` response field

**Goal.** Add Osaka's per-tx gas cap (16,777,216) as an additional DoS layer. Add `MaxUsedGas` to `SimCallResult`.

**Design.**
- **Per-tx gas cap.** In `sanitizeSimCall`, after defaulting, assert `call.Gas <= 16_777_216`. Violation → structured error (await upstream code assignment; reserve slot in `-380xx` range).
- **`MaxUsedGas`.** Post-call, populate from the `ExecutionResult.MaxUsedGas` field introduced in geth v1.16.9 (PR #32789). Add to `SimCallResult` struct + JSON marshaling.

**Files.**
- EDIT `x/evm/keeper/simulate_v1.go` — per-tx 16M gas cap check in `sanitizeSimCall`; populate `MaxUsedGas` from `ExecutionResult`.
- EDIT `x/evm/types/simulate_v1.go` — add `MaxUsedGas hexutil.Uint64` to `SimCallResult`.
- EDIT `x/evm/types/simulate_v1_errors.go` — add per-tx cap violation code + `NewSim*` constructor.

**Risks.** Negligible — the cap is a bound, not new surface.

**Verification.**
- `grpc_query_test.go`: `TestSimulateV1_PerTxGasCap` — `call.Gas = 20_000_000` → structured error.
- `x/evm/types/simulate_v1_test.go`: `TestSimCallResult_MaxUsedGas_RoundTrip`.
- System: extend `SimulateV1_Limits.test.ts` with the per-tx cap case.

**DoD.**
- Per-tx gas cap enforced at 16,777,216.
- `MaxUsedGas` appears in response, matching geth v1.16.9 shape.

---

# End-to-end verification strategy

1. **Go unit tests** — keeper internals, pure functions, override semantics, tracer semantics, DoS guards. Run via `make test-unit`.
2. **Go backend tests** (`rpc/backend/simulate_v1_test.go`) — mocks the query client.
3. **Hardhat system tests** (`tests/system/test/SimulateV1_*.test.ts`) — full JSON-RPC stack against a running localnet. Run via `tests/system/system-tests.sh`.
4. **Spec conformance** — port high-signal fixtures from `ethereum/execution-apis/tests/eth_simulateV1/` (Phase 12).
5. **Fuzz** — Go fuzz target to guard against panics (Phase 12).
6. **Manual localnet verification** — LAST RESORT, used only in Phases 7 and 8.
7. **Security reviews** — Phases 7, 8, 10, 15; final release-cut review at Phase 12. Uses `/security-review` against the feature branch.

# Known divergences from the execution-apis spec (documented to users)

### Part 1 (v1.14.8, Cancun)

1. **EIP-4844 / 4788 / 2935 / 7685 not supported.** Overrides for `BeaconRoot`, `Withdrawals`, blob gas fields are rejected.
2. **Custom mezo precompiles immovable.** `MovePrecompileTo` for any of the 8 addresses at `0x7b7c…` returns a structured `-32602` error (spec does not assign a dedicated -38xxx code; geth uses the same mapping for "source is not a precompile").
3. **`GasUsed` honors `MinGasMultiplier`.** Reported gas matches mezod on-chain receipts, not raw EVM gas. Documented for callers comparing across chains.
4. **`stateRoot` is always the zero hash.** Mezod's `statedb.StateDB` wraps a Cosmos cached multistore and has no Merkle Patricia Trie, so there is no `IntermediateRoot()` to call after a simulated block executes. Geth's reference implementation populates this field from `state.IntermediateRoot()`; mezod cannot replicate that without a chain-level architecture change. Echoing `base.Root` would be misleading (ignores everything the simulation did) and is explicitly rejected. Callers parsing the simulate response MUST NOT treat `stateRoot` as semantically meaningful on mezod. Pinned by `TestSimulateV1_BlockEnvelope_StateRootZero` so any accidental future change surfaces loudly.

### Part 2 (post-upgrade, Prague + Osaka)

- **Divergence (1) narrows.** EIP-2935 (Phase 14) and EIP-7702 (Phase 15) become supported. **EIP-4844, EIP-4788, EIP-7685, EIP-6110, EIP-7002, EIP-7251 stay rejected permanently** because mezod has no DA layer, no beacon chain (uses CometBFT), and no EL↔CL messaging. Rejection reason text updated to reflect mezo-specific rationale (not "EIP inactive" but "mezod chain model does not include [beacon chain / DA layer / validator queues]").
- **BLOCKHASH range extends to 8192.** EIP-2935's parent-hash contract serves the 257..8192 canonical range in simulated blocks, lifting the standard 256-block `BLOCKHASH` cap. Zero-hash fallback only for `N > 8192`.
- **Divergences (2), (3), and (4) unchanged.** Custom precompiles stay immovable; `MinGasMultiplier` gas reporting continues; `stateRoot` stays zero.

# Follow-ups / out of scope

- **EIP-4844 blob-tx simulation** — chain policy rejects blob txs.
- **EIP-4788 / EIP-7685 support** — no beacon chain (CometBFT) and no EL↔CL framework. Requires chain-level architecture changes first.
- **EIP-6110 / EIP-7002 / EIP-7251 validator queues** — validator ops via `x/poa` and `x/bridge`, not EL↔CL.
- **Relaxing custom-precompile `MovePrecompileTo` restriction** — requires per-precompile safety audit, especially for `BTCToken` (`0x7b7c…00`), `AssetsBridge` (`0x7b7c…12`), and `ValidatorPool` (`0x7b7c…11`) which interact with Cosmos modules outside EVM state.
- **Richer per-feature DoS config** (`SimulateGasCap`, `SimulateEVMTimeout`, `SimulateMaxBlocks`) if operational experience shows shared-with-`eth_call` knobs are too coarse.
- **Streaming / paginated responses** for very large simulations — spec doesn't support this today.
