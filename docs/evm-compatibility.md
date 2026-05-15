# EVM compatibility

## Overview

Mezo achieves EVM compatibility by implementing components that collectively
support EVM state transitions and maintaining a developer experience similar to
Ethereum.

## EVM forks support

### EVM forks up to London

Mezo offers EVM compatibility, supporting all Ethereum features
up to the London fork. For more information about the London fork, please see
[here](https://ethereum.org/en/history/#london).

### EVM forks post-London

Mezo sets post-London forks in its chain config. In some cases, Mezo's runtime
behavior can deviate from Ethereum. For example, Mezo does not support
PREVRANDAO (EIP-4399) and does not support blob transactions (EIP-4844).
See the fork/EIP notes below for details.

#### Arrow Glacier

Arrow Glacier (EIP-4345) only delayed Ethereum's PoW difficulty bomb and did
not change EVM execution.

#### Gray Glacier

Gray Glacier (EIP-5133) did the same; Mezo does not use a PoW difficulty bomb, so
neither Glacier fork affects Mezo's EVM compatibility.

#### Paris (The Merge)

Paris is Ethereum's Merge upgrade. It changed contract-visible EVM behavior
and also changed Ethereum consensus. Mezo runs on CometBFT, so only the EVM
execution semantics matter directly.

- EIP-4399 (PREVRANDAO / opcode `0x44`)
    - Description: changes opcode `0x44` from `DIFFICULTY` (PoW mining
      difficulty) to `PREVRANDAO` (beacon-chain randomness value).
    - Mezo implementation: PREVRANDAO is not supported. Mezo does not provide
      an EVM randomness value, so contracts observe `0` for opcode `0x44`.
      Contracts must not use `DIFFICULTY`/`PREVRANDAO` as a randomness source
      on Mezo.
    - Ref: https://eips.ethereum.org/EIPS/eip-4399

- EIP-3675 (The Merge consensus transition)
    - Description: transitions Ethereum block production from PoW to PoS.
    - Mezo implementation: not applicable to Mezo block production (Mezo runs
      on CometBFT).
    - Ref: https://eips.ethereum.org/EIPS/eip-3675

#### Shanghai (Shapella - execution part)

Shanghai is an execution-layer upgrade that changed EVM behavior and gas rules.

- EIP-3651 (Warm COINBASE)
    - Description: treats the block `COINBASE` address as 'warm' at the start
      of each tx, so the first access costs less gas. This only changes gas
      usage (except edge cases where a tx runs out of gas).
    - Mezo implementation: implemented in `StateDB.Prepare` by adding
      `coinbase` to the access list when Shanghai is active.
    - Ref: https://eips.ethereum.org/EIPS/eip-3651

- EIP-3855 (PUSH0)
    - Description: pre-Shanghai, contracts pushed `0` as `PUSH1 0x00`. Shanghai
      adds `PUSH0` (`0x5f`) to push `0` directly, making bytecode slightly
      smaller and cheaper without changing contract logic.
    - Mezo implementation: supported in the underlying VM when Shanghai is
      active.
    - Ref: https://eips.ethereum.org/EIPS/eip-3855

- EIP-3860 (Initcode size limit and metering)
    - Description: pre-Shanghai, there was no explicit initcode size limit.
      Shanghai caps initcode to 49152 bytes and charges extra gas (2 gas per
      32-byte word) so very large deployments cost more. If initcode is too
      large, contract creation fails. This only affects contract creation
      (deployments and CREATE/CREATE2), not normal execution of already
      deployed contracts.
    - Mezo implementation: supported in the underlying VM when Shanghai is
      active (create-tx size checks and initcode metering).
    - Ref: https://eips.ethereum.org/EIPS/eip-3860

- EIP-4895 (Beacon chain withdrawals)
    - Description: adds a system-level list of balance credits ("withdrawals")
      to blocks, used to pay out validator staking withdrawals into normal
      accounts. This is block processing, not a user transaction: the client
      applies these balance credits before transactions run.
    - Mezo implementation: Mezo does not have Beacon-chain withdrawals, so
      blocks won't include EIP-4895 withdrawals processing.
    - Ref: https://eips.ethereum.org/EIPS/eip-4895

- EIP-6049 (SELFDESTRUCT deprecation notice)
    - Description: adds an official warning that `SELFDESTRUCT` is deprecated
      and its behavior may change in future forks. EIP-6049 itself does not
      change EVM execution, so it does not change transaction outcomes.
    - Mezo implementation: `SELFDESTRUCT` is supported. See EIP-6780 below for
      the current Cancun fork semantics on Mezo.
    - Ref: https://eips.ethereum.org/EIPS/eip-6049

#### Cancun (Dencun - execution part)

Cancun is the execution-layer part of Dencun on Ethereum. It adds new opcodes
and runtime behavior, including transient storage and blob-transaction support.

Note: the full Dencun upgrade also includes consensus-layer EIPs. Mezo runs on
CometBFT, so only the execution-layer EIPs below apply directly.
See EIP-7569 for the full Dencun EIP list.

- EIP-1153 (Transient storage opcodes)
    - Description: adds `TLOAD` (`0x5c`) and `TSTORE` (`0x5d`) for
      "transaction-scoped" storage: contracts can store a value and read it
      back later in the same transaction (including across internal calls),
      but it is always cleared after the transaction finishes. This enables
      cheap per-tx caches and reentrancy locks without writing permanent storage.
    - Mezo implementation: supported via `StateDB` transient storage and reset
      at the start of each transaction.
    - Ref: https://eips.ethereum.org/EIPS/eip-1153

- EIP-4788 (Beacon block root in the EVM)
    - Description: makes a piece of Ethereum consensus-layer data available to
      contracts: the parent beacon block root. The execution client writes it
      to a fixed "system contract" address during block processing so
      contracts can query recent roots.
    - Mezo implementation: Mezo does not run Ethereum consensus, so there is
      no beacon root source. It is reasonable to treat EIP-4788 as out-of-scope
      because it depends on consensus-layer data that Mezo does not have.
    - Ref: https://eips.ethereum.org/EIPS/eip-4788

- EIP-4844 (Shard blob transactions)
    - Description: adds type-3 "blob" transactions and blob gas accounting.
      It also adds `BLOBHASH` and the KZG point evaluation precompile at `0x0a`.
    - Mezo implementation: Mezo shims EIP-4844 opcodes for compatibility but
      does not support blob transactions. Type-3 transactions are rejected.
      RPC block fields `blobGasUsed` and `excessBlobGas` are always `nil`.
      `BLOBHASH` returns `0` because blob hashes are never present.
      `BLOBBASEFEE` returns `0` because the blob gas market does not exist.
    - Ref: https://eips.ethereum.org/EIPS/eip-4844

- EIP-5656 (MCOPY)
    - Description: adds `MCOPY` (`0x5e`) to copy a range of EVM memory bytes
      in one instruction. It replaces slower patterns like loops over
      `MLOAD`/`MSTORE` or calling the identity precompile for memory copying.
      This is mainly a gas/performance change (except edge cases where a tx
      runs out of gas).
    - Mezo implementation: supported in the underlying VM when Cancun is
      active.
    - Ref: https://eips.ethereum.org/EIPS/eip-5656

- EIP-6780 (SELFDESTRUCT only in same transaction)
    - Description: `SELFDESTRUCT` transfers a contract's ETH balance to a
      beneficiary. Before Cancun, it also deleted the contract's code and
      storage at the end of the transaction. Since Cancun (EIP-6780), deletion
      only happens if the contract was created in the same transaction;
      otherwise only the balance is transferred.
    - Mezo implementation: supported in the underlying VM when Cancun is
      active.
    - Ref: https://eips.ethereum.org/EIPS/eip-6780

- EIP-7516 (BLOBBASEFEE opcode)
    - Description: adds `BLOBBASEFEE` to read the current blob base fee from
      the block header.
    - Mezo implementation: since Mezo rejects blob transactions, there is no
      real blob base fee. `BLOBBASEFEE` returns `0`.
    - Ref: https://eips.ethereum.org/EIPS/eip-7516

Reference list: Dencun meta EIP (execution + consensus): https://eips.ethereum.org/EIPS/eip-7569

#### Prague (Pectra - execution part)

Prague is the execution-layer part of Pectra on Ethereum. It adds new
precompiles, a new transaction type, block-header fields tied to a
beacon-chain side that Mezo does not have, and several gas-accounting
adjustments. The full Pectra upgrade also includes consensus-layer EIPs
(EIP-7251, EIP-7549, EIP-7691); since Mezo runs on CometBFT, those CL
EIPs are out of scope and not covered below. See EIP-7600 for the full
Pectra EIP list.

Prague is activated on live Mezo chains via the v11.0 upgrade handler
(see [`docs/upgrades.md`](./upgrades.md) and
`app/upgrades/v11_0/upgrades.go`), and is active from genesis on fresh
chains.

- EIP-2537 (BLS12-381 precompiles)
    - Description: adds seven precompiled contracts at addresses `0x0b`
      through `0x11` implementing BLS12-381 curve operations (G1Add,
      G1MSM, G2Add, G2MSM, PairingCheck, MapFpToG1, MapFp2ToG2).
    - Mezo implementation: supported via the vendored
      `mezo-org/go-ethereum` fork once Prague is active. Mezo's vendored
      geth was migrated from the pre-finalization 9-precompile draft
      layout (which placed standalone `G1Mul`/`G2Mul` at `0x0b`–`0x13`)
      to the finalized 7-precompile layout (`0x0b`–`0x11`). The
      `0x0b`–`0x11` slot range does not overlap with Mezo's custom
      precompile address space at `0x7b7c…`, so there is no clash. The
      live precompile surface is pinned by the
      `tests/system/test/Bls12381Check.test.ts` system test, which
      asserts presence, output sizes, and the empty-account state at
      the obsolete draft slots `0x12`/`0x13`.
    - Ref: https://eips.ethereum.org/EIPS/eip-2537

- EIP-2935 (Historical block hashes from state)
    - Description: deploys a "history storage" system contract at
      `0x0000F90827F1C53a10cb7A02335B175320002935` that stores the last
      8192 block hashes in its storage slots. The contract is meant to
      be deployed at fork activation. Upstream geth's payload processor
      calls `core.ProcessParentBlockHash` at the start of each block
      to update the ring buffer, and the `BLOCKHASH` opcode falls
      through to it for heights older than 256 blocks.
    - Mezo implementation: the system contract is not deployed and
      `core.ProcessParentBlockHash` is not called from Mezo's block
      processing path; the upstream BLOCKHASH-via-system-contract
      mechanism is not needed because Mezo resolves `BLOCKHASH` through
      its own `Keeper.GetHashFn`
      (`x/evm/keeper/state_transition.go:209`), which reads CometBFT
      headers persisted in `x/poa`'s historical info store. That store
      keeps up to `DefaultHistoricalEntries = 10000` entries
      (`x/poa/types/params.go:13`), which exceeds the 8192-block
      window EIP-2935 was designed to serve, so `BLOCKHASH` continues
      to work uniformly without the EIP-2935 system contract. The only
      observable divergence is that contracts reading the history
      storage address directly via `SLOAD` will find an empty account.
    - Ref: https://eips.ethereum.org/EIPS/eip-2935

- EIP-6110 (Validator deposits via EL system contract)
    - Description: replaces beacon-chain deposit tree processing with
      an execution-layer deposit system contract at
      `0x00000000219ab540356cBB839Cbe05303d7705Fa` whose `DepositEvent`
      logs are gathered by the EL into the block-level requests list.
    - Mezo implementation: not applicable. Mezo runs on CometBFT and
      has no beacon chain consuming deposits; the deposit system
      contract is not deployed. Mezo's transaction execution path
      (`x/evm/keeper/state_transition.go:ApplyTransaction`) does not
      invoke upstream geth's `StateProcessor.Process`, so no
      deposit-log parsing or requests collection is performed during
      block production and the missing contract cannot fail the block.
    - Ref: https://eips.ethereum.org/EIPS/eip-6110

- EIP-7002 (Execution-layer triggerable withdrawals)
    - Description: adds a withdrawal-request system contract at
      `0x00000961Ef480Eb55e80D19ad83579A64c007002` so a validator's
      withdrawal credential holder (an EL account) can trigger a
      partial or full exit by sending value to the contract; the
      contract emits requests gathered into the block-level requests
      list for the beacon chain.
    - Mezo implementation: not applicable. Same reason as EIP-6110 —
      no beacon chain, the withdrawal-request system contract is not
      deployed, and Mezo's block-processing path does not collect
      requests.
    - Ref: https://eips.ethereum.org/EIPS/eip-7002

- EIP-7623 (Calldata gas cost floor) — **IN PROGRESS**
    - Description: adds an additive intrinsic-gas floor charged per
      transaction based on its calldata token weight (10 gas per zero
      byte, 40 gas per non-zero byte), in addition to the existing
      per-byte cost. The floor only changes a transaction's gas
      consumption (except edge cases where a tx runs out of gas).
      Upstream geth enforces the floor inside `StateTransition.execute`
      by calling a separate `FloorDataGas` helper — `IntrinsicGas` on
      its own does not include the floor.
    - Mezo implementation: not yet applied. Mezo's transaction
      execution path bypasses upstream `StateTransition.execute` in
      favor of `Keeper.applyMessageWithConfig`
      (`x/evm/keeper/state_transition.go:498`), which calls
      `core.IntrinsicGas` via `Keeper.GetEthIntrinsicGas`
      (`x/evm/keeper/gas.go:34`) but never invokes
      `core.FloorDataGas`. As a result, data-heavy transactions on
      Mezo are charged the legacy intrinsic cost without the EIP-7623
      Prague floor. Bringing this in requires adding a Prague-gated
      `FloorDataGas` check to the keeper (and the matching ante-handler
      gas validation) so data-heavy transactions pay at least the floor
      amount and refunds are clipped against it the way upstream does
      it at `state_transition.go:532`.
    - Ref: https://eips.ethereum.org/EIPS/eip-7623

- EIP-7685 (General-purpose execution-layer requests)
    - Description: introduces a `requests` block-level list and a
      `requestsHash` header field that commits to all requests
      produced during block execution (EIP-6110 deposits, EIP-7002
      withdrawals, EIP-7251 consolidations). Engine-API consumers and
      the beacon chain rely on `requestsHash` to verify EL-CL
      communication.
    - Mezo implementation: not supported. Mezo has no EL→CL messaging
      (no beacon chain) and produces no requests, so there is nothing
      to commit. Block JSON-RPC responses built by
      `rpc/types/utils.go:FormatBlock` omit the `requestsHash` field
      entirely (the field is not present in the response map — it is
      not emitted as `null`), and `eth_simulateV1` leaves the
      geth-side `Header.RequestsHash` nil
      (`x/evm/keeper/simulate_v1.go:165`). Tooling that hard-requires
      `requestsHash` on a Prague-active chain will need to treat its
      absence as the Mezo-specific shape.
    - Ref: https://eips.ethereum.org/EIPS/eip-7685

- EIP-7702 (Set Code Transactions)
    - Description: adds transaction type `0x04` ("set code"), which lets
      an externally owned account (EOA) sign one or more authorization
      tuples that install a per-account delegation designator pointing at
      a contract address. Calls to the delegated EOA execute the target
      contract's code in the EOA's storage and balance context, with the
      EOA's address as `CALLER`. The designator is a 23-byte code value
      `0xef0100 || target_address` written to the authority's code field;
      the authority rotates or clears it by signing a fresh
      authorization (clearing is `target = 0x0`).
    - Mezo implementation: type-`0x04` transactions are accepted once
      Prague is active. The keeper applies authorization tuples to mezod's
      `statedb.StateDB` (delegation install, rotation via re-signing,
      clearing via `target = 0x0`); `EthAccountVerificationDecorator` gains
      a delegation-aware EIP-3607 exemption so a delegated EOA can still
      send transactions; `EXTCODESIZE`/`EXTCODECOPY`/`EXTCODEHASH` return
      the raw 23-byte designator so on-chain contracts can detect a
      delegation; authorizations whose target is in `evm.Precompiles()`
      (the union of stock and Mezo custom precompiles) are rejected per
      the EIP's per-tuple rule. See
      [`docs/spec/eip7702-set-code.md`](./spec/eip7702-set-code.md) for
      the canonical mezod behavior, configuration, and Mezo-specific
      divergences from upstream geth.
    - Ref: https://eips.ethereum.org/EIPS/eip-7702

- EIP-7840 (Blob schedule configuration)
    - Description: makes the per-fork blob target/maximum/base-fee
      update fraction part of `ChainConfig` rather than client-hardcoded
      constants, so chains can carry different schedules without
      forking client code.
    - Mezo implementation: Mezo's `ChainConfig.EthereumConfig()` wires
      `params.DefaultBlobSchedule` (see
      `x/evm/types/chain_config.go:64`) so the geth side sees a valid
      schedule covering Cancun, Prague, and Osaka — required to pass
      `CheckConfigForkOrder`. The schedule is dormant: Mezo rejects
      EIP-4844 blob transactions entirely, so no blob fee market
      operates against the schedule values.
    - Ref: https://eips.ethereum.org/EIPS/eip-7840

- EIP-7642 (eth/69 wire-protocol cleanup)
    - Description: drops the legacy `td` field from devp2p `eth/68`
      messages and bumps the protocol version to `eth/69`. It is a
      peer-to-peer wire-protocol cleanup with no contract-visible
      effect.
    - Mezo implementation: not applicable. Mezo's peer-to-peer layer is
      CometBFT, not devp2p; there is no `eth/68`/`eth/69` handler.
      `eth_protocolVersion` returns a static value
      (`types.ProtocolVersion = eth65`,
      `rpc/namespaces/ethereum/eth/api.go:312`) and is not a real
      negotiated wire version.
    - Ref: https://eips.ethereum.org/EIPS/eip-7642

Reference list: Pectra meta EIP (execution + consensus): https://eips.ethereum.org/EIPS/eip-7600

#### Osaka (Fusaka - execution part)

Osaka is the execution-layer part of Fusaka on Ethereum. It tightens
gas accounting around MODEXP, caps per-transaction gas, adds a new
opcode and precompile, and exposes chain-config metadata via JSON-RPC.
The full Fusaka upgrade also includes consensus-layer EIPs (EIP-7594
PeerDAS, EIP-7917 proposer lookahead); since Mezo runs on CometBFT,
those CL EIPs are out of scope and not covered below. See EIP-7607 for
the full Fusaka EIP list.

Osaka is **not yet active on Mezo**. `OsakaTime` is intentionally
left nil in `x/evm/types/chain_config.go` (see the comment at
lines 89-103). The vendored `mezo-org/go-ethereum v1.16.9-mezo1` fork
already carries the Osaka opcode, precompile, and gas-schedule
changes, so once `OsakaTime` is set the EVM will fire them
automatically without further code changes — the work that remains
before flipping the switch is audit of the keeper, ante handler, and
RPC surface for Osaka's behavior changes (tracked under MEZO-4014),
followed by a planned upgrade handler that writes `OsakaTime` to the
stored chain config (mirroring how `v11.0` set `PragueTime`). Until
that lands, every Osaka EIP below is flagged **IN PROGRESS** to make
clear that, even though the upstream code is present, no contract or
RPC caller on Mezo currently observes the Osaka behavior — Mezo's
chain rules report `IsOsaka == false` at every height.

- EIP-7823 (MODEXP input upper bound) — **IN PROGRESS**
    - Description: rejects MODEXP precompile calls whose base, exponent,
      or modulus length exceeds 8192 bits, capping inputs that would
      otherwise dominate block validation cost.
    - Mezo implementation: present in the vendored geth fork, dormant
      until `OsakaTime` is set. Today, the upstream `modexp`
      precompile does not enforce the 8192-bit input bound for Mezo
      callers because the check is gated by `IsOsaka` inside geth.
      Mezo applies no additional MODEXP overrides, so the cap will
      take effect uniformly the moment an upgrade handler activates
      Osaka.
    - Ref: https://eips.ethereum.org/EIPS/eip-7823

- EIP-7825 (Transaction gas limit cap) — **IN PROGRESS**
    - Description: caps the maximum gas a single transaction can
      declare to `2^24 = 16_777_216` gas (≈16.78M), independent of
      block gas limit, to bound worst-case transaction validation.
    - Mezo implementation: present in the vendored geth fork, dormant
      until `OsakaTime` is set. Today, transactions with declared
      `gasLimit` above 16.78M are not rejected on Osaka-cap grounds —
      they are only constrained by Mezo's existing ante-handler check
      against `maxGasWanted` and the block gas limit
      (`app/ante/evm/eth.go:EthGasConsumeDecorator`). Those Mezo
      checks compare against the configured per-tx ceiling and the
      block ceiling, not against a hardcoded Ethereum value, so
      switching `OsakaTime` on does not collide with them — the
      upstream 16.78M cap will simply apply as an additional, lower
      ceiling inside the EVM.
    - Ref: https://eips.ethereum.org/EIPS/eip-7825

- EIP-7883 (MODEXP gas cost increase) — **IN PROGRESS**
    - Description: bumps MODEXP's gas cost: the minimum charge rises
      from 200 to 500 gas, and the per-iteration cost for large
      operands doubles. Targets the same DoS-shape that EIP-7823
      addresses by input bound.
    - Mezo implementation: present in the vendored geth fork, dormant
      until `OsakaTime` is set. Today, MODEXP calls are still charged
      under the Cancun gas schedule; the new minimum and the doubled
      per-iteration cost will apply once Osaka activates. Mezo applies
      no additional MODEXP overrides on top of upstream. The TODO in
      `x/evm/types/chain_config.go:89-93` calls this out as a
      reason to audit the keeper, ante handler, and RPC surface
      before defaulting `OsakaTime` to zero at genesis.
    - Ref: https://eips.ethereum.org/EIPS/eip-7883

- EIP-7934 (RLP block-size limit)
    - Description: caps the RLP-encoded size of a block at 10 MiB to
      bound network-propagation and storage costs.
    - Mezo implementation: not applicable. Mezo blocks are
      CometBFT-encoded; an RLP block representation only exists at the
      JSON-RPC boundary in `rpc/types/utils.go:FormatBlock`. Block
      size limits on Mezo come from CometBFT consensus parameters
      (`cmd/mezod/init.go:118` sets `MaxGas`), not the Ethereum RLP
      cap.
    - Ref: https://eips.ethereum.org/EIPS/eip-7934

- EIP-7935 (Default gas limit raise to 60M)
    - Description: raises Ethereum's default block gas limit target
      from 30M to 60M.
    - Mezo implementation: not applicable. Mezo's block gas limit is
      governed by `x/feemarket` (see `app.init()` in `app/app.go`
      wiring `MainnetMinGasPrices` and `MainnetMinGasMultiplier`) and
      the CometBFT consensus parameter `Block.MaxGas`, not Ethereum's
      default.
    - Ref: https://eips.ethereum.org/EIPS/eip-7935

- EIP-7939 (`CLZ` opcode) — **IN PROGRESS**
    - Description: adds opcode `0x1e` `CLZ` (count leading zeros) on
      a 256-bit stack word, returning the number of leading zero bits.
      Mirrors `BIT.popcnt`-style helpers used by gas-tight algorithms.
    - Mezo implementation: present in the vendored geth fork, dormant
      until `OsakaTime` is set. Today, executing opcode `0x1e` in
      Mezo bytecode reverts as `invalid opcode`; once an upgrade
      handler activates Osaka the EVM jump table will route `0x1e` to
      the `CLZ` implementation automatically. No Mezo-side override is
      required.
    - Ref: https://eips.ethereum.org/EIPS/eip-7939

- EIP-7951 (secp256r1 `P256VERIFY` precompile) — **IN PROGRESS**
    - Description: adds a precompile at
      `0x0000000000000000000000000000000000000100` that verifies an
      ECDSA signature on the secp256r1 (NIST P-256) curve.
    - Mezo implementation: present in the vendored geth fork, dormant
      until `OsakaTime` is set. Today, a `STATICCALL` to `0x0100`
      lands on an empty account and returns no data; once Osaka
      activates, the precompile registered by upstream geth will
      respond. The `0x0100` slot sits in the standard Ethereum
      precompile range and does not overlap Mezo's custom precompile
      space at `0x7b7c…` (see `x/evm/types/precompile.go:4-57`), so
      no address relocation is required — a sister-check exercise to
      MEZO-4004 (the BLS12-381 address conflict resolution) confirmed
      the slot is free on Mezo.
    - Ref: https://eips.ethereum.org/EIPS/eip-7951

- EIP-7892 (Blob-parameter-only forks)
    - Description: introduces "BPO" forks whose only effect is to swap
      the active blob schedule entry, without invoking a full
      execution-layer fork.
    - Mezo implementation: not applicable. `BPO1Time..BPO5Time` fields
      exist on Mezo's `ChainConfig` purely to satisfy
      `params.CheckConfigForkOrder` against the upstream struct
      (`x/evm/types/chain_config.go:54-58`) and are intentionally
      left nil (see the comment at lines 95-103). Mezo rejects blob
      transactions, so a BPO schedule swap would have no
      contract-visible effect even if activated.
    - Ref: https://eips.ethereum.org/EIPS/eip-7892

- EIP-7918 (Blob base fee floor)
    - Description: changes the blob fee market so the base fee per
      blob gas cannot fall below an execution-block-related floor,
      preventing pathological zero-fee blob inclusion.
    - Mezo implementation: not applicable. Mezo rejects EIP-4844 blob
      transactions entirely; `BLOBBASEFEE` returns `0` and there is no
      blob fee market on which to apply a floor.
    - Ref: https://eips.ethereum.org/EIPS/eip-7918

- EIP-7642 (eth/69 wire-protocol cleanup)
    - Description: see the Prague section above. EIP-7642 is referenced
      by both Pectra and Fusaka meta-EIPs.
    - Mezo implementation: not applicable for the same reason — Mezo
      has no devp2p layer.
    - Ref: https://eips.ethereum.org/EIPS/eip-7642

- EIP-7910 (`eth_config` JSON-RPC method)
    - Description: adds the `eth_config` JSON-RPC method, which
      returns the active chain configuration (fork activation times,
      blob schedule, chain ID, precompile list) so clients and tracers
      can discover post-Fusaka behavior without out-of-band metadata.
    - Mezo implementation: out of scope for now. The `EthereumAPI`
      interface in `rpc/namespaces/ethereum/eth/api.go` does not
      declare an `EthConfig`/`Config` method and the `PublicAPI`
      struct does not register one, so callers receive a
      method-not-found error. Support may be added in a later release
      once tooling demand for `eth_config` justifies the surface.
    - Ref: https://eips.ethereum.org/EIPS/eip-7910

Reference list: Fusaka meta EIP (execution + consensus): https://eips.ethereum.org/EIPS/eip-7607

## EVM JSON-RPC API reference

The `mezod` node exposes the following JSON-RPC API. The reference is split into specific namespaces.

### `web3` namespace

#### web3_clientVersion

- **Description**: Returns the current client version.
- **Parameters**: None.
- **Returns**: `String` - The current client version.

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"web3_clientVersion","params":[],"id":1}' -H "Content-Type: application/json" https://rpc.test.mezo.org
```

#### web3_sha3

- **Description**: Returns Keccak-256 of the given data.
- **Parameters**:
    - `String` - Data to convert into a SHA3 hash.
- **Returns**: `String` - The SHA3 result of the given data.

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"web3_sha3","params":["0x68656c6c6f20776f726c64"],"id":1}' -H "Content-Type: application/json" https://rpc.test.mezo.org
```

### `net` namespace

#### net_version

- **Description**: Returns the current network ID.
- **Parameters**: None.
- **Returns**: `String` - The current network ID.

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"net_version","params":[],"id":1}' -H "Content-Type: application/json" https://rpc.test.mezo.org
```

#### net_listening

- **Description**: Returns `true` if the client is actively listening for network connections.
- **Parameters**: None.
- **Returns**: `Boolean` - `true` if listening, `false` otherwise.

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"net_listening","params":[],"id":1}' -H "Content-Type: application/json" https://rpc.test.mezo.org
```

#### net_peerCount

- **Description**: Returns the number of peers currently connected to the client.
- **Parameters**: None.
- **Returns**: `String` - Number of connected peers.

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"net_peerCount","params":[],"id":1}' -H "Content-Type: application/json" https://rpc.test.mezo.org
```

### `eth` namespace

#### eth_protocolVersion

- **Description**: Returns the current Ethereum protocol version.
- **Parameters**: None.
- **Returns**: `String` - The Ethereum protocol version.

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_protocolVersion","params":[],"id":1}' -H "Content-Type: application/json" https://rpc.test.mezo.org
```

#### eth_syncing

- **Description**: Returns an object with data about the sync status or `false` if not syncing.
- **Parameters**: None.
- **Returns**:
    - `Object | Boolean` - An object with sync status or `false`.

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_syncing","params":[],"id":1}' -H "Content-Type: application/json" https://rpc.test.mezo.org
```

#### eth_coinbase

- **Description**: Returns the client's coinbase address (mining beneficiary).
This address is where any mining rewards will be sent if the node is mining.
- **Parameters**: None.
- **Returns**: `String` - Coinbase address.

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_coinbase","params":[],"id":1}' -H "Content-Type: application/json" https://rpc.test.mezo.org
```

#### eth_chainId

- **Description**: Returns the client's chain ID.
- **Parameters**: None.
- **Returns**: `String` - Chain ID.

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_chainId","params":[],"id":1}' -H "Content-Type: application/json" https://rpc.test.mezo.org
```

#### eth_gasPrice

- **Description**: Returns the current price per gas in wei.
- **Parameters**: None.
- **Returns**: `String` - The current gas price in wei.

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_gasPrice","params":[],"id":1}' -H "Content-Type: application/json" https://rpc.test.mezo.org
```

#### eth_accounts

- **Description**: Returns a list of addresses owned by the client.
- **Parameters**: None.
- **Returns**: `Array` - Array of addresses.

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_accounts","params":[],"id":1}' -H "Content-Type: application/json" https://rpc.test.mezo.org
```

#### eth_blockNumber

- **Description**: Returns the number of the most recent block.
- **Parameters**: None.
- **Returns**: `String` - The block number.

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' -H "Content-Type: application/json" https://rpc.test.mezo.org
```

#### eth_getBalance

- **Description**: Returns the balance of the account at the given address.
- **Parameters**:
    - `String` - Address to check for balance.
    - `String` - Block number.
- **Returns**: `String` - The balance in abtc, as a hexadecimal.

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_getBalance","params":["0x0504d82efb7db7a8c05e8df8cea575d8c9f48bb2","latest"],"id":1}' -H "Content-Type: application/json" https://rpc.test.mezo.org
```

#### eth_getStorageAt

- **Description**: Returns the value from a storage position at a given address.
- **Parameters**:
    - `String` - Address of the storage.
    - `String` - Storage position.
    - `String` - Block number.
- **Returns**: `String` - The value at this storage position.

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_getStorageAt","params":["0x0504d82efb7db7a8c05e8df8cea575d8c9f48bb2","0x0","latest"],"id":1}' -H "Content-Type: application/json" https://rpc.test.mezo.org
```

#### eth_getTransactionCount

- **Description**: Returns the number of transactions sent from an address.
- **Parameters**:
    - `String` - Address to check for transaction count.
    - `String` - Block number.
- **Returns**: `String` - The transaction count as a hexadecimal number.

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_getTransactionCount","params":["0x0504d82efb7db7a8c05e8df8cea575d8c9f48bb2","latest"],"id":1}' -H "Content-Type: application/json" https://rpc.test.mezo.org
```

#### eth_getBlockTransactionCountByHash

- **Description**: Returns the number of transactions in a block from a block
matching the given block hash.
- **Parameters**:
    - `String` - Block hash.
- **Returns**: `String` - The number of transactions in the block as a hexadecimal.

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_getBlockTransactionCountByHash","params":["0x41175c10b68dd0bfa27f2533a23979445a5d643427e0ffd1870d11806f31b291"],"id":1}' -H "Content-Type: application/json" https://rpc.test.mezo.org
```

#### eth_getBlockTransactionCountByNumber

- **Description**: Returns the number of transactions in a block matching the given block number.
- **Parameters**:
    - `String` - Block number.
- **Returns**: `String` - The number of transactions in the block as a hexadecimal.

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_getBlockTransactionCountByNumber","params":["0x1"],"id":1}' -H "Content-Type: application/json" https://rpc.test.mezo.org
```

#### eth_getCode

- **Description**: Returns the code stored at a given address. For a contract
account this is the deployed bytecode; for a plain externally owned account
(EOA) this is empty (`0x`). On a Prague-active chain a third case applies:
an EOA that has signed an EIP-7702 authorization returns its 23-byte
delegation designator `0xef0100 || target_address` rather than empty bytes.
Callers that key behavior off "code present means contract" must account for
the designator case. See [`docs/spec/eip7702-set-code.md`](./spec/eip7702-set-code.md)
and the [Prague](#prague) section for details.
- **Parameters**:
    - `String` - Address to get code from.
    - `String` - Block number.
- **Returns**: `String` - The code at the given address: deployed bytecode for
a contract, `0x` for a plain EOA, or a 23-byte EIP-7702 delegation designator
for a delegated EOA.

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_getCode","params":["0x0504d82efb7db7a8c05e8df8cea575d8c9f48bb2","latest"],"id":1}' -H "Content-Type: application/json" https://rpc.test.mezo.org
```

#### eth_sign

- **Description**: Signs data with a given address, resulting in a signature. The address to sign with must be unlocked.
- **Parameters**:
    - `String` - Address to sign with.
    - `String` - Data to sign.
- **Returns**: `String` - The signature.

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_sign","params":["0x0504d82efb7db7a8c05e8df8cea575d8c9f48bb2","0xdeadbeef"],"id":1}' -H "Content-Type: application/json" https://rpc.test.mezo.org
```

#### eth_sendTransaction

- **Description**: Creates and sends a new transaction.
- **Parameters**:
    The transaction call object:
    - `from`: DATA, 20 Bytes - (optional) The address the transaction is sent from.
    - `to`: DATA, 20 Bytes - (optional when creating new contract) The address the
      transaction is directed to.
    - `gas`: QUANTITY - (optional, default: 90000) Integer of the gas provided
      for the transaction execution. It will return unused gas.
    - `gasPrice`: QUANTITY - (optional) Integer of the gasPrice used for each paid gas
    - `value`: QUANTITY - (optional) Integer of the value sent with this transaction
    - `input`: DATA The compiled code of a contract OR the hash of the invoked method
      signature and encoded parameters.
    - `nonce`: QUANTITY - (optional) Integer of a nonce. This allows to overwrite
      your own pending transactions that use the same nonce.
    - `authorizationList`: Array of objects - (optional) EIP-7702 authorization
      tuples (`chainId`, `address`, `nonce`, `v`, `r`, `s`) applied before
      the transaction executes. Submitting a non-empty list produces a
      type-`0x04` transaction; each entry adds an intrinsic gas charge. See
      [`docs/spec/eip7702-set-code.md`](./spec/eip7702-set-code.md).
- **Returns**: `String` - The transaction hash.

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_sendTransaction","params":[{see above}],"id":1}' -H "Content-Type: application/json" https://rpc.test.mezo.org
```

#### eth_sendRawTransaction

- **Description**: Creates new message call transaction or a contract creation
for signed transactions. On a Prague-active chain, signed type-`0x04`
(EIP-7702 set-code) transactions are accepted alongside legacy, access-list
and dynamic-fee types.
- **Parameters**:
    - `String` - The signed transaction data.
- **Returns**: `String` - The transaction hash.

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_sendRawTransaction","params":["0xd46e8dd67c5d32be8d46e8dd67c5d32be8058bb8eb970870f072445675058bb8eb970870f072445675"],"id":1}' -H "Content-Type: application/json" https://rpc.test.mezo.org
```

#### eth_call

- **Description**: Executes a new message call immediately without creating a
transaction on the blockchain. Often used for executing read-only smart contract
functions.
- **Parameters**:
    - `Object` - The transaction call object
        - from: DATA, 20 Bytes - (optional) The address the transaction is sent from.
        - to: DATA, 20 Bytes - The address the transaction is directed to.
        - gas: QUANTITY - (optional) Integer of the gas provided for the transaction
          execution. eth_call consumes zero gas, but this parameter may be needed by
          some executions.
        - gasPrice: QUANTITY - (optional) Integer of the gasPrice used for each paid gas
        - value: QUANTITY - (optional) Integer of the value sent with this transaction
        - input: DATA - (optional) Hash of the method signature and encoded parameters.
          For details see Ethereum Contract ABI in the Solidity documentation(opens in
          a new tab).
        - authorizationList: Array of objects - (optional) EIP-7702 authorization
          tuples (`chainId`, `address`, `nonce`, `v`, `r`, `s`) applied before
          the call executes. See
          [`docs/spec/eip7702-set-code.md`](./spec/eip7702-set-code.md).
    - `String` (optional) - Block number.
    - `Object` (optional) - State override set. A mapping of addresses to
      override objects. Each override object may contain:
        - balance: QUANTITY - (optional) Balance to set for the account before
          executing the call.
        - nonce: QUANTITY - (optional) Nonce to set for the account.
        - code: DATA - (optional) EVM bytecode to inject at the account address.
        - state: Object - (optional) Key-value mapping of storage slots to
          override. Replaces the entire storage of the account.
        - stateDiff: Object - (optional) Key-value mapping of individual storage
          slots to override. Merges with existing storage. Cannot be combined
          with `state`.
- **Returns**: `String` - The return value of the executed contract.

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_call","params":[{see above}],"id":1}' -H "Content-Type: application/json" https://rpc.test.mezo.org
```

#### eth_simulateV1

- **Description**: Simulates a chain of transactions across one or more
synthetic blocks, with state and block overrides. See
[`docs/spec/eth-simulate-v1.md`](./spec/eth-simulate-v1.md) for the full spec.

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_simulateV1","params":[{"blockStateCalls":[{"stateOverrides":{"0xc100000000000000000000000000000000000001":{"balance":"0x56bc75e2d63100000"}},"calls":[{"from":"0xc100000000000000000000000000000000000001","to":"0xc100000000000000000000000000000000000002","value":"0x1"}]}]},"latest"],"id":1}' -H "Content-Type: application/json" https://rpc.test.mezo.org
```

#### eth_estimateGas

- **Description**: Estimates the gas necessary to execute a transaction.
- **Parameters**:
    - `Object` - The transaction call object
        - from: DATA, 20 Bytes - The address the transaction is sent from.
        - to: DATA, 20 Bytes - (optional when creating new contract) The address the
          transaction is directed to.
        - gas: QUANTITY - (optional) Integer of the gas provided for the transaction
          execution.
        - gasPrice: QUANTITY - (optional) Integer of the gasPrice used for each paid gas
        - value: QUANTITY - (optional) Integer of the value sent with this transaction
        - input: DATA - (optional) Hash of the method signature and encoded parameters.
        - authorizationList: Array of objects - (optional) EIP-7702 authorization
          tuples (`chainId`, `address`, `nonce`, `v`, `r`, `s`) included in the
          estimated transaction. Each entry adds an intrinsic gas charge to the
          returned estimate. See
          [`docs/spec/eip7702-set-code.md`](./spec/eip7702-set-code.md).
    - `String` (optional) - Block number.
    - `Object` (optional) - State override set. A mapping of addresses to
      override objects. Each override object may contain:
        - balance: QUANTITY - (optional) Balance to set for the account before
          executing the call.
        - nonce: QUANTITY - (optional) Nonce to set for the account.
        - code: DATA - (optional) EVM bytecode to inject at the account address.
        - state: Object - (optional) Key-value mapping of storage slots to
          override. Replaces the entire storage of the account.
        - stateDiff: Object - (optional) Key-value mapping of individual storage
          slots to override. Merges with existing storage. Cannot be combined
          with `state`.
- **Returns**: `String` - The estimated gas amount.

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_estimateGas","params":[{"from":"0xFF3014B077D307E7B0bf262d072B25dbE19E2Be3","to":"0xd3CdA913deB6f67967B99D67aCDFa1712C293601","value":"0x186a0"}],"id":1}' -H "Content-Type: application/json" https://rpc.test.mezo.org
```

#### eth_getBlockByHash

- **Description**: Returns information about a block by hash.
- **Parameters**:
    - `String` - Block hash.
    - `Boolean` - If `true`, returns full transaction objects; if `false`, returns only hashes.
- **Returns**: `Object` - Block information.

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_getBlockByHash","params":["0x8d80d1a8ac12c5e57c17c580afbb4c03987649934b60ce04ec89fcd336e3a186", true],"id":1}' -H "Content-Type: application/json" https://rpc.test.mezo.org
```

#### eth_getBlockByNumber

- **Description**: Returns information about a block by number.
- **Parameters**:
    - `String` - Block number.
    - `Boolean` - If `true`, returns full transaction objects; if `false`, returns only hashes.
- **Returns**: `Object` - Block information.

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["0x1b4", true],"id":1}' -H "Content-Type: application/json" https://rpc.test.mezo.org
```

#### eth_getTransactionByHash

- **Description**: Returns the information about a transaction requested by transaction hash.
- **Parameters**:
    - `String` - Transaction hash.
- **Returns**: `Object` - Transaction information.

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_getTransactionByHash","params":["0x5c504ed432cb51138bcf09aa5e8a410dd4a1e204ef84bfed1be16dfba1b22060"],"id":1}' -H "Content-Type: application/json" https://rpc.test.mezo.org
```

#### eth_getTransactionByBlockHashAndIndex

- **Description**: Returns information about a transaction by block hash and transaction index position.
- **Parameters**:
    - `String` - Block hash.
    - `String` - Transaction index position.
- **Returns**: `Object` - Transaction information.

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_getTransactionByBlockHashAndIndex","params":["0x8d80d1a8ac12c5e57c17c580afbb4c03987649934b60ce04ec89fcd336e3a186", "0x0"],"id":1}' -H "Content-Type: application/json" https://rpc.test.mezo.org
```

#### eth_getTransactionByBlockNumberAndIndex

- **Description**: Returns information about a transaction by block number and transaction index position.
- **Parameters**:
    - `String` - Block number.
    - `String` - Transaction index position.
- **Returns**: `Object` - Transaction information.

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_getTransactionByBlockNumberAndIndex","params":["0x1b4", "0x0"],"id":1}' -H "Content-Type: application/json" https://rpc.test.mezo.org
```

#### eth_getTransactionReceipt

- **Description**: Returns the receipt of a transaction by transaction hash.
- **Parameters**:
    - `String` - Transaction hash.
- **Returns**: `Object` - Transaction receipt.

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_getTransactionReceipt","params":["0x1758f2ad26d448ecdcc2f225432c520bc77c03194536e76f6776f8c5dabce9a9"],"id":1}' -H "Content-Type: application/json" https://rpc.test.mezo.org
```

#### eth_getLogs

- **Description**: Returns an array of logs matching the filter options.
- **Parameters**:
    Filter options
    - `fromBlock`: QUANTITY|TAG - (optional, default: "latest") Integer block number, or "latest" for the last mined
    block or "pending", "earliest" for not yet mined transactions.
    - `toBlock`: QUANTITY|TAG - (optional, default: "latest") Integer block number, or "latest" for the last mined block
    or "pending", "earliest" for not yet mined transactions.
    - `address`: DATA|Array, 20 Bytes - (optional) Contract address or a list of addresses from which logs should originate.
    - `topics`: Array of DATA, - (optional) Array of 32 Bytes DATA topics. Topics are order-dependent. Each topic can
    also be an array of DATA with "or" options.
    - `blockhash`: (optional, future) With the addition of EIP-234, blockHash will be a new filter option which restricts
    the logs returned to the single block with the 32-byte hash blockHash. Using blockHash is equivalent to fromBlock =
    toBlock = the block number with hash blockHash. If blockHash is present in in the filter criteria, then neither fromBlock
    nor toBlock are allowed.
- **Returns**: `Array` - Array of log objects.

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_getLogs","params":[{"fromBlock": "0x1", "toBlock": "0x2"}],"id":1}' -H "Content-Type: application/json" https://rpc.test.mezo.org
```

#### eth_newFilter

- **Description**: Create new filter using topics of some kind.
- **Parameters**:
    - `String` - hash of a transaction
- **Returns**: `String` - Filter ID.

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_newFilter","params":[{"fromBlock": "0x1", "toBlock": "0x2"}],"id":1}' -H "Content-Type: application/json" https://rpc.test.mezo.org
```

#### eth_newBlockFilter

- **Description**: Creates a filter in the node to notify when a new block arrives.
- **Parameters**: None.
- **Returns**: `String` - A filter ID.

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_newBlockFilter","params":[],"id":1}' -H "Content-Type: application/json" https://rpc.test.mezo.org
```

#### eth_newPendingTransactionFilter

- **Description**: Creates a filter in the node to notify when new pending transactions arrive.
- **Parameters**: None.
- **Returns**: `String` - A filter ID.

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_newPendingTransactionFilter","params":[],"id":1}' -H "Content-Type: application/json" https://rpc.test.mezo.org
```

#### eth_uninstallFilter

- **Description**: Uninstalls a filter with the given ID.
- **Parameters**:
    - `String` - The filter ID.
- **Returns**: `Boolean` - `true` if the filter was successfully uninstalled.

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_uninstallFilter","params":["0x1"],"id":1}' -H "Content-Type: application/json" https://rpc.test.mezo.org
```

#### eth_getFilterChanges

- **Description**: Checks for changes to a filter since the last call.
- **Parameters**:
    - `String` - The filter ID.
- **Returns**: `Array` - An array of logs that have occurred since the last poll.

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_getFilterChanges","params":["0x1"],"id":1}' -H "Content-Type: application/json" https://rpc.test.mezo.org
```

#### eth_getFilterLogs

- **Description**: Returns an array of all logs for a given filter ID, containing all past logs matching the filter.
- **Parameters**:
    - `String` - The filter ID.
- **Returns**: `Array` - An array of log objects that match the filter, providing historical log data.

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_getFilterLogs","params":["0x1"],"id":1}' -H "Content-Type: application/json" https://rpc.test.mezo.org
```

#### eth_getProof

- **Description**: Returns the account and storage values of a specified account, including the Merkle proof.
- **Parameters**:
    - `String` - Address of the account.
    - `Array` of `String` - An array of storage keys that you want the values and proof for.
    - `String` - Block number to get the proof for.
- **Returns**: `Object` - An object containing the account details, storage proof, and relevant Merkle proofs.

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_getProof","params":["0x1234567890123456789012345678901234567890",["0x0000000000000000000000000000000000000000000000000000000000000000","0x0000000000000000000000000000000000000000000000000000000000000001"],`"latest"`],"id":1}' -H "Content-type:application/json" https://rpc.test.mezo.org
```

### `eth` namespace - not supported methods

Due to the nature of PoA consensus, the following methods might revert or return empty results.

#### eth_getUncleCountByBlockHash

#### eth_getUncleCountByBlockNumber

#### eth_getUncleByBlockHashAndIndex

#### eth_getUncleByBlockNumberAndIndex

#### eth_getWork

#### eth_submitWork

#### eth_submitHashrate

#### eth_hashrate

#### eth_mining

### `debug` namespace

#### debug_traceTransaction

- **Description**: Replays a transaction, returning the detailed execution trace.
- **Parameters**:
    - `String` - Transaction hash of the transaction to trace.
    - `Object` (optional) - Options to configure tracing properties like disabling memory or stack capturing for less
      verbose output.
- **Returns**: `Object` - Detailed information about the transaction execution, such as step-by-step state changes.

```bash
curl -X POST -H "Content-Type: application/json" --data '{"jsonrpc": "2.0", "method": "debug_traceTransaction", "params": ["0x14bd9cd554b725129a6c86f916490f52060644dc9627414bf9c62e1889130bf1", {"disableMemory": true, "disableStorage": false, "disableStack": false}], "id": 1}' https://rpc.test.mezo.org
```

#### debug_traceBlockByNumber

- **Description**: Replays all transactions in a block and returns detailed execution traces for each transaction.
- **Parameters**:
    - `String` - Block number (in hexadecimal) for the block to be traced.
    - `Object` (optional) - Options to control the verbosity of the trace, similar to individual transaction tracing.
- **Returns**: `Array` - An array of execution traces, each corresponding to a transaction within the block, detailing
step-by-step state changes.

```bash
curl -X POST -H "Content-Type: application/json" --data '{"jsonrpc": "2.0", "method": "debug_traceBlockByNumber", "params": ["0x29C7C", {"tracer": "callTracer"}], "id": 1}'  https://rpc.test.mezo.org
```

### `txpool` namespace

#### txpool_content

- **Description**: Returns detailed information about all transactions currently in the transaction pool.
- **Parameters**: None.
- **Returns**: `Object` - An object categorized by pending and queued transactions, including sender addresses and
detailed transaction data.

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"txpool_content","params":[],"id":1}' -H "Content-Type: application/json" https://rpc.test.mezo.org
```

#### txpool_inspect

- **Description**: Similar to `txpool_content`, but returns a more human-readable summary of the transactions in the
transaction pool.
- **Parameters**: None.
- **Returns**: `Object` - An object categorized by pending and queued transactions, showing a simplified listing of
transactions with sender addresses and brief details.

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"txpool_inspect","params":[],"id":1}' -H "Content-Type: application/json" https://rpc.test.mezo.org
```

#### txpool_status

- **Description**: Returns the status of the transaction pool, including the number of pending and queued transactions.
- **Parameters**: None.
- **Returns**: `Object` - An object with keys for `pending` and `queued`, indicating the total number of transactions in
each state.

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"txpool_status","params":[],"id":1}' -H "Content-Type: application/json" https://rpc.test.mezo.org
```

### `personal` namespace

The `personal` namespace is exposed, however it has been
[deprecated](https://geth.ethereum.org/docs/interacting-with-geth/rpc/ns-personal).
It will be removed in the future releases and any usage is discouraged.

### `mezo` namespace

The `mezo` namespace is a custom one that exposes additional methods.

#### mezo_estimateCost

- **Description**: Estimates the cost necessary to execute a transaction.
- **Parameters**:
    Object:
    - `from`: DATA, 20 Bytes - The address the transaction is send from.
    - `to`: DATA, 20 Bytes - (optional when creating new contract) The address the transaction is directed to.
    - `value`: QUANTITY - value sent with this transaction
- **Returns**:
    Object:
    - `decimals` - The decimals cost values are presented with.
    - `usdCost` - Estimated USD cost presented using `decimals`. Divide by `10^decimals`
                  to obtain the base USD value.
    - `btcCost` - Estimated BTC cost presented using `decimals`. Divide by `10^decimals`
                  to obtain the base BTC value.

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"mezo_estimateCost","params":[{"from":"0xFF3014B077D307E7B0bf262d072B25dbE19E2Be3","to":"0xd3CdA913deB6f67967B99D67aCDFa1712C293601","value":"0x186a0"}],"id":1}' -H "Content-Type: application/json" https://rpc.test.mezo.org
```
