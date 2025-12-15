// SPDX-License-Identifier: GPL-3.0

pragma solidity 0.8.29;

import {TransparentUpgradeableProxy} from "@openzeppelin/contracts/proxy/transparent/TransparentUpgradeableProxy.sol";

import "./xMEZO.sol";

/// @title Cross-chain MEZO Deployer
/// @notice xMEZODeployer allows to deploy MEZO token to a foreign EVM chain using a stable address.
///         The xMEZODeployer should be deployed to the chain using the EIP2470 singleton factory.
/// @dev Based on: https://github.com/mezo-org/musd/blob/main/solidity/contracts/token/TokenDeployer.sol
contract xMEZODeployer {
    bytes32 public constant SALT =
        keccak256("Bank on yourself. Bring everyday finance to your Bitcoin.");

    /// @notice The governance address receiving the control over the token;
    /// @dev This is the same multisig as the one used to control Mezo contracts
    ///      upgradeability and some protocol parameters of the chain.
    address public constant GOVERNANCE = 0x98D8899c3030741925BE630C710A98B57F397C7a;    

    /// @notice The address of the deployed MEZO token contract.
    /// @dev Zero address before the contract is deployed.
    address public token;

    event TokenDeployed(address token);

    error TokenAlreadyDeployed();

    /// @notice Deploys the MEZO token to the chain via create2, initializes it,
    ///         and initiates the 2-step ownership transfer process to the governance.
    /// @dev IMPORTANT NOTES: 
    ///      - This method can be called by anyone
    ///      - The initial owner of the deployed token is this deployer contract
    ///      - The initial minter of the deployed token is the governance address
    ///      - The governance must finalize the 2-step ownership transfer process 
    ///        by issuing an acceptance transaction
    ///      - The governance must set the minter according to the needs
    function deployToken() external {
        if (token != address(0)) {
            revert TokenAlreadyDeployed();
        }

        // Deploy the implementation contract using CREATE2 and prepare the initialization data.
        // Desipite the fact that implementations may change over time, CREATE2 is used here for 
        // consistency to ensure the same initial implementation address for all chains.
        // Note that {salt: SALT} is syntactic sugar assembling a CREATE2 call under the hood.
        xMEZO impl = new xMEZO{salt: SALT}();
        bytes memory initData = abi.encodeWithSelector(
            impl.initialize.selector,
            "MEZO",
            "MEZO",
            18,
            GOVERNANCE // Set governance as initial minter.
        );

        // Deploy the transparent proxy contract using CREATE2. Using CREATE2 here is crucial
        // as the proxy address is the primary identifier for the token and won't change over time.
        TransparentUpgradeableProxy proxy = new TransparentUpgradeableProxy{salt: SALT}(
            address(impl),
            GOVERNANCE, // Set governance as ProxyAdmin owner.
            initData
        );

        // Store the address of the deployed proxy contract.
        token = address(proxy);

        emit TokenDeployed(token);
        
        // Initiate the 2-step ownership transfer process to the governance.
        xMEZO(token).transferOwnership(GOVERNANCE);
    }
}
