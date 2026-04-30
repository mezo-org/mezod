> **Disclaimer:** Temporary implementation file for MEZO-4336 (`eth_simulateV1`,
> post-upgrade Prague + Osaka). Remove once the feature is complete.

# Implementation plan: `eth_simulateV1` post-geth-v1.16.9 phases

**Status.** Continuation of MEZO-4227 (Phases 1-12, against geth v1.14.8 / Cancun).
Phases 1-12 shipped: see [#658](https://github.com/mezo-org/mezod/pull/658) (scaffold),
[#660](https://github.com/mezo-org/mezod/pull/660) (`MovePrecompileTo`),
[#662](https://github.com/mezo-org/mezod/pull/662) (keeper seams + proto + single-call),
[#664](https://github.com/mezo-org/mezod/pull/664) (multi-call + multi-block + simulated `GetHashFn`),
[#665](https://github.com/mezo-org/mezod/pull/665) (DoS guards + kill switch + request-level fatal lifts),
[#667](https://github.com/mezo-org/mezod/pull/667) (`traceTransfers`, `validation=true`,
`returnFullTransactions` + full block envelope), and the Phase 12 close-out PR
(spec-conformance + divergence pinning + fuzzing + docs).

This plan covers the post-upgrade phases (13-16) that pick up Prague/Osaka behavior
once the geth v1.14.8 → v1.16.9 chain upgrade lands.

## Context

`eth_simulateV1` is already live on the v1.14.8 branch with full feature parity to the
`ethereum/execution-apis` Cancun-era spec. The remaining work is mechanical port +
selective Prague/Osaka feature pickup. Reference implementation: go-ethereum v1.16.9
`internal/ethapi/simulate.go`.

**⚠ Blocked on** the
[geth v1.16 upgrade project](https://linear.app/thesis-co/project/chain-geth-v116-upgrade-and-osaka-fork-compatibility-b08591b25fb5)
merging to `main`. Target: 2026-05-15. Do not start Phase 13 before the upgrade merges.

## Decisions carried forward from MEZO-4227

| Decision | Choice |
|---|---|
| Architectural seam | Bare types in `x/evm/types/`, flow logic in `x/evm/keeper/`. No `simulate/` sub-package — driver and helpers continue to live in `x/evm/keeper/simulate_v1.go` |
| Error handling | Single typed `*SimError{Code, Message, Data}` end-to-end. Catalog in `x/evm/types/simulate_v1_errors.go`. Riding gRPC on a dedicated `SimError error = 2` field of `SimulateV1Response`. Genuine internals collapse to `status.Error(codes.Internal, …)` |
| `MovePrecompileTo` for custom mezo precompiles (`0x7b7c…`) | Blocked — stdlib precompiles only. Custom rejected with structured `-32602` |
| DoS config | Kill switch only: `JSONRPCConfig.SimulateDisabled`; reuses existing `RPCGasCap` + `RPCEVMTimeout`; hard-coded 256-block + 1000-call envelope caps |
| Gas numerics | mezod-native — reported `GasUsed` honors `MinGasMultiplier` (matches on-chain receipts) |
| System test layout | Exactly two files: `tests/system/test/SimulateV1_SpecCompliance.test.ts` (spec-conformance) and `tests/system/test/SimulateV1_MezoDivergence.test.ts` (divergences). Phases 14/15 extend these — do NOT add per-EIP `SimulateV1_*.test.ts` files |

## Scope discipline

Prague/Osaka activates many EIPs simultaneously. We pick up only those that fit mezo's
chain model:

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
| EIP-4895 validator withdrawals (`BlockOverrides.Withdrawals`) | **No** — no validator-withdrawal queue on mezod | Continue rejecting `BlockOverrides.Withdrawals` at parse time |
| EIP-7685 EL requests + EIP-6110/7002/7251 (deposits/exits/consolidations) | **No** — no EL↔CL messaging; validator ops via `x/poa`+`x/bridge` | Skip post-block `ProcessWithdrawalQueue`/`ProcessConsolidationQueue`/`ParseDepositLogs`; `RequestsHash` stays nil |

The port cost measured against `git diff v1.14.8..v1.16.9` on the interfaces simulate
touches is ~15 mechanical lines (net -10 LOC after replacing the custom
`FinaliseBetweenCalls` helper with geth's new `StateDB.Finalise(true)`).

## Phase 13 — Port simulate to v1.16.9 interfaces (mechanical)

**Goal.** Update call sites where v1.16.9's signatures differ from v1.14.8's. Pure
mechanical edits. ~15-20 lines modified, ~10 lines deleted net.

**What changes** (measured from `git diff v1.14.8..v1.16.9` on `core/vm/evm.go`,
`core/vm/contracts.go`, `core/vm/interface.go`, `core/state/statedb.go`,
`core/state_transition.go`):

| Interface | v1.14.8 → v1.16.9 change | Simulate-code fix |
|---|---|---|
| `vm.NewEVM` | drops `TxContext` param | Update `NewEVMWithOverrides` (`x/evm/keeper/state_transition.go`); call `evm.SetTxContext(core.NewEVMTxContext(msg))` separately where TxContext was passed |
| `vm.StateDB.SetNonce` | gains `tracing.NonceChangeReason` param | `applyStateOverrides` (`x/evm/keeper/state_override.go`) + driver helpers: pass `tracing.NonceChangeUnspecified` |
| `vm.StateDB.SetCode` | gains `tracing.CodeChangeReason` param, returns prev code | Same; ignore return |
| `vm.StateDB.Finalise(bool)` | **NEW on interface** | **Simplification**: remove the custom `FinaliseBetweenCalls()` helper from `x/evm/statedb/statedb.go`; call `stateDB.Finalise(true)` (matches geth's own `simulate.go:299-303`). Saves ~20 lines |
| `evm.Call`/`Create` first param | `ContractRef` → `common.Address` | Simulate invokes `core.ApplyMessage`, not these directly; no fix |
| `core.IntrinsicGas` | gains `authList []types.SetCodeAuthorization` param | Absorbed by `k.GetEthIntrinsicGas` keeper wrapper (updated by upgrade project) |
| `vm.PrecompiledContract` | gains `Name() string` method | Mezo custom precompiles get `Name()` via upgrade project |
| `ExecutionResult.RefundedGas` | renamed to `MaxUsedGas` | Handled in Phase 16 |

**Files.**
- EDIT `x/evm/keeper/state_override.go` — add `tracing.*ChangeReason` params to affected setters.
- EDIT `x/evm/keeper/state_transition.go` — update `NewEVMWithOverrides` to the new
  `NewEVM` signature; insert `evm.SetTxContext(...)` calls. In `NewEVM`, swap the inline
  clone-and-layer precompile build for `evm.WithCustomPrecompiles(k.customPrecompiles, ...)`.
  In `SimulateMessage`, replace the duplicated precompile-registry rebuild with
  `precompiles := evm.Precompiles()` (live map), apply moves, `evm.SetPrecompiles(precompiles)` —
  the explicit `evm.WithPrecompiles(...)` re-attach goes away.
- EDIT `x/evm/statedb/statedb.go` — **remove** `FinaliseBetweenCalls` **only after the
  verification gate below**.
- EDIT `x/evm/keeper/simulate_v1.go` — replace `stateDB.FinaliseBetweenCalls()` call
  sites with `stateDB.Finalise(true)`.

**⚠ VERIFY BEFORE DELETING `FinaliseBetweenCalls`.** The helper does two things:
(a) standard finalise (clear logs/refund/transientStorage, preserve stateObjects), and
(b) reset mezod's custom `ongoingPrecompilesCallsCounter`. Geth's new
`StateDB.Finalise(true)` covers (a). Whether it also performs (b) depends on how
mezod's StateDB override of `Finalise` is written on the upgrade branch. Before
removing the helper:
1. Read mezod's `Finalise(true)` impl on the post-upgrade branch.
2. If it resets `ongoingPrecompilesCallsCounter`, remove the helper as planned.
3. If it does NOT, either fold the counter reset into mezod's `Finalise` override,
   or keep a thin wrapper that resets the counter and then calls `Finalise(true)`.

Skipping this check will silently break any simulate request that exceeds
`maxPrecompilesCallsPerExecution` across call boundaries.

**Risks.** None new in the type-safe sense — purely mechanical — but see the
counter-reset gate above.

**Verification.**
- All MEZO-4227 (Phase 1-12) tests pass unchanged.
- `TestSimulateV1_MultiBlock_PrecompileStateChains` (`x/evm/keeper/grpc_query_test.go`)
  + the system-side btctoken state-chain cases in `SimulateV1_MezoDivergence.test.ts`
  still pass — canary for the counter-reset gap.
- `go build ./...` clean; `make test-unit` green.

**DoD.**
- Simulate compiles clean against v1.16.9.
- All MEZO-4227 (Phase 1-12) behavior tests green.
- No functional delta.

---

## Phase 14 — EIP-2935 parent-hash state contract

**Goal.** Post-Prague, `BLOCKHASH` can be served from the system contract at
`0x…fffffffffffffffffffffffffffffffffffffffe` for up to the last 8192 blocks. Simulate
must invoke `core.ProcessParentBlockHash` at the top of each simulated block (matches
go-ethereum `simulate.go:267-272`) so BLOCKHASH works across the full 1..8192 range.

**Design.** In `processSimBlock`, after EVM construction and before executing user calls:

```go
if cfg.ChainConfig.IsPrague(header.Number, header.Time) {
    core.ProcessParentBlockHash(header.ParentHash, evm)
}
```

The existing `newSimGetHashFn` closure stays — it covers the `[base, base+N]`
simulated-sibling range that the parent-hash contract cannot serve. Post-Prague split:
- `height > base` (simulated siblings) — `newSimGetHashFn` from in-memory headers.
- `height == base` — `newSimGetHashFn`.
- `height ∈ [base-256, base-1]` (recent canonical) — EVM `BLOCKHASH` opcode via
  `GetHashFn` delegating to `k.GetHashFn(ctx)`.
- `height ∈ [base-8192, base-257]` (older canonical) — parent-hash contract state.
- `height < base-8192` — zero hash.

**Files.**
- EDIT `x/evm/keeper/simulate_v1.go` — add Prague-gated `ProcessParentBlockHash` call
  in `processSimBlock`.

**Risks.**
- **Fork-gate correctness.** Use `cfg.ChainConfig.IsPrague(...)`; firing pre-Prague
  produces nonsensical state writes.
- **No divergence** with real block processing — the upgrade project adds the same
  call to `ApplyTransaction`; we mirror.

**Verification.**
- `x/evm/keeper/simulate_v1_test.go`: `TestProcessSimBlock_Prague_BlockHashRange` —
  `BLOCKHASH(base - N)` for N = 100, 500, 5000, 9000 → first three return real hashes,
  last returns zero.
- System: extend `tests/system/test/SimulateV1_SpecCompliance.test.ts` with an EIP-2935
  scenario — multi-block simulate; inside block 3 read `BLOCKHASH(base - 1000)`;
  cross-check against `eth_getBlockByNumber(base - 1000).hash`. Update the file's
  top-of-file scenario table.

**DoD.**
- BLOCKHASH 257..8192 range works in simulated blocks (lifting the standard 256-block
  cap for Prague-activated simulations).

---

## Phase 15 — EIP-7702 SetCode transactions ⚠️ SECURITY-CRITICAL

**Goal.** Accept type-4 (SetCode) transactions in `calls[]`. Handle delegation-prefix
(`0xef0100…`) state overrides correctly. Validate authorization lists when
`validation=true`.

**Depends on** the upgrade project's "EIP-7702 SetCode transaction support" scope item —
that lands Type-4 tx handling, authorization validation in ante handlers, and
delegation-prefix handling in `statedb.StateDB`. Simulate extends the new machinery;
we don't build it from scratch.

**Design.**
- **Input.** `TransactionArgs.AuthorizationList` is populated by the upgrade project.
  Simulate's JSON unmarshal passes it through unchanged; `call.ToMessage` at the keeper
  level absorbs it.
- **Validation mode.** When `validation=true`, validate each auth per EIP-7702:
  `chainID ∈ {0, chain.ID}`, nonce matches current state, signer not a contract
  (unless already delegated), signature recoverable. Any invalid auth → top-level fatal
  with new structured code (await upstream assignment; add to `simulate_v1_errors.go`).
- **State overrides + delegation.** `OverrideAccount.Code` set to `0xef0100` + 20-byte
  address is a delegation. `applyStateOverrides` passes through unchanged — mezod's
  upgraded StateDB handles the prefix semantics.
- **Cross-call nonce consistency.** Auth nonces reference current state; between calls
  in a simulated block, nonce advances. Validation must consult the shared StateDB,
  not a snapshot.

**Files.**
- EDIT `x/evm/keeper/simulate_v1.go` — recognize `authList` in the call loop; invoke
  per-call auth validation when `validation=true`.
- EDIT `x/evm/types/simulate_v1.go` — allow `authorizationList` in `SimBlock` calls
  JSON unmarshal.
- EDIT `x/evm/types/transaction_args.go` (or equivalent owned by the upgrade project) —
  surface `AuthorizationList` in the serializable call-args shape if not already from
  the upgrade.
- EDIT `x/evm/types/simulate_v1_errors.go` — add EIP-7702 auth-invalid error codes +
  `NewSim*` constructors.

**Risks.**
- **Delegation amplification in state overrides.** A caller could chain delegations
  across N EOAs to inflate storage reads per call. Bounded by the existing 256-block /
  1000-call envelope caps + the new Phase 16 per-tx 16M cap.
- **Signature verification cost.** ~40-50μs per auth (ecdsa); 100 auths = 5ms.
  Negligible vs wall-clock timeout.
- **Auth signature replay across simulated blocks.** Each auth has a nonce; replay
  bounded by nonce increments; test explicitly that auth N in block 1 cannot be
  replayed in block 2.
- **`/security-review` before merge** — new tx type + auth-list validation is a rich
  attack surface.

**Verification.**
- `x/evm/keeper/grpc_query_test.go`: `TestSimulateV1_EIP7702_ValidAuth` — single-auth
  type-4 tx → delegation installed; call to authorizer's address reaches delegate.
- `x/evm/keeper/grpc_query_test.go`: `TestSimulateV1_EIP7702_InvalidSig_Fatal` —
  invalid auth signature + `validation=true` → top-level fatal.
- `x/evm/keeper/grpc_query_test.go`: `TestSimulateV1_EIP7702_InvalidNonce_Fatal` —
  invalid auth nonce + `validation=true` → top-level fatal.
- `x/evm/keeper/grpc_query_test.go`: `TestSimulateV1_EIP7702_Revocation` — auth to
  `0x0000…` → subsequent call reverts to EOA.
- `x/evm/keeper/grpc_query_test.go`: `TestSimulateV1_EIP7702_NoValidationProceeds` —
  `validation=false` + invalid auth → call proceeds.
- `x/evm/keeper/grpc_query_test.go`: `TestSimulateV1_EIP7702_AuthReplayBlocked` — same
  auth in two blocks; second fails.
- System: extend `tests/system/test/SimulateV1_SpecCompliance.test.ts` with end-to-end
  EIP-7702 delegation scenarios; update its top-of-file scenario table.
- Port upstream spec conformance fixtures for 7702 once
  `ethereum/execution-apis/tests/eth_simulateV1/` publishes them.

**DoD.**
- Type-4 tx round-trips end-to-end.
- Auth-list validation matches spec conformance.
- Security review clean.

---

## Phase 16 — EIP-7825 per-tx gas cap + `MaxUsedGas` response field

**Goal.** Add Osaka's per-tx gas cap (16,777,216) as an additional DoS layer. Add
`MaxUsedGas` to `SimCallResult`.

**Design.**
- **Per-tx gas cap.** In `resolveSimCallGas`, after defaulting, assert
  `call.Gas <= 16_777_216`. Violation → structured error (await upstream code
  assignment; reserve slot in `-380xx` range).
- **`MaxUsedGas`.** Post-call, populate from the `ExecutionResult.MaxUsedGas` field
  introduced in geth v1.16.9 (PR #32789). Add to `SimCallResult` struct + JSON
  marshaling.

**Files.**
- EDIT `x/evm/keeper/simulate_v1.go` — per-tx 16M gas cap check in `resolveSimCallGas`;
  populate `MaxUsedGas` from `ExecutionResult`.
- EDIT `x/evm/types/simulate_v1.go` — add `MaxUsedGas hexutil.Uint64` to
  `SimCallResult`.
- EDIT `x/evm/types/simulate_v1_errors.go` — add per-tx cap violation code +
  `NewSim*` constructor.

**Risks.** Negligible — the cap is a bound, not new surface.

**Verification.**
- `x/evm/keeper/grpc_query_test.go`: `TestSimulateV1_PerTxGasCap` — `call.Gas =
  20_000_000` → structured error.
- `x/evm/types/simulate_v1_test.go`: `TestSimCallResult_MaxUsedGas_RoundTrip`.
- System: extend `tests/system/test/SimulateV1_SpecCompliance.test.ts` with a
  per-tx-cap and a `maxUsedGas` round-trip case; update its top-of-file scenario table.

**DoD.**
- Per-tx gas cap enforced at 16,777,216.
- `MaxUsedGas` appears in response, matching geth v1.16.9 shape.

---

# End-to-end verification strategy

1. **Go unit tests** — keeper internals, pure functions, override semantics, tracer
   semantics, DoS guards. Run via `make test-unit`.
2. **Go backend tests** (`rpc/backend/simulate_v1_test.go`) — mocks the query client.
3. **Hardhat system tests** — `tests/system/test/SimulateV1_SpecCompliance.test.ts`
   and `tests/system/test/SimulateV1_MezoDivergence.test.ts`. Full JSON-RPC stack
   against a running localnode. Run via `tests/system/system-tests.sh`.
4. **Spec conformance** — port any new
   `ethereum/execution-apis/tests/eth_simulateV1/` fixtures published for Prague/Osaka
   (EIP-2935, EIP-7702, EIP-7825) into the SpecCompliance file.
5. **Fuzz** — extend the existing `FuzzSimulateV1Opts` corpus with new fixtures.
6. **Manual localnet verification** — used as a sanity step at the close of each phase.
7. **Security reviews** — Phase 15 requires `/security-review` before merge.

# Known divergences from the execution-apis spec (post-upgrade)

Carried forward from MEZO-4227 with Prague/Osaka deltas applied:

1. **EIP-4844 (blob txs), EIP-4788 (parent beacon root), and EIP-4895
   (validator withdrawals — `BlockOverrides.Withdrawals`) stay rejected
   permanently** because mezod has no DA layer, no beacon chain (uses
   CometBFT), and no validator-withdrawal queue. The Prague-era requests
   EIPs — EIP-7685 (general EL requests) and the specific
   EIP-6110/7002/7251 (deposits, exits, consolidations) — are a separate
   forward-looking divergence: they require EL↔CL messaging that mezo
   does not have, since validator ops live in `x/poa`+`x/bridge`.
   Rejection reason text reflects mezo-specific rationale (not "EIP
   inactive" but "mezod chain model does not include [beacon chain / DA
   layer / validator queues]"). EIP-2935 (Phase 14) and EIP-7702
   (Phase 15) become supported.
2. **Custom mezo precompiles immovable.** `MovePrecompileTo` for any of the addresses
   at `0x7b7c…` returns a structured `-32602` error. Continues unchanged.
3. **`GasUsed` honors `MinGasMultiplier`.** Reported gas matches mezod on-chain
   receipts, not raw EVM gas. Continues unchanged.
4. **`stateRoot` is always the zero hash.** Mezod's `statedb.StateDB` wraps a Cosmos
   cached multistore and has no Merkle Patricia Trie, so there is no
   `IntermediateRoot()` to call after a simulated block executes. Pinned by
   `TestSimulateV1_StateRootZero_KeeperLayer` (`x/evm/keeper/grpc_query_test.go`) +
   `TestSimBlockResult_MarshalJSON_StateRootIsZero` (`x/evm/types/simulate_v1_test.go`)
   + the divergence-tripwire case in `SimulateV1_MezoDivergence.test.ts`. Continues
   unchanged.
5. **BLOCKHASH range extends to 8192 (post-Phase 14).** EIP-2935's parent-hash contract
   serves the 257..8192 canonical range in simulated blocks, lifting the standard
   256-block `BLOCKHASH` cap. Zero-hash fallback only for `N > 8192`.

# Follow-ups / out of scope

- **EIP-4844 blob-tx simulation** — chain policy rejects blob txs.
- **EIP-4788 / EIP-7685 support** — no beacon chain (CometBFT) and no EL↔CL framework.
  Requires chain-level architecture changes first.
- **EIP-6110 / EIP-7002 / EIP-7251 validator queues** — validator ops via `x/poa` and
  `x/bridge`, not EL↔CL.
- **Relaxing custom-precompile `MovePrecompileTo` restriction** — requires per-precompile
  safety audit, especially for `BTCToken` (`0x7b7c…00`), `AssetsBridge`
  (`0x7b7c…12`), and `ValidatorPool` (`0x7b7c…11`) which interact with Cosmos modules
  outside EVM state.
- **Richer per-feature DoS config** (`SimulateGasCap`, `SimulateEVMTimeout`,
  `SimulateMaxBlocks`) if operational experience shows shared-with-`eth_call` knobs are
  too coarse.
- **Streaming / paginated responses** for very large simulations — spec doesn't support
  this today.
