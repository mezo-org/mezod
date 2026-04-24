import { expect } from "chai"
import hre, { ethers } from "hardhat"
import { SimpleToken } from "../typechain-types/SimpleToken"
import { getDeployedContract } from "./helpers/contract"

describe("SimulateV1_SingleCall", function () {
  const { deployments } = hre

  const AMOUNT = ethers.parseUnits("1", 18)
  const HUGE_AMOUNT = AMOUNT * 1_000_000n
  const TRANSFER_TOPIC = ethers.id("Transfer(address,address,uint256)")
  // Solidity 0.8 panic selector: Panic(uint256). Underflow emits code 0x11.
  const PANIC_SELECTOR = "0x4e487b71"

  let simpleToken: SimpleToken
  let tokenAddr: string
  let senderAddr: string
  let recipientAddr: string

  let happyResult: any
  let revertResult: any

  before(async function () {
    await deployments.fixture(["BridgeOutDelegate"])
    simpleToken = await getDeployedContract<SimpleToken>("SimpleToken")
    tokenAddr = await simpleToken.getAddress()

    const [deployer] = await ethers.getSigners()
    senderAddr = deployer.address
    recipientAddr = ethers.Wallet.createRandom().address

    // Seed the sender's ERC-20 balance for the happy-path transfer. The
    // revert-path case uses a larger transfer amount that underflows this
    // balance inside the contract.
    const mintTx = await simpleToken.connect(deployer).mint(senderAddr, AMOUNT * 10n)
    await mintTx.wait()

    const happyCallData = simpleToken.interface.encodeFunctionData(
      "transfer",
      [recipientAddr, AMOUNT],
    )
    const revertCallData = simpleToken.interface.encodeFunctionData(
      "transfer",
      [recipientAddr, HUGE_AMOUNT],
    )

    // Balance override is redundant with the localnode's pre-funded dev
    // account, but we set it anyway so the stateOverrides path is
    // exercised and regressions in it fail the suite.
    const stateOverrides = {
      [senderAddr]: { balance: ethers.toQuantity(ethers.parseEther("100")) },
    }

    const happyOpts = {
      blockStateCalls: [
        {
          stateOverrides,
          calls: [{ from: senderAddr, to: tokenAddr, data: happyCallData }],
        },
      ],
    }
    const happyBlocks: any[] = await ethers.provider.send(
      "eth_simulateV1",
      [happyOpts, "latest"],
    )
    happyResult = happyBlocks[0].calls[0]

    const revertOpts = {
      blockStateCalls: [
        {
          stateOverrides,
          calls: [{ from: senderAddr, to: tokenAddr, data: revertCallData }],
        },
      ],
    }
    const revertBlocks: any[] = await ethers.provider.send(
      "eth_simulateV1",
      [revertOpts, "latest"],
    )
    revertResult = revertBlocks[0].calls[0]
  })

  context("happy path (ERC-20 transfer with balance override)", function () {
    it("returns status 0x1 with positive gasUsed and no error", function () {
      expect(happyResult.status).to.equal("0x1")
      expect(BigInt(happyResult.gasUsed)).to.be.greaterThan(0n)
      expect(happyResult.error).to.be.undefined
    })

    it("returnData decodes to true", function () {
      const [decoded] = ethers.AbiCoder.defaultAbiCoder().decode(
        ["bool"],
        happyResult.returnData,
      )
      expect(decoded).to.equal(true)
    })

    it("emits the ERC-20 Transfer log with sender, recipient, amount", function () {
      expect(happyResult.logs).to.have.lengthOf(1)
      const log = happyResult.logs[0]
      expect(log.address.toLowerCase()).to.equal(tokenAddr.toLowerCase())
      expect(log.topics[0]).to.equal(TRANSFER_TOPIC)
      expect(BigInt(log.topics[1])).to.equal(BigInt(senderAddr))
      expect(BigInt(log.topics[2])).to.equal(BigInt(recipientAddr))
      expect(BigInt(log.data)).to.equal(AMOUNT)
    })
  })

  context("revert path (transfer exceeds balance, Solidity 0.8 underflow)", function () {
    it("returns status 0x0", function () {
      expect(revertResult.status).to.equal("0x0")
    })

    it("per-call error carries spec-reserved code 3", function () {
      expect(revertResult.error).to.exist
      expect(revertResult.error.code).to.equal(3)
    })

    it("per-call error.message is 'execution reverted'", function () {
      expect(revertResult.error.message).to.equal("execution reverted")
    })

    it("per-call error.data is the hex-encoded Panic(0x11) payload", function () {
      expect(revertResult.error.data).to.be.a("string")
      expect(revertResult.error.data.toLowerCase().startsWith(PANIC_SELECTOR)).to.equal(true)
    })
  })
})
