import { task } from 'hardhat/config'
import '@nomicfoundation/hardhat-toolbox'

import abi from '../../validatorpool/abi.json'
const precompileAddress = '0x7b7c000000000000000000000000000000000011'

task('validatorPool:owner', 'Returns the current contract owner', async (taskArguments, hre, runSuper) => {
  const validatorPool = new hre.ethers.Contract(precompileAddress, abi, hre.ethers.provider)
  const owner: string = await validatorPool.owner()
  console.log(owner)
})

task('validatorPool:candidateOwner', 'Returns the current contract candidateOwner', async (taskArguments, hre, runSuper) => {
  const validatorPool = new hre.ethers.Contract(precompileAddress, abi, hre.ethers.provider)
  const candidateOwner: string = await validatorPool.candidateOwner()
  console.log(candidateOwner)
})

task('validatorPool:validators', 'Returns an array of operator addresses for current validators', async (taskArguments, hre, runSuper) => {
  const validatorPool = new hre.ethers.Contract(precompileAddress, abi, hre.ethers.provider)
  const validators: string[] = await validatorPool.validators()
  console.log(validators)
})

task('validatorPool:validator', "Returns a validator's consensus public key & description")
  .addParam('operator', "The validator's operator address")
  .setAction(async (taskArguments, hre, runSuper) => {
    const validatorPool = new hre.ethers.Contract(precompileAddress, abi, hre.ethers.provider)
    const validator = await validatorPool.validator(taskArguments.operator)
    console.log(validator)
  })

task('validatorPool:application', "Returns an application's consensus public key & description")
  .addParam('operator', "The application's operator address")
  .setAction(async (taskArguments, hre, runSuper) => {
    const validatorPool = new hre.ethers.Contract(precompileAddress, abi, hre.ethers.provider)
    const application = await validatorPool.application(taskArguments.operator)
    console.log(application)
  })

task('validatorPool:applications', 'Returns an array of operator addresses for current applications', async (taskArguments, hre, runSuper) => {
  const validatorPool = new hre.ethers.Contract(precompileAddress, abi, hre.ethers.provider)
  const candidates = await validatorPool.applications()
  console.log(candidates)
})

task('validatorPool:leave', 'Removes the signers validator from the pool')
  .addParam('signer', 'The signer address (msg.sender)')
  .setAction(async (taskArguments, hre, runSuper) => {
    const signer = await hre.ethers.getSigner(taskArguments.signer)
    const validatorPool = new hre.ethers.Contract(precompileAddress, abi, signer)
    const pending = await validatorPool.leave({ gasLimit: 46128n })
    const confirmed = await pending.wait()
    console.log(confirmed.hash)
  })

task('validatorPool:kick', 'Kicks a validator from the pool')
  .addParam('signer', 'The signer address (msg.sender)')
  .addParam('operator', 'The operator address of the validator to be kicked')
  .setAction(async (taskArguments, hre, runSuper) => {
    const signer = await hre.ethers.getSigner(taskArguments.signer)
    const validatorPool = new hre.ethers.Contract(precompileAddress, abi, signer)
    const pending = await validatorPool.kick(taskArguments.operator)
    const confirmed = await pending.wait()
    console.log(confirmed.hash)
  })

task('validatorPool:transferOwnership', 'Begins the ownership transfer flow (owner)')
  .addParam('signer', 'The signer address (msg.sender)')
  .addParam('to', 'Address to transfer ownership to')
  .setAction(async (taskArguments, hre, runSuper) => {
    const signer = await hre.ethers.getSigner(taskArguments.signer)
    const validatorPool = new hre.ethers.Contract(precompileAddress, abi, signer)
    const pending = await validatorPool.transferOwnership(taskArguments.to)
    const confirmed = await pending.wait()
    console.log(confirmed.hash)
  })

task('validatorPool:acceptOwnership', 'Accepts a pending ownership transfer (candidateOwner)')
  .addParam('signer', 'The signer address (msg.sender)')
  .setAction(async (taskArguments, hre, runSuper) => {
    const signer = await hre.ethers.getSigner(taskArguments.signer)
    const validatorPool = new hre.ethers.Contract(precompileAddress, abi, signer)
    const pending = await validatorPool.acceptOwnership({ gasLimit: 50000 })
    const confirmed = await pending.wait()
    console.log(confirmed.hash)
  })

task('validatorPool:submitApplication', 'Submit a new validator application')
  .addParam('signer', 'The signer address (msg.sender)')
  .addParam('conspubkey', "The validator's consensus pub key")
  .addParam('moniker', "The validator's name")
  .addOptionalParam('identity', 'Optional identity signature (ex. UPort or Keybase)', '')
  .addOptionalParam('website', 'Optional website link', '')
  .addOptionalParam('security', 'Optional security contact information', '')
  .addOptionalParam('details', 'Optional details about the validator', '')
  .setAction(async (taskArguments, hre, runSuper) => {
    const signer = await hre.ethers.getSigner(taskArguments.signer)
    const validatorPool = new hre.ethers.Contract(precompileAddress, abi, signer)
    const description: string[] = [
      taskArguments.moniker,
      taskArguments.identity,
      taskArguments.website,
      taskArguments.security,
      taskArguments.details
    ]
    const pending = await validatorPool.submitApplication(taskArguments.conspubkey, description, { gasLimit: 50000 })
    const confirmed = await pending.wait()
    console.log(confirmed.hash)
  })

task('validatorPool:approveApplication', 'Approves a pending validator application (owner)')
  .addParam('signer', 'The signer address (msg.sender)')
  .addParam('operator', "The validator's operator address of the application to approve")
  .setAction(async (taskArguments, hre, runSuper) => {
    const signer = await hre.ethers.getSigner(taskArguments.signer)
    const validatorPool = new hre.ethers.Contract(precompileAddress, abi, signer)
    const pending = await validatorPool.approveApplication(taskArguments.operator, { gasLimit: 50000 })
    const confirmed = await pending.wait()
    console.log(confirmed.hash)
  })
