import { expect } from "chai"
import { ethers } from "hardhat"

describe("SimulateV1_Stub", function () {
  // Phase 1 ships a stub that returns JSON-RPC -32603 (internal error)
  // with a "not yet implemented" message. The method IS registered on
  // the eth_ namespace, so -32603 is the spec-listed internal-error
  // code (per execution-apis `src/eth/execute.yaml`), not -32601 which
  // would incorrectly signal an unknown method.
  // Later phases replace this stub with the full implementation.

  it("should be registered as eth_simulateV1 and return not-yet-implemented", async function () {
    let raisedError: any
    try {
      await ethers.provider.send("eth_simulateV1", [
        { blockStateCalls: [] },
        "latest",
      ])
      expect.fail("eth_simulateV1 should have returned an error in Phase 1")
    } catch (err: any) {
      raisedError = err
    }

    // Providers wrap JSON-RPC errors in various shapes. Drill through to find
    // the original `{code, message}` object.
    const code = extractCode(raisedError)
    const message = extractMessage(raisedError)

    expect(code, `expected numeric JSON-RPC code, got ${JSON.stringify(raisedError)}`).to.equal(-32603)
    expect(message).to.be.a("string")
    expect(message.toLowerCase()).to.include("eth_simulatev1")
    expect(message.toLowerCase()).to.include("not yet implemented")
  })
})

function extractCode(err: any): number | undefined {
  if (!err) return undefined
  if (typeof err.code === "number") return err.code
  if (err.error && typeof err.error.code === "number") return err.error.code
  if (err.info && err.info.error && typeof err.info.error.code === "number") return err.info.error.code
  if (err.data && typeof err.data.code === "number") return err.data.code
  return undefined
}

function extractMessage(err: any): string {
  if (!err) return ""
  if (err.error && typeof err.error.message === "string") return err.error.message
  if (err.info && err.info.error && typeof err.info.error.message === "string") return err.info.error.message
  if (typeof err.message === "string") return err.message
  return String(err)
}
