# Upgrades

This document describes the process for breaking chain upgrades.
Moreover, it describes all historical upgrades performed on the Mezo chain.

## The process

Mezo chain client defines two basic primitives for breaking upgrades. They are  
both part of the `app/upgrades` module:

- `Fork`: This component represents a hard fork upgrade that executes a one-time
  code block upon a programmatically enforced block. A `Fork` can schedule an
  `Upgrade` or activate state/logic changes at a specific height, using the
  feature flag approach. `Fork` changes are executed on a running chain.
- `Upgrade`: This component typically defines state changes that require migrations
  of the store schema. Moreover, it can be used to roll out complex logic changes
  that cannot be applied using a feature flag. `Upgrade` changes require
  a coordinated chain halt to be executed. An `Upgrade` can be scheduled either
  through a `Fork` or through the `Upgrade` precompile. The store migrations are
  executed using the built-in
  [Cosmos In-Place Store Migrations mechanism](https://docs.cosmos.network/v0.52/learn/advanced/upgrade).

Those two primitives can be combined to perform different upgrade procedures
depending on the requirements of the upgrade. Each procedure assumes
all Mezo chain clients run on version `v1.0.0` initially.

### Hard fork upgrade without chain halt

This procedure is suitable for upgrades that do not require any store migrations
nor complex logic changes. A good example is an upgrade of consensus parameters.

The procedure is as follows:

1. Create a package for the new upgrade: `app/upgrades/v2`.
2. Define a new `v2.Fork` in the `app/upgrades/v2/constants.go` file and the
   fork logic in the `app/upgrades/v2/forks.go` file.
3. Register the `v2.Fork` in the `app.Forks` list, in the `app/upgrades.go` file.
4. Release version `v2.0.0` of the Mezo chain client and ask validators to upgrade.
5. Once validators reach the fork block, the fork code will be executed.

After the upgrade, the `v2.0.0` clients are still able to validate past blocks produced
by `v1.0.0` clients. New nodes can sync from the genesis block using `v2.0.0` directly.

### Hard fork upgrade with chain halt

This procedure is suitable for urgent upgrades that require store migrations
or complex logic changes. A good example is adding/removing a module or changing
the store schema as response to a critical vulnerability.

The procedure is as follows:

1. Create a package for the new upgrade: `app/upgrades/v2`.
2. Define a new `v2.Fork` in the `app/upgrades/v2/constants.go` file and the
   fork logic in the `app/upgrades/v2/forks.go` file. The fork logic
   should schedule the `v2` upgrade plan in the `x/upgrade` module. The plan
   should have the same height as the fork block.
3. Define a new `v2.Upgrade` in the `app/upgrades/v2/constants.go` file and the
   upgrade handler in the `app/upgrades/v2/upgrades.go` file. The upgrade handler
   should perform all necessary store migrations.
4. Register the `v2.Fork` in the `app.Forks` list, in the `app/upgrades.go` file.
5. Register the `v2.Upgrade` in the `app.Upgrades` list, in the `app/upgrades.go` file.
6. Release version `v2.0.0` of the Mezo chain client.
7. Backport the `v2.Fork` to `v1.0.0` and release version `v1.0.1` of the Mezo
   chain client. Ask validators to upgrade to `v1.0.1` first. It is important
   to NOT BACKPORT the `v2.Upgrade` to `v1.0.0` as it will break the
   Cosmos In-Place Store Migrations mechanism.
8. Once validators reach the fork block, the fork code will be executed and the
   chain will halt due to the `v2` upgrade plan.
9. After the halt, validators should upgrade to `v2.0.0` and restart their nodes.

After the upgrade, the `v2.0.0` clients are no longer able to validate past blocks
produced by `v1.0.0` clients. New nodes must sync from the genesis block using
`v1.0.1` first and then upgrade to `v2.0.0` once the fork block is reached.
This process can be automated using Cosmovisor. Alternatively, nodes can
use the state sync process that starts syncing from a snapshot block
compatible with the latest version.

### Planned upgrade with chain halt

This procedure is suitable for planned upgrades that require store migrations
or complex logic changes. A good example is adding/removing a module or changing
the store schema as part of a planned feature.

The procedure is as follows:

1. Create a package for the new upgrade: `app/upgrades/v2`.
2. Define a new `v2.Upgrade` in the `app/upgrades/v2/constants.go` file and the
   upgrade handler in the `app/upgrades/v2/upgrades.go` file. The upgrade handler
   should perform all necessary store migrations.
3. Register the `v2.Upgrade` in the `app.Upgrades` list, in the `app/upgrades.go` file.
4. Release version `v2.0.0` of the Mezo chain client.
5. The governance schedules the `v2` upgrade plan through the `Upgrade` precompile.
6. Once validators reach the upgrade block, the chain will halt due to the `v2` upgrade plan.
7. After the halt, validators should upgrade to `v2.0.0` and restart their nodes.

After the upgrade, the `v2.0.0` clients are no longer able to validate past blocks
produced by `v1.0.0` clients. New nodes must sync from the genesis block using
`v1.0.0` first and then upgrade to `v2.0.0` once the fork block is reached.
This process can be automated using Cosmovisor. Alternatively, nodes can
use the state sync process that starts syncing from a snapshot block
compatible with the latest version.

## The Upgrade precompile

The `Upgrade` precompile (`precompile/upgrade`), serves as an EVM interface to the
[x/upgrade module](https://docs.cosmos.network/main/build/modules/upgrade) and is used by
the [Planned upgrade with chain halt](#planned-upgrade-with-chain-halt) scenario described above.

### Upgrade Plan

The `x/upgrade` module defines a `Plan` type in which a live upgrade is scheduled to occur.
A `Plan` can be scheduled at a specific block `Height`. A `Plan` is submitted via the `Upgrade`
precompile after a Mezo chain client release with an appropriate upgrade handler.

An upgrade `Plan` has the following values:

- `Name`: The upgrade name (corresponds to the handler name)
- `Height`: The block height old clients should halt at to prevent state corruption
- `Info`: Any metadata about the upgrade (e.g. urls to updated binaries or git commit hash)

### Upgrade precompile API

- Address: `0x7b7c000000000000000000000000000000000014`
- ABI: `precompile/upgrade/abi.json`
- Interface: `precompile/upgrade/IUpgrade.sol`

The `Upgrade` precompile provides 3 methods:

- `plan`: Returns the latest upgrade plan
- `submitPlan`: Submits a new upgrade plan (restricted to the validator pool owner)
- `cancelPlan`: Cancels an upgrade plan (restricted to the validator pool owner)

And emits 2 events:

- `PlanSubmitted`: Emitted when a new upgrade plan is submitted
- `PlanCanceled`: Emitted when an upgrade plan is canceled

### Interaction via the Hardhat toolbox

Hardhat tasks are used to simplify interacting with Mezo precompiles. More information,
including setup steps and usage can be found in `precompile/hardhat/README.md`

After the hardhat environment and accounts have been configured, interaction with the
`Upgrade` precompile is possible with the following commands:

```
npx hardhat upgrade:submitPlan --signer OWNER --name NAME --height HEIGHT --info "{}"
npx hardhat upgrade:plan
npx hardhat upgrade:cancelPlan --signer OWNER
```

## Historical upgrades

### Versioning

Mezo client was using `v0.Y.Z-rcN` versioning pattern for the initial phase
of the project where only testnet was available. The major version was fixed
to `0` before mainnet readiness was reached. Moreover, all testnet-only
versions were release candidates (`-rcN`) to indicate that the software was
still in the testing phase.

Since mainnet readiness was reached, the major version was bumped to `1`
and proper semantic versioning (`vX.Y.Z`) started to be used from there.
The first mainnet release is `v1.0.0`.

### Testnet

Here is the list of upgrades performed on the Mezo Matsnet testnet.
For testnet-only versions, the `-rcN` suffix is omitted for brevity.
In that case, always assume the latest `-rcN` suffix for the given version.

Consult the [tags list](https://github.com/mezo-org/mezod/tags) for full version information.

| Version  | Block   | Type                                 | Details                                                                                                                                                                                 |
|----------|---------|--------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `v0.1.0` | 1       | N/A                                  | Initial genesis version.                                                                                                                                                                |
| `v0.2.0` | 496901  | Hard fork upgrade without chain halt | Change gas formula for the `ValidatorPool` precompile. <br/>This change was done before the `Fork` primitive was introduced. <br/>It was executed by introducing versioned precompiles. |
| `v0.3.0` | 1093500 | Hard fork upgrade with chain halt    | Introduce the Connect price oracle.                                                                                                                                                     |
| `v0.4.0` | 1745000 | Hard fork upgrade without chain halt | Update EVM storage root strategy (fix for Mezo Passport create2 problem) and introduce EVM observability for the BTC bridge.                                                            |
| `v0.5.0` | 2213000 | Hard fork upgrade without chain halt | On-chain precompile versioning. New Upgrade and PriceOracle precompiles and upgrade of the existing Maintenance precompile.                                                             |
| `v0.6.0` | 2563000 | Hard fork upgrade without chain halt | Introduce the ERC20 bridge and the BTC supply assertion.                                                                                                                                |
| `v0.7.0` | 3078794 | Hard fork upgrade without chain halt | Fix security issues in the EVM state DB. Introduce proper reverts for precompiles. Add chain fee splitter support. Disable Cosmos transactions.                                         |
| `v1.0.0-rc0` | 3569000 | Planned upgrade with chain halt    | Patch for a DoS vector in the bridge. Fix for the precompile revert mechanism.                                                                                                           |
| `v1.0.0-rc1` | 3712500 | Planned upgrade with chain halt    | Patch for mixed precompile addresses.                                                                                                           |
| `v2.0.2` | 5559500 | Planned upgrade with chain halt    | Bring back parity with mainnet.                                                                                                           |
| `v3.0.0` | 5695000 | Planned upgrade with chain halt    | [v3.0.0 release notes](https://github.com/mezo-org/mezod/releases/tag/v3.0.0) |
| `v4.0.0` | 6854500 | Planned upgrade with chain halt    | [v4.0.0 release notes](https://github.com/mezo-org/mezod/releases/tag/v4.0.0) |

### Mainnet

| Version  | Block   | Type                                 | Details                                                                       |
|----------|---------|--------------------------------------|-------------------------------------------------------------------------------|
| `v1.0.0` | 1       | N/A                                  | Initial genesis version.                                                      |
| `v2.0.0` | 706500  | Planned upgrade with chain halt      | [v2.0.0 release notes](https://github.com/mezo-org/mezod/releases/tag/v2.0.0) |
| `v3.0.0` | 1735000 | Planned upgrade with chain halt      | [v3.0.0 release notes](https://github.com/mezo-org/mezod/releases/tag/v3.0.0) |
