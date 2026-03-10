import { expect } from "chai"
import hre from "hardhat"
import type { McopyCheck } from "../typechain-types"
import { getDeployedContract } from "./helpers/contract"
import { getDeployedOpcodes } from "./helpers/opcodes"

describe("McopyCheck", function () {
  const { deployments } = hre
  let mcopyCheck: McopyCheck

  before(async function () {
    await deployments.fixture(["McopyCheck"])
    mcopyCheck = await getDeployedContract("McopyCheck")
  })

  it("should copy bytes with MCOPY", async function () {
    const input = "0x11223344556677889900aabbccddeeff"
    const output = await mcopyCheck.copy.staticCall(input)

    expect(output).to.equal(input)
  })

  it("should handle overlapping copy (destination after source)", async function () {
    const output = await mcopyCheck.overlapCopyForward.staticCall()
    expect(output).to.equal("0x0102010203040506")
  })

  it("should handle overlapping copy (destination before source)", async function () {
    const output = await mcopyCheck.overlapCopyBackward.staticCall()
    expect(output).to.equal("0x0304050607080708")
  })

  it("should compile with MCOPY in opcode listing", async function () {
    const opcodes = await getDeployedOpcodes("McopyCheck")
    expect(opcodes).to.include("MCOPY")
  })
})
