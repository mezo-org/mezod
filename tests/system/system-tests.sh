#!/bin/sh
# This script runs the system tests for the Mezo blockchain.

# Create interfaces dir if it doesn't already exist
mkdir -p ./contracts/interfaces

# Copy BTC Token interfaces
cp ../../precompile/btctoken/IBTC.sol ./contracts/interfaces
cp -R ../../precompile/btctoken/solidity ./contracts/interfaces

# Install dependencies
npm i

# Run the system tests
npm run test