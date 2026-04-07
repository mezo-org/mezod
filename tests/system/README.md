# System Tests

End-to-end tests that run against a live Mezo chain (localnode or testnet).

## Prerequisites

- Node.js and npm
- A running Mezo node reachable over JSON-RPC

## Running tests

All commands are run from the `tests/system/` directory.

### Localnode (default)

Start the localnode first:

```bash
make localnode-bin-start   # from the repo root
```

Then run the tests:

```bash
./system-tests.sh                              # all suites
./system-tests.sh TripartyBridge               # single suite
./system-tests.sh AssetsBridge TripartyBridge   # multiple suites
```

The localnode seed files (`.localnode/dev*_key_seed.json`) are loaded
automatically. Extra keys can be appended via the `PRIVATE_KEYS` env var:

```bash
PRIVATE_KEYS=0xabc...,0xdef... ./system-tests.sh
```

A custom RPC endpoint can be specified if the localnode listens on a
non-default address:

```bash
RPC_URL=http://192.168.1.10:8545 ./system-tests.sh
```

### Testnet

Use the `testnet` network. `PRIVATE_KEYS` is mandatory (there are no seed
files to fall back on). See [Account requirements](#account-requirements)
for the number of keys needed per suite.

```bash
NETWORK=testnet PRIVATE_KEYS=0xabc...,0xdef... ./system-tests.sh
```

A custom RPC endpoint overrides the default (`https://rpc.test.mezo.org`):

```bash
NETWORK=testnet RPC_URL=https://custom-rpc.example.com PRIVATE_KEYS=0xabc... ./system-tests.sh
```

## Account requirements

Tests use accounts from `ethers.getSigners()`, which are derived from the
private keys supplied via seed files and/or the `PRIVATE_KEYS` env var (in
that order). Running the full suite requires **at least 3 private keys**.
On the localnode this is satisfied by the three seed files.

Some suites derive a **pool owner** signer at runtime via
`ethers.getSigner(await validatorPool.owner())`. This looks up the PoA
validator-pool owner address on-chain and expects to find a matching
private key among the configured accounts. On the localnode the `dev0`
seed key is the pool owner.

On the testnet all three keys must be supplied via `PRIVATE_KEYS`, the
first key must correspond to the PoA owner, and every account must hold
BTC for gas.

## Available test suites

| Suite | File | Description |
|-------|------|-------------|
| `AssetsBridge` | `AssetsBridge.test.ts` | Native bridge in/out operations |
| `BTCTransfers` | `BTCTransfers.test.ts` | BTC transfer mechanics |
| `MEZOTransfers` | `MEZOTransfers.test.ts` | MEZO token transfers |
| `TripartyBridge` | `TripartyBridge.test.ts` | Triparty BTC minting path |
| `Push0Check` | `Push0Check.test.ts` | EVM PUSH0 opcode support |
| `RandaoCheck` | `RandaoCheck.test.ts` | RANDAO/PREVRANDAO opcode |
| `McopyCheck` | `McopyCheck.test.ts` | EVM MCOPY opcode support |
| `TransientStorageCheck` | `TransientStorageCheck.test.ts` | EIP-1153 transient storage |
| `Selfdestruct6780Check` | `Selfdestruct6780Check.test.ts` | EIP-6780 SELFDESTRUCT behavior |
| `InitcodeLimitCheck` | `InitcodeLimitCheck.test.ts` | EIP-3860 initcode size limit |

## Environment variables

| Variable | Networks | Description |
|----------|----------|-------------|
| `NETWORK` | — | Hardhat network name (`localhost` or `testnet`, default: `localhost`) |
| `RPC_URL` | both | JSON-RPC endpoint (defaults: `http://127.0.0.1:8545` for localhost, `https://rpc.test.mezo.org` for testnet) |
| `PRIVATE_KEYS` | both | Comma-separated hex private keys. Appended to localnode seed keys on `localhost`; sole source on `testnet` |
