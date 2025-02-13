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

const config: HardhatUserConfig = {
  solidity: "0.8.28",

  namedAccounts: {
    deployer: {
      default: 0,
    },
  },

  networks: {
    localhost: {
      url: 'http://127.0.0.1:8545', // localnode listens on this specific interface
      chainId: 31611,
      accounts: getPrivKeys(),
      gas: 'auto'
    },
  }
};

export default config;
