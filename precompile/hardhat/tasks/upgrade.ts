import { task } from 'hardhat/config'
import '@nomicfoundation/hardhat-toolbox'

import abi from '../../upgrade/abi.json'
const precompileAddress = '0x7b7c000000000000000000000000000000000014'

task('upgrade:plan', 'Returns the current upgrade plan', async (taskArguments, hre) => {
  const upgrade = new hre.ethers.Contract(precompileAddress, abi, hre.ethers.provider)
  const plan: string = await upgrade.plan()
  console.log(plan)
})

task('upgrade:cancelPlan', 'Cancels the current upgrade plan (owner)')
  .addParam('signer', 'The signer address (msg.sender)')
  .setAction(async (taskArguments, hre) => {
    const signer = await hre.ethers.getSigner(taskArguments.signer)
    const upgrade = new hre.ethers.Contract(precompileAddress, abi, signer)
    const pending = await upgrade.cancelPlan()
    const confirmed = await pending.wait()
    console.log(confirmed.hash)
  })

task('upgrade:submitPlan', 'Submits a new upgrade plan (owner)')
  .addParam('signer', 'The signer address (msg.sender)')
  .addParam('name', 'The name of the upgrade plan')
  .addParam('height', 'The block height to activate the upgrade')
  .addParam('info', 'The upgrade plan information')
  .setAction(async (taskArguments, hre) => {
    const signer = await hre.ethers.getSigner(taskArguments.signer)
    const upgrade = new hre.ethers.Contract(precompileAddress, abi, signer)
    const pending = await upgrade.submitPlan(taskArguments.name, taskArguments.height, taskArguments.info)
    const confirmed = await pending.wait()
    console.log(confirmed.hash)
  })