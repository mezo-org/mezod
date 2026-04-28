import { expect } from "chai"
import hre, { ethers } from "hardhat"
import { SimpleToken } from "../typechain-types/SimpleToken"
import { getDeployedContract } from "./helpers/contract"
import btcabi from "../../../precompile/btctoken/abi.json"

// SimulateV1_MultiCall covers multiple calls in a single simulated
// block against one shared StateDB. The contract state written by one
// call must be visible to the next, block-gas-limit enforcement must be
// cumulative, and a revert in one call must not leak state to the next.
describe("SimulateV1_MultiCall", function () {
  const { deployments } = hre

  const BTC_TOKEN_PRECOMPILE = "0x7b7c000000000000000000000000000000000000"

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

  // Custom Mezo precompiles (btctoken, mezotoken, …) write through the
  // Cosmos bank keeper, not the EVM journal — their mutations ride in
  // the StateDB's cached Cosmos context and must survive call
  // boundaries. balance stateOverrides only touch the EVM state object
  // and do not propagate to bankKeeper, so this case leans on the
  // localnode dev0 signer's real pre-funded BTC balance.
  context("custom Mezo precompile state chains across calls", function () {
    it("btctoken.transfer → btctoken.balanceOf surfaces the transferred amount", async function () {
      const btcToken: any = new hre.ethers.Contract(
        BTC_TOKEN_PRECOMPILE,
        btcabi,
        ethers.provider,
      )
      const transferAmount = ethers.parseEther("0.5")

      const transferData = btcToken.interface.encodeFunctionData("transfer", [
        recipientAddr,
        transferAmount,
      ])
      const balanceOfData = btcToken.interface.encodeFunctionData("balanceOf", [
        recipientAddr,
      ])

      const opts = {
        blockStateCalls: [
          {
            calls: [
              { from: senderAddr, to: BTC_TOKEN_PRECOMPILE, data: transferData },
              { from: senderAddr, to: BTC_TOKEN_PRECOMPILE, data: balanceOfData },
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
      expect(calls).to.have.lengthOf(2)

      expect(calls[0].status).to.equal("0x1", "btctoken.transfer must succeed")
      expect(calls[1].status).to.equal("0x1", "btctoken.balanceOf must succeed")

      const [balance] = ethers.AbiCoder.defaultAbiCoder().decode(
        ["uint256"],
        calls[1].returnData,
      )
      expect(balance).to.equal(transferAmount)
    })
  })
})
