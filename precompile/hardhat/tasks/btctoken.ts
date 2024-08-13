import { task } from 'hardhat/config'
import '@nomicfoundation/hardhat-toolbox'

import abi from '../../btctoken/abi.json'
const precompileAddress = '0x7b7c000000000000000000000000000000000000'

task('btcToken:name', 'Returns the token name', async (taskArguments, hre, runSuper) => {
  const btctoken = new hre.ethers.Contract(precompileAddress, abi, hre.ethers.provider)
  const name = await btctoken.name()
  console.log(name)
})
