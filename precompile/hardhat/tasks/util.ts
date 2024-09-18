import { task } from 'hardhat/config'
import '@nomicfoundation/hardhat-toolbox'

task('util:getCode', 'Returns the evm bytecode associated with the given address')
  .addParam('address', 'The smart contract/precompile address')
  .setAction(async (taskArguments, hre) => {
    console.log(await hre.ethers.provider.getCode(taskArguments.address, 'latest'))
  })
