import { deployments, ethers } from "hardhat"

import type { BaseContract } from "ethers"


/**
 * Get instance of a contract from Hardhat Deployments.
 * @param deploymentName Name of the contract deployment.
 * @returns Deployed Ethers contract instance.
 */
export async function getDeployedContract<T extends BaseContract>(
  deploymentName: string,
): Promise<T> {
  const { address, abi } = await deployments.get(deploymentName)
  const [defaultSigner] = await ethers.getSigners()

  return new ethers.BaseContract(address, abi, defaultSigner) as T
}
