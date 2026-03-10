import { expect } from "chai"
import hre from "hardhat"
import type { InitcodeLimitCheck } from "../typechain-types"
import { getDeployedContract } from "./helpers/contract"

describe("InitcodeLimitCheck", function () {
  const { deployments } = hre
  const maxInitcodeSize = 49152n
  const oversizedInitcodeSize = 49153n
  let initcodeLimitCheck: InitcodeLimitCheck
  let senderSigner: any

  const fixture = async function () {
    await deployments.fixture(["InitcodeLimitCheck"])
    initcodeLimitCheck = await getDeployedContract("InitcodeLimitCheck")
    const signers = await hre.ethers.getSigners()
    senderSigner = signers[1]
  }

  describe("deployWithInitcodeSize", function () {
    before(async function () {
      await fixture()
    })

    it("should deploy small initcode", async function () {
      const tx = await initcodeLimitCheck
        .connect(senderSigner)
        .deployWithInitcodeSize(5n, { gasLimit: 1_000_000 })
      await tx.wait()

      const deployed = await initcodeLimitCheck.lastDeployed()
      expect(deployed).to.not.equal("0x0000000000000000000000000000000000000000")
    })

    it("should deploy initcode at 49152 bytes", async function () {
      const tx = await initcodeLimitCheck
        .connect(senderSigner)
        .deployWithInitcodeSize(maxInitcodeSize, { gasLimit: 10_000_000 })
      await tx.wait()

      const deployed = await initcodeLimitCheck.lastDeployed()
      expect(deployed).to.not.equal("0x0000000000000000000000000000000000000000")
    })

    it("should reject initcode above 49152 bytes", async function () {
      await expect((async () => {
        const tx = await initcodeLimitCheck
          .connect(senderSigner)
          .deployWithInitcodeSize(oversizedInitcodeSize, { gasLimit: 10_000_000 })
        await tx.wait()
      })()).to.be.rejected
    })
  })

  describe("top-level create transaction", function () {
    before(async function () {
      await fixture()
    })

    it("should allow contract creation tx at 49152 bytes", async function () {
      const tx = await senderSigner.sendTransaction({
        to: null,
        data: "0x" + "00".repeat(Number(maxInitcodeSize)),
        gasLimit: 10_000_000,
      })
      const receipt = await tx.wait()

      expect(receipt?.status).to.equal(1)
      expect(receipt?.contractAddress).to.not.equal(null)
    })

    it("should reject oversized contract creation tx", async function () {
      await expect(
        senderSigner.sendTransaction({
          to: null,
          data: "0x" + "00".repeat(Number(oversizedInitcodeSize)),
          gasLimit: 10_000_000,
        }),
      ).to.be.rejected
    })
  })
})
