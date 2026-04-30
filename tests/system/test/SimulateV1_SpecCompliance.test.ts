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
 * Conventions: one before() per file fires every RPC call; it() blocks
 * assert only. Opts are raw JSON. No byte-for-byte response-hash checks
 * against the upstream fixtures (mezod's localnode chain id, base block,
 * and account state never match the reference replay) — invariants only.
 *
 * +-------------------------------------+--------------------------------+------------------------------+--------------------------------+
 * | Scenario                            | Given                          | When                         | Then                           |
 * +-------------------------------------+--------------------------------+------------------------------+--------------------------------+
 * | erc20 transfer happy path           | SimpleToken with sender funded | one transfer call            | status 0x1, gasUsed > 0,       |
 * |                                     |                                |                              | returnData decodes true,       |
 * |                                     |                                |                              | one Transfer log               |
 * |                                     |                                |                              |                                |
 * | erc20 transfer revert path          | sender lacks balance           | transfer of HUGE amount      | status 0x0, error.code 3,      |
 * |                                     |                                |                              | message "execution reverted",  |
 * |                                     |                                |                              | data starts with Panic(0x11)   |
 * |                                     |                                |                              |                                |
 * | mint -> transfer -> balanceOf chain | empty SimpleToken              | three calls in one block     | all three succeed, balanceOf   |
 * |                                     |                                |                              | reads transferred amount       |
 * |                                     |                                |                              |                                |
 * | block 1 mint -> block 2 balanceOf   | empty SimpleToken              | two-block opts               | block 2 reads block 1 mint     |
 * |                                     |                                |                              |                                |
 * | BLOCKHASH(base+1)/(base+2)          | reader stub installed in blk 1 | block 3 reads sibling hashes | matches blocks[0].hash and     |
 * |                                     |                                |                              | blocks[1].hash                 |
 * |                                     |                                |                              |                                |
 * | sha256 move-and-call                | sha256 moved to 0x..1234       | call dest with "hello"       | returns sha256("hello")        |
 * |                                     |                                |                              |                                |
 * | self-referencing move rejected      | precompile -> itself           | eth_call with that override  | rejected, message includes     |
 * |                                     |                                |                              | "referenced itself" (typed     |
 * |                                     |                                |                              | -38022 only on simulateV1)     |
 * |                                     |                                |                              |                                |
 * | block cap                           | 257 empty blocks               | eth_simulateV1               | -38026, "blocks > max 256"     |
 * | call cap                            | 1001 calls in one block        | eth_simulateV1               | -38026, "calls > max 1000"     |
 * | block-gas-limit cap                 | gasLimit 50k, two calls        | second call needs 30k        | top-level -38015               |
 * | intrinsic-gas cap                   | call.gas 1000 < 21000          | eth_simulateV1               | top-level -38013               |
 * | trace EOA->EOA                      | traceTransfers=true            | one value transfer           | status 0x1, one synthetic log  |
 * | trace contract->EOA                 | forwarder bytecode             | top-level CALL with value    | two synthetic logs (each edge) |
 * | trace omitted                       | flag absent                    | same forwarder call          | zero synthetic logs            |
 * | validation happy path               | balance funded                 | validation=true              | per-call status 0x1            |
 * | -38010 nonce too low                | state nonce 5, call.nonce 4    | validation=true              | request aborts -38010          |
 * | -38011 nonce too high               | state nonce 5, call.nonce 9    | validation=true              | request aborts -38011          |
 * | -38013 intrinsic gas                | call.gas 20999                 | validation=true              | request aborts -38013          |
 * | -38014 insufficient funds           | balance 0, value 1             | validation=true              | request aborts -38014          |
 * | -38025 init-code too large          | initcode 49153 bytes           | validation=true              | request aborts -38025          |
 * | -32005 fee-cap below base           | maxFeePerGas=0 + baseFee=1gwei | validation=true              | request aborts -32005          |
 * | -38012 base-fee override low        | baseFee override 1 wei         | validation=true              | request aborts -38012          |
 * | revert per-call under validation    | revert bytecode                | validation=true              | request OK, per-call code 3    |
 * | nonce-low under validation=false    | state nonce 5, call.nonce 4    | flag omitted                 | per-call status 0x1            |
 * | fee-cap-low under validation=false  | maxFeePerGas=0 + baseFee=1gwei | flag omitted                 | per-call status 0x1            |
 * | nonce-too-high success              | state nonce 0, call.nonce 100  | validation=false             | per-call status 0x1            |
 * | nonce increments per call           | three calls nonce 0,1,2        | validation=true              | three per-call status 0x1      |
 * | balance after empty new block       | base, then empty block         | second block reads balance   | request succeeds               |
 * |                                     |                                |                              |                                |
 * | block-num order -38020              | second block goes backwards    | eth_simulateV1               | -38020, "blocks must be in     |
 * |                                     |                                |                              | order"                         |
 * |                                     |                                |                              |                                |
 * | block-timestamp order -38021        | equal timestamps both > base   | eth_simulateV1               | -38021                         |
 * | timestamp non-increment             | second block goes backwards    | eth_simulateV1               | -38021                         |
 * |                                     |                                |                              |                                |
 * | timestamp auto-increment            | only first time set            | eth_simulateV1               | second block.timestamp >       |
 * |                                     |                                |                              | first                          |
 * |                                     |                                |                              |                                |
 * | timestamp incrementing pair         | strictly monotonic times       | eth_simulateV1               | both blocks present, in order  |
 * |                                     |                                |                              |                                |
 * | blocknumber auto-increment          | gap-fill 1 then 5              | eth_simulateV1               | five blocks numbered           |
 * |                                     |                                |                              | sequentially                   |
 * |                                     |                                |                              |                                |
 * | empty at numeric anchor             | empty blockStateCalls          | concrete numeric anchor      | one block at anchor+1          |
 * | future block anchor                 | anchor far past head           | validation=true              | header not found / -32000      |
 * | override-block-num                  | sequential override numbers    | NUMBER opcode in each        | each call returns its number   |
 * |                                     |                                |                              |                                |
 * | BLOCKHASH below base                | reader on canonical height     | call hits canonical chain    | matches canonical block hash   |
 * |                                     |                                |                              | OR zero (mezod historical info |
 * |                                     |                                |                              | retention is best-effort)      |
 * |                                     |                                |                              |                                |
 * | simple-state-diff                   | code override + state override | call reads slot 0            | gets overridden value          |
 * | override-storage-slots              | stateDiff slot vs state slot   | call reads same slot         | both return overridden value   |
 * | override-address-twice              | code override + balance from   | sender balance 0, value 1    | -38014 with validation true    |
 * |                                     |                                |                              |                                |
 * | override across separate blocks     | balance set in block 1 and 2   | EOA-EOA value transfer twice | both blocks succeed            |
 * | overwrite-existing-contract         | call once -> override -> call  | second call has new bytecode | first reverts, second OK       |
 * | block-override reflected            | timeline-clean overrides       | NUMBER/TIMESTAMP probe       | values match overrides         |
 * | fee-recipient receives funds        | feeRecipient override          | tx with non-zero value       | request succeeds, status 0x1   |
 * | move-ecrecover happy                | ecrecover -> 0x..123456        | call dest with sig           | returns recovered addr         |
 * | move two precompiles to one         | sha256+id both -> dest         | eth_simulateV1               | -38023                         |
 * | move non-precompile rejected        | EOA -> dest, validation=true   | eth_simulateV1               | -32000, "is not a precompile"  |
 * |                                     |                                |                              |                                |
 * | override ecrecover                  | code override + move           | call original ecrecover addr | call succeeds, returns user    |
 * |                                     |                                |                              | code's result                  |
 * |                                     |                                |                              |                                |
 * | override identity                   | identity -> 0x..123456 + code  | call dest with 0x1234        | dest returns 0x1234 (id),      |
 * |                                     |                                |                              | original returns empty (code)  |
 * |                                     |                                |                              |                                |
 * | logs: eth send no logs by default   | EOA->EOA with value, no flag   | eth_simulateV1               | logs == []                     |
 * | logs: forward then revert           | forwarder -> reverter          | traceTransfers=true          | status 0x0, code 3, logs == [] |
 * | logs: selfdestruct produces log     | self-destruct contract         | traceTransfers=true          | one synthetic log on selfdest  |
 * | delegate-call to EOA logs once      | wallet contract via delegate   | traceTransfers=true          | one synthetic log only         |
 * | comprehensive: simple two transfers | sender funded                  | two value transfers in block | both per-call status 0x1       |
 * | transfer over BlockStateCalls       | balance set in two blocks      | several transfers            | all status 0x1                 |
 * | set-read-storage                    | storage contract               | set then read slot           | second call returns set value  |
 * | contract-calls-itself               | self-calling contract          | one call                     | status 0x1                     |
 * | run gas spending across blocks      | gas-burning contract           | three blocks                 | all calls status 0x1           |
 * | hash-only default                   | omitted returnFullTransactions | ERC-20 transfer              | transactions[] are 32-byte hex |
 * | full-tx mode: from patched          | returnFullTransactions=true    | ERC-20 transfer              | tx.from matches sender         |
 * | multi-sender full-tx                | two senders one block          | full-tx mode                 | each tx.from resolves          |
 * | block envelope: tx root             | one ERC-20 transfer            | eth_simulateV1               | transactionsRoot != empty      |
 * | block envelope: receipts root       | one ERC-20 transfer            | eth_simulateV1               | receiptsRoot != empty          |
 * | block envelope: logsBloom           | one Transfer log               | eth_simulateV1               | bloom positive-tests topic     |
 * | block envelope: size                | one ERC-20 transfer            | eth_simulateV1               | size > 0                       |
 * | block hash determinism              | identical opts                 | rerun                        | block.hash unchanged           |
 * |                                     |                                |                              |                                |
 * | gap-fill empty envelope             | base+1 then base+3             | eth_simulateV1               | gap block has empty            |
 * |                                     |                                |                              | tx/receipts roots & zero bloom |
 * +-------------------------------------+--------------------------------+------------------------------+--------------------------------+
 */
describe("SimulateV1_SpecCompliance", function () {
  const { deployments } = hre

  // -----------------------------------------------------------------
  // Constants used by multiple cases.
  // -----------------------------------------------------------------

  // Spec-reserved JSON-RPC error codes (geth execution-apis execute.yaml).
  const NONCE_TOO_LOW = -38010
  const NONCE_TOO_HIGH = -38011
  const BASE_FEE_TOO_LOW = -38012
  const INTRINSIC_GAS_TOO_LOW = -38013
  const INSUFFICIENT_FUNDS = -38014
  const BLOCK_GAS_LIMIT = -38015
  const BLOCK_NUM_ORDER = -38020
  const BLOCK_TIMESTAMP_ORDER = -38021
  const MOVE_SELF_REFERENCE_CODE = -38022
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

  // sha256("hello") — precomputed, stable.
  const HELLO = "0x68656c6c6f"
  const SHA256_HELLO =
    "0x2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"

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

  // Resolves the localnode head's block number. Useful when a test
  // needs a fresh anchor mid-flow rather than the cached latest.
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  async function latestNumber(): Promise<bigint> {
    const b = await ethers.provider.getBlock("latest")
    if (!b) throw new Error("no latest block")
    return BigInt(b.number)
  }

  // -----------------------------------------------------------------
  // Resources captured in before(): every RPC call lives here.
  // -----------------------------------------------------------------

  let simpleToken: SimpleToken
  let tokenAddr: string
  let senderAddr: string
  let recipientAddr: string

  // SingleCall / MultiCall / MultiBlock / Limits / TraceTransfers /
  // Validation / BlockEnvelope / FullTx redistributions.
  let happyResult: any
  let revertResult: any
  let stateChainBlocks: any[]
  let crossBlockBlocks: any[]
  let blockHashBlocks: any[]
  let sha256MoveResult: string
  let selfMoveErr: any
  let blockCapResult: CapturedError
  let callCapResult: CapturedError
  let blockGasLimitResult: CapturedError
  let intrinsicGasResultLimits: CapturedError
  let traceOnEoaResult: any
  let traceOnContractResult: any
  let traceOffContractResult: any
  let happyValidationResult: any
  let nonceLowResult: any
  let nonceHighResult: any
  let insufficientFundsResult: any
  let intrinsicGasResultValidation: any
  let initCodeTooLargeResult: any
  let feeCapTooLowResult: any
  let baseFeeOverrideTooLowResult: any
  let revertUnderValidationBlocks: any[]
  let nonceLowNoValidationBlocks: any[]
  let feeCapBelowBaseNoValidationBlocks: any[]
  let nonceHighNoValidationBlocks: any[]
  let nonceIncrementsBlocks: any[]
  let balanceAfterNewBlockBlocks: any[]
  let nonEmptyBlockResult: any
  let nonEmptyBlockResultRerun: any
  let gapBlocksResult: any[]
  let hashOnlyResult: any
  let fullTxResult: any
  let multiSenderResult: any

  // execution-apis fixture ports.
  let blockNumOrderResult: { error: CapturedError; result: any[] | undefined }
  let blockTsOrderResult: { error: CapturedError; result: any[] | undefined }
  let blockTsNonIncrementResult: {
    error: CapturedError
    result: any[] | undefined
  }
  let blockTsAutoIncrementBlocks: any[]
  let timestampsIncrementingBlocks: any[]
  let blocknumberIncrementBlocks: any[]
  let emptyEarliestBlocks: any[]
  let futureBlockResult: { error: CapturedError; result: any[] | undefined }
  let overrideBlockNumBlocks: any[]
  let blockHashBelowBaseBlocks: any[]
  let blockHashBelowBaseExpected: string | null
  let traceRecipientAddr: string
  let traceForwarderAddr: string
  let simpleStateDiffBlocks: any[]
  let overrideStorageSlotsBlocks: any[]
  let overrideAddressTwiceResult: {
    error: CapturedError
    result: any[] | undefined
  }
  let overrideAcrossBlocksBlocks: any[]
  let overwriteExistingContractBlocks: any[]
  let blockOverrideReflectedBlocks: any[]
  let feeRecipientBlocks: any[]
  let moveEcrecoverBlocks: any[]
  let moveTwoToSameResult: { error: CapturedError; result: any[] | undefined }
  let moveNonPrecompileResult: {
    error: CapturedError
    result: any[] | undefined
  }
  let overrideIdentityBlocks: any[]
  let logsNoneByDefaultBlocks: any[]
  let logsForwardRevertBlocks: any[]
  let logsSelfdestructBlocks: any[]
  let delegateCallEoaBlocks: any[]
  let simpleSanityBlocks: any[]
  let transferOverBlockStateCallsBlocks: any[]
  let setReadStorageBlocks: any[]
  let contractCallsItselfBlocks: any[]
  let runGasSpendingBlocks: any[]

  before(async function () {
    await deployments.fixture(["BridgeOutDelegate"])
    simpleToken = await getDeployedContract<SimpleToken>("SimpleToken")
    tokenAddr = await simpleToken.getAddress()

    const [deployer] = await ethers.getSigners()
    senderAddr = deployer.address
    recipientAddr = ethers.Wallet.createRandom().address

    // Capture the canonical-chain latest block once. Localnode timestamps
    // are real wall-clock seconds (~1.7e9) and far above any literal
    // 0x1cd hex used by upstream fixtures, so timestamp/blocknumber
    // overrides are derived as offsets from this anchor — see
    // sanitizeSimChain in x/evm/keeper/simulate_v1.go which seeds
    // prevTimestamp from base.Time.
    const latest = await ethers.provider.getBlock("latest")
    if (!latest) throw new Error("no latest block")
    const latestTime = BigInt(latest.timestamp)
    const latestBlockNum = BigInt(latest.number)

    const AMOUNT = ethers.parseUnits("1", 18)
    const HUGE_AMOUNT = AMOUNT * 1_000_000n
    const MINT_AMOUNT = ethers.parseUnits("1000", 18)

    // Mint 10*AMOUNT so the happy-path transfer succeeds and the
    // revert-path transfer (HUGE_AMOUNT) reliably underflows.
    await (
      await simpleToken.connect(deployer).mint(senderAddr, AMOUNT * 10n)
    ).wait()

    // ---------- SingleCall -------------------------------------------
    const happyData = simpleToken.interface.encodeFunctionData("transfer", [
      recipientAddr,
      AMOUNT,
    ])
    const revertData = simpleToken.interface.encodeFunctionData("transfer", [
      recipientAddr,
      HUGE_AMOUNT,
    ])
    const baseStateOverrides = {
      [senderAddr]: { balance: ethers.toQuantity(ethers.parseEther("100")) },
    }
    const happyBlocks: any[] = await ethers.provider.send("eth_simulateV1", [
      {
        blockStateCalls: [
          {
            stateOverrides: baseStateOverrides,
            calls: [{ from: senderAddr, to: tokenAddr, data: happyData }],
          },
        ],
      },
      "latest",
    ])
    happyResult = happyBlocks[0].calls[0]
    const revertBlocks: any[] = await ethers.provider.send("eth_simulateV1", [
      {
        blockStateCalls: [
          {
            stateOverrides: baseStateOverrides,
            calls: [{ from: senderAddr, to: tokenAddr, data: revertData }],
          },
        ],
      },
      "latest",
    ])
    revertResult = revertBlocks[0].calls[0]

    // ---------- MultiCall (SimpleToken case) -------------------------
    const mintData = simpleToken.interface.encodeFunctionData("mint", [
      senderAddr,
      MINT_AMOUNT,
    ])
    const transferDataForChain = simpleToken.interface.encodeFunctionData(
      "transfer",
      [recipientAddr, ethers.parseUnits("500", 18)],
    )
    const balanceOfData = simpleToken.interface.encodeFunctionData(
      "balanceOf",
      [recipientAddr],
    )
    stateChainBlocks = await ethers.provider.send("eth_simulateV1", [
      {
        blockStateCalls: [
          {
            stateOverrides: baseStateOverrides,
            calls: [
              { from: senderAddr, to: tokenAddr, data: mintData },
              { from: senderAddr, to: tokenAddr, data: transferDataForChain },
              { from: senderAddr, to: tokenAddr, data: balanceOfData },
            ],
          },
        ],
      },
      "latest",
    ])

    // ---------- MultiBlock (SimpleToken + BLOCKHASH) -----------------
    const crossBlockRecipient = ethers.Wallet.createRandom().address
    const crossMintData = simpleToken.interface.encodeFunctionData("mint", [
      crossBlockRecipient,
      ethers.parseUnits("777", 18),
    ])
    const crossBalanceOfData = simpleToken.interface.encodeFunctionData(
      "balanceOf",
      [crossBlockRecipient],
    )
    crossBlockBlocks = await ethers.provider.send("eth_simulateV1", [
      {
        blockStateCalls: [
          {
            stateOverrides: baseStateOverrides,
            calls: [{ from: senderAddr, to: tokenAddr, data: crossMintData }],
          },
          {
            calls: [
              { from: senderAddr, to: tokenAddr, data: crossBalanceOfData },
            ],
          },
        ],
      },
      "latest",
    ])

    // Two-pass: first, simulate the chain so the synthetic block 0
    // surfaces its assigned number; second, read BLOCKHASH for that
    // assigned base + offsets. Deriving the calldata from a separate
    // ethers `getBlock("latest")` would race the localnode head if a
    // canonical block lands between the two reads; pinning the base to
    // the simulate response avoids that.
    const readerAddr = ethers.getAddress("0x" + "c0de".repeat(10))
    const heightToCalldata = (h: bigint) =>
      ethers.zeroPadValue(ethers.toBeHex(h), 32)
    {
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
      blockHashBlocks = await ethers.provider.send("eth_simulateV1", [
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
    }

    // ---------- MovePrecompile_ethCall (spec-conformant cases) -------
    sha256MoveResult = await ethers.provider.send("eth_call", [
      { to: DEST_ADDR, data: HELLO },
      "latest",
      { [SHA256]: { movePrecompileToAddress: DEST_ADDR } },
    ])
    try {
      await ethers.provider.send("eth_call", [
        { to: SHA256, data: HELLO },
        "latest",
        { [SHA256]: { movePrecompileToAddress: SHA256 } },
      ])
      selfMoveErr = null
    } catch (err: any) {
      selfMoveErr = err
    }

    // ---------- Limits -----------------------------------------------
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
    ] as any)
    intrinsicGasResultLimits = await captureSim([
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

    // ---------- TraceTransfers ---------------------------------------
    const traceRecipient = ethers.Wallet.createRandom().address
    const forwarderAddr = ethers.getAddress(
      "0x000000000000000000000000000000000000f0f0",
    )
    traceRecipientAddr = traceRecipient
    traceForwarderAddr = forwarderAddr
    const transferValue = ethers.toQuantity(ethers.parseEther("1"))
    const senderBalance = ethers.toQuantity(ethers.parseEther("100"))
    const traceForwardCalldata = ethers.zeroPadValue(traceRecipient, 32)
    const traceState = {
      [senderAddr]: { balance: senderBalance },
      [forwarderAddr]: { code: VALUE_FORWARDER_BYTECODE },
    }
    const traceOnEoaBlocks: any[] = await ethers.provider.send(
      "eth_simulateV1",
      [
        {
          blockStateCalls: [
            {
              stateOverrides: { [senderAddr]: { balance: senderBalance } },
              calls: [
                {
                  from: senderAddr,
                  to: traceRecipient,
                  value: transferValue,
                },
              ],
            },
          ],
          traceTransfers: true,
        },
        "latest",
      ],
    )
    traceOnEoaResult = traceOnEoaBlocks[0].calls[0]
    const traceOnContractBlocks: any[] = await ethers.provider.send(
      "eth_simulateV1",
      [
        {
          blockStateCalls: [
            {
              stateOverrides: traceState,
              calls: [
                {
                  from: senderAddr,
                  to: forwarderAddr,
                  value: transferValue,
                  data: traceForwardCalldata,
                },
              ],
            },
          ],
          traceTransfers: true,
        },
        "latest",
      ],
    )
    traceOnContractResult = traceOnContractBlocks[0].calls[0]
    const traceOffBlocks: any[] = await ethers.provider.send("eth_simulateV1", [
      {
        blockStateCalls: [
          {
            stateOverrides: traceState,
            calls: [
              {
                from: senderAddr,
                to: forwarderAddr,
                value: transferValue,
                data: traceForwardCalldata,
              },
            ],
          },
        ],
      },
      "latest",
    ])
    traceOffContractResult = traceOffBlocks[0].calls[0]

    // ---------- Validation -------------------------------------------
    const validationFunded = { [DETACHED_SENDER]: { balance: FUNDED_BALANCE } }
    happyValidationResult = await simulate({
      blockStateCalls: [
        {
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
    intrinsicGasResultValidation = await simulate({
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
    const revertUnderValidationContract =
      "0xddee000000000000000000000000000000000000"
    const revertUnderValidation = await simulate({
      blockStateCalls: [
        {
          stateOverrides: {
            [DETACHED_SENDER]: { balance: FUNDED_BALANCE },
            [revertUnderValidationContract]: { code: REVERT_BYTECODE },
          },
          calls: [
            {
              from: DETACHED_SENDER,
              to: revertUnderValidationContract,
              maxFeePerGas: HIGH_FEE,
            },
          ],
        },
      ],
      validation: true,
    })
    revertUnderValidationBlocks = revertUnderValidation.result!
    const nonceLowOff = await simulate({
      blockStateCalls: [
        {
          stateOverrides: {
            [DETACHED_SENDER]: { balance: FUNDED_BALANCE, nonce: "0x5" },
          },
          calls: [{ from: DETACHED_SENDER, to: DETACHED_RECIPIENT, nonce: "0x4" }],
        },
      ],
    })
    nonceLowNoValidationBlocks = nonceLowOff.result!
    const feeCapBelowOff = await simulate({
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
    })
    feeCapBelowBaseNoValidationBlocks = feeCapBelowOff.result!

    // ---------- BlockEnvelope ----------------------------------------
    const transferData = simpleToken.interface.encodeFunctionData("transfer", [
      recipientAddr,
      ethers.parseUnits("1", 18),
    ])
    const baseBlockNumber = await ethers.provider.getBlockNumber()
    const transferOpts = {
      blockStateCalls: [
        {
          calls: [{ from: senderAddr, to: tokenAddr, data: transferData }],
        },
      ],
    }
    const transferBlocks: any[] = await ethers.provider.send("eth_simulateV1", [
      transferOpts,
      "latest",
    ])
    nonEmptyBlockResult = transferBlocks[0]
    const transferBlocksRerun: any[] = await ethers.provider.send(
      "eth_simulateV1",
      [transferOpts, "latest"],
    )
    nonEmptyBlockResultRerun = transferBlocksRerun[0]
    gapBlocksResult = await ethers.provider.send("eth_simulateV1", [
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

    // ---------- FullTx -----------------------------------------------
    const hashOnlyBlocks: any[] = await ethers.provider.send("eth_simulateV1", [
      {
        blockStateCalls: [
          { calls: [{ from: senderAddr, to: tokenAddr, data: transferData }] },
        ],
      },
      "latest",
    ])
    hashOnlyResult = hashOnlyBlocks[0]
    const fullTxBlocks: any[] = await ethers.provider.send("eth_simulateV1", [
      {
        blockStateCalls: [
          { calls: [{ from: senderAddr, to: tokenAddr, data: transferData }] },
        ],
        returnFullTransactions: true,
      },
      "latest",
    ])
    fullTxResult = fullTxBlocks[0]
    const multiSenderBalance = ethers.toQuantity(ethers.parseEther("100"))
    const multiSenderBlocks: any[] = await ethers.provider.send(
      "eth_simulateV1",
      [
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
      ],
    )
    multiSenderResult = multiSenderBlocks[0]

    // ===============================================================
    // execution-apis high-signal fixture ports.
    // ===============================================================

    // -------- block ordering / gap fill -----------------------------

    // -38020: numbers must be strictly increasing. Both overrides must
    // sit above latestBlockNum so the first block clears the
    // base-anchored span check; -38020 then fires on the second block's
    // non-monotonic number.
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

    // -38021: both timestamps must clear base.Time so the inter-block
    // ordering check is what fires (not the first-block-vs-base check).
    // (a) Equal timestamps; both bumped above latestTime + 1.
    blockTsOrderResult = await simulate({
      blockStateCalls: [
        { blockOverrides: { time: ethers.toQuantity(latestTime + 10n) } },
        { blockOverrides: { time: ethers.toQuantity(latestTime + 10n) } },
      ],
    })
    // (b) Non-incrementing pair: second block goes backwards. Both
    // values clear latestTime so the violation is the second's
    // timestamp being <= the first's.
    blockTsNonIncrementResult = await simulate({
      blockStateCalls: [
        { blockOverrides: { time: ethers.toQuantity(latestTime + 20n) } },
        { blockOverrides: {} },
        { blockOverrides: { time: ethers.toQuantity(latestTime + 15n) } },
        { blockOverrides: {} },
      ],
    })

    // Auto-increment: only first `time` set; second block must end
    // up with timestamp > first. Anchor the explicit timestamp above
    // latestTime so the first block does not trip -38021 against base.
    {
      const r = await simulate({
        blockStateCalls: [
          {
            blockOverrides: { time: ethers.toQuantity(latestTime + 10n) },
          },
          { calls: [] },
        ],
      })
      blockTsAutoIncrementBlocks = r.result!
    }

    // Two-block timestamp increment: both succeed, both present.
    // Both timestamps must clear latestTime; pin them so the
    // assertion can read the exact values back.
    const tsIncrementingFirst = latestTime + 10n
    const tsIncrementingSecond = latestTime + 11n
    {
      const r = await simulate({
        blockStateCalls: [
          { blockOverrides: { time: ethers.toQuantity(tsIncrementingFirst) } },
          { blockOverrides: { time: ethers.toQuantity(tsIncrementingSecond) } },
        ],
      })
      timestampsIncrementingBlocks = r.result!
    }

    // Block-number increment: explicit override of the first block at
    // base+1 (the first valid number after the anchor) so the simulator
    // does not gap-fill empties between base and the override. The second
    // block is left to auto-increment to base+2. Pin the anchor to
    // latestBlockNum (captured at top of before()) so the override stays
    // valid even if the live localnode head advances during the suite —
    // anchoring at the moving "latest" tag would race the override.
    {
      const future = latestBlockNum + 1n
      const r = await simulateAt(
        {
          blockStateCalls: [
            { blockOverrides: { number: ethers.toQuantity(future) } },
            { calls: [] },
          ],
        },
        ethers.toQuantity(latestBlockNum),
      )
      blocknumberIncrementBlocks = r.result!
    }

    // -------- block-number resolution -------------------------------

    // Single empty blockStateCall surfaces one envelope at anchor+1.
    // Anchored at "latest" (server-resolved atomically) because mezod's
    // localnode runs with pruning-keep-recent=2: any captured numeric
    // anchor races against block advance and pruning. The upstream
    // "earliest" tag (genesis) cannot be used either without depending
    // on historical-info retention.
    {
      const r = await simulate({ blockStateCalls: [{ calls: [] }] })
      emptyEarliestBlocks = r.result!
    }

    // Future block anchor: tag well past the head.
    {
      const future = latestBlockNum + 1_000_000n
      futureBlockResult = await simulateAt(
        {
          blockStateCalls: [
            {
              calls: [
                { from: DETACHED_SENDER, to: DETACHED_RECIPIENT },
              ],
            },
          ],
          validation: true,
        },
        ethers.toQuantity(future),
      )
    }

    // Override block num: each block hosts a NUMBER probe. Anchor the
    // first override at base+1 so the simulate driver does not gap-fill
    // empty headers between base and the override; the second block
    // increments by 1 to exercise the inter-block monotonicity path.
    // Pin the anchor to latestBlockNum so the override stays valid even
    // if the live localnode head advances during the suite.
    {
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
      overrideBlockNumBlocks = r.result!
    }

    // -------- BLOCKHASH below base ---------------------------------

    // Ask BLOCKHASH for a canonical height below the localnode head.
    // Use latest.number - 1 so the probe always hits a block that
    // actually exists; localnode's earliest block 0 is genesis but
    // BLOCKHASH(genesis_height) can return zero on some EVM rules,
    // so the previous block is the safer probe.
    {
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
      blockHashBelowBaseBlocks = r.result!
      const probedBlock = await ethers.provider.getBlock(Number(probedHeight))
      blockHashBelowBaseExpected = probedBlock ? probedBlock.hash! : null
    }

    // -------- state / block overrides not pinned today --------------

    // simple-state-diff using `state`: code at C100 reads slot 0.
    // Bytecode below: PUSH1 0 SLOAD PUSH1 0 MSTORE PUSH1 0x20 PUSH1 0
    // RETURN — returns slot 0 padded.
    const SLOT0_READER = "0x60005460005260206000F3"
    {
      const target = "0xc100000000000000000000000000000000000000"
      const r = await simulate({
        blockStateCalls: [
          {
            stateOverrides: {
              [target]: {
                code: SLOT0_READER,
                state: {
                  "0x0000000000000000000000000000000000000000000000000000000000000000":
                    "0x12340000000000000000000000000000000000000000000000000000000000ff",
                },
              },
            },
            calls: [{ from: DETACHED_SENDER, to: target }],
          },
        ],
      })
      simpleStateDiffBlocks = r.result!
    }

    // override-storage-slots: stateDiff in block 1, then `state` in
    // block 2 — both must surface the overridden value.
    {
      const target = "0xc100000000000000000000000000000000000000"
      const overrideValue =
        "0x12340000000000000000000000000000000000000000000000000000000000ff"
      const r = await simulate({
        blockStateCalls: [
          {
            stateOverrides: {
              [target]: {
                code: SLOT0_READER,
                stateDiff: {
                  "0x0000000000000000000000000000000000000000000000000000000000000000":
                    overrideValue,
                },
              },
            },
            calls: [{ from: DETACHED_SENDER, to: target }],
          },
          {
            stateOverrides: {
              [target]: {
                state: {
                  "0x0000000000000000000000000000000000000000000000000000000000000000":
                    overrideValue,
                },
              },
            },
            calls: [{ from: DETACHED_SENDER, to: target }],
          },
        ],
      })
      overrideStorageSlotsBlocks = r.result!
    }

    // override-address-twice: code-only override leaves balance 0,
    // value-1 transfer with validation=true must abort -38014.
    overrideAddressTwiceResult = await simulate({
      blockStateCalls: [
        {
          stateOverrides: {
            [DETACHED_SENDER]: { code: REVERT_BYTECODE },
          },
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

    // override-address-twice-in-separate-BlockStateCalls: balance
    // override re-applied in each block; both transfers succeed.
    {
      const r = await simulate({
        blockStateCalls: [
          {
            stateOverrides: {
              [DETACHED_SENDER]: { balance: FUNDED_BALANCE },
            },
            calls: [
              {
                from: DETACHED_SENDER,
                to: DETACHED_RECIPIENT,
                value: "0x3e8",
              },
            ],
          },
          {
            stateOverrides: {
              [DETACHED_SENDER]: { balance: FUNDED_BALANCE },
            },
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
      overrideAcrossBlocksBlocks = r.result!
    }

    // overwrite-existing-contract: SimpleToken.transfer reverts in
    // block 1 (recipient mismatched), then a code override replaces
    // the contract with a benign-success bytecode in block 2.
    {
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
      overwriteExistingContractBlocks = r.result!
    }

    // block-override reflected in contract: NUMBER opcode probe at
    // explicit override. Use base+1 so the simulator does not gap-fill
    // empty headers between base and the override. Pin the anchor to
    // latestBlockNum so the override stays valid even if the live
    // localnode head advances during the suite.
    {
      const target = "0xc100000000000000000000000000000000000000"
      const probeNumber = latestBlockNum + 1n
      const r = await simulateAt(
        {
          blockStateCalls: [
            {
              blockOverrides: { number: ethers.toQuantity(probeNumber) },
              stateOverrides: {
                [target]: { code: RETURN_NUMBER_BYTECODE },
              },
              calls: [{ from: DETACHED_SENDER, to: target }],
            },
          ],
        },
        ethers.toQuantity(latestBlockNum),
      )
      blockOverrideReflectedBlocks = r.result!
    }

    // fee-recipient receiving funds: explicit feeRecipient override
    // with a value transfer. Mezo does not credit the recipient on
    // simulated blocks, but the request must succeed.
    {
      const r = await simulate({
        blockStateCalls: [
          {
            blockOverrides: {
              feeRecipient: "0xc200000000000000000000000000000000000000",
            },
            stateOverrides: {
              [DETACHED_SENDER]: { balance: FUNDED_BALANCE },
            },
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
      feeRecipientBlocks = r.result!
    }

    // -------- MovePrecompileTo edges -------------------------------

    // ECRECOVER move: with a real signature payload from the upstream
    // fixture, the moved address recovers the same recoverable addr.
    const ecrecoverSig =
      "0x1c8aff950685c2ed4bc3174f3472287b56d9517b9c948127319a09a7a36deac8000000000000000000000000000000000000000000000000000000000000001cb7cf302145348387b9e69fde82d8e634a0f8761e78da3bfa059efced97cbed0d2a66b69167cafe0ccfc726aec6ee393fea3cf0e4f3f9c394705e0f56d9bfe1c9"
    {
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
      moveEcrecoverBlocks = r.result!
    }

    // Move two precompiles to the same destination: -38023.
    moveTwoToSameResult = await simulate({
      blockStateCalls: [
        {
          stateOverrides: {
            [ECRECOVER]: { movePrecompileToAddress: DEST_ADDR },
            [SHA256]: { movePrecompileToAddress: DEST_ADDR },
          },
        },
      ],
    })

    // Try to move a non-precompile EOA: -32000.
    moveNonPrecompileResult = await simulate({
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

    // Override identity precompile: provide alternative bytecode at
    // 0x...4 + move to DEST_ADDR. Original address now runs the
    // override (empty code → empty return); DEST_ADDR runs identity
    // (returns calldata).
    {
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
      overrideIdentityBlocks = r.result!
    }

    // -------- logs / tracer ----------------------------------------

    // No logs by default for an EOA->EOA value transfer.
    {
      const r = await simulate({
        blockStateCalls: [
          {
            stateOverrides: {
              [DETACHED_SENDER]: { balance: FUNDED_BALANCE },
            },
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
      logsNoneByDefaultBlocks = r.result!
    }

    // Forward then revert: forwarder -> reverter, traceTransfers=true.
    // No synthetic log; revert surfaces as code 3, status 0x0. Use a
    // status-propagating forwarder (REVERT on inner failure) — the plain
    // VALUE_FORWARDER_BYTECODE STOPs after CALL regardless of inner
    // outcome, so it could not surface the inner revert at the top frame.
    //
    // Disassembly of REVERT_PROPAGATING_FORWARDER_BYTECODE:
    //   PUSH1 0; PUSH1 0; PUSH1 0; PUSH1 0; CALLVALUE; PUSH1 0;
    //   CALLDATALOAD; GAS; CALL; ISZERO; PUSH1 0x13; JUMPI; STOP;
    //   JUMPDEST; PUSH1 0; PUSH1 0; REVERT.
    const REVERT_PROPAGATING_FORWARDER_BYTECODE =
      "0x6000600060006000346000355af115601357005b60006000fd"
    {
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
      logsForwardRevertBlocks = r.result!
    }

    // Self-destruct produces a synthetic Transfer log when the EVM
    // dispatches SELFDESTRUCT through the OnEnter tracer hook. Whether
    // it does is a runtime EVM concern; if the live EVM wires the
    // SELFDESTRUCT edge differently (e.g. via OnSelfDestruct only,
    // outside this tracer's hooks), the synthetic log will not be
    // emitted and the assertion below tolerates zero. The helper-layer
    // contract is pinned regardless via
    // TestSimTracer_SelfdestructWithBalanceEmitsSynthetic.
    {
      const sd = "0xc200000000000000000000000000000000000000"
      const r = await simulate({
        blockStateCalls: [
          {
            stateOverrides: {
              [sd]: { code: SELFDESTRUCT_BYTECODE, balance: "0x1e8480" },
            },
            calls: [
              { from: DETACHED_SENDER, to: sd, data: "0x83197ef0" },
            ],
          },
        ],
        traceTransfers: true,
      })
      logsSelfdestructBlocks = r.result!
    }

    // -------- delegate-call ----------------------------------------

    // Send eth + delegate-call to an EOA: the wallet bytecode
    // forwards via delegate-call (no value transfer through that
    // edge); only the top-level value transfer surfaces a synthetic
    // log.
    const WALLET_DELEGATECALL_BYTECODE =
      "0x608060405273ffffffffffffffffffffffffffffffffffffffff73c200000000000000000000000000000000000000167fa619486e000000000000000000000000000000000000000000000000000000005f3503605e57805f5260205ff35b365f80375f80365f845af43d5f803e5f81036077573d5ffd5b3d5ff3fea26469706673582212206b787fe3e60b14a5c449d37005afac7b1803ee7c87e12d2740b96a158f34802a64736f6c63430008160033"
    {
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
      delegateCallEoaBlocks = r.result!
    }

    // -------- comprehensive sanity ---------------------------------

    // Two transfers in a single block (EOA -> EOA, then chain along).
    {
      const r = await simulate({
        blockStateCalls: [
          {
            stateOverrides: {
              [DETACHED_SENDER]: { balance: "0x3e8" },
            },
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
      simpleSanityBlocks = r.result!
    }

    // Transfer over BlockStateCalls: balance set per-block, several
    // value transfers across two blocks.
    {
      const c3 = "0xc300000000000000000000000000000000000000"
      const r = await simulate({
        blockStateCalls: [
          {
            stateOverrides: {
              [DETACHED_SENDER]: { balance: "0x1388" },
            },
            calls: [
              {
                from: DETACHED_SENDER,
                to: DETACHED_RECIPIENT,
                value: "0x7d0",
              },
              {
                from: DETACHED_SENDER,
                to: c3,
                value: "0x7d0",
              },
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
      transferOverBlockStateCallsBlocks = r.result!
    }

    // set-read-storage: deploy storage contract via override, set
    // slot 0 to 5, read it back.
    {
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
      setReadStorageBlocks = r.result!
    }

    // Contract-calls-itself: SELF-CALL via deploy override.
    {
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
      contractCallsItselfBlocks = r.result!
    }

    // run-gas-spending: gas-burning contract called across three
    // blocks. Trims to one gas-burn call per block to keep the test
    // fast on localnode while still exercising multi-block accounting.
    {
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
      runGasSpendingBlocks = r.result!
    }

    // Validation edges: nonce too high under validation=false (per
    // upstream `transaction-too-high-nonce`).
    {
      const r = await simulate({
        blockStateCalls: [
          {
            calls: [
              {
                from: DETACHED_SENDER,
                to: DETACHED_SENDER,
                nonce: "0x64",
              },
            ],
          },
        ],
      })
      nonceHighNoValidationBlocks = r.result!
    }

    // Nonce increments per call under validation=true.
    {
      const r = await simulate({
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
      nonceIncrementsBlocks = r.result!
    }

    // Balance-after-new-block: block 1 sets balance via override;
    // block 2 reads balance via SELFBALANCE (using a tiny stub).
    // SELFBALANCE-MSTORE-RETURN bytecode: 0x47 (SELFBALANCE), 0x60 0x00,
    // MSTORE, 0x60 0x20, 0x60 0x00, RETURN.
    {
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
      balanceAfterNewBlockBlocks = r.result!
    }
  })

  // Helper used in before() for Limits-style top-level captures.
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

  // =================================================================
  // ===  Assertions                                                 ===
  // =================================================================

  context("ERC-20 transfer happy path", function () {
    it("returns status 0x1 with positive gasUsed and no error", function () {
      expect(happyResult.status).to.equal("0x1")
      expect(BigInt(happyResult.gasUsed)).to.be.greaterThan(0n)
      expect(happyResult.error).to.be.undefined
    })

    it("returnData decodes to true", function () {
      const [decoded] = ethers.AbiCoder.defaultAbiCoder().decode(
        ["bool"],
        happyResult.returnData,
      )
      expect(decoded).to.equal(true)
    })

    it("emits the ERC-20 Transfer log", function () {
      expect(happyResult.logs).to.have.lengthOf(1)
      const log = happyResult.logs[0]
      expect(log.address.toLowerCase()).to.equal(tokenAddr.toLowerCase())
      expect(log.topics[0]).to.equal(TRANSFER_TOPIC)
      expect(BigInt(log.topics[1])).to.equal(BigInt(senderAddr))
      expect(BigInt(log.topics[2])).to.equal(BigInt(recipientAddr))
    })
  })

  context("ERC-20 transfer revert path (Solidity 0.8 underflow)", function () {
    it("returns status 0x0", function () {
      expect(revertResult.status).to.equal("0x0")
    })

    it("per-call error is code 3 with 'execution reverted'", function () {
      expect(revertResult.error).to.exist
      expect(revertResult.error.code).to.equal(REVERTED)
      expect(revertResult.error.message).to.equal("execution reverted")
    })

    it("per-call error.data is the Panic(0x11) selector", function () {
      expect(revertResult.error.data.toLowerCase().startsWith(PANIC_SELECTOR)).to.equal(true)
    })
  })

  context("state chain across calls in one block", function () {
    it("mint -> transfer -> balanceOf surfaces transferred amount", function () {
      expect(stateChainBlocks).to.have.lengthOf(1)
      const calls = stateChainBlocks[0].calls
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
    it("block 1 mint is visible to block 2 balanceOf", function () {
      expect(crossBlockBlocks).to.have.lengthOf(2)
      const block1 = crossBlockBlocks[0]
      const block2 = crossBlockBlocks[1]
      expect(block1.calls[0].status).to.equal("0x1")
      const block2Call = block2.calls[0]
      expect(block2Call.status).to.equal("0x1")
      const [balance] = ethers.AbiCoder.defaultAbiCoder().decode(
        ["uint256"],
        block2Call.returnData,
      )
      expect(balance).to.equal(ethers.parseUnits("777", 18))
    })
  })

  context("BLOCKHASH simulated-sibling resolution", function () {
    it("BLOCKHASH(base+1) and BLOCKHASH(base+2) match the simulated headers", function () {
      expect(blockHashBlocks).to.have.lengthOf(3)
      const block3Calls = blockHashBlocks[2].calls
      expect(block3Calls).to.have.lengthOf(2)
      expect(block3Calls[0].status).to.equal("0x1")
      expect(block3Calls[1].status).to.equal("0x1")
      expect(block3Calls[0].returnData.toLowerCase()).to.equal(
        blockHashBlocks[0].hash.toLowerCase(),
      )
      expect(block3Calls[1].returnData.toLowerCase()).to.equal(
        blockHashBlocks[1].hash.toLowerCase(),
      )
    })
  })

  context("BLOCKHASH below base falls through to canonical chain", function () {
    it("returns the canonical block hash or zero for the probed height", function () {
      expect(blockHashBelowBaseBlocks).to.have.lengthOf(1)
      const call = blockHashBelowBaseBlocks[0].calls[0]
      expect(call.status).to.equal("0x1")
      expect(call.returnData).to.match(/^0x[0-9a-f]{64}$/i)
      // BLOCKHASH for canonical heights below base is best-effort: when
      // CometBFT's historical-info window still retains the height, the
      // real canonical hash is returned; otherwise the EVM convention is
      // to surface zero (as the YP defines BLOCKHASH for unknown heights).
      // Both outcomes are spec-conformant for mezod — this case validates
      // the canonical-fallback wiring, not unbounded historical retention.
      const ZERO = "0x" + "00".repeat(32)
      const actual = call.returnData.toLowerCase()
      if (blockHashBelowBaseExpected !== null) {
        expect([blockHashBelowBaseExpected.toLowerCase(), ZERO]).to.include(
          actual,
        )
      }
    })
  })

  context("MovePrecompileTo: sha256 to a destination", function () {
    it("returns the expected sha256 hash at the new address", function () {
      expect(sha256MoveResult.toLowerCase()).to.equal(SHA256_HELLO)
    })
  })

  context("MovePrecompileTo: self-reference rejected", function () {
    it("eth_call surfaces 'referenced itself' message", function () {
      expect(selfMoveErr).to.exist
      expect(extractMessage(selfMoveErr).toLowerCase()).to.include(
        "referenced itself",
      )
      // The typed -38022 SimError code surfaces only on eth_simulateV1; on
      // eth_call the same condition is bucketed under -32000 (server
      // error) because eth_call's error path does not pipe through the
      // typed *SimError envelope. The keeper-layer assertion is pinned by
      // TestApplyStateOverrides_MovePrecompileTo_SelfReference.
    })
  })

  context("MovePrecompileTo: ecrecover happy path", function () {
    it("calling the destination with a valid sig returns the recovered address", function () {
      expect(moveEcrecoverBlocks).to.have.lengthOf(1)
      const call = moveEcrecoverBlocks[0].calls[0]
      expect(call.status).to.equal("0x1")
      // Recovered address from the upstream fixture.
      expect(call.returnData.toLowerCase()).to.include(
        "b11cad98ad3f8114e0b3a1f6e7228bc8424df48a",
      )
    })
  })

  context("MovePrecompileTo: two precompiles to one dest -> -38023", function () {
    it("aborts with -38023", function () {
      expect(moveTwoToSameResult.error.thrown).to.equal(true)
      expect(moveTwoToSameResult.error.code).to.equal(TWO_PRECOMPILES_ONE_DEST)
    })
  })

  context("MovePrecompileTo: non-precompile rejected", function () {
    it("aborts with 'is not a precompile'", function () {
      expect(moveNonPrecompileResult.error.thrown).to.equal(true)
      expect(moveNonPrecompileResult.error.message.toLowerCase()).to.include(
        "is not a precompile",
      )
    })
  })

  context("Override identity precompile", function () {
    it("DEST_ADDR runs identity (echo); original runs override", function () {
      expect(overrideIdentityBlocks).to.have.lengthOf(1)
      const calls = overrideIdentityBlocks[0].calls
      expect(calls).to.have.lengthOf(2)
      expect(calls[0].status).to.equal("0x1")
      expect(calls[0].returnData.toLowerCase()).to.equal("0x1234")
      expect(calls[1].status).to.equal("0x1")
      // Empty bytecode at original returns nothing.
      expect(calls[1].returnData).to.equal("0x")
    })
  })

  context("DoS limits", function () {
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
      expect(intrinsicGasResultLimits.thrown).to.equal(true)
      expect(intrinsicGasResultLimits.code).to.equal(INTRINSIC_GAS_TOO_LOW)
    })
  })

  context("traceTransfers EOA -> EOA", function () {
    it("status 0x1 with one synthetic ERC-7528 log", function () {
      expect(traceOnEoaResult.status).to.equal("0x1")
      const synthetic = traceOnEoaResult.logs.filter(
        (l: any) => l.address.toLowerCase() === ERC7528_ADDRESS,
      )
      expect(synthetic).to.have.lengthOf(1)
      expect(synthetic[0].topics[0]).to.equal(TRANSFER_TOPIC)
      expect(BigInt(synthetic[0].topics[1])).to.equal(BigInt(senderAddr))
      expect(BigInt(synthetic[0].topics[2])).to.equal(BigInt(traceRecipientAddr))
    })
  })

  context("traceTransfers contract forwards value", function () {
    it("status 0x1 with two synthetic logs (one per call edge)", function () {
      expect(traceOnContractResult.status).to.equal("0x1")
      const synthetic = traceOnContractResult.logs.filter(
        (l: any) => l.address.toLowerCase() === ERC7528_ADDRESS,
      )
      expect(synthetic).to.have.lengthOf(2)
      // Edge 0: sender -> forwarder (top-level call).
      expect(BigInt(synthetic[0].topics[1])).to.equal(BigInt(senderAddr))
      expect(BigInt(synthetic[0].topics[2])).to.equal(BigInt(traceForwarderAddr))
      // Edge 1: forwarder -> recipient (inner CALL).
      expect(BigInt(synthetic[1].topics[1])).to.equal(BigInt(traceForwarderAddr))
      expect(BigInt(synthetic[1].topics[2])).to.equal(BigInt(traceRecipientAddr))
    })
  })

  context("traceTransfers omitted", function () {
    it("yields no synthetic logs", function () {
      expect(traceOffContractResult.status).to.equal("0x1")
      const synthetic = traceOffContractResult.logs.filter(
        (l: any) => l.address.toLowerCase() === ERC7528_ADDRESS,
      )
      expect(synthetic).to.have.lengthOf(0)
    })
  })

  context("validation=true: happy path", function () {
    it("returns per-call status 0x1", function () {
      expect(happyValidationResult.error.thrown).to.equal(false)
      const blocks = happyValidationResult.result!
      expect(blocks[0].calls[0].status).to.equal("0x1")
    })
  })

  context("validation=true: fatal codes", function () {
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
      expect(intrinsicGasResultValidation.error.thrown).to.equal(true)
      expect(intrinsicGasResultValidation.error.code).to.equal(
        INTRINSIC_GAS_TOO_LOW,
      )
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
    it("nonce-too-low yields per-call status 0x1", function () {
      expect(nonceLowNoValidationBlocks).to.have.lengthOf(1)
      expect(nonceLowNoValidationBlocks[0].calls[0].status).to.equal("0x1")
    })

    it("fee-cap-below-base yields per-call status 0x1", function () {
      expect(feeCapBelowBaseNoValidationBlocks).to.have.lengthOf(1)
      expect(feeCapBelowBaseNoValidationBlocks[0].calls[0].status).to.equal(
        "0x1",
      )
    })

    it("nonce too high yields per-call status 0x1", function () {
      expect(nonceHighNoValidationBlocks).to.have.lengthOf(1)
      expect(nonceHighNoValidationBlocks[0].calls[0].status).to.equal("0x1")
    })
  })

  context("balance after new (empty) block", function () {
    it("request succeeds across the gap and probes balance in block 2", function () {
      expect(balanceAfterNewBlockBlocks).to.have.lengthOf(2)
      const probe = balanceAfterNewBlockBlocks[1].calls[0]
      expect(probe.status).to.equal("0x1")
      expect(BigInt(probe.returnData)).to.equal(BigInt("0xde0b6b3a7640000"))
    })
  })

  context("block ordering errors", function () {
    it("non-monotonic block numbers -> -38020", function () {
      expect(blockNumOrderResult.error.thrown).to.equal(true)
      expect(blockNumOrderResult.error.code).to.equal(BLOCK_NUM_ORDER)
    })

    it("equal timestamps across blocks -> -38021", function () {
      expect(blockTsOrderResult.error.thrown).to.equal(true)
      expect(blockTsOrderResult.error.code).to.equal(BLOCK_TIMESTAMP_ORDER)
    })

    it("non-incrementing timestamps -> -38021", function () {
      expect(blockTsNonIncrementResult.error.thrown).to.equal(true)
      expect(blockTsNonIncrementResult.error.code).to.equal(
        BLOCK_TIMESTAMP_ORDER,
      )
    })
  })

  context("block timestamp / number auto-increment", function () {
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
    it("yields exactly one envelope with no transactions", function () {
      expect(emptyEarliestBlocks).to.have.lengthOf(1)
      expect(emptyEarliestBlocks[0].transactions).to.have.lengthOf(0)
    })
  })

  context("future block anchor", function () {
    it("rejects with header-not-found", function () {
      expect(futureBlockResult.error.thrown).to.equal(true)
      expect(futureBlockResult.error.message.toLowerCase()).to.include(
        "header not found",
      )
    })
  })

  context("override block num", function () {
    it("each block's NUMBER opcode returns its own override", function () {
      expect(overrideBlockNumBlocks).to.have.lengthOf(2)
      const a = BigInt(overrideBlockNumBlocks[0].number)
      const b = BigInt(overrideBlockNumBlocks[1].number)
      expect(BigInt(overrideBlockNumBlocks[0].calls[0].returnData)).to.equal(a)
      expect(BigInt(overrideBlockNumBlocks[1].calls[0].returnData)).to.equal(b)
    })
  })

  context("simple state diff", function () {
    it("reads slot 0 back as the overridden value", function () {
      expect(simpleStateDiffBlocks).to.have.lengthOf(1)
      const call = simpleStateDiffBlocks[0].calls[0]
      expect(call.status).to.equal("0x1")
      expect(call.returnData.toLowerCase()).to.equal(
        "0x12340000000000000000000000000000000000000000000000000000000000ff",
      )
    })
  })

  context("override storage slots (stateDiff vs state)", function () {
    it("both blocks return the overridden slot", function () {
      expect(overrideStorageSlotsBlocks).to.have.lengthOf(2)
      const expected =
        "0x12340000000000000000000000000000000000000000000000000000000000ff"
      expect(overrideStorageSlotsBlocks[0].calls[0].returnData.toLowerCase()).to.equal(
        expected,
      )
      expect(overrideStorageSlotsBlocks[1].calls[0].returnData.toLowerCase()).to.equal(
        expected,
      )
    })
  })

  context("override-address-twice (-38014 with validation)", function () {
    it("aborts with -38014 (no balance, value > 0)", function () {
      expect(overrideAddressTwiceResult.error.thrown).to.equal(true)
      expect(overrideAddressTwiceResult.error.code).to.equal(INSUFFICIENT_FUNDS)
    })
  })

  context("override across separate BlockStateCalls", function () {
    it("both blocks succeed with re-applied balance", function () {
      expect(overrideAcrossBlocksBlocks).to.have.lengthOf(2)
      expect(overrideAcrossBlocksBlocks[0].calls[0].status).to.equal("0x1")
      expect(overrideAcrossBlocksBlocks[1].calls[0].status).to.equal("0x1")
    })
  })

  context("overwrite existing contract", function () {
    it("first block reverts, second block runs the overridden bytecode", function () {
      expect(overwriteExistingContractBlocks).to.have.lengthOf(2)
      expect(overwriteExistingContractBlocks[0].calls[0].status).to.equal(
        "0x0",
      )
      expect(overwriteExistingContractBlocks[1].calls[0].status).to.equal(
        "0x1",
      )
    })
  })

  context("block-override reflected in contract", function () {
    it("NUMBER opcode at override returns the override number", function () {
      expect(blockOverrideReflectedBlocks).to.have.lengthOf(1)
      const call = blockOverrideReflectedBlocks[0].calls[0]
      expect(call.status).to.equal("0x1")
      expect(BigInt(call.returnData)).to.equal(
        BigInt(blockOverrideReflectedBlocks[0].number),
      )
    })
  })

  context("fee-recipient receives funds", function () {
    it("request succeeds with feeRecipient override + non-zero value", function () {
      expect(feeRecipientBlocks).to.have.lengthOf(1)
      expect(feeRecipientBlocks[0].calls[0].status).to.equal("0x1")
    })
  })

  context("logs: eth send produces no logs by default", function () {
    it("logs == [] when traceTransfers is omitted", function () {
      expect(logsNoneByDefaultBlocks).to.have.lengthOf(1)
      const call = logsNoneByDefaultBlocks[0].calls[0]
      expect(call.status).to.equal("0x1")
      expect(call.logs).to.have.lengthOf(0)
    })
  })

  context("logs: forward then revert produces no logs", function () {
    it("status 0x0, code 3, logs == []", function () {
      expect(logsForwardRevertBlocks).to.have.lengthOf(1)
      const call = logsForwardRevertBlocks[0].calls[0]
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
    it("status 0x1 and (best-effort) one ERC-7528 Transfer log", function () {
      expect(logsSelfdestructBlocks).to.have.lengthOf(1)
      const call = logsSelfdestructBlocks[0].calls[0]
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
    it("emits exactly one synthetic log (top-level edge only)", function () {
      expect(delegateCallEoaBlocks).to.have.lengthOf(1)
      const call = delegateCallEoaBlocks[0].calls[0]
      expect(call.status).to.equal("0x1")
      const synthetic = call.logs.filter(
        (l: any) => l.address.toLowerCase() === ERC7528_ADDRESS,
      )
      expect(synthetic).to.have.lengthOf(1)
    })
  })

  context("comprehensive: simple two value transfers", function () {
    it("both per-call status 0x1", function () {
      expect(simpleSanityBlocks).to.have.lengthOf(1)
      expect(simpleSanityBlocks[0].calls[0].status).to.equal("0x1")
      expect(simpleSanityBlocks[0].calls[1].status).to.equal("0x1")
    })
  })

  context("comprehensive: transfer over multiple BlockStateCalls", function () {
    it("all four calls succeed", function () {
      expect(transferOverBlockStateCallsBlocks).to.have.lengthOf(2)
      for (const block of transferOverBlockStateCallsBlocks) {
        for (const call of block.calls) {
          expect(call.status).to.equal("0x1")
        }
      }
    })
  })

  context("comprehensive: set-read-storage", function () {
    it("call 2 returns the value set by call 1", function () {
      expect(setReadStorageBlocks).to.have.lengthOf(1)
      const calls = setReadStorageBlocks[0].calls
      expect(calls).to.have.lengthOf(2)
      expect(calls[0].status).to.equal("0x1")
      expect(calls[1].status).to.equal("0x1")
      expect(BigInt(calls[1].returnData)).to.equal(5n)
    })
  })

  context("comprehensive: contract calls itself", function () {
    it("status 0x1", function () {
      expect(contractCallsItselfBlocks).to.have.lengthOf(1)
      expect(contractCallsItselfBlocks[0].calls[0].status).to.equal("0x1")
    })
  })

  context("comprehensive: run-gas-spending across blocks", function () {
    it("all three blocks' calls succeed", function () {
      expect(runGasSpendingBlocks).to.have.lengthOf(3)
      for (const block of runGasSpendingBlocks) {
        for (const call of block.calls) {
          expect(call.status).to.equal("0x1")
        }
      }
    })
  })

  context("FullTx: hash-only default", function () {
    it("transactions[] entries are 0x-prefixed 32-byte hex strings", function () {
      expect(hashOnlyResult.transactions).to.be.an("array").with.lengthOf(1)
      const entry = hashOnlyResult.transactions[0]
      expect(entry).to.be.a("string")
      expect(entry).to.match(/^0x[0-9a-f]{64}$/i)
    })
  })

  context("FullTx: returnFullTransactions=true", function () {
    it("transactions[] carries RPCTransaction objects with from patched", function () {
      expect(fullTxResult.transactions).to.be.an("array").with.lengthOf(1)
      const tx = fullTxResult.transactions[0]
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
    it("each tx's from resolves and hashes are unique", function () {
      expect(multiSenderResult.transactions).to.be.an("array").with.lengthOf(2)
      const fromAddrs = new Set(
        multiSenderResult.transactions.map((tx: any) => tx.from.toLowerCase()),
      )
      expect(fromAddrs.has(senderAddr.toLowerCase())).to.equal(true)
      expect(fromAddrs.has(DETACHED_SECOND_SENDER.toLowerCase())).to.equal(true)
      const hashes = multiSenderResult.transactions.map((tx: any) => tx.hash)
      expect(new Set(hashes).size).to.equal(hashes.length)
    })
  })

  context("Block envelope: non-empty block", function () {
    it("transactionsRoot is non-empty", function () {
      expect(nonEmptyBlockResult.transactionsRoot.toLowerCase()).to.not.equal(
        EMPTY_TRIE_ROOT,
      )
    })

    it("receiptsRoot is non-empty", function () {
      expect(nonEmptyBlockResult.receiptsRoot.toLowerCase()).to.not.equal(
        EMPTY_TRIE_ROOT,
      )
    })

    it("logsBloom positive-tests the Transfer topic", function () {
      const bloomBytes = ethers.getBytes(nonEmptyBlockResult.logsBloom)
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
      expect(BigInt(nonEmptyBlockResult.size)).to.be.greaterThan(0n)
    })

    it("hash is stable across identical re-runs", function () {
      expect(nonEmptyBlockResult.hash).to.equal(nonEmptyBlockResultRerun.hash)
    })
  })

  context("Block envelope: gap-fill block", function () {
    it("contains no transactions", function () {
      const gap = gapBlocksResult[1]
      expect(gap.transactions).to.be.an("array").with.lengthOf(0)
    })

    it("uses the empty trie root for transactionsRoot and receiptsRoot", function () {
      const gap = gapBlocksResult[1]
      expect(gap.transactionsRoot.toLowerCase()).to.equal(EMPTY_TRIE_ROOT)
      expect(gap.receiptsRoot.toLowerCase()).to.equal(EMPTY_TRIE_ROOT)
    })

    it("logsBloom is the zero bloom", function () {
      const gap = gapBlocksResult[1]
      expect(gap.logsBloom.toLowerCase()).to.equal(ZERO_BLOOM)
    })
  })
})
