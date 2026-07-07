import type { HardhatRuntimeEnvironment } from "hardhat/types"
import type { DeployFunction } from "hardhat-deploy/types"

const func: DeployFunction = async (hre: HardhatRuntimeEnvironment) => {
  const { getNamedAccounts, deployments } = hre
  const { deployer } = await getNamedAccounts()

  console.log(`deployer is ${deployer}`)
  console.log("Deploying P256VerifyCheck contract...")

  await deployments.deploy("P256VerifyCheck", {
    from: deployer,
    args: [],
    log: true,
    waitConfirmations: 1,
  })
}

export default func

func.tags = ["P256VerifyCheck"]
