import type { DeployFunction } from "hardhat-deploy/types"
import type { HardhatRuntimeEnvironment } from "hardhat/types"

const func: DeployFunction = async (hre: HardhatRuntimeEnvironment) => {
  const { getNamedAccounts, deployments } = hre
  const { deployer } = await getNamedAccounts()

  console.log(`deployer is ${deployer}`)
  console.log("Deploying Selfdestruct6780Check contract...")

  await deployments.deploy("Selfdestruct6780Check", {
    from: deployer,
    args: [],
    log: true,
    waitConfirmations: 1,
  })
}

export default func

func.tags = ["Selfdestruct6780Check"]
