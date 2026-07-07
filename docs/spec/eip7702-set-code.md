# EIP-7702 Set Code Transactions

## Overview

EIP-7702 introduces transaction type `0x04` ("set code"), which lets an
externally owned account (EOA) attach a per-account delegation designator
to one or more contract addresses. Calls to a delegated EOA execute the
target contract's code in the EOA's storage and balance context, with the
EOA's address in `CALLER`. The delegation is a 23-byte code value of the
form `0xef0100 || target_address`; the EOA can rotate or clear it by
signing a fresh authorization.

The EIP shipped with the Ethereum Prague hardfork.

Mezo ships EIP-7702 so that the broader EVM tooling ecosystem (account
abstraction-aware wallets, viem/ethers v6 sponsored-tx flows, MetaMask
batched-action UX) — which now assumes the feature is available on any
post-Prague chain — works against mezod without external relayers.

## Implementation summary

- **Fork activation.** A new `prague_time` field on the EVM module's
  `ChainConfig` (`proto/ethermint/evm/v1/evm.proto`) plumbs through to
  `params.ChainConfig.PragueTime`. The genesis default activates Prague
  immediately (`prague_time = 0`), matching how Shanghai and Cancun are
  wired today; the live mainnet and testnet networks pick up activation
  through a dedicated upgrade handler that writes a chosen unix
  timestamp into the stored chain config.
- **Transaction type.** `SetCodeTx` (type `0x04`) is added as a first-class
  Cosmos-encoded `TxData` alongside `LegacyTx`, `AccessListTx`, and
  `DynamicFeeTx`. A new `SetCodeAuthorization` message carries the
  per-tuple `(chain_id, address, nonce, v, r, s)` payload. The proto
  field is named `auth_list`; it surfaces over JSON-RPC as
  `authorizationList`. Strict validation rejects type-`0x04`
  transactions whose `to` is nil or whose `auth_list` is empty before
  they reach the keeper.
- **Authorization processing.** The keeper's state-transition driver
  reimplements geth 1.16.x's `validateAuthorization` and
  `applyAuthorization` semantics on top of mezod's `statedb.StateDB`,
  pinned by version comment to a specific upstream commit. Authorization
  tuples are processed after intrinsic-gas accounting and before EVM
  call execution; invalid tuples are silently skipped and do not abort
  the transaction. Successful authorizations bump the authority's nonce,
  install (or clear, if `target == 0x0`) the 23-byte delegation
  designator on the authority's code field, and warm the delegation
  target in the access list. Self-sponsored authorizations (where the
  authority equals the transaction sender) sign their tuple against the
  post-bump sender nonce; on consensus paths the ante's
  `EthIncrementSenderSequenceDecorator` supplies the bump, and on
  keeper-internal simulate / trace paths (`EthCall`, `TraceTx`,
  `traceTx`, `SimulateV1`) each entry point mirrors the bump inline so
  authorization validation reads the same post-bump value regardless
  of path.
- **Post-loop call-target warming.** After the authorization loop, if
  `msg.To`'s code resolves to a delegation designator, the resolved
  target is also added to the access list — mirroring upstream
  `core.ApplyMessage` post-loop warming. This complements the
  per-tuple warming of the delegation target inside the loop and
  covers the case where the tx itself just installed a delegation on
  `msg.To`, so the subsequent `evm.Call` does not pay cold-access gas
  for the freshly delegated address.
- **EVM call resolution.** The geth EVM (`v1.16.9-mezo0`) handles
  delegation resolution natively: `Call`, `CallCode`, `DelegateCall`, and
  `StaticCall` follow exactly one level of delegation when reading code.
  `EXTCODESIZE`, `EXTCODECOPY`, and `EXTCODEHASH` deliberately return the
  raw 23-byte designator so on-chain contracts can detect delegations.
  Mezo inherits this behavior unchanged.
- **Ante handler exemption.** `EthAccountVerificationDecorator` rejects
  contract-coded senders today (the long-standing EIP-3607 enforcement).
  The decorator gains a delegation-aware exemption: an account whose code
  parses as a valid EIP-7702 delegation designator is treated as an EOA
  for sender purposes. Accounts with non-delegation contract code are
  still rejected.
- **JSON-RPC.** `eth_sendRawTransaction` flows through unchanged once
  `NewTxDataFromTx` learns to map `SetCodeTxType`. `eth_sendTransaction`
  and `eth_signTransaction` accept an `authorizationList` field on
  `TransactionArgs`. The transaction-formatting layer
  (`rpc/types/utils.go`) emits `authorizationList` and the type-`0x04`
  envelope for both subscriptions and historical lookups.
- **Indexing.** Mezod runs two transaction indexers and `SetCodeTx`
  must surface through both on equal footing with the existing types.
  The always-on CometBFT tx indexer indexes by tx events emitted by
  the EVM module and is the default backing for
  `eth_getTransactionByHash`, `eth_getTransactionReceipt`, and the
  `tx_search` flows when the custom indexer is disabled. The opt-in
  custom KV indexer (`indexer/kv_indexer.go`, enabled per node via
  `enable-indexer = true` in `app.toml`) is the same surface plus the
  pseudo-transaction observability path (so block explorers can see
  bridge activity at index 0) and the `mezo_*` RPC additions. Both
  indexers are tx-type-agnostic at the indexing layer: they ride on
  the `MsgEthereumTx` event-emission path, which set-code transactions
  use unchanged.

## Conformance with the EIP

The authoritative spec is [EIP-7702][eip-7702] in the
`ethereum/EIPs` repository, anchored to its post-Prague freeze; the
reference implementation lives in `core/types/tx_setcode.go` and
`core/state_transition.go` of go-ethereum.

[eip-7702]: https://eips.ethereum.org/EIPS/eip-7702

Mezo's conformance posture mirrors the project's stance for prior
EVM forks: the in-VM behavior (delegation resolution, gas costs,
intrinsic-gas charges, opcode semantics, signer recovery) is delegated
to geth verbatim by way of the upgraded `mezo-org/go-ethereum`
fork. The application layer added in this scope (proto types, ante
handlers, RPC plumbing, upgrade handler) is exercised by:

- A dedicated system test suite (`Eip7702*.test.ts`) that drives every
  documented spec path (delegation install, rotation, clearing, replay
  rejection, gas accounting, EXTCODE\* opacity, EIP-3607 exemption,
  precompile interaction) end-to-end via `eth_sendRawTransaction` against
  localnode.
- Keeper-level unit and table tests covering authorization
  validation/application, intrinsic-gas plumbing, the chain-config
  `PragueTime` migration, and ante-handler exemption decisions.
- A fuzz target on the `SetCodeTx` proto unmarshaler that asserts no
  panics and that every error carries a typed mezod error code.

The system test asserts per-field invariants (status, gasUsed, codeHash,
nonce, log topics) rather than byte-level response equality: mezod's
chain id, base block, fee market, and account state never match the
upstream replay corpus, so byte-for-byte hashes are not a meaningful
check. Fuzz seeds and any spec-conformance fixtures the upstream
project ships are wired in as the seed corpus.

## Mezo-specific divergences

Each divergence has a tripwire in `Eip7702_MezoDivergence.test.ts` so any
accidental flip surfaces loudly in CI.

1. **No mempool-level authority reservation or single-slot delegation
   limit.** Geth's txpool rejects a second pending transaction from a
   delegated account, and reserves each authority across all in-flight
   set-code transactions. Mezo runs on CometBFT, where validators control
   block proposal directly and there is no peer-to-peer transaction pool
   in the geth sense. The hardenings are not ported. The risk model is
   covered by the existing per-block gas envelope plus the per-tx
   intrinsic-gas charge of `CallNewAccountGas` per authorization.
2. **Authorization processing reimplemented in mezod's keeper.** Geth's
   `validateAuthorization` and `applyAuthorization` are unexported
   methods on an unexported struct. Mezo reimplements them on the
   keeper's `statedb.StateDB` so that authorization side-effects (nonce
   bump, code write, target warming) commit through the same store
   path as the rest of the transaction. The implementation is pinned by
   comment to a specific upstream geth commit and is exercised by both
   keeper-level differential tests and the system suite.
3. **No new tx-type signer adapter beyond what `MakeSigner` returns.**
   Once Prague is active, `ethtypes.MakeSigner` returns geth's
   `PragueSigner`, which already handles type-`0x04` recovery. Mezo's
   ante and RPC paths use `MakeSigner` (with `LatestSignerForChainID`
   in caching paths), so no Mezo-specific signer is introduced.
4. **Authorization targets that are precompiles are rejected.** Geth's
   `validateAuthorization` accepts any 20-byte address as
   `auth.Address`, including the stock precompile space (0x01..0x12).
   On stock geth those addresses hold no stored bytecode, so a
   delegation pointing there is a no-op at runtime. Mezo additionally
   stores facade bytecode at the custom precompile addresses
   (`0x7b7c…` range, registered via `CustomPrecompileGenesisAccounts`
   in `x/evm/keeper/keeper.go`), so a delegation pointing at a custom
   precompile would actually execute in the authority's context —
   surprising semantics and a fragile surface to leave open. To keep
   the rule uniform across both precompile families, mezod rejects any
   authorization whose target is in `evm.Precompiles()` (the union of
   fork-active stock precompiles and Mezo custom precompiles for the
   current EVM). Rejection follows the EIP's per-tuple rule: the
   offending tuple is silently skipped, the rest of the transaction
   proceeds. The clear-delegation path (`auth.Address == 0x0`) is
   unaffected — the zero address is not a precompile in any fork or in
   Mezo's custom set.

## Key decisions

- **Genesis-default activation, upgrade-driven for live chains.** The
  default chain config sets `prague_time = 0` so new chains and devnets
  inherit Prague immediately on next restart. Activation on the live
  mainnet and testnet networks happens through a dedicated upgrade
  handler that writes a chosen unix timestamp into the stored EVM
  module params, so every node enables 7702 at the same height.
- **Reimplement, don't refactor.** mezod keeps its custom
  `applyMessageWithConfig` (which preserves `MinGasMultiplier`-aware
  gas accounting, custom precompile registries, and the simulate
  driver's call hooks) and reimplements just the authorization-validation
  and authorization-application steps from geth. Folding mezod onto
  `core.ApplyMessage` is a much larger refactor and is explicitly out of
  scope.
- **One typed `*EvmError` family for set-code paths.** Authorization
  validation failures, missing `to`, empty `auth_list`, malformed RLP,
  and PragueTime-not-active rejections each get a distinct code in the
  existing EVM error catalog. Per-call validation errors stay per-tuple
  (silently skipped at execution time, per spec); transaction-level
  validation failures become top-level fatals at the same boundary as
  for prior types.
- **No txpool hardening.** The geth-style single-slot and authority
  reservation protections are not ported. Documented as a divergence.
- **Precompile-target authorizations rejected.** Mezo special-cases the
  precompile address space at the authorization-validation step:
  authorizations whose target is any precompile (stock or custom) are
  rejected at validate time and silently skipped per the EIP's
  per-tuple rule. The custom mezo precompiles enumerated in
  `x/evm/types/precompile.go` (`DefaultPrecompilesVersions`) remain at
  their canonical addresses and are still callable from a delegated EOA
  through normal contract calls — only authorization targets are
  restricted. See divergence #4 for the rationale.

## Configuration

- **`prague_time` (proto field 24, EVM `ChainConfig`).** Mirrors the
  shape of `shanghai_time` and `cancun_time`. Maps directly onto
  `params.ChainConfig.PragueTime` in the geth runtime config. `nil`
  means the fork is not yet scheduled; `0` means already active. The
  genesis default is `0`. The field is read by the keeper on every
  message-application path (`applyMessageWithConfig`,
  `GetEthIntrinsicGas`, `feeChecker`), so flipping it through the
  upgrade handler is sufficient — no per-call wiring elsewhere.
- **No new `json-rpc.*` knob.** The feature is governed by chain config
  alone; operators do not get a per-node toggle. Once Prague is active,
  every node accepts type-`0x04` transactions. This matches how
  Shanghai/Cancun-era surfaces were rolled out and avoids creating a
  validator-vs-RPC consensus split.
- **Existing `json-rpc.gas-cap` and `json-rpc.evm-timeout` apply
  unchanged.** Per-tx gas budget for `eth_call` / `eth_estimateGas` /
  `eth_simulateV1` is enforced as today; the per-authorization
  intrinsic charge is included in the same accounting.

## Usage examples

Install a delegation. The EOA at `0xc100…01` self-sponsors a type-`0x04`
transaction that delegates its own code to the contract at `0xc200…01`.
The `to` field is the EOA itself; `authorizationList` carries a single
tuple naming the delegation target, and the tuple's nonce is the EOA's
post-bump sender nonce because the authority equals the sender. The
`r` / `s` fields are placeholders here — the caller produces them
off-chain by signing the `(chainId, address, nonce)` payload per
EIP-7702 §1 before submitting the request.

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "eth_sendTransaction",
  "params": [
    {
      "from": "0xc100000000000000000000000000000000000001",
      "to":   "0xc100000000000000000000000000000000000001",
      "gas":  "0x186a0",
      "authorizationList": [
        {
          "chainId": "0x7b8b",
          "address": "0xc200000000000000000000000000000000000001",
          "nonce":   "0x1",
          "v":       "0x0",
          "r":       "0x...",
          "s":       "0x..."
        }
      ]
    }
  ]
}
```

Call the delegated EOA. After the install example above is included,
calling the EOA's address from any contract executes the delegate's
code in the EOA's storage and balance context:

```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "eth_call",
  "params": [
    { "to": "0xc100000000000000000000000000000000000001", "data": "0x..." },
    "latest"
  ]
}
```

Read the delegation designator. `eth_getCode` on the EOA returns the
23-byte designator `0xef0100 || target_address` — for the install
example above, the bytes are `0xef0100` followed by
`0xc200…01`. On-chain contracts detect a delegation by reading this
prefix via `EXTCODECOPY`. To clear a delegation, the EOA signs a fresh
authorization with `address = 0x0000000000000000000000000000000000000000`;
after inclusion `eth_getCode` returns `0x` and subsequent calls execute
as plain value-only transfers. Storage is not cleared on clear:
slots written by the previous delegate persist, and a later
re-delegation of the same authority will read them.

```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "method": "eth_getCode",
  "params": ["0xc100000000000000000000000000000000000001", "latest"]
}
```

For end-to-end examples that actually construct and submit `0x04`
transactions against a localnode, see the system suite under
`tests/system/test/Eip7702_*.test.ts` and the shared helpers in
`tests/system/test/helpers/eip7702.ts`.
