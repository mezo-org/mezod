import { task } from 'hardhat/config'
import '@nomicfoundation/hardhat-toolbox'

import abi from '../../assetsbridge/abi.json'
const precompileAddress = '0x7b7c000000000000000000000000000000000012'

task('assetsBridge:createERC20TokenMapping', 'Creates a new ERC20 token mapping')
  .addParam('sourceToken', 'The address of the ERC20 token on the source chain')
  .addParam('mezoToken', 'The address of the ERC20 token on the Mezo chain')
  .addParam('signer', 'The signer address (msg.sender)')
  .setAction(async (taskArguments, hre) => {
    const signer = await hre.ethers.getSigner(taskArguments.signer)
    const bridge = new hre.ethers.Contract(precompileAddress, abi, signer)
    const pending = await bridge.createERC20TokenMapping(
      taskArguments.sourceToken,
      taskArguments.mezoToken
    )
    const confirmed = await pending.wait()
    console.log(confirmed.hash)
  })

task('assetsBridge:deleteERC20TokenMapping', 'Deletes an existing ERC20 token mapping')
  .addParam('sourceToken', 'The address of the ERC20 token on the source chain')
  .addParam('signer', 'The signer address (msg.sender)')
  .setAction(async (taskArguments, hre) => {
    const signer = await hre.ethers.getSigner(taskArguments.signer)
    const bridge = new hre.ethers.Contract(precompileAddress, abi, signer)
    const pending = await bridge.deleteERC20TokenMapping(
      taskArguments.sourceToken
    )
    const confirmed = await pending.wait()
    console.log(confirmed.hash)
  })

task(
  'assetsBridge:getERC20TokenMapping',
  'Returns the ERC20 token mapping by source token address.'
)
  .addParam('sourceToken', 'The address of the ERC20 token on the source chain')
  .setAction(async (taskArguments, hre) => {
    const bridge = new hre.ethers.Contract(precompileAddress, abi, hre.ethers.provider)
    const result: string = await bridge.getERC20TokenMapping(taskArguments.sourceToken)
    console.log(result)
  })

task(
  'assetsBridge:getERC20TokensMappings',
  'Returns the list of all ERC20 token mappings supported by the bridge',
  async (_, hre) => {
    const bridge = new hre.ethers.Contract(precompileAddress, abi, hre.ethers.provider)
    const result: string = await bridge.getERC20TokensMappings()
    console.log(result)
  }
)

task(
  'assetsBridge:getMaxERC20TokensMappings',
  'Returns the maximum number of ERC20 token mappings supported by the bridge',
  async (_, hre) => {
    const bridge = new hre.ethers.Contract(precompileAddress, abi, hre.ethers.provider)
    const result: string = await bridge.getMaxERC20TokensMappings()
    console.log(result)
  }
)

task(
  'assetsBridge:getSourceBTCToken',
  'Returns the address of the BTC token on the source chain',
  async (_, hre) => {
    const bridge = new hre.ethers.Contract(precompileAddress, abi, hre.ethers.provider)
    const result: string = await bridge.getSourceBTCToken()
    console.log(result)
  }
)