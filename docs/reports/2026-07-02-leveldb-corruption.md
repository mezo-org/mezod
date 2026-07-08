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
  of live, actively writing nodes, and the default stop timeouts (such as those of
  [validator-kit] deployments before version 12.1.0) turned out to be too short: when
  the timeout expires, the node is killed mid-write. Several operators hit the
  corruption. Their
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

If you have no backup of your own, you can restore from someone else's: third-party
snapshot services publish downloadable archives of the data directories of their own
nodes. Restoring works the same way, except you replace your `data/` directory with the
content of the archive, and you are trusting a stranger's copy of the chain data. Known
services for Mezo mainnet are
[Nodes Hub](https://services.nodeshub.online/mainnet/mezo/snapshot) and
[Imperator](https://www.imperator.co/services/chain-services/mainnets/mezo).
**These services are not vetted or endorsed by the Mezo development team; use them at
your own risk.** As with your own backup, put your saved `priv_validator_state.json`
back after unpacking the archive.

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

State sync downloads the state snapshot automatically from other nodes over the
peer-to-peer network and verifies it against the trusted block, so it needs peers that
serve state sync snapshots and expose their CometBFT RPC:

- **Testnet:** the Mezo team serves state sync snapshots from its testnet nodes; see the
  [validator-kit README](https://github.com/mezo-org/validator-kit#state-sync-from-snapshot)
  for a ready-made configuration.
- **Mainnet:** the Mezo team does **not** run public state sync sources; reach out to
  other node operators for nodes that serve state sync snapshots. If you cannot find
  any, restoring a third-party archive of the data directory (see option 1) is the
  alternative.

**Do not use state sync on an archive node.** A state-synced node starts from a recent
height and has no historical state; the archive history would be lost.

### 3. Repair the database (EXPERIMENTAL, last resort)

If the node holds data that exists nowhere else (as with our testnet archive node) and
there is no backup, the corrupted database can be repaired offline. **This is
experimental and dangerous: it operates directly on the database files and can destroy
data that a more careful attempt would have saved.** Only proceed on a copy, and keep the
snapshot from before you started.

All commands below use the rebuild tool from the
[appendix](#appendix-offline-rebuild-tool); build it first. Run everything with the node
stopped, on a machine (or helper container) that has the data directory mounted.

Step by step:

1. Identify the corrupted databases. The startup panic names the first one (the
   `MANIFEST-XXXXXXX` file lives inside the affected database directory). Check all six
   databases listed in the [Symptoms](#symptoms) section, because several can be
   corrupted at once; repairing one and starting the node reveals the next.
2. Rebuild the `MANIFEST` of each corrupted database:

   ```bash
   ldb-recover recover <mezod-home>/data/application.db
   ```

   This runs goleveldb's built-in `leveldb.RecoverFile`, which scans the data files and
   writes a fresh `MANIFEST`, then checks that the database opens and is readable.
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
   organized copy. Stop the node again and make sure you have free disk space of
   roughly twice the database size, then:
    1. Merge-sort the table files into intermediate sorted databases ("runs"):

       ```bash
       ldb-recover stage1 <mezod-home>/data/application.db <scratch-dir>/runs
       ```

       This reads the `.ldb` files directly, in groups of 256, and writes each group
       through a merged iterator into its own run. For our 60,272 table files it
       produced 236 runs in about one hour. Interrupted? Just rerun; finished runs are
       skipped.
    2. Flush the runs:

       ```bash
       ldb-recover flushruns <scratch-dir>/runs
       ```

       This reopens each run once so its journal is flushed into table files. Without
       it, the final merge holds every journal in memory and runs out of it.
    3. Merge all runs into the fresh final database:

       ```bash
       ldb-recover stage2 <scratch-dir>/runs <mezod-home>/data/application.db.new
       ```

       For each key, this keeps only the newest version and drops deleted keys. For our
       database it wrote 1.27 billion entries in about two hours.
    4. Swap the fresh database in place of the corrupted one, keeping the old one until
       the node is verified:

       ```bash
       cd <mezod-home>/data
       mv application.db application.db.old
       mv application.db.new application.db
       ```

6. Start the node and verify as in step 3, then let it re-sync the blocks produced since
   the corruption. On an archive node, additionally verify that history survived: query
   historical state at old heights (for example `eth_call` or `eth_getBalance` at old
   block numbers) and confirm you get values instead of errors. Once verified, delete
   `application.db.old`.

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

## Appendix: offline rebuild tool

<details>
<summary>The complete rebuild tool used in recovery option 3 (click to expand)</summary>

The single-file Go program below implements recovery option 3. It is the tool that was
used to repair the testnet archive node, trimmed to the four needed modes:

- `recover <db>`: rebuild the `MANIFEST` of a corrupted database in place and check that
  it opens and is readable.
- `stage1 <db> <runs-dir>`: merge-sort the database's table files into intermediate
  sorted databases ("runs").
- `flushruns <runs-dir>`: flush the runs' journals into table files.
- `stage2 <runs-dir> <new-db>`: merge all runs into a fresh database, keeping only the
  newest version of each key and dropping deleted keys.

Build it with the **exact goleveldb version `mezod` uses**, taken from the `replace`
directive for `github.com/syndtr/goleveldb` in `mezod`'s `go.mod`:

```bash
mkdir ldb-recover && cd ldb-recover
go mod init ldb-recover
go get github.com/syndtr/goleveldb@<version-from-mezod-go.mod>
# save the program below as main.go, then:
go build -o ldb-recover .
```

The tool runs in bounded memory (we ran it with a 12 GiB limit next to a 90 GiB
database). Entries in stage 1 and 2 are ordered by LevelDB's *internal* key: user key
ascending, then sequence number descending, so the newest version of every key comes
first; that ordering is what lets stage 2 deduplicate versions and drop deletion markers
in a single pass.

```go
package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/storage"
	"github.com/syndtr/goleveldb/leveldb/table"
	"github.com/syndtr/goleveldb/leveldb/util"
)

func main() {
	switch {
	case len(os.Args) == 3 && os.Args[1] == "recover":
		recoverMain(os.Args[2])
	case len(os.Args) == 4 && os.Args[1] == "stage1":
		stage1Main(os.Args[2], os.Args[3])
	case len(os.Args) == 3 && os.Args[1] == "flushruns":
		flushrunsMain(os.Args[2])
	case len(os.Args) == 4 && os.Args[1] == "stage2":
		stage2Main(os.Args[2], os.Args[3])
	default:
		fmt.Fprintln(os.Stderr, "usage: ldb-recover recover <db> | stage1 <db> <runs-dir> |",
			"flushruns <runs-dir> | stage2 <runs-dir> <new-db>")
		os.Exit(1)
	}
}

func fatalIf(err error, msg string) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", msg, err)
		os.Exit(1)
	}
}

// recoverMain rebuilds the MANIFEST of a corrupted database in place and
// checks that the result opens and is readable.
func recoverMain(path string) {
	fmt.Println("recovering", path)
	db, err := leveldb.RecoverFile(path, nil)
	fatalIf(err, "recover failed")

	it := db.NewIterator(nil, nil)
	n := 0
	for it.Next() && n < 10 {
		n++
	}
	it.Release()
	fatalIf(it.Error(), "iterator error after recovery")
	fatalIf(db.Close(), "close failed")
	fmt.Println("recovery OK, sample keys readable:", n)
}

// icmp orders goleveldb internal keys: user key ascending, then
// (sequence<<8|kind) descending, so the newest version of a key comes first.
type icmp struct{}

func (icmp) Name() string { return "rebuild.icmp" }

func (icmp) Compare(a, b []byte) int {
	ua, ub := a[:len(a)-8], b[:len(b)-8]
	if r := bytes.Compare(ua, ub); r != 0 {
		return r
	}
	na := binary.LittleEndian.Uint64(a[len(a)-8:])
	nb := binary.LittleEndian.Uint64(b[len(b)-8:])
	switch {
	case na > nb:
		return -1
	case na < nb:
		return 1
	}
	return 0
}

func (icmp) Separator(dst, a, b []byte) []byte { return nil }
func (icmp) Successor(dst, b []byte) []byte    { return nil }

func listTables(src string) []string {
	entries, err := os.ReadDir(src)
	fatalIf(err, "readdir failed")
	var files []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".ldb") || strings.HasSuffix(e.Name(), ".sst") {
			files = append(files, filepath.Join(src, e.Name()))
		}
	}
	sort.Strings(files)
	return files
}

func fileNum(path string) int64 {
	base := strings.TrimSuffix(strings.TrimSuffix(filepath.Base(path), ".ldb"), ".sst")
	n, _ := strconv.ParseInt(base, 10, 64)
	return n
}

const groupSize = 256

// stage1Main reads the source database's table files directly and
// merge-sorts them, in groups, into intermediate run databases. Finished
// runs are skipped on rerun, so an interrupted stage 1 can be resumed.
func stage1Main(src, runsDir string) {
	files := listTables(src)
	fmt.Printf("stage1: %d tables in %s\n", len(files), src)
	fatalIf(os.MkdirAll(runsDir, 0o755), "mkdir runs")

	bpool := util.NewBufferPool(opt.DefaultBlockSize + 5)
	start := time.Now()
	totalEntries := 0

	for gi := 0; gi*groupSize < len(files); gi++ {
		runPath := filepath.Join(runsDir, fmt.Sprintf("run-%05d", gi))
		if _, err := os.Stat(filepath.Join(runPath, "CURRENT")); err == nil {
			fmt.Printf("stage1: run %d exists, skipping\n", gi)
			continue
		}
		group := files[gi*groupSize : min((gi+1)*groupSize, len(files))]

		var (
			iters   []iterator.Iterator
			readers []*table.Reader
			fhs     []*os.File
		)
		for _, f := range group {
			fh, err := os.Open(f)
			fatalIf(err, "open table "+f)
			st, err := fh.Stat()
			fatalIf(err, "stat table "+f)
			tr, err := table.NewReader(
				fh, st.Size(),
				storage.FileDesc{Type: storage.TypeTable, Num: fileNum(f)},
				nil, bpool, &opt.Options{},
			)
			fatalIf(err, "table reader "+f)
			iters = append(iters, tr.NewIterator(nil, nil))
			readers = append(readers, tr)
			fhs = append(fhs, fh)
		}

		mi := iterator.NewMergedIterator(iters, icmp{}, true)
		rdb, err := leveldb.OpenFile(runPath, &opt.Options{
			Comparer:            icmp{},
			WriteBuffer:         128 * opt.MiB,
			CompactionTableSize: 32 * opt.MiB,
		})
		fatalIf(err, "open run db")

		batch := new(leveldb.Batch)
		entries := 0
		for mi.Next() {
			batch.Put(mi.Key(), mi.Value())
			entries++
			if batch.Len() >= 10000 {
				fatalIf(rdb.Write(batch, nil), "write batch")
				batch.Reset()
			}
		}
		fatalIf(mi.Error(), "merged iterator")
		if batch.Len() > 0 {
			fatalIf(rdb.Write(batch, nil), "write final batch")
		}
		mi.Release()
		for i := range readers {
			readers[i].Release()
			fhs[i].Close()
		}
		fatalIf(rdb.Close(), "close run db")
		totalEntries += entries
		fmt.Printf("stage1: run %d/%d done, %d entries, elapsed %s\n",
			gi+1, (len(files)+groupSize-1)/groupSize, entries, time.Since(start).Round(time.Second))
	}
	fmt.Printf("stage1 OK: %d entries in %s\n", totalEntries, time.Since(start).Round(time.Second))
}

// flushrunsMain reopens each run read-write with a tiny write buffer so
// journal replay flushes pending writes into table files. Without this,
// stage2's read-only opens must hold every run's journal tail in memory.
func flushrunsMain(runsDir string) {
	entries, err := os.ReadDir(runsDir)
	fatalIf(err, "readdir runs")
	n := 0
	for _, e := range entries {
		if !e.IsDir() || !strings.HasPrefix(e.Name(), "run-") {
			continue
		}
		rp := filepath.Join(runsDir, e.Name())
		rdb, err := leveldb.OpenFile(rp, &opt.Options{
			Comparer:    icmp{},
			WriteBuffer: 64 * opt.KiB,
		})
		fatalIf(err, "flush open "+rp)
		fatalIf(rdb.Close(), "flush close "+rp)
		n++
	}
	fmt.Printf("flushruns OK: %d runs flushed\n", n)
}

// stage2Main merges all runs into a fresh database. Thanks to the icmp
// ordering, the first occurrence of a user key is its newest version:
// older versions are skipped and keys whose newest version is a deletion
// marker are dropped entirely.
func stage2Main(runsDir, dst string) {
	entries, err := os.ReadDir(runsDir)
	fatalIf(err, "readdir runs")
	var runPaths []string
	for _, e := range entries {
		if e.IsDir() && strings.HasPrefix(e.Name(), "run-") {
			runPaths = append(runPaths, filepath.Join(runsDir, e.Name()))
		}
	}
	sort.Strings(runPaths)
	fmt.Printf("stage2: merging %d runs into %s\n", len(runPaths), dst)

	var (
		iters []iterator.Iterator
		rdbs  []*leveldb.DB
	)
	for _, rp := range runPaths {
		rdb, err := leveldb.OpenFile(rp, &opt.Options{
			Comparer:               icmp{},
			ReadOnly:               true,
			ErrorIfMissing:         true,
			OpenFilesCacheCapacity: 64,
			BlockCacheCapacity:     4 * opt.MiB,
		})
		fatalIf(err, "open run "+rp)
		iters = append(iters, rdb.NewIterator(nil, nil))
		rdbs = append(rdbs, rdb)
	}
	mi := iterator.NewMergedIterator(iters, icmp{}, true)

	out, err := leveldb.OpenFile(dst, &opt.Options{
		WriteBuffer:         128 * opt.MiB,
		CompactionTableSize: 8 * opt.MiB,
	})
	fatalIf(err, "open dst db")

	start := time.Now()
	batch := new(leveldb.Batch)
	var lastUkey []byte
	have := false
	kept, skippedOld, skippedDel := 0, 0, 0

	for mi.Next() {
		k := mi.Key()
		uk := k[:len(k)-8]
		num := binary.LittleEndian.Uint64(k[len(k)-8:])
		kind := num & 0xff

		if have && bytes.Equal(uk, lastUkey) {
			skippedOld++
			continue
		}
		lastUkey = append(lastUkey[:0], uk...)
		have = true

		if kind == 0 { // newest version is a deletion marker
			skippedDel++
			continue
		}
		batch.Put(uk, mi.Value())
		kept++
		if batch.Len() >= 10000 {
			fatalIf(out.Write(batch, nil), "write batch")
			batch.Reset()
		}
		if kept%5000000 == 0 {
			fmt.Printf("stage2: kept %dM entries, elapsed %s\n",
				kept/1000000, time.Since(start).Round(time.Second))
		}
	}
	fatalIf(mi.Error(), "merged iterator")
	if batch.Len() > 0 {
		fatalIf(out.Write(batch, nil), "write final batch")
	}
	mi.Release()
	for _, rdb := range rdbs {
		rdb.Close()
	}
	fatalIf(out.Close(), "close dst db")
	fmt.Printf("stage2 OK: kept=%d oldVersions=%d tombstones=%d in %s\n",
		kept, skippedOld, skippedDel, time.Since(start).Round(time.Second))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
```

</details>

[v11.0.1]: https://github.com/mezo-org/mezod/releases/tag/v11.0.1
[validator-kit]: https://github.com/mezo-org/validator-kit
