import type { HardhatRuntimeEnvironment } from "hardhat/types"
import type { DeployFunction } from "hardhat-deploy/types"

const func: DeployFunction = async (hre: HardhatRuntimeEnvironment) => {
  const { getNamedAccounts, deployments } = hre
  const { deployer } = await getNamedAccounts()

  console.log(`deployer is ${deployer}`)

  console.log("Deploying MyERC20 contract...")

  // TODO:
  // THIS IS JUST A TEST ERC20 CONTRACT. REPLACE IT WITH REAL MAPPED ERC20(s)
  // Deploy as upgradable proxy - deployProxy()
  await deployments.deploy("TestERC20", {
    from: deployer,
    args: [],
    log: true,
    waitConfirmations: 1,
  })
}

export default func

func.tags = ["TestERC20"]
