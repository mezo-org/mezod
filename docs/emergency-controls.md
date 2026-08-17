# Emergency controls

This document describes the emergency controls of the Mezo chain: the Emergency
Team role, the lockdown mode it drives, and the audit trail both leave behind.

The Emergency Team role and the bridge lockdown ship in `v13.0.0`. They replace
the `AssetsBridge` pauser role and the triparty pause flag, which the same
release retires.

## Overview

Mezo has two control planes.

| Control plane | Holder         | Character                                     |
|---------------|----------------|-----------------------------------------------|
| Governance    | PoA owner      | slow, high signature threshold, broad powers  |
| Emergency     | Emergency Team | fast, low signature threshold, narrow powers  |

The governance plane owns every configuration change: validator management,
bridge limits and mappings, precompile versions and bytecode, chain parameters,
and upgrade plans. On mainnet the PoA owner is a governance Safe with a high
signature threshold, so collecting the signatures takes time.

The emergency plane exists because some incidents cannot wait for the
governance quorum. The Emergency Team is the quick technical multisig held by
the engineers on call: a Safe with a low signature threshold. Its power set is
deliberately small: it enables and disables lockdown levels and nothing else.
The PoA owner grants the role, rotates or revokes it at any time, and can take
every emergency action itself.

The `Maintenance` precompile is the single emergency surface. The emergency
methods need `Maintenance` precompile version 6, which the `v13.0.0` upgrade
activates. The same release removes the `AssetsBridge` methods `setPauser`,
`getPauser`, `pauseBridgeOut`, `pauseTriparty`, and `isTripartyPaused` from the
codebase and the ABI; calls to their selectors revert with
`method not found in ABI`.

## The Emergency Team role

The chain stores one Emergency Team address in the `x/poa` module, next to the
owner. An absent entry means the role is not granted. The genesis field is
`emergency_team`, a bech32 address, and an empty value means the role is not
granted. See [the PoA state spec](../x/poa/spec/01_state.md) for the store keys.

The role is a single address. The quorum lives inside the Safe, not on the
chain. The grant is a single step, unlike the 2-step PoA ownership transfer,
because the role must be quick to rotate. The owner corrects a mistake with a
second call.

### Checking the current holders

The chain is the source of truth for both roles. Through the Hardhat toolbox
(see `precompile/hardhat/README.md` for the setup):

```
npx hardhat maintenance:getEmergencyTeam
npx hardhat validatorPool:owner
```

`getEmergencyTeam` returns the zero address when the role is not granted. When
a returned address is a Safe, open it on
[safe.mezo.org](https://safe.mezo.org) to see the signer set and the signature
threshold.

The `v13.0.0` upgrade handler grants the role to the retired `AssetsBridge`
pauser. On mainnet the pauser is the quick technical multisig, so the ability
to stop the bridge stays continuous across the upgrade. A zero or absent pauser
grants nothing.

### Grant, rotate, and revoke

All three actions use one method of the `Maintenance` precompile:

- Address: `0x7b7c000000000000000000000000000000000013`
- ABI: `precompile/maintenance/abi.json`
- Interface: `precompile/maintenance/IMaintenance.sol`

```solidity
function setEmergencyTeam(address team) external returns (bool);
function getEmergencyTeam() external view returns (address team);
```

`setEmergencyTeam` is restricted to the PoA owner. It grants the role to `team`,
rotates the role when another address holds it, and revokes the role when `team`
is the zero address. Every call emits `EmergencyTeamSet`. `getEmergencyTeam`
returns the zero address when the role is not granted and is callable by anyone.

Through the Hardhat toolbox (see `precompile/hardhat/README.md` for the setup):

```
npx hardhat maintenance:setEmergencyTeam --signer OWNER --team <TEAM_ADDRESS>
npx hardhat maintenance:getEmergencyTeam
npx hardhat maintenance:setEmergencyTeam --signer OWNER --team 0x0000000000000000000000000000000000000000
```

On mainnet the PoA owner is a Safe, so the call is a Safe transaction. Use
[safe.mezo.org](https://safe.mezo.org) to craft it from the PoA owner Safe and
fill the fields as follows:

| Field  | Value                                                 |
|--------|-------------------------------------------------------|
| To     | `0x7b7c000000000000000000000000000000000013`          |
| Value  | `0`                                                   |
| ABI    | the contents of `precompile/maintenance/abi.json`     |
| Method | `setEmergencyTeam`                                    |
| `team` | the new Emergency Team, or the zero address to revoke |

Pass the transaction JSON to the Mezo Governance Safe signers, the same way as
an upgrade plan. See [Governance](./release-process.md#governance).

## Lockdown mode

Lockdown is the only power of the Emergency Team. It comes in levels, and each
level stops more activity than the level below it. MEZO-5000 defines the model;
`v13.0.0` implements level 1.

| Level | Name                    | Method                                        | Disable on chain | Status             |
|-------|-------------------------|-----------------------------------------------|------------------|--------------------|
| 1     | Bridge lockdown         | `setBridgeLockdown(bool,bool)`                | yes              | `v13.0.0`          |
| 2     | Restricted transactions | `setTxLockdown` plus `setTxLockdownAllowlist` | yes              | planned, MEZO-5000 |
| 3     | No transactions         | `setTxLockdown`                               | yes              | planned, MEZO-5000 |
| 4     | Chain lockdown          | `setChainLockdown`                            | no               | planned, MEZO-5000 |

### The interface contract

Every level follows one shape:

- one setter named `set<Level>Lockdown`, taking the enable state plus the
  arguments of that level;
- one view named `get<Level>Lockdown`, callable by anyone;
- one event named `<Level>LockdownSet`.

Level 4 is the exception. `setChainLockdown` takes no arguments, has no view,
and cannot be disabled on chain.

Both the Emergency Team and the PoA owner can call every setter.

### Level 1: bridge lockdown

```solidity
function setBridgeLockdown(bool bridgeIn, bool bridgeOut) external returns (bool);
function getBridgeLockdown() external view returns (bool bridgeIn, bool bridgeOut);
```

`setBridgeLockdown` is restricted to the Emergency Team and the PoA owner. It
writes both flags on every call, so pass the full wanted state:
`setBridgeLockdown(true, true)` stops both directions and
`setBridgeLockdown(false, false)` disables the level. Every call emits
`BridgeLockdownSet`.

`bridgeOut = true` stops every bridge-out: BTC and mapped ERC20 tokens, to
Ethereum and to Bitcoin. The check sits in `SaveAssetsUnlocked` in the
`x/bridge` module, before the outflow limit check. A blocked bridge-out fails
with `bridge-out is paused`, the sequence tip does not advance, and no outflow
is recorded.

`bridgeIn = true` stops both inbound paths: the triparty path and the
validator-observed path.

On the triparty path, `bridgeTriparty` calls fail with `bridge-in is paused`
and the `PreBlocker` stops processing pending triparty requests. Pending
requests stay in state.

On the validator-observed path, the check sits in the proposal handlers of the
`x/bridge` module. `PrepareProposal` skips the injection of the `AssetsLocked`
pseudo-transaction, and `ProcessProposal` rejects any proposal that carries
one. No `AssetsLocked` event observed on Ethereum is executed while `bridgeIn`
is true, and the sequence tip does not advance.

Lockdown never touches the bridge outflow limits. Limits stay a separate PoA
owner tool, and the lockdown state is readable through `getBridgeLockdown`
instead of masquerading as zero limits.

Through the Hardhat toolbox:

```
npx hardhat maintenance:setBridgeLockdown --signer TEAM --bridge-in true --bridge-out true
npx hardhat maintenance:getBridgeLockdown
npx hardhat maintenance:setBridgeLockdown --signer TEAM --bridge-in false --bridge-out false
```

### Recovery

Read `getBridgeLockdown` first. The setter writes both flags, so a call that
only releases one direction must pass the current value of the other flag.

To release bridge-out, call `setBridgeLockdown(bridgeIn, false)` with the
current `bridgeIn`. Bridge-outs are accepted again immediately and the outflow
limits apply as before. Nothing needs a replay, because the rejected
bridge-outs never entered the state.

To release bridge-in, call `setBridgeLockdown(false, bridgeOut)` with the
current `bridgeOut`. The `PreBlocker` resumes the pending triparty requests,
subject to the block delay and the triparty limits. Requests that controllers
attempted during the lockdown were rejected at submission, so those controllers
must submit them again.

The validator-observed path needs no manual step. The lockdown defers the
`AssetsLocked` events; it does not drop them. The sequence tip stays frozen
while `bridgeIn` is true, so the Ethereum sidecar keeps serving the events from
the tip. After the release, the next proposal injects them again and the chain
processes them in sequence order. No event is lost.

Confirm the result with `getBridgeLockdown` and with the `BridgeLockdownSet`
event of the transaction.

## Audit trail

Every write of the emergency controls emits an event from the `Maintenance`
precompile at `0x7b7c000000000000000000000000000000000013`. The views emit
nothing.

| Event               | Signature                           | Topic 0                                                              |
|---------------------|-------------------------------------|----------------------------------------------------------------------|
| `EmergencyTeamSet`  | `EmergencyTeamSet(address,address)` | `0xc22e0a53d80be0ed688451dd96632727b45a62185ad4bcf8c30014ba1bceb04b` |
| `BridgeLockdownSet` | `BridgeLockdownSet(bool,bool)`      | `0x60257cbb964d7216fa05325ba9832a1c24fcf2999621d6cd1ec6f751ff119f9f` |

`EmergencyTeamSet(address indexed previous, address indexed current)` records
every grant, rotation, and revocation. `previous` is the holder before the call
and is the zero address when the role was not granted. `current` is the new
holder and is the zero address on a revocation. Both arguments are indexed, so
`previous` is topic 1 and `current` is topic 2.

`BridgeLockdownSet(bool bridgeIn, bool bridgeOut)` records the full lockdown
state after the call. Neither argument is indexed, so both sit in the log data.

Query the whole history of the role with `eth_getLogs`:

```
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_getLogs","params":[{"fromBlock":"0x0","toBlock":"latest","address":"0x7b7c000000000000000000000000000000000013","topics":["0xc22e0a53d80be0ed688451dd96632727b45a62185ad4bcf8c30014ba1bceb04b"]}],"id":1}' -H "Content-Type: application/json" https://rpc.test.mezo.org
```

Add a third topic to find only the calls that granted the role to one address.
The topic is the address left-padded with zeros to 32 bytes:

```
"topics":[
  "0xc22e0a53d80be0ed688451dd96632727b45a62185ad4bcf8c30014ba1bceb04b",
  null,
  "0x000000000000000000000000<TEAM_ADDRESS_WITHOUT_0x>"
]
```

The transactions themselves carry the rest of the trail: the sender is the
Emergency Team Safe or the PoA owner, so the Safe history names the signers.

## Boundary: Emergency Team and PoA owner

| Action                                                                             | Emergency Team | PoA owner |
|------------------------------------------------------------------------------------|----------------|-----------|
| Enable / disable bridge lockdown (level 1)                                         | yes            | yes       |
| Enable / disable future lockdown levels (MEZO-5000; chain lockdown is enable-only) | yes            | yes       |
| Grant / revoke the Emergency Team role                                             | no             | yes       |
| Outflow limits, min amounts, ERC20 mappings                                        | no             | yes       |
| Triparty controllers and limits                                                    | no             | yes       |
| Upgrade plans, precompile bytecode, EVM/feemarket params                           | no             | yes       |
| Validator applications, kick, privileges                                           | no             | yes       |

Both parties can enable and disable every level, so no action waits for a
separate governance confirmation. Level 4 is the exception: once chain lockdown
is enabled, nobody disables it on chain.

## Roadmap

MEZO-5000 continues the level model with three reserved method names. None of
them exists yet; the names are fixed here so the vocabulary stays stable.

- `setTxLockdown(bool enabled)`: transaction lockdown. Enabling it installs the
  defaults. Allowed senders are the Emergency Team and the PoA owner. Allowed
  targets are the precompiles needed to diagnose and recover the chain.
- `setTxLockdownAllowlist(address[] senders, address[] targets)`: adds senders
  and targets beyond the defaults of `setTxLockdown`. Level 2 is
  `setTxLockdown(true)` plus this allowlist, and level 3 is `setTxLockdown(true)`
  with the defaults only. There is no separate level 2 method.
- `setChainLockdown()`: halts consensus through an immediate upgrade. It takes no
  arguments and is one-way on chain. The only recovery is validators restarting
  with the next-version binary, so this level has no on-chain disable and no
  view.
