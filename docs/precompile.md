# Precompiles

## Overview

Precompiles are built-in smart contracts that offer optimized operations within
the EVM. They provide specialized functionalities that extend beyond standard
Solidity-based contracts.

## Implementation

TODO

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
