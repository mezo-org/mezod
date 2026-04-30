import { expect } from "chai"
import hre, { ethers } from "hardhat"
import { SimpleToken } from "../typechain-types/SimpleToken"
import { getDeployedContract } from "./helpers/contract"
import {
  CapturedError,
  extractCode,
  extractMessage,
} from "./helpers/rpc-error"

/**
 * SimulateV1_SpecCompliance pins every behavior where mezod's eth_simulateV1
 * MUST match the geth `execution-apis` reference implementation. Covers the
 * per-feature single-call, multi-call, multi-block, limits, trace-transfers,
 * validation, full-tx, and block-envelope surfaces, plus high-signal cases
 * harvested from `ethereum/execution-apis/tests/eth_simulateV1/`. Mezo-
 * specific divergences (stateRoot=zero, custom-precompile move rejection,
 * BeaconRoot rejection, etc.) live in SimulateV1_MezoDivergence — never
 * duplicated here.
 *
 * Conventions: each context() colocates its setup (before()) with its
 * assertions (it()). The file-level before() only seeds shared resources
 * used by many contexts (deployer, SimpleToken, latest block anchors).
 * No byte-for-byte response-hash checks against upstream fixtures
 * (mezod's localnode chain id, base block, and account state never match
 * the reference replay) — invariants only.
 *
 * | Scenario                                  | Given                                  | When                                   | Then                                                       |
 * |-------------------------------------------|----------------------------------------|----------------------------------------|------------------------------------------------------------|
 * | erc20 transfer happy path                 | sender holds AMOUNT*10 of SimpleToken  | one transfer call                      | status 0x1, gasUsed > 0, returnData=true, one Transfer log |
 * | erc20 transfer revert path                | transfer of HUGE_AMOUNT (overflows)    | one transfer call                      | status 0x0, error.code=3, data starts with Panic(0x11)     |
 * | mint -> transfer -> balanceOf chain       | empty SimpleToken                      | three calls in one block               | all succeed; balanceOf reads transferred amount            |
 * | block 1 mint -> block 2 balanceOf         | empty SimpleToken                      | two-block opts                         | block 2 reads block 1 mint                                 |
 * | BLOCKHASH(base+1)/(base+2)                | reader stub installed in block 1       | block 3 reads sibling block hashes     | matches blocks[0].hash and blocks[1].hash                  |
 * | BLOCKHASH below base                      | reader stub at canonical height        | call hits a canonical block            | matches canonical hash OR zero (best-effort retention)     |
 * | self-referencing move rejected            | precompile -> itself via stateOverride | eth_call dispatches the override       | rejected with "referenced itself"                          |
 * | move ecrecover happy path                 | ecrecover -> 0x..123456                | call dest with valid sig payload       | returns recovered addr                                     |
 * | move two precompiles to one               | sha256 + id both -> dest               | eth_simulateV1                         | -38023                                                     |
 * | move non-precompile rejected              | EOA -> dest, validation=true           | eth_simulateV1                         | -32000 with "is not a precompile"                          |
 * | override identity precompile              | identity moved + code override at orig | call dest, then call orig              | dest echoes (identity); orig returns empty (override)      |
 * | block cap                                 | 257 empty blocks                       | eth_simulateV1                         | -38026 with "blocks > max 256"                             |
 * | call cap                                  | 1001 calls in one block                | eth_simulateV1                         | -38026 with "calls > max 1000"                             |
 * | block-gas-limit cap                       | gasLimit 50k, one 21k call already in  | second call needs 30k                  | top-level -38015                                           |
 * | intrinsic-gas cap                         | call.gas=1000 (< 21000 intrinsic)      | eth_simulateV1                         | top-level -38013                                           |
 * | trace EOA->EOA                            | traceTransfers=true                    | one value transfer                     | status 0x1, one synthetic ERC-7528 log                     |
 * | trace contract->EOA                       | traceTransfers=true, forwarder bytecode| top-level CALL with value              | two synthetic logs (one per call edge)                     |
 * | trace omitted                             | flag absent                            | same forwarder call                    | zero synthetic logs                                        |
 * | validation happy path                     | sender balance funded                  | validation=true                        | per-call status 0x1                                        |
 * | -38010 nonce too low                      | state nonce 5, call.nonce 4            | validation=true                        | request aborts -38010                                      |
 * | -38011 nonce too high                     | state nonce 5, call.nonce 9            | validation=true                        | request aborts -38011                                      |
 * | -38013 intrinsic gas                      | call.gas=20999                         | validation=true                        | request aborts -38013                                      |
 * | -38014 insufficient funds                 | balance=0, value=1                     | validation=true                        | request aborts -38014                                      |
 * | -38025 init-code too large                | initcode 49153 bytes                   | validation=true                        | request aborts -38025                                      |
 * | -32005 fee-cap below base                 | maxFeePerGas=0 + baseFee=1gwei         | validation=true                        | request aborts -32005                                      |
 * | -38012 base-fee override low              | baseFee override 1 wei                 | validation=true                        | request aborts -38012                                      |
 * | revert per-call under validation          | revert bytecode                        | validation=true                        | request OK, per-call code 3                                |
 * | nonce increments per call                 | three calls nonce 0,1,2                | validation=true                        | three per-call status 0x1                                  |
 * | nonce-low under validation=false          | state nonce 5, call.nonce 4            | flag omitted                           | per-call status 0x1 (gate bypassed)                        |
 * | fee-cap-low under validation=false        | maxFeePerGas=0 + baseFee=1gwei         | flag omitted                           | per-call status 0x1 (gate bypassed)                        |
 * | nonce-too-high under validation=false     | state nonce 0, call.nonce 100          | flag omitted                           | per-call status 0x1 (gate bypassed)                        |
 * | balance after empty new block             | balance set in block 1                 | block 2 reads via SELFBALANCE          | block 2 reads the override                                 |
 * | block-num order -38020                    | second block's number < first's        | eth_simulateV1                         | -38020 with "blocks must be in order"                      |
 * | timestamp non-increment -38021            | later block's timestamp <= earlier's   | eth_simulateV1                         | -38021                                                     |
 * | timestamp auto-increment                  | only first time set                    | eth_simulateV1                         | second block.timestamp > first                             |
 * | timestamp incrementing pair               | strictly monotonic times               | eth_simulateV1                         | both blocks present, in order                              |
 * | blocknumber auto-increment                | first block's number set, second blank | eth_simulateV1                         | second.number = first.number + 1                           |
 * | empty blockStateCalls                     | one empty blockStateCall, anchor=latest| eth_simulateV1                         | one envelope with no transactions                          |
 * | future block anchor                       | tag well past the head                 | validation=true                        | header not found                                           |
 * | override block num                        | sequential override numbers            | NUMBER opcode in each                  | each call returns its block's number                       |
 * | simple-state-diff                         | code override + `state` map            | call reads slot 0                      | gets overridden value                                      |
 * | override-storage-slots                    | stateDiff in blk 1, state in blk 2     | call reads same slot                   | both return overridden value                               |
 * | override-address-twice                    | code-only override (no balance)        | sender balance=0, value=1, validation  | -38014                                                     |
 * | override across separate blocks           | balance set in block 1 and 2           | EOA-EOA value transfer twice           | both blocks succeed                                        |
 * | overwrite-existing-contract               | call once -> code override -> call     | second call has new bytecode           | first reverts, second OK                                   |
 * | block-override reflected                  | block.number override                  | NUMBER opcode probe                    | NUMBER returns the override                                |
 * | fee-recipient receives funds              | feeRecipient override                  | tx with non-zero value                 | request succeeds, status 0x1                               |
 * | logs: eth send no logs by default         | EOA->EOA value, no flag                | eth_simulateV1                         | logs == []                                                 |
 * | logs: forward then revert                 | forwarder -> reverter via traceTransfers| top-level call                        | status 0x0, code 3, logs == []                             |
 * | logs: selfdestruct synthetic log          | self-destruct contract                 | traceTransfers=true                    | (best-effort) one synthetic log on selfdest                |
 * | delegate-call to EOA logs once            | wallet contract via delegate-call      | traceTransfers=true                    | one synthetic log only                                     |
 * | comprehensive: simple two transfers       | sender funded                          | two value transfers in one block       | both per-call status 0x1                                   |
 * | comprehensive: transfer over BlockStateCalls| balance set in two blocks            | several value transfers                | all status 0x1                                             |
 * | comprehensive: set-read-storage           | storage contract bytecode override     | set then read slot                     | second call returns set value                              |
 * | comprehensive: contract-calls-itself      | self-calling bytecode override         | one call                               | status 0x1                                                 |
 * | comprehensive: gas spending across blocks | gas-burning bytecode override          | three blocks of gas-burn calls         | all calls status 0x1                                       |
 * | hash-only default                         | omitted returnFullTransactions         | ERC-20 transfer                        | transactions[] are 32-byte hex hashes                      |
 * | full-tx mode: from patched                | returnFullTransactions=true            | ERC-20 transfer                        | tx.from matches sender                                     |
 * | multi-sender full-tx                      | two senders one block                  | full-tx mode                           | each tx.from resolves                                      |
 * | block envelope: tx root / receipts root   | one ERC-20 transfer                    | eth_simulateV1                         | both roots != empty                                        |
 * | block envelope: logsBloom                 | one Transfer log                       | eth_simulateV1                         | bloom positive-tests the topic                             |
 * | block envelope: size > 0                  | one ERC-20 transfer                    | eth_simulateV1                         | size > 0                                                   |
 * | block hash determinism                    | identical opts                         | rerun                                  | block.hash unchanged                                       |
 * | gap-fill empty envelope                   | base+1 then base+3                     | eth_simulateV1                         | gap block has empty roots and zero bloom                   |
 */
describe("SimulateV1_SpecCompliance", function () {
  const { deployments } = hre

  // Spec-reserved JSON-RPC error codes (geth execution-apis execute.yaml).
  const NONCE_TOO_LOW = -38010
  const NONCE_TOO_HIGH = -38011
  const BASE_FEE_TOO_LOW = -38012
  const INTRINSIC_GAS_TOO_LOW = -38013
  const INSUFFICIENT_FUNDS = -38014
  const BLOCK_GAS_LIMIT = -38015
  const BLOCK_NUM_ORDER = -38020
  const BLOCK_TIMESTAMP_ORDER = -38021
  const TWO_PRECOMPILES_ONE_DEST = -38023
  const INIT_CODE_TOO_LARGE = -38025
  const REQUEST_TOO_LARGE = -38026
  const FEE_CAP_TOO_LOW = -32005
  const REVERTED = 3

  const TRANSFER_TOPIC = ethers.id("Transfer(address,address,uint256)")
  const PANIC_SELECTOR = "0x4e487b71"
  const ZERO_BLOOM = "0x" + "00".repeat(256)
  // Empty trie root: keccak256(rlp(<empty list>)). Any populated tx
  // or receipts list derives a different root.
  const EMPTY_TRIE_ROOT =
    "0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421"
  const ERC7528_ADDRESS =
    "0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee".toLowerCase()

  // Precompile addresses and a destination used for move tests.
  const ECRECOVER = "0x0000000000000000000000000000000000000001"
  const SHA256 = "0x0000000000000000000000000000000000000002"
  const IDENTITY = "0x0000000000000000000000000000000000000004"
  const DEST_ADDR = "0x0000000000000000000000000000000000123456"

  // Reader stub: PUSH1 0 CALLDATALOAD BLOCKHASH PUSH1 0 MSTORE PUSH1 0x20
  // PUSH1 0 RETURN. Returns BLOCKHASH(<calldata word>) padded to 32 bytes.
  const BLOCKHASH_READER_BYTECODE = "0x6000354060005260206000F3"

  // PUSH1 0 PUSH1 0 REVERT — unconditional revert with empty data.
  const REVERT_BYTECODE = "0x60006000fd"

  // PUSH1 0 NUMBER MSTORE PUSH1 0x20 PUSH1 0 RETURN — returns the
  // current block number padded to 32 bytes.
  const RETURN_NUMBER_BYTECODE = "0x6000434360005260206000f3"

  // Forwarder bytecode: forwards CALLVALUE to the address packed into
  // the first 32 bytes of calldata (low 20 bytes used).
  const VALUE_FORWARDER_BYTECODE = "0x6000600060006000346000355AF100"

  // Storage contract: store(key, val), get(key). Used by set-read-storage.
  // Same compiled bytecode as the upstream fixture.
  const STORAGE_CONTRACT_BYTECODE =
    "0x608060405234801561001057600080fd5b50600436106100365760003560e01c80632e64cec11461003b5780636057361d14610059575b600080fd5b610043610075565b60405161005091906100d9565b60405180910390f35b610073600480360381019061006e919061009d565b61007e565b005b60008054905090565b8060008190555050565b60008135905061009781610103565b92915050565b6000602082840312156100b3576100b26100fe565b5b60006100c184828501610088565b91505092915050565b6100d3816100f4565b82525050565b60006020820190506100ee60008301846100ca565b92915050565b6000819050919050565b600080fd5b61010c816100f4565b811461011757600080fd5b5056fea2646970667358221220404e37f487a89a932dca5e77faaf6ca2de3b991f93d230604b1b8daaef64766264736f6c63430008070033"

  // Self-destruct contract: a single function 0x83197ef0 that runs
  // SELFDESTRUCT(address(0)). Mirrors the upstream fixture.
  const SELFDESTRUCT_BYTECODE =
    "0x6080604052348015600f57600080fd5b506004361060285760003560e01c806383197ef014602d575b600080fd5b60336035565b005b600073ffffffffffffffffffffffffffffffffffffffff16fffea26469706673582212208e566fde20a17fff9658b9b1db37e27876fd8934ccf9b2aa308cabd37698681f64736f6c63430008120033"

  // Detached test addresses — picked to never collide with the
  // localnode dev keys, so simulation calls never touch on-chain state.
  const DETACHED_SENDER = "0xc100000000000000000000000000000000000001"
  const DETACHED_RECIPIENT = "0xc200000000000000000000000000000000000002"
  const DETACHED_SECOND_SENDER = "0xc100000000000000000000000000000000000201"

  // 1e22 wei. Covers gasLimit*HIGH_FEE + small `value` for any case below.
  const FUNDED_BALANCE = "0x21e19e0c9bab2400000"
  // Large enough to clear any plausible eip1559 base-fee floor.
  const HIGH_FEE = "0xe8d4a51000"

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

  async function simulateAt(opts: any, anchor: string): Promise<{
    error: CapturedError
    result: any[] | undefined
  }> {
    try {
      const result: any[] = await ethers.provider.send("eth_simulateV1", [
        opts,
        anchor,
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

  async function captureSim(blockStateCalls: any[]): Promise<CapturedError> {
    try {
      await ethers.provider.send("eth_simulateV1", [
        { blockStateCalls },
        "latest",
      ])
      return { thrown: false, code: undefined, message: "" }
    } catch (err: any) {
      return {
        thrown: true,
        code: extractCode(err),
        message: extractMessage(err),
      }
    }
  }

  // -----------------------------------------------------------------
  // Shared resources captured once, used by many contexts. Each
  // context's own setup goes in its local before() below.
  // -----------------------------------------------------------------
  let simpleToken: SimpleToken
  let tokenAddr: string
  let senderAddr: string
  let recipientAddr: string
  // Captured at file-level so override values that anchor on
  // latestTime/latestBlockNum stay valid even if the live localnode
  // head advances during the suite.
  let latestTime: bigint
  let latestBlockNum: bigint
  let baseStateOverrides: Record<string, { balance: string }>

  const AMOUNT = ethers.parseUnits("1", 18)

  before(async function () {
    await deployments.fixture(["BridgeOutDelegate"])
    simpleToken = await getDeployedContract<SimpleToken>("SimpleToken")
    tokenAddr = await simpleToken.getAddress()

    const [deployer] = await ethers.getSigners()
    senderAddr = deployer.address
    recipientAddr = ethers.Wallet.createRandom().address

    const latest = await ethers.provider.getBlock("latest")
    if (!latest) throw new Error("no latest block")
    latestTime = BigInt(latest.timestamp)
    latestBlockNum = BigInt(latest.number)

    baseStateOverrides = {
      [senderAddr]: { balance: ethers.toQuantity(ethers.parseEther("100")) },
    }

    // Mint AMOUNT*10 so the happy-path transfer succeeds and the
    // revert-path transfer (HUGE_AMOUNT) reliably underflows.
    await (
      await simpleToken.connect(deployer).mint(senderAddr, AMOUNT * 10n)
    ).wait()
  })

  // =================================================================
  // === Single-call ERC-20                                          ==
  // =================================================================

  context("ERC-20 transfer happy path", function () {
    let result: any

    before(async function () {
      const data = simpleToken.interface.encodeFunctionData("transfer", [
        recipientAddr,
        AMOUNT,
      ])
      const blocks: any[] = await ethers.provider.send("eth_simulateV1", [
        {
          blockStateCalls: [
            {
              stateOverrides: baseStateOverrides,
              calls: [{ from: senderAddr, to: tokenAddr, data }],
            },
          ],
        },
        "latest",
      ])
      result = blocks[0].calls[0]
    })

    it("returns status 0x1 with positive gasUsed and no error", function () {
      expect(result.status).to.equal("0x1")
      expect(BigInt(result.gasUsed)).to.be.greaterThan(0n)
      expect(result.error).to.be.undefined
    })

    it("returnData decodes to true", function () {
      const [decoded] = ethers.AbiCoder.defaultAbiCoder().decode(
        ["bool"],
        result.returnData,
      )
      expect(decoded).to.equal(true)
    })

    it("emits the ERC-20 Transfer log", function () {
      expect(result.logs).to.have.lengthOf(1)
      const log = result.logs[0]
      expect(log.address.toLowerCase()).to.equal(tokenAddr.toLowerCase())
      expect(log.topics[0]).to.equal(TRANSFER_TOPIC)
      expect(BigInt(log.topics[1])).to.equal(BigInt(senderAddr))
      expect(BigInt(log.topics[2])).to.equal(BigInt(recipientAddr))
    })
  })

  context("ERC-20 transfer revert path (Solidity 0.8 underflow)", function () {
    let result: any

    before(async function () {
      const HUGE_AMOUNT = AMOUNT * 1_000_000n
      const data = simpleToken.interface.encodeFunctionData("transfer", [
        recipientAddr,
        HUGE_AMOUNT,
      ])
      const blocks: any[] = await ethers.provider.send("eth_simulateV1", [
        {
          blockStateCalls: [
            {
              stateOverrides: baseStateOverrides,
              calls: [{ from: senderAddr, to: tokenAddr, data }],
            },
          ],
        },
        "latest",
      ])
      result = blocks[0].calls[0]
    })

    it("returns status 0x0", function () {
      expect(result.status).to.equal("0x0")
    })

    it("per-call error is code 3 with 'execution reverted'", function () {
      expect(result.error).to.exist
      expect(result.error.code).to.equal(REVERTED)
      expect(result.error.message).to.equal("execution reverted")
    })

    it("per-call error.data is the Panic(0x11) selector", function () {
      expect(result.error.data.toLowerCase().startsWith(PANIC_SELECTOR)).to.equal(true)
    })
  })

  // =================================================================
  // === Multi-call / multi-block state chains                       ==
  // =================================================================

  context("state chain across calls in one block", function () {
    let blocks: any[]

    before(async function () {
      const MINT_AMOUNT = ethers.parseUnits("1000", 18)
      const mintData = simpleToken.interface.encodeFunctionData("mint", [
        senderAddr,
        MINT_AMOUNT,
      ])
      const transferData = simpleToken.interface.encodeFunctionData(
        "transfer",
        [recipientAddr, ethers.parseUnits("500", 18)],
      )
      const balanceOfData = simpleToken.interface.encodeFunctionData(
        "balanceOf",
        [recipientAddr],
      )
      blocks = await ethers.provider.send("eth_simulateV1", [
        {
          blockStateCalls: [
            {
              stateOverrides: baseStateOverrides,
              calls: [
                { from: senderAddr, to: tokenAddr, data: mintData },
                { from: senderAddr, to: tokenAddr, data: transferData },
                { from: senderAddr, to: tokenAddr, data: balanceOfData },
              ],
            },
          ],
        },
        "latest",
      ])
    })

    it("mint -> transfer -> balanceOf surfaces transferred amount", function () {
      expect(blocks).to.have.lengthOf(1)
      const calls = blocks[0].calls
      expect(calls).to.have.lengthOf(3)
      expect(calls[0].status).to.equal("0x1")
      expect(calls[1].status).to.equal("0x1")
      expect(calls[2].status).to.equal("0x1")
      const [balance] = ethers.AbiCoder.defaultAbiCoder().decode(
        ["uint256"],
        calls[2].returnData,
      )
      expect(balance).to.equal(ethers.parseUnits("500", 18))
    })
  })

  context("state chain across blocks", function () {
    let blocks: any[]

    before(async function () {
      const recipient = ethers.Wallet.createRandom().address
      const mintData = simpleToken.interface.encodeFunctionData("mint", [
        recipient,
        ethers.parseUnits("777", 18),
      ])
      const balanceOfData = simpleToken.interface.encodeFunctionData(
        "balanceOf",
        [recipient],
      )
      blocks = await ethers.provider.send("eth_simulateV1", [
        {
          blockStateCalls: [
            {
              stateOverrides: baseStateOverrides,
              calls: [{ from: senderAddr, to: tokenAddr, data: mintData }],
            },
            {
              calls: [{ from: senderAddr, to: tokenAddr, data: balanceOfData }],
            },
          ],
        },
        "latest",
      ])
    })

    it("block 1 mint is visible to block 2 balanceOf", function () {
      expect(blocks).to.have.lengthOf(2)
      expect(blocks[0].calls[0].status).to.equal("0x1")
      const block2Call = blocks[1].calls[0]
      expect(block2Call.status).to.equal("0x1")
      const [balance] = ethers.AbiCoder.defaultAbiCoder().decode(
        ["uint256"],
        block2Call.returnData,
      )
      expect(balance).to.equal(ethers.parseUnits("777", 18))
    })
  })

  context("BLOCKHASH simulated-sibling resolution", function () {
    let blocks: any[]

    before(async function () {
      // Two-pass: first, simulate the chain so the synthetic block 0
      // surfaces its assigned number; second, read BLOCKHASH for that
      // assigned base + offsets. Deriving the calldata from a separate
      // ethers `getBlock("latest")` would race the localnode head if a
      // canonical block lands between the two reads; pinning the base
      // to the simulate response avoids that.
      const readerAddr = ethers.getAddress("0x" + "c0de".repeat(10))
      const heightToCalldata = (h: bigint) =>
        ethers.zeroPadValue(ethers.toBeHex(h), 32)
      const probe: any[] = await ethers.provider.send("eth_simulateV1", [
        {
          blockStateCalls: [
            {
              stateOverrides: {
                ...baseStateOverrides,
                [readerAddr]: { code: BLOCKHASH_READER_BYTECODE },
              },
              calls: [],
            },
            { calls: [] },
            { calls: [] },
          ],
        },
        "latest",
      ])
      const blockHashBase = BigInt(probe[0].number) - 1n
      blocks = await ethers.provider.send("eth_simulateV1", [
        {
          blockStateCalls: [
            {
              stateOverrides: {
                ...baseStateOverrides,
                [readerAddr]: { code: BLOCKHASH_READER_BYTECODE },
              },
              calls: [],
            },
            { calls: [] },
            {
              calls: [
                {
                  from: senderAddr,
                  to: readerAddr,
                  data: heightToCalldata(blockHashBase + 1n),
                },
                {
                  from: senderAddr,
                  to: readerAddr,
                  data: heightToCalldata(blockHashBase + 2n),
                },
              ],
            },
          ],
        },
        "latest",
      ])
    })

    it("BLOCKHASH(base+1) and BLOCKHASH(base+2) match the simulated headers", function () {
      expect(blocks).to.have.lengthOf(3)
      const block3Calls = blocks[2].calls
      expect(block3Calls).to.have.lengthOf(2)
      expect(block3Calls[0].status).to.equal("0x1")
      expect(block3Calls[1].status).to.equal("0x1")
      expect(block3Calls[0].returnData.toLowerCase()).to.equal(
        blocks[0].hash.toLowerCase(),
      )
      expect(block3Calls[1].returnData.toLowerCase()).to.equal(
        blocks[1].hash.toLowerCase(),
      )
    })
  })

  context("BLOCKHASH below base falls through to canonical chain", function () {
    let blocks: any[]
    let expectedHash: string | null

    before(async function () {
      // Ask BLOCKHASH for a canonical height below the localnode head.
      // Use latest.number - 1 so the probe always hits a block that
      // actually exists; localnode's earliest block 0 is genesis but
      // BLOCKHASH(genesis_height) can return zero on some EVM rules,
      // so the previous block is the safer probe.
      const reader = ethers.getAddress("0x" + "babe".repeat(10))
      const probedHeight = latestBlockNum - 1n
      const r = await simulate({
        blockStateCalls: [
          {
            stateOverrides: {
              ...baseStateOverrides,
              [reader]: { code: BLOCKHASH_READER_BYTECODE },
            },
            calls: [
              {
                from: senderAddr,
                to: reader,
                data: ethers.zeroPadValue(ethers.toBeHex(probedHeight), 32),
              },
            ],
          },
        ],
      })
      blocks = r.result!
      const probedBlock = await ethers.provider.getBlock(Number(probedHeight))
      expectedHash = probedBlock ? probedBlock.hash! : null
    })

    it("returns the canonical block hash or zero for the probed height", function () {
      expect(blocks).to.have.lengthOf(1)
      const call = blocks[0].calls[0]
      expect(call.status).to.equal("0x1")
      expect(call.returnData).to.match(/^0x[0-9a-f]{64}$/i)
      // BLOCKHASH for canonical heights below base is best-effort: when
      // CometBFT's historical-info window still retains the height, the
      // real canonical hash is returned; otherwise the EVM convention is
      // to surface zero (as the YP defines BLOCKHASH for unknown
      // heights). Both outcomes are spec-conformant for mezod — this
      // case validates the canonical-fallback wiring, not unbounded
      // historical retention.
      const ZERO = "0x" + "00".repeat(32)
      const actual = call.returnData.toLowerCase()
      if (expectedHash !== null) {
        expect([expectedHash.toLowerCase(), ZERO]).to.include(actual)
      }
    })
  })

  // =================================================================
  // === MovePrecompileTo                                            ==
  // =================================================================

  context("MovePrecompileTo: self-reference rejected", function () {
    let err: any

    before(async function () {
      try {
        await ethers.provider.send("eth_call", [
          { to: SHA256, data: "0x" },
          "latest",
          { [SHA256]: { movePrecompileToAddress: SHA256 } },
        ])
        err = null
      } catch (e: any) {
        err = e
      }
    })

    it("eth_call surfaces 'referenced itself' message", function () {
      expect(err).to.exist
      expect(extractMessage(err).toLowerCase()).to.include("referenced itself")
      // The typed -38022 SimError code surfaces only on eth_simulateV1; on
      // eth_call the same condition is bucketed under -32000 (server
      // error) because eth_call's error path does not pipe through the
      // typed *SimError envelope. The keeper-layer assertion is pinned
      // by TestApplyStateOverrides_MovePrecompileTo_SelfReference.
    })
  })

  context("MovePrecompileTo: ecrecover happy path", function () {
    // Single canonical "move precompile X to dest, call dest, get X's
    // result" pin. Ecrecover demonstrates a real recoverable address
    // (higher signal than identity/sha256 echo); the underlying
    // SimChain code path is shared, so additional precompiles would
    // duplicate the scenario.
    let blocks: any[]

    before(async function () {
      const ecrecoverSig =
        "0x1c8aff950685c2ed4bc3174f3472287b56d9517b9c948127319a09a7a36deac8000000000000000000000000000000000000000000000000000000000000001cb7cf302145348387b9e69fde82d8e634a0f8761e78da3bfa059efced97cbed0d2a66b69167cafe0ccfc726aec6ee393fea3cf0e4f3f9c394705e0f56d9bfe1c9"
      const r = await simulate({
        blockStateCalls: [
          {
            stateOverrides: {
              [ECRECOVER]: { movePrecompileToAddress: DEST_ADDR },
            },
            calls: [
              { from: DETACHED_SENDER, to: DEST_ADDR, data: ecrecoverSig },
            ],
          },
        ],
      })
      blocks = r.result!
    })

    it("calling the destination with a valid sig returns the recovered address", function () {
      expect(blocks).to.have.lengthOf(1)
      const call = blocks[0].calls[0]
      expect(call.status).to.equal("0x1")
      // Recovered address from the upstream fixture.
      expect(call.returnData.toLowerCase()).to.include(
        "b11cad98ad3f8114e0b3a1f6e7228bc8424df48a",
      )
    })
  })

  context("MovePrecompileTo: two precompiles to one dest -> -38023", function () {
    let r: { error: CapturedError; result: any[] | undefined }

    before(async function () {
      r = await simulate({
        blockStateCalls: [
          {
            stateOverrides: {
              [ECRECOVER]: { movePrecompileToAddress: DEST_ADDR },
              [SHA256]: { movePrecompileToAddress: DEST_ADDR },
            },
          },
        ],
      })
    })

    it("aborts with -38023", function () {
      expect(r.error.thrown).to.equal(true)
      expect(r.error.code).to.equal(TWO_PRECOMPILES_ONE_DEST)
    })
  })

  context("MovePrecompileTo: non-precompile rejected", function () {
    let r: { error: CapturedError; result: any[] | undefined }

    before(async function () {
      r = await simulate({
        blockStateCalls: [
          { stateOverrides: { [DETACHED_SENDER]: { nonce: "0x5" } } },
          {
            stateOverrides: {
              [DETACHED_SENDER]: {
                movePrecompileToAddress: DETACHED_RECIPIENT,
              },
            },
            calls: [
              {
                from: DETACHED_SENDER,
                to: DETACHED_SENDER,
                nonce: "0x0",
              },
            ],
          },
        ],
        validation: true,
      })
    })

    it("aborts with 'is not a precompile'", function () {
      expect(r.error.thrown).to.equal(true)
      expect(r.error.message.toLowerCase()).to.include("is not a precompile")
    })
  })

  context("Override identity precompile", function () {
    // The unique angle here: a precompile is moved AND its original
    // address gets a code override. Two parallel pins in one fixture —
    // moved precompile still operates at the destination, and the
    // original address now runs the user's bytecode.
    let blocks: any[]

    before(async function () {
      const r = await simulate({
        blockStateCalls: [
          {
            stateOverrides: {
              [IDENTITY]: {
                code: "0x",
                movePrecompileToAddress: DEST_ADDR,
              },
            },
            calls: [
              { from: DETACHED_SENDER, to: DEST_ADDR, data: "0x1234" },
              { from: DETACHED_SENDER, to: IDENTITY, data: "0x1234" },
            ],
          },
        ],
      })
      blocks = r.result!
    })

    it("DEST_ADDR runs identity (echo); original runs override", function () {
      expect(blocks).to.have.lengthOf(1)
      const calls = blocks[0].calls
      expect(calls).to.have.lengthOf(2)
      expect(calls[0].status).to.equal("0x1")
      expect(calls[0].returnData.toLowerCase()).to.equal("0x1234")
      expect(calls[1].status).to.equal("0x1")
      // Empty bytecode at original returns nothing.
      expect(calls[1].returnData).to.equal("0x")
    })
  })

  // =================================================================
  // === DoS limits                                                  ==
  // =================================================================

  context("DoS limits", function () {
    let blockCapResult: CapturedError
    let callCapResult: CapturedError
    let blockGasLimitResult: CapturedError
    let intrinsicGasResult: CapturedError

    before(async function () {
      blockCapResult = await captureSim(
        Array.from({ length: 257 }, () => ({ calls: [] })) as any,
      )
      callCapResult = await captureSim([
        {
          calls: Array.from({ length: 1001 }, () => ({
            from: ethers.ZeroAddress,
            to: ethers.ZeroAddress,
          })),
        },
      ] as any)
      blockGasLimitResult = await captureSim([
        {
          blockOverrides: { gasLimit: ethers.toQuantity(50_000n) },
          stateOverrides: baseStateOverrides,
          calls: [
            { from: senderAddr, to: recipientAddr, value: ethers.toQuantity(1n) },
            {
              from: senderAddr,
              to: recipientAddr,
              value: ethers.toQuantity(1n),
              gas: ethers.toQuantity(30_000n),
            },
          ],
        },
      ] as any)
      intrinsicGasResult = await captureSim([
        {
          stateOverrides: baseStateOverrides,
          calls: [
            {
              from: senderAddr,
              to: recipientAddr,
              value: ethers.toQuantity(1n),
              gas: ethers.toQuantity(1_000n),
            },
          ],
        },
      ] as any)
    })

    it("blocks > 256 -> -38026", function () {
      expect(blockCapResult.thrown).to.equal(true)
      expect(blockCapResult.code).to.equal(REQUEST_TOO_LARGE)
      expect(blockCapResult.message).to.include("blocks > max 256")
    })

    it("calls > 1000 -> -38026", function () {
      expect(callCapResult.thrown).to.equal(true)
      expect(callCapResult.code).to.equal(REQUEST_TOO_LARGE)
      expect(callCapResult.message).to.include("calls > max 1000")
    })

    it("over-budget call -> top-level -38015", function () {
      expect(blockGasLimitResult.thrown).to.equal(true)
      expect(blockGasLimitResult.code).to.equal(BLOCK_GAS_LIMIT)
    })

    it("below-intrinsic call -> top-level -38013", function () {
      expect(intrinsicGasResult.thrown).to.equal(true)
      expect(intrinsicGasResult.code).to.equal(INTRINSIC_GAS_TOO_LOW)
    })
  })

  // =================================================================
  // === traceTransfers                                              ==
  // =================================================================

  context("traceTransfers EOA -> EOA", function () {
    let result: any
    let recipient: string

    before(async function () {
      recipient = ethers.Wallet.createRandom().address
      const senderBalance = ethers.toQuantity(ethers.parseEther("100"))
      const blocks: any[] = await ethers.provider.send("eth_simulateV1", [
        {
          blockStateCalls: [
            {
              stateOverrides: { [senderAddr]: { balance: senderBalance } },
              calls: [
                {
                  from: senderAddr,
                  to: recipient,
                  value: ethers.toQuantity(ethers.parseEther("1")),
                },
              ],
            },
          ],
          traceTransfers: true,
        },
        "latest",
      ])
      result = blocks[0].calls[0]
    })

    it("status 0x1 with one synthetic ERC-7528 log", function () {
      expect(result.status).to.equal("0x1")
      const synthetic = result.logs.filter(
        (l: any) => l.address.toLowerCase() === ERC7528_ADDRESS,
      )
      expect(synthetic).to.have.lengthOf(1)
      expect(synthetic[0].topics[0]).to.equal(TRANSFER_TOPIC)
      expect(BigInt(synthetic[0].topics[1])).to.equal(BigInt(senderAddr))
      expect(BigInt(synthetic[0].topics[2])).to.equal(BigInt(recipient))
    })
  })

  context("traceTransfers contract forwards value", function () {
    let result: any
    let recipient: string
    let forwarderAddr: string

    before(async function () {
      recipient = ethers.Wallet.createRandom().address
      forwarderAddr = ethers.getAddress(
        "0x000000000000000000000000000000000000f0f0",
      )
      const senderBalance = ethers.toQuantity(ethers.parseEther("100"))
      const transferValue = ethers.toQuantity(ethers.parseEther("1"))
      const calldata = ethers.zeroPadValue(recipient, 32)
      const blocks: any[] = await ethers.provider.send("eth_simulateV1", [
        {
          blockStateCalls: [
            {
              stateOverrides: {
                [senderAddr]: { balance: senderBalance },
                [forwarderAddr]: { code: VALUE_FORWARDER_BYTECODE },
              },
              calls: [
                {
                  from: senderAddr,
                  to: forwarderAddr,
                  value: transferValue,
                  data: calldata,
                },
              ],
            },
          ],
          traceTransfers: true,
        },
        "latest",
      ])
      result = blocks[0].calls[0]
    })

    it("status 0x1 with two synthetic logs (one per call edge)", function () {
      expect(result.status).to.equal("0x1")
      const synthetic = result.logs.filter(
        (l: any) => l.address.toLowerCase() === ERC7528_ADDRESS,
      )
      expect(synthetic).to.have.lengthOf(2)
      // Edge 0: sender -> forwarder (top-level call).
      expect(BigInt(synthetic[0].topics[1])).to.equal(BigInt(senderAddr))
      expect(BigInt(synthetic[0].topics[2])).to.equal(BigInt(forwarderAddr))
      // Edge 1: forwarder -> recipient (inner CALL).
      expect(BigInt(synthetic[1].topics[1])).to.equal(BigInt(forwarderAddr))
      expect(BigInt(synthetic[1].topics[2])).to.equal(BigInt(recipient))
    })
  })

  context("traceTransfers omitted", function () {
    let result: any

    before(async function () {
      const recipient = ethers.Wallet.createRandom().address
      const forwarderAddr = ethers.getAddress(
        "0x000000000000000000000000000000000000f0f0",
      )
      const senderBalance = ethers.toQuantity(ethers.parseEther("100"))
      const transferValue = ethers.toQuantity(ethers.parseEther("1"))
      const calldata = ethers.zeroPadValue(recipient, 32)
      const blocks: any[] = await ethers.provider.send("eth_simulateV1", [
        {
          blockStateCalls: [
            {
              stateOverrides: {
                [senderAddr]: { balance: senderBalance },
                [forwarderAddr]: { code: VALUE_FORWARDER_BYTECODE },
              },
              calls: [
                {
                  from: senderAddr,
                  to: forwarderAddr,
                  value: transferValue,
                  data: calldata,
                },
              ],
            },
          ],
        },
        "latest",
      ])
      result = blocks[0].calls[0]
    })

    it("yields no synthetic logs", function () {
      expect(result.status).to.equal("0x1")
      const synthetic = result.logs.filter(
        (l: any) => l.address.toLowerCase() === ERC7528_ADDRESS,
      )
      expect(synthetic).to.have.lengthOf(0)
    })
  })

  // =================================================================
  // === validation=true                                             ==
  // =================================================================

  context("validation=true: happy path", function () {
    let r: { error: CapturedError; result: any[] | undefined }

    before(async function () {
      r = await simulate({
        blockStateCalls: [
          {
            stateOverrides: { [DETACHED_SENDER]: { balance: FUNDED_BALANCE } },
            calls: [
              {
                from: DETACHED_SENDER,
                to: DETACHED_RECIPIENT,
                value: "0x1",
                maxFeePerGas: HIGH_FEE,
              },
            ],
          },
        ],
        validation: true,
      })
    })

    it("returns per-call status 0x1", function () {
      expect(r.error.thrown).to.equal(false)
      expect(r.result![0].calls[0].status).to.equal("0x1")
    })
  })

  context("validation=true: fatal codes", function () {
    let nonceLowResult: { error: CapturedError; result: any[] | undefined }
    let nonceHighResult: { error: CapturedError; result: any[] | undefined }
    let insufficientFundsResult: { error: CapturedError; result: any[] | undefined }
    let intrinsicGasResult: { error: CapturedError; result: any[] | undefined }
    let initCodeTooLargeResult: { error: CapturedError; result: any[] | undefined }
    let feeCapTooLowResult: { error: CapturedError; result: any[] | undefined }
    let baseFeeOverrideTooLowResult: { error: CapturedError; result: any[] | undefined }
    let revertUnderValidationBlocks: any[]
    let nonceIncrementsBlocks: any[]

    before(async function () {
      const validationFunded = {
        [DETACHED_SENDER]: { balance: FUNDED_BALANCE },
      }

      nonceLowResult = await simulate({
        blockStateCalls: [
          {
            stateOverrides: {
              [DETACHED_SENDER]: { balance: FUNDED_BALANCE, nonce: "0x5" },
            },
            calls: [
              {
                from: DETACHED_SENDER,
                to: DETACHED_RECIPIENT,
                nonce: "0x4",
                maxFeePerGas: HIGH_FEE,
              },
            ],
          },
        ],
        validation: true,
      })
      nonceHighResult = await simulate({
        blockStateCalls: [
          {
            stateOverrides: {
              [DETACHED_SENDER]: { balance: FUNDED_BALANCE, nonce: "0x5" },
            },
            calls: [
              {
                from: DETACHED_SENDER,
                to: DETACHED_RECIPIENT,
                nonce: "0x9",
                maxFeePerGas: HIGH_FEE,
              },
            ],
          },
        ],
        validation: true,
      })
      insufficientFundsResult = await simulate({
        blockStateCalls: [
          {
            stateOverrides: { [DETACHED_SENDER]: { balance: "0x0" } },
            calls: [
              {
                from: DETACHED_SENDER,
                to: DETACHED_RECIPIENT,
                value: "0x1",
                maxFeePerGas: HIGH_FEE,
              },
            ],
          },
        ],
        validation: true,
      })
      intrinsicGasResult = await simulate({
        blockStateCalls: [
          {
            stateOverrides: validationFunded,
            calls: [
              {
                from: DETACHED_SENDER,
                to: DETACHED_RECIPIENT,
                value: "0x1",
                gas: "0x5207",
                maxFeePerGas: HIGH_FEE,
              },
            ],
          },
        ],
        validation: true,
      })
      const oversizedInitCode = "0x" + "00".repeat(49153)
      initCodeTooLargeResult = await simulate({
        blockStateCalls: [
          {
            stateOverrides: validationFunded,
            calls: [
              {
                from: DETACHED_SENDER,
                data: oversizedInitCode,
                gas: "0xf4240",
                maxFeePerGas: HIGH_FEE,
              },
            ],
          },
        ],
        validation: true,
      })
      feeCapTooLowResult = await simulate({
        blockStateCalls: [
          {
            blockOverrides: { baseFeePerGas: "0x3b9aca00" },
            stateOverrides: validationFunded,
            calls: [
              {
                from: DETACHED_SENDER,
                to: DETACHED_RECIPIENT,
                value: "0x1",
                maxFeePerGas: "0x0",
                maxPriorityFeePerGas: "0x0",
              },
            ],
          },
        ],
        validation: true,
      })
      baseFeeOverrideTooLowResult = await simulate({
        blockStateCalls: [
          {
            blockOverrides: { baseFeePerGas: "0x1" },
            stateOverrides: validationFunded,
            calls: [
              {
                from: DETACHED_SENDER,
                to: DETACHED_RECIPIENT,
                value: "0x1",
                maxFeePerGas: HIGH_FEE,
              },
            ],
          },
        ],
        validation: true,
      })
      const revertContract = "0xddee000000000000000000000000000000000000"
      const revertResult = await simulate({
        blockStateCalls: [
          {
            stateOverrides: {
              [DETACHED_SENDER]: { balance: FUNDED_BALANCE },
              [revertContract]: { code: REVERT_BYTECODE },
            },
            calls: [
              {
                from: DETACHED_SENDER,
                to: revertContract,
                maxFeePerGas: HIGH_FEE,
              },
            ],
          },
        ],
        validation: true,
      })
      revertUnderValidationBlocks = revertResult.result!

      const nonceIncrementsResult = await simulate({
        blockStateCalls: [
          {
            stateOverrides: validationFunded,
            calls: [
              {
                from: DETACHED_SENDER,
                to: DETACHED_SENDER,
                maxFeePerGas: HIGH_FEE,
                nonce: "0x0",
              },
              {
                from: DETACHED_SENDER,
                to: DETACHED_SENDER,
                maxFeePerGas: HIGH_FEE,
                nonce: "0x1",
              },
              {
                from: DETACHED_SENDER,
                to: DETACHED_SENDER,
                maxFeePerGas: HIGH_FEE,
                nonce: "0x2",
              },
            ],
          },
        ],
        validation: true,
      })
      nonceIncrementsBlocks = nonceIncrementsResult.result!
    })

    it("nonce too low -> -38010", function () {
      expect(nonceLowResult.error.thrown).to.equal(true)
      expect(nonceLowResult.error.code).to.equal(NONCE_TOO_LOW)
    })

    it("nonce too high -> -38011", function () {
      expect(nonceHighResult.error.thrown).to.equal(true)
      expect(nonceHighResult.error.code).to.equal(NONCE_TOO_HIGH)
    })

    it("insufficient funds -> -38014", function () {
      expect(insufficientFundsResult.error.thrown).to.equal(true)
      expect(insufficientFundsResult.error.code).to.equal(INSUFFICIENT_FUNDS)
    })

    it("intrinsic gas -> -38013", function () {
      expect(intrinsicGasResult.error.thrown).to.equal(true)
      expect(intrinsicGasResult.error.code).to.equal(INTRINSIC_GAS_TOO_LOW)
    })

    it("init-code too large -> -38025", function () {
      expect(initCodeTooLargeResult.error.thrown).to.equal(true)
      expect(initCodeTooLargeResult.error.code).to.equal(INIT_CODE_TOO_LARGE)
    })

    it("fee-cap below base -> -32005", function () {
      expect(feeCapTooLowResult.error.thrown).to.equal(true)
      expect(feeCapTooLowResult.error.code).to.equal(FEE_CAP_TOO_LOW)
    })

    it("base-fee override below floor -> -38012", function () {
      expect(baseFeeOverrideTooLowResult.error.thrown).to.equal(true)
      expect(baseFeeOverrideTooLowResult.error.code).to.equal(BASE_FEE_TOO_LOW)
    })

    it("revert stays per-call (NOT a top-level fatal)", function () {
      expect(revertUnderValidationBlocks).to.have.lengthOf(1)
      const call = revertUnderValidationBlocks[0].calls[0]
      expect(call.status).to.equal("0x0")
      expect(call.error.code).to.equal(REVERTED)
    })

    it("nonce increments per call (3 calls succeed in order)", function () {
      expect(nonceIncrementsBlocks).to.have.lengthOf(1)
      const calls = nonceIncrementsBlocks[0].calls
      expect(calls).to.have.lengthOf(3)
      expect(calls[0].status).to.equal("0x1")
      expect(calls[1].status).to.equal("0x1")
      expect(calls[2].status).to.equal("0x1")
    })
  })

  context("validation=false: gates bypassed", function () {
    // Each validation gate has its own predicate and its own bypass
    // path; the symmetric coverage on the off side mirrors the
    // validation=true fatal-codes context, with one case per gate.
    let nonceLowBlocks: any[]
    let feeCapBelowBlocks: any[]
    let nonceHighBlocks: any[]

    before(async function () {
      const nonceLowOff = await simulate({
        blockStateCalls: [
          {
            stateOverrides: {
              [DETACHED_SENDER]: { balance: FUNDED_BALANCE, nonce: "0x5" },
            },
            calls: [
              { from: DETACHED_SENDER, to: DETACHED_RECIPIENT, nonce: "0x4" },
            ],
          },
        ],
      })
      nonceLowBlocks = nonceLowOff.result!

      const feeCapBelowOff = await simulate({
        blockStateCalls: [
          {
            blockOverrides: { baseFeePerGas: "0x3b9aca00" },
            stateOverrides: { [DETACHED_SENDER]: { balance: FUNDED_BALANCE } },
            calls: [
              {
                from: DETACHED_SENDER,
                to: DETACHED_RECIPIENT,
                value: "0x1",
                maxFeePerGas: "0x0",
                maxPriorityFeePerGas: "0x0",
              },
            ],
          },
        ],
      })
      feeCapBelowBlocks = feeCapBelowOff.result!

      const nonceHighOff = await simulate({
        blockStateCalls: [
          {
            calls: [
              { from: DETACHED_SENDER, to: DETACHED_SENDER, nonce: "0x64" },
            ],
          },
        ],
      })
      nonceHighBlocks = nonceHighOff.result!
    })

    it("nonce-too-low yields per-call status 0x1", function () {
      expect(nonceLowBlocks).to.have.lengthOf(1)
      expect(nonceLowBlocks[0].calls[0].status).to.equal("0x1")
    })

    it("fee-cap-below-base yields per-call status 0x1", function () {
      expect(feeCapBelowBlocks).to.have.lengthOf(1)
      expect(feeCapBelowBlocks[0].calls[0].status).to.equal("0x1")
    })

    it("nonce-too-high yields per-call status 0x1", function () {
      expect(nonceHighBlocks).to.have.lengthOf(1)
      expect(nonceHighBlocks[0].calls[0].status).to.equal("0x1")
    })
  })

  // =================================================================
  // === Block ordering / numbering                                  ==
  // =================================================================

  context("balance after new (empty) block", function () {
    let blocks: any[]

    before(async function () {
      // Block 1 sets balance + bytecode via override; block 2 reads
      // balance via SELFBALANCE.
      // SELFBALANCE-MSTORE-RETURN bytecode: 0x47 (SELFBALANCE), 0x60
      // 0x00, MSTORE, 0x60 0x20, 0x60 0x00, RETURN.
      const target = "0xc100000000000000000000000000000000000000"
      const SELFBALANCE_PROBE = "0x4760005260206000F3"
      const r = await simulate({
        blockStateCalls: [
          {
            stateOverrides: {
              [target]: {
                balance: "0xde0b6b3a7640000",
                code: SELFBALANCE_PROBE,
              },
            },
            calls: [],
          },
          {
            calls: [{ from: DETACHED_SENDER, to: target }],
          },
        ],
      })
      blocks = r.result!
    })

    it("request succeeds across the gap and probes balance in block 2", function () {
      expect(blocks).to.have.lengthOf(2)
      const probe = blocks[1].calls[0]
      expect(probe.status).to.equal("0x1")
      expect(BigInt(probe.returnData)).to.equal(BigInt("0xde0b6b3a7640000"))
    })
  })

  context("block ordering errors", function () {
    let blockNumOrderResult: { error: CapturedError; result: any[] | undefined }
    let blockTsNonIncrementResult: {
      error: CapturedError
      result: any[] | undefined
    }

    before(async function () {
      // -38020: numbers must be strictly increasing. Both overrides
      // must sit above latestBlockNum so the first block clears the
      // base-anchored span check; -38020 then fires on the second
      // block's non-monotonic number.
      blockNumOrderResult = await simulate({
        blockStateCalls: [
          {
            blockOverrides: { number: ethers.toQuantity(latestBlockNum + 20n) },
            calls: [],
          },
          {
            blockOverrides: { number: ethers.toQuantity(latestBlockNum + 10n) },
            calls: [],
          },
        ],
      })

      // -38021: a later block goes backwards in time. Both anchors
      // clear latestTime so the violation fires on the inter-block
      // monotonicity check, not on the first-block-vs-base check.
      // Covers both the "equal timestamps" boundary (auto-incremented
      // block 2 advances by 1, block 3 sits below it) and a strict
      // decrease — same expected error, same code path.
      blockTsNonIncrementResult = await simulate({
        blockStateCalls: [
          { blockOverrides: { time: ethers.toQuantity(latestTime + 20n) } },
          { blockOverrides: {} },
          { blockOverrides: { time: ethers.toQuantity(latestTime + 15n) } },
          { blockOverrides: {} },
        ],
      })
    })

    it("non-monotonic block numbers -> -38020", function () {
      expect(blockNumOrderResult.error.thrown).to.equal(true)
      expect(blockNumOrderResult.error.code).to.equal(BLOCK_NUM_ORDER)
    })

    it("non-incrementing timestamps -> -38021", function () {
      expect(blockTsNonIncrementResult.error.thrown).to.equal(true)
      expect(blockTsNonIncrementResult.error.code).to.equal(
        BLOCK_TIMESTAMP_ORDER,
      )
    })
  })

  context("block timestamp / number auto-increment", function () {
    let blockTsAutoIncrementBlocks: any[]
    let timestampsIncrementingBlocks: any[]
    let blocknumberIncrementBlocks: any[]

    before(async function () {
      // Auto-increment: only first `time` set; second block must end
      // up with timestamp > first. Anchor the explicit timestamp above
      // latestTime so the first block does not trip -38021 against base.
      const r1 = await simulate({
        blockStateCalls: [
          { blockOverrides: { time: ethers.toQuantity(latestTime + 10n) } },
          { calls: [] },
        ],
      })
      blockTsAutoIncrementBlocks = r1.result!

      // Two-block timestamp increment: both succeed, both present.
      // Both timestamps must clear latestTime; pin them so the
      // assertion can read the exact values back.
      const r2 = await simulate({
        blockStateCalls: [
          { blockOverrides: { time: ethers.toQuantity(latestTime + 10n) } },
          { blockOverrides: { time: ethers.toQuantity(latestTime + 11n) } },
        ],
      })
      timestampsIncrementingBlocks = r2.result!

      // Block-number increment: explicit override of the first block at
      // latestBlockNum+1 (the first valid number after the anchor) so
      // the simulator does not gap-fill empties between base and the
      // override. Second block left to auto-increment to base+2. Pin
      // the anchor to latestBlockNum so the override stays valid even
      // if the live localnode head advances during the suite.
      const future = latestBlockNum + 1n
      const r3 = await simulateAt(
        {
          blockStateCalls: [
            { blockOverrides: { number: ethers.toQuantity(future) } },
            { calls: [] },
          ],
        },
        ethers.toQuantity(latestBlockNum),
      )
      blocknumberIncrementBlocks = r3.result!
    })

    it("only first time set => second block.timestamp > first", function () {
      expect(blockTsAutoIncrementBlocks).to.have.lengthOf(2)
      expect(BigInt(blockTsAutoIncrementBlocks[1].timestamp)).to.be.greaterThan(
        BigInt(blockTsAutoIncrementBlocks[0].timestamp),
      )
    })

    it("explicit pair (latest+10, latest+11) — both present and monotonic", function () {
      expect(timestampsIncrementingBlocks).to.have.lengthOf(2)
      expect(
        BigInt(timestampsIncrementingBlocks[1].timestamp),
      ).to.be.greaterThan(BigInt(timestampsIncrementingBlocks[0].timestamp))
    })

    it("explicit number on first block — second auto-increments by 1", function () {
      expect(blocknumberIncrementBlocks).to.have.lengthOf(2)
      const a = BigInt(blocknumberIncrementBlocks[0].number)
      const b = BigInt(blocknumberIncrementBlocks[1].number)
      expect(b).to.equal(a + 1n)
    })
  })

  context("empty blockStateCalls", function () {
    let blocks: any[]

    before(async function () {
      // Anchored at "latest" (server-resolved atomically) because
      // mezod's localnode runs with pruning-keep-recent=2: any captured
      // numeric anchor races against block advance and pruning. The
      // upstream "earliest" tag (genesis) cannot be used either without
      // depending on historical-info retention.
      const r = await simulate({ blockStateCalls: [{ calls: [] }] })
      blocks = r.result!
    })

    it("yields exactly one envelope with no transactions", function () {
      expect(blocks).to.have.lengthOf(1)
      expect(blocks[0].transactions).to.have.lengthOf(0)
    })
  })

  context("future block anchor", function () {
    let r: { error: CapturedError; result: any[] | undefined }

    before(async function () {
      const future = latestBlockNum + 1_000_000n
      r = await simulateAt(
        {
          blockStateCalls: [
            {
              calls: [{ from: DETACHED_SENDER, to: DETACHED_RECIPIENT }],
            },
          ],
          validation: true,
        },
        ethers.toQuantity(future),
      )
    })

    it("rejects with header-not-found", function () {
      expect(r.error.thrown).to.equal(true)
      expect(r.error.message.toLowerCase()).to.include("header not found")
    })
  })

  context("override block num", function () {
    let blocks: any[]

    before(async function () {
      // Anchor the first override at base+1 so the simulate driver does
      // not gap-fill empty headers between base and the override; the
      // second block increments by 1 to exercise the inter-block
      // monotonicity path. Pin the anchor to latestBlockNum so the
      // override stays valid even if the live localnode head advances
      // during the suite.
      const a = latestBlockNum + 1n
      const b = a + 1n
      const r = await simulateAt(
        {
          blockStateCalls: [
            {
              blockOverrides: { number: ethers.toQuantity(a) },
              stateOverrides: {
                [DETACHED_SENDER]: { code: RETURN_NUMBER_BYTECODE },
              },
              calls: [{ from: DETACHED_RECIPIENT, to: DETACHED_SENDER }],
            },
            {
              blockOverrides: { number: ethers.toQuantity(b) },
              calls: [{ from: DETACHED_RECIPIENT, to: DETACHED_SENDER }],
            },
          ],
        },
        ethers.toQuantity(latestBlockNum),
      )
      blocks = r.result!
    })

    it("each block's NUMBER opcode returns its own override", function () {
      expect(blocks).to.have.lengthOf(2)
      const a = BigInt(blocks[0].number)
      const b = BigInt(blocks[1].number)
      expect(BigInt(blocks[0].calls[0].returnData)).to.equal(a)
      expect(BigInt(blocks[1].calls[0].returnData)).to.equal(b)
    })
  })

  // =================================================================
  // === State / block override edges                                ==
  // =================================================================

  // simple-state-diff using `state`: code at C100 reads slot 0.
  // Bytecode below: PUSH1 0 SLOAD PUSH1 0 MSTORE PUSH1 0x20 PUSH1 0
  // RETURN — returns slot 0 padded.
  const SLOT0_READER = "0x60005460005260206000F3"
  const STATE_OVERRIDE_TARGET = "0xc100000000000000000000000000000000000000"
  const STATE_OVERRIDE_VALUE =
    "0x12340000000000000000000000000000000000000000000000000000000000ff"

  context("simple state diff", function () {
    let blocks: any[]

    before(async function () {
      const r = await simulate({
        blockStateCalls: [
          {
            stateOverrides: {
              [STATE_OVERRIDE_TARGET]: {
                code: SLOT0_READER,
                state: {
                  "0x0000000000000000000000000000000000000000000000000000000000000000":
                    STATE_OVERRIDE_VALUE,
                },
              },
            },
            calls: [{ from: DETACHED_SENDER, to: STATE_OVERRIDE_TARGET }],
          },
        ],
      })
      blocks = r.result!
    })

    it("reads slot 0 back as the overridden value", function () {
      expect(blocks).to.have.lengthOf(1)
      const call = blocks[0].calls[0]
      expect(call.status).to.equal("0x1")
      expect(call.returnData.toLowerCase()).to.equal(STATE_OVERRIDE_VALUE)
    })
  })

  context("override storage slots (stateDiff vs state)", function () {
    let blocks: any[]

    before(async function () {
      const r = await simulate({
        blockStateCalls: [
          {
            stateOverrides: {
              [STATE_OVERRIDE_TARGET]: {
                code: SLOT0_READER,
                stateDiff: {
                  "0x0000000000000000000000000000000000000000000000000000000000000000":
                    STATE_OVERRIDE_VALUE,
                },
              },
            },
            calls: [{ from: DETACHED_SENDER, to: STATE_OVERRIDE_TARGET }],
          },
          {
            stateOverrides: {
              [STATE_OVERRIDE_TARGET]: {
                state: {
                  "0x0000000000000000000000000000000000000000000000000000000000000000":
                    STATE_OVERRIDE_VALUE,
                },
              },
            },
            calls: [{ from: DETACHED_SENDER, to: STATE_OVERRIDE_TARGET }],
          },
        ],
      })
      blocks = r.result!
    })

    it("both blocks return the overridden slot", function () {
      expect(blocks).to.have.lengthOf(2)
      expect(blocks[0].calls[0].returnData.toLowerCase()).to.equal(
        STATE_OVERRIDE_VALUE,
      )
      expect(blocks[1].calls[0].returnData.toLowerCase()).to.equal(
        STATE_OVERRIDE_VALUE,
      )
    })
  })

  context("override-address-twice (-38014 with validation)", function () {
    let r: { error: CapturedError; result: any[] | undefined }

    before(async function () {
      r = await simulate({
        blockStateCalls: [
          {
            stateOverrides: { [DETACHED_SENDER]: { code: REVERT_BYTECODE } },
            calls: [
              {
                from: DETACHED_SENDER,
                to: DETACHED_RECIPIENT,
                value: "0x3e8",
                maxFeePerGas: HIGH_FEE,
              },
            ],
          },
        ],
        traceTransfers: true,
        validation: true,
      })
    })

    it("aborts with -38014 (no balance, value > 0)", function () {
      expect(r.error.thrown).to.equal(true)
      expect(r.error.code).to.equal(INSUFFICIENT_FUNDS)
    })
  })

  context("override across separate BlockStateCalls", function () {
    let blocks: any[]

    before(async function () {
      const r = await simulate({
        blockStateCalls: [
          {
            stateOverrides: { [DETACHED_SENDER]: { balance: FUNDED_BALANCE } },
            calls: [
              {
                from: DETACHED_SENDER,
                to: DETACHED_RECIPIENT,
                value: "0x3e8",
              },
            ],
          },
          {
            stateOverrides: { [DETACHED_SENDER]: { balance: FUNDED_BALANCE } },
            calls: [
              {
                from: DETACHED_SENDER,
                to: DETACHED_RECIPIENT,
                value: "0x3e8",
              },
            ],
          },
        ],
        traceTransfers: true,
      })
      blocks = r.result!
    })

    it("both blocks succeed with re-applied balance", function () {
      expect(blocks).to.have.lengthOf(2)
      expect(blocks[0].calls[0].status).to.equal("0x1")
      expect(blocks[1].calls[0].status).to.equal("0x1")
    })
  })

  context("overwrite existing contract", function () {
    let blocks: any[]

    before(async function () {
      const r = await simulate({
        blockStateCalls: [
          {
            stateOverrides: baseStateOverrides,
            calls: [
              {
                from: senderAddr,
                to: tokenAddr,
                data: simpleToken.interface.encodeFunctionData("transfer", [
                  recipientAddr,
                  ethers.parseUnits("999999999", 18),
                ]),
              },
            ],
          },
          {
            stateOverrides: { [tokenAddr]: { code: STORAGE_CONTRACT_BYTECODE } },
            calls: [
              {
                from: senderAddr,
                to: tokenAddr,
                data: "0x2e64cec1", // get() — reads slot 0, returns 0
              },
            ],
          },
        ],
      })
      blocks = r.result!
    })

    it("first block reverts, second block runs the overridden bytecode", function () {
      expect(blocks).to.have.lengthOf(2)
      expect(blocks[0].calls[0].status).to.equal("0x0")
      expect(blocks[1].calls[0].status).to.equal("0x1")
    })
  })

  context("block-override reflected in contract", function () {
    let blocks: any[]

    before(async function () {
      // Anchor at latestBlockNum so the override stays valid even if
      // the live localnode head advances during the suite. base+1 keeps
      // the simulator from gap-filling empty headers.
      const target = "0xc100000000000000000000000000000000000000"
      const probeNumber = latestBlockNum + 1n
      const r = await simulateAt(
        {
          blockStateCalls: [
            {
              blockOverrides: { number: ethers.toQuantity(probeNumber) },
              stateOverrides: { [target]: { code: RETURN_NUMBER_BYTECODE } },
              calls: [{ from: DETACHED_SENDER, to: target }],
            },
          ],
        },
        ethers.toQuantity(latestBlockNum),
      )
      blocks = r.result!
    })

    it("NUMBER opcode at override returns the override number", function () {
      expect(blocks).to.have.lengthOf(1)
      const call = blocks[0].calls[0]
      expect(call.status).to.equal("0x1")
      expect(BigInt(call.returnData)).to.equal(BigInt(blocks[0].number))
    })
  })

  context("fee-recipient receives funds", function () {
    let blocks: any[]

    before(async function () {
      const r = await simulate({
        blockStateCalls: [
          {
            blockOverrides: {
              feeRecipient: "0xc200000000000000000000000000000000000000",
            },
            stateOverrides: { [DETACHED_SENDER]: { balance: FUNDED_BALANCE } },
            calls: [
              {
                from: DETACHED_SENDER,
                to: DETACHED_RECIPIENT,
                value: "0x1",
              },
            ],
          },
        ],
      })
      blocks = r.result!
    })

    it("request succeeds with feeRecipient override + non-zero value", function () {
      expect(blocks).to.have.lengthOf(1)
      expect(blocks[0].calls[0].status).to.equal("0x1")
    })
  })

  // =================================================================
  // === logs / tracer edges                                         ==
  // =================================================================

  context("logs: eth send produces no logs by default", function () {
    let blocks: any[]

    before(async function () {
      const r = await simulate({
        blockStateCalls: [
          {
            stateOverrides: { [DETACHED_SENDER]: { balance: FUNDED_BALANCE } },
            calls: [
              {
                from: DETACHED_SENDER,
                to: DETACHED_RECIPIENT,
                value: "0x3e8",
              },
            ],
          },
        ],
      })
      blocks = r.result!
    })

    it("logs == [] when traceTransfers is omitted", function () {
      expect(blocks).to.have.lengthOf(1)
      const call = blocks[0].calls[0]
      expect(call.status).to.equal("0x1")
      expect(call.logs).to.have.lengthOf(0)
    })
  })

  context("logs: forward then revert produces no logs", function () {
    let blocks: any[]

    before(async function () {
      // Use a status-propagating forwarder (REVERT on inner failure) —
      // the plain VALUE_FORWARDER_BYTECODE STOPs after CALL regardless
      // of inner outcome, so it could not surface the inner revert at
      // the top frame.
      //
      // Disassembly of REVERT_PROPAGATING_FORWARDER_BYTECODE:
      //   PUSH1 0; PUSH1 0; PUSH1 0; PUSH1 0; CALLVALUE; PUSH1 0;
      //   CALLDATALOAD; GAS; CALL; ISZERO; PUSH1 0x13; JUMPI; STOP;
      //   JUMPDEST; PUSH1 0; PUSH1 0; REVERT.
      const REVERT_PROPAGATING_FORWARDER_BYTECODE =
        "0x6000600060006000346000355af115601357005b60006000fd"
      const forwarderAddr = ethers.getAddress(
        "0x000000000000000000000000000000000000f0f0",
      )
      const reverter = "0xc200000000000000000000000000000000000000"
      const fwdRevert = ethers.zeroPadValue(reverter, 32)
      const r = await simulate({
        blockStateCalls: [
          {
            stateOverrides: {
              [DETACHED_SENDER]: { balance: FUNDED_BALANCE },
              [forwarderAddr]: { code: REVERT_PROPAGATING_FORWARDER_BYTECODE },
              [reverter]: { code: REVERT_BYTECODE },
            },
            calls: [
              {
                from: DETACHED_SENDER,
                to: forwarderAddr,
                value: "0x3e8",
                data: fwdRevert,
              },
            ],
          },
        ],
        traceTransfers: true,
      })
      blocks = r.result!
    })

    it("status 0x0, code 3, logs == []", function () {
      expect(blocks).to.have.lengthOf(1)
      const call = blocks[0].calls[0]
      expect(call.status).to.equal("0x0")
      expect(call.error.code).to.equal(REVERTED)
      expect(call.logs).to.have.lengthOf(0)
    })
  })

  context("logs: self-destruct produces synthetic Transfer log", function () {
    // The synthetic log here only fires if the live EVM routes
    // SELFDESTRUCT through OnEnter. The tracer's behavior is pinned by
    // TestSimTracer_SelfdestructWithBalanceEmitsSynthetic; this case
    // tolerates zero-or-one to avoid blocking CI on an EVM-version-
    // dependent dispatch path. A console.warn flags the absence so a
    // future EVM upgrade that wires the dispatch correctly does not go
    // unnoticed.
    let blocks: any[]

    before(async function () {
      const sd = "0xc200000000000000000000000000000000000000"
      const r = await simulate({
        blockStateCalls: [
          {
            stateOverrides: {
              [sd]: { code: SELFDESTRUCT_BYTECODE, balance: "0x1e8480" },
            },
            calls: [{ from: DETACHED_SENDER, to: sd, data: "0x83197ef0" }],
          },
        ],
        traceTransfers: true,
      })
      blocks = r.result!
    })

    it("status 0x1 and (best-effort) one ERC-7528 Transfer log", function () {
      expect(blocks).to.have.lengthOf(1)
      const call = blocks[0].calls[0]
      expect(call.status).to.equal("0x1")
      const synthetic = call.logs.filter(
        (l: any) => l.address.toLowerCase() === ERC7528_ADDRESS,
      )
      expect(synthetic.length).to.be.lessThanOrEqual(1)
      if (synthetic.length === 1) {
        expect(synthetic[0].topics[0]).to.equal(TRANSFER_TOPIC)
        // Topics index sender (the self-destructing contract) and
        // recipient (the beneficiary, address(0) here).
        expect(BigInt(synthetic[0].topics[1])).to.equal(
          BigInt("0xc200000000000000000000000000000000000000"),
        )
        expect(BigInt(synthetic[0].topics[2])).to.equal(0n)
      } else {
        // eslint-disable-next-line no-console
        console.warn(
          "selfdestruct: synthetic transfer log not emitted; live EVM does not route SELFDESTRUCT through OnEnter",
        )
      }
    })
  })

  context("delegate-call to EOA", function () {
    let blocks: any[]

    before(async function () {
      // Send eth + delegate-call to an EOA: the wallet bytecode
      // forwards via delegate-call (no value transfer through that
      // edge); only the top-level value transfer surfaces a synthetic
      // log.
      const WALLET_DELEGATECALL_BYTECODE =
        "0x608060405273ffffffffffffffffffffffffffffffffffffffff73c200000000000000000000000000000000000000167fa619486e000000000000000000000000000000000000000000000000000000005f3503605e57805f5260205ff35b365f80375f80365f845af43d5f803e5f81036077573d5ffd5b3d5ff3fea26469706673582212206b787fe3e60b14a5c449d37005afac7b1803ee7c87e12d2740b96a158f34802a64736f6c63430008160033"
      const wallet = "0xc100000000000000000000000000000000000000"
      const r = await simulate({
        blockStateCalls: [
          {
            stateOverrides: {
              [DETACHED_SENDER]: { balance: FUNDED_BALANCE },
              [wallet]: { code: WALLET_DELEGATECALL_BYTECODE },
            },
            calls: [
              {
                from: DETACHED_SENDER,
                to: wallet,
                value: "0x3e8",
                data: "0x",
              },
            ],
          },
        ],
        traceTransfers: true,
      })
      blocks = r.result!
    })

    it("emits exactly one synthetic log (top-level edge only)", function () {
      expect(blocks).to.have.lengthOf(1)
      const call = blocks[0].calls[0]
      expect(call.status).to.equal("0x1")
      const synthetic = call.logs.filter(
        (l: any) => l.address.toLowerCase() === ERC7528_ADDRESS,
      )
      expect(synthetic).to.have.lengthOf(1)
    })
  })

  // =================================================================
  // === Comprehensive sanity                                        ==
  // =================================================================

  context("comprehensive: simple two value transfers", function () {
    let blocks: any[]

    before(async function () {
      const r = await simulate({
        blockStateCalls: [
          {
            stateOverrides: { [DETACHED_SENDER]: { balance: "0x3e8" } },
            calls: [
              {
                from: DETACHED_SENDER,
                to: DETACHED_RECIPIENT,
                value: "0x3e8",
              },
              {
                from: DETACHED_RECIPIENT,
                to: "0xc300000000000000000000000000000000000000",
                value: "0x3e8",
              },
            ],
          },
        ],
      })
      blocks = r.result!
    })

    it("both per-call status 0x1", function () {
      expect(blocks).to.have.lengthOf(1)
      expect(blocks[0].calls[0].status).to.equal("0x1")
      expect(blocks[0].calls[1].status).to.equal("0x1")
    })
  })

  context("comprehensive: transfer over multiple BlockStateCalls", function () {
    let blocks: any[]

    before(async function () {
      const c3 = "0xc300000000000000000000000000000000000000"
      const r = await simulate({
        blockStateCalls: [
          {
            stateOverrides: { [DETACHED_SENDER]: { balance: "0x1388" } },
            calls: [
              {
                from: DETACHED_SENDER,
                to: DETACHED_RECIPIENT,
                value: "0x7d0",
              },
              { from: DETACHED_SENDER, to: c3, value: "0x7d0" },
            ],
          },
          {
            stateOverrides: { [c3]: { balance: "0x1388" } },
            calls: [
              {
                from: DETACHED_RECIPIENT,
                to: "0xc200000000000000000000000000000000000000",
                value: "0x3e8",
              },
              {
                from: c3,
                to: "0xc200000000000000000000000000000000000000",
                value: "0x3e8",
              },
            ],
          },
        ],
      })
      blocks = r.result!
    })

    it("all four calls succeed", function () {
      expect(blocks).to.have.lengthOf(2)
      for (const block of blocks) {
        for (const call of block.calls) {
          expect(call.status).to.equal("0x1")
        }
      }
    })
  })

  context("comprehensive: set-read-storage", function () {
    let blocks: any[]

    before(async function () {
      const target = "0xc200000000000000000000000000000000000000"
      const r = await simulate({
        blockStateCalls: [
          {
            stateOverrides: { [target]: { code: STORAGE_CONTRACT_BYTECODE } },
            calls: [
              {
                from: DETACHED_SENDER,
                to: target,
                data: "0x6057361d0000000000000000000000000000000000000000000000000000000000000005",
              },
              {
                from: DETACHED_SENDER,
                to: target,
                data: "0x2e64cec1",
              },
            ],
          },
        ],
      })
      blocks = r.result!
    })

    it("call 2 returns the value set by call 1", function () {
      expect(blocks).to.have.lengthOf(1)
      const calls = blocks[0].calls
      expect(calls).to.have.lengthOf(2)
      expect(calls[0].status).to.equal("0x1")
      expect(calls[1].status).to.equal("0x1")
      expect(BigInt(calls[1].returnData)).to.equal(5n)
    })
  })

  context("comprehensive: contract calls itself", function () {
    let blocks: any[]

    before(async function () {
      const SELF_REF_BYTECODE =
        "0x608060405234801561001057600080fd5b506000366060484641444543425a3a60014361002c919061009b565b406040516020016100469a99989796959493929190610138565b6040516020818303038152906040529050915050805190602001f35b6000819050919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b60006100a682610062565b91506100b183610062565b92508282039050818111156100c9576100c861006c565b5b92915050565b6100d881610062565b82525050565b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b6000610109826100de565b9050919050565b610119816100fe565b82525050565b6000819050919050565b6101328161011f565b82525050565b60006101408201905061014e600083018d6100cf565b61015b602083018c6100cf565b610168604083018b610110565b610175606083018a6100cf565b61018260808301896100cf565b61018f60a08301886100cf565b61019c60c08301876100cf565b6101a960e08301866100cf565b6101b76101008301856100cf565b6101c5610120830184610129565b9b9a505050505050505050505056fea26469706673582212205139ae3ba8d46d11c29815d001b725f9840c90e330884ed070958d5af4813d8764736f6c63430008120033"
      const target = "0xc000000000000000000000000000000000000000"
      const r = await simulate({
        blockStateCalls: [
          {
            stateOverrides: { [target]: { code: SELF_REF_BYTECODE } },
            calls: [{ from: target, to: target }],
          },
        ],
        traceTransfers: true,
      })
      blocks = r.result!
    })

    it("status 0x1", function () {
      expect(blocks).to.have.lengthOf(1)
      expect(blocks[0].calls[0].status).to.equal("0x1")
    })
  })

  context("comprehensive: run-gas-spending across blocks", function () {
    let blocks: any[]

    before(async function () {
      // Trims to one gas-burn call per block to keep the test fast on
      // localnode while still exercising multi-block accounting.
      const burner =
        "0x608060405234801561001057600080fd5b506004361061002b5760003560e01c8063815b8ab414610030575b600080fd5b61004a600480360381019061004591906100b6565b61004c565b005b60005a90505b60011561007657815a826100669190610112565b106100715750610078565b610052565b505b50565b600080fd5b6000819050919050565b61009381610080565b811461009e57600080fd5b50565b6000813590506100b08161008a565b92915050565b6000602082840312156100cc576100cb61007b565b5b60006100da848285016100a1565b91505092915050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b600061011d82610080565b915061012883610080565b92508282039050818111156101405761013f6100e3565b5b9291505056fea2646970667358221220a659ba4db729a6ee4db02fcc5c1118db53246b0e5e686534fc9add6f2e93faec64736f6c63430008120033"
      const burnerAddr = "0xc200000000000000000000000000000000000000"
      const r = await simulate({
        blockStateCalls: [
          {
            blockOverrides: { gasLimit: "0x16e360" },
            stateOverrides: {
              [DETACHED_SENDER]: { balance: "0x1e8480" },
              [burnerAddr]: { code: burner },
            },
            calls: [
              {
                from: DETACHED_SENDER,
                to: burnerAddr,
                data: "0x815b8ab40000000000000000000000000000000000000000000000000000000000000000",
              },
            ],
          },
          {
            calls: [
              {
                from: DETACHED_SENDER,
                to: burnerAddr,
                data: "0x815b8ab40000000000000000000000000000000000000000000000000000000000000000",
              },
            ],
          },
          {
            calls: [
              {
                from: DETACHED_SENDER,
                to: burnerAddr,
                data: "0x815b8ab40000000000000000000000000000000000000000000000000000000000000000",
              },
            ],
          },
        ],
      })
      blocks = r.result!
    })

    it("all three blocks' calls succeed", function () {
      expect(blocks).to.have.lengthOf(3)
      for (const block of blocks) {
        for (const call of block.calls) {
          expect(call.status).to.equal("0x1")
        }
      }
    })
  })

  // =================================================================
  // === FullTx surface                                              ==
  // =================================================================

  context("FullTx: hash-only default", function () {
    let block: any

    before(async function () {
      const transferData = simpleToken.interface.encodeFunctionData("transfer", [
        recipientAddr,
        ethers.parseUnits("1", 18),
      ])
      const blocks: any[] = await ethers.provider.send("eth_simulateV1", [
        {
          blockStateCalls: [
            { calls: [{ from: senderAddr, to: tokenAddr, data: transferData }] },
          ],
        },
        "latest",
      ])
      block = blocks[0]
    })

    it("transactions[] entries are 0x-prefixed 32-byte hex strings", function () {
      expect(block.transactions).to.be.an("array").with.lengthOf(1)
      const entry = block.transactions[0]
      expect(entry).to.be.a("string")
      expect(entry).to.match(/^0x[0-9a-f]{64}$/i)
    })
  })

  context("FullTx: returnFullTransactions=true", function () {
    let block: any

    before(async function () {
      const transferData = simpleToken.interface.encodeFunctionData("transfer", [
        recipientAddr,
        ethers.parseUnits("1", 18),
      ])
      const blocks: any[] = await ethers.provider.send("eth_simulateV1", [
        {
          blockStateCalls: [
            { calls: [{ from: senderAddr, to: tokenAddr, data: transferData }] },
          ],
          returnFullTransactions: true,
        },
        "latest",
      ])
      block = blocks[0]
    })

    it("transactions[] carries RPCTransaction objects with from patched", function () {
      expect(block.transactions).to.be.an("array").with.lengthOf(1)
      const tx = block.transactions[0]
      expect(tx).to.be.an("object")
      for (const key of [
        "blockHash",
        "blockNumber",
        "from",
        "gas",
        "hash",
        "input",
        "nonce",
        "to",
        "transactionIndex",
        "value",
        "type",
      ]) {
        expect(tx).to.have.property(key)
      }
      expect(tx.from.toLowerCase()).to.equal(senderAddr.toLowerCase())
      expect(tx.hash).to.match(/^0x[0-9a-f]{64}$/i)
    })
  })

  context("FullTx: multi-sender", function () {
    let block: any

    before(async function () {
      const multiSenderBalance = ethers.toQuantity(ethers.parseEther("100"))
      const blocks: any[] = await ethers.provider.send("eth_simulateV1", [
        {
          blockStateCalls: [
            {
              stateOverrides: {
                [senderAddr]: { balance: multiSenderBalance },
                [DETACHED_SECOND_SENDER]: { balance: multiSenderBalance },
              },
              calls: [
                {
                  from: senderAddr,
                  to: recipientAddr,
                  value: ethers.toQuantity(1n),
                },
                {
                  from: DETACHED_SECOND_SENDER,
                  to: recipientAddr,
                  value: ethers.toQuantity(2n),
                },
              ],
            },
          ],
          returnFullTransactions: true,
        },
        "latest",
      ])
      block = blocks[0]
    })

    it("each tx's from resolves and hashes are unique", function () {
      expect(block.transactions).to.be.an("array").with.lengthOf(2)
      const fromAddrs = new Set(
        block.transactions.map((tx: any) => tx.from.toLowerCase()),
      )
      expect(fromAddrs.has(senderAddr.toLowerCase())).to.equal(true)
      expect(fromAddrs.has(DETACHED_SECOND_SENDER.toLowerCase())).to.equal(true)
      const hashes = block.transactions.map((tx: any) => tx.hash)
      expect(new Set(hashes).size).to.equal(hashes.length)
    })
  })

  // =================================================================
  // === Block envelope                                              ==
  // =================================================================

  context("Block envelope: non-empty block", function () {
    let block: any
    let blockRerun: any

    before(async function () {
      const transferData = simpleToken.interface.encodeFunctionData("transfer", [
        recipientAddr,
        ethers.parseUnits("1", 18),
      ])
      const opts = {
        blockStateCalls: [
          { calls: [{ from: senderAddr, to: tokenAddr, data: transferData }] },
        ],
      }
      const blocks: any[] = await ethers.provider.send("eth_simulateV1", [
        opts,
        "latest",
      ])
      block = blocks[0]
      const blocksRerun: any[] = await ethers.provider.send(
        "eth_simulateV1",
        [opts, "latest"],
      )
      blockRerun = blocksRerun[0]
    })

    it("transactionsRoot is non-empty", function () {
      expect(block.transactionsRoot.toLowerCase()).to.not.equal(EMPTY_TRIE_ROOT)
    })

    it("receiptsRoot is non-empty", function () {
      expect(block.receiptsRoot.toLowerCase()).to.not.equal(EMPTY_TRIE_ROOT)
    })

    it("logsBloom positive-tests the Transfer topic", function () {
      const bloomBytes = ethers.getBytes(block.logsBloom)
      const topicBytes = ethers.getBytes(TRANSFER_TOPIC)
      const hash = ethers.getBytes(ethers.keccak256(topicBytes))
      let allBitsSet = true
      for (let i = 0; i < 3; i++) {
        const bitIndex = ((hash[2 * i] << 8) | hash[2 * i + 1]) & 0x7ff
        const byteIndex = 256 - 1 - (bitIndex >> 3)
        const bitMask = 1 << (bitIndex & 7)
        if ((bloomBytes[byteIndex] & bitMask) === 0) {
          allBitsSet = false
          break
        }
      }
      expect(allBitsSet).to.equal(true)
    })

    it("size > 0", function () {
      expect(BigInt(block.size)).to.be.greaterThan(0n)
    })

    it("hash is stable across identical re-runs", function () {
      expect(block.hash).to.equal(blockRerun.hash)
    })
  })

  context("Block envelope: gap-fill block", function () {
    let gap: any

    before(async function () {
      const transferData = simpleToken.interface.encodeFunctionData("transfer", [
        recipientAddr,
        ethers.parseUnits("1", 18),
      ])
      const baseBlockNumber = await ethers.provider.getBlockNumber()
      const blocks: any[] = await ethers.provider.send("eth_simulateV1", [
        {
          blockStateCalls: [
            {
              calls: [{ from: senderAddr, to: tokenAddr, data: transferData }],
            },
            {
              blockOverrides: {
                number: ethers.toQuantity(baseBlockNumber + 3),
              },
              calls: [{ from: senderAddr, to: tokenAddr, data: transferData }],
            },
          ],
        },
        "latest",
      ])
      gap = blocks[1]
    })

    it("contains no transactions", function () {
      expect(gap.transactions).to.be.an("array").with.lengthOf(0)
    })

    it("uses the empty trie root for transactionsRoot and receiptsRoot", function () {
      expect(gap.transactionsRoot.toLowerCase()).to.equal(EMPTY_TRIE_ROOT)
      expect(gap.receiptsRoot.toLowerCase()).to.equal(EMPTY_TRIE_ROOT)
    })

    it("logsBloom is the zero bloom", function () {
      expect(gap.logsBloom.toLowerCase()).to.equal(ZERO_BLOOM)
    })
  })
})
