import { task } from 'hardhat/config'
import '@nomicfoundation/hardhat-toolbox'

task('util:getCode', 'Returns the evm bytecode associated with the given address')
  .addParam('address', 'The smart contract/precompile address')
  .setAction(async (taskArguments, hre) => {
    console.log(await hre.ethers.provider.getCode(taskArguments.address, 'latest'))
  })

task('util:getERC20Supply', '')
  .setAction(async (taskArguments, hre) => {
    const erc20 = await hre.deployments.read('TestERC20', 'totalSupply')
    console.log(`total supply of TestERC20 is ${erc20}`)
  })