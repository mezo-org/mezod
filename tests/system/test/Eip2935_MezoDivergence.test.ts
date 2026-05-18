import { expect } from "chai"
import hre, { ethers } from "hardhat"
import type { BlockhashCheck } from "../typechain-types"
import { getDeployedContract } from "./helpers/contract"

/**
 * Eip2935_MezoDivergence — pins the deliberate gap between Mezo and the
 * Prague EIP-2935 history-storage layout.
 *
 * Upstream Ethereum deploys a "history storage" system contract at
 *   0x0000F90827F1C53a10cb7A02335B175320002935
 * at fork activation, and the BLOCKHASH opcode falls through to that
 * contract for heights older than the most recent 256 blocks. Mezo
 * resolves BLOCKHASH through x/poa's historical-info store
 * (`x/evm/keeper/state_transition.go:Keeper.GetHashFn`), so the EIP-2935
 * system contract is never deployed. This file pins two divergences from
 * that choice:
 *
 *   1. The 0x...2935 slot remains an empty account on chain. Any contract
 *      reading it via SLOAD/EXTCODESIZE must see zero state.
 *   2. BLOCKHASH for the most recent block still returns the parent hash
 *      exactly as eth_getBlockByNumber reports it. The opcode is not
 *      degraded by the absence of the history-storage contract.
 */

const HISTORY_STORAGE_ADDR = ethers.getAddress(
  "0x0000F90827F1C53a10cb7A02335B175320002935",
)

describe("Eip2935_MezoDivergence", function () {
  const { deployments } = hre
  let bh: BlockhashCheck

  before(async function () {
    await deployments.fixture(["BlockhashCheck"])
    bh = await getDeployedContract("BlockhashCheck")
  })

  it("EIP-2935 history-storage contract is not deployed", async function () {
    const code = await ethers.provider.getCode(HISTORY_STORAGE_ADDR)
    expect(code).to.equal("0x")
  })

  it("history-storage account has zero balance and nonce", async function () {
    const balance = await ethers.provider.getBalance(HISTORY_STORAGE_ADDR)
    const nonce = await ethers.provider.getTransactionCount(
      HISTORY_STORAGE_ADDR,
    )
    expect(balance).to.equal(0n)
    expect(nonce).to.equal(0)
  })

  it("BLOCKHASH(N-1) matches eth_getBlockByNumber(N-1).hash", async function () {
    // BLOCKHASH(block.number) returns zero by spec; the most recent
    // queryable height is current - 1. We deliberately don't probe
    // > 256 blocks back here — that path is also Mezo-divergent (served
    // by x/poa rather than the EIP-2935 ring buffer) but needs a chain
    // with >256 blocks already produced, which the localnode warm-up
    // doesn't reliably give the test harness.
    const current = await ethers.provider.getBlockNumber()
    expect(current, "no blocks produced yet").to.be.gte(1)
    const parent = await ethers.provider.getBlock(current - 1)
    if (parent === null) {
      throw new Error("parent block not found")
    }
    const got = await bh.blockHashOf(current - 1)
    expect(got).to.equal(parent.hash)
  })
})
