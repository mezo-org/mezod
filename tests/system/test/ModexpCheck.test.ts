import { expect } from "chai"
import hre from "hardhat"
import type { ModexpCheck } from "../typechain-types"
import { getDeployedContract } from "./helpers/contract"

const MODEXP = hre.ethers.getAddress(
  "0x0000000000000000000000000000000000000005",
)

// Build a MODEXP payload from raw byte buffers. Layout per EIP-198:
//   <base_len(32 BE)><exp_len(32 BE)><mod_len(32 BE)><base><exp><mod>
function buildModexpInput(
  base: Uint8Array,
  exp: Uint8Array,
  mod: Uint8Array,
  // Allows the test to lie about a buffer's declared length to drive
  // EIP-7823's "declared length above 1024 bytes" rejection path without
  // requiring the test to actually allocate 1025+ bytes of body content.
  lenOverride?: { base?: number; exp?: number; mod?: number },
): string {
  const baseLen = lenOverride?.base ?? base.length
  const expLen = lenOverride?.exp ?? exp.length
  const modLen = lenOverride?.mod ?? mod.length

  const uint256BE = (n: number): Uint8Array => {
    const out = new Uint8Array(32)
    let v = n
    for (let i = 31; i >= 0 && v > 0; i--) {
      out[i] = v & 0xff
      v = Math.floor(v / 256)
    }
    return out
  }

  return hre.ethers.concat([
    uint256BE(baseLen),
    uint256BE(expLen),
    uint256BE(modLen),
    base,
    exp,
    mod,
  ])
}

describe("ModexpCheck (Osaka)", function () {
  const { deployments } = hre
  let modexp: ModexpCheck

  before(async function () {
    await deployments.fixture(["ModexpCheck"])
    modexp = await getDeployedContract("ModexpCheck")
  })

  describe("baseline (pre-EIP-7823 sanity)", function () {
    it("computes 3^4 mod 7 = 4", async function () {
      const payload = buildModexpInput(
        new Uint8Array([0x03]),
        new Uint8Array([0x04]),
        new Uint8Array([0x07]),
      )
      const [ok, out] = await modexp.staticCall.staticCall(MODEXP, payload)
      expect(ok, "MODEXP staticcall reverted").to.be.true
      expect(out).to.equal("0x04")
    })

    it("accepts inputs at the 1024-byte boundary", async function () {
      // base_len = mod_len = 1024 (the EIP-7823 upper bound). Body is
      // zero-padded; the actual result is 0 since base = 0. The point is
      // that the precompile must accept the request without reverting.
      const buf1024 = new Uint8Array(1024)
      const payload = buildModexpInput(
        buf1024,
        new Uint8Array([0x01]),
        buf1024,
      )
      const [ok] = await modexp.staticCall.staticCall(MODEXP, payload)
      expect(ok, "boundary-size MODEXP staticcall reverted").to.be.true
    })
  })

  describe("EIP-7823 (input upper bound 8192 bits)", function () {
    it("rejects base_len = 1025", async function () {
      const payload = buildModexpInput(
        new Uint8Array(1025),
        new Uint8Array([0x01]),
        new Uint8Array([0x07]),
      )
      const [ok] = await modexp.staticCall.staticCall(MODEXP, payload)
      expect(ok, "oversized base accepted").to.be.false
    })

    it("rejects exp_len = 1025", async function () {
      const payload = buildModexpInput(
        new Uint8Array([0x02]),
        new Uint8Array(1025),
        new Uint8Array([0x07]),
      )
      const [ok] = await modexp.staticCall.staticCall(MODEXP, payload)
      expect(ok, "oversized exponent accepted").to.be.false
    })

    it("rejects mod_len = 1025", async function () {
      const payload = buildModexpInput(
        new Uint8Array([0x02]),
        new Uint8Array([0x01]),
        new Uint8Array(1025),
      )
      const [ok] = await modexp.staticCall.staticCall(MODEXP, payload)
      expect(ok, "oversized modulus accepted").to.be.false
    })
  })

  describe("EIP-7883 (MODEXP gas cost floor)", function () {
    // Pre-Osaka the MODEXP minimum cost was 200 gas. EIP-7883 raises the
    // minimum to 500 gas. Probing with the smallest valid 1-byte/1-byte/
    // 1-byte input is what pins the floor: pre-Osaka it would charge 200,
    // post-Osaka it must charge at least 500. The assertion below would
    // fail on a chain still running the pre-Osaka schedule.
    it("charges at least 500 gas for the minimum-size call", async function () {
      const payload = buildModexpInput(
        new Uint8Array([0x02]),
        new Uint8Array([0x01]),
        new Uint8Array([0x07]),
      )
      const [ok, , gasUsed] = await modexp.staticCallWithGas.staticCall(
        MODEXP,
        payload,
      )
      expect(ok, "minimum-size MODEXP staticcall reverted").to.be.true
      expect(gasUsed).to.be.gte(500n)
    })
  })
})
