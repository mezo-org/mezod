import type { DeployFunction } from "hardhat-deploy/types"
import type { HardhatRuntimeEnvironment } from "hardhat/types"
import { mERC20DeployFunctionFactory } from "../helpers/deploy-helpers"

const tokenContract = "mTestERC20"
const tokenName = "Test Token"
const tokenSymbol = "TEST"
const decimals = 8

const func: DeployFunction = mERC20DeployFunctionFactory(tokenContract, tokenName, tokenSymbol, decimals)

export default func

func.tags = [tokenContract]
func.skip = async (hre: HardhatRuntimeEnvironment): Promise<boolean> => 
  hre.network.name !== "hardhat"
