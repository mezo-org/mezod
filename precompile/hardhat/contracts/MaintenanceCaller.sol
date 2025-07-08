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
}