# RFC-4: Native non-Bitcoin asset Bridge

## Background
[RFC-3: Bridging non-Bitcoin assets to Mezo](./rfc-3.md) was the first attempt 
to describe the mechanism to support the operation of bridging non-Bitcoin
assets to Mezo minimizing the fragmentation of liquidity . While all of the
concepts presented in RFC-3 are still technically correct, the RFC-3 proposal is 
based on the existence of a Bridging Partner ready to perform to-Mezo bridging 
on day one, and this assumption is quite risky in practice.

RFC-4 proposes an extension to the Bitcoin bridge mechanism described in
[RFC-2: Bridging Bitcoin to Mezo](./rfc-2.md) to support a small, selected group
of non-Bitcoin assets in the native bridge. The non-Bitcoin assets not supported
by the native bridge described in RFC-4 will be bridgable in the future, most
probably using a mechanism described in RFC-3.

## Proposal

The goal of the proposal is to perform the minimum necessary changes to the
existing RFC-2 Bitcoin bridging protocol, without adding too much overhead to
the validator, both in terms of the code and chain performance.

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

This RFC does not enforce any specific implementation choices in regards to how
to organize the contract code but we could potentially separate the common
bridge code like the `AssetsLocked` event and the sequence nonce, so that they
can be reused by the `BitcoinBridge` and `ERC20Bridge` contracts, both extended
by the `MezoBridge` contract. In this setup, some of the fields, events, and
errors in the `BitcoinBridge` contract will have to be renamed to clearly
indicate they are used for Bitcoin bridging.

Only a small selected set of ERC20 tokens should be accepted by the `MezoBridge`
contract. The contract should expose a set of functions for the governance to
add and remove the to-Mezo bridging support for selected ERC20s.

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