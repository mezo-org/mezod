import { expect } from "chai"
import hre from "hardhat"
import type { McopyCheck } from "../typechain-types"
import { getDeployedContract } from "./helpers/contract"
import { getContractOpcodes } from "./helpers/opcodes"

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

  it("should copy empty bytes with MCOPY", async function () {
    const output = await mcopyCheck.copy.staticCall("0x")
    expect(output).to.equal("0x")
  })

  it("should handle overlapping copy (destination after source)", async function () {
    const output = await mcopyCheck.overlapCopyForward.staticCall()
    expect(output).to.equal("0x0102010203040506")
  })

  it("should handle overlapping copy (destination before source)", async function () {
    const output = await mcopyCheck.overlapCopyBackward.staticCall()
    expect(output).to.equal("0x0304050607080708")
  })

  it("should treat zero-length overlapping copy as a no-op", async function () {
    const output = await mcopyCheck.zeroLengthOverlapCopy.staticCall()
    expect(output).to.equal("0x0102030405060708")
  })

  it("should compile with MCOPY in opcode listing", async function () {
    const opcodes = await getContractOpcodes("McopyCheck")
    expect(opcodes).to.include("MCOPY")
  })
})
