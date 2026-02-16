import type { HardhatRuntimeEnvironment } from "hardhat/types"
import type { DeployFunction } from "hardhat-deploy/types"

const tokens = [
  "mSolvBTC",
  "mT",
  "mxSolvBTC",
  "mUSDC",
  "mUSDT",
  "mUSDe",
  "mFBTC",
  "mcbBTC",
  "mDAI",
  "mswBTC",
]

const func: DeployFunction = async (hre: HardhatRuntimeEnvironment) => {
  const { helpers, deployments } = hre

  const { deployer } = await helpers.signers.getNamedSigners()

  for (const token of tokens) {
    const existingDeployment = await deployments.getOrNull(token)
    const isValidDeployment = existingDeployment &&
      helpers.address.isValid(existingDeployment.address)

    if (!isValidDeployment) {
      deployments.log(`Skipping ${token} - no valid deployment found`)
      continue
    }

    deployments.log(
      `Preparing upgrade for ${token} (proxy: ${existingDeployment.address})`,
    )

    // NOTE: The `initilizeV2` function should be updated/omitted for future
    //       upgrades. It is only needed for the upgrade that introduced EIP-712
    //       support to mERC20.
    const callData = (
      await hre.ethers.getContractFactory(
        `contracts/${token}.sol:${token}`,
        deployer,
      )
    ).interface.encodeFunctionData("initializeV2")

    const { newImplementationAddress, preparedTransaction } =
      await helpers.upgrades.prepareProxyUpgrade(token, token, {
        contractName: `contracts/${token}.sol:${token}`,
        callData,
        factoryOpts: { signer: deployer },
      })

    if (hre.network.name !== "mainnet") {
      deployments.log("Sending transaction to upgrade implementation...")
      await deployer.sendTransaction(preparedTransaction)
    }

    if (hre.network.name !== "hardhat") {
      // We use `verify` instead of `verify:verify` as the `verify` task is defined
      // in "@openzeppelin/hardhat-upgrades" to verify the proxy’s implementation
      // contract, the proxy itself and any proxy-related contracts, as well as
      // link the proxy to the implementation contract’s ABI on (Ether)scan.
      await hre.run("verify", {
        address: newImplementationAddress,
      })
    }
  }
}

export default func

func.tags = ["UpgradeMTokens"]

// Comment this line when running an upgrade.
func.skip = async () => true
