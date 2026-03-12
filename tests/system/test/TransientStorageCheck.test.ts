import { expect } from "chai"
import hre from "hardhat"
import type { TransientStorageCheck } from "../typechain-types"
import { getDeployedContract } from "./helpers/contract"

describe("TransientStorageCheck", function () {
  const { deployments } = hre
  let transientStorageCheck: TransientStorageCheck
  let senderSigner: any

  before(async function () {
    await deployments.fixture(["TransientStorageCheck"])
    transientStorageCheck = await getDeployedContract("TransientStorageCheck")
    const signers = await hre.ethers.getSigners()
    senderSigner = signers[1]
  })

  it("should load transient value in the same transaction", async function () {
    const tx = await transientStorageCheck.connect(senderSigner).setAndLoad(1n, 123n)
    await tx.wait()

    const value = await transientStorageCheck.lastLoaded()
    expect(value).to.equal(123n)
  })

  it("should clear transient value between transactions", async function () {
    const tx = await transientStorageCheck.connect(senderSigner).setAndLoad(1n, 123n)
    await tx.wait()
    const nextTxValue = await transientStorageCheck.load.staticCall(1n)

    expect(nextTxValue).to.equal(0n)
  })

  it("should keep transient value across same-contract call frames in one transaction", async function () {
    const tx = await transientStorageCheck
      .connect(senderSigner)
      .setAndExternalLoad(2n, 456n)
    await tx.wait()

    const value = await transientStorageCheck.lastLoaded()
    expect(value).to.equal(456n)
  })

  it("should fail TSTORE in STATICCALL context", async function () {
    const lastLoadedBefore = await transientStorageCheck.lastLoaded()
    const [success] = await transientStorageCheck.staticCallSetAndLoad.staticCall(
      3n,
      789n,
    )
    const lastLoadedAfter = await transientStorageCheck.lastLoaded()

    expect(success).to.equal(false)
    expect(lastLoadedAfter).to.equal(lastLoadedBefore)
  })
})
