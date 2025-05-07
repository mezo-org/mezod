import type { DeployFunction } from "hardhat-deploy/types"
import type { HardhatRuntimeEnvironment } from "hardhat/types"
import { mERC20DeployFunctionFactory } from "../helpers/deploy-helpers"

const tokenName = "mTestERC20"
const tokenSymbol = "TEST42"
const decimals = 18

const func: DeployFunction = mERC20DeployFunctionFactory(tokenName, tokenSymbol, decimals)

export default func

func.tags = [tokenName]
func.skip = async (hre: HardhatRuntimeEnvironment): Promise<boolean> => 
  hre.network.name !== "hardhat"
