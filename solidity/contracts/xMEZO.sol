// SPDX-License-Identifier: LGPL-3.0
//
//  __       __  ________  ________   ______  
// |  \     /  \|        \|        \ /      \ 
// | $$\   /  $$| $$$$$$$$ \$$$$$$$$|  $$$$$$\
// | $$$\ /  $$$| $$__        /  $$ | $$  | $$
// | $$$$\  $$$$| $$  \      /  $$  | $$  | $$
// | $$\$$ $$ $$| $$$$$     /  $$   | $$  | $$
// | $$ \$$$| $$| $$_____  /  $$___ | $$__/ $$
// | $$  \$ | $$| $$     \|  $$    \ \$$    $$
//  \$$      \$$ \$$$$$$$$ \$$$$$$$$  \$$$$$$ 
                                           
pragma solidity 0.8.29;

import "./mERC20.sol";

// @title xMEZO
// @notice Cross-chain representation of the MEZO token.
// @dev MEZO token is a native precompile of the Mezo chain. 
//      This contract serves as a representation of the MEZO 
//      token on foreign EVM chains for bridging purposes.
contract xMEZO is mERC20 {}
