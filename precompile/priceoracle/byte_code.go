package priceoracle

// EvmByteCode is the EVM bytecode of the Price Oracle precompile. This code is
// returned by eth_getCode and ensures the precompile address is detected as a
// smart contract by external services. note: It should NOT contain a 0x prefix
//
// This bytecode was generated by compiling the PriceOracleCaller contract
// found in the precompile/hardhat package. Then extracting `deployedBytecode`
// from the build artifacts
const EvmByteCode = "608060405234801561000f575f80fd5b5060043610610034575f3560e01c8063313ce56714610038578063feaf968c14610057575b5f80fd5b610040610096565b60405160ff90911681526020015b60405180910390f35b61005f610104565b6040805169ffffffffffffffffffff968716815260208101959095528401929092526060830152909116608082015260a00161004e565b5f6015611edf60921b016001600160a01b031663313ce5676040518163ffffffff1660e01b8152600401602060405180830381865afa1580156100db573d5f803e3d5ffd5b505050506040513d601f19601f820116820180604052508101906100ff9190610182565b905090565b5f805f805f6015611edf60921b016001600160a01b031663feaf968c6040518163ffffffff1660e01b815260040160a060405180830381865afa15801561014d573d5f803e3d5ffd5b505050506040513d601f19601f8201168201806040525081019061017191906101c7565b945094509450945094509091929394565b5f60208284031215610192575f80fd5b815160ff811681146101a2575f80fd5b9392505050565b805169ffffffffffffffffffff811681146101c2575f80fd5b919050565b5f805f805f60a086880312156101db575f80fd5b6101e4866101a9565b9450602086015193506040860151925060608601519150610207608087016101a9565b9050929550929590935056fea26469706673582212208f4288cb3fc04c3fb45b488f5bfcf391ecedb7c2be14bcd046d1cc64130824d364736f6c63430008180033"
