import type { DeployFunction } from "hardhat-deploy/types"
import type { HardhatRuntimeEnvironment } from "hardhat/types"
import waitForTransaction from "../helpers/deploy-helpers"

const tokenName = "TestERC20"

const func: DeployFunction = async (hre: HardhatRuntimeEnvironment) => {
  const { ethers, getNamedAccounts, deployments, helpers, network } = hre
  const { deployer } = await getNamedAccounts()
  const { log } = deployments

  const tokenSymbol = "TEST42"
  const initialMinter = "0x17F29B073143D8cd97b5bBe492bDEffEC1C5feE5" // x/bridge module account

  console.log(`deployer is ${deployer}`)
  console.log(`Deploying ${tokenName} contract...`)

  const TestERC20 = await deployments.getOrNull(tokenName)
  const isValidDeployment =
    TestERC20 && helpers.address.isValid(TestERC20.address)

  if (isValidDeployment) {
    log(`Using ${tokenName} at ${TestERC20.address}`)
  } else {
    const [_, testERC20Deployment] = await helpers.upgrades.deployProxy(
      tokenName,
      {
        contractName: tokenName,
        initializerArgs: [
          tokenName,
          tokenSymbol,
          initialMinter,
        ],
        factoryOpts: { signer: await ethers.getSigner(deployer) },
        proxyOpts: {
          kind: "transparent",
        },
      },
    )

    if (
      testERC20Deployment.transactionHash
    ) {
      const confirmationsByChain: Record<string, number> = {
        mainnet: 3,
        testnet: 3,
      }
  
      await waitForTransaction(
        hre,
        testERC20Deployment.transactionHash,
        confirmationsByChain[network.name],
      )

      // TODO: fix verification for proxy admin.
      if (hre.network.tags.verify) {
        await hre.run("verify", {
          address: testERC20Deployment.address,
          constructorArgsParams: testERC20Deployment.args,
        })
      }
    }
  }
}

export default func

func.tags = [tokenName]
