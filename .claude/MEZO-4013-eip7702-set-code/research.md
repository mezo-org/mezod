> **Disclaimer:** This is a temporary file used during implementation of
> MEZO-4013 (EIP-7702 Set Code Transactions). It should be removed once the feature is
> complete.

# EIP-7702 ("Set EOA Account Code") Implementation in go-ethereum

## 1. Architecture & Implementation Overview

EIP-7702 introduces **transaction type 0x04** that allows Externally Owned Accounts (EOAs) to delegate their code execution to a contract address. This effectively gives EOAs smart contract capabilities while retaining their private key control. It was activated as part of the **Prague hardfork**.

### Core Data Structures

**Authorization Tuple** — `core/types/tx_setcode.go:71-79`
```go
type SetCodeAuthorization struct {
    ChainID uint256.Int     // 0 = wildcard (any chain), or specific chain ID
    Address common.Address  // Target contract to delegate to (zero = clear)
    Nonce   uint64          // Must match authority's current nonce
    V       uint8           // Signature y-parity (0 or 1)
    R       uint256.Int
    S       uint256.Int
}
```

**Delegation Designator** — `core/types/tx_setcode.go:32-47`
- Format: `0xef0100` (3-byte prefix) + 20-byte target address = **23 bytes total**
- Stored directly in the account's code field
- `ParseDelegation(code)` — validates and extracts the target address
- `AddressToDelegation(addr)` — creates the 23-byte designator
- Setting `Address` to zero clears the delegation

**SetCodeTx (type 0x04)** — `core/types/tx_setcode.go:49-67`
- Extends EIP-1559 fields with an `AuthList []SetCodeAuthorization`
- Must have a non-nil `To` field (cannot create contracts)
- RLP-encoded with type prefix byte `0x04`

### Control Flow: Transaction Lifecycle

```
1. TX SUBMISSION (txpool)
   └─ validation.go:85    — Prague fork gate
   └─ validation.go:149   — AuthList must be non-empty
   └─ legacypool.go:590   — Delegation limit checks (1-slot restriction)
   └─ legacypool.go:614   — Authority reservation (1 in-flight tx per authority)

2. INTRINSIC GAS (state_transition.go:113-114)
   └─ Base cost: len(authList) × CallNewAccountGas (25,000 per auth)
   └─ Refund of 12,500 if authority account already exists

3. PRE-CHECKS (state_transition.go:310-339)
   └─ EIP-3607 exemption: sender MAY have delegation code (but not contract code)

4. AUTHORIZATION PROCESSING (state_transition.go:505-520)
   └─ For each auth tuple:
      ├─ validateAuthorization() — chain ID, nonce, signature, code checks
      ├─ applyAuthorization()   — increment nonce, set delegation code
      └─ Errors silently skipped (invalid auths don't abort the tx)
   └─ Post-processing: warm the delegation target in access list

5. CALL EXECUTION (vm/evm.go:289)
   └─ resolveCode(addr) follows one level of delegation
   └─ Delegated code executes in the context of the delegating EOA
```

### Key Integration Points

| Component | File | What It Does |
|-----------|------|-------------|
| **Transaction type** | `core/types/tx_setcode.go` | Defines `SetCodeTx`, authorization struct, encoding |
| **State transition** | `core/state_transition.go:576-632` | Validates & applies authorizations |
| **EVM code resolution** | `core/vm/evm.go:626-653` | `resolveCode`/`resolveCodeHash` follow delegation |
| **Gas (calls)** | `core/vm/operations_acl.go:253-311` | EIP-7702 cold/warm costs for delegation resolution |
| **Gas (intrinsic)** | `core/state_transition.go:113-114` | Per-auth-tuple gas charge |
| **Jump table** | `core/vm/eips.go:553-558` | `enable7702` modifies CALL/CALLCODE/STATICCALL/DELEGATECALL gas |
| **Txpool** | `core/txpool/legacypool/legacypool.go:586-648` | Delegation limit + authority reservation |
| **Pool validation** | `core/txpool/validation.go:85,149-153` | Fork gating + non-empty auth list |
| **RPC** | `internal/ethapi/transaction_args.go:518-538` | `AuthorizationList` field in tx args |
| **Errors** | `core/error.go:130-145` | Seven EIP-7702-specific error types |
| **Fork config** | `params/config.go:63` | `PragueTime` (mainnet: `1746612311`) |

### How Delegation Resolution Works at the EVM Level

`core/vm/evm.go:628-637`:
```go
func (evm *EVM) resolveCode(addr common.Address) []byte {
    code := evm.StateDB.GetCode(addr)
    if !evm.chainRules.IsPrague { return code }
    if target, ok := types.ParseDelegation(code); ok {
        return evm.StateDB.GetCode(target) // only ONE level
    }
    return code
}
```

- Used by: `Call`, `CallCode`, `DelegateCall`, `StaticCall`
- **Not** used by: `EXTCODESIZE`, `EXTCODECOPY`, `EXTCODEHASH` — these return the raw delegation designator (23 bytes)
- Single-hop only: prevents infinite delegation chains

---

## 2. Spec Compliance Assessment

### Fully Implemented

| Spec Requirement | Status | Evidence |
|-----------------|--------|---------|
| Transaction type 0x04 with auth list | **Complete** | `tx_setcode.go`, `transaction.go:52` |
| Authorization structure (chainId, address, nonce, sig) | **Complete** | `tx_setcode.go:71-79` |
| Chain ID validation (0 = wildcard, or current) | **Complete** | `state_transition.go:579` |
| Nonce matching and increment | **Complete** | `state_transition.go:601-603, 621` |
| Signature recovery via ECDSA | **Complete** | `tx_setcode.go:118-139` |
| Authority must be EOA or already-delegated | **Complete** | `state_transition.go:598` |
| Delegation designator format (0xef0100 + addr) | **Complete** | `tx_setcode.go:32-47` |
| Clear delegation via zero address | **Complete** | `state_transition.go:622-625` |
| CALL/STATICCALL/DELEGATECALL/CALLCODE resolve delegation | **Complete** | `evm.go:289, 355, 399, 450` |
| EIP-3607 exemption for delegated senders | **Complete** | `state_transition.go:334-338` |
| Per-authorization intrinsic gas (25,000) | **Complete** | `state_transition.go:114` |
| Refund for existing accounts | **Complete** | `state_transition.go:616-617` |
| Cold/warm access costs for delegation resolution | **Complete** | `operations_acl.go:274-287` |
| Auth list must be non-empty | **Complete** | `validation.go:150-151` |
| Must have non-nil `To` | **Complete** | `state_transition.go:499-500` |
| Invalid auths silently skipped (not tx failure) | **Complete** | `state_transition.go:506-510` |
| Single-level delegation resolution | **Complete** | `evm.go:634` ("Note we only follow one level") |
| Convenience warming of delegation target | **Complete** | `state_transition.go:513-520` |
| Nonce overflow check (EIP-2681) | **Complete** | `state_transition.go:583-584` |

### EXTCODE* Opcode Behavior

`EXTCODESIZE`, `EXTCODECOPY`, and `EXTCODEHASH` do **not** resolve delegation — they return the raw 23-byte designator. This is consistent with the implementation design:

- `enable7702()` (`eips.go:553-558`) only modifies gas functions for CALL-family opcodes
- `resolveCodeHash` comment explicitly says: *"Although this is not accessible in the EVM it is used internally to associate jumpdest analysis to code"* (`evm.go:642-643`)

This behavior allows contracts to **detect** delegations by inspecting the raw code (checking for the `0xef0100` prefix). Given that Prague is live on mainnet and go-ethereum is the reference client, this is the canonical behavior.

### Gas Model

The intrinsic gas charges `CallNewAccountGas` (25,000) per authorization tuple (`state_transition.go:114`), with a refund of `CallNewAccountGas - TxAuthTupleGas` (25,000 - 12,500 = 12,500) for already-existing accounts (`state_transition.go:617`). This means:
- **New account**: 25,000 gas (net)
- **Existing account**: 12,500 gas (net, after refund)

The `TxAuthTupleGas` constant (12,500) is defined at `params/protocol_params.go:101`.

---

## 3. Security Analysis

### Replay Protection — Secure

| Vector | Protection | Location |
|--------|-----------|----------|
| Cross-chain replay | Chain ID must be 0 (wildcard) or match current | `state_transition.go:579` |
| Same-chain replay | Nonce must exactly match authority's current nonce; incremented after use | `state_transition.go:601, 621` |
| Nonce overflow | Checked: `auth.Nonce+1 < auth.Nonce` rejects at 2^64-1 | `state_transition.go:583` |
| Wildcard chain ID (0) | Intentional by spec — enables cross-chain authorization. Users must understand the trust model | By design |

### Signature Validation — Secure

- `ValidateSignatureValues` with `homestead=true` enforces `s <= secp256k1halfN` (EIP-2), preventing signature malleability (`crypto/crypto.go:246`)
- V must be 0 or 1 (not legacy 27/28) (`tx_setcode.go:120`)
- Full ECDSA recovery via `crypto.SigToPub` (`tx_setcode.go:129`)
- Authorization signature hash uses unique prefix `0x05` to prevent cross-domain replay (`tx_setcode.go:105`)

### Delegation Chain / Reentrancy — Secure

- **Hard single-hop limit**: `resolveCode` and `resolveCodeHash` follow exactly one level (`evm.go:634, 648`)
- No possibility of infinite loops or circular chains
- Self-delegation is allowed but harmless (the 23-byte designator points back to the same account)

### Transaction Pool DoS — Secure with documented caveats

The txpool implements several restrictions (`legacypool.go:586-648`):

1. **Single-slot restriction**: Delegated accounts (or those with pending delegation) get only 1 pending tx slot instead of the normal limit (`legacypool.go:590-610`)
2. **No gapped nonces**: Transactions from delegated accounts must be continuous (`legacypool.go:600-601`)
3. **Authority reservation**: Each authority in a SetCode tx can only appear in one in-flight tx (`legacypool.go:622-646`)
4. **Cross-pool race condition**: Acknowledged in comments (`legacypool.go:634-641`) — a SetCode tx may be accepted in one pool while conflicting txs are accepted in others. Deemed acceptable as it primarily limits deliberate attack stacking.

### State Integrity — Secure, with one caveat

| Concern | Status | Detail |
|---------|--------|--------|
| Delegation designator format | **Strict** | Must be exactly 23 bytes with 0xef0100 prefix (`tx_setcode.go:37-42`) |
| Contract creation blocked | **Enforced** | SetCode txs must have `To != nil` (`state_transition.go:499-500`) |
| SELFDESTRUCT interaction | **Safe** | EIP-6780 restricts SELFDESTRUCT to same-tx only; cannot corrupt delegating accounts |
| **Storage persistence after clear** | **Caveat** | Clearing delegation (`Address = 0x0`) sets code to nil but does **not** wipe storage. Storage written by delegated code persists on the EOA even after delegation is removed. Future delegations or direct access could see stale storage. |

### Privilege Escalation — Secure by design

- **Revocation**: Only the EOA holder can sign new authorizations (requires their private key + current nonce)
- **No re-delegation by contract**: Delegated code cannot issue new authorizations — only the original EOA's signature can create them
- **Permanent takeover impossible**: User can always revoke by signing an authorization with `Address = 0x0`
- **Trust model**: Users must trust the delegation target contract. A malicious contract can spend the EOA's ETH or make arbitrary calls from its address, but cannot prevent revocation.

### Gas-Related Attacks — Mitigated

- **No explicit auth list size cap**, but effectively bounded by:
  - Block gas limit (~30M): at 25,000 gas per auth, max ~1,200 auths per block
  - Transaction size limit: 128KB (`legacypool.go:56`)
- Refunds capped by EIP-3529 (max refund = gasUsed/5)
- Delegation resolution gas properly charged: cold access (2,600) or warm access (100) for the target (`operations_acl.go:274-287`)

### EIP-3607 Interaction — Secure

`state_transition.go:334-338`:
```go
code := st.state.GetCode(msg.From)
_, delegated := types.ParseDelegation(code)
if len(code) > 0 && !delegated {
    return ErrSenderNoEOA  // only reject if it's REAL contract code
}
```
Accounts with delegation designators (23 bytes, 0xef0100 prefix) are allowed as senders. Only accounts with actual deployed contract code are rejected.

### ORIGIN/CALLER Semantics — Correct

- `tx.origin` remains the transaction sender throughout all call depths
- When delegated code executes, `msg.sender` / `CALLER` is the immediate caller (the account that called the EOA)
- Value transfers to delegated EOAs trigger delegated code execution with correct `msg.value`

---

## Summary

### Implementation Quality: Comprehensive and Production-Ready

The EIP-7702 implementation in go-ethereum is thorough, spanning ~15 files across the transaction type system, state transition engine, EVM, txpool, and RPC layer. It is already **live on mainnet** (Prague activated at timestamp 1746612311).

### Key Design Decisions

1. **Single-hop delegation**: Prevents complexity/DoS from chained delegations
2. **Silent auth skip**: Invalid authorizations don't fail the transaction — allows partial success
3. **EXTCODE* returns raw designator**: Allows on-chain detection of delegations
4. **Txpool hardening**: 1-slot limit + authority reservation prevents mempool-level attacks

### Risks to Be Aware Of

| Risk | Severity | Notes |
|------|----------|-------|
| Storage persistence after delegation clear | Medium | Storage is NOT wiped when delegation is removed; orphaned state can be read by future code |
| Chain ID = 0 wildcard | Low (by design) | Authorizations valid on all chains; users must understand cross-chain implications |
| Cross-pool race condition in txpool | Low | Documented and accepted trade-off vs. performance |
| No explicit auth list size limit | Low | Implicit gas/size limits provide sufficient protection |
