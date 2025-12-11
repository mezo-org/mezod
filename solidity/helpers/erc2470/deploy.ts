import type { Signer, BytesLike, Contract } from "ethers"
import type { Deployment } from "hardhat-deploy/dist/types"
import { ensureSingletonFactory } from "./singleton-factory"
import { saveDeploymentArtifact } from "../deploy-helpers"
import { HardhatRuntimeEnvironment } from "hardhat/types"

/**
 * Options for deploying a contract using a singleton factory.
 * @property {Signer} from - The signer object initiating the contract deployment.
 * @property {BytesLike} salt - A unique salt value used in the deployment process.
 * @property {string} [contractName] - Optional. Name of the contract to deploy.
 * @property {FactoryOptions} [factoryOpts] - Optional. Custom options for the contract factory.
 * @property {unknown[]} [constructorArgs] - Optional. Arguments for the contract's constructor.
 */
type Options = {
  from: Signer
  salt: BytesLike
  contractName?: string
  constructorArgs?: unknown[]
  confirmations?: number
}

/**
 * Deploys a contract using a singleton factory with optional constructor arguments.
 * This approach can be used to deploy contracts with deterministic addresses using the CREATE2 opcode.
 *
 * @async
 * @param {string} deploymentName - The name of the deployment.
 * @param {Options} opts - The deployment options.
 * @returns {Promise<{ contractInstance: T; deployment: Deployment }>} The deployment
 * details.
 * @throws Will throw an error if the deployment fails.
 */
export async function deployWithSingletonFactory(
  hre: HardhatRuntimeEnvironment,
  deploymentName: string,
  opts: Options,
): Promise<{ contractInstance: Contract; deployment: Deployment }> {
  const { ethers, deployments } = hre
  const contractName = opts?.contractName || deploymentName

  const singletonFactory = await ensureSingletonFactory(opts.from)

  const contractFactory = await ethers.getContractFactory(
    contractName,
    opts?.from,
  )

  const creationCode = contractFactory.bytecode
  const constructorArgs = opts?.constructorArgs
    ? contractFactory.interface.encodeDeploy(opts.constructorArgs)
    : ""

  // The initcode is the contract bytecode concatenated with the encoded constructor arguments.
  const initCode = creationCode + constructorArgs

  // Simulate contract deployment with SingletonFactory.
  const contractAddress = await singletonFactory.deploy.staticCall(
    initCode,
    opts.salt,
  )

  if (contractAddress === ethers.ZeroAddress) {
    throw new Error("Deployment simulation failed")
  }

  const contractInstance: Contract = contractFactory.attach(contractAddress) as Contract

  // Increase gas limit as the estimates are not accurate.
  // const gasEstimate = await singletonFactory.deploy.estimateGas(
  //  initCode,
  //  opts.salt,
  // )

  // FIXME: There is something off with gas estimation on matsnet and the gas
  // limit is severely underestimated. This TX was executed with the commented out
  // code and an estimate of ~360k gas:
  // https://explorer.test.mezo.org/tx/0x989a34b8b7924549f9d5eb11d39f5a785e7bc83250831f5330f33074b19d9bf6
  //
  // const gasLimit = (gasEstimate * 150n) / 100n

  // FIXME: This transaction was executed with a hardcoded limit of 3M gas - which
  // was way closer to the expectation - and it succeeded:
  // https://explorer.test.mezo.org/tx/0x88aec2c7a857a6f0b3b2536bc0029e4c71d8205c6bf5e473e62402f36c9d517f
  const gasLimit = 3_000_000

  // Deploy contract with SingletonFactory.
  const deploymentTransaction = await singletonFactory.deploy(
    initCode,
    opts.salt,
    { gasLimit },
  )

  deployments.log(
    `Waiting for ${deploymentName} deployment transaction ${deploymentTransaction.hash} confirmations...`,
  )

  const transactionReceipt = await deploymentTransaction.wait(
    opts.confirmations || 1,
  )

  // Check the contract was deployed.
  const existingRuntimeCode = await opts.from.provider!.getCode(contractAddress)
  if (
    typeof existingRuntimeCode === "undefined" ||
    existingRuntimeCode === null ||
    existingRuntimeCode === "0x" ||
    existingRuntimeCode === "0x0"
  ) {
    throw new Error(`Deployment at ${contractAddress} failed`)
  }

  const transactionHash = transactionReceipt?.hash

  if (!transactionReceipt || !transactionHash) {
    throw new Error(
      `Could not find transaction receipt for transaction hash: ${transactionHash}`,
    )
  }

  deployments.log(
    `Deployed ${deploymentName} contract (address: ${contractAddress}) ` +
      `in transaction: ${transactionHash}`,
  )

  const deployment: Deployment = await saveDeploymentArtifact(
    hre,
    deploymentName,
    contractAddress,
    deploymentTransaction.hash,
    {
      contractName,
      constructorArgs: opts?.constructorArgs,
    },
  )

  return { contractInstance, deployment }
}
