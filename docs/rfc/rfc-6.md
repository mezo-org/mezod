# RFC-6: Triparty BTC minting

## Background

BTC is bridged from Ethereum to Mezo using the native bridge described in
[RFC-2](./rfc-2.md) and extended for non-Bitcoin assets in [RFC-4](./rfc-4.md).
In this model, `AssetsLocked` events emitted by the `MezoBridge`
contract on Ethereum are observed by the Ethereum sidecar, propagated through
vote extensions, and executed in the `PreBlocker` which mints BTC using the
`x/bank` module. The bridge-out mechanism described in [RFC-5](./rfc-5.md)
completes the picture by allowing assets to flow from Mezo back to Ethereum and
Bitcoin.

Currently, BTC minting on Mezo is exclusively driven by the native bridge
infrastructure. There is no mechanism for minting BTC based on signals originating
from outside of the sidecar.

This RFC proposes an alternative BTC minting path that allows authorized addresses
to request BTC minting on Mezo through the existing bridge infrastructure.
The design accommodates any triparty setup and a specific triparty mechanism is
out of the scope of this document. The only requirement for a triparty mechanism
is that it should be able to issue transactions to the precompile.

## Proposal

The proposal reuses the existing bridge infrastructure, making triparty minting
an extension to the bridge rather than a parallel system. In this model, all BTC
minted on the chain goes through the same flow, and the existing total supply
checks in the `EndBlock` are preserved. Additionally, this approach enables
implementing a veto mechanism at the consensus level in the future, as there is
a natural window between a mint request and its execution.

### Overview

The mechanism works in two phases with a configurable block delay between them:

```
Block N:           Authorized party calls AssetsBridge.bridgeTriparty(recipient, amount)
                   and a TripartyBridgeRequest is written to x/bridge module state
                   with the current block height recorded as the request height.

Block N+D (D>=1):  PreBlocker reads pending TripartyBridgeRequests whose request
                   height is at least D blocks in the past and mints BTC for each
                   mature request using mintBTC(). Requests that have not yet
                   reached the required delay are left in state for future blocks.
```

The delay `D` is configurable via `AssetsBridge.setTripartyBlockDelay` and
defaults to 1.

The advantage of this approach is a single minting point - all BTC minting happens
in the `PreBlocker` through the same `mintBTC()` function, regardless of the
originating signal. This allows maintaining the existing `verifyBTCSupply`
invariant and provides extensibility like tracking supplies separately or adding
a consensus-level veto mechanism.

### `AssetsBridge` precompile

The existing `AssetsBridge` precompile is the entry point for bridge operations
and should expose the following functions:

```solidity
function bridgeTriparty(address recipient, uint256 amount) external returns (bool);

function allowTripartyController(address controller, bool isAllowed) external returns (bool);

function pauseTriparty(bool isPaused) external returns (bool);

function setTripartyBlockDelay(uint256 delay) external returns (bool);
```

Only an allowed triparty controller should be able to call the `bridgeTriparty`
function. The `allowTripartyController` and `setTripartyBlockDelay` functions
should only be callable by the same address that can set the pauser or outflow
limits, which is `poaKeeper.CheckOwner()`). The `pauseTriparty` should only be
callable by the `AssetsBridge` pauser.

`setTripartyBlockDelay` sets the number of blocks that must pass between a
triparty mint request and its execution by the `PreBlocker`. The delay must be
at least 1 (the request and execution always happen in different blocks).

Additionally, `bridgeTriparty` should respect the paused state and revert if
triparty bridging is paused.

### `x/bridge` module

#### Triparty state

The `x/bridge` module should store the following new state:

* Triparty controller addresses: the addresses authorized to submit triparty
  mint requests
* Triparty block delay: the number of blocks that must elapse between a request
  and its execution.
* Triparty sequence tip: the sequence number of the last processed triparty
  request, analogous to the `AssetsLockedSequenceTip`
* Pending triparty mint requests: a list of `TripartyBridgeRequest` entries
  awaiting processing by the `PreBlocker`. Each entry is assigned a
  monotonically increasing sequence number at creation time and records the
  block height at which it was created so the `PreBlocker` can determine
  maturity. The sequence number ensures deterministic processing order across
  all validators, following the same pattern used for `AssetsLocked` events.

#### `PreBlocker` extension

The bridge `PreBlocker` currently processes `AssetsLockedEvents` extracted from
the injected pseudo-transaction. After processing bridge events, the `PreBlocker`
should additionally:

1. Read the configured triparty block delay `D` and the current triparty
   sequence tip from state
2. Read up to `TripartyBatch` pending `TripartyBridgeRequest` entries from the
   module state, starting from the request whose sequence number is one greater
   than the current tip and proceeding in strictly increasing sequence order.
   `TripartyBatch` is a compile-time constant set to 5. While triparty mints
   are expected to be rare, capping the batch size provides defense in depth to
   ensure stable block times.
3. For each request in the batch whose recorded block height satisfies
   `currentHeight - requestHeight >= D`, call the existing `mintBTC()` function
   which mints coins through the `x/bank` module and updates the `BTCMinted`
   counter. Stop at the first immature request to preserve sequential processing.
   No request can be processed ahead of an earlier one that is not yet mature.
4. Update the triparty sequence tip to the sequence number of the last processed
   request and clear all processed requests from state.

This extension is deliberately minimal. The `mintBTC()` function is reused
without modification, ensuring the same minting logic and supply tracking apply
to both paths.

The code processing minting requests should skip requests minting to blocked
addresses like module addresses. This follows the same logic used for processing
`AssetsLocked` events.

### Provenance tracking and bridging out

To distinguish BTC minted through the native bridge from BTC minted through the
triparty path, the `x/bridge` module should maintain a separate counter tracking
the total BTC minted from triparty requests. This counter is informational
and does not affect the supply invariant - the existing `BTCMinted` counter
remains the source of truth for the `verifyBTCSupply` check. The provenance
information is useful for monitoring purposes.

We are not going to consider the provenance information during bridge outs. All
BTC is fungible on the chain and the triparty mechanism needs to ensure BTC
minted based on triparty agreements is not mixed with other BTC in a way allowing
triparty BTC to leave the chain making the bridge insolvent. This can be achieved,
for example, by minting only to vaults that use triparty BTC only in very specific
way, like locking it into veBTC.

### Triparty limits

The regular bridge path requires 2/3+ validator consensus to mint BTC. The
triparty path requires a call from a specific address. This is a fundamentally
different security model, and even though the triparty controller address is
considered trusted, we should implement per-request and per-24h limits for
BTC minting via triparty.

### Safety mechanisms

The triparty minting path benefits from the existing safety mechanisms of the
bridge module:

* Supply invariant: the `verifyBTCSupply` check in `EndBlock` verifies that
  `btc_supply = total_btc_minted - total_btc_burnt` and panics if violated.
  Since `mintBTC()` is reused, this invariant covers both paths without
  modification.
* Pause mechanism: when triparty minting is paused, the `PreBlocker` should
  skip processing existing triparty requests and new requests should be rejected.
* Access control: only configured triparty controller addresses can submit requests.
* Per-request limits: a maximum amount per individual triparty mint request.
* Per-24h limits: an aggregate cap on triparty minting within a 24h reset period,
  following the existing outflow limit pattern.
