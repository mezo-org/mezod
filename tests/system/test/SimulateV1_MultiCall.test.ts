import { expect } from "chai"
import hre, { ethers } from "hardhat"
import { SimpleToken } from "../typechain-types/SimpleToken"
import { getDeployedContract } from "./helpers/contract"

// SimulateV1_MultiCall exercises Phase 6: multiple calls in a single
// simulated block against one shared StateDB. The contract state written
// by one call must be visible to the next, block-gas-limit enforcement
// must be cumulative, and a revert in one call must not leak state to
// the next.
describe("SimulateV1_MultiCall", function () {
  const { deployments } = hre

  const MINT_AMOUNT = ethers.parseUnits("1000", 18)
  const TRANSFER_AMOUNT = ethers.parseUnits("500", 18)

  let simpleToken: SimpleToken
  let tokenAddr: string
  let senderAddr: string
  let recipientAddr: string

  before(async function () {
    await deployments.fixture(["BridgeOutDelegate"])
    simpleToken = await getDeployedContract<SimpleToken>("SimpleToken")
    tokenAddr = await simpleToken.getAddress()

    const [deployer] = await ethers.getSigners()
    senderAddr = deployer.address
    recipientAddr = ethers.Wallet.createRandom().address
  })

  context("state chains across calls within a block", function () {
    it("mint → transfer → balanceOf surfaces the transferred amount", async function () {
      const mintData = simpleToken.interface.encodeFunctionData("mint", [
        senderAddr,
        MINT_AMOUNT,
      ])
      const transferData = simpleToken.interface.encodeFunctionData(
        "transfer",
        [recipientAddr, TRANSFER_AMOUNT],
      )
      const balanceOfData = simpleToken.interface.encodeFunctionData(
        "balanceOf",
        [recipientAddr],
      )

      const opts = {
        blockStateCalls: [
          {
            stateOverrides: {
              [senderAddr]: { balance: ethers.toQuantity(ethers.parseEther("100")) },
            },
            calls: [
              { from: senderAddr, to: tokenAddr, data: mintData },
              { from: senderAddr, to: tokenAddr, data: transferData },
              { from: senderAddr, to: tokenAddr, data: balanceOfData },
            ],
          },
        ],
      }

      const blocks: any[] = await ethers.provider.send(
        "eth_simulateV1",
        [opts, "latest"],
      )
      expect(blocks).to.have.lengthOf(1)
      const calls = blocks[0].calls
      expect(calls).to.have.lengthOf(3)

      // All three calls succeed.
      expect(calls[0].status).to.equal("0x1", "mint must succeed")
      expect(calls[1].status).to.equal("0x1", "transfer must succeed")
      expect(calls[2].status).to.equal("0x1", "balanceOf must succeed")

      // Call 3's return data decodes to the transferred amount — proves
      // both prior calls' state mutations reached the balanceOf read.
      const [balance] = ethers.AbiCoder.defaultAbiCoder().decode(
        ["uint256"],
        calls[2].returnData,
      )
      expect(balance).to.equal(TRANSFER_AMOUNT)
    })
  })

  context("cumulative block gas limit enforced across calls", function () {
    it("over-budget call is rejected with -38015 while preceding calls stand", async function () {
      const recipient = ethers.Wallet.createRandom().address
      const opts = {
        blockStateCalls: [
          {
            blockOverrides: { gasLimit: ethers.toQuantity(50_000n) },
            stateOverrides: {
              [senderAddr]: { balance: ethers.toQuantity(ethers.parseEther("100")) },
            },
            calls: [
              {
                from: senderAddr,
                to: recipient,
                value: ethers.toQuantity(1n),
              },
              {
                from: senderAddr,
                to: recipient,
                value: ethers.toQuantity(1n),
                gas: ethers.toQuantity(30_000n),
              },
            ],
          },
        ],
      }

      const blocks: any[] = await ethers.provider.send(
        "eth_simulateV1",
        [opts, "latest"],
      )
      const calls = blocks[0].calls
      expect(calls).to.have.lengthOf(2)

      expect(calls[0].status).to.equal("0x1")

      expect(calls[1].error).to.exist
      expect(calls[1].error.code).to.equal(-38015)
    })
  })
})
