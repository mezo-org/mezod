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

task('maintenance:setPrecompileByteCode', 'Updates the byte code associated with a precompile')
  .addParam('signer', 'The signer address (msg.sender)')
  .addParam('precompile', 'The precompile contract address')
  .addParam('code', 'The new byte code')
  .setAction(async (taskArguments, hre) => {
    const signer = await hre.ethers.getSigner(taskArguments.signer)
    const maintenance = new hre.ethers.Contract(precompileAddress, abi, signer)
    const pending = await maintenance.setPrecompileByteCode(taskArguments.precompile, taskArguments.code)
    const confirmed = await pending.wait()
    console.log(confirmed)
  })

task('maintenance:getPrecompileByteCode', '')
  .addParam('precompile', 'The precompile contract address')
  .setAction(async (taskArguments, hre) => {
    const code = await hre.ethers.provider.getCode(taskArguments.precompile)
    console.log(code)
  })

task('maintenance:setChainFeeSplitterAddress', 'Sets the chain fee splitter address')
  .addParam('signer', 'The owner address (msg.sender)')
  .addParam('address', 'Address of the chain fee splitter contract')
  .setAction(async (taskArguments, hre) => {
    const signer = await hre.ethers.getSigner(taskArguments.signer)
    const maintenance = new hre.ethers.Contract(precompileAddress, abi, signer)
    const pending = await maintenance.setChainFeeSplitterAddress(taskArguments.address)
    const confirmed = await pending.wait()
    console.log(confirmed.hash)
  })

task('maintenance:getChainFeeSplitterAddress', 'Gets the chain fee splitter address')
  .setAction(async (taskArguments, hre) => {
    const maintenance = new hre.ethers.Contract(precompileAddress, abi, hre.ethers.provider)
    const address = await maintenance.getChainFeeSplitterAddress()
    console.log(address)
  })