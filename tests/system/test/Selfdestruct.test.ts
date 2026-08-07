import { expect } from "chai"
import hre, { ethers } from "hardhat"
import maintenanceAbi from "../../../precompile/maintenance/abi.json"
import { getDeployedContract } from "./helpers/contract"

// The maintenance precompile (0x7b7c...0013) exposes the SELFDESTRUCT toggle.
const maintenanceAddress = "0x7b7c000000000000000000000000000000000013"

// EIP-90000 disables SELFDESTRUCT at genesis. The maintenance precompile lets
// the PoA owner toggle it at runtime.
describe("Selfdestruct", function () {
  const { deployments } = hre
  let selfdestructCheck: any
  let owner: any
  let sender: any
  let beneficiary: any

  before(async function () {
    await deployments.fixture(["Selfdestruct6780Check"])
    selfdestructCheck = await getDeployedContract("Selfdestruct6780Check")

    const signers = await ethers.getSigners()
    owner = signers[0]
    sender = signers[1]
    beneficiary = signers[2]
  })

  const maintenanceAs = (signer: any) =>
    new ethers.Contract(maintenanceAddress, maintenanceAbi, signer)

  // Leave the chain in its default (disabled) state for other suites.
  after(async function () {
    await (await maintenanceAs(owner).setSelfDestructDisabled(true)).wait()
  })

  it("reports SELFDESTRUCT disabled by default", async function () {
    expect(await maintenanceAs(owner).getSelfDestructDisabled()).to.equal(true)
  })

  it("rejects deploying a contract that self-destructs in its constructor", async function () {
    // PUSH1 0x00; SELFDESTRUCT -- init code that runs SELFDESTRUCT immediately.
    const initCode = "0x6000ff"

    await expect(
      (async () => {
        const tx = await sender.sendTransaction({
          data: initCode,
          gasLimit: 200_000,
        })
        await tx.wait()
      })(),
    ).to.be.rejectedWith("transaction execution reverted")
  })

  it("rejects calling a self-destruct method while disabled", async function () {
    const createTx = await selfdestructCheck
      .connect(sender)
      .createDestructible()
    await createTx.wait()
    const destructible = await selfdestructCheck.lastDestructible()

    await expect(
      (async () => {
        const tx = await selfdestructCheck
          .connect(sender)
          .destroyExisting(destructible, beneficiary.address, {
            gasLimit: 200_000,
          })
        await tx.wait()
      })(),
    ).to.be.rejectedWith("transaction execution reverted")
  })

  it("rejects SELFDESTRUCT in a contract deployed before it is disabled", async function () {
    await (await maintenanceAs(owner).setSelfDestructDisabled(false)).wait()

    const createTx = await selfdestructCheck
      .connect(sender)
      .createDestructible()
    await createTx.wait()
    const destructible = await selfdestructCheck.lastDestructible()

    await (await maintenanceAs(owner).setSelfDestructDisabled(true)).wait()

    await expect(
      (async () => {
        const tx = await selfdestructCheck
          .connect(sender)
          .destroyExisting(destructible, beneficiary.address, {
            gasLimit: 200_000,
          })
        await tx.wait()
      })(),
    ).to.be.rejectedWith("transaction execution reverted")
  })

  it("allows SELFDESTRUCT once enabled, and blocks it again once disabled", async function () {
    const createTx = await selfdestructCheck
      .connect(sender)
      .createDestructible()
    await createTx.wait()
    const destructible = await selfdestructCheck.lastDestructible()

    // Enable the opcode -> destroy runs successfully.
    await (await maintenanceAs(owner).setSelfDestructDisabled(false)).wait()
    expect(await maintenanceAs(owner).getSelfDestructDisabled()).to.equal(false)

    const okTx = await selfdestructCheck
      .connect(sender)
      .destroyExisting(destructible, beneficiary.address, {
        gasLimit: 200_000,
      })
    const receipt = await okTx.wait()
    expect(receipt.status).to.equal(1)

    // Disable again -> the same contract's destroy reverts.
    await (await maintenanceAs(owner).setSelfDestructDisabled(true)).wait()
    await expect(
      (async () => {
        const tx = await selfdestructCheck
          .connect(sender)
          .destroyExisting(destructible, beneficiary.address, {
            gasLimit: 200_000,
          })
        await tx.wait()
      })(),
    ).to.be.rejectedWith("transaction execution reverted")
  })
})
