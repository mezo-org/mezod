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
const EvmByteCode = "608060405234801561000f575f80fd5b50600436106100cb575f3560e01c80633644e5151161008857806395d89b411161006357806395d89b4114610189578063a9059cbb14610191578063d505accf146101a4578063dd62ed3e146101b7575f80fd5b80633644e5151461015b57806370a082311461016357806370ae92d214610176575f80fd5b806306fdde03146100cf578063095ea7b3146100ed57806318160ddd1461011057806323b872dd1461012657806330adf81f14610139578063313ce56714610141575b5f80fd5b6100d76101ca565b6040516100e491906106eb565b60405180910390f35b6101006100fb366004610738565b610238565b60405190151581526020016100e4565b6101186102b5565b6040519081526020016100e4565b610100610134366004610760565b61031b565b6101186103a0565b6101496103e2565b60405160ff90911681526020016100e4565b610118610448565b610118610171366004610799565b61048a565b610118610184366004610799565b6104fe565b6100d7610531565b61010061019f366004610738565b610573565b6101006101b23660046107c3565b6105ad565b6101186101c536600461082b565b610654565b6060611edf60921b6001600160a01b03166306fdde036040518163ffffffff1660e01b81526004015f60405180830381865afa15801561020c573d5f803e3d5ffd5b505050506040513d5f823e601f3d908101601f191682016040526102339190810190610870565b905090565b60405163095ea7b360e01b81526001600160a01b0383166004820152602481018290525f90611edf60921b9063095ea7b3906044015b6020604051808303815f875af115801561028a573d5f803e3d5ffd5b505050506040513d601f19601f820116820180604052508101906102ae9190610918565b9392505050565b5f611edf60921b6001600160a01b03166318160ddd6040518163ffffffff1660e01b8152600401602060405180830381865afa1580156102f7573d5f803e3d5ffd5b505050506040513d601f19601f820116820180604052508101906102339190610937565b6040516323b872dd60e01b81526001600160a01b03808516600483015283166024820152604481018290525f90611edf60921b906323b872dd906064016020604051808303815f875af1158015610374573d5f803e3d5ffd5b505050506040513d601f19601f820116820180604052508101906103989190610918565b949350505050565b5f611edf60921b6001600160a01b03166330adf81f6040518163ffffffff1660e01b8152600401602060405180830381865afa1580156102f7573d5f803e3d5ffd5b5f611edf60921b6001600160a01b031663313ce5676040518163ffffffff1660e01b8152600401602060405180830381865afa158015610424573d5f803e3d5ffd5b505050506040513d601f19601f82011682018060405250810190610233919061094e565b5f611edf60921b6001600160a01b0316633644e5156040518163ffffffff1660e01b8152600401602060405180830381865afa1580156102f7573d5f803e3d5ffd5b6040516370a0823160e01b81526001600160a01b03821660048201525f90611edf60921b906370a08231906024015b602060405180830381865afa1580156104d4573d5f803e3d5ffd5b505050506040513d601f19601f820116820180604052508101906104f89190610937565b92915050565b604051633857496960e11b81526001600160a01b03821660048201525f90611edf60921b906370ae92d2906024016104b9565b6060611edf60921b6001600160a01b03166395d89b416040518163ffffffff1660e01b81526004015f60405180830381865afa15801561020c573d5f803e3d5ffd5b60405163a9059cbb60e01b81526001600160a01b0383166004820152602481018290525f90611edf60921b9063a9059cbb9060440161026e565b60405163d505accf60e01b81526001600160a01b03808916600483015287166024820152604481018690526064810185905260ff8416608482015260a4810183905260c481018290525f90611edf60921b9063d505accf9060e4016020604051808303815f875af1158015610624573d5f803e3d5ffd5b505050506040513d601f19601f820116820180604052508101906106489190610918565b98975050505050505050565b604051636eb1769f60e11b81526001600160a01b038084166004830152821660248201525f90611edf60921b9063dd62ed3e90604401602060405180830381865afa1580156106a5573d5f803e3d5ffd5b505050506040513d601f19601f820116820180604052508101906102ae9190610937565b5f5b838110156106e35781810151838201526020016106cb565b50505f910152565b602081525f82518060208401526107098160408501602087016106c9565b601f01601f19169190910160400192915050565b80356001600160a01b0381168114610733575f80fd5b919050565b5f8060408385031215610749575f80fd5b6107528361071d565b946020939093013593505050565b5f805f60608486031215610772575f80fd5b61077b8461071d565b92506107896020850161071d565b9150604084013590509250925092565b5f602082840312156107a9575f80fd5b6102ae8261071d565b60ff811681146107c0575f80fd5b50565b5f805f805f805f60e0888a0312156107d9575f80fd5b6107e28861071d565b96506107f06020890161071d565b95506040880135945060608801359350608088013561080e816107b2565b9699959850939692959460a0840135945060c09093013592915050565b5f806040838503121561083c575f80fd5b6108458361071d565b91506108536020840161071d565b90509250929050565b634e487b7160e01b5f52604160045260245ffd5b5f60208284031215610880575f80fd5b815167ffffffffffffffff80821115610897575f80fd5b818401915084601f8301126108aa575f80fd5b8151818111156108bc576108bc61085c565b604051601f8201601f19908116603f011681019083821181831017156108e4576108e461085c565b816040528281528760208487010111156108fc575f80fd5b61090d8360208301602088016106c9565b979650505050505050565b5f60208284031215610928575f80fd5b815180151581146102ae575f80fd5b5f60208284031215610947575f80fd5b5051919050565b5f6020828403121561095e575f80fd5b81516102ae816107b256fea2646970667358221220c2f6371598a91ac86bb0ab656ef59c039c475baed01b08c9ad75207fdd5ab20f64736f6c63430008180033"
