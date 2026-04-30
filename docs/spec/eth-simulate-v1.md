# `eth_simulateV1`

## Overview

`eth_simulateV1` runs an ordered sequence of simulated transactions across one
or more synthetic blocks, with per-block state and block-header overrides. It
returns per-call results (return data, logs, gas usage, status) and a
per-block envelope shaped like a normal Ethereum block.

Mezo ships `eth_simulateV1` so that the broader EVM tooling ecosystem
(ethers v6, viem, MetaMask, Rabby, debug UIs) — which increasingly assumes the
method is available — works against mezod without an external simulation
service. Mezo is the first Cosmos-SDK / Evmos-derived chain to implement it.

Compared to `eth_call`, `eth_simulateV1`:

- Executes a chain of calls instead of a single one, with state propagating
  across calls and across blocks within the same request.
- Supports synthetic blocks with header overrides (number, timestamp, base
  fee, gas limit, …) on top of a real anchor block.
- Returns per-call `logs` and `gasUsed` (`eth_call` returns only return data).
- Optionally validates each call as if it were a real transaction
  (`validation: true`).
- Optionally emits synthetic ERC-7528 `Transfer` events for every value-bearing
  call frame (`traceTransfers: true`).
- Allows `MovePrecompileTo` to relocate stdlib precompiles before the calls
  run.

## Implementation summary

- **Request envelope.** The typed JSON shapes (`SimOpts`, `SimBlock`,
  `SimCallResult`, `SimBlockResult`) and their strict unmarshaler live in
  `x/evm/types/`. Strict unmarshaling rejects fields mezod cannot honor before
  the request reaches the keeper.
- **Keeper driver.** `x/evm/keeper/simulate_v1.go` drives execution. A single
  `*statedb.StateDB` is shared across every call of every block in the
  request, so EVM-side and Cosmos-side writes (custom precompile state) chain
  correctly across calls. The driver sanitizes the input chain, applies state
  and block overrides, runs each call through `applyMessageWithConfig`, and
  assembles each simulated block via `ethtypes.NewBlock` so that header roots
  and bloom derive from the synthetic transactions and receipts.
- **JSON-RPC entry point.** `rpc/namespaces/ethereum/eth/simulate_v1.go`
  exposes the method on the `eth` namespace. `rpc/backend/simulate_v1.go`
  resolves the caller's anchor block to a concrete numeric height, plumbs the
  node-wide `RPCGasCap` and `RPCEVMTimeout`, and adapts the keeper's gRPC
  response back to a JSON-RPC response. Typed `*SimError` values surface
  end-to-end; the JSON-RPC server emits `{code, message, data}` directly.

## Conformance with the geth spec

The authoritative spec is the
[`ethereum/execution-apis`](https://github.com/ethereum/execution-apis)
repository:

- Schema: `src/eth/execute.yaml`.
- Conformance fixtures: `tests/eth_simulateV1/*.io` — 91 `.io` files pinning
  request shape, response shape, and error codes for every documented case.

Mezod ports a high-signal subset of those fixtures plus the per-feature
coverage cases into `tests/system/test/SimulateV1_SpecCompliance.test.ts`.
The system test asserts per-field invariants (status, gasUsed, log topics,
error code, block envelope shape) rather than byte-level response equality:
mezod's localnode chain id, base block, and account state never match the
upstream reference replay, so byte-for-byte response hashes are not a
meaningful check.

The same 91 fixtures are wired into the Go fuzz target
(`x/evm/keeper/simulate_v1_fuzz_test.go`) as the seed corpus for
`FuzzSimulateV1Opts`. The fuzz invariant is "no panics and every error
carries a documented `*SimError` code from
`x/evm/types/simulate_v1_errors.go`".

## Mezo-specific divergences

Each divergence has a tripwire test in
`tests/system/test/SimulateV1_MezoDivergence.test.ts` so that any accidental
spec-conformant flip surfaces loudly.

1. **EIP-4844, EIP-4788, EIP-4895 overrides rejected.** Mezo runs on
   CometBFT and has no beacon chain, no DA layer, and no validator-withdrawal
   queue. Overrides for `BlockOverrides.BeaconRoot` (EIP-4788),
   `BlockOverrides.Withdrawals` (EIP-4895), and the blob gas fields
   `BlobBaseFee` (EIP-4844) are rejected at parse time as
   `-32602 invalid params`. EIP-2935 (parent-hash storage contract) and
   EIP-7685 (general EL requests) are Prague-era surfaces that this Cancun-
   era build does not yet support; their post-upgrade status is tracked in
   [`.claude/MEZO-4336-eth-simulate-v1-geth116/plan.md`](../../.claude/MEZO-4336-eth-simulate-v1-geth116/plan.md).
2. **Custom mezo precompiles immovable.** `MovePrecompileTo` works for the
   standard precompiles at `0x01..0x0a`, but is rejected for any of the
   mezo custom precompiles enumerated in `x/evm/types/precompile.go`
   (`DefaultPrecompilesVersions`). The rejection is a structured `-32602`
   error (the geth spec does not assign a dedicated `-380xx` code; geth
   uses the same mapping for "source is not a precompile").
3. **`GasUsed` honors `MinGasMultiplier`.** Reported `gasUsed` matches what
   would appear on a mezod on-chain receipt:
   `gasUsed = max(gasLimit * MinGasMultiplier, raw_evm_gas)`. Raw EVM gas is
   used only for internal pool accounting. Callers comparing simulate results
   across chains should not assume mezod's `gasUsed` matches geth's.
4. **`stateRoot` is always the zero hash.** Mezod's `statedb.StateDB` wraps a
   Cosmos cached multistore and has no Merkle Patricia Trie, so there is no
   `IntermediateRoot()` to call after a simulated block executes. Echoing
   `base.Root` would be misleading (it would ignore everything the simulation
   did) and is explicitly rejected. Callers parsing the simulate response
   MUST NOT treat `stateRoot` as semantically meaningful on mezod.
5. **Insufficient-funds is per-call when `validation` is omitted.** Geth
   promotes a value-transfer balance failure to a top-level fatal `-38014`
   even with `validation=false`; mezod's `CanTransfer` fires per-call
   regardless of validation, surfacing the failure as per-call `status=0x0`
   with error code `-32015`. Pinned by
   `tests/system/test/SimulateV1_MezoDivergence.test.ts`.

## Key decisions

- **Full feature parity with `execution-apis`.** `traceTransfers`,
  `validation`, `returnFullTransactions`, and `MovePrecompileTo` are all
  supported, modulo the divergences above.
- **Single typed `*SimError` end-to-end.** One error type carries
  `{code, message, data}` from constructor through keeper, gRPC, and backend
  to the JSON-RPC server. Genuine internals collapse to
  `status.Error(codes.Internal, …)`. Error codes are catalogued in
  `x/evm/types/simulate_v1_errors.go`; no codes are declared anywhere else.
- **DoS config: kill switch only.** Operators get a single boolean
  (`SimulateDisabled`) plus the existing `RPCGasCap` and `RPCEVMTimeout`
  knobs. The 256-block span and 1000-call cumulative caps are hard-coded;
  per-feature `Simulate*` knobs are deferred until operational experience
  shows the shared-with-`eth_call` knobs are too coarse.
- **Validation error semantics are spec-conformant.** Tx-level validation
  failures (`-38010..-38025`) are top-level fatal errors that abort the whole
  request. Reverts and VM errors stay per-call (`3` and `-32015`).

## Configuration

- **`json-rpc.simulate-disabled`** (TOML key, configured per node). Maps to
  `JSONRPCConfig.SimulateDisabled`. When `true`, the JSON-RPC method returns
  `-32601 method not found` so a disabled node is indistinguishable from one
  that does not implement the method. The flag does **not** affect the SDK
  gRPC port (default 9090): direct gRPC peers can still invoke the keeper
  handler. To suppress simulate entirely, restrict that port at the network
  layer.
- **`json-rpc.gas-cap`** (`JSONRPCConfig.GasCap`). Used as the request-wide
  gas budget: the sum of `gas` consumed across every call of every block in
  one simulate request cannot exceed this cap.
- **`json-rpc.evm-timeout`** (`JSONRPCConfig.EVMTimeout`). Used as the
  request deadline. On expiry, the in-flight call is cancelled and the
  request returns `-32016 execution timeout`.

## Usage examples

Single-call happy path. Pre-fund the sender via `stateOverrides`, then simulate
one transfer at the latest block:

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "eth_simulateV1",
  "params": [
    {
      "blockStateCalls": [
        {
          "stateOverrides": {
            "0xc100000000000000000000000000000000000001": {
              "balance": "0x56bc75e2d63100000"
            }
          },
          "calls": [
            {
              "from": "0xc100000000000000000000000000000000000001",
              "to":   "0xc100000000000000000000000000000000000002",
              "value": "0x1"
            }
          ]
        }
      ]
    },
    "latest"
  ]
}
```

Multi-block with state and block overrides. Pre-fund a sender via
`stateOverrides`, then simulate two blocks with a bumped base fee in the
second. Per-block `time` is intentionally omitted — the driver auto-increments
synthetic-block timestamps from the parent block, and any explicit override
must be strictly greater than the parent (otherwise the request fails
`-38021`):

```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "eth_simulateV1",
  "params": [
    {
      "blockStateCalls": [
        {
          "stateOverrides": {
            "0xc100000000000000000000000000000000000001": {
              "balance": "0x56bc75e2d63100000"
            }
          },
          "calls": [
            {
              "from": "0xc100000000000000000000000000000000000001",
              "to":   "0xc100000000000000000000000000000000000002",
              "value": "0x1"
            }
          ]
        },
        {
          "blockOverrides": {
            "baseFeePerGas": "0x3b9aca00"
          },
          "calls": [
            {
              "from": "0xc100000000000000000000000000000000000001",
              "to":   "0xc100000000000000000000000000000000000003",
              "value": "0x2"
            }
          ]
        }
      ],
      "traceTransfers": true,
      "validation": false,
      "returnFullTransactions": false
    },
    "latest"
  ]
}
```

For more examples — including every documented error path, override shape,
and edge case — see the conformance fixtures under
[`ethereum/execution-apis/tests/eth_simulateV1/*.io`](https://github.com/ethereum/execution-apis/tree/main/tests/eth_simulateV1).
