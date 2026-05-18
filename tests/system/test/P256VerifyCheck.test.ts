import { expect } from "chai"
import hre from "hardhat"
import { webcrypto } from "node:crypto"
import type { P256VerifyCheck } from "../typechain-types"
import { getDeployedContract } from "./helpers/contract"

// P256VERIFY precompile, EIP-7951.
const P256VERIFY = hre.ethers.getAddress(
  "0x0000000000000000000000000000000000000100",
)
// Empty slot used to pin "is a precompile" vs. "is an empty account":
// 0x0101 is one byte past the precompile and must not behave like
// P256VERIFY.
const EMPTY_NEIGHBOR = hre.ethers.getAddress(
  "0x0000000000000000000000000000000000000101",
)

const ZERO_HEX = (n: number): string => "0x" + "00".repeat(n)

const SUCCESS_OUTPUT = "0x" + "00".repeat(31) + "01"

// Build the 160-byte EIP-7951 input: hash || r || s || x || y.
function buildP256Input(parts: {
  hash: Uint8Array
  r: Uint8Array
  s: Uint8Array
  x: Uint8Array
  y: Uint8Array
}): string {
  for (const [k, v] of Object.entries(parts)) {
    if (v.length !== 32) {
      throw new Error(`P256VERIFY ${k}: expected 32 bytes, got ${v.length}`)
    }
  }
  return hre.ethers.hexlify(
    hre.ethers.concat([parts.hash, parts.r, parts.s, parts.x, parts.y]),
  )
}

// Sign `message` with a freshly generated P-256 keypair and return the
// raw 32-byte components needed to build a P256VERIFY input.
async function signWithFreshP256Key(message: Uint8Array): Promise<{
  hash: Uint8Array
  r: Uint8Array
  s: Uint8Array
  x: Uint8Array
  y: Uint8Array
}> {
  const keyPair = await webcrypto.subtle.generateKey(
    { name: "ECDSA", namedCurve: "P-256" },
    true,
    ["sign", "verify"],
  )

  // WebCrypto's ECDSA sign with SHA-256 returns raw IEEE P1363 (r||s).
  const sigRaw = new Uint8Array(
    await webcrypto.subtle.sign(
      { name: "ECDSA", hash: "SHA-256" },
      keyPair.privateKey,
      message,
    ),
  )
  // exportKey "raw" gives the uncompressed point: 0x04 || x || y.
  const rawPub = new Uint8Array(
    await webcrypto.subtle.exportKey("raw", keyPair.publicKey),
  )
  if (rawPub.length !== 65 || rawPub[0] !== 0x04) {
    throw new Error("unexpected raw public-key encoding")
  }
  const hash = new Uint8Array(await webcrypto.subtle.digest("SHA-256", message))

  return {
    hash,
    r: sigRaw.subarray(0, 32),
    s: sigRaw.subarray(32, 64),
    x: rawPub.subarray(1, 33),
    y: rawPub.subarray(33, 65),
  }
}

describe("P256VerifyCheck (EIP-7951)", function () {
  const { deployments } = hre
  let p256: P256VerifyCheck

  before(async function () {
    await deployments.fixture(["P256VerifyCheck"])
    p256 = await getDeployedContract("P256VerifyCheck")
  })

  describe("surface", function () {
    it("invalid all-zero input returns empty (precompile alive)", async function () {
      // 160 bytes of zeros is well-formed in shape but not a valid
      // signature (zero r/s, point (0,0) not on curve). The precompile
      // must surface the failure as "no output", not as a revert. The
      // ok=true path is what distinguishes it from EMPTY_NEIGHBOR below.
      const [ok, out] = await p256.staticCall.staticCall(
        P256VERIFY,
        ZERO_HEX(160),
      )
      expect(ok, "staticcall reverted").to.be.true
      expect(out).to.equal("0x")
    })

    it("wrong-size input (159 bytes) returns empty", async function () {
      const [ok, out] = await p256.staticCall.staticCall(
        P256VERIFY,
        ZERO_HEX(159),
      )
      expect(ok, "staticcall reverted").to.be.true
      expect(out).to.equal("0x")
    })

    it("wrong-size input (161 bytes) returns empty", async function () {
      const [ok, out] = await p256.staticCall.staticCall(
        P256VERIFY,
        ZERO_HEX(161),
      )
      expect(ok, "staticcall reverted").to.be.true
      expect(out).to.equal("0x")
    })

    it("empty input returns empty (well-defined for invalid size)", async function () {
      const [ok, out] = await p256.staticCall.staticCall(P256VERIFY, "0x")
      expect(ok, "staticcall reverted").to.be.true
      expect(out).to.equal("0x")
    })

    // Address 0x0101 must remain an empty account, not bleed P256VERIFY
    // semantics into a neighbor slot.
    it("0x0101 is not a precompile (empty account)", async function () {
      const [ok, out] = await p256.staticCall.staticCall(
        EMPTY_NEIGHBOR,
        ZERO_HEX(160),
      )
      expect(ok, "staticcall reverted").to.be.true
      expect(out).to.equal("0x")
    })
  })

  describe("cryptographic correctness", function () {
    it("accepts a valid signature produced by a fresh keypair", async function () {
      // The verification surface is the responsibility of upstream geth;
      // this test only proves Mezo routes calls to the precompile and
      // does not strip the 32-byte success output on the way back.
      const message = new TextEncoder().encode(
        "mezo p256verify system test",
      )
      const parts = await signWithFreshP256Key(message)
      const payload = buildP256Input(parts)
      const [ok, out] = await p256.staticCall.staticCall(P256VERIFY, payload)
      expect(ok, "staticcall reverted").to.be.true
      expect(out).to.equal(SUCCESS_OUTPUT)
    })

    it("rejects a signature that has been tampered with", async function () {
      const message = new TextEncoder().encode(
        "mezo p256verify system test (tampered)",
      )
      const parts = await signWithFreshP256Key(message)
      // Flip a bit in r so the signature stops verifying. Copy first —
      // subarray views share the underlying buffer.
      const tamperedR = new Uint8Array(parts.r)
      tamperedR[0] ^= 0x01
      const payload = buildP256Input({ ...parts, r: tamperedR })
      const [ok, out] = await p256.staticCall.staticCall(P256VERIFY, payload)
      expect(ok, "staticcall reverted").to.be.true
      expect(out).to.equal("0x")
    })
  })
})
