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

The mechanism works in two phases spanning two blocks, similarly to bridge outs:

```
Block N:   Authorized party calls AssetsBridge.bridgeTriparty(recipient, amount)
           and a TripartyBridgeRequest is written to x/bridge module state

Block N+1: PreBlocker reads pending TripartyBridgeRequests from state and mints
           BTC for each request using mintBTC()
```

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

function allowTripartyController(address controller, bool isAllowed) external;

function pauseTriparty(bool isPaused);
```

Only an allowed triparty controller should be able to call the `bridgeTriparty`
function. The `allowTripartyController` function should only be callable by the
the same address that can set the pauser or outflow limits, which is
`poaKeeper.CheckOwner()`). The `pauseTriparty` should only be callable by the
`AssetsBridge` pauser.

Additionally, `bridgeTriparty` should respect the paused state and revert if
triparty bridging is paused.

### `x/bridge` module

#### Triparty state

The `x/bridge` module should store the following new state:

* Triparty controller addresses: the addresses authorized to submit triparty
  mint requests
* Pending triparty mint requests: a list of `TripartyBridgeRequest` entries
  awaiting processing by the `PreBlocker`

#### `PreBlocker` extension

The bridge `PreBlocker` currently processes `AssetsLockedEvents` extracted from
the injected pseudo-transaction. After processing bridge events, the `PreBlocker`
should additionally:

1. Read all pending `TripartyBridgeRequest` entries from the module state
2. For each valid request, call the existing `mintBTC()` function which mints coins
   through the `x/bank` module and updates the `BTCMinted` counter
3. Clear all processed requests from state

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
