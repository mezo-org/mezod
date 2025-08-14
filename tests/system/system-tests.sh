#!/bin/sh
# This script runs the system tests for the Mezo blockchain.

# Create interfaces dir if it doesn't already exist
mkdir -p ./contracts/interfaces/solidity

# Copy ERC20 interfaces to the solidity sub-directory.
cp ../../precompile/erc20/IERC20.sol ./contracts/interfaces/solidity/
cp ../../precompile/erc20/IERC20Metadata.sol ./contracts/interfaces/solidity/
cp ../../precompile/erc20/IERC20WithPermit.sol ./contracts/interfaces/solidity/

# Copy BTC Token interfaces
cp ../../precompile/btctoken/IBTC.sol ./contracts/interfaces
# Adjust imports in IBTC.sol
sed -i '' 's|../erc20/|./solidity/|g' ./contracts/interfaces/IBTC.sol

# Copy MEZO Token interface
cp ../../precompile/mezotoken/IMEZO.sol ./contracts/interfaces
# Adjust imports in IMEZO.sol
sed -i '' 's|../erc20/|./solidity/|g' ./contracts/interfaces/IMEZO.sol

# Copy Validator Pool interface
cp ../../precompile/validatorpool/IValidatorPool.sol ./contracts/interfaces
# Copy Asset Bridge interface
cp ../../precompile/assetsbridge/IAssetsBridge.sol ./contracts/interfaces

# Install dependencies
npm i

# Run the system tests
npm run test
