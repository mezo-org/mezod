import { expect } from "chai"
import { ethers } from "hardhat"

describe("SimulateV1_RejectedOverrides", function () {
  type Captured = {
    thrown: boolean
    code: number | undefined
    message: string
  }

  async function simulateWithBlockOverrides(overrides: any): Promise<Captured> {
    const opts = {
      blockStateCalls: [
        { blockOverrides: overrides, calls: [] },
      ],
    }
    try {
      await ethers.provider.send("eth_simulateV1", [opts, "latest"])
      return { thrown: false, code: undefined, message: "" }
    } catch (err: any) {
      return {
        thrown: true,
        code: extractCode(err),
        message: extractMessage(err),
      }
    }
  }

  let beaconRootResult: Captured
  let withdrawalsResult: Captured
  let blobBaseFeeResult: Captured

  before(async function () {
    beaconRootResult = await simulateWithBlockOverrides({ beaconRoot: ethers.ZeroHash })
    withdrawalsResult = await simulateWithBlockOverrides({
      withdrawals: [
        {
          index: "0x0",
          validatorIndex: "0x0",
          address: ethers.ZeroAddress,
          amount: "0x1",
        },
      ],
    })
    blobBaseFeeResult = await simulateWithBlockOverrides({ blobBaseFee: "0x1" })
  })

  context("BlockOverrides.beaconRoot", function () {
    it("is rejected with -32602 and a BeaconRoot-specific message", function () {
      expect(beaconRootResult.thrown).to.equal(true)
      expect(beaconRootResult.code).to.equal(-32602)
      expect(beaconRootResult.message).to.include(
        "BlockOverrides.BeaconRoot is not supported",
      )
    })
  })

  context("BlockOverrides.withdrawals", function () {
    it("is rejected with -32602 and a Withdrawals-specific message", function () {
      expect(withdrawalsResult.thrown).to.equal(true)
      expect(withdrawalsResult.code).to.equal(-32602)
      expect(withdrawalsResult.message).to.include(
        "BlockOverrides.Withdrawals is not supported",
      )
    })
  })

  context("BlockOverrides.blobBaseFee", function () {
    it("is rejected with -32602 and a BlobBaseFee-specific message", function () {
      expect(blobBaseFeeResult.thrown).to.equal(true)
      expect(blobBaseFeeResult.code).to.equal(-32602)
      expect(blobBaseFeeResult.message).to.include(
        "BlockOverrides.BlobBaseFee is not supported",
      )
    })
  })
})

// Providers wrap JSON-RPC errors through several layers; drill through to find
// the original {code, message} object regardless of where it landed.
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
