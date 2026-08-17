// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import { IMaintenance } from "../interfaces/IMaintenance.sol";

contract MaintenanceCaller is IMaintenance {
    address private constant maintenancePrecompile = 0x7B7C000000000000000000000000000000000013;

    function getSupportNonEIP155Txs() external view returns (bool) {
        return IMaintenance(maintenancePrecompile).getSupportNonEIP155Txs();
    }

    function setSupportNonEIP155Txs(bool value) external returns (bool) {
        return IMaintenance(maintenancePrecompile).setSupportNonEIP155Txs(value);
    }

    function setPrecompileByteCode(address precompile, bytes calldata code) external returns (bool) {
        return IMaintenance(maintenancePrecompile).setPrecompileByteCode(precompile, code);
    }

    function setChainFeeSplitterAddress(address chainFeeSplitterAddress) external returns (bool) {
        return IMaintenance(maintenancePrecompile).setChainFeeSplitterAddress(chainFeeSplitterAddress);
    }

    function getChainFeeSplitterAddress() external view returns (address) {
        return IMaintenance(maintenancePrecompile).getChainFeeSplitterAddress();
    }

    function setMinGasPrice(uint256 minGasPrice) external returns (bool) {
        return IMaintenance(maintenancePrecompile).setMinGasPrice(minGasPrice);
    }

    function getMinGasPrice() external view returns (uint256) {
        return IMaintenance(maintenancePrecompile).getMinGasPrice();
    }

    function setMaxPrecompilesCallsPerExecution(uint32 value) external returns (bool) {
        return IMaintenance(maintenancePrecompile).setMaxPrecompilesCallsPerExecution(value);
    }

    function getMaxPrecompilesCallsPerExecution() external view returns (uint32) {
        return IMaintenance(maintenancePrecompile).getMaxPrecompilesCallsPerExecution();
    }

    function setSelfDestructDisabled(bool disabled) external returns (bool) {
        return IMaintenance(maintenancePrecompile).setSelfDestructDisabled(disabled);
    }

    function getSelfDestructDisabled() external view returns (bool) {
        return IMaintenance(maintenancePrecompile).getSelfDestructDisabled();
    }

    function setEmergencyTeam(address team) external returns (bool) {
        return IMaintenance(maintenancePrecompile).setEmergencyTeam(team);
    }

    function getEmergencyTeam() external view returns (address team) {
        return IMaintenance(maintenancePrecompile).getEmergencyTeam();
    }

    function setBridgeLockdown(bool bridgeIn, bool bridgeOut) external returns (bool) {
        return IMaintenance(maintenancePrecompile).setBridgeLockdown(bridgeIn, bridgeOut);
    }

    function getBridgeLockdown() external view returns (bool bridgeIn, bool bridgeOut) {
        return IMaintenance(maintenancePrecompile).getBridgeLockdown();
    }

    function setTxLockdown(bool enabled) external returns (bool) {
        return IMaintenance(maintenancePrecompile).setTxLockdown(enabled);
    }

    function getTxLockdown() external view returns (bool enabled) {
        return IMaintenance(maintenancePrecompile).getTxLockdown();
    }

    function setTxLockdownAllowlist(
        address[] calldata senders,
        address[] calldata targets
    ) external returns (bool) {
        return IMaintenance(maintenancePrecompile).setTxLockdownAllowlist(senders, targets);
    }

    function getTxLockdownAllowlist()
        external
        view
        returns (address[] memory senders, address[] memory targets)
    {
        return IMaintenance(maintenancePrecompile).getTxLockdownAllowlist();
    }

    function setChainLockdown() external returns (bool) {
        return IMaintenance(maintenancePrecompile).setChainLockdown();
    }
}
