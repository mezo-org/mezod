package upgrade

// EvmByteCode is the EVM bytecode of the upgrade precompile. This code is
// returned by eth_getCode and ensures the precompile address is detected as a
// smart contract by external services. note: It should NOT contain a 0x prefix

// This bytecode was generated by compiling the UpgradeCaller contract
// found in the precompile/hardhat package. Then extracting `deployedBytecode`
// from the build artifacts
const EvmByteCode = ""
