# RFC-3: Bridging non-Bitcoin assets to Mezo

>[!WARNING]
> This document was superseded by [RFC-4](./rfc-4.md).

## Background

[RFC-2: Bridging Bitcoin to Mezo](./rfc-2.md) describes the mechanism of
bridging Bitcoin represented on EVM as tBTC to Mezo chain. Bitcoin is the base
asset of Mezo and requires a separate bridging path that may in the future be
transformed into direct tBTC minting on Mezo by moving the tBTC Bridge ledger
there.

The RFC-3 focuses on bridging non-tBTC assets from Ethereum to Mezo. The
mechanism described in this RFC is going to be used in the native
Ethereum-to-Mezo bridge. An important requirement for the proposal is to
minimize the liquidity fragmentation for major assets by establishing clear
canonical token addresses.

## Proposal

### Canonical Token

The bridging itself may not require any work from the Mezo development team.
Various bridging solutions such as Wormhole, LayerZero, or Axelar exist
on the market. The development teams of bridges could integrate with Mezo chain
and enable minting bridged tokens. The problem with this out-of-the-box solution
is liquidity fragmentation given each bridge mints its own representation of
the token. As an example, we could have three or more representations of USDC on
Mezo, all with different addresses.

This proposal addresses the liquidity fragmentation problem by introducing
a canonical ERC20 token contract. The canonical token implementation should be
modeled after [tBTC `L2TBTC` token](https://github.com/keep-network/tbtc-v2/blob/main/solidity/contracts/l2/L2TBTC.sol)
and be flexible enough to:

- Delegate restricted minting authority to a short list of bridging gateways.
- Be paused by any one of n guardians, allowing avoidance of contagion in case
  of a gateway-specific incident.
- Initially be governed by the same body as the ValidatorPool precompile.

### Bridging Partners

The canonical tokens will be mintable by bridging gateways integrated with
specific bridging partners. An example of a bridging partner integration is
tBTC v2 integration with Wormhole to mint canonical tBTC address on Arbitrum,
Optimism, Polygon, Base, and Solana. Instead of minting Wormhole wrapped tokens,
the canonical representations of tBTC are minted by Wormhole L2 gateway contract
on each mentioned chain.

For the Mezo chain launch, we will integrate with at least one well-established
bridging partner to bridge all tokens but Bitcoin to Mezo. Bitcoin bridging will
remain controlled exclusively by our native bridge.

The first bridging partner we will integrate with is Wormhole but we should add
more bridging partners over time.

### Bridge Gateway

A bridge gateway is a smart contract specific to the given bridging partner
implementation. The bridge gateway is added to the canonical token as one of the
minters. A canonical token can have one or multiple bridge gateways authorized
to mint. There should be a separate gateway deployed for each canonical token
minted. The gateway receives a wrapped token minted by the bridging partner and
mints the same amount of canonical token to the user.

Gateways can be added and removed over time. Special care will have to be taken
in the UI to select the gateway that has enough wrapped token locked to serve
the bridge-back request. Some large bridge-back operations may need to be split
into several transactions.

Each gateway should be deployed as an upgradeable contract, controlled by the
same entity as the ValidatorPool precompile.

A battle-proven implementation of a Wormhole-specific gateway is available in
tBTC repository as [`L2WormholeGateway`](https://github.com/keep-network/tbtc-v2/blob/main/solidity/contracts/l2/L2WormholeGateway.sol)
and can be reused for Mezo.

The RFC recommends keeping the gateway implementations in the mezo-portal
repository, next to the Bitcoin Bridge contract described in RFC-2.

### Architecture Graph

The diagram below presents an example architecture for three canonical tokens
deployed: USDC, DAI, and USDT, each with its own minter attached.
The bridge gateways serving as minters are implemented for the Wormhole Bridge.
Wormhole deposits wrapped tokens into a specific gateway that is capable of
minting canonical token representation.

```
                                         +-------------------------------------------------------------------------------+
                                         |                                      Mezo                                     |
                                         |                                                                               |
                                         |                                   +---------------------+  +--------------+   |
+----------------------------+           |                               +---| USDCWormholeGateway |--|     USDC     |   |
|          Ethereum          |           |                               |   +---------------------+  +--------------+   |
|                            |           |                               |                                               |
|  +----------------------+  |           |  +----------------------+     |   +----------------------+  +-------------+   |
|  | Wormhole TokenBridge |--|-----------|--| Wormhole TokenBridge |-----|---|  DaiWormholeGateway  |--|     DAI     |   |
|  +----------------------+  |           |  +----------------------+     |   +----------------------+  +-------------+   |
|                            |           |                               |                                               |
+----------------------------+           |                               |   +---------------------+  +--------------+   |
                                         |                               +---| USDTWormholeGateway |--|     USDT     |   |
                                         |                                   +---------------------+  +--------------+   |   
                                         |                                                                               |
                                         +-------------------------------------------------------------------------------+
```

### Canonical Token Factory

All canonical tokens and all partner-specific gateways should follow the same
implementation. We can facilitate the process of registering new canonical
tokens utilizing a proxy factory for gateways and canonical tokens, all pointing
to the same implementation. An audited implementation of a proxy and mechanism
for deploying such proxies could be borrowed from the
[OrangeKit Safe factory](https://github.com/thesis/orangekit/blob/4285a3dbacf944a43914a182ba5313ce818416b8/solidity/contracts/OrangeKitSafeFactory.sol#L270-L314).
The canonical token factory should be owned by the same entity as the one owning
ValidatorPool and provide an on-chain mapping between Ethereum and Mezo token
addresses.

### Mezo Portal UI

The UI abstracting out the complexity of the bridging operations needs to be
integrated into the Mezo Portal. The complexity of integrating the bridging
partner with our custom UI should be one of the deciding factors when selecting
bridging partners. Wormhole provides a [TypeScript SDK](https://wormhole.com/products/sdk)
and [Wormhole Connect](https://docs.wormhole.com/wormhole/wormhole-connect/overview)
to integrate with their bridge. When bridging to Mezo from Ethereum, users would
interact with one of the bridging partner's infrastructure but this interaction
would be hidden under the hood and the users would only have to interact with
the Mezo Portal.

### Relaying post-bridge redemptions

One common usability issue when bridging tokens to another chain is the
necessity of redeeming them on the target chain once the bridging operation is
completed. In the case of Mezo, redemption requires calling the receive function
of the gateway to redeem wrapped tokens from the bridge and mint canonical tokens
on Mezo. This requires users to already have Bitcoin on Mezo and remember to
execute this operation once the bridging is completed.

This additional complexity can be avoided by utilizing relayers. Wormhole
provides a standard relayer, we were already able to
[integrate successfully](https://github.com/keep-network/tbtc-v2/blob/e0a0bd46d783b616805815fc9840d5cc09fe79fd/solidity/contracts/l2/L1BitcoinDepositor.sol#L651-L660)
for the L2 tBTC bridging. The standard relayer should be integrated with the
ERC20 Mezo Bridge allowing users to pay for Mezo token redemption when
initiating a bridge operation on Ethereum.

This will require having the bridge gateways deployed on Mezo chain to support
the [`IWormholeReceiver`](https://github.com/wormhole-foundation/wormhole-solidity-sdk/blob/bacbe82e6ae3f7f5ec7cdcd7d480f1e528471bbb/src/interfaces/IWormholeReceiver.sol#L44-L50)
interface.

### Bridging tokens locked in the Portal

For existing deposits locked in the Portal contract, we should provide a path to
bridge them to the same EVM address on Mezo as the `depositOwner` stored in
Portal. This path should be added with an upgrade to the Portal contract
deployed on Ethereum.

Note that not all tokens locked in the Portal will be bridgable and some of them
will have to be transformed to other assets before bridging. This particular
mechanism may require cooperation with a market maker and is out of the scope of
this RFC. The ERC20 bridge, as designed in this RFC, should be flexible enough
to accommodate any ERC20 bridging needs coming from the Portal contract.

The list of tokens to be bridged will be provided in a separate document. For
all of them, we should establish canonical token addresses.

Tokens locked in the Portal contract will not be automatically bridged. User's
action will be required to initiate bridging to Mezo.

## Open Questions

### Ethereum ERC20 Bridge contract

After the Mezo chain launch users will no longer be expected to lock their
assets in the Portal for a specific time but instead to bridge them to the Mezo
chain. Since a separate mechanism handles Bitcoin bridging, no Ethereum contract
for ERC20 bridging should be necessary if integrating standard relayer calls
will be possible using the Wormhole SDK. If this proves not to be the case
during the implementation, we should be able to reuse
[the same code](https://github.com/keep-network/tbtc-v2/blob/e0a0bd46d783b616805815fc9840d5cc09fe79fd/solidity/contracts/l2/L1BitcoinDepositor.sol#L600)
for transferring the token to the Wormhole bridge and initiating relaying as for
the tBTC bridge. However, avoiding another on-chain building block is preferred.
