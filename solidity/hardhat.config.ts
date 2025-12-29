import { HardhatUserConfig } from "hardhat/config"
import "@keep-network/hardhat-helpers"
import "@nomicfoundation/hardhat-toolbox"
import "@openzeppelin/hardhat-upgrades"
import "hardhat-deploy"
import * as dotenv from "dotenv";

// Load .env file
dotenv.config();

function parseCommaDelimitedString(input: string): string[] {
  if (!input) return [];
  return input.split(',').map(item => item.trim());
}

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
    poaOwner: {
      default: 3,
      testnet: "0x0504d82Efb7Db7A8C05e8df8Cea575d8C9f48bb2",
      mainnet: "0x123694886DBf5Ac94DDA07135349534536D14cAf"
    }
  },
  defaultNetwork: 'hardhat',
  networks: {
    testnet: {
      chainId: 31611,
      url: process.env.TESTNET_RPC_URL || "",
      accounts: parseCommaDelimitedString(process.env.TESTNET_PRIVATE_KEY as string),
      tags: ['verify'],
      companionNetworks: {
        ethereum: 'ethereumSepolia'
      }
    },
    mainnet: {
      chainId: 31612,
      url: process.env.MAINNET_RPC_URL || "",
      accounts: parseCommaDelimitedString(process.env.MAINNET_PRIVATE_KEY as string),
      tags: ['verify'],
      companionNetworks: {
        ethereum: 'ethereumMainnet'
      }
    },
    ethereumSepolia: {
      url: process.env.ETHEREUM_SEPOLIA_RPC_URL || "",
      accounts: parseCommaDelimitedString(process.env.ETHEREUM_SEPOLIA_PRIVATE_KEY as string),
      tags: ['verify'],
      chainId: 11155111,
    },
    ethereumMainnet: {
      url: process.env.ETHEREUM_MAINNET_RPC_URL || "",
      accounts: parseCommaDelimitedString(process.env.ETHEREUM_MAINNET_PRIVATE_KEY as string),
      tags: ['verify'],
      chainId: 1,
    },
    baseSepolia: {
      url: process.env.BASE_SEPOLIA_RPC_URL || "",
      accounts: parseCommaDelimitedString(process.env.BASE_SEPOLIA_PRIVATE_KEY as string),
      tags: ['verify'],
      chainId: 84532,
    },
    baseMainnet: {
      url: process.env.BASE_MAINNET_RPC_URL || "",
      accounts: parseCommaDelimitedString(process.env.BASE_MAINNET_PRIVATE_KEY as string),
      tags: ['verify'],
      chainId: 8453,
    },
    bscMainnet: {
      url: process.env.BSC_MAINNET_RPC_URL || "",
      accounts: parseCommaDelimitedString(process.env.BSC_MAINNET_PRIVATE_KEY as string),
      tags: ['verify'],
      chainId: 56,
    },
  },
  external: {
    deployments: {
      testnet: ["./external/testnet"],
      mainnet: ["./external/mainnet"],
      ethereumSepolia: ["./external/ethereum-sepolia"],
      ethereumMainnet: ["./external/ethereum-mainnet"],
    },
  },
  // FIXME: Since the introduction of Etherscan v2 API, the hardhat-verify plugin
  //        requires a single API key for all networks. This is necessary to
  //        make Etherscan verification working. At the same time, this breaks
  //        the BlockScout verification for the Mezo networks. We can live
  //        with this for now but let's revisit and try to fix this next time
  //        we touch this code.
  etherscan: {
    apiKey: process.env.ETHERSCAN_API_KEY || "",
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
