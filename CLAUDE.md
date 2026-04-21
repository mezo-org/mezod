# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

Mezo (`mezod`) is a Bitcoin-first Cosmos SDK-based L1 with EVM compatibility, running on CometBFT consensus.
The client codebase is a heavy fork of the LGPL version of Evmos. The chain uses BTC as gas (denom `abtc`,
"atto BTC"), MEZO as a governance/staking token (denom `amezo`), bridges with tBTC, and follows a dual
staking (PoA validators + delegated staking) model.

## Commands

### Build and install

- `make build` — builds `build/mezod`, `build/bridge-worker`, `build/metrics-scraper`. Cleans `build/` first.
- `make install` — installs binaries to `$GOPATH/bin`.
- `make bindings` — fetches NPM contract packages (tmp/contracts) and runs `go generate ./...` to regenerate
  Ethereum contract bindings. Required before `make build` after checkout because bindings are gitignored.
- `make build-docker-linux` — build linux/amd64 Docker image (needed for release smoke tests).

### Tests

- `make test-unit` — all Go unit tests with `-race`, 15m timeout, excludes `tests/e2e`.
- `make test-race` — same but includes simulation packages.
- `make test-unit-cover` — with coverage profile.
- `make benchmark` — runs `go test -bench=.` on non-simulation packages.
- Single test: `go test -run TestName ./path/to/package` or `go test -run TestName -race ./x/bridge/keeper/...`.
- System/E2E tests (Hardhat, TypeScript) live in `tests/system/`. Run via
  `./tests/system/system-tests.sh [SuiteName...]`. They require a running localnode (`make localnode-bin-start`)
  or testnet access (`NETWORK=testnet PRIVATE_KEYS=...`).

### Lint and format

- `make lint` / `make lint-fix` — golangci-lint (config in `.golangci.yml`; enables gofumpt, revive, gosec,
  staticcheck, etc.).
- `make format-markdown` / `make format-markdown-fix` — markdownlint.
- `make proto-lint` / `make proto-format` / `make proto-gen` — dockerized buf-based Protobuf tooling.

### Local networks

Two distinct local setups (don't confuse them):

- **Localnode** (single node, simplest): `make localnode-bin-start`. Creates/clears `.localnode/`, generates
  three dev keys (`dev0..dev2`), starts one validator. `dev0` is the PoA owner. Bridge is effectively disabled
  (bogus `source_btc_token`).
- **Localnet** (four nodes, bridge-capable): `make localnet-bin-init` → `make localnet-bin-sidecars-start` →
  `make localnet-bin-start` (called once per node). Needs `ETH_SIDECAR_RPC_PROVIDER` in `.env` for the Ethereum
  sidecar. Connect sidecar reads node0's gRPC, so node0 must be up. Default chain ID is `mezo_31611-10`, source
  BTC token is tBTC on Sepolia.

### Other

- `make vulncheck` — govulncheck scan.
- `make generate` — `go generate ./...` with a `GOFLAGS=-mod=mod` dance; used by `make bindings`.
- Pre-commit: `pip install pre-commit && pre-commit install` to enable the hooks listed in
  `.pre-commit-config.yaml`.
- Go toolchain: **go 1.22.11** (pinned in `go.mod` and Dockerfile).

## Architecture

### Binary entry points (`cmd/`)

- `cmd/mezod` — the chain client. `main.go` wires up `cmdcfg` (bech32/bip44 setup) then delegates to
  `svrcmd.Execute` with `NewRootCmd()`. `root.go` also registers custom commands: `testnet init-files`, genesis
  helpers, PoA commands, `toml` editor used by `entrypoint.sh`.
- `cmd/bridge-worker` — standalone BTC withdrawal worker process (see "Bridging" below).
- `cmd/metrics-scraper` — pulls metrics from bridge contracts for Prometheus.

### App wiring (`app/`)

`app.Mezo` in `app/app.go` is the `baseapp`-derived ABCI application. The `ModuleBasics` list and
`keys`/`tkeys` declare the Cosmos modules that make up the chain. The app composes:

- Core Cosmos modules: `auth`, `bank`, `consensus`, `crisis`, `authz`, `params`, `upgrade`.
- Mezo modules (`x/`): `poa`, `evm`, `feemarket`, `bridge`.
- Skip `connect/v2` modules: `marketmap`, `oracle` — off-chain price oracle wiring.

Key non-module components:

- `app/abci` — custom ABCI++ `PreBlockHandler`, `ProposalHandler`, and vote-extension handlers that coordinate
  bridge & oracle processing.
- `app/ante` — ante handlers; `ante/evm` has EVM-specific checks.
- `app/upgrades/vX_Y` — one package per historical upgrade. Two primitives:
    - `Fork` — code block keyed to a block height for upgrades without chain halt; registered in `app.Forks`.
    - `Upgrade` — store migration handler for upgrades with chain halt; registered in `app.Upgrades`.
  See `docs/upgrades.md` for the three upgrade procedures and the historical table.
- `app/oracle.go` — Connect (Skip) oracle client/metrics setup.

### Cosmos modules (`x/`)

- `x/bridge` — Bitcoin bridging. Consumes `AssetsLocked` events from Ethereum via the sidecar, mints/burns on
  Mezo, emits pseudo-transactions (at most one per block, always at index 0) so block explorers can show bridge
  activity. Submodules: `keeper` (bank/mint, outflow limits, pause, triparty), `abci` (vote extensions so
  validators agree on finalized Ethereum events), `client/cli`.
- `x/evm` — forked ethermint EVM. `keeper/state_transition.go`, `statedb/`, `migrations/`. Handles EVM tx
  processing, gas accounting, access list behavior. Registers precompiles.
- `x/feemarket` — EIP-1559-style base fee. `MainnetMinGasPrices`/`MainnetMinGasMultiplier` are wired globally
  in `app.init()`.
- `x/poa` — Proof-of-Authority validator set. Owner-gated applications, privilege flags, CometBFT historical
  info compatibility (`connect_compat.go`).

### Precompiles (`precompile/`)

EVM-native precompiled contracts sit in the `0x7b7c...` address space. Each precompile package contains Go
logic, an `abi.json`, and a `byte_code.go`; the ABI/bytecode are generated by compiling the matching Solidity
interface in `precompile/hardhat/contracts` and copying `deployedBytecode`. Precompiles are versioned via
`version_map.go`; the active version is selected by chain height. Current precompiles:

| Address | Package |
|---|---|
| `0x7b7c...0000` | `btctoken` (ERC-20-like wrapper over native BTC gas token) |
| `0x7b7c...0001` | `mezotoken` |
| `0x7b7c...0011` | `validatorpool` (PoA management) |
| `0x7b7c...0012` | `assetsbridge` (bridge observability + ERC20 mapping) |
| `0x7b7c...0013` | `maintenance` |
| `0x7b7c...0014` | `upgrade` (submits/cancels `x/upgrade` plans) |
| `0x7b7c...0015` | `priceoracle` |
| (testbed-only) | `testbed`, enabled only via `--enable-testbed-precompile` |

Adding a precompile requires: Solidity interface + caller, ABI/bytecode generation via Hardhat, Go
implementation implementing the `Method` interface, registration in `app.go`, and optional Hardhat tasks under
`precompile/hardhat/tasks`. See `docs/precompile.md`.

### Ethereum integration (`ethereum/`, `bridge-worker/`)

- `ethereum/sidecar` — long-running companion process that watches Ethereum for `AssetsLocked` events on the
  `MezoBridge` contract, waits for finality (~13–14 min), and serves them via gRPC to `mezod`. Also batches
  attestations for cross-chain messages.
- `ethereum/bindings` — auto-generated Go bindings for `portal`, `tbtc`, shared `common`. Regenerated via
  `make bindings`.
- `bridge-worker/` — separate binary that observes the chain and completes BTC withdrawal PSBTs (sign,
  broadcast). Uses `bridge-worker/bitcoin` for BTC RPC, `bridge-worker/ethereum` for reads.

### JSON-RPC (`rpc/`, `server/`)

Ethereum-compatible JSON-RPC lives under `rpc/`. Namespaces: `eth`, `net`, `web3`, `debug`, `txpool`,
`personal` (deprecated), `mezo` (custom — e.g. `mezo_estimateCost`). The server is started from
`server/start.go`; `server/json_rpc.go` mounts the HTTP/WS handlers; `indexer/kv_indexer.go` powers the custom
tx indexer that enables pseudo-transaction observability (`enable-indexer = true` in `app.toml`). Full method
reference in `docs/evm-compatibility.md`.

### EVM fork semantics

Mezo advertises post-London forks but deviates in several places (PREVRANDAO returns 0, EIP-4844 blob txs are
rejected, `BLOBHASH`/`BLOBBASEFEE` return 0, no EIP-4895 withdrawals, EIP-4788 beacon root is out of scope).
Full list: `docs/evm-compatibility.md`. System tests under `tests/system/` include opcode-level checks
(`Push0Check`, `McopyCheck`, `TransientStorageCheck`, `Selfdestruct6780Check`, `InitcodeLimitCheck`,
`RandaoCheck`).

### Solidity (`solidity/`, `tests/system/`, `precompile/hardhat/`)

Three separate Hardhat projects — don't cross-import:

- `solidity/` — the chain's native Mezo ERC-20 asset contracts (mcbBTC, mUSDC, mDAI, MEZO, …). Deployed
  addresses in `solidity/README.md`.
- `precompile/hardhat/` — contracts/callers used only to produce precompile ABIs + bytecode. Also hosts
  Hardhat *tasks* for interacting with precompiles on live networks (`npx hardhat <precompile>:<method>`).
  Account keys configured via `npx hardhat vars set MEZO_ACCOUNTS`.
- `tests/system/` — end-to-end test suites that run against localnode or testnet.

## Conventions

- Commits **must** be GPG-signed (CI enforces; signing in the sandbox won't work — sign outside).
- Each PR should update `docs/` or `x/<module>/spec/` when behavior changes. Follow the
  `.github/pull_request_template.md` sections (Introduction / Changes / Testing).
- Module dependencies are blocked by `gomodguard` in `.golangci.yml` (etcd < 3.4.10, dgrijalva/jwt-go ≥
  4.0.0-preview1).
- Commit message style: Chris Beams guidelines; avoid `-m`, write body text, don't reference ticket numbers
  (summarize context instead).
- When touching consensus logic, an upgrade is likely needed — read `docs/upgrades.md` before starting.
