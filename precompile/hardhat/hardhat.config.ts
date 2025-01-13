import { vars, HardhatUserConfig } from 'hardhat/config'
import { ethers } from 'ethers'
import '@nomicfoundation/hardhat-toolbox'
import "hardhat-deploy"
// import precompile tasks
import './tasks/validatorpool'
import './tasks/btctoken'
import './tasks/util'
import './tasks/maintenance'
import fs from 'fs'
import path from 'path'

const MEZO_LOCALNET_DIR = process.env.MEZO_LOCALNET_DIR || "/Users/lukasz-zimnoch/go/src/github.com/thesis/mezo/.localnet"
const COUNT = 4

function getPrivKeys (): string[] {
  const keys: string[] = []

  for (let i = 0; i < COUNT; i++) {
    const filePath = path.resolve(`${MEZO_LOCALNET_DIR}/node${i}/mezod/key_seed.json`)
    const seed = JSON.parse(fs.readFileSync(filePath, 'utf8'))
    const pk: string = ethers.Wallet.fromPhrase(seed.secret).privateKey
    keys.push(pk)
  }

  return keys
}

const config: HardhatUserConfig = {
  solidity: {
    version: '0.8.24',
    settings: {
      optimizer: {
        enabled: true,
        runs: 200
      },
      evmVersion: 'london'
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
    mezo_testnet: {
      url: 'http://mezo-node-0.test.mezo.org:8545',
      chainId: 31611,
      accounts: getPrivKeys()
    }
  },
  namedAccounts: {
    deployer: 0,
  },
}

export default config
