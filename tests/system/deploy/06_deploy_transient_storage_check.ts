import type { DeployFunction } from "hardhat-deploy/types"
import type { HardhatRuntimeEnvironment } from "hardhat/types"

const func: DeployFunction = async (hre: HardhatRuntimeEnvironment) => {
  const { getNamedAccounts, deployments } = hre
  const { deployer } = await getNamedAccounts()

  console.log(`deployer is ${deployer}`)
  console.log("Deploying TransientStorageCheck contract...")

  await deployments.deploy("TransientStorageCheck", {
    from: deployer,
    args: [],
    log: true,
    waitConfirmations: 1,
  })

  console.log("Deploying TransientStorageReader contract...")

  await deployments.deploy("TransientStorageReader", {
    from: deployer,
    args: [],
    log: true,
    waitConfirmations: 1,
  })
}

export default func

func.tags = ["TransientStorageCheck"]
