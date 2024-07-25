# RFC-3: Bridging non-Bitcoin assets to Mezo

## Background

[RFC-2: Bridging Bitcoin to Mezo](./rfc-2.md) describes the mechanism of
bridging Bitcoin represented on EVM as tBTC to Mezo chain. Bitcoin is the base
asset of Mezo and requires a separate bridging path that may in the future be
transformed into direct tBTC minting on Mezo by moving the tBTC Bridge ledger
there.

The RFC-3 focuses on bridging non-tBTC assets from Ethereum to Mezo.
Specifically, it aims to enable bridging of all non-Bitcoin assets locked
currently in the Portal contract on Ethereum, such as USDC, USDT, thUSD, crvUSD,
DAI, and others.

An important requirement for the proposal is to minimize the liquidity
fragmentation for major assets by establishing clear canonical token addresses.

## Proposal

### Canonical Token

The bridging itself may not require any work from the Mezo development team.
Various bridging solutions such as Wormhole, LayerZero, or Axelar are available
on the market. The development teams of the bridges could integrate with Mezo
and enable minting bridged tokens on the chain. The problem with this
out-of-the-box solution is liquidity fragmentation given each bridge will mint
its own representation of the token. As an example, we can have three or more
representations of USDC on Mezo, all with different addresses.

This proposal addresses the liquidity fragmentation problem introducing
a canonical ERC20 token precompile available under the same Mezo EVM address as
on Ethereum mainnet. At a minimum, we want canonical token precompiles to be
available for all ERC20 non-Bitcoin tokens supported currently in the Portal
contract on Ethereum. As of the time of writing this RFC, those tokens are:

| Token    | Address                                    |
| -------- | ------------------------------------------ |
| USDC     | 0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48 |
| USDT     | 0xdac17f958d2ee523a2206206994597c13d831ec7 |
| thUSD    | 0xCFC5bD99915aAa815401C5a41A927aB7a38d29cf |
| USDe     | 0x4c9EDD5852cd905f086C759E8383e09bff1E68B3 |
| crvUSD   | 0xf939E0A03FB07F59A73314E73794Be0E57ac1b4E |
| DAI      | 0x6B175474E89094C44Da98b954EedeAC495271d0F |

The canonical token precompile implementation should be modeled after
[tBTC `L2TBTC` token](https://github.com/keep-network/tbtc-v2/blob/main/solidity/contracts/l2/L2TBTC.sol)
and be flexible enough to:

- Delegate restricted minting authority to a short list of bridging gateways.
- Be paused by any one of n guardians, allowing avoidance of contagion in case
  of a gateway-specific incident.
- Initially be governed by the same body as the ValidatorPool precompile.

### Bridging Partners

The canonical tokens will be mintable by bridging gateways integrated with
specific bridging partners. An example of a bridging partner integration is
tBTC v2 integration with Wormhole to mint canonical tBTC address on Abirtrum,
Optimism, Polygon, Base, and Solana. Instead of minting Wormhole wrapped tokens,
the canonical representations of tBTC are minted by Wormhole L2 gateway contract
on each mentioned chain.

For the Mezo chain launch, we will integrate with at least one well-established
bridging partner to bridge all tokens but Bitcoin to Mezo. Bitcoin bridging will
remain controlled exclusively by our native bridge.

### Bridge Gateway

A bridge gateway is a smart contract specific to the given bridging partner
implementation. The bridge gateway is added to the canonical token as one of the
minters. A canonical token can have one or multiple bridge gateways authorized
to mint. The bridge gateway address is not sacrosanct so it does not have to be
a precompile. Having the gateway implemented as a precompile could lower the
operational cost of bridging but is not a requirement for the launch. There
should be a separate gateway deployed for each token minted. The gateway
receives a wrapped token minted by the bridging partner and mints the same
amount of canonical token to the user.

Gateways can be added and removed over time. Special care will have to be taken
in the UI to select the gateway that has enough wrapped token locked to serve
the bridge-back request. Some bridge-back operations may need to be split
between several transactions.

Each gateway should be deployed as an upgradeable contract, controlled by the
same body as the ValidatorPool precompile.

A battle-proven implementation of a Wormhole-specific gateway is available in
tBTC repository as [`L2WormholeGateway`](https://github.com/keep-network/tbtc-v2/blob/main/solidity/contracts/l2/L2WormholeGateway.sol)
and can be reused for Mezo.

The RFC recommends keeping the gateway implementations in the mezo-portal
repository, next to the Bitcoin Bridge contract described in RFC-2.

### Architecture Graph

The diagram below presents an example architecture for three canonical tokens
deployed. USDC, thUSD, and USDT are precompiles with three minters attach. The
bridge gateways serving as minters are implemented for the Wormhole Bridge.
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
|  | Wormhole TokenBridge |--|-----------|--| Wormhole TokenBridge |-----|---| ThUSDWormholeGateway |--|    thUSD    |   |
|  +----------------------+  |           |  +----------------------+     |   +----------------------+  +-------------+   |
|                            |           |                               |                                               |
+----------------------------+           |                               |   +---------------------+  +--------------+   |
                                         |                               +---| USDTWormholeGateway |--|     USDT     |   |
                                         |                                   +---------------------+  +--------------+   |   
                                         |                                                                               |
                                         +-------------------------------------------------------------------------------+
```

### Mezo Portal UI

The UI abstracting out the complexity of the bridging operations needs to be
integrated into the Mezo Portal. The complexity of integrating the bridging
partner with our custom UI should be one of the deciding factors when choosing
a bridging partner. Wormhole provides a [TypeScript SDK](https://wormhole.com/products/sdk)
to integrate with their bridge, for example. When bridging to Mezo from
Ethereum, users would interact with one of the bridging partner's infrastructure
directly but this interaction would be hidden under the hood and the users would
only have to interact with the Mezo Portal.

For existing deposits locked in the Portal contract, we should provide a path to
bridge them to the same EVM address on Mezo as the `depositOwner` stored in
Portal. This path should be added with an upgrade to the Portal contract
deployed on Ethereum.

## Open Questions

### Bridging automatically on the chain launch

It is currently under the debate if assets locked in the Mezo Portal contract
should be bridged automatically on the chain launch. The option when the user
triggers bridging feels to be less complex technically. For non-Bitcoin tokens,
Mezo Portal does not currently offer minting receipt tokens so all locked
non-Bitcoin tokens should be bridgable.

## Future Work

### Relaying post-bridge redemptions

One common usability issue when bridging tokens to another chain is the
necessity of redeeming them on the target chain once the bridging operation
completes. In case of Mezo, redemption requires calling the receive function of
the gateway to redeem wrapped tokens from the bridge and mint canonical tokens
on the chain. This requires users to already have Bitcoin on Mezo and remember
to execute this operation once the bridging is completed. Having this UX flaw is
acceptable for the initial implementation but should be improved over time by
introducing a relayer bot. This aspect is out of the scope of this RFC. One
option for the future is to extend the possibilities of the Ethereum sidecar to
return bridging events from the gRPC API and have the block proposer inject
zero-fee EVM transaction finalizing bridge operations. This solution bounds the
validator implementation with a specific bridging partner so keeping the relayer
as a separate process and binary makes more sense.
