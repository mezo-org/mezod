import { expect } from "chai"
import { ethers } from "hardhat"
import { extractMessage } from "./helpers/rpc-error"

describe("SimulateV1_MovePrecompile_ethCall", function () {
  const SHA256_ADDR = "0x0000000000000000000000000000000000000002"
  const DEST_ADDR = "0x0000000000000000000000000000000000001234"
  const BTC_CUSTOM_PRECOMPILE = "0x7b7c000000000000000000000000000000000000"

  // sha256("hello") — precomputed, stable.
  const HELLO = "0x68656c6c6f" // "hello" as bytes
  const EXPECTED =
    "0x2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"

  it("moves sha256 to a destination address and returns the hash when called there", async function () {
    const result: string = await ethers.provider.send("eth_call", [
      { to: DEST_ADDR, data: HELLO },
      "latest",
      {
        [SHA256_ADDR]: { movePrecompileToAddress: DEST_ADDR },
      },
    ])

    expect(result.toLowerCase()).to.equal(EXPECTED)
  })

  it("rejects moving a mezo custom precompile (BTC token) with invalid params", async function () {
    let raisedError: any
    try {
      await ethers.provider.send("eth_call", [
        { to: DEST_ADDR, data: "0x" },
        "latest",
        {
          [BTC_CUSTOM_PRECOMPILE]: { movePrecompileToAddress: DEST_ADDR },
        },
      ])
      expect.fail("eth_call should have rejected the mezo custom precompile move")
    } catch (err: any) {
      raisedError = err
    }

    expect(raisedError, "eth_call should have raised").to.exist
    const message = extractMessage(raisedError).toLowerCase()
    expect(message).to.include("cannot move mezo custom precompile")
  })

  it("rejects a self-referencing move (source == destination)", async function () {
    let raisedError: any
    try {
      await ethers.provider.send("eth_call", [
        { to: SHA256_ADDR, data: HELLO },
        "latest",
        {
          [SHA256_ADDR]: { movePrecompileToAddress: SHA256_ADDR },
        },
      ])
      expect.fail("eth_call should have rejected a self-referencing move")
    } catch (err: any) {
      raisedError = err
    }

    expect(raisedError, "eth_call should have raised").to.exist
    const message = extractMessage(raisedError).toLowerCase()
    expect(message).to.include("referenced itself")
  })
})
