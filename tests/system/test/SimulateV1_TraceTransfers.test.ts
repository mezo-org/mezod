import { expect } from "chai"
import hre, { ethers } from "hardhat"

// SimulateV1_TraceTransfers covers eth_simulateV1's `traceTransfers`
// option: when enabled, the response logs include synthetic ERC-7528
// Transfer entries (emitter address = 0xEeeE..eEeE) for every native
// BTC value-transfer call edge. When the option is omitted or set to
// false, only real EVM logs surface.
describe("SimulateV1_TraceTransfers", function () {
  // ERC-7528 pseudo-address used as the emitter for synthetic native
  // value-transfer logs. Lowercased here so case-insensitive comparisons
  // line up with whatever the JSON-RPC response casing is.
  const ERC7528_ADDRESS =
    "0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee".toLowerCase()
  const TRANSFER_TOPIC = ethers.id("Transfer(address,address,uint256)")

  // Minimal "value forwarder" contract bytecode: forwards the entire
  // CALLVALUE to the address packed (left-padded) into the first 32
  // bytes of calldata. The only side effect is the inner CALL frame —
  // exactly the tracer's onEnter trigger surface.
  //
  // CALL stack layout the EVM expects (top-of-stack first): gas,
  // address, value, argsOffset, argsSize, retOffset, retSize. The push
  // sequence below populates these in reverse so the operand at the
  // top of the stack is `gas` once CALL executes.
  //
  //   @00  PUSH1 0x00          ; retSize
  //   @02  PUSH1 0x00          ; retOffset
  //   @04  PUSH1 0x00          ; argsSize
  //   @06  PUSH1 0x00          ; argsOffset
  //   @08  CALLVALUE           ; value (passed value)
  //   @09  PUSH1 0x00          ; CALLDATALOAD offset
  //   @0B  CALLDATALOAD        ; recipient (low 20 bytes of calldata[0..32])
  //   @0C  GAS                 ; forward all remaining gas
  //   @0D  CALL
  //   @0E  STOP
  //
  // Encoded: 60 00 60 00 60 00 60 00 34 60 00 35 5A F1 00 (15 bytes).
  const VALUE_FORWARDER_BYTECODE = "0x6000600060006000346000355AF100"

  let senderAddr: string
  let recipientAddr: string
  let forwarderAddr: string

  // Encoded calldata: a 32-byte left-padded recipient address that
  // CALLDATALOAD picks up at offset 0.
  let forwardCalldata: string

  // Captured response state for each scenario, populated once in
  // `before`. The skill-mandated convention (no beforeEach, no
  // execute-in-it) drives all transactions through this hook.
  let traceOnEoaToEoaResult: any
  let traceOnContractCallResult: any
  let traceOffContractCallResult: any

  before(async function () {
    senderAddr = (await ethers.getSigners())[0].address
    recipientAddr = ethers.Wallet.createRandom().address
    forwarderAddr = ethers.getAddress(
      "0x000000000000000000000000000000000000f0f0",
    )

    // 32-byte left-padded recipient address packed into calldata. The
    // forwarder bytecode reads CALLDATALOAD(0) and uses the low 20
    // bytes as the recipient.
    forwardCalldata = ethers.zeroPadValue(recipientAddr, 32)

    const transferValue = ethers.toQuantity(ethers.parseEther("1"))
    const senderBalance = ethers.toQuantity(ethers.parseEther("100"))
    const stateOverrides = {
      [senderAddr]: { balance: senderBalance },
      [forwarderAddr]: { code: VALUE_FORWARDER_BYTECODE },
    }

    // Scenario 1: traceTransfers=true, top-level EOA-to-EOA value
    // transfer. The driver opens a single CALL frame at depth 0 carrying
    // the value, which the tracer must surface as a synthetic log.
    const eoaOpts = {
      blockStateCalls: [
        {
          stateOverrides: { [senderAddr]: { balance: senderBalance } },
          calls: [
            {
              from: senderAddr,
              to: recipientAddr,
              value: transferValue,
            },
          ],
        },
      ],
      traceTransfers: true,
    }
    const eoaBlocks: any[] = await ethers.provider.send("eth_simulateV1", [
      eoaOpts,
      "latest",
    ])
    traceOnEoaToEoaResult = eoaBlocks[0].calls[0]

    // Scenario 2: traceTransfers=true, contract -> EOA via inner CALL.
    // Sender funds the forwarder contract with `value`; the forwarder
    // immediately re-calls the recipient with the same value. Two
    // value-transfer call edges -> two synthetic logs.
    const contractOpts = {
      blockStateCalls: [
        {
          stateOverrides,
          calls: [
            {
              from: senderAddr,
              to: forwarderAddr,
              value: transferValue,
              data: forwardCalldata,
            },
          ],
        },
      ],
      traceTransfers: true,
    }
    const contractBlocks: any[] = await ethers.provider.send(
      "eth_simulateV1",
      [contractOpts, "latest"],
    )
    traceOnContractCallResult = contractBlocks[0].calls[0]

    // Scenario 3: same call as scenario 2 but with traceTransfers
    // omitted. No synthetic logs must be present. Real EVM logs would
    // surface here too — the forwarder bytecode emits none, so the
    // expected logs list is empty.
    const traceOffOpts = {
      blockStateCalls: [
        {
          stateOverrides,
          calls: [
            {
              from: senderAddr,
              to: forwarderAddr,
              value: transferValue,
              data: forwardCalldata,
            },
          ],
        },
      ],
      // traceTransfers intentionally absent (default false)
    }
    const traceOffBlocks: any[] = await ethers.provider.send(
      "eth_simulateV1",
      [traceOffOpts, "latest"],
    )
    traceOffContractCallResult = traceOffBlocks[0].calls[0]
  })

  context("traceTransfers=true, top-level EOA-to-EOA value transfer", function () {
    it("returns status 0x1", function () {
      expect(traceOnEoaToEoaResult.status).to.equal("0x1")
    })

    it("emits exactly one synthetic ERC-7528 log", function () {
      const synthetic = traceOnEoaToEoaResult.logs.filter(
        (log: any) => log.address.toLowerCase() === ERC7528_ADDRESS,
      )
      expect(synthetic).to.have.lengthOf(1)
    })

    it("synthetic log carries the canonical Transfer topic and indexed sender/recipient", function () {
      const synthetic = traceOnEoaToEoaResult.logs.filter(
        (log: any) => log.address.toLowerCase() === ERC7528_ADDRESS,
      )
      const log = synthetic[0]
      expect(log.topics[0]).to.equal(TRANSFER_TOPIC)
      expect(BigInt(log.topics[1])).to.equal(BigInt(senderAddr))
      expect(BigInt(log.topics[2])).to.equal(BigInt(recipientAddr))
      expect(BigInt(log.data)).to.equal(ethers.parseEther("1"))
    })
  })

  context("traceTransfers=true, contract forwards value to EOA", function () {
    it("returns status 0x1", function () {
      expect(traceOnContractCallResult.status).to.equal("0x1")
    })

    it("emits two synthetic logs — one per value-transfer call edge", function () {
      const synthetic = traceOnContractCallResult.logs.filter(
        (log: any) => log.address.toLowerCase() === ERC7528_ADDRESS,
      )
      // Edge 1: sender -> forwarder (top-level CALL with value).
      // Edge 2: forwarder -> recipient (inner CALL with value).
      expect(synthetic).to.have.lengthOf(2)
    })

    it("first synthetic log records sender -> forwarder", function () {
      const synthetic = traceOnContractCallResult.logs.filter(
        (log: any) => log.address.toLowerCase() === ERC7528_ADDRESS,
      )
      const log = synthetic[0]
      expect(log.topics[0]).to.equal(TRANSFER_TOPIC)
      expect(BigInt(log.topics[1])).to.equal(BigInt(senderAddr))
      expect(BigInt(log.topics[2])).to.equal(BigInt(forwarderAddr))
      expect(BigInt(log.data)).to.equal(ethers.parseEther("1"))
    })

    it("second synthetic log records forwarder -> recipient", function () {
      const synthetic = traceOnContractCallResult.logs.filter(
        (log: any) => log.address.toLowerCase() === ERC7528_ADDRESS,
      )
      const log = synthetic[1]
      expect(log.topics[0]).to.equal(TRANSFER_TOPIC)
      expect(BigInt(log.topics[1])).to.equal(BigInt(forwarderAddr))
      expect(BigInt(log.topics[2])).to.equal(BigInt(recipientAddr))
      expect(BigInt(log.data)).to.equal(ethers.parseEther("1"))
    })
  })

  context("traceTransfers omitted (regression guard)", function () {
    it("returns status 0x1", function () {
      expect(traceOffContractCallResult.status).to.equal("0x1")
    })

    it("emits NO synthetic ERC-7528 logs", function () {
      const synthetic = traceOffContractCallResult.logs.filter(
        (log: any) => log.address.toLowerCase() === ERC7528_ADDRESS,
      )
      expect(synthetic).to.have.lengthOf(0)
    })
  })
})
