import { expect } from "chai"
import hre, { ethers } from "hardhat"
import { SimpleToken } from "../typechain-types/SimpleToken"
import { getDeployedContract } from "./helpers/contract"

// SimulateV1_MultiBlock exercises Phase 7: multiple simulated blocks
// share a single StateDB so state mutations from one block are visible
// in the next, and BLOCKHASH resolves consistently across both the
// canonical base block and the simulated-sibling range.
describe("SimulateV1_MultiBlock", function () {
  const { deployments } = hre

  // Raw bytecode that reads calldata[0..32] as a uint256 height and
  // returns BLOCKHASH(height) padded to 32 bytes. Installed via
  // stateOverrides in block 1 so block 3 can call it.
  //
  //   PUSH1 0 CALLDATALOAD BLOCKHASH PUSH1 0 MSTORE PUSH1 0x20 PUSH1 0 RETURN
  const BLOCKHASH_READER_BYTECODE = "0x6000354060005260206000F3"

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

  context("state chains across blocks", function () {
    it("block 1 mint is visible to block 2 balanceOf", async function () {
      const mintAmount = ethers.parseUnits("777", 18)
      const mintData = simpleToken.interface.encodeFunctionData("mint", [
        recipientAddr,
        mintAmount,
      ])
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
            calls: [{ from: senderAddr, to: tokenAddr, data: mintData }],
          },
          {
            calls: [{ from: senderAddr, to: tokenAddr, data: balanceOfData }],
          },
        ],
      }

      const blocks: any[] = await ethers.provider.send(
        "eth_simulateV1",
        [opts, "latest"],
      )
      expect(blocks).to.have.lengthOf(2)

      const [block1, block2] = blocks
      expect(block1.calls[0].status).to.equal("0x1", "block 1 mint must succeed")

      const block2Call = block2.calls[0]
      expect(block2Call.status).to.equal("0x1", "block 2 balanceOf must succeed")
      const [balance] = ethers.AbiCoder.defaultAbiCoder().decode(
        ["uint256"],
        block2Call.returnData,
      )
      expect(balance).to.equal(
        mintAmount,
        "block 2 must observe block 1's mint through the shared StateDB",
      )
    })
  })

  context("BLOCKHASH chains canonical + simulated-sibling ranges", function () {
    it("block 3 resolves BLOCKHASH(base+1) and BLOCKHASH(base+2) to the simulated headers", async function () {
      const latest = await ethers.provider.getBlock("latest")
      if (!latest) throw new Error("no latest block")
      const baseHeight = BigInt(latest.number)

      const readerAddr = ethers.getAddress(
        "0x" + "c0de".repeat(10),
      )
      const heightToCalldata = (h: bigint) =>
        ethers.zeroPadValue(ethers.toBeHex(h), 32)

      const opts = {
        blockStateCalls: [
          {
            stateOverrides: {
              [senderAddr]: { balance: ethers.toQuantity(ethers.parseEther("100")) },
              [readerAddr]: { code: BLOCKHASH_READER_BYTECODE },
            },
            calls: [],
          },
          { calls: [] },
          {
            calls: [
              {
                from: senderAddr,
                to: readerAddr,
                data: heightToCalldata(baseHeight + 1n),
              },
              {
                from: senderAddr,
                to: readerAddr,
                data: heightToCalldata(baseHeight + 2n),
              },
            ],
          },
        ],
      }

      const blocks: any[] = await ethers.provider.send(
        "eth_simulateV1",
        [opts, "latest"],
      )
      expect(blocks).to.have.lengthOf(3)

      const block3Calls = blocks[2].calls
      expect(block3Calls).to.have.lengthOf(2)

      // Both reads succeed.
      expect(block3Calls[0].status).to.equal("0x1")
      expect(block3Calls[1].status).to.equal("0x1")

      // Each simulated-sibling BLOCKHASH must match the corresponding
      // block envelope's `hash` field. This is the load-bearing
      // invariant of Phase 7: a simulated block N can observe a past
      // simulated sibling M's hash via BLOCKHASH, and the response
      // envelope surfaces the same hash.
      expect(block3Calls[0].returnData.toLowerCase()).to.equal(
        blocks[0].hash.toLowerCase(),
        "BLOCKHASH(base+1) must match block[0].hash",
      )
      expect(block3Calls[1].returnData.toLowerCase()).to.equal(
        blocks[1].hash.toLowerCase(),
        "BLOCKHASH(base+2) must match block[1].hash",
      )
    })
  })
})
