import { expect } from "chai"
import hre from "hardhat"
import type { Push0Check } from "../typechain-types"
import { getDeployedContract } from "./helpers/contract"
import { getContractOpcodes } from "./helpers/opcodes"

describe("Push0Check", function () {
  const { deployments } = hre
  let push0Check: Push0Check

  before(async function () {
    await deployments.fixture(["Push0Check"])
    push0Check = await getDeployedContract("Push0Check")
  })

  it("should deploy and return zero", async function () {
    const value = await push0Check.zero()
    expect(value).to.equal(0n)
  })

  it("should compile with PUSH0 in opcode listing", async function () {
    const opcodes = await getContractOpcodes("Push0Check")
    expect(opcodes).to.include("PUSH0")
  })
})
