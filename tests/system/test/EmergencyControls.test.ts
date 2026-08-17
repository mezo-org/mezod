import { expect } from "chai"
import { ethers } from "hardhat"
import maintenanceabi from "../../../precompile/maintenance/abi.json"
import validatorpoolabi from "../../../precompile/validatorpool/abi.json"

const validatorPoolPrecompileAddress =
  "0x7b7c000000000000000000000000000000000011"
const assetsBridgePrecompileAddress =
  "0x7b7c000000000000000000000000000000000012"
const maintenancePrecompileAddress =
  "0x7b7c000000000000000000000000000000000013"

// The maintenance precompile (0x7b7c...0013) is the single emergency surface:
// the PoA owner grants the Emergency Team role there and both the owner and the
// Emergency Team drive the bridge lockdown from there.
describe("EmergencyControls", function () {
  let maintenance: any
  let validatorPool: any
  let poolOwner: any
  let emergencyTeam: any
  let thirdParty: any

  const maintenanceAs = (signer: any) =>
    new ethers.Contract(maintenancePrecompileAddress, maintenanceabi, signer)

  // eventArguments returns the arguments of the named event emitted by the
  // transaction. It throws when the transaction does not emit that event.
  const eventArguments = (receipt: any, name: string) => {
    const iface = new ethers.Interface(maintenanceabi)

    for (const log of receipt.logs) {
      try {
        const parsed = iface.parseLog({
          topics: log.topics as string[],
          data: log.data,
        })
        if (parsed && parsed.name === name) {
          return parsed.args
        }
      } catch {
        // Not a decodable event from our ABI, skip.
      }
    }

    throw new Error(`the transaction did not emit the ${name} event`)
  }

  before(async function () {
    const signers = await ethers.getSigners()

    validatorPool = new ethers.Contract(
      validatorPoolPrecompileAddress,
      validatorpoolabi,
      ethers.provider,
    )
    maintenance = new ethers.Contract(
      maintenancePrecompileAddress,
      maintenanceabi,
      ethers.provider,
    )

    poolOwner = await ethers.getSigner(await validatorPool.owner())
    emergencyTeam = signers[1]
    thirdParty = signers[2]
  })

  // Leave the chain without a lockdown and without an Emergency Team for the
  // other suites.
  after(async function () {
    await (
      await maintenanceAs(poolOwner).setBridgeLockdown(false, false)
    ).wait()
    await (
      await maintenanceAs(poolOwner).setEmergencyTeam(ethers.ZeroAddress)
    ).wait()
  })

  it("reports no emergency team by default", async function () {
    expect(await maintenance.getEmergencyTeam()).to.equal(ethers.ZeroAddress)
  })

  it("reports no lockdown by default", async function () {
    const lockdown = await maintenance.getBridgeLockdown()
    expect(lockdown.bridgeIn).to.equal(false)
    expect(lockdown.bridgeOut).to.equal(false)
  })

  it("rejects granting the role from a non-owner", async function () {
    let errorMessage: string = ""

    try {
      await maintenanceAs(thirdParty).setEmergencyTeam.staticCall(
        thirdParty.address,
      )
    } catch (error: any) {
      errorMessage = error.message
    }

    expect(errorMessage).to.include("not the owner")
    expect(await maintenance.getEmergencyTeam()).to.equal(ethers.ZeroAddress)
  })

  it("rejects the lockdown before the role is granted", async function () {
    let errorMessage: string = ""

    try {
      await maintenanceAs(emergencyTeam).setBridgeLockdown.staticCall(
        true,
        true,
      )
    } catch (error: any) {
      errorMessage = error.message
    }

    expect(errorMessage).to.include("emergency team address is empty")
  })

  it("lets the owner grant the role", async function () {
    const tx = await maintenanceAs(poolOwner).setEmergencyTeam(
      emergencyTeam.address,
    )
    const receipt = await tx.wait()

    expect(receipt.status).to.equal(1)

    const args = eventArguments(receipt, "EmergencyTeamSet")
    expect(args.previous).to.equal(ethers.ZeroAddress)
    expect(args.current).to.equal(emergencyTeam.address)

    expect(await maintenance.getEmergencyTeam()).to.equal(
      emergencyTeam.address,
    )
  })

  it("lets the emergency team enable the lockdown", async function () {
    const tx = await maintenanceAs(emergencyTeam).setBridgeLockdown(true, true)
    const receipt = await tx.wait()

    expect(receipt.status).to.equal(1)

    const args = eventArguments(receipt, "BridgeLockdownSet")
    expect(args.bridgeIn).to.equal(true)
    expect(args.bridgeOut).to.equal(true)

    const lockdown = await maintenance.getBridgeLockdown()
    expect(lockdown.bridgeIn).to.equal(true)
    expect(lockdown.bridgeOut).to.equal(true)
  })

  it("rejects the lockdown from an account without the role", async function () {
    let errorMessage: string = ""

    try {
      await maintenanceAs(thirdParty).setBridgeLockdown.staticCall(false, false)
    } catch (error: any) {
      errorMessage = error.message
    }

    expect(errorMessage).to.include("not the emergency team")

    // The lockdown stays enabled.
    const lockdown = await maintenance.getBridgeLockdown()
    expect(lockdown.bridgeIn).to.equal(true)
    expect(lockdown.bridgeOut).to.equal(true)
  })

  it("lets the emergency team narrow the lockdown to the bridge-out", async function () {
    const tx = await maintenanceAs(emergencyTeam).setBridgeLockdown(false, true)
    const receipt = await tx.wait()

    const args = eventArguments(receipt, "BridgeLockdownSet")
    expect(args.bridgeIn).to.equal(false)
    expect(args.bridgeOut).to.equal(true)

    const lockdown = await maintenance.getBridgeLockdown()
    expect(lockdown.bridgeIn).to.equal(false)
    expect(lockdown.bridgeOut).to.equal(true)
  })

  it("lets the owner disable the lockdown", async function () {
    const tx = await maintenanceAs(poolOwner).setBridgeLockdown(false, false)
    const receipt = await tx.wait()

    const args = eventArguments(receipt, "BridgeLockdownSet")
    expect(args.bridgeIn).to.equal(false)
    expect(args.bridgeOut).to.equal(false)

    const lockdown = await maintenance.getBridgeLockdown()
    expect(lockdown.bridgeIn).to.equal(false)
    expect(lockdown.bridgeOut).to.equal(false)
  })

  it("lets the owner revoke the role", async function () {
    const tx = await maintenanceAs(poolOwner).setEmergencyTeam(
      ethers.ZeroAddress,
    )
    const receipt = await tx.wait()

    const args = eventArguments(receipt, "EmergencyTeamSet")
    expect(args.previous).to.equal(emergencyTeam.address)
    expect(args.current).to.equal(ethers.ZeroAddress)

    expect(await maintenance.getEmergencyTeam()).to.equal(ethers.ZeroAddress)
  })

  it("rejects the lockdown after the role is revoked", async function () {
    let errorMessage: string = ""

    try {
      await maintenanceAs(emergencyTeam).setBridgeLockdown.staticCall(
        true,
        true,
      )
    } catch (error: any) {
      errorMessage = error.message
    }

    expect(errorMessage).to.include("emergency team address is empty")
  })

  describe("retired assets bridge methods", function () {
    // The retired methods are removed from the assets bridge ABI, so the test
    // calls their raw 4-byte selectors.
    const retiredSelectors: Array<[string, string]> = [
      ["setPauser(address)", "0x2d88af4a"],
      ["getPauser()", "0x7008b548"],
      ["pauseBridgeOut()", "0x2f1d448f"],
      ["pauseTriparty(bool)", "0x9afc02ff"],
      ["isTripartyPaused()", "0x5d1b3fc4"],
    ]

    retiredSelectors.forEach(([signature, selector]) => {
      it(`reverts ${signature}`, async function () {
        let errorMessage: string = ""

        try {
          await ethers.provider.call({
            to: assetsBridgePrecompileAddress,
            data: selector,
          })
        } catch (error: any) {
          errorMessage = error.message
        }

        expect(errorMessage).to.include("method not found in ABI")
      })
    })
  })
})
