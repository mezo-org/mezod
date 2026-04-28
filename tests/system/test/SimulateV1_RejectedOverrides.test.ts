import { expect } from "chai"
import { ethers } from "hardhat"
import {
  CapturedError,
  extractCode,
  extractMessage,
} from "./helpers/rpc-error"

describe("SimulateV1_RejectedOverrides", function () {
  async function simulateWithBlockOverrides(overrides: any): Promise<CapturedError> {
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

  let beaconRootResult: CapturedError
  let withdrawalsResult: CapturedError
  let blobBaseFeeResult: CapturedError

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
