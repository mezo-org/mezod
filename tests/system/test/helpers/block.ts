import { ethers } from "hardhat"

export async function waitForBlock(targetHeight: number): Promise<void> {
  while ((await ethers.provider.getBlockNumber()) < targetHeight) {
    await new Promise((resolve) => setTimeout(resolve, 500))
  }
}
