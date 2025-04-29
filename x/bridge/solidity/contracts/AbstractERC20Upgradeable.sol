// SPDX-License-Identifier: MIT
pragma solidity 0.8.29;

import "@openzeppelin/contracts-upgradeable/token/ERC20/ERC20Upgradeable.sol";
import "@openzeppelin/contracts-upgradeable/token/ERC20/extensions/ERC20BurnableUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/access/Ownable2StepUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";

abstract contract AbstractERC20Upgradeable is Initializable, ERC20Upgradeable, ERC20BurnableUpgradeable, Ownable2StepUpgradeable {
    /// @dev The address of the minter.
    address public minter;

    /// @dev This empty reserved space is put in place to allow future versions to add new
    /// variables without shifting down storage in the inheritance chain.
    /// The convention from OpenZeppelin
    /// suggests the storage space should add up to 50 slots.
    /// See https://docs.openzeppelin.com/contracts/4.x/upgradeable#storage_gaps
    uint256[49] private __gap;

    /// @dev Emitted when the minter is changed.
    event MinterChanged(address indexed oldMinter, address indexed newMinter);

    /// @dev Throws if the minter is the zero address.
    error ZeroAddressMinter();

    /// @dev Throws if the caller is not the minter.
    error NotMinter();

    /// @custom:oz-upgrades-unsafe-allow constructor
    constructor() {
        _disableInitializers();
    }

    /// @dev Initializes the contract.
    function initialize(
        string memory name,
        string memory symbol,
        address initialMinter
    ) public initializer {
        __ERC20_init(name, symbol);
        __Ownable_init(_msgSender());
        
        if (initialMinter == address(0)) {
            revert ZeroAddressMinter();
        }

        minter = initialMinter;
        emit MinterChanged(address(0), initialMinter);
    }

    /// @dev Throws if called by any account other than the minter.
    modifier onlyMinter() {
        if (minter != _msgSender()) {
            revert NotMinter();
        }
        _;
    }

    /// @dev Set minter role to a new account.
    function setMinter(address newMinter) public onlyOwner {
        if (newMinter == address(0)) {
            revert ZeroAddressMinter();
        }
        address oldMinter = minter;
        minter = newMinter;
        emit MinterChanged(oldMinter, newMinter);
    }

    /// @dev Mints `amount` tokens and assigns them to `account`.
    function mint(address account, uint256 amount) public onlyMinter {
        _mint(account, amount);
    }
}