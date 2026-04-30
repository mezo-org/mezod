import { expect } from "chai"
import hre, { ethers } from "hardhat"
import { SimpleToken } from "../typechain-types/SimpleToken"
import { getDeployedContract } from "./helpers/contract"

// SimulateV1_FullTx covers the `returnFullTransactions` option:
// when true the response's `transactions` field carries
// RPCTransaction objects (with the `from` address patched from the
// driver's senders-by-hash map); when omitted it carries plain
// 32-byte hash strings.
describe("SimulateV1_FullTx", function () {
  const { deployments } = hre

  let simpleToken: SimpleToken
  let tokenAddr: string
  let senderAddr: string
  let secondSenderAddr: string
  let recipientAddr: string

  let hashOnlyResult: any
  let fullTxResult: any
  let multiSenderResult: any

  before(async function () {
    await deployments.fixture(["BridgeOutDelegate"])
    simpleToken = await getDeployedContract<SimpleToken>("SimpleToken")
    tokenAddr = await simpleToken.getAddress()

    const [deployer] = await ethers.getSigners()
    senderAddr = deployer.address
    // Detach the second sender from the localnode dev keys so the
    // simulate request can fake any nonce/balance state without
    // colliding with on-chain reality.
    secondSenderAddr = "0xc100000000000000000000000000000000000201"
    recipientAddr = ethers.Wallet.createRandom().address

    // Mint once so the deployer can exercise transfer().
    const mintAmount = ethers.parseUnits("1000", 18)
    const mintTx = await simpleToken
      .connect(deployer)
      .mint(senderAddr, mintAmount)
    await mintTx.wait()

    const transferData = simpleToken.interface.encodeFunctionData(
      "transfer",
      [recipientAddr, ethers.parseUnits("1", 18)],
    )

    // Hash-only path: returnFullTransactions omitted (default false).
    const hashOnlyOpts = {
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
    const hashOnlyBlocks: any[] = await ethers.provider.send(
      "eth_simulateV1",
      [hashOnlyOpts, "latest"],
    )
    hashOnlyResult = hashOnlyBlocks[0]

    // Full-tx path: returnFullTransactions=true.
    const fullTxOpts = {
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
      returnFullTransactions: true,
    }
    const fullTxBlocks: any[] = await ethers.provider.send("eth_simulateV1", [
      fullTxOpts,
      "latest",
    ])
    fullTxResult = fullTxBlocks[0]

    // Multi-sender single block: two callers in the same block,
    // both with state overrides funding their balance. Asserts the
    // Senders-by-hash patch matches each tx individually.
    const balance = ethers.toQuantity(ethers.parseEther("100"))
    const multiSenderOpts = {
      blockStateCalls: [
        {
          stateOverrides: {
            [senderAddr]: { balance },
            [secondSenderAddr]: { balance },
          },
          calls: [
            {
              from: senderAddr,
              to: recipientAddr,
              value: ethers.toQuantity(1n),
            },
            {
              from: secondSenderAddr,
              to: recipientAddr,
              value: ethers.toQuantity(2n),
            },
          ],
        },
      ],
      returnFullTransactions: true,
    }
    const multiSenderBlocks: any[] = await ethers.provider.send(
      "eth_simulateV1",
      [multiSenderOpts, "latest"],
    )
    multiSenderResult = multiSenderBlocks[0]
  })

  context("returnFullTransactions omitted (default)", function () {
    it("transactions[] carries hash strings, not objects", function () {
      expect(hashOnlyResult.transactions).to.be.an("array").with.lengthOf(1)
      const entry = hashOnlyResult.transactions[0]
      expect(entry).to.be.a("string", "hash-only mode must yield strings")
      expect(entry).to.match(
        /^0x[0-9a-f]{64}$/i,
        "tx hash must be 0x-prefixed 32-byte hex",
      )
    })
  })

  context("returnFullTransactions=true", function () {
    it("transactions[] carries objects (RPCTransaction shape)", function () {
      expect(fullTxResult.transactions).to.be.an("array").with.lengthOf(1)
      const tx = fullTxResult.transactions[0]
      expect(tx).to.be.an("object")
      // Pin a representative subset of the RPCTransaction fields
      // — drift on any of these would imply the upstream JSON shape
      // changed.
      for (const key of [
        "blockHash",
        "blockNumber",
        "from",
        "gas",
        "gasPrice",
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
    })

    it("from is patched from the senders-by-hash map", function () {
      const tx = fullTxResult.transactions[0]
      // Lower-cased compare to dodge checksum-vs-lowercase encoding
      // differences between providers.
      expect(tx.from.toLowerCase()).to.equal(senderAddr.toLowerCase())
    })

    it("hash is a 32-byte hex string", function () {
      const tx = fullTxResult.transactions[0]
      expect(tx.hash).to.match(/^0x[0-9a-f]{64}$/i)
    })
  })

  context("multi-sender single block (returnFullTransactions=true)", function () {
    it("each tx's from resolves to its own sender (matched by hash)", function () {
      expect(multiSenderResult.transactions).to.be.an("array").with.lengthOf(2)
      const fromAddrs = new Set(
        multiSenderResult.transactions.map((tx: any) => tx.from.toLowerCase()),
      )
      expect(fromAddrs.has(senderAddr.toLowerCase())).to.equal(
        true,
        "first sender must appear",
      )
      expect(fromAddrs.has(secondSenderAddr.toLowerCase())).to.equal(
        true,
        "second sender must appear",
      )
    })

    it("each tx hash is unique (synthetic txs distinct per call)", function () {
      const hashes = multiSenderResult.transactions.map((tx: any) => tx.hash)
      expect(new Set(hashes).size).to.equal(hashes.length)
    })
  })
})
