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
  const [roundId, answer, startedAt, updatedAt, answeredInRound] = await priceOracle.latestRoundData()
  console.log({
    roundId: roundId.toString(),
    answer: answer.toString(),
    startedAt: new Date(parseInt(startedAt.toString(), 10) * 1000).toISOString(),
    updatedAt: new Date(parseInt(updatedAt.toString(), 10) * 1000).toISOString(),
    answeredInRound: answeredInRound.toString(),
  })
})
