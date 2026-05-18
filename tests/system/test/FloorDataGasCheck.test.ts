import { expect } from "chai"
import hre, { ethers } from "hardhat"

// EIP-7623 calldata gas floor:
//   floor = 21000 + (zeroBytes + nonZeroBytes * 4) * 10
//
// 4096 zero-byte calldata bytes give an intrinsic of 21000 + 4096*4 = 37384
// and a floor of 21000 + 4096*10 = 61960 — well separated so the floor is
// observably distinct from the intrinsic gate.
describe("FloorDataGasCheck", function () {
  const dataLen = 4096
  const calldata = "0x" + "00".repeat(dataLen)
  const floor = 21000n + BigInt(dataLen) * 10n
  let senderSigner: any
  let recipient: any

  before(async function () {
    const signers = await ethers.getSigners()
    senderSigner = signers[1]
    recipient = signers[2]
  })

  it("rejects a tx with gasLimit one below the floor", async function () {
    await expect(
      senderSigner.sendTransaction({
        to: recipient.address,
        data: calldata,
        gasLimit: floor - 1n,
      }),
    ).to.be.rejected
  })

  it("includes a tx with gasLimit equal to the floor", async function () {
    const tx = await senderSigner.sendTransaction({
      to: recipient.address,
      data: calldata,
      gasLimit: floor,
    })
    const receipt = await tx.wait()
    expect(receipt?.status).to.equal(1)
    expect(receipt?.gasUsed).to.be.gte(floor)
  })

  it("charges exactly the floor when calling an EOA", async function () {
    const headroom = 50_000n
    const tx = await senderSigner.sendTransaction({
      to: recipient.address,
      data: calldata,
      gasLimit: floor + headroom,
    })
    const receipt = await tx.wait()
    expect(receipt?.status).to.equal(1)
    // An EOA target executes no code; intrinsic-plus-execution gas stays
    // below the floor, so the EIP-7623 clamp is the binding rule and
    // gasUsed must land on the floor exactly.
    //
    // Mezo's MinGasMultiplier (0.5) can lift gasUsed above the floor when
    // 0.5 * gasLimit > floor. Pick `headroom` small enough that
    // floor + headroom < 2 * floor to keep the floor the dominant clamp.
    expect(receipt?.gasUsed).to.equal(floor)
  })
})
