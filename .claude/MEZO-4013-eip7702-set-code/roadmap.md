> **Note:** This is the phased delivery plan for MEZO-4013 (EIP-7702 Set
> Code Transactions). The canonical source of truth for **what** mezod
> ships is [`spec.md`](./spec.md) — this roadmap describes **how** we get
> there, sequenced into mid-size PRs. Each phase below is sized to one
> reviewable PR.

# Roadmap — EIP-7702 Set Code Transactions

## Current progress

- **Phase 1.** `prague_time` plumbed through `ChainConfig`, plus the
  eight additional fork-time placeholders bundled in via MEZO-4007.
  PR: [#673](https://github.com/mezo-org/mezod/pull/673).
- **Phase 2.** `SetCodeTx` and `SetCodeAuthorization` proto messages
  plus the Cosmos `TxData` implementation, with stateless validation
  and a pre-Prague ante gate.
  PR: [#674](https://github.com/mezo-org/mezod/pull/674).
- **Phase 3.** Authorization processing in the EVM state transition:
  intrinsic-gas wiring for `SetCodeAuthorizations`, reimplemented
  `validateAuthorization`/`applyAuthorization` mirroring geth
  `v1.16.9`, and exposure of the auth list on the `TxData` interface.
  PR: [#680](https://github.com/mezo-org/mezod/pull/680).
- **Phase 4.** Ante handler EIP-3607 exemption: delegated EOAs
  (accounts whose code parses as a valid `0xef0100‖target` designator)
  bypass the contract-coded-sender reject in
  `EthAccountVerificationDecorator` while genuine contract code is
  still rejected.
  PR: [#681](https://github.com/mezo-org/mezod/pull/681).
- **Phase 5.** JSON-RPC surface: `TransactionArgs` gains
  `authorizationList`, the formatting layer (`NewRPCTransaction`,
  `RPCTransaction`) emits the field and the type-`0x04` envelope on
  subscriptions and historical lookups, receipts surface
  `effectiveGasPrice` for SetCodeTx via the existing
  `TxData.EffectiveGasPrice` interface, and the keeper's
  `applyMessageWithConfig` gates `applySetCodeAuthorizations` on
  `rules.IsPrague` so simulate / `eth_call` / `eth_estimateGas` match
  consensus on a pre-Prague chain. End-to-end coverage via a Hardhat
  `Eip7702SendRawTx` system test.
  PR: [#683](https://github.com/mezo-org/mezod/pull/683).

## Status of prerequisites

The geth bump to `mezo-org/go-ethereum v1.16.9-mezo0` is on `main`
(merged via PR #671, commit `6fea1605`). All EIP-7702 machinery on the
geth side — `SetCodeTx`, `SetCodeAuthorization`,
`ParseDelegation`/`AddressToDelegation`, `PragueSigner`, the
`enable7702` jump-table mods, and `core.IntrinsicGas` with
`SetCodeAuthorizations` plumbing — is therefore present in the vendored
fork as of recent `main`. This roadmap assumes the implementing branch
is cut from a `main` HEAD that includes the merge.

`x/evm/keeper/{gas,fees}.go`, `x/evm/statedb/statedb.go`, and
`x/evm/types/tx_data.go` carry explicit `TODO (geth-upgrade)` markers at
exactly the integration points the spec calls out. These markers
disappear as the corresponding phases land.

## Sequencing

```
Phase 1: Fork-time plumbing (PragueTime in ChainConfig)
   └─ Phase 2: SetCodeTx proto + Cosmos TxData
        ├─ Phase 3: Intrinsic gas + state-transition authorization processing
        │     └─ Phase 4: Ante handler EIP-3607 exemption
        │           └─ Phase 5: JSON-RPC surface
        │                 └─ Phase 6: System tests + divergence tripwires
        │                       └─ Phase 7: Activation upgrade handler
        └─ (independent) `MEZO-4336 eth_simulateV1` Phase 15
              extends Phases 3+5 to the simulate driver
```

Phases 1–4 are non-functional from the user's perspective: nothing on the
RPC surface accepts type-`0x04` until Phase 5 lands. Phases 1–6 can ship
to a devnet (with `prague_time = 0` at genesis) before Phase 7 chooses
mainnet/testnet activation timestamps.

---

## Phase 1 — Plumb `prague_time` through `ChainConfig`

**Goal.** Make Prague a configurable fork on mezod the same way Shanghai
and Cancun are, defaulted to active at genesis.

**Scope.**
- Add `prague_time` field (proto field 24) to `ChainConfig` in
  `proto/ethermint/evm/v1/evm.proto`, mirroring `shanghai_time` and
  `cancun_time`.
- Wire the new field through `EthereumConfig()`, `DefaultChainConfig()`,
  and `Validate()` in `x/evm/types/chain_config.go` (genesis default is
  zero, i.e. active at genesis).
- Regenerate protobufs.
- Cover the new field in `chain_config_test.go` (validation, fork
  ordering, round-trip).

**Out of scope.** Any consumer of the new field; changes outside
`x/evm/types` and the proto tree.

**Why this is its own PR.** Pure plumbing, no behavior change, easy to
review, unblocks every later phase.

**References.** Spec §"Implementation summary" and §"Configuration".

**PR size estimate.** Small.

---

## Phase 2 — `SetCodeTx` Cosmos `TxData` + proto

**Goal.** Land the type-`0x04` transaction shape in mezod's
EVM-message protobuf and `TxData` interface so the rest of the stack
can carry it without translating to/from raw RLP at every layer.

**Scope.**
- Add `SetCodeTx` and `SetCodeAuthorization` proto messages to
  `proto/ethermint/evm/v1/tx.proto`, mirroring `DynamicFeeTx`.
  `SetCodeTx.to` is required (no contract creation); `auth_list` is
  required and non-empty.
- New file `x/evm/types/set_code_tx.go` implementing the `TxData`
  interface (return `ethtypes.SetCodeTxType`; conversion via
  `AsEthereumData`; `Validate()` enforcing non-nil `to` and non-empty
  `auth_list`; signature/fee math identical to `DynamicFeeTx`).
- Extend the `TxData` interface (or use a typed accessor) to expose
  the authorization list to consumers that need it.
- Register the new type in `NewTxDataFromTx` in `x/evm/types/tx_data.go`
  and any other tx-type switches under `x/evm/types/` that today
  enumerate `LegacyTxType`/`AccessListTxType`/`DynamicFeeTxType`.
- Unit tests under `x/evm/types/`: round-trip, validate, signing
  helpers, fee math.

**Out of scope.** Authorization processing inside the keeper; ante
handlers; RPC formatting; activation. Until Phase 3 lands, a `SetCodeTx`
that reaches the keeper still fails on intrinsic-gas wiring.

**Why this is its own PR.** Defines the on-the-wire shape and the
in-Cosmos shape independently of execution semantics; reviewers can
focus on encoding correctness.

**References.** Spec §"Implementation summary".

**PR size estimate.** Medium.

---

## Phase 3 — Authorization processing in the state transition

**Goal.** Make a type-`0x04` transaction validate, charge correct
intrinsic gas, apply its authorizations to state, and execute the call
in the delegated context.

**Scope.**
- Update `x/evm/keeper/gas.go` (`GetEthIntrinsicGas`) and
  `x/evm/keeper/fees.go` to pass `msg.SetCodeAuthorizations` into
  `core.IntrinsicGas`. Both call sites still pass `nil` for non-set-code
  paths.
- In `x/evm/keeper/state_transition.go` (`applyMessageWithConfig`),
  reimplement `validateAuthorization` and `applyAuthorization` mirroring
  geth `v1.16.9` (`core/state_transition.go:577,608`). Pin the
  reimplementation to that upstream commit by comment so future bumps
  re-audit it.
  - Validation rules: chain-id must be `0` or current; nonce must match
    authority's current nonce; nonce + 1 must not overflow; signature
    must recover; signer must not have non-delegation contract code;
    invalid tuples are silently skipped.
  - Application: bump authority nonce; install `0xef0100 || target` (or
    clear when `target == 0x0`); warm the delegation target in the
    access list; account for the per-existing-account intrinsic-gas
    refund.
- Run authorization processing after intrinsic-gas charging and before
  `evm.Call`, against the same `*statedb.StateDB` instance the call uses.
- Drop the `TODO (geth-upgrade)` markers in `gas.go` and `fees.go`.
- Differential keeper tests: validation rules, application side-effects,
  silent-skip of invalid tuples, refund accounting, target warming.

**Out of scope.** Ante handler exemption; RPC; user-visible activation.
This phase makes the keeper correct; ante still rejects delegated
senders until Phase 4.

**Why this is its own PR.** Highest-risk change in the project. The
phase touches consensus-critical paths, introduces a new signature
recovery surface (the per-auth secp256k1 verification), and writes to
account code via the keeper for the first time outside contract
deployment. Reviewing it in isolation is the safest path. Run
`/security-review` before merge.

**References.** Spec §"Implementation summary",
§"Mezo-specific divergences" item 2, §"Key decisions".

**PR size estimate.** Large (mid-size at the upper end).

---

## Phase 4 — Ante handler EIP-3607 exemption

**Goal.** Stop the existing `EthAccountVerificationDecorator` from
rejecting delegated EOAs.

**Scope.**
- In `app/ante/evm/eth.go`, replace the unconditional
  `acct.IsContract()` reject with: reject only if the account has code
  AND the code is not a delegation designator. The check is inlined
  inside `EthAccountVerificationDecorator.AnteHandle` using
  `evmKeeper.GetCode` + `ethtypes.ParseDelegation`; no helper is added
  to `statedb.Account` because the exemption is purely an ante-layer
  concern. A shared helper can be extracted in Phase 5 if RPC
  formatting needs the same detection.
- Audit the rest of `app/ante/` and the cosmos-side decorators for
  other assumptions about EOA-vs-contract on the sender. Outcome:
  the only sender EOA/contract assumption in the ante chain is the
  decorator we're modifying; no other carve-outs needed.
- Unit tests for the new branch (delegated sender accepted; balance
  shortfall on a delegated sender to pin that the exemption only
  bypasses the EOA check). Contract-sender rejection and uncoded
  sender behavior are already covered by the existing
  `TestNewEthAccountVerificationDecorator` cases.

**Out of scope.** Delegated-account behavior beyond the sender check;
RPC.

**Why this is its own PR.** Tight, single-decorator change with a
focused security review surface.

**References.** Spec §"Implementation summary" (ante handler exemption
paragraph), §"Key decisions" item on EIP-3607.

**PR size estimate.** Small.

---

## Phase 5 — JSON-RPC surface

**Goal.** Make every public RPC entry point understand
type-`0x04`: accept it on inbound, surface it on outbound.

**Scope.**
- `x/evm/types/tx_args.go`: add `AuthorizationList` to
  `TransactionArgs`; thread through `ToTransaction()` and the
  `String()` formatter; surface the field in any `MsgEthereumTx`
  conversion path.
- `rpc/types/utils.go`: extend `NewRPCTransaction` and
  `NewTransactionFromMsg` switches with an `ethtypes.SetCodeTxType`
  case that emits `authorizationList` and the type byte.
- `rpc/backend/sign_tx.go` and `rpc/backend/call_tx.go`: confirm
  signer selection picks up `PragueSigner` (already returned by
  `MakeSigner`/`LatestSignerForChainID` once Prague is active) and
  exercise type-`0x04` paths.
- KV indexer (`indexer/kv_indexer.go`): make sure `SetCodeTx`
  transactions index alongside the existing types.
- Verify `eth_sendRawTransaction` end-to-end against a local node with
  `prague_time = 0`.
- RPC-layer unit tests: marshal/unmarshal of `authorizationList`,
  type-`0x04` formatting, raw-tx submission flow.

**Out of scope.** `eth_simulateV1` validation-mode handling for auth
lists (that lives in `MEZO-4336` Phase 15); `eth_signTransaction` is
in scope only insofar as the existing implementation already handles
signer dispatch — no Mezo-specific signing UX.

**Why this is its own PR.** RPC plumbing is mostly mechanical but
touches several files; landing it together avoids a half-public surface
where some endpoints accept the type and others don't.

**References.** Spec §"Implementation summary" (JSON-RPC paragraph).

**PR size estimate.** Medium.

---

## Phase 6 — System tests and divergence tripwires

**Goal.** Lock down end-to-end behavior under TypeScript system tests
and ensure regressions are visible.

**Scope.**
- New suite `tests/system/test/Eip7702Delegation.test.ts`:
  install delegation, call delegated EOA, observe storage change in
  EOA's slot, rotate delegation, clear delegation.
- New suite `tests/system/test/Eip7702Gas.test.ts`: per-auth
  intrinsic gas (new account vs. existing), refund accounting, gas
  cost of CALL into a delegated EOA (cold vs. warm).
- New suite `tests/system/test/Eip7702Security.test.ts`: replay
  protection, cross-chain id rejection (non-zero, non-current),
  invalid-signature silent skip, EIP-3607 exemption for delegated
  senders, EXTCODECOPY/EXTCODESIZE/EXTCODEHASH return raw 23-byte
  designator.
- New suite `tests/system/test/Eip7702MezoDivergence.test.ts`:
  pin every spec divergence (no mempool authority reservation;
  reimplemented validate/apply path in keeper; no Mezo-specific
  signer adapter).
- Keeper-level fuzz target on the `SetCodeTx` proto unmarshaler.
- Add a Prague section to `docs/evm-compatibility.md` following the
  Cancun template, listing EIP-7702 as supported with a back-link to
  `spec.md`.

**Out of scope.** Production activation and any change to live chain
config.

**Why this is its own PR.** Tests-only PR after the implementation has
stabilized; reviewers can focus on assertion quality and tripwire
coverage, not behavior. Routes through the `test-creator` agent per
local convention.

**References.** Spec §"Conformance with the EIP",
§"Mezo-specific divergences".

**PR size estimate.** Medium-Large.

---

## Phase 7 — Activation upgrade handler

**Goal.** Activate Prague on the live mainnet and testnet networks at
chosen unix timestamps without breaking historical block replay on
existing nodes.

**Scope.**
- Create the next `app/upgrades/vN_0/` package (numbered after the
  current latest, currently `v9_0`). The handler writes the chosen
  `PragueTime` into the EVM module's stored `ChainConfig` params and
  runs `mm.RunMigrations` for any incidental module-version bumps.
- Wire the new `Upgrade` value into `app/upgrades.go`.
- A simple unit test that invokes the handler against a fresh keeper
  and verifies the stored chain config exposes the expected
  `PragueTime`.
- Coordinate timestamp choice with operations; document the chosen
  values per network in `docs/upgrades.md` alongside the existing
  historical table.

**Out of scope.** Behavior changes; everything else this scope ships
must already be live in the binary by the time Prague activates.

**Why this is its own PR.** Activation is an operational decision
distinct from feature delivery; isolating it lets us merge Phases 1–6
on cadence and gate the live-chain switch on its own review.

**References.** Spec §"Configuration" (upgrade-driven activation),
§"Key decisions" (genesis default vs. live-chain rollout).

**PR size estimate.** Small.

---

## Cross-references

- [`spec.md`](./spec.md) — canonical behavior. Every phase references it
  rather than restating decisions inline.
- [`../MEZO-4336-eth-simulate-v1-geth116/plan.md`](../MEZO-4336-eth-simulate-v1-geth116/plan.md)
  — Phase 15 of that plan adds simulate-specific behavior **on top of**
  Phases 3 + 5 of this roadmap. The relationship is:
    - Phase 3 lands keeper-level authorization processing inside
      `applyMessageWithConfig`. Because the simulate driver routes
      every call through `applyMessageWithConfig`, EVM-level
      delegation install/clear, target warming, and intrinsic-gas
      accounting work in `eth_simulateV1` automatically as soon as
      Phase 3 ships — Phase 15 does not redo this.
    - Phase 5 adds `AuthorizationList` to `TransactionArgs`, the
      shared input shape that simulate's `SimBlock.calls[]` extends.
      Once Phase 5 ships, simulate's call shape carries auth lists
      end-to-end without further wiring — Phase 15 does not
      rebuild this either.
    - What Phase 15 does add is genuinely simulate-specific: per-call
      authorization validation under `validation: true` (normal tx
      flow silently skips invalid auths; simulate's validation mode
      promotes them to top-level fatal errors), structured
      `*SimError` codes for invalid-auth cases in
      `simulate_v1_errors.go`, and recognition of the `0xef0100`
      delegation prefix in `OverrideAccount.Code` overrides.
  Phase 15 is sequenced after this scope ships.

## Out of scope for this initiative

- Mempool hardening (single-slot delegation limit, authority
  reservation across in-flight txs). Per spec divergence #1.
- Folding `applyMessageWithConfig` onto `core.ApplyMessage`. Per spec
  §"Key decisions" (reimplement, don't refactor).
- New per-node configuration for EIP-7702 enable/disable. Per spec
  §"Configuration".
- Storage-clearing extensions when a delegation is cleared. Slots
  written by a prior delegate persist after the delegation is cleared
  and remain readable by a subsequent re-delegation of the same
  authority.
- `eth_simulateV1` validation-mode handling for auth lists. Tracked in
  `MEZO-4336` Phase 15.
