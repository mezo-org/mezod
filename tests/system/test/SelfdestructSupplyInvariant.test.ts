import { expect } from "chai"
import hre, { ethers } from "hardhat"
import btcabi from "../../../precompile/btctoken/abi.json"
import maintenanceAbi from "../../../precompile/maintenance/abi.json"

const btcTokenPrecompileAddress =
  "0x7b7c000000000000000000000000000000000000"

// SELFDESTRUCT is disabled by default; the PoA owner (accounts[0]) toggles it via
// the maintenance precompile so this regression can exercise the opcode.
const maintenanceAddress = "0x7b7c000000000000000000000000000000000013"

async function setSelfDestructDisabled(disabled: boolean) {
  const accounts = (hre.network.config as any).accounts
  const owner = new ethers.Wallet(accounts[0], ethers.provider)
  const maintenance = new ethers.Contract(
    maintenanceAddress,
    maintenanceAbi,
    owner,
  )
  await (await maintenance.setSelfDestructDisabled(disabled)).wait()
}

describe("SelfdestructSupplyInvariant", function () {
  let btcToken: any
  let senderSigner: any
  let initialSupply: bigint
  let updatedSupply: bigint
  let initialBlock: number
  let targetBlock: number
  let updatedBlock: number

  const fixture = async function () {
    btcToken = new hre.ethers.Contract(
      btcTokenPrecompileAddress,
      btcabi,
      ethers.provider,
    )

    const accounts = (hre.network.config as any).accounts
    senderSigner = new ethers.Wallet(accounts[0], ethers.provider)
  }

  const sendRawTransaction = async function (rawTx: string): Promise<void> {
    const controller = new AbortController()
    const timeoutId = setTimeout(() => controller.abort(), 10_000)

    try {
      await fetch((hre.network.config as any).url, {
        method: "POST",
        headers: { "content-type": "application/json" },
        body: JSON.stringify({
          jsonrpc: "2.0",
          method: "eth_sendRawTransaction",
          params: [rawTx],
          id: 1,
        }),
        signal: controller.signal,
      })
    } catch {
      // Rejected txs do not always return a regular response.
    } finally {
      clearTimeout(timeoutId)
    }
  }

  const waitForBlock = async function (
    targetHeight: number,
    timeoutMs: number,
  ): Promise<void> {
    const startedAt = Date.now()

    while (
      (await ethers.provider.getBlockNumber()) < targetHeight &&
      Date.now() - startedAt < timeoutMs
    ) {
      await new Promise((resolve) => setTimeout(resolve, 500))
    }
  }

  before(async function () {
    await setSelfDestructDisabled(false)
  })

  after(async function () {
    await setSelfDestructDisabled(true)
  })

  describe("selfdestruct to self during contract creation", function () {
    before(async function () {
      await fixture()

      initialSupply = await btcToken.totalSupply()
      initialBlock = await ethers.provider.getBlockNumber()

      // This creates and self-destructs a funded contract in one tx; without
      // the BTC supply invariant guard, that can halt consensus.
      const tx = await senderSigner.populateTransaction({
        to: null,
        value: 1n,
        data: "0x30ff",
        gasLimit: 100_000,
      })
      const rawTx = await senderSigner.signTransaction(tx)

      await sendRawTransaction(rawTx)

      targetBlock = initialBlock + 2
      await waitForBlock(targetBlock, 10_000)
      updatedBlock = await ethers.provider.getBlockNumber()
      updatedSupply = await btcToken.totalSupply()
    })

    it("should keep BTC totalSupply unchanged", async function () {
      expect(updatedSupply).to.equal(initialSupply)
    })

    it("should keep producing blocks", async function () {
      expect(updatedBlock).to.be.greaterThanOrEqual(targetBlock)
    })
  })
})
