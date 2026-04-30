import { expect } from "chai"
import hre, { ethers } from "hardhat"
import { SimpleToken } from "../typechain-types/SimpleToken"
import { getDeployedContract } from "./helpers/contract"

// SimulateV1_BlockEnvelope covers the full block envelope returned by
// eth_simulateV1: real *ethtypes.Block backing produces correct
// logsBloom, transactionsRoot, receiptsRoot, and size, while stateRoot
// stays at the zero hash (Mezo divergence; see assembleSimBlock).
describe("SimulateV1_BlockEnvelope", function () {
  const { deployments } = hre

  // 256-byte zero bloom serialized as `0x` + 512 hex chars. Used to
  // pin the empty-block logsBloom case.
  const ZERO_BLOOM = "0x" + "00".repeat(256)

  // ERC-20 Transfer event topic:
  // keccak256("Transfer(address,address,uint256)").
  const TRANSFER_TOPIC = ethers.id("Transfer(address,address,uint256)")

  let simpleToken: SimpleToken
  let tokenAddr: string
  let senderAddr: string
  let recipientAddr: string

  let baseBlockNumber: number
  let nonEmptyBlockResult: any
  let nonEmptyBlockResultRerun: any
  let gapBlocksResult: any[]

  before(async function () {
    await deployments.fixture(["BridgeOutDelegate"])
    simpleToken = await getDeployedContract<SimpleToken>("SimpleToken")
    tokenAddr = await simpleToken.getAddress()

    const [deployer] = await ethers.getSigners()
    senderAddr = deployer.address
    recipientAddr = ethers.Wallet.createRandom().address

    // Mint enough to the sender so the transfer below succeeds.
    const mintAmount = ethers.parseUnits("1000", 18)
    const mintTx = await simpleToken
      .connect(deployer)
      .mint(senderAddr, mintAmount)
    await mintTx.wait()

    baseBlockNumber = await ethers.provider.getBlockNumber()

    // Single-block scenario: one ERC-20 transfer call. Drives the
    // happy-path assertions on the envelope.
    const transferData = simpleToken.interface.encodeFunctionData(
      "transfer",
      [recipientAddr, ethers.parseUnits("1", 18)],
    )
    const transferOpts = {
      blockStateCalls: [
        {
          calls: [
            {
              from: senderAddr,
              to: tokenAddr,
              data: transferData,
            },
          ],
        },
      ],
    }
    const transferBlocks: any[] = await ethers.provider.send(
      "eth_simulateV1",
      [transferOpts, "latest"],
    )
    nonEmptyBlockResult = transferBlocks[0]

    // Re-run the same opts to assert hash determinism.
    const transferBlocksRerun: any[] = await ethers.provider.send(
      "eth_simulateV1",
      [transferOpts, "latest"],
    )
    nonEmptyBlockResultRerun = transferBlocksRerun[0]

    // Gap-fill scenario: jump from base+1 to base+3, sanitizeSimChain
    // inserts an empty block at base+2 with no calls. Inspect that
    // gap block to assert the empty-block envelope shape.
    const gapOpts = {
      blockStateCalls: [
        {
          calls: [
            {
              from: senderAddr,
              to: tokenAddr,
              data: transferData,
            },
          ],
        },
        {
          blockOverrides: {
            number: ethers.toQuantity(baseBlockNumber + 3),
          },
          calls: [
            {
              from: senderAddr,
              to: tokenAddr,
              data: transferData,
            },
          ],
        },
      ],
    }
    gapBlocksResult = await ethers.provider.send("eth_simulateV1", [
      gapOpts,
      "latest",
    ])
  })

  context("non-empty block envelope", function () {
    it("returns a non-empty transactionsRoot", function () {
      // Empty-trie root = keccak256(rlp(empty list)). DeriveSha for a
      // populated tx list cannot collide with this value.
      const EMPTY_TX_ROOT =
        "0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421"
      expect(nonEmptyBlockResult.transactionsRoot.toLowerCase()).to.not.equal(
        EMPTY_TX_ROOT,
      )
    })

    it("returns a non-empty receiptsRoot", function () {
      const EMPTY_RECEIPTS_ROOT =
        "0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421"
      expect(nonEmptyBlockResult.receiptsRoot.toLowerCase()).to.not.equal(
        EMPTY_RECEIPTS_ROOT,
      )
    })

    it("returns a logsBloom containing the Transfer event topic", function () {
      // ethers' BloomFilter helper isn't easily available here;
      // perform the bitwise probe manually. The Bloom encodes a
      // topic by hashing it three times into 256-byte slots. A
      // mismatch in any of the three positions -> the topic is
      // definitively absent.
      const bloom = nonEmptyBlockResult.logsBloom
      const bloomBytes = ethers.getBytes(bloom)
      const topicBytes = ethers.getBytes(TRANSFER_TOPIC)

      // Replicate go-ethereum's bloom indexing (see core/types/bloom9):
      // for each of the three pairs (i=0..2) take 16-bit index from
      // keccak256(topic)[2i..2i+2] mod 2048, locate that bit.
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
      expect(allBitsSet).to.equal(
        true,
        "logsBloom must affirmative-test the Transfer topic",
      )
    })

    it("returns size > 0", function () {
      expect(BigInt(nonEmptyBlockResult.size)).to.be.greaterThan(0n)
    })

    it("returns stateRoot equal to the zero hash (Mezo divergence)", function () {
      const zeroHash = "0x" + "00".repeat(32)
      expect(nonEmptyBlockResult.stateRoot.toLowerCase()).to.equal(zeroHash)
    })

    it("returns block.hash that is stable across identical re-runs", function () {
      expect(nonEmptyBlockResult.hash).to.equal(nonEmptyBlockResultRerun.hash)
    })
  })

  context("empty (gap-fill) block envelope", function () {
    it("the gap block contains no transactions", function () {
      // Result layout: [base+1 block], [gap base+2 block], [base+3 block].
      const gap = gapBlocksResult[1]
      expect(gap.transactions).to.be.an("array").with.lengthOf(0)
    })

    it("the gap block uses the empty transactionsRoot", function () {
      const EMPTY_TX_ROOT =
        "0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421"
      const gap = gapBlocksResult[1]
      expect(gap.transactionsRoot.toLowerCase()).to.equal(EMPTY_TX_ROOT)
    })

    it("the gap block uses the empty receiptsRoot", function () {
      const EMPTY_RECEIPTS_ROOT =
        "0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421"
      const gap = gapBlocksResult[1]
      expect(gap.receiptsRoot.toLowerCase()).to.equal(EMPTY_RECEIPTS_ROOT)
    })

    it("the gap block surfaces a zero logsBloom", function () {
      const gap = gapBlocksResult[1]
      expect(gap.logsBloom.toLowerCase()).to.equal(ZERO_BLOOM)
    })
  })
})
