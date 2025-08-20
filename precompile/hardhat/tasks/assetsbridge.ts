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
  'assetsBridge:getCurrentSequenceTip',
  'Returns the current assets lock sequence tip of the bridge',
  async (_, hre) => {
    const bridge = new hre.ethers.Contract(precompileAddress, abi, hre.ethers.provider)
    const result: string = await bridge.getCurrentSequenceTip()
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

task(
  'assetsBridge:bridgeOut', 
  'Initiates the bridge out process by unlocking the given assets on Mezo'
)
  .addParam('token', 'The address of the ERC20 token on Mezo')
  .addParam('amount', 'The amount of the ERC20 token to unlock')
  .addParam('chain', 'The target chain to bridge out to (uint8)')
  .addParam('recipient', 'The target address to send the funds to (hex encoded bytes)')
  .addParam('signer', 'The signer address (msg.sender)')
  .setAction(
    async (taskArguments, hre) => {
      const signer = await hre.ethers.getSigner(taskArguments.signer)
      const bridge = new hre.ethers.Contract(precompileAddress, abi, signer)
      const pending = await bridge.bridgeOut(
        taskArguments.token,
        taskArguments.amount,
        taskArguments.chain,
        taskArguments.recipient
      )
      const confirmed = await pending.wait()
      console.log(confirmed.hash)
    } 
  )

task('assetsBridge:setOutflowLimit', 'Sets the outflow limit for a specific token')
  .addParam('token', 'The address of the token to set the limit for')
  .addParam('limit', 'The maximum amount that can be bridged out in a 25,000 block window (set to 0 to remove limit)')
  .addParam('signer', 'The signer address (msg.sender) - must be PoA owner')
  .setAction(async (taskArguments, hre) => {
    const signer = await hre.ethers.getSigner(taskArguments.signer)
    const bridge = new hre.ethers.Contract(precompileAddress, abi, signer)
    const pending = await bridge.setOutflowLimit(
      taskArguments.token,
      taskArguments.limit
    )
    const confirmed = await pending.wait()
    console.log(confirmed.hash)
  })

task(
  'assetsBridge:getOutflowLimit',
  'Gets the current outflow limit for a specific token'
)
  .addParam('token', 'The address of the token to check the limit for')
  .setAction(async (taskArguments, hre) => {
    const bridge = new hre.ethers.Contract(precompileAddress, abi, hre.ethers.provider)
    const result: string = await bridge.getOutflowLimit(taskArguments.token)
    console.log(result)
  })

task(
  'assetsBridge:getOutflowCapacity',
  'Gets the outflow capacity for a specific token'
)
  .addParam('token', 'The address of the token to check the capacity for')
  .setAction(async (taskArguments, hre) => {
    const bridge = new hre.ethers.Contract(precompileAddress, abi, hre.ethers.provider)
    const result = await bridge.getOutflowCapacity(taskArguments.token)
    console.log('capacity:', result[0].toString())
    console.log('reset height:', result[1].toString())
  })
