import type { DeployFunction } from "hardhat-deploy/types"
import { mERC20DeployFunctionFactory } from "../helpers/deploy-helpers"

const tokenContract = "mswBTC"
const tokenName = "Mezo swBTC"
const tokenSymbol = "mswBTC"
const decimals = 8

const func: DeployFunction = mERC20DeployFunctionFactory(tokenContract, tokenName, tokenSymbol, decimals)

export default func

func.tags = [tokenContract]
