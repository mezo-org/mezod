# Connect MEZO NTT on Mezo, Ethereum, Base and BSC with Robinhood Chain

Safe Transaction Builder batches for the Mezo governance multisig
[`0x98D8899c3030741925BE630C710A98B57F397C7a`](https://etherscan.io/address/0x98D8899c3030741925BE630C710A98B57F397C7a).
They register the new Robinhood Chain MEZO NTT deployment as a peer of the
existing MEZO Wormhole NTT deployments on Mezo, Ethereum, Base and BSC.

The Robinhood side (NTT Manager and Wormhole Transceiver on Robinhood Chain)
already points at our supported chains. Its ownership has not been transferred
to the multisig yet, so the deployer still owns it. These four batches add the
other direction. Until they run, MEZO can leave Robinhood Chain but Mezo,
Ethereum, Base and BSC will not accept transfers from it.

Linear: [SUP-234 Support Robinhood for MEZO](https://linear.app/thesis-co/issue/SUP-234/support-robinhood-for-mezo).
The MEZO token deployment on Robinhood Chain is
[mezo-org/mezod#740](https://github.com/mezo-org/mezod/pull/740).

## Files

| File | Network | EVM chain ID | Safe batch checksum |
| ---- | ------- | ------------ | ------------------- |
| [`MezoRobinhood.json`](./MezoRobinhood.json) | Mezo mainnet | 31612 | `0x4f865ff4e7b7c43e91e511d377f5af3051a093766d342b552f3ded43f56544d6` |
| [`EthereumRobinhood.json`](./EthereumRobinhood.json) | Ethereum | 1 | `0x237c1a190a9514d078c5626611b1fd5ee6f03a97a67659eeb4581cc98367dbb2` |
| [`BaseRobinhood.json`](./BaseRobinhood.json) | Base | 8453 | `0xc5ffe71afb7fb62bdc3ac982855b0c3358198cb1eed132f8aa981204dbdaafab` |
| [`BscRobinhood.json`](./BscRobinhood.json) | BNB Smart Chain | 56 | `0xe8192548649b884d6b1bffcca5a45254dd4d95467c0de5ec3e18d81e0c29731c` |

The files are byte-for-byte the batches exported from Safe Transaction Builder
2.1.0. The `meta.checksum` of each one was recomputed and matches, so the files
were not edited after export.

## What every batch does

Each batch has the same two calls, both with `value` 0. Only the target
addresses differ between chains.

1. **`NttManager.setPeer(uint16 peerChainId, bytes32 peerContract, uint8 decimals, uint256 inboundLimit)`**
   (selector `0x7c918634`)
    - `peerChainId` = `72`, the Wormhole chain ID of Robinhood Chain (EVM chain ID 4663). See
      [Wormhole SDK chain IDs](https://github.com/wormhole-foundation/wormhole-sdk-ts/blob/main/core/base/src/constants/chains.ts)
      and `ChainIDRobinhoodChain = 72` in
      [wormhole `sdk/vaa/structs.go`](https://github.com/wormhole-foundation/wormhole/blob/main/sdk/vaa/structs.go).
    - `peerContract` =
      `0x0000000000000000000000006c84a8f1c29108f47a79964b5fe888d4f4d0de40`, the
      Robinhood Chain NTT Manager
      [`0x6c84a8f1c29108F47a79964b5Fe888D4f4D0dE40`](https://robinhoodchain.blockscout.com/address/0x6c84a8f1c29108F47a79964b5Fe888D4f4D0dE40),
      left-padded to 32 bytes.
    - `decimals` = `18`, the decimals of MEZO on Robinhood Chain
      ([`0x8e4cbBcc33dB6c0a18561fDE1F6bA35906d4848b`](https://robinhoodchain.blockscout.com/address/0x8e4cbBcc33dB6c0a18561fDE1F6bA35906d4848b),
      the standard 18-decimal `MEZO.sol`).
    - `inboundLimit` = `50000000000000000000000000`, which is **50,000,000 MEZO**
      (18 decimals). This is the inbound rate limit for transfers that come from
      Robinhood Chain: at most 50M MEZO in each rolling rate-limit window (24h by
      default in NTT). Transfers above the remaining capacity are queued, not lost.
2. **`WormholeTransceiver.setWormholePeer(uint16 peerChainId, bytes32 peerContract)`**
   (selector `0x7ab56403`)
    - `peerChainId` = `72` (Robinhood Chain).
    - `peerContract` =
      `0x000000000000000000000000236aa50979d5f3de3bd1eeb40e81137f22ab794b`, the
      Robinhood Chain Wormhole Transceiver
      [`0x236aa50979D5f3De3Bd1Eeb40E81137F22ab794b`](https://robinhoodchain.blockscout.com/address/0x236aa50979D5f3De3Bd1Eeb40E81137F22ab794b).

Raw calldata, identical on all four chains:

```text
setPeer:
0x7c9186340000000000000000000000000000000000000000000000000000000000000048000000000000000000000000
6c84a8f1c29108f47a79964b5fe888d4f4d0de400000000000000000000000000000000000000000000000000000000000000012
000000000000000000000000000000000000000000295be96e64066972000000

setWormholePeer:
0x7ab564030000000000000000000000000000000000000000000000000000000000000048000000000000000000000000
236aa50979d5f3de3bd1eeb40e81137f22ab794b
```

(Each call is one line in the JSON. It is wrapped here for readability.)

Both functions are `onlyOwner`. The Wormhole Transceiver checks ownership
against the owner of its NTT Manager.

## Targets per chain

The addresses match the MEZO NTT deployments listed in the
[Mezo contracts reference](https://github.com/mezo-org/documentation/blob/main/src/content/docs/docs/users/resources/contracts-reference.md#mezo-token-and-bridge).

| Network | Tx | Target | Call |
| ------- | -- | ------ | ---- |
| Mezo | 1 | NTT Manager [`0x5E668D912F69db0762CD0fE00647dA9fa6b29591`](https://explorer.mezo.org/address/0x5E668D912F69db0762CD0fE00647dA9fa6b29591) | `setPeer(72, Robinhood NttManager, 18, 50M MEZO)` |
| Mezo | 2 | Wormhole Transceiver [`0x528edf2dBbC6Ae6978c88E54B260AB86B682D18c`](https://explorer.mezo.org/address/0x528edf2dBbC6Ae6978c88E54B260AB86B682D18c) | `setWormholePeer(72, Robinhood WormholeTransceiver)` |
| Ethereum | 1 | NTT Manager [`0x13916D0daB357DCBAa1600b594d62C641840686A`](https://etherscan.io/address/0x13916D0daB357DCBAa1600b594d62C641840686A) | `setPeer(72, Robinhood NttManager, 18, 50M MEZO)` |
| Ethereum | 2 | Wormhole Transceiver [`0x920871aF2D4106E76d204fEA7122fA129C9283b1`](https://etherscan.io/address/0x920871aF2D4106E76d204fEA7122fA129C9283b1) | `setWormholePeer(72, Robinhood WormholeTransceiver)` |
| Base | 1 | NTT Manager [`0x0c46f496C410465975a427e34a976fc15a2edE4f`](https://basescan.org/address/0x0c46f496C410465975a427e34a976fc15a2edE4f) | `setPeer(72, Robinhood NttManager, 18, 50M MEZO)` |
| Base | 2 | Wormhole Transceiver [`0x27321f84704a599aB740281E285cc4463d89A3D5`](https://basescan.org/address/0x27321f84704a599aB740281E285cc4463d89A3D5) | `setWormholePeer(72, Robinhood WormholeTransceiver)` |
| BSC | 1 | NTT Manager [`0x09959798B95d00a3183d20FaC298E4594E599eab`](https://bscscan.com/address/0x09959798B95d00a3183d20FaC298E4594E599eab) | `setPeer(72, Robinhood NttManager, 18, 50M MEZO)` |
| BSC | 2 | Wormhole Transceiver [`0xa10aD2570ea7b93d19fDae6Bd7189fF4929Bc747`](https://bscscan.com/address/0xa10aD2570ea7b93d19fDae6Bd7189fF4929Bc747) | `setWormholePeer(72, Robinhood WormholeTransceiver)` |

The Safe address is the same on every chain:
[Mezo](https://explorer.mezo.org/address/0x98D8899c3030741925BE630C710A98B57F397C7a),
[Ethereum](https://etherscan.io/address/0x98D8899c3030741925BE630C710A98B57F397C7a),
[Base](https://basescan.org/address/0x98D8899c3030741925BE630C710A98B57F397C7a),
[BSC](https://bscscan.com/address/0x98D8899c3030741925BE630C710A98B57F397C7a).

`0xa10aD2570ea7b93d19fDae6Bd7189fF4929Bc747` is also the address of mSolvBTC
on Mezo. That is a coincidence of deployer and nonce on different chains. On
BSC it is the MEZO Wormhole Transceiver.

## Expected state

| Check (per chain) | Before | After |
| ----------------- | ------ | ----- |
| `NttManager.owner()` and `WormholeTransceiver.owner()` | Safe `0x98D8…7C7a` | unchanged |
| `NttManager.getPeer(72)` | `(0x00…00, 0)` | `(0x…6c84a8f1c29108f47a79964b5fe888d4f4d0de40, 18)` |
| `NttManager.getInboundLimitParams(72).limit` | 0 | 50,000,000 MEZO (stored trimmed to 8 decimals: `5000000000000000` with decimals `8`) |
| `WormholeTransceiver.getWormholePeer(72)` | `0x00…00` | `0x…236aa50979d5f3de3bd1eeb40e81137f22ab794b` |

`setWormholePeer` reverts with `PeerAlreadySet` if a peer for chain 72 is
already set. If the "before" value is not zero, stop and ask.

## How to verify before signing

1. **Check the batch in the Safe UI.** Open Transaction Builder on the right
   network and drag in the JSON. Expect two calls, each with value 0, in this
   order: NTT Manager, then Wormhole Transceiver. Compare the targets with the
   table above and the calldata with the raw calldata above.
2. **Decode the calldata yourself** with [Foundry](https://getfoundry.sh/) `cast`:

    ```sh
    cast calldata-decode "setPeer(uint16,bytes32,uint8,uint256)" \
      0x7c91863400000000000000000000000000000000000000000000000000000000000000480000000000000000000000006c84a8f1c29108f47a79964b5fe888d4f4d0de400000000000000000000000000000000000000000000000000000000000000012000000000000000000000000000000000000000000295be96e64066972000000
    # 72, 0x0000…6c84a8f1c29108f47a79964b5fe888d4f4d0de40, 18, 50000000000000000000000000 [5e25]

    cast calldata-decode "setWormholePeer(uint16,bytes32)" \
      0x7ab564030000000000000000000000000000000000000000000000000000000000000048000000000000000000000000236aa50979d5f3de3bd1eeb40e81137f22ab794b
    # 72, 0x0000…236aa50979d5f3de3bd1eeb40e81137f22ab794b

    cast to-unit 50000000000000000000000000 ether   # 50000000
    ```

3. **Check the current state** on each chain. Example for Ethereum. Swap in the
   addresses from the table and the RPC for the other chains (Mezo:
   `https://rpc-http.mezo.org`).

    ```sh
    RPC=https://ethereum-rpc.publicnode.com
    MGR=0x13916D0daB357DCBAa1600b594d62C641840686A
    XCVR=0x920871aF2D4106E76d204fEA7122fA129C9283b1

    cast call $MGR  "owner()(address)" --rpc-url $RPC                  # Safe 0x98D8…7C7a
    cast call $XCVR "owner()(address)" --rpc-url $RPC                  # Safe 0x98D8…7C7a
    cast call $MGR  "getPeer(uint16)((bytes32,uint8))" 72 --rpc-url $RPC   # zero before
    cast call $XCVR "getWormholePeer(uint16)(bytes32)" 72 --rpc-url $RPC   # zero before
    cast call $XCVR "nttManager()(address)" --rpc-url $RPC             # $MGR
    ```

4. **Check the Robinhood side** that we are pointing at:

    ```sh
    RH=https://rpc.mainnet.chain.robinhood.com
    RH_MGR=0x6c84a8f1c29108F47a79964b5Fe888D4f4D0dE40
    RH_XCVR=0x236aa50979D5f3De3Bd1Eeb40E81137F22ab794b

    cast call $RH_MGR  "chainId()(uint16)" --rpc-url $RH                 # 72
    cast call $RH_MGR  "token()(address)" --rpc-url $RH                  # MEZO 0x8e4cbBcc33dB6c0a18561fDE1F6bA35906d4848b
    cast call $RH_MGR  "getTransceivers()(address[])" --rpc-url $RH      # includes $RH_XCVR
    cast call $RH_XCVR "nttManager()(address)" --rpc-url $RH             # $RH_MGR
    # Robinhood already points back at us (Wormhole IDs: Ethereum 2, BSC 4, Base 30, Mezo 50):
    cast call $RH_MGR  "getPeer(uint16)((bytes32,uint8))" 2 --rpc-url $RH    # Ethereum NTT Manager
    cast call $RH_XCVR "getWormholePeer(uint16)(bytes32)" 2 --rpc-url $RH    # Ethereum Wormhole Transceiver
    ```

5. **Check the Safe transaction hash** on your hardware wallet before you sign,
   for example with
   [`safe-tx-hashes-util`](https://github.com/pcaversaccio/safe-tx-hashes-util)
   (`./safe_hashes.sh --network ethereum --address 0x98D8899c3030741925BE630C710A98B57F397C7a --nonce <nonce>`).
   The Safe UI sends the batch as one `multiSend` delegatecall to the
   MultiSendCallOnly contract.

## Follow-ups not covered by these batches

- Ownership of the Robinhood Chain NTT Manager and Wormhole Transceiver will
  move to the multisig in a separate step (SUP-234).
- Solana is listed in SUP-234 but has no batch here. It has its own
  owner/upgrade authority.
