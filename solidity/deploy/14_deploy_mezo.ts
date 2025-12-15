import type { HardhatRuntimeEnvironment } from "hardhat/types"
import type { DeployFunction } from "hardhat-deploy/types"
import { saveDeploymentArtifact, waitForTransaction } from "../helpers/deploy-helpers"

const func: DeployFunction = async function (hre: HardhatRuntimeEnvironment) {
  const { deployments, getNamedAccounts, upgrades, helpers } = hre
  const { execute, read, log } = deployments
  const { deployer } = await getNamedAccounts()

  const existingDeployment = await deployments.getOrNull("MEZO")
  const isValidDeployment = existingDeployment &&
    helpers.address.isValid(existingDeployment.address)

  if (isValidDeployment) {
    log(`Using MEZO at ${existingDeployment.address}`)
    return
  }

  const deployTx = await execute(
    "MEZODeployer",
    { from: deployer, log: true },
    "deployToken",
  )

  await waitForTransaction(hre, deployTx.transactionHash, 12)

  const MEZOProxyAddress = await read("MEZODeployer", "token")

  log(`MEZO transparent proxy deployed at ${MEZOProxyAddress}`)

  // Get the implementation address from the proxy using OpenZeppelin's upgrades plugin
  const MEZOImplAddress = await upgrades.erc1967.getImplementationAddress(MEZOProxyAddress)

  log(`MEZO implementation at ${MEZOImplAddress}`)

  const deployment = await saveDeploymentArtifact(
    hre,
    "MEZO",
    MEZOProxyAddress,
    deployTx.transactionHash,
    {
      implementation: MEZOImplAddress,
      log: true,
    },
  )

  if (hre.network.tags.verify) {
    await helpers.etherscan.verify(deployment, "contracts/MEZO.sol:MEZO")
  }
}

export default func

func.tags = ["MEZO"]
func.dependencies = ["MEZODeployer"]
