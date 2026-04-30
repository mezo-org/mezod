import { expect } from "chai"
import hre, { ethers } from "hardhat"
import { SimpleToken } from "../typechain-types/SimpleToken"
import { getDeployedContract } from "./helpers/contract"
import {
  CapturedError,
  extractCode,
  extractMessage,
} from "./helpers/rpc-error"
import btcabi from "../../../precompile/btctoken/abi.json"

/**
 * SimulateV1_MezoDivergence — pins mezod's intentional deviations from
 * the spec, so any drift back toward it surfaces as a test failure.
 *
 * Each `it()` here pins a behavior where mezod intentionally deviates
 * from `execution-apis` and geth's reference `eth_simulateV1`. If one of
 * these starts failing, mezod has drifted toward the spec; investigate
 * the divergence before flipping the assertion to make it green again.
 *
 * The companion suite `SimulateV1_SpecCompliance.test.ts` pins the
 * spec-conformant surface; the two files are mutually exclusive — no
 * assertion is duplicated across them.
 *
 * The kill switch (`JSONRPCConfig.SimulateDisabled`, TOML key
 * `simulate-disabled`) is covered Go-side by
 * `TestSimulateV1_KillSwitch{,Off}` in
 * `rpc/namespaces/ethereum/eth/simulate_v1_test.go`. A TS twin would
 * require a node restart with the flag flipped, which the harness can't
 * orchestrate, so the case is intentionally skipped here.
 *
 * Test scenarios:
 *
 * | Scenario                            | Given                                | When                                              | Then                                                                    |
 * |-------------------------------------|--------------------------------------|---------------------------------------------------|-------------------------------------------------------------------------|
 * | btctoken state chains across calls  | dev0 pre-funded with native BTC      | btctoken.transfer + .balanceOf in one block       | both succeed; balanceOf returns the transferred amount                  |
 * | btctoken state chains across blocks | dev0 pre-funded with native BTC      | btctoken.transfer in block 1; .balanceOf in 2     | block 2 reads block 1's amount via the cached Cosmos context            |
 * | BTC custom precompile move rejected | btctoken precompile (immovable)      | eth_call with movePrecompileToAddress             | request rejected; message contains "cannot move mezo custom precompile" |
 * | BlockOverrides.beaconRoot rejected  | post-Cancun field unsupported        | eth_simulateV1 with blockOverrides.beaconRoot     | -32602 with a BeaconRoot-specific message                               |
 * | BlockOverrides.withdrawals rejected | post-Cancun field unsupported        | eth_simulateV1 with blockOverrides.withdrawals    | -32602 with a Withdrawals-specific message                              |
 * | BlockOverrides.blobBaseFee rejected | post-Cancun field unsupported        | eth_simulateV1 with blockOverrides.blobBaseFee    | -32602 with a BlobBaseFee-specific message                              |
 * | stateRoot is the zero hash          | sender holds non-zero SimpleToken    | eth_simulateV1 runs an ERC-20 transfer            | block.stateRoot equals the 32-byte zero hash                            |
 * | gasUsed honors MinGasMultiplier     | default MinGasMultiplier=0.5         | value transfer with gas=100k (raw cost 21k)       | call.gasUsed = 50000 (gasLimit*MinGasMultiplier), not raw 21000         |
 * | insufficient-funds is per-call      | sender balance=0, validation omitted | eth_simulateV1 dispatches the value-transfer call | top-level success; per-call status=0x0; per-call error.code=-32015      |
 */
describe("SimulateV1_MezoDivergence", function () {
  const { deployments } = hre

  // The full custom-precompile range lives at 0x7b7c…00 through
  // 0x7b7c…15. The btctoken (0x7b7c…00) is the canonical pin: it touches
  // the bank keeper, so it surfaces both the "immovable" divergence and
  // the "cached Cosmos context survives call/block boundaries"
  // divergence.
  const BTC_TOKEN_PRECOMPILE = "0x7b7c000000000000000000000000000000000000"

  // Destination used by the eth_call move-precompile attempt below.
  const MOVE_DEST_ADDR = "0x0000000000000000000000000000000000001234"

  // 32-byte zero hash, used to pin the stateRoot divergence.
  const ZERO_HASH = "0x" + "00".repeat(32)

  // Detached EOA used by the validation=false insufficient-funds case.
  // Kept off the dev keys so the simulate request never collides with
  // real on-chain state.
  const ZERO_BALANCE_SENDER = "0xc100000000000000000000000000000000000001"
  const VALUE_RECIPIENT = "0xc200000000000000000000000000000000000002"

  // Per-call code emitted when the EVM's CanTransfer guard rejects a
  // value transfer with validation=false. NewSimVMError emits -32015
  // (SimErrCodeVMError) for non-revert VM errors; CanTransfer surfaces
  // as a non-revert VM failure rather than as the ExecutionReverted
  // path, so the per-call code is -32015, NOT the spec-reserved
  // request-level -38014.
  const SIM_VM_ERROR = -32015

  let simpleToken: SimpleToken
  let tokenAddr: string
  let senderAddr: string
  let recipientAddr: string

  async function simulateWithBlockOverrides(
    overrides: any,
  ): Promise<CapturedError> {
    const opts = {
      blockStateCalls: [{ blockOverrides: overrides, calls: [] }],
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

  before(async function () {
    await deployments.fixture(["BridgeOutDelegate"])
    simpleToken = await getDeployedContract<SimpleToken>("SimpleToken")
    tokenAddr = await simpleToken.getAddress()

    const [deployer] = await ethers.getSigners()
    senderAddr = deployer.address
    recipientAddr = ethers.Wallet.createRandom().address
  })

  context("custom Mezo precompile state chains across calls", function () {
    // Custom Mezo precompiles (btctoken, mezotoken, …) write through the
    // Cosmos bank keeper, not the EVM journal — their mutations ride in
    // the StateDB's cached Cosmos context and must survive call
    // boundaries. balance stateOverrides only touch the EVM state object
    // and do not propagate to bankKeeper, so this case leans on the
    // localnode dev0 signer's real pre-funded BTC balance.
    let result: any[]

    before(async function () {
      const btcToken: any = new hre.ethers.Contract(
        BTC_TOKEN_PRECOMPILE,
        btcabi,
        ethers.provider,
      )
      const transferAmount = ethers.parseEther("0.5")
      const transferData = btcToken.interface.encodeFunctionData("transfer", [
        recipientAddr,
        transferAmount,
      ])
      const balanceOfData = btcToken.interface.encodeFunctionData("balanceOf", [
        recipientAddr,
      ])

      result = await ethers.provider.send("eth_simulateV1", [
        {
          blockStateCalls: [
            {
              calls: [
                { from: senderAddr, to: BTC_TOKEN_PRECOMPILE, data: transferData },
                { from: senderAddr, to: BTC_TOKEN_PRECOMPILE, data: balanceOfData },
              ],
            },
          ],
        },
        "latest",
      ])
    })

    it("btctoken.transfer → btctoken.balanceOf surfaces the transferred amount", function () {
      expect(result).to.have.lengthOf(1)
      const calls = result[0].calls
      expect(calls).to.have.lengthOf(2)

      expect(calls[0].status).to.equal("0x1", "btctoken.transfer must succeed")
      expect(calls[1].status).to.equal("0x1", "btctoken.balanceOf must succeed")

      const [balance] = ethers.AbiCoder.defaultAbiCoder().decode(
        ["uint256"],
        calls[1].returnData,
      )
      expect(balance).to.equal(ethers.parseEther("0.5"))
    })
  })

  context("custom Mezo precompile state chains across blocks", function () {
    let result: any[]

    before(async function () {
      // Use a fresh recipient so this case doesn't observe the
      // single-block case's transfer through any shared persistent state.
      const recipient = ethers.Wallet.createRandom().address
      const btcToken: any = new hre.ethers.Contract(
        BTC_TOKEN_PRECOMPILE,
        btcabi,
        ethers.provider,
      )
      const transferAmount = ethers.parseEther("0.5")
      const transferData = btcToken.interface.encodeFunctionData("transfer", [
        recipient,
        transferAmount,
      ])
      const balanceOfData = btcToken.interface.encodeFunctionData("balanceOf", [
        recipient,
      ])

      result = await ethers.provider.send("eth_simulateV1", [
        {
          blockStateCalls: [
            {
              calls: [
                { from: senderAddr, to: BTC_TOKEN_PRECOMPILE, data: transferData },
              ],
            },
            {
              calls: [
                { from: senderAddr, to: BTC_TOKEN_PRECOMPILE, data: balanceOfData },
              ],
            },
          ],
        },
        "latest",
      ])
    })

    it("block 1 btctoken.transfer is visible to block 2 btctoken.balanceOf", function () {
      expect(result).to.have.lengthOf(2)

      const [block1, block2] = result
      expect(block1.calls[0].status).to.equal(
        "0x1",
        "block 1 btctoken.transfer must succeed",
      )

      const block2Call = block2.calls[0]
      expect(block2Call.status).to.equal(
        "0x1",
        "block 2 btctoken.balanceOf must succeed",
      )
      const [balance] = ethers.AbiCoder.defaultAbiCoder().decode(
        ["uint256"],
        block2Call.returnData,
      )
      expect(balance).to.equal(
        ethers.parseEther("0.5"),
        "block 2 must observe block 1's btctoken.transfer through the StateDB's cached Cosmos context",
      )
    })
  })

  context("custom Mezo precompiles are immovable", function () {
    let movePrecompileError: any

    before(async function () {
      try {
        await ethers.provider.send("eth_call", [
          { to: MOVE_DEST_ADDR, data: "0x" },
          "latest",
          {
            [BTC_TOKEN_PRECOMPILE]: { movePrecompileToAddress: MOVE_DEST_ADDR },
          },
        ])
        movePrecompileError = undefined
      } catch (err: any) {
        movePrecompileError = err
      }
    })

    it("rejects moving the BTC custom precompile with a Mezo-specific message", function () {
      expect(movePrecompileError, "eth_call should have raised").to.exist
      const message = extractMessage(movePrecompileError).toLowerCase()
      expect(message).to.include("cannot move mezo custom precompile")
    })
  })

  context("rejected post-Cancun BlockOverrides fields", function () {
    let beaconRootResult: CapturedError
    let withdrawalsResult: CapturedError
    let blobBaseFeeResult: CapturedError

    before(async function () {
      beaconRootResult = await simulateWithBlockOverrides({
        beaconRoot: ethers.ZeroHash,
      })
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
      blobBaseFeeResult = await simulateWithBlockOverrides({
        blobBaseFee: "0x1",
      })
    })

    it("BlockOverrides.beaconRoot is rejected with -32602 and a BeaconRoot-specific message", function () {
      expect(beaconRootResult.thrown).to.equal(true)
      expect(beaconRootResult.code).to.equal(-32602)
      expect(beaconRootResult.message).to.include(
        "BlockOverrides.BeaconRoot is not supported",
      )
    })

    it("BlockOverrides.withdrawals is rejected with -32602 and a Withdrawals-specific message", function () {
      expect(withdrawalsResult.thrown).to.equal(true)
      expect(withdrawalsResult.code).to.equal(-32602)
      expect(withdrawalsResult.message).to.include(
        "BlockOverrides.Withdrawals is not supported",
      )
    })

    it("BlockOverrides.blobBaseFee is rejected with -32602 and a BlobBaseFee-specific message", function () {
      expect(blobBaseFeeResult.thrown).to.equal(true)
      expect(blobBaseFeeResult.code).to.equal(-32602)
      expect(blobBaseFeeResult.message).to.include(
        "BlockOverrides.BlobBaseFee is not supported",
      )
    })
  })

  context("block envelope: stateRoot is zero", function () {
    let stateRootBlock: any

    before(async function () {
      // Mint enough to senderAddr so the simulated transfer below
      // succeeds against any live SimpleToken state. The mint goes
      // through real on-chain submission (not eth_simulateV1) so the
      // simulated transfer operates on a non-empty token balance and
      // produces a non-empty logsBloom + non-empty receipt root.
      // stateRoot, by contrast, is expected to remain zero — that's the
      // divergence this case pins.
      const [deployer] = await ethers.getSigners()
      const mintTx = await simpleToken
        .connect(deployer)
        .mint(senderAddr, ethers.parseUnits("1000", 18))
      await mintTx.wait()

      const transferData = simpleToken.interface.encodeFunctionData(
        "transfer",
        [recipientAddr, ethers.parseUnits("1", 18)],
      )
      const blocks: any[] = await ethers.provider.send("eth_simulateV1", [
        {
          blockStateCalls: [
            {
              calls: [{ from: senderAddr, to: tokenAddr, data: transferData }],
            },
          ],
        },
        "latest",
      ])
      stateRootBlock = blocks[0]
    })

    it("returns stateRoot equal to the 32-byte zero hash", function () {
      // Mezod's StateDB wraps a Cosmos cached multistore and has no
      // Merkle Patricia Trie, so there is no IntermediateRoot() to
      // call after a simulated block executes. Geth populates this
      // field from state.IntermediateRoot(); mezod cannot replicate
      // that without a chain-level architecture change. Echoing
      // base.Root would be misleading (ignores everything the
      // simulation did) and is explicitly rejected.
      expect(stateRootBlock.stateRoot.toLowerCase()).to.equal(ZERO_HASH)
    })
  })

  context("gasUsed honors MinGasMultiplier", function () {
    let blocks: any[]

    before(async function () {
      // Mezod applies a floor:
      //   gasUsed = max(gasLimit * MinGasMultiplier, raw_evm_gas)
      //
      // Default MinGasMultiplier = 0.5 (set in app/gas.go as
      // MainnetMinGasMultiplier and wired through feemarkettypes
      // DefaultMinGasMultiplier). For a value transfer the raw EVM gas
      // is 21000. Choosing gasLimit = 100000 (0x186a0) makes the floor
      // (50000) strictly greater than the raw cost, so the reported
      // value pins the floor. Geth reports raw 21000 here.
      blocks = await ethers.provider.send("eth_simulateV1", [
        {
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
                  value: "0x1",
                  gas: "0x186a0", // 100_000
                },
              ],
            },
          ],
        },
        "latest",
      ])
    })

    it("reports gasLimit * MinGasMultiplier when raw EVM gas is below the floor", function () {
      // gasLimit = 100_000 and MinGasMultiplier = 0.5 → floor = 50_000.
      // Raw value-transfer cost is 21_000, so the floor wins.
      expect(blocks).to.have.lengthOf(1)
      const calls = blocks[0].calls
      expect(calls).to.have.lengthOf(1)
      expect(calls[0].status).to.equal("0x1", "value transfer must succeed")
      expect(BigInt(calls[0].gasUsed)).to.equal(50000n)
    })
  })

  context("insufficient-funds is per-call when validation is omitted", function () {
    // Geth promotes insufficient-funds to the top-level fatal -38014
    // even with validation=false (its CanTransfer simply doesn't fire
    // because the validation gate caught it first). Mezod's CanTransfer
    // remains live regardless of the validation flag, so the per-call
    // status is 0x0 with the VM-error code (-32015) — the request
    // itself never aborts. Pinned: top-level success, per-call failure
    // with code -32015.
    let captured: { error: CapturedError; result: any[] | undefined }

    before(async function () {
      try {
        const result: any[] = await ethers.provider.send("eth_simulateV1", [
          {
            blockStateCalls: [
              {
                stateOverrides: {
                  [ZERO_BALANCE_SENDER]: { balance: "0x0" },
                },
                calls: [
                  {
                    from: ZERO_BALANCE_SENDER,
                    to: VALUE_RECIPIENT,
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

    it("surfaces the failure as a per-call status 0x0 with the VM-error code", function () {
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
