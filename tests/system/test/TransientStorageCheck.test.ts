import { expect } from "chai"
import hre from "hardhat"
import type {
  TransientStorageCheck,
  TransientStorageReader,
} from "../typechain-types"
import { getDeployedContract } from "./helpers/contract"

describe("TransientStorageCheck", function () {
  const { deployments } = hre
  let transientStorageCheck: TransientStorageCheck
  let transientStorageReader: TransientStorageReader
  let senderSigner: any

  before(async function () {
    await deployments.fixture(["TransientStorageCheck"])
    transientStorageCheck = await getDeployedContract("TransientStorageCheck")
    transientStorageReader = await getDeployedContract("TransientStorageReader")
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
    const nextTxValue = await transientStorageCheck.load(1n)

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

  it("should not share transient value across different contracts", async function () {
    const tx = await transientStorageCheck
      .connect(senderSigner)
      .setAndOtherContractLoad(await transientStorageReader.getAddress(), 4n, 999n)
    await tx.wait()

    const value = await transientStorageCheck.lastLoaded()
    expect(value).to.equal(0n)
  })

  it("should use caller transient storage across DELEGATECALL", async function () {
    const tx = await transientStorageCheck
      .connect(senderSigner)
      .delegateCallSetAndLoad(await transientStorageReader.getAddress(), 5n, 321n)
    await tx.wait()

    expect(await transientStorageCheck.lastLoaded()).to.equal(321n)
    expect(await transientStorageReader.lastLoaded()).to.equal(0n)
  })

  it("should use caller transient storage across CALLCODE", async function () {
    const tx = await transientStorageCheck
      .connect(senderSigner)
      .callCodeSetAndLoad(await transientStorageReader.getAddress(), 6n, 654n)
    await tx.wait()

    expect(await transientStorageCheck.lastLoaded()).to.equal(654n)
    expect(await transientStorageReader.lastLoaded()).to.equal(0n)
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
