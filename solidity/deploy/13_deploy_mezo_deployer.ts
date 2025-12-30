
import { DeployFunction } from "hardhat-deploy/dist/types"
import { HardhatRuntimeEnvironment } from "hardhat/types"
import { deployWithSingletonFactory } from "../helpers/erc2470"

const func: DeployFunction = async (hre: HardhatRuntimeEnvironment) => {
  const { ethers, helpers, deployments } = hre
  const { log } = deployments
  const { deployer } = await helpers.signers.getNamedSigners()

  const existingDeployment = await deployments.getOrNull("MEZODeployer")
  const isValidDeployment = existingDeployment &&
    helpers.address.isValid(existingDeployment.address)

  if (isValidDeployment) {
    log(`Using MEZODeployer at ${existingDeployment.address}`)
  } else {
    log("Deploying the MEZODeployer...")

    const deployTx = await deployWithSingletonFactory(
      hre,
      "MEZODeployer",
      {
        contractName: "contracts/MEZODeployer.sol:MEZODeployer",
        from: deployer,
        salt: ethers.keccak256(
          // Note that this is the salt for deploying the MEZODeployer contract.
          // The salt for MEZO token contract is defined inside the MEZODeployer
          // as a SALT constant. Both salts doesn't have to be the same but we keep
          // them as such for consistency.
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

func.tags = ["MEZODeployer"]
