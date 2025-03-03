package testbeds

// EvmByteCode is the EVM bytecode of the BTC token precompile. This code is
// returned by eth_getCode and ensures the precompile address is detected as a
// smart contract by external services. note: It should NOT contain a 0x prefix
//
// WARNING! This value is used in the InitChain ABCI hook and affects the app state hash.
// DO NOT change this value on a live chain, instead, use `setPrecompileByteCode`
// provided by the `Maintenance` precompile
//
// This bytecode was generated by compiling the BTCCaller contract
// found in the precompile/hardhat package. Then extracting `deployedBytecode`
// from the build artifacts
const EvmByteCode = ""
