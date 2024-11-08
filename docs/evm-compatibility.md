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

### eth_coinbase

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
- **Returns**: `String` - The balance in abtc, as a hexadecimal.

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_getBalance","params":["0x0504d82efb7db7a8c05e8df8cea575d8c9f48bb2","latest"],"id":1}' -H "Content-Type: application/json" http://mezo-node-0.test.mezo.org:8545
```

### eth_getStorageAt

- **Description**: Returns the value from a storage position at a given address.
- **Parameters**:
    - `String` - Address of the storage.
    - `String` - Storage position.
    - `String` - Block number.
- **Returns**: `String` - The value at this storage position.

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_getStorageAt","params":["0x0504d82efb7db7a8c05e8df8cea575d8c9f48bb2","0x0","latest"],"id":1}' -H "Content-Type: application/json" http://mezo-node-0.test.mezo.org:8545
```

### eth_getTransactionCount

- **Description**: Returns the number of transactions sent from an address.
- **Parameters**:
    - `String` - Address to check for transaction count.
    - `String` - Block number.
- **Returns**: `String` - The transaction count as a hexadecimal number.

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_getTransactionCount","params":["0x0504d82efb7db7a8c05e8df8cea575d8c9f48bb2","latest"],"id":1}' -H "Content-Type: application/json" http://mezo-node-0.test.mezo.org:8545
```

### eth_getBlockTransactionCountByHash

- **Description**: Returns the number of transactions in a block from a block
matching the given block hash.
- **Parameters**:
    - `String` - Block hash.
- **Returns**: `String` - The number of transactions in the block as a hexadecimal.

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_getBlockTransactionCountByHash","params":["0x41175c10b68dd0bfa27f2533a23979445a5d643427e0ffd1870d11806f31b291"],"id":1}' -H "Content-Type: application/json" http://mezo-node-0.test.mezo.org:8545
```

### eth_getBlockTransactionCountByNumber

- **Description**: Returns the number of transactions in a block matching the given block number.
- **Parameters**:
    - `String` - Block number.
- **Returns**: `String` - The number of transactions in the block as a hexadecimal.

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_getBlockTransactionCountByNumber","params":["0x1"],"id":1}' -H "Content-Type: application/json" http://mezo-node-0.test.mezo.org:8545
```

### eth_getCode

- **Description**: Returns the code (compiled bytecode of a smart contract) at a given address.
- **Parameters**:
    - `String` - Address to get code from.
    - `String` - Block number.
- **Returns**: `String` - The code from the given address. If the response is empty ("0x"),
it indicates the address is likely an externally owned account rather than a contract account.

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_getCode","params":["0x0504d82efb7db7a8c05e8df8cea575d8c9f48bb2","latest"],"id":1}' -H "Content-Type: application/json" http://mezo-node-0.test.mezo.org:8545
```

### eth_sign

- **Description**: Signs data with a given address, resulting in a signature. The address to sign with must be unlocked.
- **Parameters**:
    - `String` - Address to sign with.
    - `String` - Data to sign.
- **Returns**: `String` - The signature.

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_sign","params":["0x0504d82efb7db7a8c05e8df8cea575d8c9f48bb2","0xdeadbeef"],"id":1}' -H "Content-Type: application/json" http://mezo-node-0.test.mezo.org:8545
```

### eth_sendTransaction

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
- **Returns**: `String` - The transaction hash.

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_sendTransaction","params":[{see above}],"id":1}' -H "Content-Type: application/json" http://mezo-node-0.test.mezo.org:8545
```

### eth_sendRawTransaction

- **Description**: Creates new message call transaction or a contract creation for signed transactions.
- **Parameters**:
    - `String` - The signed transaction data.
- **Returns**: `String` - The transaction hash.

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_sendRawTransaction","params":["0xd46e8dd67c5d32be8d46e8dd67c5d32be8058bb8eb970870f072445675058bb8eb970870f072445675"],"id":1}' -H "Content-Type: application/json" http://mezo-node-0.test.mezo.org:8545
```

### eth_call

- **Description**: Executes a new message call immediately without creating a
transaction on the blockchain. Often used for executing read-only smart contract
functions.
- **Parameters**:
    - `Object` - The transaction call object
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
    Object:
    - `from`: DATA, 20 Bytes - The address the transaction is send from.
    - `to`: DATA, 20 Bytes - (optional when creating new contract) The address the transaction is directed to.
    - `value`: QUANTITY - value sent with this transaction
- **Returns**: `String` - The estimated gas amount.

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_estimateGas","params":[{see above}],"id":1}'
```

### eth_getBlockByHash

- **Description**: Returns information about a block by hash.
- **Parameters**:
    - `String` - Block hash.
    - `Boolean` - If `true`, returns full transaction objects; if `false`, returns only hashes.
- **Returns**: `Object` - Block information.

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_getBlockByHash","params":["0x8d80d1a8ac12c5e57c17c580afbb4c03987649934b60ce04ec89fcd336e3a186", true],"id":1}' -H "Content-Type: application/json" http://mezo-node-0.test.mezo.org:8545
```

### eth_getBlockByNumber

- **Description**: Returns information about a block by number.
- **Parameters**:
    - `String` - Block number.
    - `Boolean` - If `true`, returns full transaction objects; if `false`, returns only hashes.
- **Returns**: `Object` - Block information.

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["0x1b4", true],"id":1}' -H "Content-Type: application/json" http://mezo-node-0.test.mezo.org:8545
```

### eth_getTransactionByHash

- **Description**: Returns the information about a transaction requested by transaction hash.
- **Parameters**:
    - `String` - Transaction hash.
- **Returns**: `Object` - Transaction information.

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_getTransactionByHash","params":["0x5c504ed432cb51138bcf09aa5e8a410dd4a1e204ef84bfed1be16dfba1b22060"],"id":1}' -H "Content-Type: application/json" http://mezo-node-0.test.mezo.org:8545
```

### eth_getTransactionByBlockHashAndIndex

- **Description**: Returns information about a transaction by block hash and transaction index position.
- **Parameters**:
    - `String` - Block hash.
    - `String` - Transaction index position.
- **Returns**: `Object` - Transaction information.

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_getTransactionByBlockHashAndIndex","params":["0x8d80d1a8ac12c5e57c17c580afbb4c03987649934b60ce04ec89fcd336e3a186", "0x0"],"id":1}' -H "Content-Type: application/json" http://mezo-node-0.test.mezo.org:8545
```

### eth_getTransactionByBlockNumberAndIndex

- **Description**: Returns information about a transaction by block number and transaction index position.
- **Parameters**:
    - `String` - Block number.
    - `String` - Transaction index position.
- **Returns**: `Object` - Transaction information.

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_getTransactionByBlockNumberAndIndex","params":["0x1b4", "0x0"],"id":1}' -H "Content-Type: application/json" http://mezo-node-0.test.mezo.org:8545
```

### eth_getTransactionReceipt

- **Description**: Returns the receipt of a transaction by transaction hash.
- **Parameters**:
    - `String` - Transaction hash.
- **Returns**: `Object` - Transaction receipt.

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_getTransactionReceipt","params":["0x1758f2ad26d448ecdcc2f225432c520bc77c03194536e76f6776f8c5dabce9a9"],"id":1}' -H "Content-Type: application/json" http://mezo-node-0.test.mezo.org:8545
```

### eth_getLogs

- **Description**: Returns an array of logs matching the filter options.
- **Parameters**:
    Filter options
    - `fromBlock`: QUANTITY|TAG - (optional, default: "latest") Integer block number, or "latest" for the last mined
    block or "pending", "earliest" for not yet mined transactions.
    - `toBlock`: QUANTITY|TAG - (optional, default: "latest") Integer block number, or "latest" for the last mined block
    or "pending", "earliest" for not yet mined transactions.
    - `address`: DATA|Array, 20 Bytes - (optional) Contract address or a list of addresses from which logs should originate.
    - `topics`: Array of DATA, - (optional) Array of 32 Bytes DATA topics. Topics are order-dependent. Each topic can
    also be an array of DATA with “or” options.
    - `blockhash`: (optional, future) With the addition of EIP-234, blockHash will be a new filter option which restricts
    the logs returned to the single block with the 32-byte hash blockHash. Using blockHash is equivalent to fromBlock =
    toBlock = the block number with hash blockHash. If blockHash is present in in the filter criteria, then neither fromBlock
    nor toBlock are allowed.
- **Returns**: `Array` - Array of log objects.

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_getLogs","params":[{"fromBlock": "0x1", "toBlock": "0x2"}],"id":1}' -H "Content-Type: application/json" http://mezo-node-0.test.mezo.org:8545
```

### eth_newFilter

- **Description**: Create new filter using topics of some kind.
- **Parameters**:
    - `String` - hash of a transaction
- **Returns**: `String` - Filter ID.

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_newFilter","params":[{"fromBlock": "0x1", "toBlock": "0x2"}],"id":1}' -H "Content-Type: application/json" http://mezo-node-0.test.mezo.org:8545
```

### eth_newBlockFilter

- **Description**: Creates a filter in the node to notify when a new block arrives.
- **Parameters**: None.
- **Returns**: `String` - A filter ID.

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_newBlockFilter","params":[],"id":1}' -H "Content-Type: application/json" http://mezo-node-0.test.mezo.org:8545
```

### eth_newPendingTransactionFilter

- **Description**: Creates a filter in the node to notify when new pending transactions arrive.
- **Parameters**: None.
- **Returns**: `String` - A filter ID.

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_newPendingTransactionFilter","params":[],"id":1}' -H "Content-Type: application/json" http://mezo-node-0.test.mezo.org:8545
```

### eth_uninstallFilter

- **Description**: Uninstalls a filter with the given ID.
- **Parameters**:
    - `String` - The filter ID.
- **Returns**: `Boolean` - `true` if the filter was successfully uninstalled.

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_uninstallFilter","params":["0x1"],"id":1}' -H "Content-Type: application/json" http://mezo-node-0.test.mezo.org:8545
```

### eth_getFilterChanges

- **Description**: Checks for changes to a filter since the last call.
- **Parameters**:
    - `String` - The filter ID.
- **Returns**: `Array` - An array of logs that have occurred since the last poll.

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_getFilterChanges","params":["0x1"],"id":1}' -H "Content-Type: application/json" http://mezo-node-0.test.mezo.org:8545
```

### eth_getFilterLogs

- **Description**: Returns an array of all logs for a given filter ID, containing all past logs matching the filter.
- **Parameters**:
    - `String` - The filter ID.
- **Returns**: `Array` - An array of log objects that match the filter, providing historical log data.

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_getFilterLogs","params":["0x1"],"id":1}' -H "Content-Type: application/json" http://mezo-node-0.test.mezo.org:8545
```

### eth_getProof

- **Description**: Returns the account and storage values of a specified account, including the Merkle proof.
- **Parameters**:
    - `String` - Address of the account.
    - `Array` of `String` - An array of storage keys that you want the values and proof for.
    - `String` - Block number to get the proof for.
- **Returns**: `Object` - An object containing the account details, storage proof, and relevant Merkle proofs.

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_getProof","params":["0x1234567890123456789012345678901234567890",["0x0000000000000000000000000000000000000000000000000000000000000000","0x0000000000000000000000000000000000000000000000000000000000000001"],`"latest"`],"id":1}' -H "Content-type:application/json" http://mezo-node-0.test.mezo.org:8545
```

## Due to the nature of PoA consensus, the following methods are not applicable and might revert or return empty results

### eth_getUncleCountByBlockHash
### eth_getUncleCountByBlockNumber
### eth_getUncleByBlockHashAndIndex
### eth_getUncleByBlockNumberAndIndex
### eth_getWork
### eth_submitWork
### eth_submitHashrate
### eth_hashrate
### eth_mining