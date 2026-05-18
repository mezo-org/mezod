import { expect } from "chai"
import { ethers } from "hardhat"

/**
 * Eip2935_MezoDivergence — pins the deliberate gap between Mezo and the
 * Prague EIP-2935 history-storage layout.
 *
 * Upstream Ethereum deploys a "history storage" system contract at
 *   0x0000F90827F1C53a10cb7A02335B175320002935
 * at fork activation, and the BLOCKHASH opcode falls through to that
 * contract for heights older than the most recent 256 blocks. Mezo does
 * not deploy that contract — `core.ProcessParentBlockHash` is never
 * called from Mezo's block-processing path. The slot stays an empty
 * account on chain, which is what the assertions below pin.
 *
 * The EVM-compatibility doc additionally claims BLOCKHASH for older
 * blocks is served by `Keeper.GetHashFn` reading from x/poa's
 * historical-info store. That fallback is NOT exercised here: it is a
 * documented Mezo divergence from the EIP-2935 system-contract path but
 * lives in a separate code path that is not affected by Osaka
 * activation, and on a fresh localnode it does not surface a non-zero
 * result for past heights — pinning the documented behavior would have
 * the test pass against the doc but fail against the running chain.
 * A dedicated suite can hold that surface separately once the gap is
 * reconciled.
 */

const HISTORY_STORAGE_ADDR = ethers.getAddress(
  "0x0000F90827F1C53a10cb7A02335B175320002935",
)

describe("Eip2935_MezoDivergence", function () {
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
})
