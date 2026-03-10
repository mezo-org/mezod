import { expect } from "chai"
import hre, { ethers } from "hardhat"
import type { Selfdestruct6780Check } from "../typechain-types"
import { getDeployedContract } from "./helpers/contract"

describe("Selfdestruct6780Check", function () {
  const { deployments } = hre
  let selfdestructCheck: Selfdestruct6780Check
  let senderSigner: any
  let beneficiary: any

  const fixture = async function () {
    await deployments.fixture(["Selfdestruct6780Check"])
    selfdestructCheck = await getDeployedContract("Selfdestruct6780Check")
    const signers = await ethers.getSigners()
    senderSigner = signers[1]
    beneficiary = signers[2]
  }

  describe("destroy existing contract in later transaction", function () {
    let destructible: string
    let codeBefore: string
    let beneficiaryBalanceBefore: bigint
    let beneficiaryBalanceAfter: bigint
    const fundingAmount = 1n

    before(async function () {
      await fixture()

      const createTx = await selfdestructCheck.connect(senderSigner).createDestructible()
      await createTx.wait()
      destructible = await selfdestructCheck.lastDestructible()

      const fundTx = await senderSigner.sendTransaction({
        to: destructible,
        value: fundingAmount,
      })
      await fundTx.wait()

      codeBefore = await ethers.provider.getCode(destructible)

      beneficiaryBalanceBefore = await ethers.provider.getBalance(
        beneficiary.address,
      )
      const destroyTx = await selfdestructCheck.connect(senderSigner).destroyExisting(
        destructible,
        beneficiary.address,
      )
      await destroyTx.wait()
      beneficiaryBalanceAfter = await ethers.provider.getBalance(
        beneficiary.address,
      )
    })

    it("should have code before destroy", async function () {
      expect(codeBefore).to.not.equal("0x")
    })

    it("should keep code unchanged after destroy", async function () {
      const codeAfter = await ethers.provider.getCode(destructible)
      expect(codeAfter).to.equal(codeBefore)
    })

    it("should transfer balance to beneficiary", async function () {
      expect(beneficiaryBalanceAfter - beneficiaryBalanceBefore).to.equal(
        fundingAmount,
      )
    })
  })

  describe("create and destroy in same transaction", function () {
    let destructible: string
    const fundingAmount = 1n
    let beneficiaryBalanceBefore: bigint

    before(async function () {
      await fixture()
      beneficiaryBalanceBefore = await ethers.provider.getBalance(
        beneficiary.address,
      )

      const createAndDestroyTx = await selfdestructCheck
        .connect(senderSigner)
        .createAndDestroySameTx(beneficiary.address, { value: fundingAmount })
      await createAndDestroyTx.wait()

      destructible = await selfdestructCheck.lastDestructible()
    })

    it("should delete code", async function () {
      const codeAfter = await ethers.provider.getCode(destructible)
      expect(codeAfter).to.equal("0x")
    })

    it("should fail to call ping", async function () {
      const destructibleContract = await ethers.getContractAt(
        "DestructibleContract6780",
        destructible,
      )

      await expect(destructibleContract.ping()).to.be.rejected
    })

    it("should transfer balance to beneficiary", async function () {
      const beneficiaryBalanceAfter = await ethers.provider.getBalance(
        beneficiary.address,
      )
      expect(beneficiaryBalanceAfter - beneficiaryBalanceBefore).to.equal(
        fundingAmount,
      )
    })
  })
})
