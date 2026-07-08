# Incident report: 2026-07-02 LevelDB corruption

In the early morning UTC of July 2, 2026, the archive node of the Mezo testnet started
crash-looping because its main database was corrupted. In the following days, at least two
mainnet validator operators reported the same corruption on their nodes. Neither chain
halted and no funds were at risk, but the affected nodes needed days to get back to normal.

The root cause is general and affects every `mezod` node: **killing a running `mezod`
process without letting it shut down cleanly can corrupt its databases.** This report
explains how that happens, how to recognize it, how to recover, and what to do so it does
not happen to your node.

## Summary

`mezod` stores all chain data in [LevelDB](https://github.com/google/leveldb) databases
(the `data/` directory of the node home). Each database has a `MANIFEST` file that acts as
its index. If the process is killed in the middle of writing (SIGKILL, out-of-memory kill,
power loss), the `MANIFEST` can be left half-written or otherwise unreadable, and LevelDB
then refuses to open the database. The node crash-loops on startup from that point on.

Two separate chains of events led to such kills during this incident:

- **Testnet:** a resource-exhaustion vulnerability in the `EthCall`, `EstimateGas` and
  `SimulateV1` query handlers (fixed in [v11.0.1]) allowed unauthenticated queries to make
  `mezod` consume memory without bound. On July 2, a wave of such queries hit the publicly
  exposed query ports of all five testnet validators, and the operating system repeatedly
  killed the processes with an out-of-memory (OOM) kill, 33 times within 46 minutes. One
  of the five validators also serves as the testnet archive node, and its database did not
  survive its kill. Mezo's production nodes expose only the peer-to-peer port publicly and
  saw no OOM kills.
- **Mainnet:** operators restarted their nodes to apply the v11.0.1 fix. Every previous
  major upgrade was a coordinated chain-halt upgrade where `mezod` stops on its own and
  closes its databases cleanly. This was the first large wave of *hot restarts* of live,
  actively writing nodes, and the default stop timeouts (10 seconds for Docker Compose,
  30 seconds for Kubernetes) turned out to be too short: when the timeout expires, the
  node is killed mid-write. At least two of the roughly eight operators who restarted
  during those days hit the corruption. Their nodes were regular (non-archive) nodes, so
  they could be restored by re-syncing (recovery option 2 below).

The testnet kept producing blocks throughout, since four of five validators are enough for
consensus. The corrupted archive node held the only full copy of testnet history, so
instead of re-syncing it (which would lose that history) its 90 GiB database was repaired
with an experimental offline rebuild (recovery option 3 below).

## Timeline

| Day | Time (UTC) | Event |
|-----|------------|-------|
| 2026-06-25 | | Resource-exhaustion vulnerability in query handlers is reported privately to the development team |
| 2026-07-02 | 04:06–04:52 | Query wave triggers 33 OOM kills of `mezod` across all five testnet validators |
| 2026-07-02 | 04:41 | Testnet archive node enters a crash loop with a corrupted database |
| 2026-07-06 | | [v11.0.1] is released, fixing the vulnerability |
| 2026-07-06/07 | | Mainnet operators restarting for v11.0.1 report the same corruption |
| 2026-07-07 | | Shutdown grace period fix ships in [validator-kit] 12.1.0; archive database rebuilt |
| 2026-07-08 | | Archive node caught up, historical state verified, incident closed |

## Symptoms

A node with a corrupted LevelDB database crash-loops on startup. The log ends with an
error like:

```text
panic: ... leveldb: manifest corrupted (field 'comparer'): missing [file=MANIFEST-3772245]
```

The file name and the field vary. Any of the databases under `data/` can be affected:
`application.db`, `blockstore.db`, `state.db`, `tx_index.db`, `evidence.db` and
`snapshots/metadata.db`. **Corruption can hit several of them at once**, so after fixing
the first one, the node may crash again pointing at another. During this incident the
archive node had both `application.db` and `snapshots/metadata.db` corrupted.

The corruption shows up at the next startup. If your node was killed but came back up
cleanly and is syncing, it was not corrupted. A strong early-warning sign is the node
being killed instead of stopping: container exit code 137 (SIGKILL), OOM-kill events, or
`Killed` in system logs. Every such kill is a chance for corruption; during this incident,
corruption appeared within seconds of an OOM kill.

## Recovery options

Ordered from safest to most dangerous. Before touching anything, take a snapshot or copy
of the node's data volume so no recovery attempt can make things worse.

### 1. Restore from a backup (preferred)

If you have a backup or disk snapshot of the data directory, restore it and let the node
sync the missing blocks from the network. This is the safest option and usually the
fastest. Take the backup while `mezod` is stopped, or use point-in-time disk snapshots
from your cloud or storage layer.

### 2. Re-sync using state sync

If you have no backup and your node is not an archive node, wipe the data and re-sync
from a recent snapshot of the chain state. This takes hours instead of days and is
documented in the validator-kit README:
[State sync from snapshot](https://github.com/mezo-org/validator-kit#state-sync-from-snapshot).

Two warnings:

- If the node is a validator, save `data/priv_validator_state.json` before wiping and put
  it back before restarting. That file tracks what your validator has already signed;
  losing it risks signing two different blocks at the same height (double-signing), which
  endangers the safety of the whole network.
- **Do not use state sync on an archive node.** A state-synced node starts from a recent
  height and has no historical state; the archive history would be lost.

### 3. Repair the database (EXPERIMENTAL, last resort)

If the node holds data that exists nowhere else (as with our testnet archive node) and
there is no backup, the database can be repaired offline. **This is experimental and
dangerous.** Work on a copy, keep the snapshot from before you started, and expect to
verify the result carefully. Reach out to the Mezo team before attempting it.

What we learned doing it on a 90 GiB `application.db`:

- LevelDB's built-in recovery (`leveldb.RecoverFile` in goleveldb, the Go implementation
  `mezod` uses) rebuilds the `MANIFEST` and makes the database readable again, and the
  node accepts blocks correctly afterwards. For small databases this can be enough.
- On a large database it is a trap: recovery throws away the database's sorted structure,
  so every read afterwards scans tens of thousands of files. Our node processed one block
  per 9.7 seconds while the network produced one every 4.6 seconds, so it kept falling
  behind. In-place compaction to fix this measured out to about 4 days in one
  non-resumable operation.
- What worked was rebuilding the database offline: merge-sort the raw data files into a
  fresh, properly organized database, keeping only the newest version of each key. For
  1.27 billion entries this took about 3 hours plus a re-sync of the blocks written since
  the corruption.
- After the rebuild, verify the node against the network (it re-syncs and the app hashes
  must match) and, for an archive node, query historical state at old heights to confirm
  the history survived.

## Prevention

Node operators should do all of the following:

1. **Run v11.0.1 or later.** It fixes the vulnerability that caused the OOM kills. If you
   have not upgraded yet, apply step 2 first: the upgrade requires a restart, and
   restarting with the default stop timeouts is exactly what corrupted the mainnet nodes.
2. **Give `mezod` time to shut down.** Raise the stop timeout to 120 seconds so `mezod`
   can close its databases before it gets killed. [validator-kit] 12.1.0 ships this in
   all deployment variants:
    - Kubernetes: `terminationGracePeriodSeconds: 120` (chart default since 12.1.0),
    - Docker Compose: `stop_grace_period: 2m`,
    - systemd: `TimeoutStopSec=120`.

   A healthy shutdown takes well under a second (we measured about 20 ms), but a node
   under memory pressure or with a large, busy database can need much longer. The wide
   margin is cheap; a SIGKILL is not. Never stop `mezod` with `kill -9`.
3. **Back up the data volume regularly.** A daily disk snapshot turns this whole class of
   incident into a short restore. This matters most for archive nodes, whose data cannot
   be recreated from the network.
4. **Do not expose query ports publicly.** The query wave entered through the Cosmos SDK
   gRPC (9090), REST API (1317) and CometBFT RPC (26657) ports. A validator only needs
   P2P (26656) reachable from the outside.
5. **Alert on kills.** Treat container exit code 137 and OOM-kill events as incidents and
   watch the next startup closely.

## Impact

The testnet chain never halted; it kept producing blocks on four of five validators. The
testnet archive node was out of service for six days, and its recovery cost several days
of engineering work. On mainnet, the chain was unaffected; the corrupted validator nodes
were individually offline until re-synced. No test or real funds were at risk.

## Next steps

- (Done) Roll v11.0.1 and the 120-second stop timeout across the Mezo-managed fleet.
- (Done) Set up daily disk snapshots with 14-day retention for the nodes whose data
  cannot be recreated from the network (testnet archive node, mainnet RPC node).
- Notify mainnet operators who have not yet upgraded to v11.0.1 to raise their stop
  timeout before restarting (this report is part of that effort).
- Restrict the public exposure of query ports on the testnet validators.
- Add alerting for OOM kills and containers exiting with code 137, so every unclean kill
  is investigated immediately.
- Publish the security advisory for the query vulnerability once operators have had time
  to upgrade.

[v11.0.1]: https://github.com/mezo-org/mezod/releases/tag/v11.0.1
[validator-kit]: https://github.com/mezo-org/validator-kit
