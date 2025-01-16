import { task } from 'hardhat/config'
import '@nomicfoundation/hardhat-toolbox'

import abi from '../../priceoracle/abi.json'
const precompileAddress = '0x7b7c000000000000000000000000000000000015'

task('priceOracle:decimals', 'Decimal places of the precision used to represent the price', async (_, hre) => {
  const priceOracle = new hre.ethers.Contract(precompileAddress, abi, hre.ethers.provider)
  const response = await priceOracle.decimals()
  console.log(response)
})

task('priceOracle:latestRoundData', 'Returns the price from the last update tick (round) of the oracle', async (_, hre) => {
  const priceOracle = new hre.ethers.Contract(precompileAddress, abi, hre.ethers.provider)
  const response = await priceOracle.latestRoundData()
  console.log(response)
})
