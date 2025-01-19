package upgrade

// EvmByteCode is the EVM bytecode of the upgrade precompile. This code is
// returned by eth_getCode and ensures the precompile address is detected as a
// smart contract by external services. note: It should NOT contain a 0x prefix

// WARNING! This value is used in the InitChain ABCI hook and affects the app state hash.
// DO NOT change this value on a live chain, instead, use `setPrecompileByteCode`
// provided by the `Maintenance` precompile

// This bytecode was generated by compiling the UpgradeCaller contract
// found in the precompile/hardhat package. Then extracting `deployedBytecode`
// from the build artifacts
const EvmByteCode = "608060405234801561000f575f80fd5b506004361061003f575f3560e01c806353e0657114610043578063d8b13f2214610063578063f75b4ddd14610086575b5f80fd5b61004b61008e565b60405161005a93929190610243565b60405180910390f35b6100766100713660046102c6565b610108565b604051901515815260200161005a565b610076610187565b60605f60606014611edf60921b016001600160a01b03166353e065716040518163ffffffff1660e01b81526004015f60405180830381865afa1580156100d6573d5f803e3d5ffd5b505050506040513d5f823e601f3d908101601f191682016040526100fd91908101906103d4565b925092509250909192565b604051636c589f9160e11b81525f906014611edf60921b019063d8b13f229061013d908990899089908990899060040161046e565b6020604051808303815f875af1158015610159573d5f803e3d5ffd5b505050506040513d601f19601f8201168201806040525081019061017d91906104a9565b9695505050505050565b5f6014611edf60921b016001600160a01b031663f75b4ddd6040518163ffffffff1660e01b81526004016020604051808303815f875af11580156101cd573d5f803e3d5ffd5b505050506040513d601f19601f820116820180604052508101906101f191906104a9565b905090565b5f5b838110156102105781810151838201526020016101f8565b50505f910152565b5f815180845261022f8160208601602086016101f6565b601f01601f19169290920160200192915050565b606081525f6102556060830186610218565b8460070b6020840152828103604084015261017d8185610218565b5f8083601f840112610280575f80fd5b50813567ffffffffffffffff811115610297575f80fd5b6020830191508360208285010111156102ae575f80fd5b9250929050565b8060070b81146102c3575f80fd5b50565b5f805f805f606086880312156102da575f80fd5b853567ffffffffffffffff808211156102f1575f80fd5b6102fd89838a01610270565b909750955060208801359150610312826102b5565b90935060408701359080821115610327575f80fd5b5061033488828901610270565b969995985093965092949392505050565b634e487b7160e01b5f52604160045260245ffd5b5f82601f830112610368575f80fd5b815167ffffffffffffffff8082111561038357610383610345565b604051601f8301601f19908116603f011681019082821181831017156103ab576103ab610345565b816040528381528660208588010111156103c3575f80fd5b61017d8460208301602089016101f6565b5f805f606084860312156103e6575f80fd5b835167ffffffffffffffff808211156103fd575f80fd5b61040987838801610359565b94506020860151915061041b826102b5565b60408601519193508082111561042f575f80fd5b5061043c86828701610359565b9150509250925092565b81835281816020850137505f828201602090810191909152601f909101601f19169091010190565b606081525f610481606083018789610446565b8560070b6020840152828103604084015261049d818587610446565b98975050505050505050565b5f602082840312156104b9575f80fd5b815180151581146104c8575f80fd5b939250505056fea26469706673582212203f7675494375cfc0cc9c94ecc8123994aa290d55a203940346c27d2ac0d80b3564736f6c63430008180033"
