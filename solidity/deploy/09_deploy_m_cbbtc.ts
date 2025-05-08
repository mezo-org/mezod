import type { DeployFunction } from "hardhat-deploy/types"
import { mERC20DeployFunctionFactory } from "../helpers/deploy-helpers"

const tokenContract = "mcbBTC"
const tokenName = "Mezo Coinbase Wrapped BTC"
const tokenSymbol = "mcbBTC"
const decimals = 8

const func: DeployFunction = mERC20DeployFunctionFactory(tokenContract, tokenName, tokenSymbol, decimals)

export default func

func.tags = [tokenContract]
