package validatorpool

// EvmByteCode is the EVM bytecode of the ValidatorPool precompile. This code is
// returned by eth_getCode and ensures the precompile address is detected as a
// smart contract by external services. note: It should NOT contain a 0x prefix
//
// WARNING! This value is used in the InitChain ABCI hook and affects the app state hash.
// DO NOT change this value on a live chain, instead, use `setPrecompileByteCode`
// provided by the `Maintenance` precompile
//
// This bytecode was generated by compiling the ValidatorPoolCaller contract
// found in the precompile/hardhat package. Then extracting `deployedBytecode`
// from the build artifacts
const EvmByteCode = "608060405234801561000f575f80fd5b5060043610610106575f3560e01c80637ce5e82f1161009e578063c105ea2b1161006e578063c105ea2b14610210578063ca1e781914610218578063d66d9e1914610220578063e3ae4d0a14610228578063f2fde38b1461023b575f80fd5b80637ce5e82f146101c25780638bd2cc11146101ca5780638da5cb5b146101dd57806396c55175146101fd575f80fd5b8063617177f5116100d9578063617177f514610174578063644bb08b146101945780636a050f5f146101a757806379ba5097146101ba575f80fd5b8063223b3b7a1461010a5780632dea97ba14610134578063335762751461014957806354023a4e1461016c575b5f80fd5b61011d6101183660046108b0565b61024e565b60405161012b929190610918565b60405180910390f35b61013c6102fb565b60405161012b91906109b9565b61015c610157366004610a3e565b61036c565b604051901515815260200161012b565b61015c6103e5565b610187610182366004610abd565b61044f565b60405161012b9190610ad8565b61015c6101a2366004610a3e565b6104c2565b61015c6101b5366004610b24565b6104f3565b61015c610569565b6101876105af565b61011d6101d83660046108b0565b61061b565b6101e5610682565b6040516001600160a01b03909116815260200161012b565b61015c61020b3660046108b0565b6106eb565b6101e561075d565b6101876107a2565b61015c6107e7565b61015c6102363660046108b0565b61082d565b61015c6102493660046108b0565b610863565b5f6102816040518060a0016040528060608152602001606081526020016060815260200160608152602001606081525090565b60405163111d9dbd60e11b81526001600160a01b03841660048201526011611edf60921b019063223b3b7a906024015b5f60405180830381865afa1580156102cb573d5f803e3d5ffd5b505050506040513d5f823e601f3d908101601f191682016040526102f29190810190610c60565b91509150915091565b60606011611edf60921b016001600160a01b0316632dea97ba6040518163ffffffff1660e01b81526004015f60405180830381865afa158015610340573d5f803e3d5ffd5b505050506040513d5f823e601f3d908101601f191682016040526103679190810190610d83565b905090565b604051633357627560e01b81525f906011611edf60921b019063335762759061039d90879087908790600401610e75565b6020604051808303815f875af11580156103b9573d5f803e3d5ffd5b505050506040513d601f19601f820116820180604052508101906103dd9190610ece565b949350505050565b5f6011611edf60921b016001600160a01b03166354023a4e6040518163ffffffff1660e01b81526004016020604051808303815f875af115801561042b573d5f803e3d5ffd5b505050506040513d601f19601f820116820180604052508101906103679190610ece565b60405163617177f560e01b815260ff821660048201526060906011611edf60921b019063617177f5906024015f60405180830381865afa158015610495573d5f803e3d5ffd5b505050506040513d5f823e601f3d908101601f191682016040526104bc9190810190610eed565b92915050565b60405163644bb08b60e01b81525f906011611edf60921b019063644bb08b9061039d90879087908790600401610e75565b604051636a050f5f60e01b81525f906011611edf60921b0190636a050f5f906105229086908690600401610ff3565b6020604051808303815f875af115801561053e573d5f803e3d5ffd5b505050506040513d601f19601f820116820180604052508101906105629190610ece565b9392505050565b5f6011611edf60921b016001600160a01b03166379ba50976040518163ffffffff1660e01b81526004016020604051808303815f875af115801561042b573d5f803e3d5ffd5b60606011611edf60921b016001600160a01b0316637ce5e82f6040518163ffffffff1660e01b81526004015f60405180830381865afa1580156105f4573d5f803e3d5ffd5b505050506040513d5f823e601f3d908101601f191682016040526103679190810190610eed565b5f61064e6040518060a0016040528060608152602001606081526020016060815260200160608152602001606081525090565b604051638bd2cc1160e01b81526001600160a01b03841660048201526011611edf60921b0190638bd2cc11906024016102b1565b5f6011611edf60921b016001600160a01b0316638da5cb5b6040518163ffffffff1660e01b8152600401602060405180830381865afa1580156106c7573d5f803e3d5ffd5b505050506040513d601f19601f8201168201806040525081019061036791906110be565b6040516396c5517560e01b81526001600160a01b03821660048201525f906011611edf60921b01906396c55175906024015b6020604051808303815f875af1158015610739573d5f803e3d5ffd5b505050506040513d601f19601f820116820180604052508101906104bc9190610ece565b5f6011611edf60921b016001600160a01b031663c105ea2b6040518163ffffffff1660e01b8152600401602060405180830381865afa1580156106c7573d5f803e3d5ffd5b60606011611edf60921b016001600160a01b031663ca1e78196040518163ffffffff1660e01b81526004015f60405180830381865afa1580156105f4573d5f803e3d5ffd5b5f6011611edf60921b016001600160a01b031663d66d9e196040518163ffffffff1660e01b81526004016020604051808303815f875af115801561042b573d5f803e3d5ffd5b6040516371d7268560e11b81526001600160a01b03821660048201525f906011611edf60921b019063e3ae4d0a9060240161071d565b60405163f2fde38b60e01b81526001600160a01b03821660048201525f906011611edf60921b019063f2fde38b9060240161071d565b6001600160a01b03811681146108ad575f80fd5b50565b5f602082840312156108c0575f80fd5b813561056281610899565b5f5b838110156108e55781810151838201526020016108cd565b50505f910152565b5f81518084526109048160208601602086016108cb565b601f01601f19169290920160200192915050565b828152604060208201525f825160a0604084015261093960e08401826108ed565b90506020840151603f198085840301606086015261095783836108ed565b9250604086015191508085840301608086015261097483836108ed565b925060608601519150808584030160a086015261099183836108ed565b925060808601519150808584030160c0860152506109af82826108ed565b9695505050505050565b5f60208083018184528085518083526040925060408601915060408160051b8701018488015f5b83811015610a2257888303603f190185528151805160ff168452870151878401879052610a0f878501826108ed565b95880195935050908601906001016109e0565b509098975050505050505050565b60ff811681146108ad575f80fd5b5f805f60408486031215610a50575f80fd5b833567ffffffffffffffff80821115610a67575f80fd5b818601915086601f830112610a7a575f80fd5b813581811115610a88575f80fd5b8760208260051b8501011115610a9c575f80fd5b60209283019550935050840135610ab281610a30565b809150509250925092565b5f60208284031215610acd575f80fd5b813561056281610a30565b602080825282518282018190525f9190848201906040850190845b81811015610b185783516001600160a01b031683529284019291840191600101610af3565b50909695505050505050565b5f8060408385031215610b35575f80fd5b82359150602083013567ffffffffffffffff811115610b52575f80fd5b830160a08186031215610b63575f80fd5b809150509250929050565b634e487b7160e01b5f52604160045260245ffd5b60405160a0810167ffffffffffffffff81118282101715610ba557610ba5610b6e565b60405290565b6040805190810167ffffffffffffffff81118282101715610ba557610ba5610b6e565b604051601f8201601f1916810167ffffffffffffffff81118282101715610bf757610bf7610b6e565b604052919050565b5f82601f830112610c0e575f80fd5b815167ffffffffffffffff811115610c2857610c28610b6e565b610c3b601f8201601f1916602001610bce565b818152846020838601011115610c4f575f80fd5b6103dd8260208301602087016108cb565b5f8060408385031215610c71575f80fd5b82519150602083015167ffffffffffffffff80821115610c8f575f80fd5b9084019060a08287031215610ca2575f80fd5b610caa610b82565b825182811115610cb8575f80fd5b610cc488828601610bff565b825250602083015182811115610cd8575f80fd5b610ce488828601610bff565b602083015250604083015182811115610cfb575f80fd5b610d0788828601610bff565b604083015250606083015182811115610d1e575f80fd5b610d2a88828601610bff565b606083015250608083015182811115610d41575f80fd5b610d4d88828601610bff565b6080830152508093505050509250929050565b5f67ffffffffffffffff821115610d7957610d79610b6e565b5060051b60200190565b5f6020808385031215610d94575f80fd5b825167ffffffffffffffff80821115610dab575f80fd5b818501915085601f830112610dbe575f80fd5b8151610dd1610dcc82610d60565b610bce565b81815260059190911b83018401908481019088831115610def575f80fd5b8585015b83811015610e6857805185811115610e09575f80fd5b86016040818c03601f1901811315610e1f575f80fd5b610e27610bab565b89830151610e3481610a30565b8152908201519087821115610e47575f80fd5b610e558d8b84860101610bff565b818b015285525050918601918601610df3565b5098975050505050505050565b604080825281018390525f8460608301825b86811015610eb7578235610e9a81610899565b6001600160a01b0316825260209283019290910190600101610e87565b50809250505060ff83166020830152949350505050565b5f60208284031215610ede575f80fd5b81518015158114610562575f80fd5b5f6020808385031215610efe575f80fd5b825167ffffffffffffffff811115610f14575f80fd5b8301601f81018513610f24575f80fd5b8051610f32610dcc82610d60565b81815260059190911b82018301908381019087831115610f50575f80fd5b928401925b82841015610f77578351610f6881610899565b82529284019290840190610f55565b979650505050505050565b5f808335601e19843603018112610f97575f80fd5b830160208101925035905067ffffffffffffffff811115610fb6575f80fd5b803603821315610fc4575f80fd5b9250929050565b81835281816020850137505f828201602090810191909152601f909101601f19169091010190565b828152604060208201525f6110088384610f82565b60a0604085015261101d60e085018284610fcb565b91505061102d6020850185610f82565b603f1980868503016060870152611045848385610fcb565b93506110546040880188610f82565b935091508086850301608087015261106d848484610fcb565b935061107c6060880188610f82565b93509150808685030160a0870152611095848484610fcb565b93506110a46080880188610f82565b93509150808685030160c087015250610f77838383610fcb565b5f602082840312156110ce575f80fd5b81516105628161089956fea2646970667358221220fd957165b48cc4047cead21e98e2ef1a7c5d1444e4fc0e77516d2ae152d55ef164736f6c63430008180033"
