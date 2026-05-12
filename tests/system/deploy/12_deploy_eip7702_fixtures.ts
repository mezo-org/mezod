import type { HardhatRuntimeEnvironment } from "hardhat/types"
import type { DeployFunction } from "hardhat-deploy/types"

const func: DeployFunction = async (hre: HardhatRuntimeEnvironment) => {
  const { getNamedAccounts, deployments } = hre
  const { deployer } = await getNamedAccounts()

  console.log(`deployer is ${deployer}`)
  console.log("Deploying EIP-7702 fixtures (TargetV1, TargetV2, ExtCodeReader, Caller)...")

  const opts = { from: deployer, args: [], log: true, waitConfirmations: 1 }
  for (const name of [
    "Eip7702TargetV1",
    "Eip7702TargetV2",
    "Eip7702ExtCodeReader",
    "Eip7702Caller",
  ]) {
    await deployments.deploy(name, opts)
  }
}

export default func

func.tags = ["eip7702-fixtures"]
