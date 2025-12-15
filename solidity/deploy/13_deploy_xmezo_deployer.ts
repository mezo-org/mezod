
import { DeployFunction } from "hardhat-deploy/dist/types"
import { HardhatRuntimeEnvironment } from "hardhat/types"
import { deployWithSingletonFactory } from "../helpers/erc2470"

const func: DeployFunction = async (hre: HardhatRuntimeEnvironment) => {
  const { ethers, helpers, deployments } = hre
  const { log } = deployments
  const { deployer } = await helpers.signers.getNamedSigners()

  const existingDeployment = await deployments.getOrNull("xMEZODeployer")
  const isValidDeployment = existingDeployment &&
    helpers.address.isValid(existingDeployment.address)

  if (isValidDeployment) {
    log(`Using xMEZODeployer at ${existingDeployment.address}`)
  } else {
    log("Deploying the xMEZODeployer...")

    const deployTx = await deployWithSingletonFactory(
      hre,
      "xMEZODeployer",
      {
        contractName: "contracts/xMEZODeployer.sol:xMEZODeployer",
        from: deployer,
        salt: ethers.keccak256(
          ethers.toUtf8Bytes(
            "Bank on yourself. Bring everyday finance to your Bitcoin.",
          ),
        ),
        confirmations: 12,
      },
    )

    if (hre.network.tags.verify) {
      await helpers.etherscan.verify(deployTx.deployment)
    }
  }
}

export default func

func.tags = ["xMEZODeployer"]
