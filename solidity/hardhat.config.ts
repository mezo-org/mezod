import { HardhatUserConfig } from "hardhat/config"
import "@keep-network/hardhat-helpers"
import "@nomicfoundation/hardhat-toolbox"
import "@openzeppelin/hardhat-upgrades"
import "hardhat-deploy"
import * as dotenv from "dotenv";

// Load .env file
dotenv.config();

const config: HardhatUserConfig = {
  solidity: {
    version: '0.8.29',
    settings: {
      optimizer: {
        enabled: true,
        runs: 200
      },
      evmVersion: 'london'
    }
  },
  namedAccounts: {
    deployer: {
      default: 0,
    },
    governance: {
      default: 1,
      testnet: "0x6e80164ea60673d64d5d6228beb684a1274bb017", // testertesting.eth
      mainnet: "0x98D8899c3030741925BE630C710A98B57F397C7a"  // mezo multisig
    },
    minter: {
      default: 2,
      testnet: "0x17F29B073143D8cd97b5bBe492bDEffEC1C5feE5", // x/bridge module account
      mainnet: "0x17F29B073143D8cd97b5bBe492bDEffEC1C5feE5", // x/bridge module account`
    },
    
  },
  defaultNetwork: 'hardhat',
  networks: {
    testnet: {
      chainId: 31611,
      url: process.env.TESTNET_RPC_URL || "",
      accounts: process.env.TESTNET_PRIVATE_KEY ? [process.env.TESTNET_PRIVATE_KEY] : [],
      tags: ['verify'],
    },
    mainnet: {
      chainId: 31612,
      url: process.env.MAINNET_RPC_URL || "",
      accounts: process.env.MAINNET_PRIVATE_KEY ? [process.env.MAINNET_PRIVATE_KEY] : [],
      tags: ['verify'],
    }
  },
  etherscan: {
    apiKey: {
      'testnet': 'empty',
      'mainnet': 'empty'
    },
    customChains: [
      {
        network: "testnet",
        chainId: 31611,
        urls: {
          apiURL: "https://api.explorer.test.mezo.org/api",
          browserURL: "https://explorer.test.mezo.org"
        }
      },
      {
        network: "mainnet",
        chainId: 31612,
        urls: {
          apiURL: "https://api.explorer.mezo.org/api",
          browserURL: "https://explorer.mezo.org"
        }
      }
    ]
  }
};

export default config;
