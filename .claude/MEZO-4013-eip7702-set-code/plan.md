> **Disclaimer:** This is a temporary file used during implementation of
> MEZO-4013 (EIP-7702 Set Code Transactions). It should be removed once the feature is
> complete.

# EIP-7702 Implementation Research for mezod

## Executive Summary

This document maps the EIP-7702 ("Set EOA Account Code") implementation from go-ethereum 1.16.x onto the mezod codebase (an Evmos fork). It identifies every component that needs changes, the nature of each change, dependencies between them, and risks.

**Current state**: mezod uses `github.com/mezo-org/go-ethereum v1.14.8-mezo2` (forked from go-ethereum 1.14.x). This fork does **not** include EIP-7702 support — it only defines transaction types 0x00–0x03 and has no `SetCodeTx`, `SetCodeAuthorization`, delegation designator, or Prague fork processing logic. However, it does have the `PragueTime` field in `params.ChainConfig` and the `IsPrague()` method.

**Two implementation paths** exist:
1. **Upgrade the geth fork to 1.16.x** (already planned) — this brings EIP-7702 for free at the geth layer; mezod still needs the application-layer changes documented below.
2. **Backport EIP-7702 into the current v1.14.8-mezo2 fork** — significantly more work, touching ~15 geth files. Not recommended.

This document assumes **path 1**: the geth fork upgrade completes first, and then mezod needs the application-layer integration described below.

---

## 1. Dependency Gap Analysis

### What geth 1.16.x provides (that v1.14.8-mezo2 lacks)

| Component | geth 1.16.x | v1.14.8-mezo2 |
|-----------|-------------|---------------|
| `SetCodeTxType` (0x04) constant | Yes | No |
| `SetCodeTx` struct | Yes | No |
| `SetCodeAuthorization` struct | Yes | No |
| `ParseDelegation()` / `AddressToDelegation()` | Yes | No |
| `core.IntrinsicGas` with auth list param | Yes (`authList []types.SetCodeAuthorization`) | No (only `accessList`) |
| `resolveCode()` / `resolveCodeHash()` in EVM | Yes | No |
| Authorization validation in state transition | Yes | No |
| `enable7702` jump table modifications | Yes | No |
| Delegation-aware gas costs (`operations_acl.go`) | Yes | No |
| `PragueSigner` (supports type 0x04) | Yes | No (`CancunSigner` is latest) |
| Txpool delegation/authority reservation | Yes | No |

### What mezod must add on top of geth 1.16.x

Even after the geth upgrade, mezod has its own **application layer** that wraps geth. The following sections describe every mezod-specific change needed.

---

## 2. Changes Required in mezod

### 2.1 Fork Configuration — `PragueTime` activation

**Files:**
- `proto/ethermint/evm/v1/evm.proto` (lines 163-173)
- `x/evm/types/chain_config.go` (lines 31-56, 59-99, 120-180)

**What to do:**
- Add `prague_time` field (field number 24) to the `ChainConfig` protobuf message, following the pattern of `shanghai_time` (field 22) and `cancun_time` (field 23).
- Update `EthereumConfig()` to map `cc.PragueTime` → `params.ChainConfig.PragueTime` using `getTimeValue()`.
- Update `DefaultChainConfig()` to include `PragueTime` (initially `nil` for not-yet-activated).
- Update `Validate()` to validate the new field and ensure fork ordering is correct.

**Proto change:**
```protobuf
// prague_time switch block (nil = no fork, 0 = already on prague)
string prague_time = 24 [
  (gogoproto.customtype) = "cosmossdk.io/math.Int",
  (gogoproto.moretags) = "yaml:\"prague_time\""
];
```

**Alternatives:**
- Could hardcode Prague activation at a specific block/time via an upgrade handler instead of making it a chain config parameter. **Trade-off**: Less flexible but simpler governance. The protobuf approach is consistent with how Shanghai and Cancun were activated.

---

### 2.2 New Transaction Type — `SetCodeTx` (type 0x04)

**Files:**
- `x/evm/types/tx_data.go` (lines 26-89)
- New file: `x/evm/types/set_code_tx.go`
- `proto/ethermint/evm/v1/tx.proto`

**What to do:**

**a) Protobuf definition** — Define a new `SetCodeTx` message in `tx.proto`:
```protobuf
message SetCodeTx {
  string chain_id = 1 [(gogoproto.customtype) = "cosmossdk.io/math.Int"];
  uint64 nonce = 2;
  string gas_tip_cap = 3 [(gogoproto.customtype) = "cosmossdk.io/math.Int"];
  string gas_fee_cap = 4 [(gogoproto.customtype) = "cosmossdk.io/math.Int"];
  uint64 gas_limit = 5;
  string to = 6;  // REQUIRED (non-nil)
  string amount = 7 [(gogoproto.customtype) = "cosmossdk.io/math.Int"];
  bytes data = 8;
  repeated AccessTuple access_list = 9 [(gogoproto.nullable) = false];
  repeated SetCodeAuthorization auth_list = 10 [(gogoproto.nullable) = false];
  bytes v = 11;
  bytes r = 12;
  bytes s = 13;
}

message SetCodeAuthorization {
  string chain_id = 1 [(gogoproto.customtype) = "cosmossdk.io/math.Int"];
  string address = 2;
  uint64 nonce = 3;
  uint32 v = 4;  // y-parity (0 or 1)
  bytes r = 5;
  bytes s = 6;
}
```

**b) Go implementation** — Create `set_code_tx.go` implementing the `TxData` interface:
- `TxType() byte` → returns `ethtypes.SetCodeTxType` (0x04)
- `GetAuthList()` — new method returning the authorization list
- `AsEthereumData()` → converts to `ethtypes.SetCodeTx`
- `Validate()` → same as `DynamicFeeTx` plus: `To` must be non-nil, `AuthList` must be non-empty
- All other `TxData` methods (fee calculation, signature, etc.) — same pattern as `DynamicFeeTx`

**c) Registration** — Update `NewTxDataFromTx()` switch in `tx_data.go` to handle `ethtypes.SetCodeTxType`:
```go
case ethtypes.SetCodeTxType:
    txData, err = NewSetCodeTx(tx)
```

**d) Interface extension** — Consider adding `GetAuthList()` to the `TxData` interface, or handle it via type assertion where needed.

**Alternatives:**
- Could avoid the protobuf definition entirely and store SetCodeTx as opaque bytes, relying on geth's RLP encoding. **Trade-off**: Breaks Cosmos SDK's protobuf-first approach; indexers and explorers expecting protobuf would not see authorization data.
- Could extend `DynamicFeeTx` with an optional auth list field instead of a new type. **Trade-off**: Cleaner separation favors a new type.

---

### 2.3 Intrinsic Gas Calculation

**Files:**
- `x/evm/keeper/gas.go` (lines 34-42)

**What to do:**

The `GetEthIntrinsicGas` function currently calls:
```go
core.IntrinsicGas(msg.Data, msg.AccessList, isContractCreation, homestead, istanbul, isShanghai)
```

In geth 1.16.x, `core.IntrinsicGas` gains a new parameter for authorization tuples:
```go
core.IntrinsicGas(msg.Data, msg.AccessList, msg.AuthList, isContractCreation, homestead, istanbul, isShanghai)
```

Update the call to pass `msg.AuthList` (which will be `nil` for non-SetCode transactions). Also need to check if `IsPrague` is needed as a parameter.

**Note**: The authorization gas model charges 25,000 (`CallNewAccountGas`) per auth tuple, with a 12,500 refund for already-existing accounts. This is handled inside geth's `IntrinsicGas` and `applyAuthorization`, but mezod must pass the data correctly.

---

### 2.4 State Transition — Authorization Processing

**Files:**
- `x/evm/keeper/state_transition.go` (lines 367-519 — `ApplyMessageWithConfig`)

**What to do:**

In geth 1.16.x, authorization processing happens **inside** the `StateTransition.execute()` method (after intrinsic gas, before EVM call). Since mezod reimplements parts of the state transition, it needs to either:

**Option A — Delegate to geth's state transition entirely:**
Replace the manual state transition in `ApplyMessageWithConfig` with a call to geth's `core.ApplyMessage()` or `core.NewStateTransition().Execute()`. This would automatically handle authorization processing.

- **Pros**: Less custom code, automatic compatibility with future EIPs
- **Cons**: Major refactor; mezod's state transition has Cosmos-specific logic (custom gas metering, precompile handling, min gas multiplier) that doesn't exist in geth

**Option B — Add authorization processing in mezod's state transition:**
After intrinsic gas calculation and before `evm.Call()`, add authorization processing logic:

```go
// After line 448 (access list preparation)
if rules.IsPrague {
    for _, auth := range msg.AuthList {
        // 1. Validate authorization (chain ID, nonce, signature, code checks)
        // 2. Apply authorization (set delegation code on authority account)
        // 3. Warm the delegation target in access list
    }
}
```

- **Pros**: Minimal disruption to existing code
- **Cons**: Duplicates geth logic; must be kept in sync with upstream changes

**Option C (Recommended) — Leverage geth's state transition for SetCode txs:**
For type 0x04 transactions specifically, use geth's `StateTransition` to handle the authorization processing, then continue with mezod's existing call execution. This hybrid approach minimizes duplication while preserving Cosmos-specific behavior.

The key question is: **does geth 1.16.x's `StateTransition` handle authorization processing separately from call execution?** Based on the research doc, it does — `validateAuthorization()` and `applyAuthorization()` are separate methods called before `evm.Call()`. Mezod could call these directly if they're exported, or reimplement them.

**Critical detail**: Authorization processing modifies state (sets delegation code, increments nonce) and must happen within the same `StateDB` context as the rest of the transaction. Since mezod creates its own `StateDB` instance (`statedb.New(ctx, k, txConfig)`), the authorization processing must use that same instance.

---

### 2.5 StateDB — Delegation Code Storage

**Files:**
- `x/evm/statedb/statedb.go`
- `x/evm/statedb/state_object.go`
- `x/evm/statedb/interfaces.go`
- `x/evm/keeper/statedb.go`

**What to do:**

The existing `StateDB.SetCode()` and `GetCode()` methods should work for storing delegation designators (they're just 23-byte code values with the `0xef0100` prefix). However:

1. **`IsContract()` check** — The `Account.IsContract()` method (likely checking `CodeHash != emptyCodeHash`) will return `true` for accounts with delegation designators. This is correct for geth's behavior but may affect mezod's ante handlers (see section 2.6).

2. **No new StateDB methods needed** — geth 1.16.x's `resolveCode()` and `resolveCodeHash()` live on the EVM struct, not on StateDB. They read code via `StateDB.GetCode()` and parse the delegation prefix. Since mezod uses geth's EVM directly, this resolution happens automatically.

3. **Storage persistence caveat** — When delegation is cleared (target address = 0x0), the code is set to nil but storage is NOT wiped. This is by spec but worth documenting.

---

### 2.6 Ante Handlers — Account Verification

**Files:**
- `app/ante/evm/eth.go` (lines 60-105 — `EthAccountVerificationDecorator`)

**What to do:**

The current ante handler rejects transactions from contract accounts:
```go
} else if acct.IsContract() {
    return ctx, errorsmod.Wrapf(errortypes.ErrInvalidType,
        "the sender is not EOA: address %s, codeHash <%s>", fromAddr, acct.CodeHash)
}
```

With EIP-7702, an account with a delegation designator has **non-empty code** (23 bytes) but should still be allowed to send transactions. This is the equivalent of geth's EIP-3607 exemption (`state_transition.go:334-338`).

**Fix:**
```go
} else if acct.IsContract() {
    // EIP-7702: Allow accounts with delegation designators to send transactions.
    // Only reject accounts with actual contract code (not delegation code).
    code := avd.evmKeeper.GetCode(ctx, common.BytesToHash(acct.CodeHash))
    if _, isDelegated := ethtypes.ParseDelegation(code); !isDelegated {
        return ctx, errorsmod.Wrapf(errortypes.ErrInvalidType,
            "the sender is not EOA: address %s, codeHash <%s>", fromAddr, acct.CodeHash)
    }
}
```

**Alternative**: Add an `IsDelegated()` method to the account type that checks the code prefix. **Trade-off**: Cleaner but requires reading code from store in the account layer.

---

### 2.7 Ante Handlers — Signature Verification

**Files:**
- `app/ante/evm/sigverify.go`

**What to do:**

The signature verification decorator uses `ethtypes.MakeSigner()` to create a signer appropriate for the current fork. With geth 1.16.x, `MakeSigner()` will return a `PragueSigner` when Prague is active, which supports type 0x04 transactions. This should work automatically.

**Verify**: That the signer correctly handles SetCode transaction signatures. The `PragueSigner` extends `CancunSigner` and adds support for recovering the sender from type 0x04 transactions.

---

### 2.8 RPC Layer

**Files:**
- `rpc/backend/sign_tx.go` — Transaction signing/sending
- `rpc/backend/call_tx.go` — `eth_sendRawTransaction`
- `rpc/types/utils.go` — Transaction formatting for JSON-RPC
- `internal/ethapi/transaction_args.go` (in geth fork) — `TransactionArgs`

**What to do:**

**a) `eth_sendRawTransaction`** — Should work automatically. Raw transactions are RLP-decoded by geth's `Transaction.UnmarshalBinary()`, which in 1.16.x handles type 0x04. The decoded transaction flows through `FromEthereumTx()` → `NewTxDataFromTx()`, which needs the new `SetCodeTxType` case (section 2.2c).

**b) `eth_sendTransaction`** — Needs `TransactionArgs` to support an `AuthorizationList` field. This is handled in geth's `transaction_args.go` but mezod may have its own copy.

**c) Transaction formatting** — `NewTransactionFromMsg()` and related functions need to handle the new transaction type for RPC responses. The authorization list should be included in the JSON response.

**d) `eth_signTransaction`** — Must support signing type 0x04 transactions if the node holds the signing key.

---

### 2.9 Transaction Pool (Mempool) Hardening

**Files:**
- Cosmos SDK mempool (CometBFT)
- `app/ante/` handlers

**What to do:**

Geth implements specific txpool protections for EIP-7702:
1. **Single-slot restriction**: Delegated accounts get only 1 pending tx slot
2. **Authority reservation**: Each authority in a SetCode tx can only appear in one in-flight tx

Mezod uses CometBFT's mempool, not geth's txpool. The ante handler chain provides the validation layer. Consider whether these protections are needed:

- **Single-slot restriction**: Could be implemented in a new ante decorator that checks if the sender has a delegation designator and limits their pending transactions. **Trade-off**: CometBFT's mempool is simpler than geth's; the risk of mempool-level attacks is lower because Cosmos validators control block production.
- **Authority reservation**: More complex. Would need to track which authorities are referenced by pending SetCode transactions. **Trade-off**: May not be necessary for the initial implementation given Cosmos's block production model.

**Recommendation**: Skip txpool hardening for the initial implementation. Monitor for issues and add protections if needed. Document the gap.

---

### 2.10 Upgrade Handler

**Files:**
- `app/upgrades/` directory
- `app/app.go`

**What to do:**

Create a new upgrade handler that:
1. Sets `PragueTime` in the EVM module's chain config params to the desired activation time
2. Runs any necessary state migrations

This follows the pattern of previous fork activations (Shanghai, Cancun) in mezod.

---

## 3. File-by-File Change Summary

| File | Change Type | Description | Effort |
|------|------------|-------------|--------|
| `proto/ethermint/evm/v1/evm.proto` | Add field | `prague_time` field (24) in ChainConfig | Small |
| `proto/ethermint/evm/v1/tx.proto` | Add messages | `SetCodeTx`, `SetCodeAuthorization` | Medium |
| `x/evm/types/chain_config.go` | Modify | Map PragueTime, update defaults/validation | Small |
| `x/evm/types/set_code_tx.go` | **New file** | TxData implementation for SetCodeTx | Large |
| `x/evm/types/tx_data.go` | Modify | Add SetCodeTxType case in NewTxDataFromTx | Small |
| `x/evm/types/msg.go` | Modify | Handle type 0x04 in validation/encoding | Small |
| `x/evm/keeper/gas.go` | Modify | Pass auth list to IntrinsicGas | Small |
| `x/evm/keeper/state_transition.go` | Modify | Authorization processing before EVM call | Large |
| `x/evm/statedb/statedb.go` | Modify (maybe) | Prepare() may need Prague-aware access list handling | Small |
| `app/ante/evm/eth.go` | Modify | EIP-3607 exemption for delegated senders | Small |
| `app/ante/evm/sigverify.go` | Verify | Signer selection handles Prague | Small |
| `rpc/types/utils.go` | Modify | Transaction formatting for type 0x04 | Medium |
| `rpc/backend/sign_tx.go` | Modify | Handle AuthorizationList in TransactionArgs | Medium |
| `app/upgrades/` | **New file** | Prague activation upgrade handler | Small |
| Tests (various) | **New files** | System tests, unit tests | Large |

---

## 4. Implementation Order (Dependency Graph)

```
Phase 1: Foundation (no functional change)
├─ 1a. Upgrade geth fork to 1.16.x (prerequisite, separate effort)
├─ 1b. Add PragueTime to proto + chain_config.go
└─ 1c. Regenerate protobuf types

Phase 2: Transaction Type Support
├─ 2a. Define SetCodeTx protobuf messages
├─ 2b. Implement set_code_tx.go (TxData interface)
├─ 2c. Register in tx_data.go switch
└─ 2d. Update msg.go validation/encoding

Phase 3: Execution Layer
├─ 3a. Update GetEthIntrinsicGas (auth list parameter)
├─ 3b. Add authorization processing in state_transition.go
├─ 3c. Update ante handler (EIP-3607 exemption)
└─ 3d. Verify signer selection in sigverify.go

Phase 4: RPC & External Interface
├─ 4a. Update transaction formatting (rpc/types)
├─ 4b. Handle AuthorizationList in TransactionArgs
└─ 4c. Verify eth_sendRawTransaction flow

Phase 5: Activation & Testing
├─ 5a. Create upgrade handler for PragueTime activation
├─ 5b. System tests (EIP-7702 delegation, clearing, gas)
├─ 5c. Integration tests with real EVM execution
└─ 5d. Testnet deployment
```

---

## 5. Risks and Open Questions

### High Risk

| Risk | Impact | Mitigation |
|------|--------|------------|
| **State transition divergence** | Authorization processing in mezod could differ from geth's behavior, causing consensus failures | Maximize code reuse from geth; extensive differential testing |
| **Ante handler bypass** | Delegated accounts might bypass existing security checks beyond the IsContract check | Audit all ante handlers for assumptions about EOA vs. contract accounts |
| **Precompile interaction** | Custom precompiles may behave unexpectedly when called by delegated EOAs | Test all precompiles with delegated callers |

### Medium Risk

| Risk | Impact | Mitigation |
|------|--------|------------|
| **Storage persistence** | Orphaned storage from cleared delegations could be exploited | Document behavior; consider storage-clearing mechanism |
| **Cosmos-EVM nonce mismatch** | Authorization processing increments nonce at the EVM level; Cosmos ante handler also manages nonces | Verify nonce handling doesn't double-increment |
| **Gas metering misalignment** | Cosmos SDK gas meter and EVM gas meter may disagree on authorization costs | Ensure infinite gas meter covers authorization phase |

### Open Questions

1. **Should authorization processing happen in the ante handler or in ApplyMessageWithConfig?**
   - Ante handler: Earlier validation, can reject before gas consumption
   - ApplyMessageWithConfig: More consistent with geth's architecture
   - **Recommendation**: ApplyMessageWithConfig, matching geth's behavior where invalid auths are silently skipped

2. **Does `core.TransactionToMessage()` in geth 1.16.x populate `msg.AuthList`?**
   - If yes, mezod's existing flow of `core.TransactionToMessage(tx, signer, baseFee)` → `ApplyMessageWithConfig(msg, ...)` will automatically carry the auth list
   - Need to verify after geth upgrade

3. **How should the delegation designator interact with Cosmos account types?**
   - The `EthAccount` type has a `CodeHash` field. An account with a delegation designator will have a non-empty `CodeHash` but should still be recognized as an EOA-like entity
   - May need a new account type or a helper method

4. **Is CometBFT's mempool resilient enough without geth-style txpool protections?**
   - Geth's mempool protections (single-slot, authority reservation) prevent specific DoS vectors
   - CometBFT validators can reject transactions; the risk profile is different
   - Monitor post-launch; add protections if needed

---

## 6. Testing Strategy

### Unit Tests
- `set_code_tx.go`: Validate(), AsEthereumData(), fee calculations
- `chain_config.go`: PragueTime validation, fork ordering
- `gas.go`: Intrinsic gas with auth tuples (new account vs. existing)
- `eth.go` ante handler: Delegated sender acceptance

### Integration Tests
- Full transaction lifecycle: submit SetCode tx → authorize → delegated execution
- Authorization clearing: set delegation → execute → clear → verify no delegation
- Invalid authorization handling: wrong nonce, wrong chain ID, invalid signature → silently skipped
- Gas consumption: verify correct intrinsic gas charges for auth tuples

### System Tests (TypeScript, following existing patterns)
- `Eip7702Delegation.test.ts`: Create delegation, verify CALL resolution, clear delegation
- `Eip7702Gas.test.ts`: Verify gas charges for various authorization scenarios
- `Eip7702Security.test.ts`: Replay protection, cross-chain authorization, nonce handling
- `Eip7702Precompiles.test.ts`: Interaction with custom precompiles via delegation

---

## 7. Comparison with Alternative Approaches

### Approach A: Full geth state transition delegation (not recommended for now)
Replace mezod's `ApplyMessageWithConfig` with a direct call to geth's `core.ApplyMessage()`.
- **Pros**: Perfect spec compliance, zero maintenance burden for future EIPs
- **Cons**: Loses Cosmos-specific gas metering, custom precompile injection, min gas multiplier
- **Verdict**: Too disruptive for a single EIP; consider as a long-term refactor

### Approach B: Backport EIP-7702 into geth v1.14.8-mezo2 (not recommended)
Cherry-pick EIP-7702 changes from geth 1.16.x into the current fork.
- **Pros**: No geth upgrade dependency
- **Cons**: ~15 files to backport, merge conflicts, ongoing maintenance divergence
- **Verdict**: The geth upgrade to 1.16.x is already planned; backporting creates throwaway work

### Approach C: Wait for geth upgrade then add mezod layer (recommended)
Upgrade geth fork to 1.16.x first, then add the mezod-specific changes documented above.
- **Pros**: Minimal geth-layer work, clean dependency chain, most maintainable
- **Cons**: Blocked on geth upgrade timeline
- **Verdict**: Best balance of effort and correctness
