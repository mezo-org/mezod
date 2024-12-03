import { task } from 'hardhat/config'
import '@nomicfoundation/hardhat-toolbox'

import abi from '../../maintenance/abi.json'
const precompileAddress = '0x7b7c000000000000000000000000000000000013'

task('maintenance:getSupportNonEIP155Txs', 'Checks status of support for the non-EIP155 txs without replay protection', async (taskArguments, hre) => {
  const maintenance = new hre.ethers.Contract(precompileAddress, abi, hre.ethers.provider)
  const value = await maintenance.getSupportNonEIP155Txs()
  console.log(value)
})

task('maintenance:setSupportNonEIP155Txs', 'Enables/disables support for the non-EIP155 txs without replay protection')
  .addParam('signer', 'The signer address (msg.sender)')
  .addParam('value', 'The new value of the flag')
  .setAction(async (taskArguments, hre) => {
    const signer = await hre.ethers.getSigner(taskArguments.signer)
    const maintenance = new hre.ethers.Contract(precompileAddress, abi, signer)
    const pending = await maintenance.setSupportNonEIP155Txs(taskArguments.value === "true")
    const confirmed = await pending.wait()
    console.log(confirmed.hash)
  })