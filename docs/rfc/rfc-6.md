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
implementing additional safety mechanisms at the consensus level in the future,
as there is a natural window between a mint request and its execution.

### Overview

The mechanism works in two phases with a configurable block delay between them:

```
Block N:           Authorized party calls
                   AssetsBridge.bridgeTriparty(recipient, amount, callbackData)
                   and a TripartyBridgeRequest is written to x/bridge module state
                   with the current block height recorded as the request height.
                   The returned requestId identifies the request.

Block N+D (D>=1):  PreBlocker reads pending TripartyBridgeRequests whose request
                   height is at least D blocks in the past and mints BTC for each
                   mature request using mintBTC(). After a successful mint, the
                   PreBlocker issues a callback to the controller that submitted
                   the request. Requests that have not yet reached the required
                   delay are left in state for future blocks.
```

The delay `D` is configurable via `AssetsBridge.setTripartyBlockDelay` and
defaults to 1.

The advantage of this approach is a single minting point - all BTC minting happens
in the `PreBlocker` through the same `mintBTC()` function, regardless of the
originating signal. This allows maintaining the existing `verifyBTCSupply`
invariant and provides extensibility like tracking supplies separately or adding
additional safety mechanisms in the future (see [Future Work](#future-work)).

### `AssetsBridge` precompile

The existing `AssetsBridge` precompile is the entry point for bridge operations
and should expose the following functions:

```solidity
// Minting
function bridgeTriparty(address recipient, uint256 amount, bytes calldata callbackData) external returns (uint256 requestId);
function getTripartyRequestSequenceTip() external view returns (uint256 sequenceTip);
function getTripartyProcessedSequenceTip() external view returns (uint256 sequenceTip);

// Access control
function allowTripartyController(address controller, bool isAllowed) external returns (bool);
function isAllowedTripartyController(address controller) external view returns (bool);

// Pause
function pauseTriparty(bool isPaused) external returns (bool);
function isTripartyPaused() external view returns (bool isPaused);

// Block delay
function setTripartyBlockDelay(uint256 delay) external returns (bool);
function getTripartyBlockDelay() external view returns (uint256 delay);

// Limits
function setTripartyLimits(uint256 perRequestLimit, uint256 windowLimit) external returns (bool);
function getTripartyLimits() external view returns (uint256 perRequestLimit, uint256 windowLimit);
function getTripartyCapacity() external view returns (uint256 capacity, uint256 resetHeight);

// Provenance
function getTripartyTotalBTCMinted() external view returns (uint256 totalMinted);
```

* `bridgeTriparty` accepts the `recipient`, `amount`, and `callbackData`.
  The `callbackData` is arbitrary bytes forwarded to the callback, allowing the
  caller to pass context such as a lock duration or vault parameters. After BTC
  is minted, the `PreBlocker` issues a callback to the controller that submitted
  the request. The controller address is already trusted since only allowed
  triparty controllers can call `bridgeTriparty`, so there is no need for a
  separate callback address. If the callback fails, the mint still completes.
  Passing empty `callbackData` does not disable the callback. The function
  returns the `requestId` (the sequence number assigned to the request) which
  can be used to correlate the callback with the original request. Only an
  allowed triparty controller should be able to call this function.
* `getTripartyRequestSequenceTip` returns the last assigned request sequence
  number (the total number of triparty bridge requests submitted), mirroring
  `getCurrentSequenceTip` for the regular bridge path.
* `getTripartyProcessedSequenceTip` returns the last processed request sequence
  number. Together with `getTripartyRequestSequenceTip`, callers can derive the
  number of pending requests (`requestTip - processedTip`).
* `allowTripartyController` allows or disallows a triparty controller address.
  Only callable by `poaKeeper.CheckOwner()`.
* `isAllowedTripartyController` returns whether the given address is an allowed
  triparty controller.
* `pauseTriparty` sets a pause flag that prevents new triparty mint requests
  from being accepted and stops the `PreBlocker` from processing pending
  requests. Pending requests remain in state and will be processed once triparty
  is unpaused and limits allow it. Only callable by the `AssetsBridge` pauser.
* `isTripartyPaused` returns whether triparty bridging is currently paused.
* `setTripartyBlockDelay` sets the number of blocks that must pass between a
  triparty mint request and its execution by the `PreBlocker`. The delay must be
  at least 1 (the request and execution always happen in different blocks).
  Only callable by `poaKeeper.CheckOwner()`.
* `getTripartyBlockDelay` returns the configured block delay.
* `setTripartyLimits` configures the global request limits shared by all
  triparty controllers: `perRequestLimit` (the maximum BTC amount for a single
  `bridgeTriparty` call) and `windowLimit` (the maximum aggregate BTC amount
  that can be requested within a rolling block window, using the same reset
  mechanism as outflow limits). Only callable by `poaKeeper.CheckOwner()`.
* `getTripartyLimits` returns the configured per-request and window limits.
* `getTripartyCapacity` returns the remaining window capacity and the block
  height at which it resets, mirroring `getOutflowCapacity`.
* `getTripartyTotalBTCMinted` returns the cumulative BTC minted through the
  triparty path (see
  [Provenance tracking](#provenance-tracking-and-bridging-out)).

Additionally, `bridgeTriparty` should:

* Revert if the `recipient` is a blocked address (e.g. a module account).
* Revert if `amount` is below the minimum of 0.01 BTC. This hard-coded floor
  prevents a compromised controller from spamming the chain with many small
  requests and ensures each request carries enough weight for the future vote
  mechanism to be practical.
* Revert if `amount` exceeds the global per-request limit.
* Revert if `amount` would exceed the remaining request window capacity.

### `x/bridge` module

#### Triparty state

The `x/bridge` module should store the following new state:

* Triparty controller addresses: the addresses authorized to submit triparty
  mint requests
* Triparty block delay: the number of blocks that must elapse between a request
  and its execution.
* Triparty sequence tip: the sequence number of the last assigned triparty
  request, analogous to the `AssetsLockedSequenceTip`
* Pending triparty mint requests: a list of `TripartyBridgeRequest` entries
  awaiting processing by the `PreBlocker`. Each entry is assigned a
  monotonically increasing sequence number (used as the `requestId`) at creation
  time and records the block height at which it was created so the `PreBlocker`
  can determine maturity. Additionally, each entry stores the `recipient`,
  `amount`, `callbackData`, and the `controller` address that submitted the
  request. The sequence number ensures deterministic processing order across all
  validators, following the same pattern used for `AssetsLocked` events.

#### `PreBlocker` extension

The bridge `PreBlocker` currently processes `AssetsLockedEvents` extracted from
the injected pseudo-transaction. After processing bridge events, the `PreBlocker`
should additionally:

1. Read the configured triparty block delay `D` from state.
2. Read up to `TripartyBatch` pending `TripartyBridgeRequest` entries from the
   module state, starting from the lowest pending sequence number and proceeding
   in strictly increasing sequence order. `TripartyBatch` is a compile-time
   constant set to 5. While triparty mints are expected to be rare, capping the
   batch size provides defense in depth to ensure stable block times.
3. For each request in the batch whose recorded block height satisfies
   `currentHeight - requestHeight >= D`, call the existing `mintBTC()` function
   which mints coins through the `x/bank` module and updates the `BTCMinted`
   counter. Stop at the first immature request to preserve sequential processing.
   No request can be processed ahead of an earlier one that is not yet mature.
4. After a successful mint, issue an EVM call to
   `onTripartyBridgeCompleted(uint256 requestId, address recipient, uint256 amount, bytes callbackData)`
   on the controller that submitted the request, forwarding the stored
   `callbackData`. The call is executed via `ExecuteContractCall` with
   the bridge module address as the sender, following the same pattern used by
   `mintERC20`, but with a gas cap of 1,000,000. A callback failure should be
   logged but must not prevent the mint from completing or block subsequent
   requests. The BTC has already been minted and cannot be rolled back without
   risking a supply invariant violation.
5. Clear all processed requests from state.

This extension is deliberately minimal. The `mintBTC()` function is reused
without modification, ensuring the same minting logic and supply tracking apply
to both paths.

Blocked-address validation is enforced at the precompile level - `bridgeTriparty`
reverts if the recipient is a blocked address. The `PreBlocker` MUST repeat this
check before minting. Relying solely on upstream validation is dangerous -
if the precompile check is ever bypassed or a blocked address is added between
submission and processing, the `PreBlocker` would mint to a blocked address
and risk a consensus failure. Defense in depth requires the `PreBlocker` to
independently verify the recipient is not blocked.

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
considered trusted, we should implement per-request and per-window limits for
BTC minting via triparty.

Limits are global — shared across all triparty controllers — and configured via
`setTripartyLimits`. The window reset follows the same block-window mechanism
used by outflow limits (`OutflowResetBlocks`).

### Safety mechanisms

The triparty minting path benefits from the existing safety mechanisms of the
bridge module:

* Supply invariant: the `verifyBTCSupply` check in `EndBlock` verifies that
  `btc_supply = total_btc_minted - total_btc_burnt` and panics if violated.
  Since `mintBTC()` is reused, this invariant covers both paths without
  modification.
* Pause mechanism: when triparty minting is paused, the `PreBlocker` delays
  processing existing triparty requests until triparty is unpaused.
  All new requests are rejected.
* Access control: only configured triparty controller addresses can submit requests.
* Per-request limit: a global maximum amount per individual triparty mint
  request, shared across all controllers.
* Window limit: a global aggregate cap on accepted triparty request volume
  within a rolling block window, following the existing outflow limit reset
  pattern.

## Future Work

### Veto mechanism

The configurable block delay between a triparty mint request and its execution
creates a natural window for introducing a veto mechanism. A veto would allow
authorized parties to reject specific pending requests before they are processed
by the `PreBlocker`, permanently canceling the mint.

Unlike pausing (which uses a dedicated pause flag and affects all requests),
a veto would target individual requests by their `requestId`. A
vetoed request would be removed from state and never processed, and the BTC
would not be minted.

The veto mechanism is intentionally deferred to a future iteration. In the
current design, a triparty bridge request that is not vetoed has a guarantee of
being eventually processed once its block delay has elapsed and limits allow it.

### Bridge out limits

We do not consider BTC provenance during bridge outs since all BTC on the chain
is fungible. Instead, we assume triparty BTC is locked immediately in smart
contracts as, for example, collateral, and is not mixed on the chain with
BTC bridged through the locking mechanism.

In future iterations, we will revisit strengthening the validation of bridge-out
operations to check the state of reserves before processing bridge-out requests
and delay processing if there is not enough BTC locked in the bridge contract to
cover them. This mechanism will have to be aligned with other research projects,
like layered lending minting tokens through standard ERC20 operations on the
Mezo chain.
