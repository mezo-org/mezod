import { expect } from "chai"
import hre from "hardhat"
import type { RandaoCheck } from "../typechain-types"
import { getDeployedContract } from "./helpers/contract"

describe("RandaoCheck", function () {
  const { deployments } = hre
  let randaoCheck: RandaoCheck

  before(async function () {
    await deployments.fixture(["RandaoCheck"])
    randaoCheck = await getDeployedContract("RandaoCheck")
  })

  it("should return zero for difficulty and prevrandao", async function () {
    const [difficulty, prevrandao] = await randaoCheck.values()

    expect(difficulty).to.equal(0n)
    expect(prevrandao).to.equal(0n)
  })
})
