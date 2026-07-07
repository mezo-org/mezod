import { expect } from "chai"
import hre from "hardhat"
import type { Bls12381Check } from "../typechain-types"
import { getDeployedContract } from "./helpers/contract"

const G1 = 128
const G2 = 256
const FP = 64
const FP2 = 128
const SCALAR = 32

const ZERO_HEX = (n: number) => "0x" + "00".repeat(n)

type Slot = {
  addr: string
  name: string
  validInput: () => string
  validOutputSize: number
  expectedOutput?: string
  wrongSize: () => string
}

const SLOTS: Slot[] = [
  {
    addr: hre.ethers.getAddress("0x000000000000000000000000000000000000000b"),
    name: "G1Add",
    validInput: () => ZERO_HEX(2 * G1),
    validOutputSize: G1,
    expectedOutput: ZERO_HEX(G1),
    // Draft G1Mul payload shape; rejected here proves the slot is final G1Add.
    wrongSize: () => ZERO_HEX(G1 + SCALAR),
  },
  {
    addr: hre.ethers.getAddress("0x000000000000000000000000000000000000000c"),
    name: "G1MSM",
    validInput: () => ZERO_HEX(G1 + SCALAR),
    validOutputSize: G1,
    expectedOutput: ZERO_HEX(G1),
    wrongSize: () => ZERO_HEX(2 * G1),
  },
  {
    addr: hre.ethers.getAddress("0x000000000000000000000000000000000000000d"),
    name: "G2Add",
    validInput: () => ZERO_HEX(2 * G2),
    validOutputSize: G2,
    expectedOutput: ZERO_HEX(G2),
    // Draft G2Mul payload shape; rejected here proves the slot is final G2Add.
    wrongSize: () => ZERO_HEX(G2 + SCALAR),
  },
  {
    addr: hre.ethers.getAddress("0x000000000000000000000000000000000000000e"),
    name: "G2MSM",
    validInput: () => ZERO_HEX(G2 + SCALAR),
    validOutputSize: G2,
    expectedOutput: ZERO_HEX(G2),
    wrongSize: () => ZERO_HEX(2 * G2),
  },
  {
    addr: hre.ethers.getAddress("0x000000000000000000000000000000000000000f"),
    name: "PairingCheck",
    validInput: () => ZERO_HEX(G1 + G2),
    validOutputSize: 32,
    expectedOutput: "0x" + "00".repeat(31) + "01",
    wrongSize: () => ZERO_HEX(200),
  },
  {
    addr: hre.ethers.getAddress("0x0000000000000000000000000000000000000010"),
    name: "MapFpToG1",
    validInput: () => ZERO_HEX(FP),
    validOutputSize: G1,
    wrongSize: () => ZERO_HEX(FP2),
  },
  {
    addr: hre.ethers.getAddress("0x0000000000000000000000000000000000000011"),
    name: "MapFp2ToG2",
    validInput: () => ZERO_HEX(FP2),
    validOutputSize: G2,
    wrongSize: () => ZERO_HEX(FP),
  },
]

const EMPTY_SLOTS = [
  hre.ethers.getAddress("0x0000000000000000000000000000000000000012"),
  hre.ethers.getAddress("0x0000000000000000000000000000000000000013"),
]

describe("Bls12381Check (EIP-2537 layout)", function () {
  const { deployments } = hre
  let bls: Bls12381Check

  before(async function () {
    await deployments.fixture(["Bls12381Check"])
    bls = await getDeployedContract("Bls12381Check")
  })

  for (const slot of SLOTS) {
    it(`${slot.name} at ${slot.addr} accepts valid input`, async function () {
      const [ok, out] = await bls.staticCall.staticCall(
        slot.addr,
        slot.validInput(),
      )
      expect(ok, "staticcall reverted").to.be.true
      expect(hre.ethers.dataLength(out)).to.equal(slot.validOutputSize)
      if (slot.expectedOutput !== undefined) {
        expect(out).to.equal(slot.expectedOutput)
      }
    })
  }

  for (const slot of SLOTS) {
    it(`${slot.name} reverts on wrong-size payload`, async function () {
      const [ok] = await bls.staticCall.staticCall(slot.addr, slot.wrongSize())
      expect(ok, "wrong-size payload accepted").to.be.false
    })
  }

  for (const emptyAddr of EMPTY_SLOTS) {
    it(`${emptyAddr} is not a precompile (empty account)`, async function () {
      const [ok, out] = await bls.staticCall.staticCall(emptyAddr, ZERO_HEX(64))
      expect(ok, "staticcall to empty account reverted").to.be.true
      expect(out).to.equal("0x")
    })
  }
})
