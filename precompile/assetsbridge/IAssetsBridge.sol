// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

/// @title AssetsLocked
/// @notice Represents bridged assets.
struct AssetsLocked {
    uint256 sequenceNumber;
    address recipient;
    uint256 amount;
    address token;
}

/// @title ERC20TokenMapping
/// @notice Defines a mapping between an ERC20 token on the source chain
///         and on the Mezo chain.
struct ERC20TokenMapping {
    // Address of the ERC20 token on the source chain.
    address sourceToken;
    // Address of the ERC20 token on the Mezo chain.
    address mezoToken;
}

/// @title  IAssetsBridge
/// @notice Interface for the Assets Bridge precompile
interface IAssetsBridge {
    /**
     * @notice Emitted when a new ERC20 token mapping is created.
     * @param sourceToken The address of the ERC20 token on the source chain.
     * @param mezoToken The address of the ERC20 token on the Mezo chain.
     */
    event ERC20TokenMappingCreated(
        address indexed sourceToken,
        address indexed mezoToken
    );

    /**
     * @notice Emitted when an existing ERC20 token mapping is deleted.
     * @param sourceToken The address of the ERC20 token on the source chain.
     * @param mezoToken The address of the ERC20 token on the Mezo chain.
     */
    event ERC20TokenMappingDeleted(
        address indexed sourceToken,
        address indexed mezoToken
    );

    /**
     * @notice Emitted when an existing asset is unlocked on the Mezo side of the native bridge.
     * @param unlockSequenceNumber the sequence number for the specific AssetsUnlocked.
     * @param recipient The address it's bridged out to on the target chain.
     * @param token The address of the ERC20 token on the target chain.
     * @param sender The address bridging out.
     * @param amount The amount bridged out.
     * @param chain The chain to which the funds are being bridged out to, for reference
     *        please see the enum on the MezoBridge contract
     *        https://github.com/thesis/mezo-portal/blob/main/solidity/contracts/MezoBridge.sol#L22-L27
     */
    event AssetsUnlocked(
        uint256 indexed unlockSequenceNumber,
        bytes indexed recipient,
        address indexed token,
        address sender,
        uint256 amount,
        uint8 chain
    );

    /**
     * @notice Helper function used to enable bridged assets observability.
     */
    function bridge(AssetsLocked[] memory events) external returns (bool);

    /**
     * @notice Creates a new ERC20 token mapping.
     * @param sourceToken The address of the ERC20 token on the source chain.
     * @param mezoToken The address of the ERC20 token on the Mezo chain.
     * @dev Requirements:
     *      - The caller must be the contract owner,
     *      - The sourceToken address must not be the zero address,
     *      - The mezoToken address must not be the zero address,
     *      - The sourceToken address must not be already mapped,
     *      - The maximum number of mappings (getMaxERC20TokensMappings) must not be reached.
     */
    function createERC20TokenMapping(address sourceToken, address mezoToken) external returns (bool);

    /**
     * @notice Deletes an existing ERC20 token mapping.
     * @param sourceToken The address of the ERC20 token on the source chain.
     * @dev Requirements:
     *      - The caller must be the contract owner,
     *      - The source token address must correspond to an existing mapping.
     */
    function deleteERC20TokenMapping(address sourceToken) external returns (bool);

    /**
     * @notice Returns the ERC20 token mapping by source token address.
     * @param sourceToken The address of the ERC20 token on the source chain.
     * @return The ERC20 token mapping. If the source token is not mapped,
     *         the mapping will have both token addresses set to the zero address.
     */
    function getERC20TokenMapping(address sourceToken) external view returns (ERC20TokenMapping memory);

    /**
     * @notice Returns the list of all ERC20 token mappings supported by the bridge.
     * @return The list of ERC20 token mappings.
     */
    function getERC20TokensMappings() external view returns (ERC20TokenMapping[] memory);

    /**
     * @notice Returns the maximum number of ERC20 token mappings supported by the bridge.
     * @return The maximum number of ERC20 token mappings.
     */
    function getMaxERC20TokensMappings() external view returns (uint256);

    /**
     * @notice Returns the address of the BTC token on the source chain.
     * @dev AssetsLocked events carrying this token address are directly mapped
     *      to the Mezo native denomination - BTC.
     */
    function getSourceBTCToken() external view returns (address);

    /**
     * @notice Returns the current assets locked sequence tip of the bridge.
     * @return The current assets locked sequence tip of the bridge.
     */
    function getCurrentSequenceTip() external view returns (uint256);

    /**
     * @notice Initiates the bridge out process by unlocking the given assets on Mezo.
     * @param token The address of the ERC20 token on Mezo.
     * @param amount The amount of the ERC20 token to unlock.
     * @param chain The target chain to bridge out to.
     * @param recipient The target address to send the funds to.
              On Ethereum: recipient is a 20-byte EVM address
              On Bitcoin: recipient is a proper standard-type Bitcoin script
              supported by tBTC, i.e. P2PKH, P2WPKH, P2SH or P2WSH
     * @return True if the call succeeded, false otherwise.
     */
    function bridgeOut(address token, uint256 amount, uint8 chain, bytes calldata recipient) external returns (bool);
}
