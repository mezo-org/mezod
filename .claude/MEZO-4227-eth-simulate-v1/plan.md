> **Disclaimer:** This is a temporary file used during implementation of
> MEZO-4227 (`eth_simulateV1`). It should be removed once the feature is
> complete.

# Implementation plan: `eth_simulateV1` in mezod

## Context

Add the `eth_simulateV1` JSON-RPC method to mezod's EVM RPC surface. Reference impl: go-ethereum v1.16.9 `internal/ethapi/simulate.go` (full walkthrough in `research.md`). Authoritative spec: `ethereum/execution-apis`, with 92 conformance fixtures in `tests/eth_simulateV1/`.

**Why now.** Mezod is a mission-critical Cosmos-SDK EVM chain (Evmos-derived, CometBFT consensus). The broader EVM tooling ecosystem (ethers v6, viem, wallets like MetaMask/Rabby, debug UIs) increasingly assumes `eth_simulateV1` is available for multi-tx multi-block simulation with overrides — the native fit for modern "preview this transaction batch" UX. No Cosmos/Evmos-family chain (Evmos, Cronos, Kava, Canto, cosmos/evm, Sei-EVM) has shipped it yet; mezod becomes the reference implementation for the ecosystem.

**Security posture.** Chain client, mission-critical. Every phase must: (a) build green on its own, (b) ship with its own targeted tests, (c) never touch consensus-critical paths (ante handler, `msg_server.go`, `ApplyTransaction`, `ApplyMessage`). Phases 7, 8, 10, and 15 are security-critical kernels and gate a `/security-review` invocation before merge.

## Delivery sequencing

This delivery ships in **two parts**, sequenced around the separate geth v1.14.8 → v1.16.9 + Prague/Osaka upgrade project tracked in [Linear](https://linear.app/thesis-co/project/chain-geth-v116-upgrade-and-osaka-fork-compatibility-b08591b25fb5).

- **Part 1 — Phases 1-12 (this plan, ships first).** Execute now against mezod's current `v1.14.8` geth fork. No dependency on the upgrade project; merges and ships independently so users get the method ASAP.
- **Part 2 — Phases 13-16 (post-upgrade).** After the upgrade project merges to `main`, apply a small mechanical port to the new v1.16.9 interfaces, then add the Prague/Osaka features that apply to mezo.

**Why this sequencing.** Measured against the real `git diff v1.14.8..v1.16.9` on the interfaces simulate touches, the port cost is ~15 mechanical lines (and a net -10 LOC after replacing our custom `FinaliseBetweenCalls` helper with geth's new `StateDB.Finalise(true)`). The scary-looking surface-area list in the upgrade project's Linear scope is almost entirely about updating mezo's *custom* StateDB and precompile *implementations* — that's upgrade-project scope, not ours.

**Scope discipline for Part 2.** Prague/Osaka activates many EIPs simultaneously. We only pick up the ones that apply to mezo's chain model (no beacon chain because CometBFT; no blob data-availability layer; no EL↔CL request framework; bridge module handles validator-layer ops, not EL). Specifically:

| EIP / feature | Applies to mezo? | Simulate action |
|---|---|---|
| EIP-2935 parent-hash state contract | **Yes** — pure EVM | Phase 14: add `ProcessParentBlockHash` pre-block system call |
| EIP-7702 SetCode transactions | **Yes** — pure EVM tx type | Phase 15: accept type-4 txs + auth-list validation |
| EIP-7825 per-tx gas cap (16,777,216) | **Yes** — general tx bound | Phase 16: new DoS guard + new error code |
| `MaxUsedGas` response field (geth v1.16.9 PR #32789) | **Yes** — spec-conformant addition | Phase 16: add to `SimCallResult` |
| EIP-2537 BLS12-381 precompiles (0x0b-0x11) | **Yes** — stdlib precompiles | Absorbed automatically: `MovePrecompileTo` allow-list driven by `vm.DefaultPrecompiles(rules)` |
| EIP-7951 secp256r1 precompile | **Yes** — stdlib precompile | Absorbed automatically (same mechanism) |
| EIP-7623 calldata cost change | **Yes** — intrinsic gas | Absorbed by `k.GetEthIntrinsicGas` keeper wrapper |
| EIP-7883 ModExp gas bump, EIP-7939 CLZ opcode, EIP-7918 blob-base-fee bound | **Yes** — transparent | No simulate work |
| EIP-4844 blob transactions | **No** — mezo chain policy rejects blob txs | Continue rejecting blob-related overrides |
| EIP-4788 parent beacon root | **No** — mezod uses CometBFT, no beacon chain | Continue rejecting `BlockOverrides.BeaconRoot`; update reason text |
| EIP-7685 requests framework + EIP-6110/7002/7251 deposits/withdrawals/consolidations | **No** — no EL↔CL messaging; validator ops via `x/poa`+`x/bridge` | Continue rejecting `BlockOverrides.Withdrawals`; skip post-block `ProcessWithdrawalQueue`/`ProcessConsolidationQueue`/`ParseDepositLogs`; `RequestsHash` stays nil |

## Decisions (locked in)

| Decision | Choice |
|---|---|
| Feature parity | **Full parity** with the execution-apis spec (all flags: `TraceTransfers`, `Validation`, `ReturnFullTransactions`, plus `MovePrecompileTo`) |
| `MovePrecompileTo` for custom mezo precompiles (0x7b7c…00 → 0x7b7c…15, 0x7b7c1…00) | **Blocked** — stdlib precompiles (0x01–0x0A) only; custom rejects with structured error |
| DoS config | **Kill switch only**: add `SimulateDisabled bool` to `JSONRPCConfig`; reuse existing `RPCGasCap` + `RPCEVMTimeout`; hard-code 256 block cap |
| Validation error semantics | **Spec-conformant** — tx-level validation failures (`-38010`..`-38025`) are fatal top-level errors that abort the whole request. Revert / VM errors stay per-call (`3`/`-32015`, per EIP-140 + execution-apis `CallResultFailure` schema). |
| Error handling | **Single typed `*SimError{Code, Message, Data}` end-to-end.** Catalog lives in `x/evm/types/simulate_v1_errors.go`. Rides the gRPC wire on a dedicated `SimError error = 2` message on `SimulateV1Response`. No enum, no kind-string, no translator — the keeper and RPC layer share one vocabulary. Genuine internals still collapse to `status.Error(codes.Internal, …)`. |
| Gas numerics | **mezod-native** — reported `GasUsed` honors `MinGasMultiplier` (matches on-chain receipts); raw EVM gas is used only for internal pool accounting |
| EIP support (Part 1, v1.14.8) | Skip EIP-4844 / 4788 / 2935 / 7685 (not present in mezod chain config at v1.14.8); reject explicit overrides for those fields |
| EIP support (Part 2, post-upgrade) | Add EIP-2935 pre-block hook, EIP-7702 SetCode txs, EIP-7825 per-tx gas cap, `MaxUsedGas` field. Continue rejecting EIP-4844/4788/7685 and EIP-6110/7002/7251 — mezo has no beacon chain, no blob DA layer, no EL↔CL requests framework |

## Architecture summary

Simulation logic lives **inside the keeper** (new `SimulateV1` gRPC method). The RPC backend is a thin adapter that marshals `simOpts` to JSON bytes, sets up the timeout context, invokes gRPC, and marshals the response. This keeps consensus-sensitive EVM plumbing behind a single audit surface and matches the pattern of existing `EthCall`.

The driver itself — the `(k *Keeper) simulateV1` method — lives in the `keeper` package (`x/evm/keeper/simulate_v1.go`) because it needs access to unexported internals (`applyStateOverrides`, `applyMessageWithConfig`). Simulate code is split across two files along a bare-types / flow-logic seam:

- **`x/evm/types/simulate_v1.go`** — bare types and type-related helpers (constructors, converters): `SimOpts`, `SimBlock`, `SimBlockOverrides`, `SimCallResult`, `SimBlockResult` + `MarshalJSON`, `UnmarshalSimOpts`, `BuildSimCallResult`. The accompanying `x/evm/types/state_overrides.go` holds the `StateOverride` / `OverrideAccount` data types (previously keeper-private) so `SimBlock.StateOverrides` can live in types without a translation step.
- **`x/evm/types/simulate_v1_errors.go`** — unified error catalog: `SimError{Code, Message, Data}` type implementing geth's `Error()/ErrorCode()/ErrorData()` interface, `SimErrCode*` constants for every spec-reserved JSON-RPC code, and one `NewSim*` constructor per call site. Single source of truth; neither keeper nor RPC layer declares codes or translates anywhere else.
- **`x/evm/keeper/simulate_v1.go`** — flow logic: `maxSimulateBlocks`, `simTimestampIncrement`, `sanitizeSimChain`, `makeSimHeader`, `assembleSimBlock`, `computeSimTxHash`, and the `(k *Keeper) simulateV1` driver. All unexported. Driver returns `([]*SimBlockResult, error)`; `*SimError` rides the same `error` channel and callers branch with `errors.As`.

This replaces an earlier structure where every simulate symbol was keeper-private and tests reached them through an `export_test.go` bridge. Moving the bare types up lets the helper tests live alongside their targets (package `keeper` white-box for flow logic, `types_test` black-box for data types) without a test-only re-export file. `applyStateOverrides` stays in keeper — only the `StateOverride` / `OverrideAccount` data types move, not the behavior. The keeper package does not spell any `rpc/types` error code; see Override-error layering below. The precompile registry is exposed via `Keeper.activePrecompiles` (new Phase 3 helper) and wrapped by `Keeper.precompilesWithMoves` for MovePrecompileTo relocations; Phase 2 left `NewEVM` untouched and added only an unexported `evmBuilder` parameter on `applyMessageWithConfig` so `SimulateMessage` could overlay precompile relocations through a closure without changing the consensus path's execution body — Phase 3 replaced that closure with a direct `*EVMOverrides` parameter on `applyMessageWithConfig`.

**Test routing.** The `SimulateV1` gRPC handler is a dumb adapter — input/output translation plus the override-enum mapping — so its coverage is end-to-end: `TestSimulateV1_*` in `x/evm/keeper/grpc_query_test.go` exercises the full stack against a fully-wired `KeeperTestSuite` (empty opts, single-call happy path, `MovePrecompileTo` for stdlib sha256, state-override sentinel propagation, nil-request guard, unsupported-override rejection). `x/evm/keeper/simulate_v1_test.go` (package `keeper`) covers only the stateless flow helpers (sanitize × 7, make-header × 5) plus a stub pointer test documenting that the public handler is covered in `grpc_query_test.go` — re-testing `simulateV1` through a hand-built keeper would reimplement the suite setup for no additional signal.

**Error surface.** One type, one wire format, one idiom. Every spec-coded failure — unmarshal rejection, sanitize-chain violation, state-override validation, per-call preflight, VM revert, VM error — is built as `*types.SimError` at the call site via a `NewSim*` constructor. `SimError` implements `error`, so it rides any Go error channel; it also implements geth's `ErrorCode()/ErrorData()`, so the JSON-RPC server emits `{code, message, data}` verbatim when the type reaches the top of the call stack. On the gRPC wire the type rides on a dedicated `SimError error = 2` message on `SimulateV1Response`; no enum, no string kind, no translator. The keeper's `SimulateV1` handler uses `errors.As(err, &simErr)` to split typed `SimError` (populate `response.error`) from genuine internals (`status.Error(codes.Internal, …)`). The RPC backend reads `response.error` and returns `*SimError` directly — geth picks up the interface methods and emits the spec-reserved code.

The existing `NewEVM` / `ApplyMessageWithConfig` / `SimulateMessage` keep their current signatures unchanged. New variants are added alongside.

## Key references

**Existing mezod code to reuse:**
- `x/evm/keeper/state_transition.go:53` — `NewEVM` (becomes the nil-overrides path of new `NewEVMWithOverrides`)
- `x/evm/keeper/state_transition.go:133` — `GetHashFn` (fallback for canonical-range BLOCKHASH)
- `x/evm/keeper/state_transition.go:386` — `SimulateMessage` (Phase 3 rewrote the body to hand overrides through `*EVMOverrides`; public signature unchanged)
- `x/evm/keeper/state_override.go:30` — `applyStateOverrides` (extended to return validated `MovePrecompileTo` move set alongside stateDB mutations)
- `x/evm/keeper/grpc_query.go:233` — `EthCall` gRPC handler (pattern for new `SimulateV1` gRPC handler)
- `x/evm/keeper/config.go:32` — `EVMConfig` (reused as-is for simulate)
- `x/evm/statedb/statedb.go:759` — `StateDB.CacheContext()` (one shared cache ctx per simulate request)
- `x/evm/statedb/statedb.go:665-687` — `Snapshot` / `RevertToSnapshot` (per-call rollback)
- `rpc/backend/call_tx.go:428` — `DoCall` (pattern for new `SimulateV1` backend method)
- `rpc/backend/call_tx.go:462-475` — timeout context wiring (reused verbatim)
- `rpc/backend/node_info.go:299-309` — `RPCGasCap`, `RPCEVMTimeout` (reused; no new knobs)
- `rpc/types/types.go:70-85` — `StateOverride`, `OverrideAccount` (extended with `MovePrecompileTo`)
- `rpc/namespaces/ethereum/eth/api.go:283` — `PublicAPI.Call` (pattern for new `SimulateV1`)
- `proto/ethermint/evm/v1/query.proto` — add `SimulateV1` RPC alongside `EthCall`

**go-ethereum reference (read-only, for conformance checking):**
- `~/projects/ethereum/go-ethereum/internal/ethapi/simulate.go`
- `~/projects/ethereum/go-ethereum/internal/ethapi/errors.go` (error codes `-38010`..`-38026`)
- `~/projects/ethereum/go-ethereum/internal/ethapi/override/override.go`
- `~/projects/ethereum/go-ethereum/internal/ethapi/logtracer.go`

**Spec / conformance:**
- `ethereum/execution-apis` — `src/eth/execute.yaml` (schema), `docs-api/docs/ethsimulatev1-notes.mdx` (notes), `tests/eth_simulateV1/*.io` (92 fixtures)

## Phased plan

Each phase is **independently mergeable** (builds clean, tests pass, no broken contracts). Every phase ends with a concrete binary DoD. Phases 7, 8, 10, and 15 require a `/security-review` before merge.

## Part 1 — ship against v1.14.8 (Phases 1-12)

Executable today. No dependency on the geth v1.16 upgrade project.

---

### Phase 1 — Scaffolding: types + RPC registration + stubs ✅ DONE ([#658](https://github.com/mezo-org/mezod/pull/658))

**Goal.** Land the `eth_simulateV1` method name on the JSON-RPC surface returning a documented "not implemented" error. Zero behavior risk.

**Files.**
- NEW `rpc/types/simulate.go` — `SimOpts`, `SimBlock`, `BlockOverrides` (including all spec fields: `Number`, `Time`, `GasLimit`, `FeeRecipient`, `PrevRandao`, `BaseFeePerGas`, `BlobBaseFee`, `BeaconRoot`, `Withdrawals`), `SimCallResult`, `SimBlockResult`. Plain JSON-marshalable types. Custom `MarshalJSON` for `SimCallResult` forces `Logs: []` over `null` (spec-compliant).
- NEW `x/evm/types/simulate_v1_errors.go` — spec-reserved error codes (`-38010`..`-38026`, `-32005`, `-32015`, `-32016`, `-32601`..`-32603`, plus `3` for reverted) under `SimErrCode*` constants named to match the spec text in `execution-apis/src/eth/execute.yaml`. Exports `SimError{Code, Message, Data}` implementing geth's `Error()/ErrorCode()/ErrorData()` — when `*SimError` reaches the top of the RPC call stack, geth's JSON-RPC server emits `{code, message, data}` verbatim. One `NewSim*` constructor per call site (e.g. `NewSimInvalidBlockNumber`, `NewSimReverted`, `NewSimMovePrecompileSelfRef`, `NewSimVMError`).
- NEW `rpc/namespaces/ethereum/eth/simulate.go` — `PublicAPI.SimulateV1(opts SimOpts, blockNrOrHash *rpctypes.BlockNumberOrHash) ([]*SimBlockResult, error)` stub delegating to the backend.
- EDIT `rpc/namespaces/ethereum/eth/api.go` — add `SimulateV1` to `EthereumAPI` interface (new section under "EVM/Smart Contract Execution" near L89).
- EDIT `rpc/backend/backend.go` — add `SimulateV1` signature to `EVMBackend` interface (near L134 next to `DoCall`).
- NEW `rpc/backend/simulate.go` — `Backend.SimulateV1` stub returning `-32603` "eth_simulateV1 is not yet implemented". Using `-32603` (spec-listed internal error for this method) rather than `-32601` avoids misleading clients that cache "method not found" signals: the method IS registered on the `eth_` namespace, so -32601 would be semantically wrong.

**Security risks.** None — handler short-circuits.

**Verification.**
- Go unit: `rpc/types/simulate_test.go` — JSON round-trip for all types; `Logs: []` coerce-on-empty; unknown/extra fields tolerated on unmarshal.
- Go unit: `rpc/backend/simulate_test.go` — stub returns `-32603` with "not yet implemented" in the message.
- System (Hardhat): `tests/system/test/SimulateV1_Stub.test.ts` — `provider.send("eth_simulateV1", [{blockStateCalls:[]}, "latest"])` asserts correct JSON-RPC error shape.

**DoD.**
- `go build ./...` green.
- `eth_simulateV1` appears in `rpc_modules` output.
- All existing tests pass.
- New unit + system tests pass.

---

### Phase 2 — Extend state overrides with `MovePrecompileTo` (block for custom mezo precompiles) ✅ DONE ([#660](https://github.com/mezo-org/mezod/pull/660))

**Goal.** Add `MovePrecompileTo` support to the existing state-override machinery — visible to `eth_call` today, usable by simulate later. Deny-list mezo custom precompiles.

**Design.** Leave `NewEVM` untouched (byte-identical to main). Introduce an unexported `evmBuilder` function-type parameter on `applyMessageWithConfig`: the consensus-path public wrapper `ApplyMessageWithConfig` passes `k.NewEVM` directly, preserving main's execution body byte-for-byte; `SimulateMessage` supplies a closure that calls `k.NewEVM` and then — when the request has any `MovePrecompileTo` entries — rebuilds the precompile registry inline (duplicating the construction inside `NewEVM`), applies the moves, and re-attaches via `evm.WithPrecompiles(...)`. Change `applyStateOverrides` to return `(moves, error)` rather than mutating a passed-in registry; validation uses `vm.DefaultPrecompiles(rules)` to answer "is source a stdlib precompile?" and `mezoCustomPrecompileAddrs` for the deny-list. `MovePrecompileTo` is applied FIRST per spec ordering, enforcing the invariants below. The spec assigns codes `-38022` and `-38023` to two very specific conditions — the keeper marks them via distinct sentinel errors so the RPC layer can map them exactly; the remaining structured rejections map to `-32602 invalid params` (mirrors geth's `override/override.go`).

**Error surface.** `applyStateOverrides` returns `*types.SimError` (via `NewSimMovePrecompileSelfRef`, `NewSimMovePrecompileDupDest`, `NewSimNotAPrecompile`, `NewSimMoveMezoCustom`, `NewSimAccountTainted`, `NewSimDestAlreadyOverridden`, `NewSimStateAndStateDiff`) declared in `x/evm/types/simulate_v1_errors.go`. Return type is plain `error` — callers that care about the specific code (`SimulateV1` handler) branch with `errors.As(err, &simErr)`; callers that don't (`Query/EthCall`, `Query/EstimateGas`) forward the error into `status.Error(codes.Internal, err.Error())` — the user-facing message still carries the exact text the constructor chose.

The inline registry rebuild in `SimulateMessage`'s closure is the single piece of temporary duplication; it is marked `TODO (geth-upgrade)` so Phase 13's detection grep ([#651 discussion](https://github.com/mezo-org/mezod/pull/651#discussion_r3123274099)) surfaces it — v1.16.9 exposes `evm.Precompiles()` + `evm.SetPrecompiles()`, letting the closure drop the rebuild and mutate the live registry in place.

1. `movePrecompileToAddress == addr` (self-reference) → fatal `-38022` "MovePrecompileToAddress referenced itself in replacement" (`SimErrCodeMovePrecompileSelfReference`).
2. Two overrides target the same destination → fatal `-38023` "Multiple MovePrecompileToAddress referencing the same address to replace" (`SimErrCodeMovePrecompileDuplicateDest`, tracked via `dirtyAddrs`).
3. Source address is not currently a precompile → fatal `-32602` "account %s is not a precompile" (plain invalid-params; spec does not carve out a -38xxx code for this, geth uses the same mapping).
4. Source must **not** be a mezo custom precompile (0x7b7c…, check against `types.DefaultPrecompilesVersions`). Return `-32602` with message `"cannot move mezo custom precompile"`; this is a mezo-specific policy denial, not a spec case, so it rides the generic invalid-params code.

**Files.**
- EDIT `x/evm/keeper/state_override.go` — extend `overrideAccount` with `MovePrecompileTo *common.Address`; change `applyStateOverrides` to return `(map[common.Address]common.Address, error)` (the validated move set) and take `rules params.Rules` for the precompile-addr probe; apply move-validation first; enforce invariants.
- EDIT `x/evm/keeper/state_transition.go` — leave `NewEVM` untouched; add unexported `evmBuilder` function type; thread a `buildEVM evmBuilder` parameter through `applyMessageWithConfig` (the only change to its body is `evm := k.NewEVM(...)` → `evm := buildEVM(...)`). Public `ApplyMessageWithConfig` passes `k.NewEVM` so consensus callers (`ApplyTransaction`, `ApplyMessage`) see byte-identical behavior. Rewrite `SimulateMessage` to call `applyStateOverrides` (collect moves), then call `applyMessageWithConfig` with a closure that wraps `k.NewEVM` + relocates precompiles via `evm.WithPrecompiles` — the registry is rebuilt inline (duplicating `NewEVM`'s construction, flagged `TODO (geth-upgrade)`).
- EDIT `rpc/types/types.go` — add `MovePrecompileTo *common.Address` to `OverrideAccount` (L85).

**Security risks.**
- **Custom-precompile overwrite bypass** — handled by deny-list against `DefaultPrecompilesVersions`. Test matrix must cover each of the 8 custom addresses.
- **Dirty-address tracking** — prevents same-request re-overrides (port invariant from go-ethereum's `override/override.go:73-83`).
- **State sanity** — per spec, move does NOT clear source-address state. Overwriting source's `Code`/`State` is allowed. Document.

**Verification.**
- Go unit: `x/evm/keeper/state_override_test.go` — extend with ≥8 cases, pinning the exact error code returned so Phase 12's spec-conformance suite can verify wire output:
  1. Move sha256 (0x02) → 0x1234, call destination, assert sha256 output.
  2. Move from non-precompile → `-32602` "not a precompile".
  3. `movePrecompileToAddress == source` (self-reference) → `-38022` (`SimErrCodeMovePrecompileSelfReference`).
  4. Two overrides target the same destination → `-38023` (`SimErrCodeMovePrecompileDuplicateDest`).
  5. Move each of 8 mezo custom precompile addresses → rejected `-32602` "cannot move mezo custom precompile".
  6. Move + overwrite original `Code` → both applied correctly.
  7. `State` and `StateDiff` mutual exclusion preserved.
- System: `tests/system/test/SimulateV1_MovePrecompile_ethCall.test.ts` — exercise `eth_call` with `movePrecompileTo` for sha256; assert stdlib precompile works at destination.

**DoD.**
- Existing `eth_call` tests pass unchanged.
- All 8 new unit cases pass.
- Mezo custom precompiles are immovable (asserted).
- `eth_call` end-to-end accepts `movePrecompileTo` for stdlib precompiles.

---

### Phase 3 — Keeper seams: `NewEVMWithOverrides` + StateDB helpers (no behavior change) ✅ DONE ([#662](https://github.com/mezo-org/mezod/pull/662))

**Goal.** Introduce the keeper-level primitives that simulate needs, without changing any existing caller's behavior.

**Design.**
```go
// x/evm/keeper/state_transition.go
type EVMOverrides struct {
    BlockContext *vm.BlockContext                            // nil = derive from ctx
    Precompiles  map[common.Address]vm.PrecompiledContract   // nil = active registry
    NoBaseFee    *bool                                       // nil = derive from fee-market
}
func (k *Keeper) NewEVMWithOverrides(ctx sdk.Context, msg core.Message, cfg *statedb.EVMConfig,
    tracer *tracers.Tracer, stateDB vm.StateDB, over *EVMOverrides) *vm.EVM
```
`NewEVM` is refactored to call `NewEVMWithOverrides(..., nil)` — the nil-overrides path is byte-identical to today. The pointer shape keeps the consensus call site syntactically quiet (no `EVMOverrides{}` literal to review) and makes "I am not overriding anything" the default case in both reader and compiler. Per the Decisions table, `GasUsed` continues to honor `MinGasMultiplier` for both consensus and simulate paths (callers comparing across chains are documented separately).

`NewEVMWithOverrides` follows a **build-default-then-override** shape: it builds the default BlockContext / VMConfig / precompile registry exactly the way `NewEVM` always has, then replaces those with the caller's overrides when non-nil. This keeps the consensus-path construction visually unchanged and makes override behavior a small diff at the end of the function.

The previously-planned `applyMessageWithOverrides` wrapper was inlined away during implementation: `applyMessageWithConfig` directly accepts `*EVMOverrides`, threading it into `NewEVMWithOverrides`. One fewer hop, one fewer type.

The previously-planned `VMConfig(... noBaseFeeOverride *bool)` signature change was reverted: `VMConfig` keeps its original shape, and `NewEVMWithOverrides` applies the `NoBaseFee` override inline after the `VMConfig` call. Avoids leaking the simulate-only flag into a public method used by production call sites.

**Files.**
- EDIT `x/evm/keeper/state_transition.go` — add `EVMOverrides`, `NewEVMWithOverrides`, `precompilesWithMoves`, `activePrecompiles`. Refactor `NewEVM` + `applyMessageWithConfig` to delegate.
- EDIT `x/evm/statedb/statedb.go` — add `FinaliseBetweenCalls()` (clear logs, refund, transientStorage without dropping stateObjects; reset `ongoingPrecompilesCallsCounter`).

**Security risks.**
- **Regression in existing EVM construction** — eliminated by identical-behavior delegation; all existing keeper tests must pass unchanged.
- **Time cast at state_transition.go:81** — existing code does `uint64(ctx.BlockHeader().Time.Unix())` with `//nolint:gosec`. Lift into a helper with overflow check for the override path (user-supplied times can be negative).

**Verification.**
- Go unit: `x/evm/keeper/state_transition_test.go` — new cases:
  - `NewEVMWithOverrides(nil)` produces identical EVM as `NewEVM` for same inputs (byte-compare block context).
  - Override `BlockContext.BlockNumber = 999`; call a contract executing `NUMBER` opcode; assert 999.
  - Override `Precompiles = nil` → stdlib precompiles present; override with custom-only map → stdlib absent.
  - Override `NoBaseFee = &true` → fee-market param branch not consulted.
- Go unit: `x/evm/statedb/statedb_test.go` — `FinaliseBetweenCalls()` clears logs/refund but preserves stateObjects; `ongoingPrecompilesCallsCounter` resets.

**DoD.**
- All existing tests pass.
- New tests green.
- No call site outside tests uses `NewEVMWithOverrides` or `FinaliseBetweenCalls` yet; `*EVMOverrides` reaches `applyMessageWithConfig` only from `SimulateMessage` (moves-only) and the Phase 5 driver.

---

### Phase 4 — Proto + simulate-package skeleton (pure functions) ✅ DONE ([#662](https://github.com/mezo-org/mezod/pull/662))

**Goal.** Generate proto bindings for `SimulateV1` gRPC. Build the pure, side-effect-free parts of the driver: input types, `sanitizeChain`, `MakeHeader`. No execution yet.

**Design.** `simOpts` is passed as JSON bytes end-to-end (matches existing `EthCallRequest.Args` pattern at `grpc_query.go:240`) — keeps proto stable as spec evolves.

**Error surface.** Spec-coded failures travel as a **structured `SimError` message on `SimulateV1Response.error`**, not as gRPC errors. This mirrors the pre-existing VM-error pattern where reverts/EVM failures ride on `res.VmError` + `res.Failed()` (see `rpc/backend/call_tx.go:482`) — proto fields survive the gRPC boundary cleanly, Go error identity does not (`status.Error(codes.Internal, err.Error())` collapses type information). The keeper's `SimulateV1` handler branches via `errors.As(err, &simErr)` — `*types.SimError` populates `response.error`, anything else becomes a gRPC `Internal` status. The backend unpacks `response.error` back into `*evmtypes.SimError` and returns it directly; geth's JSON-RPC server reads the interface methods and emits `{code, message, data}`.

**Files.**
- EDIT `proto/ethermint/evm/v1/query.proto` — add:
  ```proto
  rpc SimulateV1(SimulateV1Request) returns (SimulateV1Response);
  message SimulateV1Request {
      bytes opts = 1;                  // JSON: {blockStateCalls, traceTransfers, validation, returnFullTransactions}
      string block_number_or_hash = 2; // "latest" / "0x..." / "0xHASH"
      uint64 gas_cap = 3;
      bytes proposer_address = 4;
      int64 chain_id = 5;
      int64 timeout_ms = 6;
  }
  message SimulateV1Response {
      bytes result = 1; // JSON: []*SimBlockResult
      // Populated when the request hits a spec-coded failure
      // (state-override validation, sanitize-chain, unsupported-EIP
      // override, preflight). Nil on success.
      SimError error = 2;
  }
  message SimError {
      int32 code = 1;     // spec-reserved JSON-RPC code (3, -32015, -32602, -38022, ...)
      string message = 2;
      string data = 3;    // hex-encoded revert payload when code == 3
  }
  ```
- Regen `x/evm/types/query.pb.go`, `rpc/types/query_client.go`.
- Simulate types + pure helpers (originally drafted as an `x/evm/keeper/simulate/` sub-package; post-Phase-5 they were collapsed into `x/evm/keeper/simulate_v1.go` as private symbols inside the `keeper` package):
  - Input types — private `simOpts`, `simBlock`, `simBlockOverrides`, `simCallResult`, `simCallError`, `simBlockResult`. JSON unmarshal via `unmarshalSimOpts` with strict validation (reject `ParentBeaconRoot` and `Withdrawals` overrides). `simBlock.StateOverrides` is typed as the existing `stateOverride` so no translation step is needed.
  - `sanitizeSimChain(base *ethtypes.Header, blocks []simBlock) ([]simBlock, error)`. Mirrors go-ethereum `simulate.go:400-459`. Rules: default number = `prev.Number + 1`; default time = `prev.Time + 12`; strict-increasing enforcement (`-38020`/`-38021`); gap-fill with empty blocks (count against `maxSimulateBlocks`); span cap (`-38026`).
  - `makeSimHeader(prev *ethtypes.Header, overrides *simBlockOverrides, rules params.Rules, chainCfg *params.ChainConfig, validation bool) *ethtypes.Header`. Pure function. Sets `UncleHash = EmptyUncleHash`, `ReceiptHash = EmptyReceiptsHash`, `TxHash = EmptyTxsHash`; `Difficulty = 0` post-merge; `MixDigest` non-nil zero post-merge (the merge-switch trigger); `BaseFee` from `eip1559.CalcBaseFee` when validation=true and override absent; `BaseFee = 0` otherwise. `ParentBeaconRoot`, `WithdrawalsHash`, `RequestsHash`, `BlobGasUsed`, `ExcessBlobGas` all nil (EIPs not active).
- EDIT `x/evm/keeper/grpc_query.go` — add `Keeper.SimulateV1` gRPC handler stub alongside the existing `EthCall` / `EstimateGas` / `TraceTx` handlers: unmarshals `opts`, sanitizes, returns a `*types.SimError` with code `-32603` "execution not yet wired" on `response.error` (short-circuit after sanitize for this phase; keeps parity with Phase 1's backend stub).
- EDIT `rpc/backend/simulate.go` — real implementation: marshals opts, sets up timeout ctx (reuse `DoCall` pattern L462-475 verbatim), invokes gRPC. On a successful gRPC response with `response.error != nil`, reconstruct `*evmtypes.SimError` and return it — geth's RPC server emits the spec-reserved code via the error-interface methods.

**Security risks.**
- **Gap-fill amplification** — caller sends `[{Number:base+1}, {Number:base+10000}]` → naive gap-fill allocates 9998 headers. Span check BEFORE allocation (research §19a).
- **JSON unmarshal size** — rely on transport-layer limit; document requirement for operators to set `RPC_MAX_REQUEST_BYTES`.
- **Negative/wrap-around time** — reject `BlockOverrides.Time` that would underflow when converted to `uint64`.

**Verification.**
- Go unit: `x/evm/keeper/simulate_v1_internal_test.go` — sanitize: port go-ethereum's `TestSimulateSanitizeBlockOrder` as `TestSanitizeSimChain_GapFill` (skip 10→13 with `Time:80` produces `[(11,62),(12,74),(13,80)]`); non-monotonic number → sentinel; non-monotonic time → sentinel; span > 256 → sentinel.
- Go unit: `x/evm/keeper/simulate_v1_internal_test.go` — header: fuzz table `TestMakeSimHeader_*` — nil overrides matches default scaffolding; post-merge zeroes `Difficulty`; `BaseFeePerGas` override wins; `validation=true` + no baseFee override → `eip1559.CalcBaseFee(prev)`.
- Go unit: `x/evm/keeper/simulate_v1_internal_test.go` — input (`TestUnmarshalSimOpts_*`): `ParentBeaconRoot` override → rejected; `Withdrawals` non-empty → rejected; `BlobBaseFee` override → rejected; unknown top-level JSON fields tolerated.
- Go backend: `rpc/backend/simulate_test.go` — mock query client, assert proto request shape + timeout ctx applied.

**DoD.**
- `make proto-gen` clean.
- `eth_simulateV1` reaches the gRPC handler and short-circuits with documented error.
- Pure-function coverage on the simulate helpers (now in `simulate_v1.go`) ≥ 90% lines.

---

### Phase 5 — Single-block simulate: one block, one call, no overrides ✅ DONE ([#662](https://github.com/mezo-org/mezod/pull/662))

**Goal.** End-to-end execution of the simplest possible `simOpts`: one `simBlock`, one call, no block overrides, state overrides honored. Ships a minimally-useful feature and proves the full pipeline.

**Design.** New unexported keeper method `(k *Keeper) simulateV1(...)` drives the end-to-end path:
```go
func (k *Keeper) simulateV1(ctx sdk.Context, cfg *statedb.EVMConfig, base *ethtypes.Header,
    opts *simulate.Opts, gasCap uint64) ([]*simulate.BlockResult, error)
```
The driver lives in the `keeper` package because it needs access to unexported internals (`applyStateOverrides`, `applyMessageWithConfig`, the `stateOverride` / `overrideAccount` types). The pure-function helpers (input parsing, sanitize-chain, header construction, block assembly) live alongside the driver in `simulate_v1.go` as private symbols — they were drafted as a separate `simulate/` sub-package but collapsed into the main keeper file once it became clear the split added a `translateOverrides` copy without buying any real isolation. The driver is unexported; external tests exercise it through the public `Keeper.SimulateV1` gRPC handler — end-to-end integration testing by default, no test-only export surface. Helper-level unit tests live in `simulate_v1_internal_test.go` (`package keeper`).

Request-level failures (sanitize-chain violations, unsupported-EIP overrides, state-override validation, preflight) are built as `*types.SimError` at the call site via a `NewSim*` constructor and returned through the single `error` channel. The gRPC handler splits with `errors.As(err, &simErr)` — typed `SimError` populates `response.error`; anything else collapses to `status.Error(codes.Internal, …)`. Per-call VM failures ride on `CallResult.Error` as `*types.SimError` directly — `BuildSimCallResult` calls `NewSimReverted(res.Ret)` or `NewSimVMError(res.VmError)` based on the EVM error, so the spec-reserved code (`3` for revert, `-32015` for other VM errors) is baked in at construction time.

Block assembly and response shaping are minimal here (spec-shaped envelope only); `returnFullTransactions` patching is Phase 11.

**Files.**
- EDIT `x/evm/keeper/grpc_query.go` — add the `SimulateV1` gRPC handler alongside the existing `EthCall` / `EstimateGas` / `TraceTx` handlers: `unmarshalSimOpts` → EVMConfig → invoke `simulateV1` → `errors.As(err, &simErr)` splits typed failures (populate `response.error` as a `SimError` proto message) from genuine internals (`status.Error(codes.Internal, …)`); on success serialize the block-result slice onto `SimulateV1Response.Result`. Also hosts `baseHeaderFromContext`.
- NEW `x/evm/keeper/simulate_v1.go` — unexported `simulateV1` driver plus all private simulate helpers (opts/block/override types, `sanitizeSimChain`, `makeSimHeader`, `assembleSimBlock`, `buildSimCallResult`, `computeSimTxHash`). `sanitizeSimChain` returns `*types.SimError` (via `NewSimInvalidBlockNumber`, `NewSimInvalidBlockTimestamp`, `NewSimClientLimitExceeded`) on its three failure modes. `assembleSimBlock` emits a minimal envelope matching `rpc/types.SimBlockResult`'s unmarshaler.
- EDIT `rpc/backend/simulate.go` — unmarshal + basic response formatting.
- EDIT `rpc/namespaces/ethereum/eth/simulate.go` — unstub.

**Security risks.**
- **First live attack surface.** Relies on existing `RPCGasCap` + `RPCEVMTimeout`. Explicit test that oversized calldata is bounded by gas cap.
- **Historical-state access** — use `rpctypes.ContextWithHeight(blockNr.Int64())` + `TendermintBlockByNumber` (existing `DoCall` pattern).

**Verification.**
- Go unit: `x/evm/keeper/simulate_v1_test.go` (`package keeper_test`, through the public gRPC handler; builds opts as raw JSON so it never touches the private driver types) — single-call happy path:
  - ERC-20 `transfer(0x..., 1)` with balance override; assert returnData, gasUsed, status.
  - Call to mezo BTC precompile (`0x7b7c…00`) `balanceOf(acct)` — assert expected value.
  - State override `Balance = 10 BTC` on sender; call `BTCToken.transfer`; assert success.
  - State override with `MovePrecompileTo` for sha256; call destination; assert correctness.
  - Call with insufficient gas limit in `TransactionArgs.Gas` → VM error reported in the per-call `error` field (per-call, not fatal).
  - Reverting call → per-call `Error.code = 3` (spec-reserved `CallResultFailure` code), `Error.data = revert reason hex`, `Error.message` prefixed with `"execution reverted"`.
- Go backend: mocked query-client integration test.
- System: `tests/system/test/SimulateV1_SingleCall.test.ts` — Hardhat: deploy contract, simulate one `transfer` with balance override, assert event logs + returnData.

**DoD.**
- Single-block single-call round-trips through JSON-RPC.
- `GasUsed` matches `eth_estimateGas` on identical input (within tolerance, since simulate uses the inflated mezod value).
- Multi-call and multi-block still return structured "not yet implemented" errors.

---

### Phase 6 — Multi-call within one block (shared state, `sanitizeCall`)

**Goal.** N calls execute in sequence inside one simulated block. State mutations from call N are visible to call N+1. Block gas limit enforced cumulatively.

**Design.** ONE `*statedb.StateDB` for the whole request built up-front. Between calls: `stateDB.FinaliseBetweenCalls()` (from Phase 3) clears logs/refund/transient without touching stateObjects. Per-call snapshot via `stateDB.Snapshot()` / `RevertToSnapshot()` if the call reverts at EVM level — but outer simulate state is preserved either way (reverts reported per-call, execution continues).

`sanitizeSimCall` (new helper in `simulate_v1.go`): default nonce via `stateDB.GetNonce(from)`; default gas via `blockCtx.GasLimit - cumGasUsed`; block-gas-limit check returns `-38015`.

**Files.**
- EDIT `x/evm/keeper/simulate_v1.go` — multi-call loop inside a single simulated block; shared StateDB; cumulative `gasUsedInBlock`; add the `sanitizeSimCall` helper.
- EDIT `x/evm/types/simulate_v1_errors.go` — add `NewSimBlockGasLimitReached(...)` constructor for `-38015`.

**Security risks.**
- **Shared StateDB journal growth** — a request with 1000 calls producing 4KB storage per call = 4MB journaled. Phase 8's global block cap (256) × per-block gas limit (mezod's `BlockMaxGasFromConsensusParams`) bounds this. Add an internal sanity check: per-request cumulative journal-size cap (e.g. 100MB hard fail).
- **Precompile call counter** — reset counter between calls so legitimate 30-call sims don't trip `maxPrecompilesCallsPerExecution`. Done via `FinaliseBetweenCalls`.
- **Per-call revert must not leak state** — cover via test.

**Verification.**
- Go unit: extend `simulate_v1_test.go` (public handler) / `simulate_v1_internal_test.go` (helpers) with multi-call cases:
  - Call 1 `transfer(B, X)`, call 2 `balanceOf(B)` → returns X.
  - Call 1 reverts, call 2 reads pre-call-1 state → unchanged.
  - Cumulative gas exceeds block gas limit → `simCallResult.Error.code = -38015` on offending call.
  - Nonce auto-increments between calls from same sender without user providing `Nonce`.
- System: `tests/system/test/SimulateV1_MultiCall.test.ts` — deploy counter contract; 3 calls each incrementing; assert final value = 3.

**DoD.**
- Multi-call works within single simulated block.
- State correctly chains call-to-call.
- Block gas limit strictly enforced; offending call gets `-38015` while preceding calls remain valid.
- Multi-block still returns "not yet implemented".

---

### Phase 7 — Multi-block chaining + simulated `GetHashFn` ⚠️ SECURITY-CRITICAL KERNEL

**Goal.** N simulated blocks in sequence. Each block's state visible to later blocks. `BLOCKHASH` inside block 3 returns hashes of simulated blocks 1 & 2.

**Design.** Block loop: `for bi, block := range sanitized { process(bi, block, headers[:bi]) }`. Shared StateDB across blocks. Between blocks: `stateDB.FinaliseBetweenCalls()`.

Custom `GetHashFn` closure (private helper in `simulate_v1.go`):
```go
func (k *Keeper) newSimGetHashFn(ctx sdk.Context, base *ethtypes.Header,
    sim []*ethtypes.Header) vm.GetHashFunc
```
Resolution order (mirror go-ethereum `simulate.go:510-563`):
1. `height == base.Number` → `base.Hash()`.
2. `height < base.Number` → delegate to existing `k.GetHashFn(ctx)` (canonical chain via `stakingKeeper.GetHistoricalInfo`).
3. `height > base.Number` → scan `sim[]` for a match. Only past siblings (slice is `headers[:bi]` from call site).
4. Not found → zero hash (matches go-ethereum and existing mezod fallback).

Pre-execution: compute preliminary headers for all sanitized blocks so `GetHashFn` can resolve future-block hashes during execution. Post-execution of each block: repair `GasUsed`, finalize hash, replace the preliminary header in place.

**Consume `BlockNumberOrHash` from the request** (carry-over from Phase 5). The backend already serializes `BlockNumberOrHash` into `SimulateV1Request.BlockNumberOrHash`, but Phase 5's driver synthesized the base header from the SDK ctx height and ignored the field. Phase 7 must parse it at the gRPC handler, fetch the real base block via `TendermintBlockByNumber` / hash lookup, and pass that as `base` into the driver — otherwise the custom `GetHashFn` cannot resolve canonical-range hashes consistently with the caller-specified anchor, and a caller passing `blockHash` gets silently rewritten to "latest".

**Files.**
- NEW `x/evm/keeper/simulate/chain.go` — `NewSimGetHashFn`.
- NEW `x/evm/keeper/simulate/process_block.go` — extract block processing into own function for testability.
- EDIT `x/evm/keeper/simulate_v1.go` — block loop; prelim header construction; post-exec repair.
- EDIT `x/evm/keeper/grpc_query.go` — resolve `BlockNumberOrHash` → real base header in the `SimulateV1` handler; stop synthesizing from ctx.

**Security risks (THIS IS THE KERNEL).**
- **Forged BLOCKHASH oracle.** Simulator-provided BLOCKHASH for future blocks is fine (by design). The critical invariant: for any **canonical** (below-base) height, MUST delegate to real `k.GetHashFn(ctx)` and MUST NOT honor any `BlockOverrides` field. Audit every block-override field for whether it could leak into the canonical range.
- **`BlockOverrides.Number < baseHeight` must be rejected.** Otherwise a caller could "simulate the past" and corrupt BLOCKHASH expectations for subsequent simulated blocks. Covered by Phase 4's `-38020` monotonic-number check in `SanitizeChain`; no additional guard needed in this phase.
- **Stale `sdk.Context.BlockHeight()`** — the context is fixed to base; the simulated block executes at `base + N` but any code reading `ctx.BlockHeight()` inside the EVM pipeline gets the wrong value. Audit: grep `ctx.BlockHeight()` within the call graph reachable from `applyMessageWithConfig` on the simulate path. Any leak must use `blockCtx.BlockNumber` instead.
- **State sprawl across blocks.** 256 blocks × 1000 calls × unbounded storage per call. Bounded by Phase 8's global gas cap + timeout, but note here for awareness.
- **BLOCKHASH depth cap at 256.** Per standard EVM semantics, the `BLOCKHASH` opcode only reaches 256 blocks back. `BLOCKHASH(base-N)` for `N > 256` returns zero. Matches go-ethereum — no mezo-specific divergence.

**Verification.**
- Go unit: `simulate/process_block_test.go` + `chain_test.go`:
  - Multi-block state: block 1 SSTORE slot, block 2 SLOAD same slot, assert observed.
  - Chain linkage: block 3 contract call `BLOCKHASH(1)`, `BLOCKHASH(2)`, `BLOCKHASH(0)` (base) — assert all three match expected simulated/base hashes. (Port `TestSimulateV1ChainLinkage` from go-ethereum `api_test.go:2466`.)
  - `BlockOverrides.Number < baseHeight` → `-38020` via Phase 4's `SanitizeChain` (regression guard; covered there, not re-added here).
  - `BLOCKHASH(base-N)` where `N ≤ 256` returns real canonical hash via `k.GetHashFn(ctx)`; `N > 256` returns zero hash.
- System: `tests/system/test/SimulateV1_MultiBlock.test.ts` — 5-block simulation, contract asserts `block.number` increments correctly.
- **Manual localnet verification** (JUSTIFIED as LAST RESORT for this phase): run against chain with ≥100 historical blocks; issue simulate that BLOCKHASHes a canonical block below base; cross-check against `eth_getBlockByNumber(height).hash`. This catches IAVL/query-at-height edge cases that mocks cannot.
- **Invoke `/security-review` on the branch before merge.**

**DoD.**
- Chained multi-block state works.
- BLOCKHASH returns consistent values across canonical + simulated range.
- Canonical-range BLOCKHASH is not influenceable by any block override.
- Manual localnet check green.
- Security review clean.

---

### Phase 8 — DoS guards + kill switch ⚠️ SECURITY-CRITICAL

**Goal.** Layered, defense-in-depth DoS bounding. Add the single operator kill-switch.

**Design.**
- **Kill switch.** New field `SimulateDisabled bool` on `JSONRPCConfig` (`server/config/config.go`). Default `false`. Checked in `PublicAPI.SimulateV1` before reaching backend.
- **Block cap.** Hard-code `maxSimulateBlocks = 256` in `x/evm/keeper/simulate/driver.go`. Enforced twice: at the RPC layer (fast fail), and inside `SanitizeChain` for defense-in-depth (span check, not just input len).
- **Gas pool.** One `uint64 gasRemaining` initialized from `b.RPCGasCap()` (existing knob). Deducted on every call's `res.GasUsed`. Exhaustion → structured `-38015`-shaped error, aborts request.
- **Timeout.** Single `context.WithTimeout(ctx, b.RPCEVMTimeout())` at the backend entry (reuse `DoCall`'s pattern L462-475). Inside the keeper loop, check `ctx.Err()` before every call. Mirror go-ethereum's `applyMessageWithEVM` goroutine (`api.go:752-754`) that calls `evm.Cancel()` on ctx-done. On ctx-done return a top-level fatal `-32016` "execution aborted (timeout = Xs)" (`SimErrCodeTimeout`, spec-reserved for this method).
- **Per-block gas limit.** Already from Phase 6 via `SanitizeCall`.
- **Cumulative call count.** Soft cap of 1000 calls per request (hard-coded constant, not configurable for v1). Prevents pathological 256-block × 10000-calls requests that would still bust memory even within gas cap.

**Files.**
- EDIT `server/config/config.go` — add `SimulateDisabled bool` to `JSONRPCConfig`; update TOML template + defaults.
- EDIT `rpc/backend/backend.go` — add `SimulateDisabled() bool` accessor.
- EDIT `rpc/backend/simulate.go` — kill-switch check; `RPCGasCap` / `RPCEVMTimeout` plumbing.
- EDIT `rpc/namespaces/ethereum/eth/simulate.go` — kill-switch at entry (short-circuit before backend).
- EDIT `x/evm/keeper/simulate/driver.go` — enforce 256 block cap, 1000 call cap, shared gas pool; `ctx.Err()` checks; `evm.Cancel()` on ctx-done.

**Security risks.**
- **Failure-open gaps.** If any guard silently absorbs a cap-reached error, others must still terminate. Test each guard in isolation.
- **Resource leak on cancel.** Deferred cancel; goroutine exits cleanly; no dangling state in StateDB.
- **Concurrent-request saturation.** Each request has its own StateDB snapshot; in-process single-threaded execution. Multiple concurrent requests are bounded by the RPC server's thread pool (`--rpc.http.threads`). Document in ops guide.

**Verification.**
- Go unit: `simulate/dos_test.go` — one test per guard:
  1. Request with >256 blocks → `-38026`.
  2. Request with ≥1000 calls total → structured error.
  3. Request that exhausts `gasRemaining` → top-level fatal `-38015`-shaped error; request aborts immediately (matches go-ethereum).
  4. Timeout fires during a long call → request returns `-32016` `"execution aborted (timeout = 5s)"` within 5.2s.
  5. Kill switch: `SimulateDisabled=true` → RPC returns `-32601 "the method eth_simulateV1 does not exist/is not available"` immediately (kill-switch intentionally impersonates "method absent" so the operator can hide the endpoint wholesale; distinct from the Phase 1 "not yet implemented" stub which uses `-32603`).
- Go unit: layered failure — each guard triggers under controlled inputs even if others are relaxed.
- System: `tests/system/test/SimulateV1_Limits.test.ts` — 257 blocks → error; kill-switch test via config reload.
- **Manual localnet verification** (JUSTIFIED): run 256-block × 1000-call simulation under `RPCEVMTimeout=5s`; capture pprof heap snapshot before/after; assert memory stable (<200MB delta, no leaks).
- **Invoke `/security-review` on the branch before merge.**

**DoD.**
- All 5 DoS guards demonstrably terminate a hostile request with documented error.
- Kill switch observed to fully disable via config reload.
- Memory load test clean.
- Security review clean.

---

### Phase 9 — `TraceTransfers`: synthetic ERC-20 logs (ERC-7528)

**Goal.** When `opts.TraceTransfers=true`, emit synthetic `Transfer(address,address,uint256)` logs at pseudo-address `0xEeeeeEeeeEeEeeEeEeEeeEEEeeeeEeeeeeeeEEeE` (ERC-7528) for every native BTC (mezod's native token) value transfer.

**Design.** New tracer in its own package to keep the existing `x/evm/types/tracer.go` focused.
```go
// x/evm/tracer/transfertracer/tracer.go
type Tracer struct {...}
func New() *Tracer
func (t *Tracer) Hooks() *tracing.Hooks  // OnEnter, OnExit, OnLog
func (t *Tracer) Reset(txHash common.Hash, txIdx int)
func (t *Tracer) Logs() []*ethtypes.Log
```
Per-frame log stack: `OnEnter` pushes new frame and emits synthetic log if `value > 0 && op != DELEGATECALL`; `OnExit` pops — on revert, drops the frame's logs; otherwise merges into parent.

Inside simulate driver: when `TraceTransfers=true`, wrap StateDB via `state.NewHookedState(stateDB, tracer.Hooks())`; pass tracer to `applyMessageWithConfig`.

**Mezo-specific** (given "Block for custom precompiles" decision): custom precompiles at `0x7b7c…` emit their own `Transfer` events via `AddLog`. Skip synthetic emission when `to` is a mezo custom precompile address to avoid double-counting. Hard-coded exclusion list from `types.DefaultPrecompilesVersions`.

**Files.**
- NEW `x/evm/tracer/transfertracer/tracer.go` — tracer implementation.
- NEW `x/evm/tracer/transfertracer/tracer_test.go` — unit tests.
- EDIT `x/evm/keeper/simulate/driver.go` — wire tracer when `opts.TraceTransfers=true`.

**Security risks.**
- **Log amplification.** Deep call stack + N transactions produces O(depth × N) synthetic logs. Bounded by Phase 8's call cap + gas cap, but note.
- **Mezo double-counting.** Exclusion list for custom precompiles. Test every custom precompile.

**Verification.**
- Go unit: `transfertracer/tracer_test.go`:
  - Plain value transfer → 1 synthetic log with correct topics + data.
  - Nested 3-level call, middle reverts → middle-level logs absent.
  - DELEGATECALL with value → no synthetic log (spec).
  - SELFDESTRUCT with balance → synthetic log emitted.
  - Value sent to mezo BTC precompile address → NO synthetic log (double-count suppression).
- System: `tests/system/test/SimulateV1_TraceTransfers.test.ts` — contract sending value to EOA; parse log at ERC-7528 address; assert topic = `keccak256("Transfer(address,address,uint256)")`.

**DoD.**
- All ERC-7528 spec cases pass.
- Mezo custom-precompile exclusion verified.
- No regression with `TraceTransfers=false`.

---

### Phase 10 — `Validation=true` mode ⚠️ SECURITY-CRITICAL (spec-conformant fatal errors)

**Goal.** Implement `validation=true` semantics per the execution-apis spec: tx-level validation failures are **fatal top-level errors** that abort the whole simulate request.

**Design.** In the driver:
- `validation=true` → before each call: run nonce check (returns `-38010`/`-38011`), balance check for `gasLimit*gasPrice + value` (returns `-38014`), intrinsic-gas check (returns `-38013`), init-code-size check (returns `-38025`). Any failure aborts the request and returns the top-level structured error.
- `validation=true` + no `BaseFee` override → compute via `eip1559.CalcBaseFee(cfg, parent)`; if `msg.GasFeeCap < baseFee` → top-level `-32005` (`SimErrCodeFeeCapTooLow`, "Transactions maxFeePerGas is too low").
- `validation=true` + `BlockOverrides.BaseFeePerGas` override supplied but below what the surrounding block context makes acceptable (e.g., negative after dropping out of bounds, or lower than an in-request prior-block baseFee the chain would reject) → top-level `-38012` (`SimErrCodeBaseFeeTooLow`, "Transactions baseFeePerGas is too low"). This is a separate code from `-32005` per the execution-apis spec: `-32005` is about the *transaction's* fee cap, `-38012` is about the *block's* overridden baseFee being rejected outright.
- `validation=true` → `EVMOverrides.NoBaseFee = &false` (force real base-fee checks regardless of fee-market `NoBaseFee` param).
- `validation=true` → `msg.SkipNonceChecks = false` (fail via core package too).
- `validation=false` (default) → preserves Phase 4 behavior: `BaseFee = 0`, `NoBaseFee = true`, `SkipNonceChecks = true`.
- Revert / VM errors (invalid opcode, OOG) stay per-call in `simCallResult.Error` regardless of validation mode (revert → code `3` per `CallResultFailure`; VM → `-32015`).

`SkipAccountChecks = true` always (EoA check off — per research §13 and spec; custom overrides may well be a contract at the from address).

**Files.**
- EDIT `x/evm/keeper/simulate/driver.go` — two mode branches; pre-call validation gates.
- EDIT `x/evm/keeper/simulate/header.go` — base-fee derivation branch.
- EDIT `x/evm/keeper/simulate/sanitize.go` — expose `skipNonceCheck` flag into the message builder.

**Security risks.**
- **Divergence from fee-market `NoBaseFee` param.** `validation=true` MUST override regardless of the node's fee-market setting, otherwise a node with `feeMarket.NoBaseFee=true` produces incorrect realism checks. Test explicitly.
- **Fatal abort is user-observable.** Ensure the abort path is deterministic (same inputs → same fatal error) — guards against attackers sowing non-determinism in a debugging flow.
- **DoS through early-rejected txs.** A caller submitting many txs with obviously-bad nonces gets cheap failures (fatal on first). Mitigated because only the first invalid is evaluated. But cost of evaluation up to that point must be bounded — already by Phase 8.

**Verification.**
- Go unit: `simulate/validation_test.go` — matrix:
  - `validation=false` + nonce-too-low call → **success** per spec (nonce check bypassed).
  - `validation=true` + nonce-too-low → top-level `-38010`.
  - `validation=true` + nonce-too-high → top-level `-38011`.
  - `validation=true` + insufficient funds (gas*price + value > balance) → top-level `-38014`.
  - `validation=true` + gasFeeCap < baseFee → top-level `-32005` (`SimErrCodeFeeCapTooLow`).
  - `validation=true` + `BlockOverrides.BaseFeePerGas` lower than the chain would accept → top-level `-38012` (`SimErrCodeBaseFeeTooLow`).
  - `validation=true` + fee-market `NoBaseFee=true` on node → still enforces base fee (spec compliance, not node config).
  - `validation=true` + reverting call → **per-call** `error.code = 3` (`SimErrCodeReverted`, spec `CallResultFailure` pins revert at `3`; not fatal — matches spec: execution-level errors stay per-call).
- **Port the relevant go-ethereum conformance fixtures** from `ethereum/execution-apis/tests/eth_simulateV1/` — specifically the `-38014` and `-38011` fatal-abort cases, plus matching `validation=false` success cases.
- System: `tests/system/test/SimulateV1_Validation.test.ts` — Hardhat, underfunded tx under both modes.
- **Invoke `/security-review` on the branch before merge.**

**DoD.**
- All spec conformance fixture behaviors match mezod.
- No regression in `validation=false` default path.
- Security review clean.

---

### Phase 11 — `ReturnFullTransactions` + sender patching + full block envelope

**Goal.** Response shape parity with spec. `returnFullTransactions=true` emits fully-populated tx objects with `from` patched from an internal `senders` map.

**Design.** Simulated txs are unsigned (no sender recoverable from signature). The driver tracks `senders map[common.Hash]common.Address` keyed by tx hash. On response marshaling:
- `returnFullTransactions=false` (default) → tx hashes only, standard shape.
- `returnFullTransactions=true` → full tx objects with `from` patched in `MarshalJSON`.

Custom `MarshalJSON` for the block envelope: invokes `RPCMarshalBlock` (existing in `rpc/backend/blocks.go`) then injects `calls` field + patches `from` (mirrors go-ethereum `simulate.go:85`).

**Files.**
- NEW `rpc/types/simulate_marshal.go` — custom `MarshalJSON` for `SimBlockResult`.
- EDIT `x/evm/keeper/simulate/driver.go` — populate `senders` map.
- EDIT `rpc/backend/simulate.go` — apply patching on unmarshaled response.
- EDIT `x/evm/keeper/simulate/assemble.go` — construct assembled block with unsigned txs.

**Security risks.** Low (cosmetic). Watch for:
- `Logs: []` vs `Logs: null` (force `[]` per spec).
- Tx hash stability: unsigned tx `Hash()` depends on all fields — ensure we don't mutate tx between hashing and block assembly.

**Verification.**
- Go unit: `simulate_marshal_test.go` — JSON round-trip:
  - `returnFullTransactions=false` → tx hashes.
  - `returnFullTransactions=true` → tx objects with correct `from`.
  - Empty `logs` serialized as `[]` not `null`.
- System: `tests/system/test/SimulateV1_FullTx.test.ts` — assert full tx shape round-trip.

**DoD.**
- Response JSON shape matches go-ethereum byte-for-byte on identical inputs (excluding fields tied to EIPs mezod doesn't support).
- All 11 phases' tests still green.

---

### Phase 12 — Spec conformance, fuzzing, operator docs

**Goal.** Catch behavior drift vs the execution-apis spec. Harden for attack. Ship operator docs.

**Tasks.**
- NEW `x/evm/keeper/simulate/fuzz_test.go` — Go fuzz target `FuzzSimulateV1Opts` mutating JSON inputs; invariant: never panic, always returns either valid response or structured error.
- NEW `tests/system/test/SimulateV1_Conformance.test.ts` — port key scenarios from `ethereum/execution-apis/tests/eth_simulateV1/`:
  - Multi-block chaining
  - State/block overrides
  - `MovePrecompileTo` (stdlib only)
  - `validation=true` fatal aborts (-38014, -38011)
  - `traceTransfers`
  - Block-gas-limit overflow (-38015)
  - Span > 256 (-38026)
- **System-test consolidation pass.** Phases 1-11 each land a focused `tests/system/test/SimulateV1_*.test.ts` file for easy attribution during development. With Phase 12's conformance suite in place, collapse the redundant ones:
  - DELETE `SimulateV1_Stub.test.ts` — Phase 5 made the stub return real data, so the test now asserts a lie.
  - DELETE each `SimulateV1_*.test.ts` whose cases the new conformance suite already covers (likely: `SingleCall`, `MultiCall`, `MultiBlock`, `MovePrecompile_ethCall`, `Validation`, `TraceTransfers`, `Limits`, `FullTx`). Do this only after confirming the conformance suite asserts the same response shapes.
  - KEEP a `SimulateV1_MezoDivergences.test.ts` (NEW — may be lifted out of existing files) for behavior the execution-apis fixtures cannot cover: custom-precompile immovability, `MinGasMultiplier` gas reporting, kill-switch returning `-32601`, rejected overrides for EIPs mezo does not support (`BeaconRoot`, `Withdrawals`, blob fields).
  - Target end state: **2 files** — `SimulateV1_Conformance.test.ts` (spec parity) + `SimulateV1_MezoDivergences.test.ts` (deliberate deltas).
- EDIT `CHANGELOG.md`, `docs/` (or README section) — document:
  - New `eth_simulateV1` method.
  - `SimulateDisabled` config flag.
  - Mezo-specific divergences: custom precompiles are immovable; gas reported with `MinGasMultiplier`; no EIP-4844/4788/2935/7685 support (rejected in overrides).
  - Operator guidance: public endpoints should front with a reverse proxy for rate limiting; bound `RPCGasCap` + `RPCEVMTimeout` for your hardware.
- **Final `/security-review` invocation** against the merged feature branch before release cut.

**Verification.**
- `go test -fuzz=FuzzSimulateV1Opts -fuzztime=10m` — no panics.
- Full system-test suite green.
- Manual: smoke test against localnet with `viem`'s `simulateCalls` equivalent (direct `eth_simulateV1` call).

**DoD.**
- CI green with new tests.
- Zero fuzz panics in 10-minute run.
- Docs merged.
- `tests/system/test/SimulateV1_*.test.ts` collapsed to the two files named above; no stub/obsolete files remain.
- Final security review clean.

---

## Part 2 — post-upgrade phases (after geth v1.16.9 + Prague/Osaka lands)

**⚠ Blocked on** the separate [geth v1.16 upgrade project](https://linear.app/thesis-co/project/chain-geth-v116-upgrade-and-osaka-fork-compatibility-b08591b25fb5) merging to mezod's `main`. Upgrade-project target: **2026-05-15**. Do not start Phase 13 before the upgrade merges.

### Phase 13 — Port simulate to v1.16.9 interfaces (mechanical)

**Goal.** Update the call sites where v1.16.9's signatures differ from v1.14.8's. Pure mechanical edits, no behavior change, ~15-20 lines modified, ~10 lines deleted net.

**What changes** (measured from `git diff v1.14.8..v1.16.9` on `core/vm/evm.go`, `core/vm/contracts.go`, `core/vm/interface.go`, `core/state/statedb.go`, `core/state_transition.go`):

| Interface | v1.14.8 → v1.16.9 change | Simulate-code fix |
|---|---|---|
| `vm.NewEVM` | drops `TxContext` param | 3 call sites in Phase 3's `NewEVMWithOverrides`; call `evm.SetTxContext(core.NewEVMTxContext(msg))` separately where TxContext was passed |
| `vm.StateDB.SetNonce` | gains `tracing.NonceChangeReason` param | `applyStateOverrides` (Phase 2) + Phase 3 helpers: pass `tracing.NonceChangeUnspecified` |
| `vm.StateDB.SetCode` | gains `tracing.CodeChangeReason` param, returns prev code | same; ignore return |
| `vm.StateDB.SetState` | returns prev value (`common.Hash`) | we don't depend on return — no change |
| `vm.StateDB.GetCommittedState` | renamed to `GetStateAndCommittedState`, returns `(current, committed)` | simulate doesn't call this — no change |
| `vm.StateDB.SubBalance`/`AddBalance`/`SelfDestruct`/`SelfDestruct6780` | return prev values | we don't depend on returns — no change |
| `vm.StateDB.Finalise(bool)` | **NEW on interface** | **simplification**: remove Phase 3's custom `FinaliseBetweenCalls()` helper; call `stateDB.Finalise(true)` (matches geth's own `simulate.go:299-303`). Saves ~20 lines. |
| `vm.StateDB.AccessEvents()` | **NEW** (Verkle witness) | mezod custom StateDB implements via upgrade project; no direct simulate use |
| `evm.Call`/`Create` first param | `ContractRef` → `common.Address` | simulate invokes `core.ApplyMessage`, not these directly; no simulate fix |
| `core.IntrinsicGas` | gains `authList []types.SetCodeAuthorization` param | absorbed by `k.GetEthIntrinsicGas` keeper wrapper (updated by upgrade project); simulate inherits |
| `vm.PrecompiledContract` | gains `Name() string` method | simulate consumes the interface; mezo custom precompiles get `Name()` via upgrade project. No simulate fix. |
| `ExecutionResult.RefundedGas` | renamed to `MaxUsedGas` | handled in Phase 16 below |

**Files.**
- EDIT `x/evm/keeper/state_override.go` — add `tracing.*ChangeReason` params to affected setters. `applyStateOverrides` already returns the move set as of Phase 2; no further signature change required.
- EDIT `x/evm/keeper/state_transition.go` — update `NewEVMWithOverrides` to the new `NewEVM` signature; insert `evm.SetTxContext(...)` calls where needed. In `NewEVM`, swap the inline clone-and-layer precompile build for `evm.WithCustomPrecompiles(k.customPrecompiles, ...)` (geth v1.16.9 folds default-precompile management into the EVM itself). Resolve the Phase 2 `TODO (geth-upgrade)` marker inside `SimulateMessage`'s `buildEVM` closure: replace the duplicated precompile-registry rebuild with `precompiles := evm.Precompiles()` (live map), apply `moves`, and call `evm.SetPrecompiles(precompiles)`. The explicit `evm.WithPrecompiles(...)` re-attach goes away.
- EDIT `x/evm/statedb/statedb.go` — **remove** custom `FinaliseBetweenCalls` helper (no longer needed) **only after confirming the behavior gap below is closed**.
- EDIT `x/evm/keeper/simulate/driver.go` — replace `stateDB.FinaliseBetweenCalls()` call sites with `stateDB.Finalise(true)`.

**⚠ VERIFY BEFORE DELETING `FinaliseBetweenCalls`.** Phase 3's helper does two things: (a) standard finalise (clear logs/refund/transientStorage, preserve stateObjects), and (b) reset mezod's custom `ongoingPrecompilesCallsCounter`. Geth's new `StateDB.Finalise(true)` covers (a). Whether it also performs (b) depends on how mezod's StateDB override of `Finalise` is written on the upgrade branch. Before removing the helper:
  1. Read mezod's `Finalise(true)` impl on the post-upgrade branch.
  2. If it resets `ongoingPrecompilesCallsCounter`, remove the helper as planned.
  3. If it does NOT, either fold the counter reset into mezod's `Finalise` override, or keep a thin wrapper that resets the counter and then calls `Finalise(true)`.
Skipping this check will silently break any simulate request that exceeds `maxPrecompilesCallsPerExecution` across call boundaries.

**Security risks.** None new in the type-safe sense — purely mechanical — but see the counter-reset verification step above; a silent behavior loss there would degrade multi-call simulations.

**Verification.**
- All Phase 1-12 tests pass unchanged on the upgraded branch.
- Multi-call simulate tests from Phase 6 (≥2 calls touching custom precompiles) still pass — this is the canary for the counter-reset gap.
- `go build ./...` clean.
- `make test-unit` green.

**DoD.**
- Simulate compiles clean against v1.16.9.
- All Phase 1-12 behavior tests green.
- No functional delta.

---

### Phase 14 — EIP-2935 parent-hash state contract

**Goal.** Post-Prague, `BLOCKHASH` can be served from the system contract at `0x…fffffffffffffffffffffffffffffffffffffffe` for up to the last 8192 blocks. Simulate must invoke `core.ProcessParentBlockHash` at the top of each simulated block (matches go-ethereum `simulate.go:267-272`) so BLOCKHASH works across the full 1..8192 range.

**Design.** In `processBlock` (from Phase 7), after EVM construction and before executing any user calls:
```go
if cfg.ChainConfig.IsPrague(header.Number, header.Time) {
    core.ProcessParentBlockHash(header.ParentHash, evm)
}
```

The Phase 7 `simulatedGetHashFn` closure stays — it still covers the `[base, base+N]` simulated-sibling range that the parent-hash contract cannot serve. Post-Prague the split is:
- `height > base` (simulated siblings) — served by `simulatedGetHashFn` from in-memory headers.
- `height == base` — served by `simulatedGetHashFn`.
- `height ∈ [base-256, base-1]` (recent canonical) — EVM `BLOCKHASH` opcode uses `GetHashFn` delegating to existing `k.GetHashFn(ctx)`.
- `height ∈ [base-8192, base-257]` (older canonical) — served by the parent-hash contract state (populated by prior real-chain blocks).
- `height < base-8192` — zero hash.

**Files.**
- EDIT `x/evm/keeper/simulate/process_block.go` — add Prague-gated `ProcessParentBlockHash` call.

**Security risks.**
- **Fork-gate correctness.** Must use `cfg.ChainConfig.IsPrague(...)`; firing pre-Prague produces nonsensical state writes.
- **No divergence** with real block processing — the upgrade project adds the same call to `ApplyTransaction`; we mirror.

**Verification.**
- Go unit: `simulate/process_block_test.go` — `BLOCKHASH(base - N)` for N = 100, 500, 5000, 9000 → first three return real hashes, last returns zero.
- System: `tests/system/test/SimulateV1_EIP2935.test.ts` — multi-block simulate; inside block 3 read `BLOCKHASH(base - 1000)`; cross-check against `eth_getBlockByNumber(base - 1000).hash`.

**DoD.**
- BLOCKHASH 257..8192 range works in simulated blocks (lifting the standard 256-block cap for Prague-activated simulations).

---

### Phase 15 — EIP-7702 SetCode transactions ⚠️ SECURITY-CRITICAL

**Goal.** Accept type-4 (SetCode) transactions in `calls[]`. Handle delegation-prefix (`0xef0100…`) state overrides correctly. Validate authorization lists when `validation=true`.

**Depends on.** Upgrade project's "EIP-7702 SetCode transaction support" scope item — that lands Type-4 tx handling, authorization validation in ante handlers, and delegation-prefix handling in `statedb.StateDB`. Simulate extends the new machinery; we don't build it from scratch.

**Design.**
- **Input.** `TransactionArgs.AuthorizationList` is populated by the upgrade project. Simulate's JSON unmarshal passes it through unchanged; `call.ToMessage` at the keeper level absorbs it.
- **Validation mode.** When `validation=true`, validate each auth per EIP-7702: `chainID ∈ {0, chain.ID}`, nonce matches current state, signer not a contract (unless already delegated to one), signature recoverable. Any invalid auth → top-level fatal error with new structured code (await upstream assignment; add to `x/evm/types/simulate_v1_errors.go`).
- **State overrides + delegation.** `OverrideAccount.Code` set to `0xef0100` + 20-byte address is interpreted as a delegation. `applyStateOverrides` passes through unchanged — mezod's upgraded StateDB handles the prefix semantics.
- **Cross-call nonce consistency.** Auth nonces reference current state; between calls in a simulated block, nonce advances. Validation must consult the shared StateDB, not a snapshot.

**Files.**
- EDIT `x/evm/keeper/simulate/driver.go` — recognize `authList` in the call loop; invoke per-call auth validation when `validation=true`.
- EDIT `x/evm/keeper/simulate/input.go` — allow `authorizationList` in JSON `calls[]` unmarshal.
- EDIT `rpc/types/simulate.go` — surface `AuthorizationList` in the serializable call-args shape if not already present from the upgrade.
- EDIT `x/evm/types/simulate_v1_errors.go` — add EIP-7702 auth-invalid error codes + `NewSim*` constructors.

**Security risks.**
- **Delegation amplification in state overrides.** A caller could set up a chain of delegations across N EOAs that inflate storage reads per call. Bounded by Phase 8's per-call gas + global request caps; the new Phase 16 per-tx 16M cap is an additional bound.
- **Signature verification cost.** ~40-50μs per auth (ecdsa); 100 auths = 5ms. Negligible vs wall-clock timeout.
- **Auth signature replay across simulated blocks.** Each auth has a nonce, so replay is bounded by nonce increments; but test explicitly that auth N in block 1 cannot be replayed in block 2.
- **Invoke `/security-review` before merge** — new tx type + auth-list validation is a rich attack surface.

**Verification.**
- Go unit: `simulate/eip7702_test.go`:
  - Valid single-auth type-4 tx → delegation installed; call to authorizer's address reaches delegate.
  - Invalid auth signature + `validation=true` → top-level fatal.
  - Invalid auth nonce + `validation=true` → top-level fatal.
  - Delegation revocation (auth to `0x0000…`) → subsequent call reverts to EOA.
  - `validation=false` + invalid auth → call proceeds (consistent with non-validation relaxation).
  - Auth replay: same auth in two blocks — second must fail.
- System: `tests/system/test/SimulateV1_EIP7702.test.ts` — Hardhat end-to-end delegation.
- Port upstream spec conformance fixtures for 7702 once `execution-apis/tests/eth_simulateV1/` publishes them.

**DoD.**
- Type-4 tx round-trips end-to-end.
- Auth-list validation matches spec conformance.
- Security review clean.

---

### Phase 16 — EIP-7825 per-tx gas cap + `MaxUsedGas` response field

**Goal.** Add Osaka's per-tx gas cap (16,777,216) as an additional DoS layer. Add `MaxUsedGas` to `SimCallResult`.

**Design.**
- **Per-tx gas cap.** In `sanitizeCall` (Phase 6), after defaulting, assert `call.Gas <= 16_777_216`. Violation → structured error (await upstream code assignment; reserve slot in `-380xx` range).
- **`MaxUsedGas`.** Post-call, populate from the `ExecutionResult.MaxUsedGas` field introduced in geth v1.16.9 (PR #32789). Add to `SimCallResult` struct + JSON marshaling. This is the spec-mandated field in modern `eth_simulateV1` responses.

**Files.**
- EDIT `x/evm/keeper/simulate/sanitize.go` — add per-tx 16M gas cap check in `sanitizeCall`.
- EDIT `rpc/types/simulate.go` — add `MaxUsedGas hexutil.Uint64` field to `SimCallResult`.
- EDIT `x/evm/keeper/simulate/driver.go` — populate `MaxUsedGas` from `ExecutionResult`.
- EDIT `x/evm/types/simulate_v1_errors.go` — add per-tx cap violation code + `NewSim*` constructor.

**Security risks.** Negligible — the cap is a bound, not new surface.

**Verification.**
- Go unit: `simulate/dos_test.go` — new case: `call.Gas = 20_000_000` → structured error.
- Go unit: `simulate_marshal_test.go` — `MaxUsedGas` round-trips through JSON.
- System: extend `SimulateV1_Limits.test.ts` with the per-tx cap case.

**DoD.**
- Per-tx gas cap enforced at 16,777,216.
- `MaxUsedGas` appears in response, matching geth v1.16.9 shape.

---

## End-to-end verification strategy

Each phase's DoD is binary; but across the whole feature:

1. **Go unit tests** (primary) — keeper internals, pure functions, override semantics, tracer semantics, DoS guards. Run via `make test-unit`.
2. **Go backend tests** (`rpc/backend/simulate_test.go`) — mocks the query client, tests marshaling/timeout plumbing.
3. **Hardhat system tests** (`tests/system/test/SimulateV1_*.test.ts`) — hit a running localnet, exercise the full JSON-RPC stack end-to-end. Run via `tests/system/system-tests.sh`.
4. **Spec conformance** — port high-signal fixtures from `ethereum/execution-apis/tests/eth_simulateV1/` into Hardhat-compatible test cases in Phase 12.
5. **Fuzz** — Go fuzz target to guard against panic-level bugs (Phase 12).
6. **Manual localnet verification** — LAST RESORT, used only in Phases 7 + 8 where state-root edge cases or memory behavior cannot be reliably mocked.
7. **Security reviews** — kernel reviews invoked after Phases 7, 8, 10, and 15 (security-critical kernels); a final release-cut review is invoked at Phase 12. Uses the `/security-review` skill against the feature branch.

## Critical files (modified or created)

### Part 1 (v1.14.8)

- `x/evm/keeper/state_transition.go` (Phase 2 — unexported `evmBuilder` type + `buildEVM` parameter on `applyMessageWithConfig`, rewritten `SimulateMessage` with precompile-override closure; Phase 3 — `NewEVMWithOverrides`, `precompilesWithMoves`, `activePrecompiles`; `applyMessageWithConfig` now threads `*EVMOverrides` directly and the `evmBuilder` parameter is gone)
- `x/evm/keeper/state_override.go` (Phase 2 — `MovePrecompileTo` support; deny-list for mezo custom precompiles)
- `x/evm/keeper/config.go` (Phase 3 — `VMConfig` accepts optional `NoBaseFee` override)
- `x/evm/statedb/statedb.go` (Phase 3 — `FinaliseBetweenCalls`)
- `x/evm/keeper/grpc_query.go` (Phases 4, 5 — new `SimulateV1` handler added alongside `EthCall`)
- `x/evm/keeper/simulate_v1.go` (Phase 5 — unexported `simulateV1` driver + helpers)
- `x/evm/keeper/simulate/` (NEW package — Phases 4–11)
  - `input.go`, `sanitize.go`, `header.go`, `chain.go`, `assemble.go`, `driver.go`, `process_block.go`
- `x/evm/tracer/transfertracer/tracer.go` (Phase 9 — NEW package)
- `rpc/types/types.go` (Phase 2 — `OverrideAccount.MovePrecompileTo`)
- `rpc/types/simulate.go` (Phase 1 — NEW, spec-shaped JSON types)
- `x/evm/types/simulate_v1_errors.go` (Phase 1 — NEW, unified `SimError` + all `-380xx` codes)
- `rpc/types/simulate_marshal.go` (Phase 11 — NEW, `MarshalJSON` with `from` patching)
- `rpc/backend/backend.go` (Phase 1 — `SimulateV1` + `SimulateDisabled` on `EVMBackend`)
- `rpc/backend/simulate.go` (Phase 1 — NEW, backend adapter)
- `rpc/namespaces/ethereum/eth/simulate.go` (Phase 1 — NEW, RPC entry)
- `rpc/namespaces/ethereum/eth/api.go` (Phase 1 — add `SimulateV1` to interface)
- `server/config/config.go` (Phase 8 — `SimulateDisabled bool`)
- `proto/ethermint/evm/v1/query.proto` (Phase 4 — add `SimulateV1` RPC)
- `tests/system/test/SimulateV1_*.test.ts` (Phases 1, 2, 5, 6, 7, 8, 9, 10, 11, 12 — system tests)

### Part 2 (post-upgrade)

- `x/evm/keeper/state_override.go` (Phase 13 — `tracing.*ChangeReason` params on setters)
- `x/evm/keeper/state_transition.go` (Phase 13 — `NewEVM` signature update, `SetTxContext` insertions)
- `x/evm/statedb/statedb.go` (Phase 13 — REMOVE custom `FinaliseBetweenCalls`; rely on interface method)
- `x/evm/keeper/simulate/process_block.go` (Phase 14 — `ProcessParentBlockHash` pre-block call)
- `x/evm/keeper/simulate/driver.go` (Phases 13, 15 — `Finalise(true)` calls; EIP-7702 auth-list processing)
- `x/evm/keeper/simulate/input.go` (Phase 15 — `authorizationList` unmarshal)
- `x/evm/keeper/simulate/sanitize.go` (Phase 16 — per-tx 16M gas cap)
- `rpc/types/simulate.go` (Phase 16 — `MaxUsedGas` field on `SimCallResult`)
- `x/evm/types/simulate_v1_errors.go` (Phases 15, 16 — EIP-7702 auth errors, per-tx cap error)
- `tests/system/test/SimulateV1_EIP2935.test.ts`, `SimulateV1_EIP7702.test.ts` (Phases 14, 15 — NEW system tests)

## Consensus path — semantics unchanged (hard requirement)

**The consensus path MUST behave byte-identically to main.** This is the governing invariant for the whole feature: no phase may change the state transitions produced by a consensus-delivered transaction. Any observed behavioral delta on the consensus path is a blocking bug, regardless of how "obviously equivalent" the refactor looked.

**Never modified — do not edit under any circumstance:**
- `app/ante/evm/*.go` — ante handler
- `x/evm/keeper/msg_server.go` — tx message server
- `x/evm/keeper/state_transition.go:185` — `ApplyTransaction`
- `x/evm/keeper/state_transition.go:319` — `ApplyMessage`

**Public signature preserved; internal body refactored but consensus behavior byte-identical.** Both of these were touched by Phases 2/3 to grow new keeper-internal seams. The rule above still holds — the consensus call path must produce the same state transitions it does today, and every phase's DoD requires all pre-existing keeper + ante tests to pass unchanged as the regression guard:
- `x/evm/keeper/state_transition.go:370` — `ApplyMessageWithConfig`. Phase 2 added an unexported `evmBuilder` parameter on the internal `applyMessageWithConfig`; the public wrapper passes `k.NewEVM` so consensus callers hit the same construction path. Phase 3 swapped the `evmBuilder` parameter for a `*EVMOverrides` parameter directly; the public wrapper passes `nil` — semantics identical.
- `x/evm/keeper/state_transition.go:386` — `SimulateMessage`. Phase 2 rewrote the body to collect `MovePrecompileTo` moves and pass a precompile-override closure through `applyMessageWithConfig`. `eth_call` / `eth_estimateGas` callers see no new required fields; with zero overrides the execution path is identical to pre-Phase-2.
- **`EVMOverrides` is a pointer argument** on `NewEVMWithOverrides` and the internal `applyMessageWithConfig` (deviation from earlier drafts that passed the struct by value). `NewEVM` and `ApplyMessageWithConfig` delegate with `nil`, so the consensus path carries neither a zero-value struct allocation nor an `EVMOverrides{}` literal — it is the compiler-verified "no overrides" case.

## Known divergences from the execution-apis spec (documented to users)

### Part 1 (v1.14.8, Cancun)

1. **EIP-4844 / 4788 / 2935 / 7685 not supported.** Overrides for `BeaconRoot`, `Withdrawals`, blob gas fields are rejected.
2. **Custom mezo precompiles immovable.** `MovePrecompileTo` for any of the 8 addresses at `0x7b7c…` returns a structured `-32602` error (spec does not assign a dedicated -38xxx code to this policy rejection; geth uses the same mapping for "source is not a precompile").
3. **`GasUsed` honors `MinGasMultiplier`.** Reported gas matches mezod on-chain receipts, not go-ethereum's raw EVM gas. Documented for callers comparing across chains.

### Part 2 (post-upgrade, Prague + Osaka)

- **Divergence (1) narrows.** EIP-2935 (Phase 14) and EIP-7702 (Phase 15) become supported. **EIP-4844, EIP-4788, EIP-7685, EIP-6110, EIP-7002, EIP-7251 stay rejected permanently** because mezod has no data-availability layer, no beacon chain (uses CometBFT), and no EL↔CL messaging framework. Rejection reason text updated in the API response to reflect the mezo-specific rationale (not "EIP inactive" but "mezod chain model does not include [beacon chain / DA layer / validator queues]").
- **BLOCKHASH range extends to 8192.** EIP-2935's parent-hash state contract serves the 257..8192 canonical range in simulated blocks, lifting the standard EVM 256-block `BLOCKHASH` cap. Zero-hash fallback only applies for `BLOCKHASH(base - N)` where `N > 8192`.
- **Divergences (2) and (3) unchanged.** Custom mezo precompiles stay immovable; `MinGasMultiplier` gas reporting continues.

## Follow-ups / out of scope

- **EIP-4844 blob-tx simulation** — mezo chain policy rejects blob txs; not a simulate problem to solve.
- **EIP-4788 / EIP-7685 support** — mezo has no beacon chain (uses CometBFT) and no EL↔CL messaging framework. Supporting these would require chain-level architecture changes first; a simulator can't fake them into existence.
- **EIP-6110 / EIP-7002 / EIP-7251 validator queues** — mezo's validator operations go through `x/poa` (PoA set) and `x/bridge` (BTC bridging), not an EL↔CL deposit/withdrawal/consolidation queue. Out of scope structurally.
- **Relaxing custom-precompile `MovePrecompileTo` restriction** — requires per-precompile safety audit, especially for `BTCToken` (0x7b7c…00), `AssetsBridge` (0x7b7c…12), and `ValidatorPool` (0x7b7c…11) which interact with Cosmos modules outside the EVM state.
- **Richer per-feature DoS config** (`SimulateGasCap`, `SimulateEVMTimeout`, `SimulateMaxBlocks`) if operational experience shows the shared-with-`eth_call` knobs are too coarse.
- **Streaming / paginated responses** for very large simulations — spec doesn't support this today.
