import { expect } from "chai"
import { ethers } from "hardhat"
import {
  CapturedError,
  extractCode,
  extractMessage,
} from "./helpers/rpc-error"

// SimulateV1_Validation covers the spec's `validation: true` mode:
// tx-level validation failures must abort the entire eth_simulateV1
// request as a top-level structured error, NOT land as per-call entries.
//
// validation=false (the default) bypasses the gate entirely, so the
// negative twins below assert that nonce-too-low / insufficient-funds /
// fee-cap-below-base do not promote to a top-level fatal when the flag
// is omitted or set to false.
//
// Per the create-tests Solidity conventions: no beforeEach, all RPC
// requests fire in `before`, flat describe -> context -> it shape.
describe("SimulateV1_Validation", function () {
  // Spec-reserved JSON-RPC error codes the gate emits.
  const NONCE_TOO_LOW = -38010
  const NONCE_TOO_HIGH = -38011
  const BASE_FEE_TOO_LOW = -38012
  const INTRINSIC_GAS_TOO_LOW = -38013
  const INSUFFICIENT_FUNDS = -38014
  const INIT_CODE_TOO_LARGE = -38025
  const FEE_CAP_TOO_LOW = -32005
  const REVERTED = 3

  // 1e12 wei is large enough to clear any plausible eip1559 baseFee
  // floor the chain might land on for a single block; this keeps the
  // fee-cap gate from masking whatever gate each test means to
  // exercise.
  const HIGH_FEE = "0xe8d4a51000" // 1e12

  // 1e22 wei (~10_000 BTC). Large enough to cover gasLimit*HIGH_FEE +
  // small `value` for every fixture below. Specific tests that need an
  // under-funded sender override balance to "0x0".
  const FUNDED_BALANCE = "0x21e19e0c9bab2400000" // 1e22

  // ERC-20-style sender + recipient. Detached from the localnode's
  // dev keys so the simulate request never collides with real on-chain
  // state.
  const SENDER = "0xc100000000000000000000000000000000000001"
  const RECIPIENT = "0xc200000000000000000000000000000000000002"

  // simulate runs eth_simulateV1 with the supplied opts and returns a
  // CapturedError indicating whether the request itself threw (which
  // is how validation=true fatals surface) or returned a result envelope
  // (which is how every per-call outcome surfaces).
  async function simulate(opts: any): Promise<{
    error: CapturedError
    result: any[] | undefined
  }> {
    try {
      const result: any[] = await ethers.provider.send("eth_simulateV1", [
        opts,
        "latest",
      ])
      return {
        error: { thrown: false, code: undefined, message: "" },
        result,
      }
    } catch (err: any) {
      return {
        error: {
          thrown: true,
          code: extractCode(err),
          message: extractMessage(err),
        },
        result: undefined,
      }
    }
  }

  const fundedState = {
    [SENDER]: { balance: FUNDED_BALANCE },
  }

  // ---- happy + fatal cases (validation=true) -----------------------------

  let happyValidationResult: { error: CapturedError; result: any[] | undefined }
  let nonceLowResult: { error: CapturedError; result: any[] | undefined }
  let nonceHighResult: { error: CapturedError; result: any[] | undefined }
  let insufficientFundsResult: { error: CapturedError; result: any[] | undefined }
  let intrinsicGasResult: { error: CapturedError; result: any[] | undefined }
  let initCodeTooLargeResult: { error: CapturedError; result: any[] | undefined }
  let feeCapTooLowResult: { error: CapturedError; result: any[] | undefined }
  let baseFeeOverrideTooLowResult: {
    error: CapturedError
    result: any[] | undefined
  }
  let revertUnderValidationResult: {
    error: CapturedError
    result: any[] | undefined
  }

  // ---- validation=false (negative twins) ---------------------------------

  let nonceLowNoValidationResult: {
    error: CapturedError
    result: any[] | undefined
  }
  let insufficientFundsNoValidationResult: {
    error: CapturedError
    result: any[] | undefined
  }
  let feeCapBelowBaseNoValidationResult: {
    error: CapturedError
    result: any[] | undefined
  }

  before(async function () {
    // Happy path: every gate clears.
    happyValidationResult = await simulate({
      blockStateCalls: [
        {
          stateOverrides: fundedState,
          calls: [
            {
              from: SENDER,
              to: RECIPIENT,
              value: "0x1",
              maxFeePerGas: HIGH_FEE,
            },
          ],
        },
      ],
      validation: true,
    })

    // -38010: state nonce 5, call.nonce 4.
    nonceLowResult = await simulate({
      blockStateCalls: [
        {
          stateOverrides: {
            [SENDER]: { balance: FUNDED_BALANCE, nonce: "0x5" },
          },
          calls: [
            {
              from: SENDER,
              to: RECIPIENT,
              nonce: "0x4",
              maxFeePerGas: HIGH_FEE,
            },
          ],
        },
      ],
      validation: true,
    })

    // -38011: state nonce 5, call.nonce 9.
    nonceHighResult = await simulate({
      blockStateCalls: [
        {
          stateOverrides: {
            [SENDER]: { balance: FUNDED_BALANCE, nonce: "0x5" },
          },
          calls: [
            {
              from: SENDER,
              to: RECIPIENT,
              nonce: "0x9",
              maxFeePerGas: HIGH_FEE,
            },
          ],
        },
      ],
      validation: true,
    })

    // -38014: balance zero, value 1.
    insufficientFundsResult = await simulate({
      blockStateCalls: [
        {
          stateOverrides: {
            [SENDER]: { balance: "0x0" },
          },
          calls: [
            {
              from: SENDER,
              to: RECIPIENT,
              value: "0x1",
              maxFeePerGas: HIGH_FEE,
            },
          ],
        },
      ],
      validation: true,
    })

    // -38013: gas below the 21k pure-transfer intrinsic floor.
    intrinsicGasResult = await simulate({
      blockStateCalls: [
        {
          stateOverrides: fundedState,
          calls: [
            {
              from: SENDER,
              to: RECIPIENT,
              value: "0x1",
              gas: "0x5207", // 20999
              maxFeePerGas: HIGH_FEE,
            },
          ],
        },
      ],
      validation: true,
    })

    // -38025: CREATE with init-code MaxInitCodeSize (49152) + 1.
    const oversizedInitCode = "0x" + "00".repeat(49153)
    initCodeTooLargeResult = await simulate({
      blockStateCalls: [
        {
          stateOverrides: fundedState,
          calls: [
            {
              from: SENDER,
              data: oversizedInitCode,
              gas: "0xf4240", // 1_000_000 — covers intrinsic + EIP-3860 word cost
              maxFeePerGas: HIGH_FEE,
            },
          ],
        },
      ],
      validation: true,
    })

    // -32005: explicit maxFeePerGas=0 with a non-zero base-fee override
    // forces the per-call fee-cap gate to fire.
    feeCapTooLowResult = await simulate({
      blockStateCalls: [
        {
          blockOverrides: { baseFeePerGas: "0x3b9aca00" }, // 1 gwei
          stateOverrides: fundedState,
          calls: [
            {
              from: SENDER,
              to: RECIPIENT,
              value: "0x1",
              maxFeePerGas: "0x0",
              maxPriorityFeePerGas: "0x0",
            },
          ],
        },
      ],
      validation: true,
    })

    // -38012: blockOverrides.baseFeePerGas (1 wei) below the chain
    // eip1559 floor.
    baseFeeOverrideTooLowResult = await simulate({
      blockStateCalls: [
        {
          blockOverrides: { baseFeePerGas: "0x1" }, // 1 wei
          stateOverrides: fundedState,
          calls: [
            {
              from: SENDER,
              to: RECIPIENT,
              value: "0x1",
              maxFeePerGas: HIGH_FEE,
            },
          ],
        },
      ],
      validation: true,
    })

    // Reverting call under validation=true: per-call code 3, NOT a
    // request-level fatal.
    //
    // Bytecode: PUSH1 0x00 PUSH1 0x00 REVERT — unconditional revert
    // with empty return data.
    const revertContract = "0xddee000000000000000000000000000000000000"
    revertUnderValidationResult = await simulate({
      blockStateCalls: [
        {
          stateOverrides: {
            [SENDER]: { balance: FUNDED_BALANCE },
            [revertContract]: { code: "0x60006000fd" },
          },
          calls: [
            {
              from: SENDER,
              to: revertContract,
              maxFeePerGas: HIGH_FEE,
            },
          ],
        },
      ],
      validation: true,
    })

    // ---- validation=false negative twins ---------------------------------

    // Same opts as nonceLowResult, but validation flag dropped.
    nonceLowNoValidationResult = await simulate({
      blockStateCalls: [
        {
          stateOverrides: {
            [SENDER]: { balance: FUNDED_BALANCE, nonce: "0x5" },
          },
          calls: [
            { from: SENDER, to: RECIPIENT, nonce: "0x4" },
          ],
        },
      ],
      // validation omitted (default false)
    })

    // Same opts as insufficientFundsResult, validation flag dropped.
    // Mezod diverges from geth here: the EVM's CanTransfer guard still
    // rejects the value transfer per-call, so status is 0x0 — but the
    // request itself does NOT abort with a top-level -38014.
    insufficientFundsNoValidationResult = await simulate({
      blockStateCalls: [
        {
          stateOverrides: {
            [SENDER]: { balance: "0x0" },
          },
          calls: [
            { from: SENDER, to: RECIPIENT, value: "0x1" },
          ],
        },
      ],
    })

    // Same opts as feeCapTooLowResult, validation flag dropped. With
    // validation=false the synthesized header carries BaseFee=0
    // regardless of the override, so the per-call fee-cap arithmetic
    // always passes.
    feeCapBelowBaseNoValidationResult = await simulate({
      blockStateCalls: [
        {
          blockOverrides: { baseFeePerGas: "0x3b9aca00" },
          stateOverrides: fundedState,
          calls: [
            {
              from: SENDER,
              to: RECIPIENT,
              value: "0x1",
              maxFeePerGas: "0x0",
              maxPriorityFeePerGas: "0x0",
            },
          ],
        },
      ],
    })
  })

  context("happy path (validation=true)", function () {
    it("returns a successful per-call result without a top-level error", function () {
      expect(happyValidationResult.error.thrown).to.equal(false)
      expect(happyValidationResult.result).to.exist
      const blocks = happyValidationResult.result!
      expect(blocks).to.have.lengthOf(1)
      expect(blocks[0].calls).to.have.lengthOf(1)
      expect(blocks[0].calls[0].status).to.equal("0x1")
    })
  })

  context("fatal: nonce too low (-38010)", function () {
    it("aborts the request at the JSON-RPC layer", function () {
      expect(nonceLowResult.error.thrown).to.equal(true)
      expect(nonceLowResult.error.code).to.equal(NONCE_TOO_LOW)
    })
  })

  context("fatal: nonce too high (-38011)", function () {
    it("aborts the request at the JSON-RPC layer", function () {
      expect(nonceHighResult.error.thrown).to.equal(true)
      expect(nonceHighResult.error.code).to.equal(NONCE_TOO_HIGH)
    })
  })

  context("fatal: insufficient funds (-38014)", function () {
    it("aborts the request at the JSON-RPC layer", function () {
      expect(insufficientFundsResult.error.thrown).to.equal(true)
      expect(insufficientFundsResult.error.code).to.equal(INSUFFICIENT_FUNDS)
    })
  })

  context("fatal: intrinsic gas too low (-38013)", function () {
    it("aborts the request at the JSON-RPC layer", function () {
      expect(intrinsicGasResult.error.thrown).to.equal(true)
      expect(intrinsicGasResult.error.code).to.equal(INTRINSIC_GAS_TOO_LOW)
    })
  })

  context("fatal: init-code too large (-38025)", function () {
    it("aborts the request at the JSON-RPC layer", function () {
      expect(initCodeTooLargeResult.error.thrown).to.equal(true)
      expect(initCodeTooLargeResult.error.code).to.equal(INIT_CODE_TOO_LARGE)
    })
  })

  context("fatal: fee-cap below base fee (-32005)", function () {
    it("aborts the request at the JSON-RPC layer", function () {
      expect(feeCapTooLowResult.error.thrown).to.equal(true)
      expect(feeCapTooLowResult.error.code).to.equal(FEE_CAP_TOO_LOW)
    })
  })

  context("fatal: base-fee override below floor (-38012)", function () {
    it("aborts the request at the JSON-RPC layer", function () {
      expect(baseFeeOverrideTooLowResult.error.thrown).to.equal(true)
      expect(baseFeeOverrideTooLowResult.error.code).to.equal(BASE_FEE_TOO_LOW)
    })
  })

  context("revert is per-call under validation=true", function () {
    it("does NOT abort the request", function () {
      expect(revertUnderValidationResult.error.thrown).to.equal(false)
    })

    it("surfaces the revert as a per-call code 3 with status 0x0", function () {
      expect(revertUnderValidationResult.result).to.exist
      const blocks = revertUnderValidationResult.result!
      expect(blocks).to.have.lengthOf(1)
      expect(blocks[0].calls).to.have.lengthOf(1)
      const call = blocks[0].calls[0]
      expect(call.status).to.equal("0x0")
      expect(call.error).to.exist
      expect(call.error.code).to.equal(REVERTED)
    })
  })

  // -----------------------------------------------------------------------
  // validation=false negative twins: every gate is OFF.
  // -----------------------------------------------------------------------

  context("nonce-too-low succeeds when validation is omitted", function () {
    it("does NOT abort the request", function () {
      expect(nonceLowNoValidationResult.error.thrown).to.equal(false)
    })

    it("returns the per-call result with status 0x1", function () {
      const blocks = nonceLowNoValidationResult.result!
      expect(blocks).to.have.lengthOf(1)
      expect(blocks[0].calls[0].status).to.equal("0x1")
    })
  })

  context("insufficient-funds is per-call (NOT fatal) when validation is omitted", function () {
    it("does NOT abort the request with a top-level -38014", function () {
      expect(insufficientFundsNoValidationResult.error.thrown).to.equal(false)
    })

    it("the call still appears in the per-call list", function () {
      // Mezod's EVM CanTransfer guard rejects the value transfer
      // per-call (status 0x0), but the request itself does NOT abort.
      // The pinned invariant is "call appears in per-call list", not a
      // specific status — geth diverges here, so we assert structure
      // rather than the per-call code.
      const blocks = insufficientFundsNoValidationResult.result!
      expect(blocks).to.have.lengthOf(1)
      expect(blocks[0].calls).to.have.lengthOf(1)
    })
  })

  context("fee-cap-below-base succeeds when validation is omitted", function () {
    it("does NOT abort the request", function () {
      expect(feeCapBelowBaseNoValidationResult.error.thrown).to.equal(false)
    })

    it("returns the per-call result with status 0x1", function () {
      const blocks = feeCapBelowBaseNoValidationResult.result!
      expect(blocks).to.have.lengthOf(1)
      expect(blocks[0].calls[0].status).to.equal("0x1")
    })
  })
})
