# EVM compatibility

## Overview

Mezo achieves EVM compatibility by implementing components that collectively
support EVM state transitions and maintaining a developer experience similar to
Ethereum.

## EVM forks

Mezo offers EVM compatibility, supporting all Ethereum features
up to the London fork. For more information about London fork please see
[here](https://ethereum.org/en/history/#london).

## Ethereum JSON-RPC endpoints

### web3_clientVersion

- **Description**: Returns the current client version.
- **Parameters**: None.
- **Returns**: `String` - The current client version.

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"web3_clientVersion","params":[],"id":1}' -H "Content-Type: application/json" http://mezo-node-0.test.mezo.org:8545
```

### web3_sha3

- **Description**: Returns Keccak-256 of the given data.
- **Parameters**:
    - `String` - Data to convert into a SHA3 hash.
- **Returns**: `String` - The SHA3 result of the given data.

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"web3_sha3","params":["0x68656c6c6f20776f726c64"],"id":1}' -H "Content-Type: application/json" http://mezo-node-0.test.mezo.org:8545
```

### net_version

- **Description**: Returns the current network ID.
- **Parameters**: None.
- **Returns**: `String` - The current network ID.

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"net_version","params":[],"id":1}' -H "Content-Type: application/json" http://mezo-node-0.test.mezo.org:8545
```

### net_listening

- **Description**: Returns `true` if the client is actively listening for network connections.
- **Parameters**: None.
- **Returns**: `Boolean` - `true` if listening, `false` otherwise.

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"net_listening","params":[],"id":1}' -H "Content-Type: application/json" http://mezo-node-0.test.mezo.org:8545
```

### net_peerCount

- **Description**: Returns the number of peers currently connected to the client.
- **Parameters**: None.
- **Returns**: `String` - Number of connected peers.

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"net_peerCount","params":[],"id":1}' -H "Content-Type: application/json" http://mezo-node-0.test.mezo.org:8545
```

### eth_protocolVersion

- **Description**: Returns the current Ethereum protocol version.
- **Parameters**: None.
- **Returns**: `String` - The Ethereum protocol version.

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_protocolVersion","params":[],"id":1}' -H "Content-Type: application/json" http://mezo-node-0.test.mezo.org:8545
```

### eth_syncing

- **Description**: Returns an object with data about the sync status or `false` if not syncing.
- **Parameters**: None.
- **Returns**:
    - `Object | Boolean` - An object with sync status or `false`.

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_syncing","params":[],"id":1}' -H "Content-Type: application/json" http://mezo-node-0.test.mezo.org:8545
```

### eth_coinbase // TODO: decide if we need to include it in the docs.

- **Description**: Returns the client’s coinbase address (mining beneficiary).
This address is where any mining rewards will be sent if the node is mining.
- **Parameters**: None.
- **Returns**: `String` - Coinbase address.

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_coinbase","params":[],"id":1}' -H "Content-Type: application/json" http://mezo-node-0.test.mezo.org:8545
```

### eth_chainId

- **Description**: Returns the client’s chian ID.
- **Parameters**: None.
- **Returns**: `String` - Chain ID.

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_chainId","params":[],"id":1}' -H "Content-Type: application/json" http://mezo-node-0.test.mezo.org:8545
```

### eth_mining

- **Description**: Returns `true` if the client is actively mining new blocks.
- **Parameters**: None.
- **Returns**: `Boolean` - `true` if mining, `false` otherwise.

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_mining","params":[],"id":1}' -H "Content-Type: application/json" http://mezo-node-0.test.mezo.org:8545
```

### eth_hashrate - TODO: decide if we need to include it here. See "returns"

- **Description**: Returns the number of hashes per second that the node is mining with.
- **Parameters**: None.
- **Returns**: `String` - Proof-of-Work specific. This endpoint always returns 0.

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_hashrate","params":[],"id":1}' -H "Content-Type: application/json" http://mezo-node-0.test.mezo.org:8545
```

### eth_gasPrice

- **Description**: Returns the current price per gas in wei.
- **Parameters**: None.
- **Returns**: `String` - The current gas price in wei.

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_gasPrice","params":[],"id":1}' -H "Content-Type: application/json" http://mezo-node-0.test.mezo.org:8545
```

### eth_accounts

- **Description**: Returns a list of addresses owned by the client.
- **Parameters**: None.
- **Returns**: `Array` - Array of addresses.

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_accounts","params":[],"id":1}' -H "Content-Type: application/json" http://mezo-node-0.test.mezo.org:8545
```

### eth_blockNumber

- **Description**: Returns the number of the most recent block.
- **Parameters**: None.
- **Returns**: `String` - The block number.

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' -H "Content-Type: application/json" http://mezo-node-0.test.mezo.org:8545
```

### eth_getBalance

- **Description**: Returns the balance of the account of given address.
- **Parameters**:
    - `String` - Address to check for balance.
    - `String` - Block number.
- **Returns**: `String` - The balance in wei, as a hexadecimal.

TODO: add example

### eth_getStorageAt

- **Description**: Returns the value from a storage position at a given address.
- **Parameters**:
    - `String` - Address of the storage.
    - `String` - Storage position.
    - `String` - Block number.
- **Returns**: `String` - The value at this storage position.

TODO: add example

### eth_getTransactionCount

- **Description**: Returns the number of transactions sent from an address.
- **Parameters**:
    - `String` - Address to check for transaction count.
    - `String` - Block number.
- **Returns**: `String` - The transaction count as a hexadecimal number.

TODO: add example

### eth_getBlockTransactionCountByHash

- **Description**: Returns the number of transactions in a block from a block matching the given block hash.
- **Parameters**:
    - `String` - Block hash.
- **Returns**: `String` - The number of transactions in the block as a hexadecimal.

TODO: add example

### eth_getBlockTransactionCountByNumber

- **Description**: Returns the number of transactions in a block matching the given block number.
- **Parameters**:
    - `String` - Block number.
- **Returns**: `String` - The number of transactions in the block as a hexadecimal.

TODO: add example

### eth_getUncleCountByBlockHash

- **Description**: Returns the number of uncles in a block from a block matching the given block hash.
- **Parameters**:
    - `String` - Block hash.
- **Returns**: `String` - The number of uncles in the block as a hexadecimal.

TODO: add example

### eth_getUncleCountByBlockNumber

- **Description**: Returns the number of uncles in a block from a block matching the given block number.
- **Parameters**:
    - `String` - Block number.
- **Returns**: `String` - The number of uncles in the block as a hexadecimal.

TODO: add example

### eth_getCode

- **Description**: Returns the code at a given address.
- **Parameters**:
    - `String` - Address to get code from.
    - `String` - Block number.
- **Returns**: `String` - The code from the given address.

TODO: add example

### eth_sign

- **Description**: Signs data with a given address, resulting in a signature.
- **Parameters**:
    - `String` - Address to sign with.
    - `String` - Data to sign.
- **Returns**: `String` - The signature.

TODO: add example

### eth_sendTransaction

- **Description**: Creates and sends a new transaction.
- **Parameters**:
    - `Object` - Transaction object.
- **Returns**: `String` - The transaction hash.

TODO: add example

### eth_sendRawTransaction

- **Description**: Sends a signed transaction.
- **Parameters**:
    - `String` - The signed transaction data.
- **Returns**: `String` - The transaction hash.

TODO: add example

### eth_call

- **Description**: Executes a new message call immediately without creating a
transaction on the blockchain. Often used for executing read-only smart contract
functions.
- **Parameters**:
    - `Object`
      Object - The transaction call object
      from: DATA, 20 Bytes - (optional) The address the transaction is sent from.
      to: DATA, 20 Bytes - The address the transaction is directed to.
      gas: QUANTITY - (optional) Integer of the gas provided for the transaction
      execution. eth_call consumes zero gas, but this parameter may be needed by
      some executions.
      gasPrice: QUANTITY - (optional) Integer of the gasPrice used for each paid gas
      value: QUANTITY - (optional) Integer of the value sent with this transaction
      input: DATA - (optional) Hash of the method signature and encoded parameters.
      For details see Ethereum Contract ABI in the Solidity documentation(opens in
      a new tab).
    - `String` (optional) - Block number.
- **Returns**: `String` - The return value of the executed contract.

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_call","params":[{see above}],"id":1}'
```

### eth_estimateGas

- **Description**: Estimates the gas necessary to execute a transaction.
- **Parameters**:
    - `Object` - Transaction call object.
- **Returns**: `String` - The estimated gas amount.

TODO: add example

### eth_getBlockByHash

- **Description**: Returns information about a block by hash.
- **Parameters**:
    - `String` - Block hash.
    - `Boolean` - If `true`, returns full transaction objects; if `false`, returns only hashes.
- **Returns**: `Object` - Block information.

TODO: add example

### eth_getBlockByNumber

- **Description**: Returns information about a block by number.
- **Parameters**:
    - `String` - Block number.
    - `Boolean` - If `true`, returns full transaction objects; if `false`, returns only hashes.
- **Returns**: `Object` - Block information.

TODO: add example

### eth_getTransactionByHash

- **Description**: Returns the information about a transaction requested by transaction hash.
- **Parameters**:
    - `String` - Transaction hash.
- **Returns**: `Object` - Transaction information.

TODO: add example

### eth_getTransactionByBlockHashAndIndex

- **Description**: Returns information about a transaction by block hash and transaction index position.
- **Parameters**:
    - `String` - Block hash.
    - `String` - Transaction index position.
- **Returns**: `Object` - Transaction information.

TODO: add example

### eth_getTransactionByBlockNumberAndIndex

- **Description**: Returns information about a transaction by block number and transaction index position.
- **Parameters**:
    - `String` - Block number.
    - `String` - Transaction index position.
- **Returns**: `Object` - Transaction information.

TODO: add example

### eth_getTransactionReceipt

- **Description**: Returns the receipt of a transaction by transaction hash.
- **Parameters**:
    - `String` - Transaction hash.
- **Returns**: `Object` - Transaction receipt.

TODO: add example

### eth_getUncleByBlockHashAndIndex

- **Description**: Returns information about an uncle of a block by hash and uncle index position.
- **Parameters**:
    - `String` - Block hash.
    - `String` - Uncle index position.
- **Returns**: `Object` - Uncle block information.

TODO: add example

### eth_getUncleByBlockNumberAndIndex

- **Description**: Returns information about an uncle of a block by number and uncle index position.
- **Parameters**:
    - `String` - Block number.
    - `String` - Uncle index position.
- **Returns**: `Object` - Uncle block information.

TODO: add example

### eth_getLogs

- **Description**: Returns an array of logs matching the filter options.
- **Parameters**:
    - `Object` - Filter options.
- **Returns**: `Array` - Array of log objects.

### eth_getWork

- **Description**: Returns the hash of the current block, the seedHash, and the boundary condition to be met.
- **Parameters**: None.
- **Returns**: `Array` - An array with data needed for mining.

### eth_submitWork

- **Description**: Used for submitting a proof-of-work solution.
- **Parameters**:
    - `String` - Nonce.
    - `String` - Header’s hash.
    - `String` - Mix digest.
- **Returns**: `Boolean` - `true` if the solution is valid, otherwise `false`.

### eth_submitHashrate

- **Description**: Used for submitting mining hash rate.
- **Parameters**:
    - `String` - Hash rate.
    - `String` - ID of the client miner.
- **Returns**: `Boolean` - `true` if accepted.