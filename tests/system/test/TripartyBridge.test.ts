import { expect } from "chai"
import hre from "hardhat"
import { ethers } from "hardhat"
import assetsbridgeabi from "../../../precompile/assetsbridge/abi.json"
import btcabi from "../../../precompile/btctoken/abi.json"
import validatorpoolabi from "../../../precompile/validatorpool/abi.json"
import { TripartyController } from "../typechain-types/TripartyController"
import { getDeployedContract } from "./helpers/contract"
import { waitForBlock } from "./helpers/block"

const validatorPoolPrecompileAddress =
  "0x7b7c000000000000000000000000000000000011"
const assetsBridgePrecompileAddress =
  "0x7b7c000000000000000000000000000000000012"
const btcTokenPrecompileAddress =
  "0x7b7c000000000000000000000000000000000000"

const BTC = (n: number | string) => ethers.parseEther(n.toString())

describe("TripartyBridge", function () {
  const { deployments } = hre
  let assetsBridge: any
  let btcToken: any
  let validatorPool: any
  let tripartyController: TripartyController
  let signers: any
  let poolOwner: any
  let eoaController: any
  let pauser: any

  const fixture = async function () {
    await deployments.fixture(["TripartyController"])
    validatorPool = new hre.ethers.Contract(
      validatorPoolPrecompileAddress,
      validatorpoolabi,
      ethers.provider,
    )
    assetsBridge = new hre.ethers.Contract(
      assetsBridgePrecompileAddress,
      assetsbridgeabi,
      ethers.provider,
    )
    btcToken = new hre.ethers.Contract(
      btcTokenPrecompileAddress,
      btcabi,
      ethers.provider,
    )
    tripartyController = await getDeployedContract("TripartyController")
    signers = await ethers.getSigners()
    poolOwner = await ethers.getSigner(await validatorPool.owner())
    eoaController = signers[1]
    pauser = signers[2]

    // Set up pauser
    await (
      await assetsBridge.connect(poolOwner).setPauser(pauser.address)
    ).wait()

    // Reset triparty state for test isolation (live chain state persists
    // across describe blocks, unlike hardhat network snapshots).
    const isPaused = await assetsBridge.isTripartyPaused()
    if (isPaused) {
      await (
        await assetsBridge.connect(pauser).pauseTriparty(false)
      ).wait()
    }
    await (
      await assetsBridge.connect(poolOwner).setTripartyBlockDelay(1)
    ).wait()
    await (
      await assetsBridge
        .connect(poolOwner)
        .allowTripartyController(eoaController.address, false)
    ).wait()
    const controllerAddr = await tripartyController.getAddress()
    await (
      await assetsBridge
        .connect(poolOwner)
        .allowTripartyController(controllerAddr, false)
    ).wait()
  }

  // ─── Configuration & Access Control ────────────────────────────────

  describe("Allow a triparty controller", function () {
    let isAllowed: boolean
    let nonOwnerError: string

    before(async function () {
      await fixture()

      await (
        await assetsBridge
          .connect(poolOwner)
          .allowTripartyController(eoaController.address, true)
      ).wait()

      isAllowed = await assetsBridge.isAllowedTripartyController(
        eoaController.address,
      )

      try {
        await assetsBridge
          .connect(eoaController)
          .allowTripartyController.staticCall(pauser.address, true)
      } catch (error: any) {
        nonOwnerError = error.message
      }
    })

    it("should return true for allowed controller", async function () {
      expect(isAllowed).to.equal(true)
    })

    it("should revert when non-owner tries to allow", async function () {
      expect(nonOwnerError).to.include("not the owner")
    })
  })

  describe("Disallow a triparty controller", function () {
    let isAllowed: boolean
    let bridgeError: string

    before(async function () {
      await fixture()

      // Allow then disallow
      await (
        await assetsBridge
          .connect(poolOwner)
          .allowTripartyController(eoaController.address, true)
      ).wait()
      await (
        await assetsBridge
          .connect(poolOwner)
          .allowTripartyController(eoaController.address, false)
      ).wait()

      isAllowed = await assetsBridge.isAllowedTripartyController(
        eoaController.address,
      )

      try {
        await assetsBridge
          .connect(eoaController)
          .bridgeTriparty.staticCall(signers[0].address, BTC(1), "0x")
      } catch (error: any) {
        bridgeError = error.message
      }
    })

    it("should return false for disallowed controller", async function () {
      expect(isAllowed).to.equal(false)
    })

    it("should revert bridgeTriparty from disallowed controller", async function () {
      expect(bridgeError).to.include("not an allowed triparty controller")
    })
  })

  describe("Set block delay", function () {
    let defaultDelay: bigint
    let updatedDelay: bigint
    let zeroDelayError: string
    let nonOwnerError: string

    before(async function () {
      await fixture()

      defaultDelay = await assetsBridge.getTripartyBlockDelay()

      await (
        await assetsBridge.connect(poolOwner).setTripartyBlockDelay(5)
      ).wait()

      updatedDelay = await assetsBridge.getTripartyBlockDelay()

      try {
        await assetsBridge
          .connect(poolOwner)
          .setTripartyBlockDelay.staticCall(0)
      } catch (error: any) {
        zeroDelayError = error.message
      }

      try {
        await assetsBridge
          .connect(eoaController)
          .setTripartyBlockDelay.staticCall(2)
      } catch (error: any) {
        nonOwnerError = error.message
      }

      // Reset delay to 1 for other tests
      await (
        await assetsBridge.connect(poolOwner).setTripartyBlockDelay(1)
      ).wait()
    })

    it("should have default delay of 1", async function () {
      expect(defaultDelay).to.equal(1n)
    })

    it("should return 5 after setting to 5", async function () {
      expect(updatedDelay).to.equal(5n)
    })

    it("should revert when setting delay to 0", async function () {
      expect(zeroDelayError).to.include("delay must be at least 1")
    })

    it("should revert when non-owner sets delay", async function () {
      expect(nonOwnerError).to.include("not the owner")
    })
  })

  describe("Set limits", function () {
    let limits: any
    let capacity: any
    let nonOwnerError: string

    before(async function () {
      await fixture()

      await (
        await assetsBridge
          .connect(poolOwner)
          .setTripartyLimits(BTC(5), BTC(50))
      ).wait()

      limits = await assetsBridge.getTripartyLimits()
      capacity = await assetsBridge.getTripartyCapacity()

      try {
        await assetsBridge
          .connect(eoaController)
          .setTripartyLimits.staticCall(BTC(1), BTC(10))
      } catch (error: any) {
        nonOwnerError = error.message
      }
    })

    it("should return correct limits", async function () {
      expect(limits.perRequestLimit).to.equal(BTC(5))
      expect(limits.windowLimit).to.equal(BTC(50))
    })

    it("should return valid capacity", async function () {
      expect(capacity.capacity).to.be.lte(BTC(50))
      expect(capacity.capacity).to.be.gt(0n)
      expect(capacity.resetHeight).to.be.gt(0n)
    })

    it("should revert when non-owner sets limits", async function () {
      expect(nonOwnerError).to.include("not the owner")
    })
  })

  describe("Setting paused flag", function () {
    let defaultPaused: boolean
    let afterPause: boolean
    let afterUnpause: boolean

    before(async function () {
      await fixture()

      defaultPaused = await assetsBridge.isTripartyPaused()

      await (
        await assetsBridge.connect(pauser).pauseTriparty(true)
      ).wait()
      afterPause = await assetsBridge.isTripartyPaused()

      await (
        await assetsBridge.connect(pauser).pauseTriparty(false)
      ).wait()
      afterUnpause = await assetsBridge.isTripartyPaused()
    })

    it("should be false by default", async function () {
      expect(defaultPaused).to.equal(false)
    })

    it("should be true after pause", async function () {
      expect(afterPause).to.equal(true)
    })

    it("should be false after unpause", async function () {
      expect(afterUnpause).to.equal(false)
    })
  })

  // ─── Happy Path — End-to-End Mint ──────────────────────────────────

  describe("Triparty mint with default delay", function () {
    let recipientBalanceBefore: bigint
    let totalSupplyBefore: bigint
    let requestTipBefore: bigint
    let recipientBalanceAfter: bigint
    let totalSupplyAfter: bigint
    let requestTipAfter: bigint
    let processedTipAfter: bigint
    let erc20BalanceAfter: bigint
    let recipient: string

    before(async function () {
      await fixture()

      const recipientWallet = ethers.Wallet.createRandom()
      recipient = recipientWallet.address

      // Allow eoaController as controller
      await (
        await assetsBridge
          .connect(poolOwner)
          .allowTripartyController(eoaController.address, true)
      ).wait()
      // Set limits
      await (
        await assetsBridge
          .connect(poolOwner)
          .setTripartyLimits(BTC(10), BTC(100))
      ).wait()

      recipientBalanceBefore = await ethers.provider.getBalance(recipient)
      totalSupplyBefore = await btcToken.totalSupply()
      requestTipBefore = await assetsBridge.getTripartyRequestSequenceTip()

      // Submit triparty request
      await (
        await assetsBridge
          .connect(eoaController)
          .bridgeTriparty(recipient, BTC(1), "0x")
      ).wait()

      requestTipAfter = await assetsBridge.getTripartyRequestSequenceTip()

      // Wait for processing (delay=1 + 1 block for processing)
      const currentBlock = await ethers.provider.getBlockNumber()
      await waitForBlock(currentBlock + 2)

      recipientBalanceAfter = await ethers.provider.getBalance(recipient)
      erc20BalanceAfter = await btcToken.balanceOf(recipient)
      totalSupplyAfter = await btcToken.totalSupply()
      processedTipAfter = await assetsBridge.getTripartyProcessedSequenceTip()
    })

    it("should increment request sequence tip", async function () {
      expect(requestTipAfter).to.equal(requestTipBefore + 1n)
    })

    it("should increase recipient balance by 1 BTC", async function () {
      expect(recipientBalanceAfter).to.equal(recipientBalanceBefore + BTC(1))
    })

    it("should have matching native and ERC20 balances", async function () {
      expect(recipientBalanceAfter).to.equal(erc20BalanceAfter)
    })

    it("should increase total supply by 1 BTC", async function () {
      expect(totalSupplyAfter).to.equal(totalSupplyBefore + BTC(1))
    })

    it("should have no pending requests", async function () {
      expect(processedTipAfter).to.equal(requestTipAfter)
    })

    it("should continue producing blocks", async function () {
      const blockBefore = await ethers.provider.getBlockNumber()
      await waitForBlock(blockBefore + 3)
      const blockAfter = await ethers.provider.getBlockNumber()
      expect(blockAfter).to.be.gte(blockBefore + 3)
    })
  })

  describe("Triparty mint with longer delay", function () {
    let recipient: string
    let recipientBalanceBefore: bigint
    let balanceAtN1: bigint
    let balanceAtN2: bigint
    let balanceAtN3: bigint
    let submitBlock: number

    before(async function () {
      await fixture()

      const recipientWallet = ethers.Wallet.createRandom()
      recipient = recipientWallet.address

      await (
        await assetsBridge
          .connect(poolOwner)
          .allowTripartyController(eoaController.address, true)
      ).wait()
      await (
        await assetsBridge
          .connect(poolOwner)
          .setTripartyLimits(BTC(10), BTC(100))
      ).wait()
      await (
        await assetsBridge.connect(poolOwner).setTripartyBlockDelay(3)
      ).wait()

      recipientBalanceBefore = await ethers.provider.getBalance(recipient)

      const tx = await assetsBridge
        .connect(eoaController)
        .bridgeTriparty(recipient, BTC(1), "0x")
      const receipt = await tx.wait()
      submitBlock = receipt.blockNumber

      await waitForBlock(submitBlock + 1)
      balanceAtN1 = await ethers.provider.getBalance(recipient)

      await waitForBlock(submitBlock + 2)
      balanceAtN2 = await ethers.provider.getBalance(recipient)

      await waitForBlock(submitBlock + 3)
      // Give one extra block for processing
      await waitForBlock(submitBlock + 4)
      balanceAtN3 = await ethers.provider.getBalance(recipient)

      // Reset delay
      await (
        await assetsBridge.connect(poolOwner).setTripartyBlockDelay(1)
      ).wait()
    })

    it("should not mint at N+1", async function () {
      expect(balanceAtN1).to.equal(recipientBalanceBefore)
    })

    it("should not mint at N+2", async function () {
      expect(balanceAtN2).to.equal(recipientBalanceBefore)
    })

    it("should mint by N+3", async function () {
      expect(balanceAtN3).to.equal(recipientBalanceBefore + BTC(1))
    })
  })

  describe("Triparty mint with callback", function () {
    let recipient: string
    let recipientBalanceBefore: bigint
    let recipientBalanceAfter: bigint
    let callbackCount: bigint
    let callbackRecord: any
    let requestId: bigint
    let callbackData: string

    before(async function () {
      await fixture()

      const recipientWallet = ethers.Wallet.createRandom()
      recipient = recipientWallet.address

      const controllerAddress = await tripartyController.getAddress()

      await (
        await assetsBridge
          .connect(poolOwner)
          .allowTripartyController(controllerAddress, true)
      ).wait()
      await (
        await assetsBridge
          .connect(poolOwner)
          .setTripartyLimits(BTC(10), BTC(100))
      ).wait()

      recipientBalanceBefore = await ethers.provider.getBalance(recipient)

      callbackData = ethers.AbiCoder.defaultAbiCoder().encode(
        ["uint256"],
        [12345],
      )

      const tx = await tripartyController.requestMint(
        recipient,
        BTC("0.5"),
        callbackData,
      )
      const receipt = await tx.wait()

      // Extract requestId from TripartyBridgeRequested event
      const iface = new ethers.Interface(assetsbridgeabi)
      for (const log of receipt.logs) {
        try {
          const parsed = iface.parseLog({
            topics: log.topics as string[],
            data: log.data,
          })
          if (parsed && parsed.name === "TripartyBridgeRequested") {
            requestId = parsed.args.requestId
            break
          }
          if (parsed) {
            console.log(`Ignoring event: ${parsed.name}`)
          }
        } catch {
          // Not a decodable event from our ABI, skip.
        }
      }

      const currentBlock = await ethers.provider.getBlockNumber()
      await waitForBlock(currentBlock + 3)

      recipientBalanceAfter = await ethers.provider.getBalance(recipient)
      callbackCount = await tripartyController.getCallbackCount()
      callbackRecord = await tripartyController.getCallback(0)
    })

    it("should mint 0.5 BTC to recipient", async function () {
      expect(recipientBalanceAfter).to.equal(
        recipientBalanceBefore + BTC("0.5"),
      )
    })

    it("should record exactly one callback", async function () {
      expect(callbackCount).to.equal(1n)
    })

    it("should record correct callback data", async function () {
      expect(callbackRecord[0]).to.equal(requestId)
      expect(callbackRecord[1]).to.equal(recipient)
      expect(callbackRecord[2]).to.equal(BTC("0.5"))
      expect(callbackRecord[3]).to.equal(callbackData)
    })
  })

  describe("Triparty mint through EOA controller", function () {
    let recipient: string
    let recipientBalanceBefore: bigint
    let recipientBalanceAfter: bigint
    let totalSupplyBefore: bigint
    let totalSupplyAfter: bigint

    before(async function () {
      await fixture()

      const recipientWallet = ethers.Wallet.createRandom()
      recipient = recipientWallet.address

      // Allow an EOA as controller — it has no code so the callback
      // call targeting onTripartyBridgeCompleted cannot succeed.
      await (
        await assetsBridge
          .connect(poolOwner)
          .allowTripartyController(eoaController.address, true)
      ).wait()
      await (
        await assetsBridge
          .connect(poolOwner)
          .setTripartyLimits(BTC(10), BTC(100))
      ).wait()

      recipientBalanceBefore = await ethers.provider.getBalance(recipient)
      totalSupplyBefore = await btcToken.totalSupply()

      await (
        await assetsBridge
          .connect(eoaController)
          .bridgeTriparty(
            recipient,
            BTC("0.1"),
            ethers.AbiCoder.defaultAbiCoder().encode(["uint256"], [99]),
          )
      ).wait()

      const currentBlock = await ethers.provider.getBlockNumber()
      await waitForBlock(currentBlock + 3)

      recipientBalanceAfter = await ethers.provider.getBalance(recipient)
      totalSupplyAfter = await btcToken.totalSupply()
    })

    it("should mint despite callback to EOA", async function () {
      expect(recipientBalanceAfter).to.equal(
        recipientBalanceBefore + BTC("0.1"),
      )
    })

    it("should increase total supply", async function () {
      expect(totalSupplyAfter).to.equal(totalSupplyBefore + BTC("0.1"))
    })
  })

  describe("Callback failure does not block mint", function () {
    let recipient: string
    let recipientBalanceBefore: bigint
    let recipientBalanceAfter: bigint
    let callbackCountBefore: bigint
    let callbackCountAfter: bigint
    let totalSupplyBefore: bigint
    let totalSupplyAfter: bigint

    before(async function () {
      await fixture()

      const recipientWallet = ethers.Wallet.createRandom()
      recipient = recipientWallet.address

      const controllerAddress = await tripartyController.getAddress()

      await (
        await assetsBridge
          .connect(poolOwner)
          .allowTripartyController(controllerAddress, true)
      ).wait()
      await (
        await assetsBridge
          .connect(poolOwner)
          .setTripartyLimits(BTC(10), BTC(100))
      ).wait()

      // Enable revert on callback
      await (await tripartyController.setRevertOnCallback(true)).wait()

      recipientBalanceBefore = await ethers.provider.getBalance(recipient)
      totalSupplyBefore = await btcToken.totalSupply()
      callbackCountBefore = await tripartyController.getCallbackCount()

      await (
        await tripartyController.requestMint(recipient, BTC("0.1"), "0x")
      ).wait()

      const currentBlock = await ethers.provider.getBlockNumber()
      await waitForBlock(currentBlock + 3)

      recipientBalanceAfter = await ethers.provider.getBalance(recipient)
      totalSupplyAfter = await btcToken.totalSupply()
      callbackCountAfter = await tripartyController.getCallbackCount()

      // Verify chain keeps producing
      const blockBefore = await ethers.provider.getBlockNumber()
      await waitForBlock(blockBefore + 3)
    })

    it("should mint despite callback revert", async function () {
      expect(recipientBalanceAfter).to.equal(
        recipientBalanceBefore + BTC("0.1"),
      )
    })

    it("should not record callback", async function () {
      expect(callbackCountAfter).to.equal(callbackCountBefore)
    })

    it("should increase total supply", async function () {
      expect(totalSupplyAfter).to.equal(totalSupplyBefore + BTC("0.1"))
    })
  })

  describe("Callback gas cap", function () {
    let recipient: string
    let recipientBalanceBefore: bigint
    let recipientBalanceAfter: bigint
    let gasSinkLength: bigint

    before(async function () {
      await fixture()

      const controllerAddress = await tripartyController.getAddress()

      await (
        await assetsBridge
          .connect(poolOwner)
          .allowTripartyController(controllerAddress, true)
      ).wait()
      await (
        await assetsBridge
          .connect(poolOwner)
          .setTripartyLimits(BTC(10), BTC(100))
      ).wait()

      // Enable gas-wasting mode: the callback writes 100 storage slots
      // (~2.2M gas), exceeding the 1M callback gas cap
      // (TripartyCallbackGasLimit in x/evm/types/call.go).
      await (await tripartyController.setWasteGasOnCallback(true)).wait()

      const recipientWallet = ethers.Wallet.createRandom()
      recipient = recipientWallet.address

      recipientBalanceBefore = await ethers.provider.getBalance(recipient)

      await (
        await tripartyController.requestMint(recipient, BTC("0.1"), "0x")
      ).wait()

      const currentBlock = await ethers.provider.getBlockNumber()
      await waitForBlock(currentBlock + 2)

      recipientBalanceAfter = await ethers.provider.getBalance(recipient)
      gasSinkLength = await tripartyController.getGasSinkLength()
    })

    it("should mint BTC despite callback exceeding gas cap", async function () {
      expect(recipientBalanceAfter).to.equal(
        recipientBalanceBefore + BTC("0.1"),
      )
    })

    it("should not persist callback state changes", async function () {
      expect(gasSinkLength).to.equal(0n)
    })
  })

  // ─── Validation & Rejection ────────────────────────────────────────

  describe("Unauthorized caller", function () {
    let unauthorizedError: string

    before(async function () {
      await fixture()

      // Set limits but do NOT allow any controller
      await (
        await assetsBridge
          .connect(poolOwner)
          .setTripartyLimits(BTC(10), BTC(100))
      ).wait()

      const recipient = ethers.Wallet.createRandom().address

      try {
        await assetsBridge
          .connect(eoaController)
          .bridgeTriparty.staticCall(recipient, BTC(1), "0x")
      } catch (error: any) {
        unauthorizedError = error.message
      }
    })

    it("should revert for unauthorized caller", async function () {
      expect(unauthorizedError).to.include(
        "not an allowed triparty controller",
      )
    })
  })

  describe("Amount below minimum", function () {
    let belowMinError: string
    let minAmountReceipt: any

    before(async function () {
      await fixture()

      await (
        await assetsBridge
          .connect(poolOwner)
          .allowTripartyController(eoaController.address, true)
      ).wait()
      await (
        await assetsBridge
          .connect(poolOwner)
          .setTripartyLimits(BTC(10), BTC(100))
      ).wait()

      const recipient = ethers.Wallet.createRandom().address

      // 0.001 BTC should fail (below 0.01 minimum)
      try {
        await assetsBridge
          .connect(eoaController)
          .bridgeTriparty.staticCall(recipient, BTC("0.001"), "0x")
      } catch (error: any) {
        belowMinError = error.message
      }

      // 0.01 BTC should succeed
      const tx = await assetsBridge
        .connect(eoaController)
        .bridgeTriparty(recipient, BTC("0.01"), "0x")
      minAmountReceipt = await tx.wait()
    })

    it("should revert for 0.001 BTC", async function () {
      expect(belowMinError).to.include("triparty amount below minimum")
    })

    it("should succeed for 0.01 BTC", async function () {
      expect(minAmountReceipt.status).to.equal(1)
    })
  })

  describe("Amount exceeds per-request limit", function () {
    let exceedsError: string
    let withinLimitReceipt: any

    before(async function () {
      await fixture()

      await (
        await assetsBridge
          .connect(poolOwner)
          .allowTripartyController(eoaController.address, true)
      ).wait()
      await (
        await assetsBridge
          .connect(poolOwner)
          .setTripartyLimits(BTC(2), BTC(100))
      ).wait()

      const recipient = ethers.Wallet.createRandom().address

      try {
        await assetsBridge
          .connect(eoaController)
          .bridgeTriparty.staticCall(recipient, BTC(3), "0x")
      } catch (error: any) {
        exceedsError = error.message
      }

      const tx = await assetsBridge
        .connect(eoaController)
        .bridgeTriparty(recipient, BTC(2), "0x")
      withinLimitReceipt = await tx.wait()
    })

    it("should revert for 3 BTC (over 2 BTC limit)", async function () {
      expect(exceedsError).to.include("triparty per-request limit exceeded")
    })

    it("should succeed for 2 BTC (at limit)", async function () {
      expect(withinLimitReceipt.status).to.equal(1)
    })
  })

  describe("Amount exceeds window capacity", function () {
    let capacityAfterFirst: any
    let exceedsWindowError: string
    let secondReceipt: any
    let exhaustedError: string

    before(async function () {
      await fixture()

      await (
        await assetsBridge
          .connect(poolOwner)
          .allowTripartyController(eoaController.address, true)
      ).wait()

      // On a live chain the 25k-block window accumulates across test runs.
      // Measure current usage, then set windowLimit so that exactly 5 BTC
      // of capacity remains, making the test deterministic.
      await (
        await assetsBridge
          .connect(poolOwner)
          .setTripartyLimits(BTC(5), BTC(1000))
      ).wait()
      const { capacity: currentCapacity } =
        await assetsBridge.getTripartyCapacity()
      const currentUsage = BTC(1000) - currentCapacity
      const windowLimit = currentUsage + BTC(5)
      await (
        await assetsBridge
          .connect(poolOwner)
          .setTripartyLimits(BTC(5), windowLimit)
      ).wait()

      const recipient = ethers.Wallet.createRandom().address

      // Submit 3 BTC — succeeds
      await (
        await assetsBridge
          .connect(eoaController)
          .bridgeTriparty(recipient, BTC(3), "0x")
      ).wait()

      capacityAfterFirst = await assetsBridge.getTripartyCapacity()

      // Submit 3 BTC — should fail (only 2 remaining)
      try {
        await assetsBridge
          .connect(eoaController)
          .bridgeTriparty.staticCall(recipient, BTC(3), "0x")
      } catch (error: any) {
        exceedsWindowError = error.message
      }

      // Submit 2 BTC — should succeed
      const tx = await assetsBridge
        .connect(eoaController)
        .bridgeTriparty(recipient, BTC(2), "0x")
      secondReceipt = await tx.wait()

      // Submit any more — should fail (window exhausted)
      try {
        await assetsBridge
          .connect(eoaController)
          .bridgeTriparty.staticCall(recipient, BTC("0.01"), "0x")
      } catch (error: any) {
        exhaustedError = error.message
      }
    })

    it("should show 2 BTC remaining after 3 BTC request", async function () {
      expect(capacityAfterFirst.capacity).to.equal(BTC(2))
    })

    it("should revert for 3 BTC when only 2 remaining", async function () {
      expect(exceedsWindowError).to.include("triparty window limit exceeded")
    })

    it("should succeed for 2 BTC (remaining capacity)", async function () {
      expect(secondReceipt.status).to.equal(1)
    })

    it("should revert when window exhausted", async function () {
      expect(exhaustedError).to.include("triparty window limit exceeded")
    })
  })

  describe("Window limits shared across controllers", function () {
    let firstReceipt: any
    let exceedsError: string
    let secondReceipt: any

    before(async function () {
      await fixture()

      const controllerAddress = await tripartyController.getAddress()

      // Allow both an EOA and the contract as controllers
      await (
        await assetsBridge
          .connect(poolOwner)
          .allowTripartyController(eoaController.address, true)
      ).wait()
      await (
        await assetsBridge
          .connect(poolOwner)
          .allowTripartyController(controllerAddress, true)
      ).wait()

      // Measure current window usage, then set windowLimit so that
      // exactly 2 BTC of capacity remains (same technique as C4).
      await (
        await assetsBridge
          .connect(poolOwner)
          .setTripartyLimits(BTC(2), BTC(1000))
      ).wait()
      const { capacity: currentCapacity } =
        await assetsBridge.getTripartyCapacity()
      const currentUsage = BTC(1000) - currentCapacity
      const windowLimit = currentUsage + BTC(2)
      await (
        await assetsBridge
          .connect(poolOwner)
          .setTripartyLimits(BTC(2), windowLimit)
      ).wait()

      const recipient = ethers.Wallet.createRandom().address

      // Contract controller requests 1.5 BTC — succeeds
      const tx1 = await tripartyController.requestMint(
        recipient,
        BTC("1.5"),
        "0x",
      )
      firstReceipt = await tx1.wait()

      // EOA controller requests 1 BTC — should fail (only 0.5 remaining).
      // Using the EOA's direct staticCall to get the precise revert reason
      // (calls through a contract wrapper lose the inner revert string).
      try {
        await assetsBridge
          .connect(eoaController)
          .bridgeTriparty.staticCall(recipient, BTC(1), "0x")
      } catch (error: any) {
        exceedsError = error.message
      }

      // EOA controller requests 0.5 BTC — succeeds
      const tx2 = await assetsBridge
        .connect(eoaController)
        .bridgeTriparty(recipient, BTC("0.5"), "0x")
      secondReceipt = await tx2.wait()
    })

    it("should allow first controller to use capacity", async function () {
      expect(firstReceipt.status).to.equal(1)
    })

    it("should reject second controller exceeding shared capacity", async function () {
      expect(exceedsError).to.include("triparty window limit exceeded")
    })

    it("should allow second controller to use remaining capacity", async function () {
      expect(secondReceipt.status).to.equal(1)
    })
  })

  describe("Blocked/invalid recipient", function () {
    let btcPrecompileError: string
    let zeroAddressError: string

    before(async function () {
      await fixture()

      await (
        await assetsBridge
          .connect(poolOwner)
          .allowTripartyController(eoaController.address, true)
      ).wait()
      await (
        await assetsBridge
          .connect(poolOwner)
          .setTripartyLimits(BTC(10), BTC(100))
      ).wait()

      // BTC precompile address as recipient
      try {
        await assetsBridge
          .connect(eoaController)
          .bridgeTriparty.staticCall(btcTokenPrecompileAddress, BTC(1), "0x")
      } catch (error: any) {
        btcPrecompileError = error.message
      }

      // Zero address as recipient
      try {
        await assetsBridge
          .connect(eoaController)
          .bridgeTriparty.staticCall(ethers.ZeroAddress, BTC(1), "0x")
      } catch (error: any) {
        zeroAddressError = error.message
      }
    })

    it("should revert for BTC precompile address", async function () {
      expect(btcPrecompileError).to.include("triparty recipient")
    })

    it("should revert for zero address", async function () {
      expect(zeroAddressError).to.include("zero")
    })
  })

  // ─── Pause Mechanism ───────────────────────────────────────────────

  describe("Pause blocks new requests", function () {
    let isPaused: boolean
    let bridgeError: string
    let nonPauserError: string

    before(async function () {
      await fixture()

      await (
        await assetsBridge
          .connect(poolOwner)
          .allowTripartyController(eoaController.address, true)
      ).wait()
      await (
        await assetsBridge
          .connect(poolOwner)
          .setTripartyLimits(BTC(10), BTC(100))
      ).wait()

      // Pause
      await (
        await assetsBridge.connect(pauser).pauseTriparty(true)
      ).wait()

      isPaused = await assetsBridge.isTripartyPaused()

      const recipient = ethers.Wallet.createRandom().address

      try {
        await assetsBridge
          .connect(eoaController)
          .bridgeTriparty.staticCall(recipient, BTC(1), "0x")
      } catch (error: any) {
        bridgeError = error.message
      }

      try {
        await assetsBridge
          .connect(eoaController)
          .pauseTriparty.staticCall(true)
      } catch (error: any) {
        nonPauserError = error.message
      }
    })

    it("should be paused", async function () {
      expect(isPaused).to.equal(true)
    })

    it("should revert bridgeTriparty when paused", async function () {
      expect(bridgeError).to.include("triparty bridging is paused")
    })

    it("should revert pause from non-pauser", async function () {
      expect(nonPauserError).to.include("caller is not the pauser")
    })
  })

  describe("Pause freezes pending, unpause resumes", function () {
    let recipient: string
    let processedBefore: bigint
    let requestTipAfterSubmit: bigint
    let processedWhilePaused: bigint
    let recipientBalanceWhilePaused: bigint
    let recipientBalanceBefore: bigint
    let isPausedAfterUnpause: boolean
    let processedAfterUnpause: bigint
    let requestTipFinal: bigint
    let recipientBalanceAfterUnpause: bigint

    before(async function () {
      await fixture()

      const recipientWallet = ethers.Wallet.createRandom()
      recipient = recipientWallet.address

      await (
        await assetsBridge
          .connect(poolOwner)
          .allowTripartyController(eoaController.address, true)
      ).wait()
      await (
        await assetsBridge
          .connect(poolOwner)
          .setTripartyLimits(BTC(10), BTC(100))
      ).wait()
      // Use delay=3 so we have time to pause before processing
      await (
        await assetsBridge.connect(poolOwner).setTripartyBlockDelay(3)
      ).wait()

      recipientBalanceBefore = await ethers.provider.getBalance(recipient)
      processedBefore = await assetsBridge.getTripartyProcessedSequenceTip()

      // Submit request
      await (
        await assetsBridge
          .connect(eoaController)
          .bridgeTriparty(recipient, BTC(1), "0x")
      ).wait()

      requestTipAfterSubmit =
        await assetsBridge.getTripartyRequestSequenceTip()

      // Immediately pause
      await (
        await assetsBridge.connect(pauser).pauseTriparty(true)
      ).wait()

      // Wait several blocks
      const currentBlock = await ethers.provider.getBlockNumber()
      await waitForBlock(currentBlock + 5)

      recipientBalanceWhilePaused =
        await ethers.provider.getBalance(recipient)
      processedWhilePaused =
        await assetsBridge.getTripartyProcessedSequenceTip()

      // --- D3: Unpause ---
      await (
        await assetsBridge.connect(pauser).pauseTriparty(false)
      ).wait()

      isPausedAfterUnpause = await assetsBridge.isTripartyPaused()

      // Wait for processing
      const blockAfterUnpause = await ethers.provider.getBlockNumber()
      await waitForBlock(blockAfterUnpause + 5)

      recipientBalanceAfterUnpause =
        await ethers.provider.getBalance(recipient)
      processedAfterUnpause =
        await assetsBridge.getTripartyProcessedSequenceTip()
      requestTipFinal = await assetsBridge.getTripartyRequestSequenceTip()

      // Reset delay
      await (
        await assetsBridge.connect(poolOwner).setTripartyBlockDelay(1)
      ).wait()
    })

    it("should not change recipient balance while paused", async function () {
      expect(recipientBalanceWhilePaused).to.equal(recipientBalanceBefore)
    })

    it("should not advance processed tip while paused", async function () {
      expect(processedWhilePaused).to.equal(processedBefore)
    })

    it("should have created the request", async function () {
      expect(requestTipAfterSubmit).to.equal(processedBefore + 1n)
    })

    it("should be unpaused after unpause", async function () {
      expect(isPausedAfterUnpause).to.equal(false)
    })

    it("should deliver BTC after unpause", async function () {
      expect(recipientBalanceAfterUnpause).to.equal(
        recipientBalanceBefore + BTC(1),
      )
    })

    it("should catch up processed tip after unpause", async function () {
      expect(processedAfterUnpause).to.equal(requestTipFinal)
    })
  })

  // ─── Batch & Ordering ──────────────────────────────────────────────

  describe("Multiple requests in a single block", function () {
    let recipients: string[]
    let balancesBefore: bigint[]
    let balancesAfter: bigint[]

    before(async function () {
      await fixture()

      const controllerAddress = await tripartyController.getAddress()
      await (
        await assetsBridge
          .connect(poolOwner)
          .allowTripartyController(controllerAddress, true)
      ).wait()
      await (
        await assetsBridge
          .connect(poolOwner)
          .setTripartyLimits(BTC(10), BTC(100))
      ).wait()

      recipients = Array.from({ length: 3 }, () =>
        ethers.Wallet.createRandom().address,
      )

      balancesBefore = await Promise.all(
        recipients.map((r) => ethers.provider.getBalance(r)),
      )

      // Submit 3 requests in a single transaction via batchRequestMint
      // to guarantee same-block delivery on a live chain.
      const amounts = [BTC("0.1"), BTC("0.2"), BTC("0.3")]
      const tx = await tripartyController.batchRequestMint(
        recipients,
        amounts,
      )
      await tx.wait()

      const currentBlock = await ethers.provider.getBlockNumber()
      await waitForBlock(currentBlock + 3)

      balancesAfter = await Promise.all(
        recipients.map((r) => ethers.provider.getBalance(r)),
      )
    })

    it("should deliver BTC to all 3 recipients", async function () {
      const amounts = [BTC("0.1"), BTC("0.2"), BTC("0.3")]
      for (let i = 0; i < 3; i++) {
        expect(balancesAfter[i]).to.equal(balancesBefore[i] + amounts[i])
      }
    })
  })

  describe("Batch cap (5 per block)", function () {
    let recipients: string[]
    let tipBefore: bigint
    let processedAfterFirst: bigint
    let processedAfterSecond: bigint
    let balancesAfterFirst: bigint[]
    let balancesAfterSecond: bigint[]

    before(async function () {
      await fixture()

      const controllerAddress = await tripartyController.getAddress()
      await (
        await assetsBridge
          .connect(poolOwner)
          .allowTripartyController(controllerAddress, true)
      ).wait()
      await (
        await assetsBridge
          .connect(poolOwner)
          .setTripartyLimits(BTC(10), BTC(100))
      ).wait()

      recipients = Array.from({ length: 7 }, () =>
        ethers.Wallet.createRandom().address,
      )

      tipBefore = await assetsBridge.getTripartyProcessedSequenceTip()

      // Submit all 7 in a single transaction via batchRequestMint so they
      // share the same submission block and mature simultaneously.
      const amounts = recipients.map(() => BTC("0.1"))
      const tx = await tripartyController.batchRequestMint(
        recipients,
        amounts,
      )
      const receipt = await tx.wait()
      const submitBlock = receipt!.blockNumber

      // delay=1: all 7 mature at submitBlock+1.
      // PreBlocker at submitBlock+1 processes the first 5 (batch cap).
      // Use blockTag to read state at the exact target block, avoiding a
      // race where the chain advances before the query lands.
      await waitForBlock(submitBlock + 1)

      ;[processedAfterFirst, ...balancesAfterFirst] = await Promise.all([
        assetsBridge.getTripartyProcessedSequenceTip({
          blockTag: submitBlock + 1,
        }),
        ...recipients.map((r: string) =>
          ethers.provider.getBalance(r, submitBlock + 1),
        ),
      ])

      // PreBlocker at submitBlock+2 processes the remaining 2.
      await waitForBlock(submitBlock + 2)

      ;[processedAfterSecond, ...balancesAfterSecond] = await Promise.all([
        assetsBridge.getTripartyProcessedSequenceTip({
          blockTag: submitBlock + 2,
        }),
        ...recipients.map((r: string) =>
          ethers.provider.getBalance(r, submitBlock + 2),
        ),
      ])
    })

    it("should process exactly 5 in first batch", async function () {
      expect(processedAfterFirst).to.equal(tipBefore + 5n)
    })

    it("should have 5 recipients with BTC after first batch", async function () {
      const funded = balancesAfterFirst.filter((b) => b > 0n).length
      expect(funded).to.equal(5)
    })

    it("should process all 7 after second batch", async function () {
      expect(processedAfterSecond).to.equal(tipBefore + 7n)
    })

    it("should have all 7 recipients with BTC after second batch", async function () {
      const funded = balancesAfterSecond.filter((b) => b > 0n).length
      expect(funded).to.equal(7)
    })
  })

  describe("Immature request blocks later requests", function () {
    let recipients: string[]
    let tipBefore: bigint
    let balanceR1AtR1Mature: bigint
    let balanceR2AtR1Mature: bigint
    let processedAtR1Mature: bigint
    let balanceR1AtBothMature: bigint
    let balanceR2AtBothMature: bigint
    let processedAtBothMature: bigint

    before(async function () {
      await fixture()

      await (
        await assetsBridge
          .connect(poolOwner)
          .allowTripartyController(eoaController.address, true)
      ).wait()
      await (
        await assetsBridge
          .connect(poolOwner)
          .setTripartyLimits(BTC(10), BTC(100))
      ).wait()
      await (
        await assetsBridge.connect(poolOwner).setTripartyBlockDelay(3)
      ).wait()

      recipients = [
        ethers.Wallet.createRandom().address,
        ethers.Wallet.createRandom().address,
      ]

      tipBefore = await assetsBridge.getTripartyProcessedSequenceTip()

      // Submit R1
      const tx1 = await assetsBridge
        .connect(eoaController)
        .bridgeTriparty(recipients[0], BTC("0.5"), "0x")
      const receipt1 = await tx1.wait()
      const submitBlockR1 = receipt1!.blockNumber

      // Wait 2 blocks to create a gap between maturities, then submit R2
      await waitForBlock(submitBlockR1 + 2)
      const tx2 = await assetsBridge
        .connect(eoaController)
        .bridgeTriparty(recipients[1], BTC("0.5"), "0x")
      const receipt2 = await tx2.wait()
      const submitBlockR2 = receipt2!.blockNumber

      // R1 matures at submitBlockR1+3, R2 matures at submitBlockR2+3.
      // Since submitBlockR2 >= submitBlockR1+2, R2 matures at least 2 blocks
      // after R1. Read state right after R1's maturity block.
      await waitForBlock(submitBlockR1 + 3)
      ;[balanceR1AtR1Mature, balanceR2AtR1Mature, processedAtR1Mature] =
        await Promise.all([
          ethers.provider.getBalance(recipients[0]),
          ethers.provider.getBalance(recipients[1]),
          assetsBridge.getTripartyProcessedSequenceTip(),
        ])

      // Now wait for R2 to mature and be processed
      await waitForBlock(submitBlockR2 + 4)
      ;[balanceR1AtBothMature, balanceR2AtBothMature, processedAtBothMature] =
        await Promise.all([
          ethers.provider.getBalance(recipients[0]),
          ethers.provider.getBalance(recipients[1]),
          assetsBridge.getTripartyProcessedSequenceTip(),
        ])

      // Reset delay
      await (
        await assetsBridge.connect(poolOwner).setTripartyBlockDelay(1)
      ).wait()
    })

    it("should process R1 when mature", async function () {
      expect(balanceR1AtR1Mature).to.equal(BTC("0.5"))
    })

    it("should not process R2 when only R1 is mature", async function () {
      expect(balanceR2AtR1Mature).to.equal(0n)
    })

    it("should advance processed tip by 1 after R1 matures", async function () {
      expect(processedAtR1Mature).to.equal(tipBefore + 1n)
    })

    it("should process both when both are mature", async function () {
      expect(balanceR1AtBothMature).to.equal(BTC("0.5"))
      expect(balanceR2AtBothMature).to.equal(BTC("0.5"))
    })

    it("should advance processed tip by 2 after both mature", async function () {
      expect(processedAtBothMature).to.equal(tipBefore + 2n)
    })
  })

  // ─── Deauthorized Controller ───────────────────────────────────────

  describe("Pending request from deauthorized controller is skipped", function () {
    let recipient: string
    let recipientBalanceBefore: bigint
    let recipientBalanceAfter: bigint
    let tipBefore: bigint
    let requestTipAfter: bigint
    let processedTipAfter: bigint

    before(async function () {
      await fixture()

      const recipientWallet = ethers.Wallet.createRandom()
      recipient = recipientWallet.address

      await (
        await assetsBridge
          .connect(poolOwner)
          .allowTripartyController(eoaController.address, true)
      ).wait()
      await (
        await assetsBridge
          .connect(poolOwner)
          .setTripartyLimits(BTC(10), BTC(100))
      ).wait()
      await (
        await assetsBridge.connect(poolOwner).setTripartyBlockDelay(3)
      ).wait()

      tipBefore = await assetsBridge.getTripartyProcessedSequenceTip()
      recipientBalanceBefore = await ethers.provider.getBalance(recipient)

      // Submit request
      const tx = await assetsBridge
        .connect(eoaController)
        .bridgeTriparty(recipient, BTC(1), "0x")
      const receipt = await tx.wait()

      requestTipAfter = await assetsBridge.getTripartyRequestSequenceTip()

      // Deauthorize before processing
      await (
        await assetsBridge
          .connect(poolOwner)
          .allowTripartyController(eoaController.address, false)
      ).wait()

      // Wait past maturity
      const submitBlock = receipt.blockNumber
      await waitForBlock(submitBlock + 5)

      recipientBalanceAfter = await ethers.provider.getBalance(recipient)
      processedTipAfter = await assetsBridge.getTripartyProcessedSequenceTip()

      // Reset delay
      await (
        await assetsBridge.connect(poolOwner).setTripartyBlockDelay(1)
      ).wait()
    })

    it("should have created the request", async function () {
      expect(requestTipAfter).to.equal(tipBefore + 1n)
    })

    it("should not mint BTC (request skipped)", async function () {
      expect(recipientBalanceAfter).to.equal(recipientBalanceBefore)
    })

    it("should advance processed tip (request dequeued)", async function () {
      expect(processedTipAfter).to.equal(requestTipAfter)
    })
  })

  // ─── Provenance Tracking ───────────────────────────────────────────

  describe("Triparty minted counter", function () {
    let totalMintedBefore: bigint
    let totalMintedAfter: bigint
    let requestTipBefore: bigint
    let requestTipAfter: bigint
    let processedTipAfter: bigint

    before(async function () {
      await fixture()

      await (
        await assetsBridge
          .connect(poolOwner)
          .allowTripartyController(eoaController.address, true)
      ).wait()
      await (
        await assetsBridge
          .connect(poolOwner)
          .setTripartyLimits(BTC(10), BTC(100))
      ).wait()

      totalMintedBefore = await assetsBridge.getTripartyTotalBTCMinted()
      requestTipBefore = await assetsBridge.getTripartyRequestSequenceTip()

      // Submit 3 requests totaling 0.6 BTC
      const amounts = [BTC("0.1"), BTC("0.2"), BTC("0.3")]
      for (const amount of amounts) {
        const recipient = ethers.Wallet.createRandom().address
        const tx = await assetsBridge
          .connect(eoaController)
          .bridgeTriparty(recipient, amount, "0x")
        await tx.wait()
      }

      // Wait for all to be processed
      const currentBlock = await ethers.provider.getBlockNumber()
      await waitForBlock(currentBlock + 5)

      totalMintedAfter = await assetsBridge.getTripartyTotalBTCMinted()
      requestTipAfter = await assetsBridge.getTripartyRequestSequenceTip()
      processedTipAfter = await assetsBridge.getTripartyProcessedSequenceTip()
    })

    it("should increase total minted by 0.6 BTC", async function () {
      expect(totalMintedAfter).to.equal(totalMintedBefore + BTC("0.6"))
    })

    it("should advance request tip by 3", async function () {
      expect(requestTipAfter).to.equal(requestTipBefore + 3n)
    })

    it("should have processed all requests", async function () {
      expect(processedTipAfter).to.equal(requestTipAfter)
    })
  })
})
