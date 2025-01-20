#!/bin/sh
# This script copies the solidity interfaces from the custom precompile
# directories to `./interfaces`. The `interfaces` directory is included
# in `.gitignore` to prevent the duplicates getting committed.

# Create interfaces dir if it doesn't already exist
mkdir -p ./interfaces

# Copy BTC Token interfaces
cp ../btctoken/IBTC.sol ./interfaces/
cp -R ../btctoken/solidity ./interfaces/

# Copy Validator Pool interface
cp ../validatorpool/IValidatorPool.sol ./interfaces/

# Copy Assets Bridge interface
cp ../assetsbridge/IAssetsBridge.sol ./interfaces/

# Copy Maintenance interface
cp ../maintenance/IMaintenance.sol ./interfaces/

# Copy Upgrade interface
cp ../upgrade/IUpgrade.sol ./interfaces/