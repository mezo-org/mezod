import type { DeployFunction } from "hardhat-deploy/types"
import { mERC20DeployFunctionFactory } from "../helpers/deploy-helpers"

const tokenContract = "mUSDC"
const tokenName = "Mezo Circle USDC"
const tokenSymbol = "mUSDC"
const decimals = 6

const func: DeployFunction = mERC20DeployFunctionFactory(tokenContract, tokenName, tokenSymbol, decimals)

export default func

func.tags = [tokenContract]
