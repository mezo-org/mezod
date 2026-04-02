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

task('assetsBridge:setPauser', 'Sets the pauser address for emergency bridge operations')
  .addParam('pauser', 'The address that will be able to pause bridge operations (can be 0x0 to remove the pauser)')
  .addParam('signer', 'The signer address (msg.sender) - must be PoA owner')
  .setAction(async (taskArguments, hre) => {
    const signer = await hre.ethers.getSigner(taskArguments.signer)
    const bridge = new hre.ethers.Contract(precompileAddress, abi, signer)
    const pending = await bridge.setPauser(taskArguments.pauser)
    const confirmed = await pending.wait()
    console.log(confirmed.hash)
  })

task(
  'assetsBridge:getPauser',
  'Gets the current pauser address',
  async (_, hre) => {
    const bridge = new hre.ethers.Contract(precompileAddress, abi, hre.ethers.provider)
    const result: string = await bridge.getPauser()
    console.log(result)
  }
)

task(
  'assetsBridge:pauseBridgeOut',
  'Pauses all bridge out operations by setting outflow limits to 0 for all supported tokens'
)
  .addParam('signer', 'The signer address (msg.sender) - must be the current pauser')
  .setAction(async (taskArguments, hre) => {
    const signer = await hre.ethers.getSigner(taskArguments.signer)
    const bridge = new hre.ethers.Contract(precompileAddress, abi, signer)
    const pending = await bridge.pauseBridgeOut()
    const confirmed = await pending.wait()
    console.log(confirmed.hash)

  })

task(
  'assetsBridge:setMinBridgeOutAmount',
  'Sets the minimum bridge-out amount for a Mezo token'
)
  .addParam('token', 'The address of the token on the Mezo chain')
  .addParam('amount', 'The new minimum amount')
  .addParam('signer', 'The signer address (msg.sender)')
  .setAction(async (taskArguments, hre) => {
    const signer = await hre.ethers.getSigner(taskArguments.signer)
    const bridge = new hre.ethers.Contract(precompileAddress, abi, signer)
    const pending = await bridge.setMinBridgeOutAmount(
      taskArguments.token,
      taskArguments.amount
    )
    const confirmed = await pending.wait()
    console.log(confirmed.hash)
  })

task(
  'assetsBridge:getMinBridgeOutAmount',
  'Returns the minimum bridge-out amount (if set) for a Mezo token'
)
  .addParam('token', 'The address of the token on the Mezo chain')
  .setAction(async (taskArguments, hre) => {
    const bridge = new hre.ethers.Contract(precompileAddress, abi, hre.ethers.provider)
    const minAmount = await bridge.getMinBridgeOutAmount(taskArguments.token)
    console.log(minAmount)
  })

task('assetsBridge:setMinBridgeOutAmountForBitcoinChain', 'Sets the minimum bridge-out amount for the Bitcoin chain')
  .addParam('amount', 'The new minimum amount for Bitcoin chain bridge-outs')
  .addParam('signer', 'The signer address (msg.sender) - must be PoA owner')
  .setAction(async (taskArguments, hre) => {
    const signer = await hre.ethers.getSigner(taskArguments.signer)
    const bridge = new hre.ethers.Contract(precompileAddress, abi, signer)
    const pending = await bridge.setMinBridgeOutAmountForBitcoinChain(
      taskArguments.amount
    )
    const confirmed = await pending.wait()
    console.log(confirmed.hash)
  })

task(
  'assetsBridge:getMinBridgeOutAmountForBitcoinChain',
  'Returns the minimum bridge-out amount for the Bitcoin chain',
  async (_, hre) => {
    const bridge = new hre.ethers.Contract(precompileAddress, abi, hre.ethers.provider)
    const minAmount = await bridge.getMinBridgeOutAmountForBitcoinChain()
    console.log(minAmount)
  }
)

task('assetsBridge:bridgeTriparty', 'Requests a triparty BTC mint through the bridge')
  .addParam('recipient', 'The address to receive the minted BTC')
  .addParam('amount', 'The amount of BTC to mint')
  .addOptionalParam('callbackData', 'Hex-encoded callback data (default: empty)', '0x')
  .addParam('signer', 'The signer address (msg.sender) - must be an allowed triparty controller')
  .setAction(async (taskArguments, hre) => {
    const signer = await hre.ethers.getSigner(taskArguments.signer)
    const bridge = new hre.ethers.Contract(precompileAddress, abi, signer)
    const pending = await bridge.bridgeTriparty(
      taskArguments.recipient,
      taskArguments.amount,
      taskArguments.callbackData
    )
    const confirmed = await pending.wait()
    console.log('tx:', confirmed.hash)
    console.log('requestId:', confirmed.logs?.[0]?.args?.[0]?.toString())
  })

task('assetsBridge:allowTripartyController', 'Allows or disallows a triparty controller address')
  .addParam('controller', 'The address of the triparty controller')
  .addParam('isAllowed', 'Whether the controller should be allowed (true/false)')
  .addParam('signer', 'The signer address (msg.sender) - must be PoA owner')
  .setAction(async (taskArguments, hre) => {
    const signer = await hre.ethers.getSigner(taskArguments.signer)
    const bridge = new hre.ethers.Contract(precompileAddress, abi, signer)
    const pending = await bridge.allowTripartyController(
      taskArguments.controller,
      taskArguments.isAllowed === 'true'
    )
    const confirmed = await pending.wait()
    console.log(confirmed.hash)
  })

task(
  'assetsBridge:isAllowedTripartyController',
  'Checks if an address is an allowed triparty controller'
)
  .addParam('controller', 'The address to check')
  .setAction(async (taskArguments, hre) => {
    const bridge = new hre.ethers.Contract(precompileAddress, abi, hre.ethers.provider)
    const result: boolean = await bridge.isAllowedTripartyController(taskArguments.controller)
    console.log(result)
  })

task('assetsBridge:pauseTriparty', 'Pauses or unpauses triparty bridging')
  .addParam('isPaused', 'Whether to pause triparty bridging (true/false)')
  .addParam('signer', 'The signer address (msg.sender) - must be the current pauser')
  .setAction(async (taskArguments, hre) => {
    const signer = await hre.ethers.getSigner(taskArguments.signer)
    const bridge = new hre.ethers.Contract(precompileAddress, abi, signer)
    const pending = await bridge.pauseTriparty(
      taskArguments.isPaused === 'true'
    )
    const confirmed = await pending.wait()
    console.log(confirmed.hash)
  })

task('assetsBridge:setTripartyBlockDelay', 'Sets the block delay for triparty mint execution')
  .addParam('delay', 'The number of blocks that must pass between mint request and execution')
  .addParam('signer', 'The signer address (msg.sender) - must be PoA owner')
  .setAction(async (taskArguments, hre) => {
    const signer = await hre.ethers.getSigner(taskArguments.signer)
    const bridge = new hre.ethers.Contract(precompileAddress, abi, signer)
    const pending = await bridge.setTripartyBlockDelay(taskArguments.delay)
    const confirmed = await pending.wait()
    console.log(confirmed.hash)
  })

task(
  'assetsBridge:getTripartyBlockDelay',
  'Gets the current block delay for triparty mint execution',
  async (_, hre) => {
    const bridge = new hre.ethers.Contract(precompileAddress, abi, hre.ethers.provider)
    const result = await bridge.getTripartyBlockDelay()
    console.log(result)
  }
)

task('assetsBridge:setTripartyLimits', 'Sets the triparty per-request and rolling window limits')
  .addParam('perRequestLimit', 'The maximum amount per triparty mint request')
  .addParam('windowLimit', 'The maximum amount in a rolling window')
  .addParam('signer', 'The signer address (msg.sender) - must be PoA owner')
  .setAction(async (taskArguments, hre) => {
    const signer = await hre.ethers.getSigner(taskArguments.signer)
    const bridge = new hre.ethers.Contract(precompileAddress, abi, signer)
    const pending = await bridge.setTripartyLimits(
      taskArguments.perRequestLimit,
      taskArguments.windowLimit
    )
    const confirmed = await pending.wait()
    console.log(confirmed.hash)
  })

task(
  'assetsBridge:getTripartyLimits',
  'Gets the current triparty per-request and rolling window limits',
  async (_, hre) => {
    const bridge = new hre.ethers.Contract(precompileAddress, abi, hre.ethers.provider)
    const result = await bridge.getTripartyLimits()
    console.log('perRequestLimit:', result[0].toString())
    console.log('windowLimit:', result[1].toString())
  }
)