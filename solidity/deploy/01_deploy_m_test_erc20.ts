import type { DeployFunction } from "hardhat-deploy/types"
import type { HardhatRuntimeEnvironment } from "hardhat/types"
import waitForTransaction from "../helpers/deploy-helpers"

const tokenName = "mTestERC20"

// TODO: Make this script reusable for all ERC20s.
const func: DeployFunction = async (hre: HardhatRuntimeEnvironment) => {
  const { ethers, getNamedAccounts, deployments, helpers, network } = hre
  const { deployer, minter } = await getNamedAccounts()
  const { log } = deployments

  const tokenSymbol = "TEST42"

  console.log(`deployer is ${deployer}`)
  console.log(`Deploying ${tokenName} contract...`)
  console.log(`Network name: ${network.name}`)

  const mTestERC20 = await deployments.getOrNull(tokenName)
  const isValidDeployment =
    mTestERC20 && helpers.address.isValid(mTestERC20.address)

  if (isValidDeployment) {
    log(`Using ${tokenName} at ${mTestERC20.address}`)
  } else {
    const [_, mTestERC20Deployment] = await helpers.upgrades.deployProxy(
      tokenName,
      {
        contractName: tokenName,
        initializerArgs: [
          tokenName,
          tokenSymbol,
          18,
          minter,
        ],
        factoryOpts: { signer: await ethers.getSigner(deployer) },
        proxyOpts: {
          kind: "transparent",
          // TODO: Set governance as the initial proxy admin owner.
        },
      },
    )

    if (
      mTestERC20Deployment.transactionHash
    ) {
      const confirmationsByChain: Record<string, number> = {
        mainnet: 3,
        testnet: 3,
      }
  
      await waitForTransaction(
        hre,
        mTestERC20Deployment.transactionHash,
        confirmationsByChain[network.name],
      )

      // TODO: fix verification for proxy admin.
      if (hre.network.tags.verify) {
        await hre.run("verify", {
          address: mTestERC20Deployment.address,
          constructorArgsParams: mTestERC20Deployment.args
        })
      }
    }
  }
}

export default func

func.tags = [tokenName]
func.skip = async (hre: HardhatRuntimeEnvironment): Promise<boolean> => 
  hre.network.name !== "hardhat"
