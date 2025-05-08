import type { DeployFunction } from "hardhat-deploy/types"
import { mERC20DeployFunctionFactory } from "../helpers/deploy-helpers"

const tokenContract = "mSolvBTC"
const tokenName = "Mezo Solv BTC"
const tokenSymbol = "mSolvBTC"
const decimals = 18

const func: DeployFunction = mERC20DeployFunctionFactory(tokenContract, tokenName, tokenSymbol, decimals)

export default func

func.tags = [tokenContract]
