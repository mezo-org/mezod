// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

import {IAssetsBridge} from "./interfaces/IAssetsBridge.sol";
import {IBTC} from "./interfaces/IBTC.sol";

contract SimpleToken {
    mapping(address => uint256) private _balances;
    mapping(address => mapping(address => uint256)) private _allowances;

    uint256 private _totalSupply;
    string public name = "Simple Token";
    string public symbol = "SIMPLE";
    uint8 public decimals = 18;

    event Transfer(address indexed from, address indexed to, uint256 value);
    event Approval(address indexed owner, address indexed spender, uint256 value);

    // View functions for basic functionality
    function balanceOf(address account) public view returns (uint256) {
        return _balances[account];
    }

    function allowance(address owner, address spender) public view returns (uint256) {
        return _allowances[owner][spender];
    }

    function totalSupply() public view returns (uint256) {
        return _totalSupply;
    }

    // Mint function - creates new tokens
    function mint(address to, uint256 amount) public {
        require(to != address(0), "ERC20: mint to zero address");

        _totalSupply += amount;
        _balances[to] += amount;

        emit Transfer(address(0), to, amount);
    }

    // Approve function - allows spender to use owner's tokens
    function approve(address spender, uint256 amount) public returns (bool) {
        require(spender != address(0), "ERC20: approve to zero address");

        _allowances[msg.sender][spender] = amount;

        emit Approval(msg.sender, spender, amount);
        return true;
    }

    // TransferFrom function - transfers tokens on behalf of another address
    function transferFrom(address from, address to, uint256 amount) public returns (bool) {
        require(from != address(0), "ERC20: transfer from zero address");
        require(to != address(0), "ERC20: transfer to zero address");
        require(_balances[from] >= amount, "ERC20: transfer amount exceeds balance");
        require(_allowances[from][msg.sender] >= amount, "ERC20: transfer amount exceeds allowance");

        _balances[from] -= amount;
        _balances[to] += amount;
        _allowances[from][msg.sender] -= amount;

        emit Transfer(from, to, amount);
        return true;
    }

    function burnFrom(address from, uint256 amount) public {
        require(from != address(0), "ERC20: burn from zero address");
        require(_balances[from] >= amount, "ERC20: burn amount exceeds balance");
        require(_allowances[from][msg.sender] >= amount, "ERC20: burn amount exceeds allowance");

        _balances[from] -= amount;
        _totalSupply -= amount;
        _allowances[from][msg.sender] -= amount;

        emit Transfer(from, address(0), amount);
    }
}

/// @title BridgeOut
/// @notice Handles various bridgeOut scenarios for Mezo native asset bridge.
contract BridgeOut {
    // AssetsBridge precompile address on Mezo.
    address private constant bridgePrecompile = 0x7B7C000000000000000000000000000000000012;
    // BTC precompile address on Mezo
    address private constant btcPrecompile = 0x7b7C000000000000000000000000000000000000;

    function bridgeOutBTCToBitcoinSuccess(bytes calldata recipient, uint256 amount) external payable {
	bool okApprove = IBTC(btcPrecompile).approve(bridgePrecompile, amount);
        require(okApprove, "couldn't approve bridge for transferFrom");

	bool okBridgeOut = IAssetsBridge(bridgePrecompile).bridgeOut(btcPrecompile, amount, 1, recipient);
        require(okBridgeOut, "couldn't bridge out btc");
    }

    function bridgeOutBTCToBitcoinReverts(bytes calldata recipient, uint256 amount) external payable {
	bool okApprove = IBTC(btcPrecompile).approve(bridgePrecompile, amount);
        require(okApprove, "couldn't approve bridge for transferFrom");

	bool okBridgeOut = IAssetsBridge(bridgePrecompile).bridgeOut(btcPrecompile, amount, 1, recipient);
        require(okBridgeOut, "couldn't bridge out btc");

	// now just revert
	revert("revert triggered");
    }

    function bridgeOutBTCToEthereum(address recipient) external payable {
        // bool ok = IAssetsBridge(precompile).bridgeOut();
        // require(ok, "No balance to transfer");
    }

    function bridgeOutERC20ToEthereum(address recipient) external payable {
        // bool ok = IAssetsBridge(precompile).bridgeOut();
        // require(ok, "No balance to transfer");
    }

    function bridgeOutInvalidERC20ToEthereum(address recipient) external payable {
        // bool ok = IAssetsBridge(precompile).bridgeOut();
        // require(ok, "No balance to transfer");
    }

    function bridgeOutBTCNotEnoughBalance(address recipient) external payable {
        // bool ok = IAssetsBridge(precompile).bridgeOut();
        // require(ok, "No balance to transfer");
    }

    function bridgeOutBTCNotApproved(address recipient) external payable {
        // bool ok = IAssetsBridge(precompile).bridgeOut();
        // require(ok, "No balance to transfer");
    }

    function bridgeOutBTCNotEnoughApproved(address recipient) external payable {
        // bool ok = IAssetsBridge(precompile).bridgeOut();
        // require(ok, "No balance to transfer");
    }

    function bridgeOutERC20NotEnoughBalance(address recipient) external payable {
        // bool ok = IAssetsBridge(precompile).bridgeOut();
        // require(ok, "No balance to transfer");
    }

    function bridgeOutERC20NotEnoughApproved(address recipient) external payable {
        // bool ok = IAssetsBridge(precompile).bridgeOut();
        // require(ok, "No balance to transfer");
    }

    function bridgeOutERC20NotApproved(address recipient) external payable {
        // bool ok = IAssetsBridge(precompile).bridgeOut();
        // require(ok, "No balance to transfer");
    }
}
