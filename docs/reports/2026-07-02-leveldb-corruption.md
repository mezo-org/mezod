# Incident report: 2026-07-02 LevelDB corruption

In the early morning UTC of July 2, 2026, the archive node of the Mezo testnet started
crash-looping because its main database was corrupted. In the following days, mainnet
validator operators reported the same corruption on their nodes. Neither chain halted and
no funds were at risk, but the affected nodes needed days to get back to normal.

The root cause is general and affects every `mezod` node: **killing a running `mezod`
process without letting it shut down cleanly can corrupt its databases.** This report
explains how that happens, how to recognize it, how to recover, and what to do so it does
not happen to your node.

## Summary

`mezod` stores all chain data in [LevelDB](https://github.com/google/leveldb) databases,
located in the `data/` directory inside the `mezod` home directory (the path passed via
`--home`, `~/.mezod` by default; when running in Docker or Kubernetes, it lives on the
volume mounted into the container). This report calls it the **data directory**. Each
database has a `MANIFEST` file that acts as its index. If the process is killed in the
middle of writing (SIGKILL, out-of-memory kill, power loss), the `MANIFEST` can be left
half-written or otherwise unreadable, and LevelDB then refuses to open the database. The
node crash-loops on startup from that point on.

Two separate chains of events led to such kills during this incident:

- **Testnet:** on July 2, `mezod` on the five testnet validators operated by the Mezo
  team repeatedly ran out of memory and was killed by the operating system, 33
  out-of-memory (OOM) kills within 46 minutes. One of those validators also serves as
  the testnet archive node, and its database did not survive its kill.
- **Mainnet:** operators restarted their nodes to apply the [v11.0.1] release. Every
  previous major upgrade was a coordinated chain-halt upgrade where `mezod` stops on its
  own and closes its databases cleanly. This was the first large wave of *hot restarts*
  of live, actively writing nodes, and the default stop timeouts (10 seconds for Docker
  Compose, 30 seconds for Kubernetes) turned out to be too short: when the timeout
  expires, the node is killed mid-write. Several operators hit the corruption. Their
  nodes were regular (non-archive) nodes, so they could be restored by re-syncing
  (recovery option 2 below).

Neither chain halted. On the testnet (eight validators at the time of writing: five
operated by the Mezo team, three by external operators), only the archive node was down
for an extended time, well within the fault tolerance of consensus. The corrupted archive
node held the only full copy of testnet history, so instead of re-syncing it (which would
lose that history) its 90 GiB database was repaired with an experimental offline rebuild
(recovery option 3 below). It was fully caught up and verified on July 8.

## Symptoms

A node with a corrupted LevelDB database crash-loops on startup. The log ends with an
error like:

```text
panic: ... leveldb: manifest corrupted (field 'comparer'): missing [file=MANIFEST-3772245]
```

The file name and the field vary. Any of the six databases in the data directory can be
affected: `application.db`, `blockstore.db`, `state.db`, `tx_index.db`, `evidence.db` and
`snapshots/metadata.db`. **Corruption can hit several of them at once**, so after fixing
the first one, the node may crash again pointing at another. During this incident the
archive node had both `application.db` and `snapshots/metadata.db` corrupted.

The corruption shows up at the next startup. If your node was killed but came back up
cleanly and is syncing, it was not corrupted. A strong early-warning sign is the node
being killed instead of stopping: container exit code 137 (SIGKILL), OOM-kill events, or
`Killed` in system logs. Every such kill is a chance for corruption; during this incident,
corruption appeared within seconds of an OOM kill.

## Recovery options

Ordered from safest to most dangerous. Paths below are relative to the `mezod` home
directory.

Before attempting any option:

1. Stop `mezod` and keep it stopped for the whole procedure.
2. Take a copy of the data directory (or a snapshot of the disk it lives on), so no
   recovery attempt can make things worse.
3. If the node is a validator, copy `data/priv_validator_state.json` to a safe place
   outside the data directory. This small JSON file records the last block your validator
   signed. It is not a LevelDB file, so it survives the corruption intact. You will put
   it back at the end of options 1 and 2. Running with an older or reset copy of this
   file risks signing two different blocks at the same height (double-signing), which
   endangers the safety of the whole network.

### 1. Restore from a backup (preferred)

If you have a backup or disk snapshot of the data directory, restore it and let the node
sync the missing blocks from the network. This is the safest option and usually the
fastest.

1. Restore the data directory from your most recent backup. With cloud disk snapshots,
   create a new disk from the snapshot and attach it in place of the corrupted one.
2. Overwrite the restored `data/priv_validator_state.json` with the copy you saved before
   starting. The backup contains an older version of this file; the saved copy is the
   only record of what your validator signed after the backup was taken.
3. Start `mezod`. It block-syncs the gap between the backup and the chain tip; watch the
   logs for steadily increasing block heights.

### 2. Re-sync using state sync

If you have no backup and your node is **not** an archive node, wipe the local state and
re-sync from a recent snapshot of the chain state. This takes hours instead of days. The
[validator-kit README](https://github.com/mezo-org/validator-kit#state-sync-from-snapshot)
documents the procedure in full; the steps are:

1. Wipe the local state:

   ```bash
   mezod tendermint unsafe-reset-all --home <mezod-home>
   ```

2. Put your saved `priv_validator_state.json` back into `data/`. The reset in step 1
   replaces it with a blank one, which loses the record of what your validator already
   signed.
3. Pick a trusted block from a node you trust: `http://<node-address>:26657/block`
   returns the latest block with its height and hash. Snapshots are taken at fixed
   intervals (every 5000 blocks on the nodes operated by the Mezo team), so choose a
   height just above a snapshot height and read its hash from
   `http://<node-address>:26657/block?height=<height>`.
4. In `config/config.toml`, section `[statesync]`, set:
    - `enable = true`,
    - `rpc_servers` to a comma-separated list of at least two trusted RPC servers,
    - `trust_height` and `trust_hash` to the block chosen in step 3.
5. Start `mezod`. It downloads the state snapshot from the network and then syncs the
   remaining blocks. Once the node is synced, you can set `enable = false` again.

Where to get snapshots from:

- **Testnet:** the Mezo team provides state sync snapshots on its testnet nodes; see the
  [validator-kit README](https://github.com/mezo-org/validator-kit#state-sync-from-snapshot)
  for a ready-made configuration.
- **Mainnet:** the Mezo team does **not** provide mainnet snapshots; you need to reach
  out to other node operators. Known third-party snapshot services for Mezo mainnet are
  [Nodes Hub](https://services.nodeshub.online/mainnet/mezo/snapshot) and
  [Imperator](https://www.imperator.co/services/chain-services/mainnets/mezo). These
  services are not vetted or endorsed by the Mezo development team; use them at your own
  risk.

**Do not use state sync on an archive node.** A state-synced node starts from a recent
height and has no historical state; the archive history would be lost.

### 3. Repair the database (EXPERIMENTAL, last resort)

If the node holds data that exists nowhere else (as with our testnet archive node) and
there is no backup, the corrupted database can be repaired offline. **This is
experimental and dangerous: it operates directly on the database files and can destroy
data that a more careful attempt would have saved.** Only proceed on a copy, keep the
snapshot from before you started, and reach out to the Mezo team before attempting it.

Step by step:

1. Identify the corrupted databases. The startup panic names the first one (the
   `MANIFEST-XXXXXXX` file lives inside the affected database directory). Check all six
   databases listed in the [Symptoms](#symptoms) section, because several can be
   corrupted at once; repairing one and starting the node reveals the next.
2. Rebuild the `MANIFEST` with goleveldb's built-in recovery. goleveldb is the Go
   implementation of LevelDB that `mezod` uses; its `leveldb.RecoverFile` function scans
   the data files and writes a fresh `MANIFEST`. Use the exact goleveldb version pinned
   in `mezod`'s `go.mod` (see the `replace` directive for `github.com/syndtr/goleveldb`).
   A minimal program:

   ```go
   package main

   import (
       "fmt"
       "os"

       "github.com/syndtr/goleveldb/leveldb"
   )

   func main() {
       db, err := leveldb.RecoverFile(os.Args[1], nil)
       if err != nil {
           fmt.Fprintln(os.Stderr, "recover failed:", err)
           os.Exit(1)
       }
       defer db.Close()

       it := db.NewIterator(nil, nil)
       defer it.Release()
       n := 0
       for it.Next() && n < 10 {
           n++
       }
       if err := it.Error(); err != nil {
           fmt.Fprintln(os.Stderr, "iterator error:", err)
           os.Exit(1)
       }
       fmt.Println("recovery OK, sample keys readable:", n)
   }
   ```

   Run it against each corrupted database directory, for example
   `go run . <mezod-home>/data/application.db`.
3. Start the node and verify it applies new blocks. The network cross-checks every block
   (app hashes must match), so if the node syncs, the recovered data is consistent with
   the chain.
4. For small databases this is the end. For a large database (tens of GiB), expect the
   node to be far too slow: `RecoverFile` throws away the database's sorted structure,
   so every read afterwards scans tens of thousands of data files. Our archive node
   processed one block per 9.7 seconds while the network produced one every 4.6 seconds,
   so it kept falling behind, and letting LevelDB compact itself back into shape in
   place measured out to about 4 days in one non-resumable operation.
5. The way out of step 4 is to rebuild the database offline into a fresh, properly
   organized copy. This requires custom Go tooling built against the same goleveldb
   version; the Mezo team has a tested tool and can assist. The procedure, which needs
   free disk space of roughly twice the database size:
    1. Merge-sort the table files into intermediate databases ("runs"). Open the `.ldb`
       files of the recovered database directly with goleveldb's `table.NewReader`, in
       groups of about 250 files, and write each group through a merged iterator into
       its own intermediate database. Order entries by LevelDB's internal key: user key
       ascending, then sequence number descending, so the newest version of each key
       comes first. For our 60,272 table files this produced 236 runs in about one hour.
    2. Reopen each run once with a small write buffer and close it. This flushes the
       run's journal into table files; without it, the final merge holds every journal
       in memory and runs out of it.
    3. Merge all runs through a single merged iterator into the fresh final database.
       For each user key, keep only the first (newest) version and drop the key entirely
       if that version is a deletion marker. For our database this wrote 1.27 billion
       entries in about two hours.
    4. Swap the fresh database in place of the corrupted one, keeping the old one until
       the node is verified.
6. Start the node and verify as in step 3, then let it re-sync the blocks produced since
   the corruption. On an archive node, additionally verify that history survived: query
   historical state at old heights (for example `eth_call` or `eth_getBalance` at old
   block numbers) and confirm you get values instead of errors.

## Prevention

Node operators should do all of the following:

1. **Give `mezod` time to shut down.** Raise the stop timeout to 120 seconds so `mezod`
   can close its databases before it gets killed. Do this before your next restart (for
   example, before an upgrade): restarting with the default timeouts is exactly what
   corrupted the mainnet nodes. Depending on how you run `mezod`:
    - Kubernetes: set `terminationGracePeriodSeconds: 120` on the pod,
    - Docker Compose: set `stop_grace_period: 2m` on the service,
    - systemd: set `TimeoutStopSec=120` in the unit.

   If you deploy with [validator-kit], version 12.1.0 ships these settings in all
   deployment variants.

   A healthy shutdown takes well under a second (we measured about 20 ms), but a node
   under memory pressure or with a large, busy database can need much longer. The wide
   margin is cheap; a SIGKILL is not. Never stop `mezod` with `kill -9`.
2. **Back up the data directory regularly.** A daily backup turns this whole class of
   incident into a short restore. Take backups while `mezod` is stopped, or use
   point-in-time snapshots of the disk the data directory lives on. This matters most
   for archive nodes, whose data cannot be recreated from the network.
3. **Alert on kills.** Treat container exit code 137 and OOM-kill events as incidents and
   watch the next startup closely.

## Impact

Neither chain halted. The testnet archive node was out of service for six days, and its
recovery cost several days of engineering work. On mainnet, the corrupted validator nodes
were individually offline until re-synced. No test or real funds were at risk.

[v11.0.1]: https://github.com/mezo-org/mezod/releases/tag/v11.0.1
[validator-kit]: https://github.com/mezo-org/validator-kit
