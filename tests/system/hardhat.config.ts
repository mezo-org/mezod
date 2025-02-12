import { vars, HardhatUserConfig } from "hardhat/config"
import "@nomicfoundation/hardhat-toolbox"
import "hardhat-deploy"
import { ethers } from 'ethers'
import fs from 'fs'
import path from 'path'

// TODO: make it work with a single node
const BUILD_DIR = '../../.localnet/'
const COUNT = 4

function getPrivKeys (): string[] {
  const keys: string[] = []
  for (let i = 0; i < COUNT; i++) {
    const filePath = path.resolve(`${BUILD_DIR}node${i}/mezod/key_seed.json`)
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
      url: 'http://localhost:8545',
      chainId: 31611,
      accounts: getPrivKeys(),
      gas: 'auto'
    },
  }
};

export default config;
