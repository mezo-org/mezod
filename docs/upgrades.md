# Upgrades

This document describes the process for breaking chain upgrades.
Moreover, it describes all historical upgrades performed on the Mezo chain.

## The process

Mezo chain client defines two basic primitives for breaking upgrades. They are  
both part of the `app/upgrades` module:

- `Fork`: This component represents a hard fork upgrade that executes a one-time
  code block upon a programmatically enforced block. A `Fork` can schedule an
  `Upgrade` with state migrations or just perform simple state changes (e.g. parameter upgrades).
- `Upgrade`: This component defines a set of state migrations. An `Upgrade`
  can be scheduled either through a `Fork` or through the `Upgrade` precompile.
  The state migrations are executed using the built-in
  [Cosmos In-Place Store Migrations mechanism](https://docs.cosmos.network/v0.52/learn/advanced/upgrade).

> [!IMPORTANT]
> The `Upgrade` precompile is still under development and not yet available in the Mezo chain client.

Those two primitives can be combined to perform different upgrade procedures
depending on the requirements of the upgrade. Each procedure assumes
all Mezo chain clients run on version `v1.0.0` initially.

### Hard fork upgrade without state migrations

This procedure is suitable for upgrades that do not require any state migrations.
A good example is an upgrade of consensus parameters.

The procedure is as follows:

1. Create a package for the new upgrade: `app/upgrades/v2`.
2. Define a new `v2.Fork` in the `app/upgrades/v2/constants.go` file and the
   fork logic in the `app/upgrades/v2/forks.go` file.
3. Register the `v2.Fork` in the `app.Forks` list, in the `app/upgrades.go` file.
4. Release version `v2.0.0` of the Mezo chain client and ask validators to upgrade.
5. Once validators reach the fork block, the fork code will be executed.

After the upgrade, the `v2.0.0` clients are still able to validate past blocks produced
by `v1.0.0` clients. New nodes can sync from the genesis block using`v2.0.0` directly.

### Hard fork upgrade with state migrations

This procedure is suitable for urgent upgrades that require state migrations.
A good example is adding/removing a module or changing the state schema as
response to a critical vulnerability.

The procedure is as follows:

1. Create a package for the new upgrade: `app/upgrades/v2`.
2. Define a new `v2.Fork` in the `app/upgrades/v2/constants.go` file and the
   fork logic in the `app/upgrades/v2/forks.go` file. The fork logic
   should schedule the `v2` upgrade plan in the `x/upgrade` module. The plan
   should have the same height as the fork block.
3. Define a new `v2.Upgrade` in the `app/upgrades/v2/constants.go` file and the
   upgrade handler in the `app/upgrades/v2/upgrades.go` file. The upgrade handler
   should perform all necessary state migrations.
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

### Planned upgrade with state migrations

This procedure is suitable for planned upgrades that require state migrations.
A good example is adding/removing a module or changing the state schema
as part of a planned feature.

The procedure is as follows:

1. Create a package for the new upgrade: `app/upgrades/v2`.
2. Define a new `v2.Upgrade` in the `app/upgrades/v2/constants.go` file and the
   upgrade handler in the `app/upgrades/v2/upgrades.go` file. The upgrade handler
   should perform all necessary state migrations.
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

## Historical upgrades

### Versioning

Mezo client uses `v0.Y.Z-rcN` versioning pattern for the initial phase
of the project where only testnet is available. The major version is fixed
to `0` before mainnet readiness is reached. Moreover, all testnet
version all release candidates (`-rcN`) to indicate that the software is
still in the testing phase.

Once mainnet readiness is reached, the major version will be bumped to `1`
and proper semantic versioning will be used (`vX.Y.Z`) from there.
The first mainnet release will be `v1.0.0`.

### Testnet

Here is the list of upgrades performed on the Mezo Matsnet testnet.
The `-rcN` suffix is omitted for brevity. Always assume the latest `-rcN` suffix
for the given version. Consult the <!-- markdown-link-check-disable-line --> [tags list](https://github.com/mezo-org/mezod/tags)
for full version information.

| Version  | Block   | Type                                       | Details                                                                                                                                                                                 |
|----------|---------|--------------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `v0.1.0` | 1       | N/A                                        | Initial genesis version.                                                                                                                                                                |
| `v0.2.0` | 496901  | Hard fork upgrade without state migrations | Change gas formula for the `ValidatorPool` precompile. <br/>This change was done before the `Fork` primitive was introduced. <br/>It was executed by introducing versioned precompiles. |
| `v0.3.0` | 1093500 | Hard fork upgrade with state migrations    | Introduce the Connect price oracle.                                                                                                                                                     |
| `v0.4.0` | 1745000 | Hard fork upgrade without state migrations | Update EVM storage root strategy (fix for Mezo Passport create2 problem) and introduce EVM observability for the BTC bridge.                                                            |

### Mainnet

Mainnet is not yet launched.
