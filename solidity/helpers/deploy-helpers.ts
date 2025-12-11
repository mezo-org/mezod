import type { DeployFunction, Deployment } from "hardhat-deploy/types"
import { HardhatRuntimeEnvironment } from "hardhat/types"

export function mERC20DeployFunctionFactory(
  tokenContract: string, 
  tokenName: string, 
  tokenSymbol: string, 
  decimals: number
): DeployFunction {
  return async (hre: HardhatRuntimeEnvironment) => {
    const { ethers, getNamedAccounts, deployments, helpers, network } = hre
    const { deployer, governance, minter } = await getNamedAccounts()
    const { log } = deployments

    console.log(`Deploying ${tokenContract} contract...`)

    console.log(`Network name: ${network.name}`)
    console.log(`Deployer is ${deployer}`)
    console.log(`Governance is ${governance}`)
    console.log(`Minter is ${minter}`)

    const existingDeployment = await deployments.getOrNull(tokenContract)
    const isValidDeployment = existingDeployment &&
      helpers.address.isValid(existingDeployment.address)

    if (isValidDeployment) {
      log(`Using ${tokenContract} at ${existingDeployment.address}`)
    } else {
      const [_, deployment] = await helpers.upgrades.deployProxy(
        tokenContract,
        {
          contractName: tokenContract,
          initializerArgs: [
            tokenName,
            tokenSymbol,
            decimals,
            minter,
          ],
          factoryOpts: { signer: await ethers.getSigner(deployer) },
          proxyOpts: {
            kind: "transparent",
            initialOwner: governance,
            // The below option is necessary to deploy distinct contract implementations 
            // in case multiple token contracts extending mERC20.sol have the same bytecode.
            redeployImplementation: "always"
          },
        },
      )

      await helpers.ownable.transferOwnership(tokenContract, governance, deployer)

      if (
        deployment.transactionHash
      ) {
        const confirmationsByChain: Record<string, number> = {
          mainnet: 3,
          testnet: 3,
        }

        await waitForTransaction(
          hre,
          deployment.transactionHash,
          confirmationsByChain[network.name],
        )

        // TODO: fix verification for proxy admin.
        if (hre.network.tags.verify) {
          const fqn = tokenContract.toLowerCase().includes("test") ? 
              `contracts/test/${tokenContract}.sol:${tokenContract}` : 
              `contracts/${tokenContract}.sol:${tokenContract}`

          await helpers.etherscan.verify(deployment, fqn)  
        }
      }
    }
  }
}

export async function waitForTransaction(
  hre: HardhatRuntimeEnvironment,
  txHash: string,
  confirmations: number = 1,
) {
  if (hre.network.name === "hardhat") {
    return
  }

  const { provider } = hre.ethers
  const transaction = await provider.getTransaction(txHash)

  if (!transaction) {
    throw new Error(`Transaction ${txHash} not found`)
  }

  let currentConfirmations = await transaction.confirmations()
  while (currentConfirmations < confirmations) {
    // wait 1s between each check to save API compute units
    // eslint-disable-next-line no-await-in-loop, no-promise-executor-return
    await new Promise((resolve) => setTimeout(resolve, 1000))
    // eslint-disable-next-line no-await-in-loop
    currentConfirmations = await transaction.confirmations()
  }
}

export async function saveDeploymentArtifact(
  hre: HardhatRuntimeEnvironment,
  deploymentName: string,
  contractAddress: string,
  transactionHash: string,
  opts?: {
    contractName?: string
    constructorArgs?: unknown[]
    implementation?: string
    log?: boolean
  },
): Promise<Deployment> {
  const artifact = await hre.artifacts.readArtifact(
    opts?.contractName || deploymentName,
  )

  const deployment: Deployment = {
    address: contractAddress,
    abi: artifact.abi,
    transactionHash,
    args: opts?.constructorArgs,
    implementation: opts?.implementation,
  }

  await hre.deployments.save(deploymentName, deployment)

  if (opts?.log) {
    hre.deployments.log(
      `Saved deployment artifact for '${deploymentName}' with address ${contractAddress}` +
        ` and deployment transaction: ${transactionHash}`,
    )
  }

  return deployment
}
