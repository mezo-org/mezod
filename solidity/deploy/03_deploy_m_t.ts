import type { DeployFunction } from "hardhat-deploy/types"
import { mERC20DeployFunctionFactory } from "../helpers/deploy-helpers"

const tokenContract = "mT"
const tokenName = "Mezo Threshold Network Token"
const tokenSymbol = "mT"
const decimals = 18

const func: DeployFunction = mERC20DeployFunctionFactory(tokenContract, tokenName, tokenSymbol, decimals)

export default func

func.tags = [tokenContract]
