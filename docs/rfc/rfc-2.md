# RFC-2: Bridging Bitcoin to Mezo

## Background

Mezo chain uses Bitcoin as a base token. The EVM-compatible version of Bitcoin
is tBTC and the tBTC Bridge will be used to bring Bitcoin to Mezo. Currently,
the tBTC Bridge ledger exists on Ethereum, so Ethereum-to-Mezo tBTC bridging
has to be achieved in the first release. The bridging mechanism has to be
compatible with the Bitcoin token precompile on Mezo and reflect the balances
in the `x/bank` module.

This RFC describes bridging tBTC from Ethereum to Mezo. Bridging back and
bridging other assets is out of the scope and will be covered in separate RFC
documents. Also, the potential future work of moving the tBTC Bridge ledger
from Ethereum to Mezo is out of the scope of this RFC.

## Proposal

### Bridge validators

Only a subset of Mezo validators participates in bridging and they have an equal
vote in deciding on bridging the asset. All those validators are expected to run
full Ethereum nodes. Having all validators in the network participate in
bridging does not make it any more secure if most of them use the same Ethereum
JSON-RPC endpoint. Additionally, for the Mezo-to-Ethereum bridge, we will reuse
the same subset of validators to generate a signature to unlock assets on
Ethereum. The governance will appoint the bridge validators using an EVM bridge
precompile. The list of addresses participating in bridging has to be known to
all other validators in the network to achieve consensus about bridging
decisions.

### Ethereum BitcoinBridge contract

Mezo Portal contract on Ethereum accepts ERC20 deposits and accumulates TVL for
the network before the chain launches. Adding bridging logic to the Portal
contract is tempting but this contract has recently incorporated more logic,
such as stBTC token minting and support for the liquidity-treasury-managed
assets. We need a separate contract for the bridge with a clear separation of
concerns - tBTC locked in the BitcoinBridge contract on Ethereum will be
reflected 1:1 on the Mezo chain and locked in the contract until not bridged
back by the user for to-Bitcoin redemption via tBTC Bridge.

The BitcoinBridge contract should be upgradeable by the governance to allow
incorporating the logic for bridging back in the future. The contract should
extend tBTC `AbstractTBTCDepositor` contract to allow direct from-Bitcoin
bridging. It should expose four external functions at a minimum:

```
/// @notice Transfer and locks the `amount` of tBTC in the contract and emits
///         `AssetsLocked` event to initiate bridging to Mezo to the `recipient`
///          address.
function bridge(uint256 amount, address recipient) external

/// @notice The same as the `bridge` function except that with EIP-2612 permit
///         instead of an on-chain approval.
function bridgeWithPermit(
    uint256 amount, 
    address recipient, 
    address owner, 
    uint256 deadline, 
    uint8 v, 
    bytes32 r, 
    bytes32 s
) 

/// @notice Registers tBTC Bitcoin deposit in the contract and reveals it to the
///         tBTC Bitcoin bridge. The `recipient` address should be encoded in 
///         the `extraData` section of the tBTC P2WSH deposit script.
function initializeDeposit(
    IBridgeTypes.BitcoinTxInfo calldata fundingTx, 
    IBridgeTypes.DepositRevealInfo calldata reveal
    address recipient
)

/// @notice Finalizes tBTC Bitcoin deposit in the tBTC Bitcoin bridge. Locks
///         the deposit in the contract and emits `AssetsLocked` event to
///         initiate bridging to Mezo to the `recipient` address.
function finalizeDeposit(uint256 depositKey, address recipient)
```

The RFC suggests placing the BitcoinBridge contract in the `thesis/mezo-portal`
monorepo, next to the Portal contract as the Portal dApp will be reworked to
expose bridging functionality.

A good example of a contract extending tBTC `BitcoinDepositor` and implementing
`initializeDeposit` and `finalizeDeposit` function is the Mezo Portal contract.
In contrast to the Mezo Portal contract, the BitcoinBridge contract should
extend the `AbstractTBTCDepositor` directly. The core functionality of the
BitcoinBridge contract is bridging native Bitcoin to Mezo and this fact should
be reflected in how the contract code is organized.

After the chain launch, the Portal contract will no longer accept deposits and
all new deposits will be bridged automatically to the Mezo chain through the
BitcoinBridge contract. Hence, the BitcoinBridge contract is what the Ethereum
sidecar should observe for events.

### Ethereum sidecar

Bridge validators need to be aware of the state of Ethereum. This will be
achieved by implementing a sidecar observing the Ethereum Mezo Bridge contract.
The sidecar may be embedded into the Mezo validator process, or run as a
separate one. Each of those two choices has its advantages. Keeping the sidecar
embedded in the validator process makes the operational work easier. Keeping the
sidecar separate makes the experience consistent with the Skip protocol sidecar
we want to integrate as a price oracle, and better prepares us for future
generations of sidecars, as described in the Future Work section. This option
also allows to incorporate an additional bridge validator logic such as Schnorr
or tECDSA key and signature generation that may not be straightforward to
implement in the validator given no support for sending arbitrary network
messages.

The RFC does not enforce a specific choice but implementing the sidecar as
a separate process in the same binary as the validator seems to be the most
future-proof approach:

- We keep the code in one place so the management is easier.
- We can reuse code so we avoid duplication of the boilerplate.
- We keep the flexibility as the sidecar running in a different process does not
  affect validator performance. Running auxiliary communication for Schnorr is
  totally separated from the consensus engine.
- This is safer. A panic in the sidecar does not kill the validator process.
- We do not close any paths for the future. If there is a need to run the
  sidecar as part of the validator process, this is just about hiding the
  sidecar CLI and running its logic along with the validator CLI command. If we
  want to extract the code to a separate binary, validators will be used to
  run a separate process.

The sidecar must expose a gRPC API returning information about confirmed
`AssetsLocked` events in the Ethereum Mezo Bridge contract.

To understand ETH2 finality, an understanding of checkpoints and epochs is
required. Each epoch has 32 slots and each slot takes 12 seconds. The checkpoint
is the first slot of an epoch. If the checkpoint at epoch E gathered 2/3
supermajority, the blocks at epoch E-1 are considered justified and the blocks
at epoch E-2 are considered finalized. The sidecar must only inform about
`AssetsLocked` events from the finalized blocks. It means it will take about 13
minutes to notify about the state change on Ethereum.

The sidecar should use periodical `eth_getLogs` calls to the Ethereum node and
cache information locally. To keep the bridging process efficient and not slow
the consensus algorithm, information about finalized `AssetsLocked` events need
to be available instantly, without the need for a call to the Ethereum node as
a part of serving the request from the validator.

### x/bridge

The `x/bridge` module should be our custom Cosmos module where all the bridging
state changes logic and the bridge Keeper should be located. Also, in the
initial version, this is the module that should interact with `x/bank` to mint
tokens for the users who bridged their Bitcoin to Mezo.

The EVM observability of bridge events should be implemented once the initial
version of the bridging mechanism works. It is possible to launch the chain
without bridging observability from the EVM state level but not being able to
see bridging events from the EVM block explorer would be a huge user experience
gap. Adding the observability most probably requires creating zero-fee EVM
transaction proposals by the block proposer.

### Consensus

To achieve the consensus about the assets being bridged to the Mezo chain, we
are going to utilize ABCI++ and vote extensions. The mechanism allows an
application to extend a pre-commit vote with arbitrary data. In our case, the
arbitrary data will be information read from the Ethereum sidecar about
`AssetsLocked` events. The mechanism is based on four handlers:
`sdk.ExtendVoteHandler`, `sdk.VerifyVoteExtensionHandler`,
`sdk.PrepareProposalHandler`,`sdk.ProcessProposalHandler`, and `PreBlocker` that
is run before any other code during the `FinalizeBlock` phase to perform the
state changes. The execution of the whole logic spans two blocks. In the first
block, data from the sidecar is gathered, published, and validated as a vote
extension. In the next block, vote extensions are consumed by the block
proposer, validated by the rest of the network, and the state is updated based
on the vote extensions accepted.

Cosmos SDK documentation contains a very good
[tutorial](https://docs.cosmos.network/v0.50/tutorials/vote-extensions/oracle/implementing-vote-extensions)
about implementing the vote extensions. An inspiration for utilizing the vote
extensions mechanism can also be taken from the
[Slinky project](https://github.com/skip-mev/slinky/).

#### Block N: Extend Vote

In the Extend Vote phase, each bridge validator queries its sidecar to retrieve
the list of `AssetsLocked` events. The events should be sorted and the events
already handled should be filtered out. `AssetsLocked` events should be
serialized and broadcast as a vote extension. This step is non-deterministic and
it is acceptable some bridge validators may retrieve other `AssetsLocked` events
than the rest of the bridge validators.

#### Block N: Verify Vote Extension

In the Verify Vote phase, vote extensions published earlier are validated by all
validators in the network. This validation has to be deterministic. At a minimum,
the unmarshaling and the fact the vote extension comes from one of the bridge
validators should be checked. Vote extensions not coming from the bridge
validators appointed in the bridge module should be rejected.

#### Block N+1: Prepare Proposal

The block proposer appointed by the consensus algorithm processes all vote
extensions from the previous block to prepare one, final proposal based on them.
In the first step, the prepare proposal handler first checks if the vote
extensions’ signatures are correct using the `ValidateVoteExtensions` helper
function from the `baseapp` package. In the second step, the prepare proposal
handler checks if each vote extension comes from a bridge validator in case the
bridge validator was removed from the set in the meantime. If all checks pass,
the block proposer takes all `AssetsLocked` events that were voted by at least
2/3 of bridge validators and aggregates them together into a pseudo-transaction
on the top of the block proposed. This pseudo-transaction should be treated just
as metadata.

#### Block N+1: Process Proposal

This step is similar to the Prepare Proposal step, except that it is executed by
all validators in the network, based on the pseudo-transaction injected by the
block proposer. The validators need to validate the vote extensions in the
pseudo-transaction in the same way as the block proposer: check if the vote
extension is supported by 2/3 of bridge validators, all signatures are valid,
and they actually come from bridge validators appointed in the bridge module.

#### Block N+1: Finalize Block

The `PreBlocker` logic is run before any other code during the Finalize Block
phase. This is where we make use of the validated vote extensions from the block
proposer and update the chain state based on them. The `PreBlocker` should mint
Bitcoin tokens using the x/Bank module to the addresses appointed in the
`AssetsLocked` events.

### Chain launch

On the chain launch, all tokens that are locked in the Portal contract on
Ethereum should be bridged automatically to the Mezo chain. In practice, it
means disabling new deposits in the Portal contract and moving tBTC from all
non-stBTC-ed deposits from Portal to BitcoinBridge for bridging. This is
a complicated operation requiring cooperation with a market maker to unwrap
deposited WBTC to Bitcoin, mint tBTC, and put it back into the Portal contract.
Special-casing stBTC-ed deposits require more attention as well. The exact
description of the transition will be covered by a separate RFC.

On a high leve, upon disabling deposits and withdrawals in the Portal contract,
a special script should iterate towards Portal's `DepositInfo` structures,
filter deposits for tBTC token for which stBTC was not minted, and generate
a module genesis JSON to set Bitcoin balances on Mezo chain for each depositor.
`DepositInfo` structures can be retrieved by scanning `Deposited` events in the
Portal contract. Note that the depositor address could be an Ethereum wallet
address or OrangeKit's EVM address derived from a Bitcoin wallet but from the
bridge's perspective there is no difference.

## Future work

### Light ETH2 client

The security model of the proposed solution relies on the fact bridge validators
run full Ethereum nodes independently instead of using a single Ethereum
JSON-RPC provider. We can take this solution one step forward and integrate the
Ethereum light client into the sidecar, possibly enabling all network validators
to participate in Ethereum-to-Mezo bridging. In this solution, the light client
would provide a trusted state root from the header via the EIP-1186
`eth_getProof` endpoint and we would perform a
[validation](https://github.com/ethereum/EIPs/issues/1186#issuecomment-401161169)
a bridging transaction happened on Ethereum. In this model, all light clients
embedded in sidecars validate Ethereum blocks independently and the
centralization risk is minimal, as long as they all chose valid checkpoints for
the light client initialization. According to Ethereum documentation as of
March 2024, none of the ETH2 light clients implementation are considered
production-ready.

# References

- [Cosmos SDK ABCI++ Vote Extensions documentation](https://docs.cosmos.network/main/build/abci/vote-extensions)
- [Cosmos SDK v0.50 Vote Extensions tutorial](https://docs.cosmos.network/v0.50/tutorials/vote-extensions/oracle/implementing-vote-extensions)

- [tBTC `AbstractTBTCDepositor` contract](https://github.com/keep-network/tbtc-v2/blob/main/solidity/contracts/integrator/AbstractTBTCDepositor.sol)
- [Ethereum Light Clients documentation](https://ethereum.org/en/developers/docs/nodes-and-clients/light-clients/)
