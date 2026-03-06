import { expect } from "chai"
import hre from "hardhat"
import { getDeployedContract } from "./helpers/contract"

describe("RandaoCheck", function () {
  const { deployments } = hre
  let randaoCheck: any

  before(async function () {
    await deployments.fixture()
    randaoCheck = await getDeployedContract("RandaoCheck")
  })

  it("should return zero for difficulty and prevrandao", async function () {
    const [difficulty, prevrandao] = await randaoCheck.values()

    expect(difficulty).to.equal(0n)
    expect(prevrandao).to.equal(0n)
  })
})
