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

  const MEZOAddress = await read("MEZODeployer", "token")

  log(`MEZO deployed at ${MEZOAddress}`)

  const deployment = await saveDeploymentArtifact(
    hre,
    "MEZO",
    MEZOAddress,
    deployTx.transactionHash,
    {
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
