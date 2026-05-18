import { expect } from "chai"
import hre, { ethers } from "hardhat"

/**
 * ClzCheck — pins the EIP-7939 CLZ (count-leading-zeros) opcode at 0x1e.
 *
 * The Hardhat compiler targets evmVersion = "cancun" (see
 * tests/system/hardhat.config.ts), so solc will not emit 0x1e from
 * Solidity source. The test sidesteps the compiler by deploying a raw
 * runtime that wraps the new opcode directly:
 *
 *   PUSH1 0x00  CALLDATALOAD  0x1e (CLZ)  PUSH1 0x00  MSTORE
 *   PUSH1 0x20  PUSH1 0x00    RETURN
 *
 * The contract reads 32 calldata bytes, runs CLZ, and returns the
 * 32-byte result. Encoding any input/expected pair as a JS bigint pair
 * is enough to pin the opcode's spec behavior at the chain surface.
 */

// Runtime: load calldata word -> CLZ -> return 32 bytes.
const RUNTIME = "6000351e60005260206000f3"
// Init: CODECOPY the trailing RUNTIME into memory and RETURN it. The
// constant 0x0c after the third PUSH1 (init length) must match
// RUNTIME.length / 2 so CODECOPY picks up the runtime tail of code.
const INIT = "600c600c600039600c6000f3"
const DEPLOY_BYTECODE = "0x" + INIT + RUNTIME

describe("ClzCheck (EIP-7939)", function () {
  let clz: string

  before(async function () {
    const [signer] = await ethers.getSigners()
    const factory = new ethers.ContractFactory([], DEPLOY_BYTECODE, signer)
    const deployed = await factory.deploy()
    await deployed.waitForDeployment()
    clz = await deployed.getAddress()
  })

  // Reference behavior for CLZ on a 256-bit word:
  //   CLZ(0)         = 256
  //   CLZ(1)         = 255
  //   CLZ(2^255)     = 0
  //   CLZ(2^255 - 1) = 1
  //   CLZ(2^240)     = 15
  //   CLZ(2^128)     = 127
  //   CLZ(2^k)       = 255 - k
  const CASES: Array<{ name: string; input: bigint; expected: bigint }> = [
    { name: "CLZ(0) = 256", input: 0n, expected: 256n },
    { name: "CLZ(1) = 255", input: 1n, expected: 255n },
    { name: "CLZ(2) = 254", input: 2n, expected: 254n },
    { name: "CLZ(0xff) = 248", input: 0xffn, expected: 248n },
    {
      name: "CLZ(2^128) = 127",
      input: 1n << 128n,
      expected: 127n,
    },
    {
      name: "CLZ(2^240) = 15",
      input: 1n << 240n,
      expected: 15n,
    },
    {
      name: "CLZ(2^255) = 0 (high bit set)",
      input: 1n << 255n,
      expected: 0n,
    },
    {
      name: "CLZ(2^256 - 1) = 0 (all bits set)",
      input: (1n << 256n) - 1n,
      expected: 0n,
    },
  ]

  for (const tc of CASES) {
    it(tc.name, async function () {
      const provider = ethers.provider
      const data = ethers.zeroPadValue(ethers.toBeHex(tc.input), 32)
      const ret = await provider.call({ to: clz, data })
      const got = BigInt(ret)
      expect(got).to.equal(tc.expected)
    })
  }
})
