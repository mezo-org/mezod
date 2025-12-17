# RFC-4: Native non-Bitcoin asset Bridge

## Background

[RFC-3: Bridging non-Bitcoin assets to Mezo](./rfc-3.md) was the first attempt
to describe the mechanism to support the operation of bridging non-Bitcoin
assets to Mezo minimizing the fragmentation of liquidity. While all of the
concepts presented in RFC-3 are still technically correct, the RFC-3 proposal is
based on the existence of a Bridging Partner ready to perform to-Mezo bridging
on day one, and this assumption is quite risky in practice.

RFC-4 proposes an extension to the Bitcoin bridge mechanism described in
[RFC-2: Bridging Bitcoin to Mezo](./rfc-2.md) to support a small, selected group
of non-Bitcoin assets in the native bridge. The non-Bitcoin assets not supported
by the native bridge described in RFC-4 will be bridgeable in the future, most
probably using a mechanism described in RFC-3.

## Proposal

The goal of the proposal is to perform the minimum necessary changes to the
existing RFC-2 Bitcoin bridging protocol, without adding too much overhead to
the Mezo validator client, both in terms of the code and chain performance.

### MezoBridge contract

The Ethereum sidecar observes the `AssetsLocked` events emitted by the
`BitcoinBridge` contracts. Sufficiently confirmed events are processed during
the Extend Vote phase and broadcast as a vote extension.

The existing `AssetsLocked` event should be extended to include the address of
the token being bridged. The `tbtcAmount` field should be renamed to just
`amount` and all other fields should remain unchanged:

```
event AssetsLocked(
  uint256 indexed sequenceNumber,
  address indexed recipient,
  address indexed token,
  uint256 amount
);
```

The `MezoBridge` is a single contract deployed on Ethereum. It is the native
bridge contract controlling both Bitcoin (tBTC) and ERC20 bridging. All bridging
operations are sequenced using the same nonce. The existing `BitcoinBridge`
contract deployed on Sepolia is going to be replaced maintaining the continuity
of the nonce.

This RFC does not enforce any specific implementation choices in regards to how
to organize the contract code. One interesting option is to separate Bitcoin and
ERC20 bridging operations into two parent contracts for `MezoBridge`:
`BitcoinBridge` and `ERC20Bridge`. In this setup, some of the fields, events,
and errors in the `BitcoinBridge` contract will have to be renamed to clearly
indicate they are used for Bitcoin bridging.

Only a small selected set of ERC20 tokens should be accepted by the `MezoBridge`
contract. The contract should expose a set of functions for the governance to
add and remove the to-Mezo bridging support for selected ERC20s. There should be
a global limit of 20 tokens supported by the native bridge protecting the chain
and bridge performance in the event of the bridge governance getting compromised.
Ideally, the minimum bridgeable amount should be tracked for each token
separately as each token has a different value.

### Ethereum sidecar

No changes to the Ethereum sidecar are necessary other than those required to
reflect the changes in the `AssetsLocked` event and pull the events from the new
`MezoBridge` contract instead of the existing `BitcoinBridge` contract.

### x/bridge module

More changes will affect the x/bridge module. Since we are going to use the same
sequence nonce for Bitcoin and ERC20 bridging, the ABCI code should remain
mostly unchanged. The `AssetsLockedEvent`'s `Equal` function will have to be
extended to include the `token` check as nothing prevents a malicious validator
from voting on an event with the given sequence nonce but a different token. The
Bridge Keeper's `AcceptAssetsLocked` function should be extended to recognize
the token and map it to the right denominator when calling the Bank module
Keeper to mint coins. The address-to-denominator mapping can be initially
hardcoded in the client as the set of non-Bitcoin tokens supported by the bridge
will be minimal. Also, the particular entries once set, should never change.

### Token contracts

Each token supported by the native bridge will have its corresponding token
contract deployed on Mezo EVM with a minting authority delegated to the bridge
module's address. The pre-blocker, upon detecting the bridging request in the
vote extension pseudo-transaction, will prepare an internal EVM transaction
triggering the token mint. Such transactions will not incur any gas costs.

Token address mapping will be held by a maintenance precompile managed by the
governance. Before the token is allowlisted in the `MezoBridge` contract,
a mapping has to be added to the maintenance precompile. In case of a governance
failure to perform those operations in the right order and the mapping entry not
being present, the bridge should ignore the bridging request, and proceed as
usual. The funds locked in the `MezoBridge` contract will remain locked there
forever unless they are manually recovered. The initial implementation will
assume governance actions are executed in the right order and will not introduce
any mechanism for token recovery from the bridging contract.

This approach allows the governance to add new tokens to the bridge without
involving mezod development teams or performing chain forks. It also enables
adding custom logic to the token contract, depending on individual needs.
Enabling IBC will require introducing a dedicated token mapping mechanism.

The bridge maintenance precompile should enforce the same global limit of 20
tokens supported by the native bridge, as the `MezoBridge` contract. This limit
can be increased by the validator development team in the future, after thorough
consideration of the performance implications.

#### Alternative approach: token precompiles

An alternative approach is to use token precompiles. We rejected this approach
given the high-maintenance cost on the validator development team as well as the
network validators. Each new token added would require code change and validator
client update.

Each token supported by the bridge would have to be represented by a precompile.
The Bitcoin token precompile contains most of the logic we need, so the code
could be abstracted out making the introduction of new token precompiles as
easy as possible. The most common differences would be the denominator for the
Bank module, the token name, and the token decimals.

### EVM observability

The proposed approach requires small changes in the EVM observability mechanism
given the transactions injected by pre-blocker are not actual EVM transaction.
The existing observability mechanism should be reused and enhanced to inform
about the asset that got bridged.

### Supported Tokens

Based on the current knowledge, about 10 ERC-20 tokens should be bridgeable to
Mezo on day one. Each token can have separate conversion rules. For example,
USDC and USDT are converted to M but then bridged separately as mUSDC and mUSDT.
ETH is wrapped to stETH but bridged as mETH. All those conversions are out of
the scope of RFC-4 and should happen before the token is deposited in the
`MezoBridge` contract.

For the PoC implementation of the bridging protocol, we will assume the
following tokens should be represented on Mezo:

* mETH
* mUSDC
* mUSDT

The remaining tokens will be added later, once more confidence about them is
obtained.
