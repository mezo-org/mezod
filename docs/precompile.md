# Precompiles

## Overview

Precompiles are built-in smart contracts that offer optimized operations within
the EVM. They provide specialized functionalities that extend beyond standard
Solidity-based contracts.

## Implementation

The precompile framework extends the CosmosSDK-based EVM implementation by
introducing built-in smart contracts executed natively by the chain. These
precompiles provide various functionalities, such as asset bridging, price
oracles, and maintenance operations.

### Main Components

1. **Precompile Contracts:** The code for precompiles is located inside the
   [precompile](https://github.com/mezo-org/mezod/tree/main/precompile) directory.
   Each precompile defines its logic, supported methods, and how they
   interact with the EVM state.

2. **ABI and bytecode generation:** A precompile must provide ABI located in
   `abi.json` and EVM bytecode located in `byte_code.go`. They both can be
   obtained by compiling a Solidity contract stored in the Hardhat's
   [contract](https://github.com/mezo-org/mezod/tree/precompiles-doc/precompile/hardhat/contracts)
   directory.

3. **Versioning:** Precompiles can have multiple versions. The versions are
   handled by a version map.

4. **Method Execution:** Each precompile registers methods that follow the
   [Method](https://github.com/mezo-org/mezod/blob/9d0d45abd839ce4d9e003536724cd8d38b0031f7/precompile/method.go#L30)
   interface.

5. **Hardhat Integration:** Precompiles can be interacted with using
   [Hardhat tasks](https://github.com/mezo-org/mezod/tree/main/precompile/hardhat/tasks),
   which provide CLI-based access for reading and writing data to the blockchain.

### Adding a New Precompile

To add a new precompile:

1. **Define the Solidity interface:** Create a Solidity interface. See example
   for [BTC Token](https://github.com/mezo-org/mezod/tree/main/precompile/btctoken).

2. **Implement the Solidity caller contract:** Create a contract that calls the
   precompile. See example for
   [BTC Token](https://github.com/mezo-org/mezod/blob/main/precompile/hardhat/contracts/BTCCaller.sol).

3. **Generate ABI and bytecode:** Use Hardhat to compile the contract and
   extract the `deployedBytecode` and `abi` from the artifacts.

4. **Implement the precompile logic:** Create a new package for the new
   precompile. Implement methods and register them. See example for
   [BTC Token](https://github.com/mezo-org/mezod/tree/main/precompile/btctoken).

5. **Register the Precompile in the Chain:** Create an instance of the new precompile
   in the [Mezo app](https://github.com/mezo-org/mezod/blob/9d0d45abd839ce4d9e003536724cd8d38b0031f7/app/app.go#L898)

6. **Write Hardhat Tasks:** Add new tasks to
   [Hardhat tasks](https://github.com/mezo-org/mezod/tree/main/precompile/hardhat/tasks).

## Available Precompiles

The following precompiles are available in the Mezo blockchain:

### BTC Token Precompile

#### Overview

The BTC Token precompile provides an interface for interacting with the Mezo's
native BTC token. It enables standard ERC-20 token operations such as balance
queries, transfers, approvals.

#### Address

```
0x7b7c000000000000000000000000000000000000
```

#### Methods

See [IBTC](https://github.com/mezo-org/mezod/blob/main/precompile/btctoken/IBTC.sol)
interface for a list of methods.

### Validator Pool Precompile

#### Overview

The Validator Pool precompile provides an interface for managing validators on
the Mezo blockchain. It allows submitting and approving validator applications,
modifying privileges, and handling validator ownership transfers.

#### Address

```
0x7b7c000000000000000000000000000000000011
```

#### Methods

See [IValidatorPool](https://github.com/mezo-org/mezod/blob/main/precompile/validatorpool/IValidatorPool.sol)
interface for a list of methods.

### Assets Bridge Precompile

#### Overview

The Assets Bridge precompile enables observability of pseudo-transactions,
allowing the Mezo blockchain explorer to decipher these transactions. It also
provides limited asset bridging functionalities, such as ERC-20 token mapping.

#### Address

```
0x7b7c000000000000000000000000000000000012
```

#### Methods

See [IAssetsBridge](https://github.com/mezo-org/mezod/blob/main/precompile/assetsbridge/IAssetsBridge.sol)
interface for a list of methods.

### Maintenance Precompile

#### Overview

The Maintenance precompile provides administrative functions for managing
certain blockchain settings, such as enabling or disabling non-EIP155
transactions and modifying precompile bytecode.

#### Address

```
0x7b7c000000000000000000000000000000000013
```

#### Methods

See [IMaintenance](https://github.com/mezo-org/mezod/blob/main/precompile/maintenance/IMaintenance.sol)
interface for a list of methods.

### Upgrade Precompile

#### Overview

The Upgrade precompile provides an interface for managing blockchain upgrade plans.
It allows for scheduling, retrieving, and canceling upgrade plans, ensuring
a controlled upgrade process for the Mezo blockchain.

#### Address

```
0x7b7c000000000000000000000000000000000014
```

#### Methods

See [IUpgrade](https://github.com/mezo-org/mezod/blob/main/precompile/upgrade/IUpgrade.sol)
interface for a list of methods.

### Price Oracle Precompile

#### Overview

The Price Oracle precompile provides access to price data for assets tracked
by the Mezo blockchain. It enables querying the latest price updates and
retrieving asset price information with a predefined level of precision.

#### Address

```
0x7b7c000000000000000000000000000000000015
```

#### Methods

See [IPriceOracle](https://github.com/mezo-org/mezod/blob/main/precompile/priceoracle/IPriceOracle.sol)
interface for a list of methods.
