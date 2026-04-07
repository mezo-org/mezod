import { vars, HardhatUserConfig } from "hardhat/config"
import "@nomicfoundation/hardhat-toolbox"
import "hardhat-deploy"
import { ethers } from 'ethers'
import fs from 'fs'
import path from 'path'

const localnodeKeySeed = (index: number) => `../../.localnode/dev${index}_key_seed.json`

const KEY_SEEDS = [
  localnodeKeySeed(0),
  localnodeKeySeed(1),
  localnodeKeySeed(2),
]

function getPrivKeys (): string[] {
  const keys: string[] = []
  for (let i = 0; i < KEY_SEEDS.length; i++) {
    const filePath = path.resolve(KEY_SEEDS[i])
    const seed = JSON.parse(fs.readFileSync(filePath, 'utf8'))
    const pk: string = ethers.Wallet.fromPhrase(seed.secret).privateKey
    keys.push(pk)
  }

  return keys
}

const rpcUrl = process.env.RPC_URL || 'http://127.0.0.1:8545'
const chainId = parseInt(process.env.CHAIN_ID || '31611')

function getAccounts(): string[] {
  if (process.env.PRIVATE_KEYS) {
    return process.env.PRIVATE_KEYS.split(',')
  }
  return getPrivKeys()
}

const config: HardhatUserConfig = {
  solidity: {
    version: "0.8.28",
    settings: {
      evmVersion: "cancun",
    },
  },
  mocha: {
    timeout: 120000,
  },

  namedAccounts: {
    deployer: {
      default: 0,
    },
  },

  networks: {
    localhost: {
      url: rpcUrl,
      chainId: chainId,
      accounts: getAccounts(),
      gas: 'auto'
    },
  }
};

export default config;
