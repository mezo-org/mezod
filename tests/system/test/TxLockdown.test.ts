import { expect } from "chai"
import { ethers } from "hardhat"
import maintenanceabi from "../../../precompile/maintenance/abi.json"
import validatorpoolabi from "../../../precompile/validatorpool/abi.json"
import { extractMessage } from "./helpers/rpc-error"

const validatorPoolPrecompileAddress =
  "0x7b7c000000000000000000000000000000000011"
const maintenancePrecompileAddress =
  "0x7b7c000000000000000000000000000000000013"

// The initcode of a contract with an empty runtime code. The deployment tests
// use it to check the contract creation path of the lockdown rule.
const emptyContractInitcode = "0x60006000f3"

// The gas limit of the transactions the lockdown is expected to reject. The
// explicit limit skips the gas estimation, which runs as a query and bypasses
// the ante handler enforcing the lockdown.
const rejectedTxGasLimit = 200000

// The transaction lockdown (levels 2 and 3 of the graduated lockdown) rejects
// every transaction that does not pass the lockdown rule. The rule always
// admits the PoA owner and the Emergency Team, as sender or as target, and
// admits the extra allowlisted pairs on top of that.
describe("TxLockdown", function () {
  let maintenance: any
  let validatorPool: any
  let poolOwner: any
  let emergencyTeam: any
  let thirdParty: any
  let fourthParty: any
  let allowedTarget: string
  let blockedTarget: string

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

  // expectRejected asserts that the lockdown rejects the transaction the given
  // function sends. The rejection surfaces as an RPC error at CheckTx.
  const expectRejected = async (send: () => Promise<any>) => {
    let errorMessage: string = ""

    try {
      const tx = await send()
      await tx.wait()
    } catch (error: any) {
      errorMessage = extractMessage(error)
    }

    expect(errorMessage).to.include("rejected by the transaction lockdown")
  }

  // callMaintenance sends a transaction that calls the named maintenance
  // precompile method. Unlike a contract call through ethers, it never runs
  // the gas estimation, so the lockdown is the only thing that can reject it.
  const callMaintenance = (signer: any, method: string) =>
    signer.sendTransaction({
      to: maintenancePrecompileAddress,
      data: new ethers.Interface(maintenanceabi).encodeFunctionData(method),
      gasLimit: rejectedTxGasLimit,
    })

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

    // The localnode exposes three funded accounts. The fourth party is a fresh
    // account funded by the owner, whose transactions always pass the rule.
    fourthParty = ethers.Wallet.createRandom().connect(ethers.provider)
    await (
      await poolOwner.sendTransaction({
        to: fourthParty.address,
        value: ethers.parseEther("1"),
      })
    ).wait()

    allowedTarget = ethers.Wallet.createRandom().address
    blockedTarget = ethers.Wallet.createRandom().address
  })

  // Leave the chain without a lockdown and without an Emergency Team for the
  // other suites.
  after(async function () {
    await (await maintenanceAs(poolOwner).setTxLockdown(false)).wait()
    await (
      await maintenanceAs(poolOwner).setEmergencyTeam(ethers.ZeroAddress)
    ).wait()
  })

  it("reports no lockdown by default", async function () {
    expect(await maintenance.getTxLockdown()).to.equal(false)
  })

  it("reports an empty allowlist by default", async function () {
    const allowlist = await maintenance.getTxLockdownAllowlist()
    expect(allowlist.senders).to.deep.equal([])
    expect(allowlist.targets).to.deep.equal([])
  })

  it("lets the owner grant the emergency team role", async function () {
    const tx = await maintenanceAs(poolOwner).setEmergencyTeam(
      emergencyTeam.address,
    )
    const receipt = await tx.wait()

    expect(receipt.status).to.equal(1)
    expect(await maintenance.getEmergencyTeam()).to.equal(emergencyTeam.address)
  })

  it("lets the emergency team enable the lockdown", async function () {
    const tx = await maintenanceAs(emergencyTeam).setTxLockdown(true)
    const receipt = await tx.wait()

    expect(receipt.status).to.equal(1)

    const args = eventArguments(receipt, "TxLockdownSet")
    expect(args.enabled).to.equal(true)

    expect(await maintenance.getTxLockdown()).to.equal(true)
  })

  it("rejects a third-party transfer", async function () {
    await expectRejected(() =>
      thirdParty.sendTransaction({
        to: blockedTarget,
        value: 1,
        gasLimit: rejectedTxGasLimit,
      }),
    )
  })

  it("rejects a third-party call to the maintenance precompile", async function () {
    // There is no default target allowlist. The maintenance precompile is not
    // reachable for a third party while the lockdown is active.
    await expectRejected(() => callMaintenance(thirdParty, "getTxLockdown"))
  })

  it("rejects a third-party contract deployment", async function () {
    await expectRejected(() =>
      thirdParty.sendTransaction({
        data: emptyContractInitcode,
        gasLimit: rejectedTxGasLimit,
      }),
    )
  })

  it("lets a third party send to a role address", async function () {
    const tx = await thirdParty.sendTransaction({
      to: poolOwner.address,
      value: 1,
    })
    const receipt = await tx.wait()

    expect(receipt.status).to.equal(1)
  })

  it("lets the emergency team call the maintenance precompile", async function () {
    const tx = await callMaintenance(emergencyTeam, "getTxLockdown")
    const receipt = await tx.wait()

    expect(receipt.status).to.equal(1)
  })

  it("lets the emergency team set both allowlist dimensions", async function () {
    const tx = await maintenanceAs(emergencyTeam).setTxLockdownAllowlist(
      [thirdParty.address],
      [allowedTarget],
    )
    const receipt = await tx.wait()

    expect(receipt.status).to.equal(1)

    const args = eventArguments(receipt, "TxLockdownAllowlistSet")
    expect(args.senders).to.deep.equal([thirdParty.address])
    expect(args.targets).to.deep.equal([allowedTarget])

    const allowlist = await maintenance.getTxLockdownAllowlist()
    expect(allowlist.senders).to.deep.equal([thirdParty.address])
    expect(allowlist.targets).to.deep.equal([allowedTarget])
  })

  it("lets the allowlisted pair through", async function () {
    const tx = await thirdParty.sendTransaction({
      to: allowedTarget,
      value: 1,
    })
    const receipt = await tx.wait()

    expect(receipt.status).to.equal(1)
  })

  it("rejects the allowlisted sender at another target", async function () {
    await expectRejected(() =>
      thirdParty.sendTransaction({
        to: blockedTarget,
        value: 1,
        gasLimit: rejectedTxGasLimit,
      }),
    )
  })

  it("rejects the allowlisted sender deploying a contract", async function () {
    // A nil target never matches a non-empty target allowlist.
    await expectRejected(() =>
      thirdParty.sendTransaction({
        data: emptyContractInitcode,
        gasLimit: rejectedTxGasLimit,
      }),
    )
  })

  it("lets the emergency team empty the sender dimension", async function () {
    const tx = await maintenanceAs(emergencyTeam).setTxLockdownAllowlist(
      [],
      [allowedTarget],
    )
    const receipt = await tx.wait()

    expect(receipt.status).to.equal(1)

    const allowlist = await maintenance.getTxLockdownAllowlist()
    expect(allowlist.senders).to.deep.equal([])
    expect(allowlist.targets).to.deep.equal([allowedTarget])
  })

  it("lets any sender reach the allowlisted target", async function () {
    const tx = await fourthParty.sendTransaction({
      to: allowedTarget,
      value: 1,
    })
    const receipt = await tx.wait()

    expect(receipt.status).to.equal(1)
  })

  it("rejects that sender at another target", async function () {
    await expectRejected(() =>
      fourthParty.sendTransaction({
        to: blockedTarget,
        value: 1,
        gasLimit: rejectedTxGasLimit,
      }),
    )
  })

  it("lets the emergency team empty the target dimension", async function () {
    const tx = await maintenanceAs(emergencyTeam).setTxLockdownAllowlist(
      [thirdParty.address],
      [],
    )
    const receipt = await tx.wait()

    expect(receipt.status).to.equal(1)

    const allowlist = await maintenance.getTxLockdownAllowlist()
    expect(allowlist.senders).to.deep.equal([thirdParty.address])
    expect(allowlist.targets).to.deep.equal([])
  })

  it("lets the allowlisted sender reach any target", async function () {
    const tx = await thirdParty.sendTransaction({
      to: blockedTarget,
      value: 1,
    })
    const receipt = await tx.wait()

    expect(receipt.status).to.equal(1)
  })

  it("lets the allowlisted sender deploy a contract", async function () {
    const tx = await thirdParty.sendTransaction({
      data: emptyContractInitcode,
    })
    const receipt = await tx.wait()

    expect(receipt.status).to.equal(1)
    expect(receipt.contractAddress).to.not.equal(null)
  })

  it("rejects a sender outside the allowlist", async function () {
    await expectRejected(() =>
      fourthParty.sendTransaction({
        to: blockedTarget,
        value: 1,
        gasLimit: rejectedTxGasLimit,
      }),
    )
  })

  it("clears the allowlist when the owner disables the lockdown", async function () {
    const tx = await maintenanceAs(poolOwner).setTxLockdown(false)
    const receipt = await tx.wait()

    const args = eventArguments(receipt, "TxLockdownSet")
    expect(args.enabled).to.equal(false)

    expect(await maintenance.getTxLockdown()).to.equal(false)

    const allowlist = await maintenance.getTxLockdownAllowlist()
    expect(allowlist.senders).to.deep.equal([])
    expect(allowlist.targets).to.deep.equal([])
  })

  it("lets the traffic flow after the lockdown is disabled", async function () {
    const tx = await fourthParty.sendTransaction({
      to: blockedTarget,
      value: 1,
    })
    const receipt = await tx.wait()

    expect(receipt.status).to.equal(1)
  })

  it("rejects the lockdown from an account without a role", async function () {
    let errorMessage: string = ""

    try {
      await maintenanceAs(thirdParty).setTxLockdown.staticCall(true)
    } catch (error: any) {
      errorMessage = extractMessage(error)
    }

    expect(errorMessage).to.include("not the emergency team")
    expect(await maintenance.getTxLockdown()).to.equal(false)
  })

  it("rejects the allowlist from an account without a role", async function () {
    let errorMessage: string = ""

    try {
      await maintenanceAs(thirdParty).setTxLockdownAllowlist.staticCall(
        [thirdParty.address],
        [],
      )
    } catch (error: any) {
      errorMessage = extractMessage(error)
    }

    expect(errorMessage).to.include("not the emergency team")
  })
})
