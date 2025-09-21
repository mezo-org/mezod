# RFC-5: Bridging assets out of Mezo

## Background

BTC and selected ERC20 assets are bridged from Ethereum (and Bitcoin) to Mezo using the 1-way native bridge
outlined in [RFC-2](./rfc-2.md) and [RFC-4](./rfc-4.md). This document describes the architecture of a solution
that allows bridging assets in the opposite direction - from Mezo to Ethereum and Bitcoin.

## Proposal

The goal of the proposal is to deliver the ability to bridge out quickly, while avoiding unnecessary complexity
and maintenance overhead. Existing components of the 1-way native bridge should be re-used wherever possible.

### `AssetsBridge` precompile on Mezo

The `AssetsBridge` precompile is an existing component that currently serves for bridge observability and
governance. We propose to make it the entry-point of the bridge out flow. To do so, the `AssetsBridge` precompile
should gain a new method named `bridgeOut`:

```solidity
function bridgeOut(
  address token, 
  uint256 amount, 
  uint8 chain, 
  bytes calldata recipient
) external returns (bool);
```

Parameters of this method are:

* `token`: address of the bridged out token on Mezo
* `amount`: amount to be bridged out in the token precision
* `chain`: Identifier of the target chain
* `recipient`: chain-specific recipient

#### Validation logic

The `bridgeOut` method should validate that:

* `chain` is either Ethereum or Bitcoin
* `token` is valid for the given `chain`:
    * Ethereum: `token` is either BTC or any ERC20 supported by the native bridge
    * Bitcoin: `token` is BTC
* `recipient` is valid for the given `chain`:
    * Ethereum: `recipient` is a 20-byte EVM address
    * Bitcoin: `recipient` is a proper standard-type Bitcoin script supported by tBTC, i.e. P2PKH, P2WPKH,
      P2SH or P2WSH
* `amount` can be spent by `AssetsBridge` precompile from the `msg.sender` account

#### Processing logic

If the validation passes, the `bridgeOut` method should:

* Burn the specified `amount` of `token` from the `msg.sender` account
    * BTC should be burned using the `x/bank` module directly. This fact should be propagated back to the EVM
      transaction execution context (similarly to [BTC transfers](https://github.com/mezo-org/mezod/blob/4b8925adccb84a5dd2a8ddc6c16d57bd91973e52/precompile/erc20/transfer.go#L212))
      and to the [BTC supply guard](https://github.com/mezo-org/mezod/blob/4b8925adccb84a5dd2a8ddc6c16d57bd91973e52/x/bridge/keeper/abci.go#L38)
    * ERC20 tokens should be burned using the `burnFrom` method exposed by the [mERC20](https://github.com/mezo-org/mezod/blob/v2.0.0/solidity/contracts/mERC20.sol)
      contract. This call should be done as an [internal EVM call](https://github.com/mezo-org/mezod/blob/4b8925adccb84a5dd2a8ddc6c16d57bd91973e52/x/evm/keeper/call.go#L19).
* Map the Mezo `token` address to the proper token address on the target `chain`:
    * If `token` is BTC, the target token should be TBTC on Ethereum, regardless of the value of the `chain`
      parameter. TBTC address is returned by the `getSourceBTCToken` method of the `AssetsBridge` precompile
    * Otherwise, the target token should be determined using bridge mappings returned by the
      `getERC20TokensMappings` method of the `AssetsBridge` precompile
* Create a new `AssetsUnlocked` storage entry in the `x/bridge` module. `AssetsUnlocked` should contain all
  input parameters of the given `bridgeOut` call and additionally:
    * Have an `unlockSequence` field allowing to uniquely identify the given `AssetsUnlocked` entry
    * Modify the `token` field to represent the token address on the **target** chain, as resolved from the mapping
* Emit appropriate EVM events

Changes described in this section should be covered with a comprehensive
[system test](https://github.com/mezo-org/mezod/tree/main/tests/system) suite.

### Ethereum sidecar

Each Mezo validator runs an instance of the Ethereum sidecar that feeds the existing 1-way bridge with events
emitted on the Ethereum chain. This proposal assumes re-using this infrastructure for bridging out of Mezo.

The Ethereum sidecar should be enhanced with an ability to fetch recent `AssetsUnlocked`entries submitted on the
Mezo chain and attest them on Ethereum. The specific logic described below is supposed to be executed only by
validators having the "bridge" privilege (the so called bridge validators). List of those validators can be
determined by calling the `validatorsByPrivilege` method (with `1` as parameter) exposed by the `ValidatorPool`
precompile.

#### Monitoring loop

The Ethereum sidecar should query its validator node privately to fetch recent `AssetsUnlocked` entries stored in
the `x/bridge` module. The query mechanism is an implementation detail and is not enforced by this document. A
possible approach is exposing `AssetsUnlocked` entries by the validator using EVM JSON-RPC or gRPC and periodically
polling them from within the sidecar monitoring loop. What is important, the sidecar should not query any third-party
node to avoid centralization and putting trust on external components.

#### Attestation process

Once `AssetsUnlocked`entries are fetched by the sidecar, all of them should be attested on Ethereum. The
attestation process requires access to the operator private key used by the corresponding validator. This is necessary
to attribute the given attestation to a specific bridge validator (as returned by the aforementioned
`validatorsByPrivilege` method of the `ValidatorPool` precompile). The integration of the private key is an
implementation detail but a possible solution may be using the keyring from within the sidecar process.

The sidecar attestation process should look as follows:

* Batch attestation phase:
    * The sidecar produces an ECDSA signature over the given `AssetsUnlocked` entry, using the aforementioned operator
      private key of the corresponding validator
    * The sidecar submits the ECDSA signature to the Bridge Worker using an HTTP call. The Bridge Worker is an
      off-chain component responsible for aggregating attestations and submitting them to the `MezoBridge` contract
      in one batch.
    * The sidecar monitors the progress of batch attestation for a specific time, defined by a new
      `batch_attestation_timeout` parameter
    * If the given `AssetsUnlocked`entry was properly attested by the Bridge Worker, the sidecar completes the
      attestation process for this entry. Otherwise, the sidecar falls back to individual attestation.
* Individual attestation phase:
    * The sidecar submits an attestation by issuing an Ethereum transaction against the `MezoBridge` contract directly.
      This step uses the operator private key to set the `msg.sender` of the issued transaction and does not require
      attaching the aforementioned ECDSA signature.
    * The sidecar confirms finality of the above Ethereum transaction and completes the attestation process for the
      given `AssetsUnlocked`entry

Such a phased attestation process has the following goals:

* The batch attestation phase reduces gas costs on Ethereum by using a centralized component that can aggregate
  signatures and submit them in a single transaction at once. However, this comes at a cost of an increased
  centralization risk.
* The individual attestation phase is a fallback allowing to mitigate the centralization problems of the first phase
  if needed. This guarantees attestations are eventually delivered, despite of the increased costs.

Moreover, the sidecar attestation process should have mechanisms allowing to cache processing data and survive
sidecar restarts. The process must guarantee that all `AssetsUnlocked` entries are processed in a reasonable time.

### `MezoBridge` contract on Ethereum

The existing [`MezoBridge`](https://github.com/thesis/mezo-portal/blob/main/solidity/contracts/MezoBridge.sol)
contract should expose a new API allowing to obtain attestations and withdraw assets.

#### Individual attestation

The `MezoBridge` contract should expose a new `attestBridgeOut`function allowing for a direct submission of a
single attestation over an `AssetsUnlocked`entry:

```
function attestBridgeOut(AssetsUnlocked calldata entry) external;
```

This function should validate that the `msg.sender` is a bridge validator and record the attestation for the given
`AssetsUnlocked` entry in the contract storage. Once the given entry meets the necessary attestation threshold of 2/3+
of the bridge validators, the `MezoBridge` contract can actually withdraw its underlying assets.

#### Batch attestation

The `MezoBridge` contract should also support batch attestations with ECDSA signatures:

```
function attestBridgeOutWithSignatures(
  AssetsUnlocked calldata entry, 
  bytes calldata signatures
) external;
```

This function should not check `msg.sender` but instead:

* Validate and parse the `signatures` vector and extract individual signatures
* Make sure the given signature matches the `entry`
* Make sure the given signature comes from a bridge validator
* Make sure the number of valid signatures meet the attestation threshold of 2/3+ of the bridge validators

If all conditions are met, the `MezoBridge` contract can actually withdraw the underlying assets of the given
`AssetsUnlocked` entry.

#### Assets withdrawal

All ERC20 assets being under control of the `MezoBridge` contract can be unlocked by simply issuing a `transfer`
call based on the `token` and `recipient` information from the given `AssetsUnlocked` entry.

However, if the `AssetsUnlocked`'s `token` is TBTC and `chain` is Bitcoin, the `MezoBridge` must request a
redemption in the tBTC bridge so the `recipient` obtains real BTC on the Bitcoin chain. This process is more
complicated as the tBTC bridge requires two additional pieces of data to request a redemption (comparing to what is
available in a single `AssetsUnlocked` entry):

* The 20-byte public key hash of the wallet supposed to handle the redemption
* The current main UTXO of the target wallet

To handle this complexity, this proposal assumes that `MezoBridge` exposes an additional method allowing to deliver
the missing data to attested `AssetsUnlocked` entries requiring tBTC to unlock funds on the Bitcoin chain:

```
function withdrawBTC(
  AssetsUnlocked calldata entry, 
  bytes20 walletPubKeyHash, 
  BitcoinTx.UTXO memory mainUtxo
) external;
```

The `walletPubKeyHash` and `mainUtxo` arguments are tBTC-specific and are expected by the
[`requestRedemption`](https://github.com/threshold-network/tbtc-v2/blob/82ba7ac1460baa9286aa211621bc1c5914ddb242/solidity/contracts/bridge/Redemption.sol#L437)
function defined in the tBTC Bridge contract.

The `withdrawBTC` method should be callable by anyone but only for `AssetsUnlocked`entries that met the attestation
threshold of 2/3+ of the bridge validators and whose `chain` is Bitcoin. It should issue a redemption request to the
tBTC bridge using the `approveAndCall` mechanism of the TBTC token (see [example from Acre contracts](https://github.com/thesis/acre/blob/2078d339f4ddce79e69c78529c9d72910dd2640a/solidity/contracts/BitcoinRedeemer.sol#L175)).
The `walletPubKeyHash` and `mainUtxo` arguments are validated as part of the `approveAndCall` so the transaction may
revert if they are wrong.

Note that the permission-less design of `withdrawBTC` makes this method prone to front-running vectors. The
validation done within this method and on the tBTC side should not allow for any negative impact on users. However,
a race between honest callers of `withdrawBTC` may lead to a gas loss for those of them that were surpassed.

#### Bridge validators management

The only source of truth about bridge validators is the `ValidatorPool` precompile on Mezo. There is no easy way to
port that automatically to the `MezoBridge` contract on Ethereum. The `MezoBridge` should allow setting and removing
bridge validators and the governance is responsible for maintaining parity between both chains.

#### Reimbursements

The `MezoBridge` contract should be integrated with a mechanism allowing to reimburse:

* `attestBridgeOut` calls made by bridge validators
* `attestBridgeOutWithSignatures` calls made by approved maintainers (e.g. Bridge Worker)
* `withdrawBTC`calls made by approved maintainers (e.g. Bridge Worker)

Consider using a [`ReimbursementPool`](https://github.com/threshold-network/keep-core/blob/406cb8954005484665422c9693dc780a7ca2d425/solidity/random-beacon/contracts/ReimbursementPool.sol)
contract, similar to the one used by tBTC.

#### Fees

The `MezoBridge` contract should cut a fee from withdrawn assets. The fee should be governable and set per-asset.
All fees should be sent to a `feeCollector` account set by the governance. Part of those fees can be used to fund
reimbursements for eligible operations.

### Bridge Worker

The Bridge Worker is an off-chain component that optimizes the attestation process and provides auxiliary services
facilitating the bridge out process. This document does not enforce any specific way to implement it. Consider using
Cloudflare Workers to fit into the existing infrastructure of Mezo off-chain components. Remember about adding
appropriate DoS protections to any public endpoints exposed by this worker.

#### Endpoints

This worker should expose a HTTP endpoint that allows submitting an `AssetsUnlocked` entry along with an attestation
signature made by a bridge validator. This endpoint should validate the incoming data to filter out spam and store
proper attestations internally for further processing. At a minimum, this endpoint should check if:

* The incoming `AssetsUnlocked`is a valid entry submitted on Mezo's `AssetsBridge` precompile
* The `AssetsUnlocked` entry was not yet processed in the `MezoBridge` on Ethereum
* The attached ECDSA signature is valid, i.e. signs the associated `AssetsUnlocked` entry and is made by a bridge
  validator

#### Background jobs

The Bridge Worker should periodically run a background job that issues a `attestBridgeOutWithSignatures` transaction
against the `MezoBridge` contract for entries that meet the attestation threshold of 2/3+ of the bridge validators.

Moreover, the Bridge Worker can also take care of calling `withdrawBTC` on `MezoBridge` for bridge outs to Bitcoin
(i.e. for `AssetsUnlocked`entries whose `chain` is Bitcoin).

## Additional considerations

* The `bridgeOut` function in the `AssetsBridge` could be `payable` to allow direct bridge out of BTC. Although this
  seems tempting, this adds additional complexity as this path must be handled separately. This document proposes to
  not include it right now and consider adding it in the future if there is an explicit need for that.
* The `AssetsBridge` precompile could expose a `bridgeOutWithPermit` function allowing to request the bridge out with
  a single transaction and spare the additional approval step. Adding this function should be considered at the end
  of the implementation to not suppress the momentum at the beginning. The goal is to make the flow working as quickly
  as possible and then adding nice-to-haves.
* Bridge out cancellation is beyond the scope of this document. What happens if the bridge out action does not
  complete in a reasonable time should be considered separately.
* This document does not outline any mechanism that automatically moves the bridge out fees from the `feeCollector`
  account to the reimbursement mechanism
* The Bridge Worker is an abstract concept introduced for the sake of this document. It can be implemented as part of
  an existing off-chain component serving the Mezo bridge, for example, [`tbtc-api`](https://github.com/thesis/mezo-portal/tree/main/tbtc-api)
