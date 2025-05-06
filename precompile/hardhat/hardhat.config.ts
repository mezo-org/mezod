import { vars, HardhatUserConfig } from 'hardhat/config'
import { ethers } from 'ethers'
import '@nomicfoundation/hardhat-toolbox'
import * as dotenv from "dotenv";
// import precompile tasks
import './tasks/validatorpool'
import './tasks/btctoken'
import './tasks/mezotoken'
import './tasks/util'
import './tasks/maintenance'
import './tasks/upgrade'
import './tasks/priceoracle'
import './tasks/assetsbridge'
import fs from 'fs'
import path from 'path'

const BUILD_DIR = '../../.localnet/'
const COUNT = 4

function getPrivKeys (): string[] {
  const strings: string[] = vars.get('MEZO_ACCOUNTS', '').split(',')
  const keys: string[] = []
  if (strings[0] !== '') {
    // Mezo accounts have been set already
    for (const str of strings) {
      if (str !== '') {
        keys.push(str)
      }
    }
  } else {
    for (let i = 0; i < COUNT; i++) {
      const filePath = path.resolve(`${BUILD_DIR}node${i}/mezod/key_seed.json`)
      const seed = JSON.parse(fs.readFileSync(filePath, 'utf8'))
      const pk: string = ethers.Wallet.fromPhrase(seed.secret).privateKey
      keys.push(pk)
    }
  }

  return keys
}

// Load .env file
dotenv.config();

const config: HardhatUserConfig = {
  solidity: {
    version: '0.8.24',
    settings: {
      optimizer: {
        enabled: true,
        runs: 200
      },
      evmVersion: 'cancun'
    }
  },
  defaultNetwork: 'localhost',
  networks: {
    localhost: {
      url: 'http://localhost:8545',
      chainId: 31611,
      accounts: getPrivKeys(),
      gas: 'auto'
    },
    testnet: {
      chainId: 31611,
      url: process.env.TESTNET_RPC_URL || "",
      accounts: process.env.TESTNET_PRIVATE_KEY ? [process.env.TESTNET_PRIVATE_KEY] : [],
    },
    mainnet: {
      chainId: 31612,
      url: process.env.MAINNET_RPC_URL || "",
      accounts: process.env.MAINNET_PRIVATE_KEY ? [process.env.MAINNET_PRIVATE_KEY] : [],
    }
  }
}

export default config
