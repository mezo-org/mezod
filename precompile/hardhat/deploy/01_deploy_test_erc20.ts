import { DeployFunction } from "hardhat-deploy/dist/types"
import { HardhatRuntimeEnvironment } from "hardhat/types"

const func: DeployFunction = async (hre: HardhatRuntimeEnvironment) => {
    const { deployments, getNamedAccounts } = hre
    const { log, read } = deployments
    const { deployer } = await getNamedAccounts()

    console.log(`deployer is ${deployer}`)

    log("deploying TestERC20 contract...")

    await deployments.deploy("TestERC20", {
      from: deployer,
      log: true,
      waitConfirmations: 1,
      args: ["0x17F29B073143D8cd97b5bBe492bDEffEC1C5feE5"]
    })

    log(`minter is ${await read("TestERC20", "minter")}`)
}

export default func
func.tags = ["TestERC20"]