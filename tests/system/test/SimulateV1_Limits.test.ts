import { expect } from "chai"
import { ethers } from "hardhat"
import {
  CapturedError,
  extractCode,
  extractMessage,
} from "./helpers/rpc-error"

// SimulateV1_Limits exercises the structured top-level errors that
// abort an eth_simulateV1 request: the request-envelope DoS caps
// (-38026) and the gas-error semantics that execute.yaml routes through
// the request channel rather than per-call entries (-38015 block gas
// limit, -38013 intrinsic gas). Each case confirms the structured
// SimError survives keeper -> gRPC -> JSON-RPC -> ethers with code and
// message intact. The DoS guards that need a node restart (kill switch)
// or runtime knob tuning the harness can't orchestrate (timeout,
// gas-pool clamp) stay covered by Go unit tests.
describe("SimulateV1_Limits", function () {
  async function simulate(opts: any): Promise<CapturedError> {
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

  let blockCapResult: CapturedError
  let callCapResult: CapturedError
  let blockGasLimitResult: CapturedError
  let intrinsicGasResult: CapturedError

  before(async function () {
    const [deployer] = await ethers.getSigners()
    const senderAddr = deployer.address
    const recipientAddr = ethers.Wallet.createRandom().address

    const tooManyBlocks = Array.from({ length: 257 }, () => ({ calls: [] }))
    blockCapResult = await simulate({ blockStateCalls: tooManyBlocks })

    const tooManyCalls = Array.from({ length: 1001 }, () => ({
      from: ethers.ZeroAddress,
      to: ethers.ZeroAddress,
    }))
    callCapResult = await simulate({
      blockStateCalls: [{ calls: tooManyCalls }],
    })

    // First call burns ~21000 of the 50000-gas block budget; second
    // call asks for 30000 against ~29000 remaining. resolveSimCallGas
    // returns NewSimBlockGasLimitReached, which the driver surfaces as
    // a top-level fatal (CallResultFailure permits only codes 3 and
    // -32015 per execute.yaml, so -38015 cannot ride per-call).
    blockGasLimitResult = await simulate({
      blockStateCalls: [
        {
          blockOverrides: { gasLimit: ethers.toQuantity(50_000n) },
          stateOverrides: {
            [senderAddr]: {
              balance: ethers.toQuantity(ethers.parseEther("100")),
            },
          },
          calls: [
            {
              from: senderAddr,
              to: recipientAddr,
              value: ethers.toQuantity(1n),
            },
            {
              from: senderAddr,
              to: recipientAddr,
              value: ethers.toQuantity(1n),
              gas: ethers.toQuantity(30_000n),
            },
          ],
        },
      ],
    })

    // Explicit gas of 1000 is well below the 21000 intrinsic baseline.
    // applyMessageWithConfig returns core.ErrIntrinsicGas; the driver
    // wraps it in NewSimIntrinsicGas and surfaces it as a top-level
    // fatal for the same execute.yaml reason as -38015.
    intrinsicGasResult = await simulate({
      blockStateCalls: [
        {
          stateOverrides: {
            [senderAddr]: {
              balance: ethers.toQuantity(ethers.parseEther("100")),
            },
          },
          calls: [
            {
              from: senderAddr,
              to: recipientAddr,
              value: ethers.toQuantity(1n),
              gas: ethers.toQuantity(1_000n),
            },
          ],
        },
      ],
    })
  })

  context("block cap", function () {
    it("rejects 257-block requests with -38026", function () {
      expect(blockCapResult.thrown).to.equal(true)
      expect(blockCapResult.code).to.equal(-38026)
      expect(blockCapResult.message).to.include("blocks > max 256")
    })
  })

  context("call cap", function () {
    it("rejects 1001-call requests with -38026", function () {
      expect(callCapResult.thrown).to.equal(true)
      expect(callCapResult.code).to.equal(-38026)
      expect(callCapResult.message).to.include("calls > max 1000")
    })
  })

  context("block gas limit", function () {
    it("rejects an over-budget call with top-level -38015", function () {
      expect(blockGasLimitResult.thrown).to.equal(true)
      expect(blockGasLimitResult.code).to.equal(-38015)
    })
  })

  context("intrinsic gas", function () {
    it("rejects a below-intrinsic call with top-level -38013", function () {
      expect(intrinsicGasResult.thrown).to.equal(true)
      expect(intrinsicGasResult.code).to.equal(-38013)
    })
  })
})
