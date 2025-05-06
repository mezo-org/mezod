import { HardhatRuntimeEnvironment } from "hardhat/types"

export default async function waitForTransaction(
  hre: HardhatRuntimeEnvironment,
  txHash: string,
  confirmations: number = 1,
) {
  if (hre.network.name === "hardhat") {
    return
  }

  const { provider } = hre.ethers
  const transaction = await provider.getTransaction(txHash)

  if (!transaction) {
    throw new Error(`Transaction ${txHash} not found`)
  }

  let currentConfirmations = await transaction.confirmations()
  while (currentConfirmations < confirmations) {
    // wait 1s between each check to save API compute units
    // eslint-disable-next-line no-await-in-loop, no-promise-executor-return
    await new Promise((resolve) => setTimeout(resolve, 1000))
    // eslint-disable-next-line no-await-in-loop
    currentConfirmations = await transaction.confirmations()
  }
}
