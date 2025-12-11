import type { HardhatRuntimeEnvironment } from "hardhat/types"
import type { DeployFunction } from "hardhat-deploy/types"
import { saveDeploymentArtifact, waitForTransaction } from "../helpers/deploy-helpers"

const func: DeployFunction = async function (hre: HardhatRuntimeEnvironment) {
  const { deployments, getNamedAccounts, upgrades, helpers } = hre
  const { execute, read, log } = deployments
  const { deployer } = await getNamedAccounts()

  const existingDeployment = await deployments.getOrNull("xMEZO")
  const isValidDeployment = existingDeployment &&
    helpers.address.isValid(existingDeployment.address)

  if (isValidDeployment) {
    log(`Using xMEZO at ${existingDeployment.address}`)
    return
  }

  const deployTx = await execute(
    "xMEZODeployer",
    { from: deployer, log: true },
    "deployToken",
  )

  await waitForTransaction(hre, deployTx.transactionHash, 12)

  const xMEZOProxyAddress = await read("xMEZODeployer", "token")

  log(`xMEZO transparent proxy deployed at ${xMEZOProxyAddress}`)

  // Get the implementation address from the proxy using OpenZeppelin's upgrades plugin
  const xMEZOImplAddress = await upgrades.erc1967.getImplementationAddress(xMEZOProxyAddress)

  log(`xMEZO implementation at ${xMEZOImplAddress}`)

  const deployment = await saveDeploymentArtifact(
    hre,
    "xMEZO",
    xMEZOProxyAddress,
    deployTx.transactionHash,
    {
      implementation: xMEZOImplAddress,
      log: true,
    },
  )

  if (hre.network.tags.verify) {
    await helpers.etherscan.verify(deployment)
  }
}

export default func

func.tags = ["xMEZO"]
func.dependencies = ["xMEZODeployer"]
