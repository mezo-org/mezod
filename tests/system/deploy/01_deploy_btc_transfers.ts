import type { HardhatRuntimeEnvironment } from "hardhat/types"
import type { DeployFunction } from "hardhat-deploy/types"

const func: DeployFunction = async (hre: HardhatRuntimeEnvironment) => {
  const { getNamedAccounts, deployments } = hre
  const { deployer } = await getNamedAccounts()

  console.log(`deployer is ${deployer}`)

  console.log("Deploying BTCTransfers contract...")

  await deployments.deploy("BTCTransfers", {
    from: deployer,
    args: [],
    log: true,
    waitConfirmations: 1,
  })

  console.log("Deploying OtherSpender contract...")
  await deployments.deploy("OtherSpender", {
    from: deployer,
    args: [],
    log: true,
    waitConfirmations: 1,
  })
}

export default func

func.tags = ["BTCTransfers"]
