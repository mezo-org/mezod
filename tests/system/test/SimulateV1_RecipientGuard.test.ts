import { expect } from "chai"
import hre, { ethers } from "hardhat"
import { getDeployedContract } from "./helpers/contract"
import {
  CapturedError,
  extractCode,
  extractMessage,
} from "./helpers/rpc-error"

/**
 * SimulateV1_RecipientGuard — pins recipient-guard behavior under
 * eth_simulateV1. The keeper's CanReceiveTransfer predicate is wired
 * into the simulated BlockContext (x/evm/keeper/simulate_v1.go), so any
 * regression that drops it (or quietly relaxes the rejection to a
 * top-level fatal) would let blocked-address credits slip past
 * simulate. The companion live-tx coverage lives in RecipientGuard.test.ts;
 * the spec-conformant / divergence halves live in
 * SimulateV1_SpecCompliance.test.ts and SimulateV1_MezoDivergence.test.ts
 * respectively — no assertion is duplicated across them.
 *
 * Test scenarios:
 *
 * | Scenario                                       | Given                                              | When                                                 | Then                                                                 |
 * |------------------------------------------------|----------------------------------------------------|------------------------------------------------------|----------------------------------------------------------------------|
 * | inner CALL via Forwarder, simulated            | Forwarder deployed; recipient is poa module addr   | eth_simulateV1 dispatches Forwarder.run(blocked)     | top-level success; outer call status=0x1; decoded (bool ok) = false  |
 * | top-level value transfer to blocked recipient  | funded sender (stateOverrides); validation omitted | eth_simulateV1 dispatches {to: blocked, value: 0x1}  | top-level success; per-call status=0x0; error.code=-32015            |
 */
describe("SimulateV1_RecipientGuard", function () {
  const { deployments } = hre

  // EVM-format address of a Cosmos module account: sha256(name)[:20].
  // Duplicated from RecipientGuard.test.ts to keep the suite
  // self-contained (no shared cosmos-module helper exists yet).
  function blockedModuleAddress(name: string): string {
    const digest = ethers.sha256(ethers.toUtf8Bytes(name))
    return ethers.getAddress("0x" + digest.slice(2, 42))
  }

  // poa is a stable module account in BankKeeper.BlockedAddrs.
  // fee_collector is also blocked but receives gas fees via
  // SendCoinsFromAccountToModule, so its balance grows during normal
  // block processing — pinning to poa keeps the assertion deterministic.
  const blockedRecipient = blockedModuleAddress("poa")

  // Detached EOA used by the validation=false top-level case. Kept off
  // the dev keys so the simulate request never collides with real
  // on-chain state. The stateOverrides funds it with enough abtc for
  // the value transfer to clear the per-call CanTransfer check; only
  // CanReceiveTransfer should reject.
  const FUNDED_SENDER = "0xc100000000000000000000000000000000000003"

  // Per-call code emitted when CanReceiveTransfer rejects a value
  // transfer. NewSimVMError emits -32015 (SimErrCodeVMError) for
  // non-revert VM errors (x/evm/types/simulate_v1_errors.go:25); the
  // guard surfaces as a non-revert VM failure rather than as the
  // ExecutionReverted path (BuildSimCallResult in
  // x/evm/types/simulate_v1.go:241), so the per-call code is -32015.
  const SIM_VM_ERROR = -32015

  let forwarder: any
  let forwarderAddress: string
  let signerAddress: string

  before(async function () {
    await deployments.fixture(["RecipientGuard"])
    forwarder = await getDeployedContract("Forwarder")
    forwarderAddress = await forwarder.getAddress()
    const [deployer] = await ethers.getSigners()
    signerAddress = deployer.address
  })

  context("inner CALL via Forwarder is simulated successfully but ok=false", function () {
    // Forwarder.run(to) performs `(ok, ) = to.call{value: msg.value}("")`
    // and returns ok. With a blocked recipient the inner CALL fails
    // (CanReceiveTransfer rejects), but the outer frame still returns
    // cleanly with ok=false. Decoding returnData as (bool ok) is the
    // unambiguous pin that the inner CALL was the failing site — outer
    // status alone wouldn't distinguish a successful forward from a
    // forward whose inner CALL failed.
    let result: any[]

    before(async function () {
      const runData = forwarder.interface.encodeFunctionData("run", [
        blockedRecipient,
      ])

      result = await ethers.provider.send("eth_simulateV1", [
        {
          blockStateCalls: [
            {
              calls: [
                {
                  from: signerAddress,
                  to: forwarderAddress,
                  data: runData,
                  value: "0x1",
                  gas: "0x30d40", // 200_000
                },
              ],
            },
          ],
        },
        "latest",
      ])
    })

    it("returns blocks without a top-level error", function () {
      expect(result).to.have.lengthOf(1)
      expect(result[0].calls).to.have.lengthOf(1)
    })

    it("reports the outer call as status 0x1", function () {
      expect(result[0].calls[0].status).to.equal(
        "0x1",
        "Forwarder.run outer frame must return cleanly",
      )
    })

    it("decodes returnData as (bool ok) = false to pin the inner CALL as the failing site", function () {
      const [ok] = ethers.AbiCoder.defaultAbiCoder().decode(
        ["bool"],
        result[0].calls[0].returnData,
      )
      expect(ok).to.equal(
        false,
        "inner CALL to blocked recipient must surface as ok=false",
      )
    })
  })

  context("top-level value transfer to blocked recipient is per-call when validation is omitted", function () {
    // validation=false skips the request-level base-fee / balance gates
    // but CanReceiveTransfer stays live (it's wired into the simulated
    // BlockContext in x/evm/keeper/simulate_v1.go). The request itself
    // must not abort; the rejection lands on the per-call result with
    // status=0x0 and the VM-error code (-32015) — the same shape as the
    // CanTransfer divergence pinned in SimulateV1_MezoDivergence.
    let captured: { error: CapturedError; result: any[] | undefined }

    before(async function () {
      try {
        const result: any[] = await ethers.provider.send("eth_simulateV1", [
          {
            blockStateCalls: [
              {
                stateOverrides: {
                  [FUNDED_SENDER]: {
                    balance: ethers.toQuantity(ethers.parseEther("1")),
                  },
                },
                calls: [
                  {
                    from: FUNDED_SENDER,
                    to: blockedRecipient,
                    value: "0x1",
                  },
                ],
              },
            ],
            // validation omitted (defaults to false)
          },
          "latest",
        ])
        captured = {
          error: { thrown: false, code: undefined, message: "" },
          result,
        }
      } catch (err: any) {
        captured = {
          error: {
            thrown: true,
            code: extractCode(err),
            message: extractMessage(err),
          },
          result: undefined,
        }
      }
    })

    it("does NOT abort the request with a top-level fatal", function () {
      expect(captured.error.thrown).to.equal(false)
      expect(captured.result).to.exist
    })

    it("surfaces the rejection as a per-call status 0x0 with the VM-error code", function () {
      const blocks = captured.result!
      expect(blocks).to.have.lengthOf(1)
      expect(blocks[0].calls).to.have.lengthOf(1)
      const call = blocks[0].calls[0]
      expect(call.status).to.equal("0x0")
      expect(call.error).to.exist
      expect(call.error.code).to.equal(SIM_VM_ERROR)
    })
  })
})
