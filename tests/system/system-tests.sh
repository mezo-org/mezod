#!/bin/sh
# This script runs the system tests for the Mezo blockchain.
#
# Usage:
#   ./system-tests.sh                           # run all suites on localhost
#   ./system-tests.sh AssetsBridge              # run a single suite
#   ./system-tests.sh TripartyBridge BTCTransfers  # run multiple suites
#   NETWORK=testnet PRIVATE_KEYS=0x... ./system-tests.sh TripartyBridge
#
# Environment variables:
#   NETWORK      - hardhat network name (default: localhost)
#   RPC_URL      - JSON-RPC endpoint (defaults per network, see hardhat.config.ts)
#   PRIVATE_KEYS - comma-separated hex private keys (appended to localnode
#                  keys on localhost, mandatory on testnet)

set -e

NETWORK="${NETWORK:-localhost}"

# Create interfaces dir if it doesn't already exist
mkdir -p ./contracts/interfaces/solidity

# Copy ERC20 interfaces to the solidity sub-directory.
cp ../../precompile/erc20/IERC20.sol ./contracts/interfaces/solidity/
cp ../../precompile/erc20/IERC20Metadata.sol ./contracts/interfaces/solidity/
cp ../../precompile/erc20/IERC20WithPermit.sol ./contracts/interfaces/solidity/

# Copy BTC Token interfaces
cp ../../precompile/btctoken/IBTC.sol ./contracts/interfaces
# Adjust imports in IBTC.sol
sed -i.bak 's|../erc20/|./solidity/|g' ./contracts/interfaces/IBTC.sol
rm -f ./contracts/interfaces/IBTC.sol.bak

# Copy MEZO Token interface
cp ../../precompile/mezotoken/IMEZO.sol ./contracts/interfaces
# Adjust imports in IMEZO.sol
sed -i.bak 's|../erc20/|./solidity/|g' ./contracts/interfaces/IMEZO.sol
rm -f ./contracts/interfaces/IMEZO.sol.bak

# Copy Validator Pool interface
cp ../../precompile/validatorpool/IValidatorPool.sol ./contracts/interfaces
# Copy Asset Bridge interface
cp ../../precompile/assetsbridge/IAssetsBridge.sol ./contracts/interfaces

# Install dependencies
npm i

# Build the test file list. When no arguments are given, run everything.
if [ $# -eq 0 ]; then
  TEST_FILES="./test/*.test.ts"
else
  TEST_FILES=""
  for suite in "$@"; do
    TEST_FILES="$TEST_FILES ./test/${suite}.test.ts"
  done
fi

# Run the system tests
npx hardhat test $TEST_FILES --network "$NETWORK"
