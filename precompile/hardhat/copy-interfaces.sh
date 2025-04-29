#!/bin/sh
# This script copies the solidity interfaces from the custom precompile
# directories to `./interfaces`. The `interfaces` directory is included
# in `.gitignore` to prevent the duplicates getting committed.

# Create interfaces dir if it doesn't already exist
mkdir -p ./interfaces/solidity

# Copy ERC20 interfaces to the solidity sub-directory.
# This is done due to legacy reasons, in order to not change
# the BTCCaller contract bytecode.
cp ../erc20/IERC20.sol ./interfaces/solidity/
cp ../erc20/IERC20Metadata.sol ./interfaces/solidity/
cp ../erc20/IERC20WithPermit.sol ./interfaces/solidity/

# Copy BTC Token interface
cp ../btctoken/IBTC.sol ./interfaces/
# Adjust imports in IBTC.sol
sed -i '' 's|../erc20/|./solidity/|g' ./interfaces/IBTC.sol

# Copy MEZO Token interface
cp ../mezotoken/IMEZO.sol ./interfaces/
# Adjust imports in IMEZO.sol
sed -i '' 's|../erc20/|./solidity/|g' ./interfaces/IMEZO.sol

# Copy Validator Pool interface
cp ../validatorpool/IValidatorPool.sol ./interfaces/

# Copy Assets Bridge interface
cp ../assetsbridge/IAssetsBridge.sol ./interfaces/

# Copy Maintenance interface
cp ../maintenance/IMaintenance.sol ./interfaces/

# Copy Upgrade interface
cp ../upgrade/IUpgrade.sol ./interfaces/

# Copy Price Oracle interface
cp ../priceoracle/IPriceOracle.sol ./interfaces/