import { expect } from "chai"
import { ethers } from "hardhat"

/**
 * Eip7685_MezoDivergence — pins the absence of EIP-7685 EL-CL requests
 * surface on Mezo.
 *
 * Upstream Prague introduces a block-level `requests` list and commits
 * to it via a new `requestsHash` header field. Mezo has no EL→CL
 * messaging (no beacon chain) so it produces no requests at all. The
 * relevant Mezo-side choice is the JSON-RPC shape: `FormatBlock`
 * (`rpc/types/utils.go`) omits the `requestsHash` field from the
 * eth_getBlockByNumber response entirely — it is not emitted as `null`,
 * and not emitted as the zero hash either. This file pins that contract.
 */

describe("Eip7685_MezoDivergence", function () {
  it("eth_getBlockByNumber('latest') omits requestsHash", async function () {
    const raw = (await ethers.provider.send("eth_getBlockByNumber", [
      "latest",
      false,
    ])) as Record<string, unknown>
    expect(raw).to.not.have.property("requestsHash")
    // Spec-compatible aliases that some clients also expose — if any of
    // these ever appears in the Mezo response, the divergence has
    // softened back toward upstream and the test should make that
    // surface as a failure.
    expect(raw).to.not.have.property("requestsRoot")
  })

  it("eth_getBlockByHash for the same block also omits requestsHash", async function () {
    const latest = (await ethers.provider.send("eth_getBlockByNumber", [
      "latest",
      false,
    ])) as Record<string, unknown>
    expect(latest.hash, "block has no hash").to.be.a("string")
    const byHash = (await ethers.provider.send("eth_getBlockByHash", [
      latest.hash as string,
      false,
    ])) as Record<string, unknown>
    expect(byHash).to.not.have.property("requestsHash")
    expect(byHash).to.not.have.property("requestsRoot")
  })

  it("eth_getBlockByNumber for a Prague-active height (0x1) omits requestsHash", async function () {
    // Fresh chains start with Prague (and Osaka) active from genesis.
    // Block 1 is the first post-genesis block; if upstream's Prague
    // requestsHash plumbing ever reaches Mezo's block-formatter, the
    // earliest Prague-active block is where it would surface first.
    const block = (await ethers.provider.send("eth_getBlockByNumber", [
      "0x1",
      false,
    ])) as Record<string, unknown> | null
    if (block === null) {
      // Chain is fresh-fresh and block 1 isn't produced yet; skip
      // rather than flake. The "latest" assertion above already covers
      // the divergence.
      this.skip()
    }
    expect(block).to.not.have.property("requestsHash")
    expect(block).to.not.have.property("requestsRoot")
  })
})
